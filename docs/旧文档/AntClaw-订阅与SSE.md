# AntClaw · 订阅与实时事件（SSE / gRPC Stream）

> 本文是 `AntClaw-重构解决方案.md` §3.4、§十三A（流事件）的可实现细化。约束范围：**所有实时事件**（告警、信号、行情 tick、任务进度、简报）在 Web 与移动端的统一契约与传输实现。

## 一、原则

1. **统一契约**：所有事件定义在 `proto/antclaw/v1/stream.proto`；Web（SSE）与移动端（gRPC server-streaming）共用同一 schema，只编译一次。
2. **禁用 WebSocket**：Web 走 SSE；移动走 gRPC server-streaming。
3. **至少一次交付 + 客户端幂等**：事件均携带 `event_id`，客户端去重。
4. **断点续传**：Web 使用 `Last-Event-ID`，移动使用请求字段 `resume_from`。
5. **SLA**：`alert`/`signal` 端到端 P95 ≤ 500 ms；`price_tick` ≤ 300 ms；`task_progress` ≤ 1s。
6. **背压策略**：SSE 满队则**丢最旧保最新**；单条事件 ≥ 64KiB 先落对象存储，只推 `snapshot_uri`。

## 二、事件模型（`stream.proto`）

### 2.1 信封

```proto
message StreamEvent {
  string event_id     = 1;   // 单调递增的字符串 id，同一 channel 内全序
  StreamChannel channel = 2; // 枚举见 §2.2
  google.protobuf.Timestamp emitted_at = 3;
  string trace_id     = 4;
  oneof payload {
    AlertEvent         alert         = 10;
    SignalEvent        signal        = 11;
    PriceTickEvent     price_tick    = 12;
    TaskProgressEvent  task_progress = 13;
    BriefingEvent      briefing      = 14;
    RegimeEvent        regime        = 15;
    CarryEvent         carry         = 16;
    SkewVixEvent       skew_vix      = 17;
    CotEvent           cot           = 18;
    SystemNoticeEvent  system_notice = 19;
  }
  string snapshot_uri = 30;  // 大 payload 时启用；客户端需 GET 完成
}
```

### 2.2 `StreamChannel` 枚举（注册表）

| 枚举 | 字符串 key | 用途 | SLA |
|---|---|---|---|
| `CHANNEL_ALERTS` | `alerts` | 用户订阅告警 | 500 ms |
| `CHANNEL_SIGNALS` | `signals` | 系统信号 | 500 ms |
| `CHANNEL_PRICE_TICKS` | `price_ticks` | 行情 tick（按 symbol 订阅） | 300 ms |
| `CHANNEL_TASK_PROGRESS` | `tasks` | 回测/任务阶段 | 1 s |
| `CHANNEL_BRIEFING` | `briefing` | 日报/简报 | 不计 |
| `CHANNEL_REGIME` | `regime` | 宏观体制切换 | 1 s |
| `CHANNEL_CARRY` | `carry` | 利差监测 | 1 s |
| `CHANNEL_SKEW_VIX` | `skew_vix` | 偏度/VIX 告警 | 500 ms |
| `CHANNEL_COT` | `cot` | COT 更新 | 不计 |
| `CHANNEL_SYSTEM_NOTICE` | `system_notice` | 系统通告（维护、强制登出） | 1 s |

> 新增 channel 必须在此表、`stream.proto` 枚举、`AntClaw-重构解决方案.md` 附录 B.1 同步登记（宪章 9）。

## 三、总线：Redis Streams

### 3.1 Stream key 规范

```
ev:<channel>:<scope>
```

- `scope` 为 `global` 或 `user:<user_id>` 或 `symbol:<sym>` 或 `task:<task_id>`。
- 例：`ev:alerts:user:abc-123`、`ev:price_ticks:symbol:EURUSD`、`ev:tasks:task:t_001`。

### 3.2 生产者

- 使用 `XADD ev:<channel>:<scope> MAXLEN ~ <N> * event_id <id> payload <protojson>`。
- `MAXLEN`（默认）：alerts/signals 10_000；price_ticks 5_000；tasks 1_000；briefing 500。
- `event_id` = `<unix_ms>-<seq>`；seq 在生产者进程内单调递增，保证同一 stream 内全序。
- 生产者必须在本地事务成功提交后再 `XADD`（**outbox 模式**）：业务事务写 `event_outbox` 表；后台 `outbox_publisher` 扫描并 `XADD`，成功即 `DELETE` outbox 行。

