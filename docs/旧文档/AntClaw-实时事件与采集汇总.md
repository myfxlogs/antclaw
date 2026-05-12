# AntClaw 实时事件与采集数据汇总

本文档描述 Admin 后台中两类“真实数据 + 实时反馈”能力的实现方式：

- **任务实时状态**：Worker 中所有采集 / 分析任务的 `running / succeeded / failed` 事件，通过 Redis Streams + SSE 推送到 Admin 前端。
- **审计日志实时**：每条审计日志在落库的同时写入事件流，前端 Audit 页面无需轮询即可实时追加。
- **采集数据汇总**：管理端可一键查看 14 个采集对象在数据库中的真实落库情况（行数、最新数据时间）。

文档遵循 `代码简洁、健壮、可靠、稳健、实用` 的总要求；接口实现尽量短小、对故障容忍。

---

## 1. 总体架构

```
+----------------+      XADD       +-----------------+        XREAD(BLOCK)        +-----+   SSE
|  antclaw-worker| --------------> |  Redis Streams  | -------------------------> | API | -------> Admin (浏览器)
|  + AuditService|                 |  jobs_events    |                            |     |
+----------------+                 |  audit_events   |                            +-----+
                                   +-----------------+

+----------------+
|  antclaw-api   |  Connect: AdminDataService/GetDataSummary → PostgreSQL → JSON 汇总
+----------------+
```

- 事件主通道：`gRPC unary（ConnectRPC）` 仍用于初始查询和管理动作；
- 实时通道：`Redis Streams + SSE`，**前端零轮询**；
- 数据汇总：**Connect** `AdminDataService`（Unary），直连 PostgreSQL 的聚合逻辑在 `data_corpus.go`。

---

## 2. 实时事件：Worker 端

文件：`backend/cmd/antclaw-worker/main.go`

### 2.1 事件结构

```go
type jobEvent struct {
    JobID      string `json:"job_id"`
    Name       string `json:"name"`
    Status     string `json:"status"`              // running / succeeded / failed
    StartedAt  int64  `json:"started_at,omitempty"`
    FinishedAt int64  `json:"finished_at,omitempty"`
    Error      string `json:"error,omitempty"`
}
```

### 2.2 统一包装器：`runWithEvent`

所有 15 个采集 / 分析任务都通过统一的包装函数执行，避免在每个 collector 内部插桩：

```go
runWithEvent(ctx, logger, "calendar-sync", "财经日历采集",
    func() error { return runCalendarCollection(ctx, calendarSvc, logger) })
```

包装器行为：

- 任务开始：发布 `running` 事件；
- 正常返回：发布 `succeeded` 事件；
- 业务错误（`fn() error`）：发布 `failed` 事件，附 `error` 详情；
- 防御性 `recover`：将运行时 `panic` 兜底为 `failed`。

错误传播以 **error 返回** 为主通道，`panic` 仅保护进程不挂；不再使用 panic 作为控制流。

### 2.3 流通道与快照

- 事件流：`stream:jobs_events`（Redis Stream）
- 快照键：`jobs:status:<job_id>`（Redis 普通 key，存最近一次完整事件 JSON）

### 2.4 已接入任务（15 个）

| 来源 | 任务 | job_id |
| --- | --- | --- |
| 采集 | 财经日历 | `calendar-sync` |
| 采集 | 宏观数据 | `macro-sync` |
| 采集 | 实际值更新 | `actuals-update` |
| 采集 | COT 持仓 | `cot-sync` |
| 采集 | 价格数据 | `price-sync` |
| 采集 | 情绪数据 | `sentiment-sync` |
| 采集 | 链上数据 | `onchain-sync` |
| 采集 | 分时价格 | `intraday-sync` |
| 采集 | DeFi 数据 | `defi-sync` |
| 采集 | VIX 期限结构 | `vix-term-sync` |
| 采集 | DVOL | `dvol-sync` |
| 分析 | COT 分析 | `cot-analysis` |
| 分析 | 宏观状态分类 | `macro-regime` |
| 分析 | 资金流向背离 | `flow-divergence` |
| 分析 | 成交量分布 | `volume-profile` |

---

## 3. 实时事件：审计日志

文件：`backend/internal/service/audit/audit.go`

`AuditService.Log` 在写入 Postgres `audit_logs` 表之后，附加一次 `XADD stream:audit_events`：

```go
_ = s.redis.Raw().XAdd(ctx, &redisv9.XAddArgs{
    Stream: "stream:audit_events",
    Values: map[string]interface{}{
        "action":   entry.Action,
        "resource": entry.Resource,
        "data":     string(payloadJSON),
    },
}).Err()
```

任意一条登录、封禁、解封、密码重置等动作都会立刻出现在 Admin Audit 页面。Redis 写入失败仅记录日志，不阻塞主路径。

---

## 4. API 层：SSE 与汇总接口

文件：

- `backend/cmd/antclaw-api/main.go`（注册 SSE 与汇总路由）
- `backend/internal/adapter/rpc/data_summary_handler.go`（汇总 handler）
- `backend/internal/adapter/storage/postgres/ensure_schema.go`（启动时确保 `audit_logs` 表存在）

### 4.1 SSE：任务事件