### 3.3 消费者（SSE 网关）

- 每进程用 **Consumer Group**：`grp:sse:<pod_id>`；
- 读：`XREADGROUP GROUP grp:sse:<pod> consumer:<conn_id> COUNT 64 BLOCK 1000 STREAMS <key> >`；
- Ack：事件成功写入 HTTP stream 后 `XACK`；未 Ack 事件由 `XCLAIM` 回收（超时 60s）。
- 历史回放：`XREVRANGE ev:<channel>:<scope> + <last_event_id>` 最多 500 条。

## 四、Web：SSE 网关

### 4.1 端点

```
GET /sse/v1/stream?channels=<csv>&token=<JWT>
Accept: text/event-stream
Last-Event-ID: <opaque>   // 断线重连时浏览器自动发送
```

- 也支持 Cookie 鉴权；`token` query 仅为 `EventSource` 不支持自定义 header 的兼容手段。
- 一次请求可订阅多个 channel，由 `channels` CSV 指定；服务端按角色与订阅清单过滤。

### 4.2 帧格式

```
id: 1700000000000-7
event: alerts
data: {"event_id":"1700000000000-7","channel":"CHANNEL_ALERTS","emitted_at":"...","alert":{...}}

```

- `event` 字段 = channel 字符串 key（小写）。
- `data` 为 protojson 序列化的 `StreamEvent`。
- 单事件 `data:` 段超过 64KiB 时，`snapshot_uri` 生效，`alert/signal/...` 内仅保留摘要字段。
- 心跳：每 15s 发 `: ping\n\n` 注释帧（不触发 `onmessage`）。

### 4.3 断点续传 `Last-Event-ID`

- 游标表 `sse_cursors(user_id, channel, scope, last_event_id, updated_at)` 用于**跨实例持久化**；服务端收到 `Last-Event-ID` 时，对比游标表与 Redis Stream 最早 id：
  - 若客户端游标 < stream 最早 id → 推 `system_notice.cursor_reset`，客户端需整屏刷新一次。
  - 否则用 `XREVRANGE` 回补，再切到实时。
- 仅在客户端显式消费成功（前端会回发 `POST /sse/v1/ack`，body `{channel, event_id}`，也接受批量）后才更新 `sse_cursors`；未 ack 的重连会重放。

### 4.4 连接管控

- **并发**：每用户按角色上限（见《用户系统与鉴权》§7.2）；超限旧连接被踢。
- **空闲**：单连接空闲（无事件、无心跳 ack）120s 自动关闭；客户端自动重连。
- **负载均衡**：SSE 反向代理 Caddy `flush_interval 0`；禁止响应缓冲。

## 五、移动端：gRPC server-streaming

### 5.1 服务方法

```proto
service StreamService {
  rpc Subscribe(SubscribeRequest) returns (stream StreamEvent);
}

message SubscribeRequest {
  repeated StreamChannel channels = 1;
  string resume_from = 2;       // = 最后一次收到的 event_id
  repeated string symbols = 3;  // 仅 CHANNEL_PRICE_TICKS 生效
  string task_id = 4;           // 仅 CHANNEL_TASK_PROGRESS 生效
}
```

- 由同一套 SSE 网关实例承担；Connect/gRPC handler 与 SSE handler 共享订阅引擎。
- 心跳：gRPC 层面每 30s 发 keepalive ping；业务层面每 15s 发 `system_notice.heartbeat`（仅内部调试信道可关闭）。

### 5.2 幂等与去重

- 客户端持久化 `last_event_id` per channel；重连用 `resume_from`。
- 服务端允许 **at-least-once**；客户端以 `event_id` 去重（`SET<256>` LRU）。

## 六、订阅管理（持久化配置）

### 6.1 `AlertService` RPC（节选）

- `ListSubscriptions(user_id)`
- `Subscribe(channel, filter)`：filter 由 channel 决定字段，例：
  - `alerts`：`pair`、`severity ≥ medium`、`quiet_hours`。
  - `price_ticks`：`symbols[]`、`min_move_pips`。
  - `tasks`：`owner_only = true`（默认）。