- 路径：`GET /sse/jobs`
- 实现：从 `stream:jobs_events` 起 `XRead(BLOCK 0)`，新事件即时通过 `text/event-stream` 推到客户端。
- ID 推进：使用 `lastID = msg.ID`，断线重连时浏览器自带 EventSource 会基于 `id:` 字段恢复（当前实现起点为 `$`，新事件优先；如需历史回放，可改为从 `0` 起）。

### 4.2 SSE：审计事件

- 路径：`GET /sse/audit`
- 行为同上，订阅 `stream:audit_events`。

### 4.3 采集数据汇总

- **Connect**：`antclaw.v1.AdminDataService/GetDataSummary`（原 `GET /admin/data/summary` 已移除）
- 输出（JSON 形状与旧接口一致，便于对照）：

```json
{
  "items": [
    {
      "job_id": "price-sync",
      "name": "价格数据(日K)",
      "table": "price_daily",
      "count": 3585,
      "latest_time": 1777178956
    }
  ],
  "updated_at": 1777200000
}
```

- 健壮性：单表查询失败仅在 `error` 字段标记，**不会影响整体响应**；
- 14 个采集对象的表名与时间列已校对真实 schema；
- 不依赖 `proto`、不动 `sqlc`，启动即可用。

### 4.4 启动自检

API 启动时执行幂等 `CREATE TABLE IF NOT EXISTS audit_logs`、`CREATE INDEX IF NOT EXISTS ...`，确保审计相关查询不会因为表缺失报错。

---

## 5. 前端

### 5.1 Jobs 实时更新

文件：`frontend/admin/src/pages/Jobs.tsx`

```ts
const evtSource = new EventSource('/sse/jobs')
evtSource.onmessage = (event) => {
  const data = JSON.parse(event.data)
  setJobs((prev) => {
    const idx = prev.findIndex(j => j.job_id === data.job_id)
    if (idx === -1) return prev
    const updated = [...prev]
    updated[idx] = {
      ...updated[idx],
      status: data.status || updated[idx].status,
      last_run: data.finished_at
        ? new Date(data.finished_at * 1000).toISOString()
        : updated[idx].last_run,
    }
    return updated
  })
}
```

- 进入页面：`listJobs()` 拉初始快照；
- 之后：完全依赖 SSE，无 `setInterval`、无轮询；
- 卸载或断连：`evtSource.close()`。

### 5.2 Audit 实时新增

文件：`frontend/admin/src/pages/Audit.tsx`

收到事件即 prepend 到列表：

```ts
setEntries((prev) => [
  {
    log_id: `sse-${Date.now()}`,
    user_id: data.user_id || '',
    action: data.action,
    resource: data.resource,
    details: data.details,
    created_at: data.timestamp,
    ip_address: data.ip_address || '',
  },
  ...prev,
])
```

### 5.3 采集数据页

文件：`frontend/admin/src/pages/DataSummary.tsx` + `components/Layout.tsx` + `App.tsx`

- 路由：`/data`
- 调用：`lib/api.ts` 的 `getDataSummary()` / `getDataPreview()`（Connect `AdminDataService`）
- 展示：每个采集对象一行，列出对应表名、行数、最新时间、状态。

### 5.4 nginx 反向代理

文件：`deploy/nginx.admin.conf`

- `location /sse/`：关闭 `proxy_buffering`、`chunked_transfer_encoding off`、长 timeout，确保 SSE 实时推送；
- `location ~ ^/antclaw\.`：Connect-RPC 反代至 API（采集汇总/预览亦经此路径，**不再**使用 `/admin/data/`）。

---

## 6. 后续可演进

- **gRPC server streaming 升级**：在 `admin.proto` 中加入 `WatchJobs` / `WatchAuditLogs`，用 ConnectRPC 替换 SSE。事件源（Redis Streams）和事件 schema 不变，前端切换 `client.watchJobs()` 即可。
- **collector 错误细分**：12 个内部仅打印日志的 collector 后续可改为返回 error，使外层事件状态比"已执行完毕"更精确。
- **Key/地址配置**：在新表 `app_settings` 中存储敏感配置（值字段加密），由 Admin 设置页 `GET/PUT /admin/settings` 读写，并在 Worker 中支持热加载。

---

## 7. 验证方法

```bash
# 1. 查看 Redis Streams 中累积的任务事件
docker exec antclaw-redis redis-cli XLEN stream:jobs_events
docker exec antclaw-redis redis-cli XREVRANGE stream:jobs_events + - COUNT 5

# 2. 监听 SSE（直连 API）
curl -N http://localhost:8082/sse/jobs

# 3. 通过 Admin 反向代理监听 SSE
curl -N http://localhost:8081/sse/jobs

# 4. 查看采集汇总（需先取 TOKEN，Connect Unary）
curl -s -X POST http://localhost:8082/antclaw.v1.AdminDataService/GetDataSummary \
  -H "Authorization: Bearer $T" -H "Content-Type: application/json" -d '{}' | python3 -m json.tool

# 5. 浏览器
http://localhost:8081/jobs    # 任务实时
http://localhost:8081/audit   # 审计实时
http://localhost:8081/data    # 采集数据
```