- `Unsubscribe(subscription_id)`
- `UpdateSubscription(subscription_id, filter)`

### 6.2 `alert_subscriptions` 表

| 列 | 说明 |
|---|---|
| `id` | UUIDv7 |
| `user_id` | FK |
| `channel` | 字符串 key |
| `filter_json` | JSONB；按 channel 固定 schema，由 worker 读取 |
| `cooldown_sec` | 整型；同一 filter 触发冷却窗口 |
| `quiet_hours` | JSONB：`{tz, ranges:[{from:"22:00",to:"07:00"}]}` |
| `created_at` / `updated_at` | — |

索引：`INDEX(user_id, channel)`、`INDEX(channel)`（扇出时按 channel 扫订阅者）。

### 6.3 匹配管线

```
[业务事件] → worker 匹配 alert_subscriptions (filter_json + cooldown + quiet_hours)
           → 命中 → 写 event_outbox → outbox_publisher → XADD ev:alerts:user:<uid>
           → SSE/gRPC 网关推送
           → 若用户离线 → 同时写站内信 notifications 表
```

- 冷却：`alerts:cooldown:<sub_id>:<filter_hash>` Redis key，TTL = `cooldown_sec`。
- 静默：服务端对比 `quiet_hours` 与事件时间，静默命中则只落站内信，不推流。

## 七、背压与大负载

- **丢弃策略**：当某 SSE 连接写入慢到积压 ≥ 100 条时，服务端**丢最旧**并计数 `stream_dropped_total{channel,reason="slow_consumer"}`；随后补发一条 `system_notice.dropped` 通知客户端。
- **大 payload**：`protojson.Marshal` 结果 > 64KiB 必须走对象存储：写 MinIO `stream-snapshots`，key = `snapshots/<yyyy-mm-dd>/<event_id>.json`；事件 `snapshot_uri` 指向预签名 GET URL（TTL 5min）；推流只含摘要。

## 八、观测

- 指标（Prometheus）：
  - `stream_events_published_total{channel,scope}`
  - `stream_events_delivered_total{channel,transport}`（transport=sse|grpc）
  - `stream_delivery_latency_seconds_bucket{channel}`（outbox 写入 → 客户端 ack）
  - `stream_active_connections{transport}`
  - `stream_dropped_total{channel,reason}`
- 追踪：`traceparent` 透传；每事件一条 `stream.publish` + `stream.deliver` span。
- 日志：每连接一次握手日志，含 `user_id`、`channels`、`resume_from`；事件级日志仅异常路径。

## 九、安全

- 订阅前统一校验：`ctx.user_id` 必须匹配 `scope=user:<user_id>`；`admin` 才可订阅 `scope=global`。
- `channels` 参数与用户角色对照白名单；超界直接 `PERMISSION_DENIED`。
- `token` 在 query 中的使用仅限 `EventSource` 初次握手，服务端收到后立即交换为 Cookie 会话；拒绝 token 在日志中出现。

## 十、验收清单（对照任务卡 P6）

- [ ] `stream.proto` 编译通过；生成物同时服务 Web 与移动端。
- [ ] k6 SSE 压测脚本（规划于 `scripts/loadtest/`，应对齐当前路径如 `/sse/jobs`；原 `sse.js` 占位已清场）：1k 并发，P95 延迟达标。
- [ ] 断线重连：`Last-Event-ID` 可回补最近 500 条，超出则触发 `cursor_reset`。
- [ ] outbox 模式下业务回滚后事件**不推送**。
- [ ] 背压：慢消费者触发丢最旧 + `system_notice.dropped`。
- [ ] 大 payload 自动走 `snapshot_uri`，客户端能访问。
- [ ] 订阅不越权（契约测试：free 订阅 global 被拒）。

## 十一、已决事项（2026-04-24）

- **S1 · `briefing` 首登补发**：**不补发**；新登录用户只接收后续事件，历史简报在页面按需拉取。
- **S2 · `price_ticks` free 用户 symbol 上限**：**无上限**；配额仅按总连接数与 RPM 控制（见《用户系统与鉴权》§7.2）。
