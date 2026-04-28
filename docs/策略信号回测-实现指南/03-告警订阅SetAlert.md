# 03 · 告警订阅 SetAlert

> ARK `/setalert` 命令：用户按 symbol/contract 订阅 COT/信号告警，触发后通过 SSE / 通知发送。AntClaw 当前完全缺失。

---

## 1. 目标

1. 用户能创建/列出/修改/删除告警订阅；
2. 支持 4 类告警条件：
   - `cot_extreme`：COT index 越过阈值
   - `signal_flip`：unified_signal recommendation 翻转
   - `regime_change`：宏观状态转换 (regime_transitions 写入新事件)
   - `price_threshold`：价格突破/跌破指定阈值
3. 后端 worker 周期评估，触发时写入 `notifications`（已有表）+ 推送 SSE 频道 `user:{user_id}:alerts`。

---

## 2. 数据模型

### 2.1 新建表 `user_signal_alerts`

```sql
CREATE TABLE IF NOT EXISTS user_signal_alerts (
    id            BIGSERIAL PRIMARY KEY,
    user_id       VARCHAR(64) NOT NULL,
    alert_type    VARCHAR(32) NOT NULL,   -- cot_extreme/signal_flip/regime_change/price_threshold
    symbol        VARCHAR(32) NOT NULL,
    params        JSONB        NOT NULL,  -- 阈值/方向等
    enabled       BOOLEAN      NOT NULL DEFAULT TRUE,
    last_fired_at TIMESTAMPTZ,
    cooldown_seconds INT       NOT NULL DEFAULT 3600,  -- 防抖
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ
);

CREATE INDEX idx_user_alerts_user ON user_signal_alerts(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_user_alerts_active ON user_signal_alerts(enabled, alert_type) WHERE deleted_at IS NULL;
```

### 2.2 params JSONB 结构（按 alert_type）

```jsonc
// cot_extreme
{ "cot_index_min": 80, "cot_index_max": 20, "direction": "any" }
// signal_flip
{ "from_directions": ["LONG","STRONG_LONG"], "to_directions": ["SHORT","STRONG_SHORT"] }
// regime_change
{ "watch_severities": ["WARN","CRITICAL"] }
// price_threshold
{ "above": 1.1000, "below": null }
```

---

## 3. RPC 设计

修改 `proto/antclaw/v1/alerts.proto`（已存在 alerts.proto，扩展即可）：

```proto
service AlertsService {
  rpc CreateAlert(CreateAlertRequest)  returns (CreateAlertResponse);
  rpc ListAlerts(ListAlertsRequest)    returns (ListAlertsResponse);
  rpc UpdateAlert(UpdateAlertRequest)  returns (UpdateAlertResponse);
  rpc DeleteAlert(DeleteAlertRequest)  returns (DeleteAlertResponse);
  rpc ToggleAlert(ToggleAlertRequest)  returns (ToggleAlertResponse);
}

message AlertRule {
  int64  id = 1;
  string user_id = 2;
  string alert_type = 3;
  string symbol = 4;
  string params_json = 5;        // 序列化的 JSON
  bool   enabled = 6;
  int64  last_fired_at = 7;      // unix ts
  int32  cooldown_seconds = 8;
}

message CreateAlertRequest {
  string alert_type = 1;
  string symbol = 2;
  string params_json = 3;
  int32  cooldown_seconds = 4;
}
message CreateAlertResponse { AlertRule alert = 1; }

message ListAlertsRequest  { string alert_type = 1; }   // 空=全部
message ListAlertsResponse { repeated AlertRule alerts = 1; }

message UpdateAlertRequest  { int64 id = 1; string params_json = 2; int32 cooldown_seconds = 3; }
message UpdateAlertResponse { AlertRule alert = 1; }

message DeleteAlertRequest  { int64 id = 1; }
message DeleteAlertResponse { bool ok = 1; }

message ToggleAlertRequest  { int64 id = 1; bool enabled = 2; }
message ToggleAlertResponse { AlertRule alert = 1; }
```

`buf generate` 后实现 handler。

---

## 4. 用户配额

每用户最多 50 条 enabled 告警；超出返回 `ERROR_CODE_ALERT_QUOTA_EXCEEDED`。

```go
const MaxActiveAlertsPerUser = 50
```

---

## 5. 后端服务

文件：`backend/internal/service/alerts/service.go`

```go
type Service struct {
    pool *pgxpool.Pool
    log  *slog.Logger
}

func (s *Service) Create(ctx context.Context, userID string, in CreateAlertInput) (*AlertRule, error)
func (s *Service) List(ctx context.Context, userID, alertType string) ([]AlertRule, error)
func (s *Service) Update(ctx context.Context, userID string, id int64, in UpdateAlertInput) (*AlertRule, error)
func (s *Service) Delete(ctx context.Context, userID string, id int64) error
func (s *Service) Toggle(ctx context.Context, userID string, id int64, enabled bool) (*AlertRule, error)

// 内部
func (s *Service) ListActiveByType(ctx context.Context, alertType string) ([]AlertRule, error)
func (s *Service) MarkFired(ctx context.Context, id int64, at time.Time) error
```

参数验证：
- `alert_type` 必须在白名单
- `params_json` 按 alert_type schema 校验（用 jsonschema 库或手写验证函数）
- `cooldown_seconds` ∈ [60, 86400]

---

## 6. 评估器（Worker）

文件：`backend/cmd/antclaw-worker/alert_evaluator.go`

调度：每 5 分钟一次（`alert_evaluator` 任务）。

流程：
```
for each alert_type ∈ {cot_extreme, signal_flip, regime_change, price_threshold}:
    rules = AlertSvc.ListActiveByType(ctx, alert_type)
    for rule in rules:
        if rule.last_fired_at + rule.cooldown_seconds > now: continue
        if !evaluateRule(rule):                              continue

        # 写入 notification
        notif = {
            user_id: rule.user_id,
            channel: "alert",
            title: titleOf(rule),
            body:  bodyOf(rule),
            payload: {alert_id, alert_type, symbol, ...},
        }
        notifyRepo.Insert(notif)
        ssePublisher.Publish("user:"+rule.user_id+":alerts", notif)
        AlertSvc.MarkFired(rule.id, now)
```

### 6.1 evaluateRule（按 type）

```go
// cot_extreme
cot := COTProvider.GetLatestAnalysis(ResolveCOTCode(rule.Symbol))
if cot.COTIndex >= params.cot_index_min || cot.COTIndex <= params.cot_index_max:
    return true

// signal_flip
recent := SignalRepo.GetRecentBySymbol(rule.Symbol, 2)
if len(recent) < 2: return false
prev, curr := recent[1].Recommendation, recent[0].Recommendation
return contains(params.from_directions, prev) && contains(params.to_directions, curr)

// regime_change
trans := RegimeProvider.GetTransitions(rule.Symbol, rule.LastFiredAt)
for t in trans:
    if t.Severity in params.watch_severities: return true

// price_threshold
last := PriceProvider.GetLatestPrice(rule.Symbol)
if params.above != nil && last.Close > params.above:  return true
if params.below != nil && last.Close < params.below:  return true
```

---

## 7. SSE 推送

复用已有 `internal/adapter/sse`。频道命名：`user:{user_id}:alerts`。事件 schema：

```json
{
  "type": "alert.fired",
  "alert_id": 123,
  "alert_type": "cot_extreme",
  "symbol": "EURUSD",
  "fired_at": "2026-04-26T12:34:56Z",
  "title": "EURUSD COT 极端值",
  "body": "COT index 达到 87，触发您的订阅阈值"
}
```

新增事件 key 必须先注册到 `proto/antclaw/v1/stream.proto`（宪章 9）。

---

## 8. 修改清单

| 文件 | 动作 |
|------|------|
| `proto/antclaw/v1/alerts.proto` | 扩展（添加 RPC + AlertRule message） |
| `proto/antclaw/v1/common.proto` | 已在 00 文档加 ERROR_CODE_ALERT_QUOTA_EXCEEDED |
| `proto/antclaw/v1/stream.proto` | 注册 `alert.fired` 事件 |
| `backend/internal/adapter/storage/postgres/ensure_schema.go` | 创建 `user_signal_alerts` |
| `backend/internal/service/alerts/service.go` | 新建 |
| `backend/internal/service/alerts/types.go` | 新建 |
| `backend/internal/service/alerts/validators.go` | 新建（params schema 校验）|
| `backend/internal/service/alerts/service_test.go` | 新建 |
| `backend/internal/adapter/rpc/alerts_handler.go` | 新建（5 个 RPC）|
| `backend/cmd/antclaw-worker/alert_evaluator.go` | 新建 |
| `backend/cmd/antclaw-worker/main.go` | 注册 alert_evaluator 任务 |
| `backend/cmd/antclaw-api/main.go` | 注册 AlertsHandler |
| `frontend/admin/src/pages/Alerts.tsx` | 新建（用户告警管理 UI） |

---

## 9. 验证

```bash
# 1. 创建一条 cot_extreme 订阅（注意 JWT）
TOKEN=$(curl -s -d '{"username":"admin","password":"admin"}' \
  http://localhost:8082/antclaw.v1.AuthService/Login | jq -r .token)

curl -s http://localhost:8082/antclaw.v1.AlertsService/CreateAlert \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"alert_type":"cot_extreme","symbol":"EURUSD",
       "params_json":"{\"cot_index_min\":80,\"cot_index_max\":20}",
       "cooldown_seconds":3600}'

# 2. 列出
curl -s http://localhost:8082/antclaw.v1.AlertsService/ListAlerts \
  -H "Authorization: Bearer $TOKEN" -d '{}' -H 'Content-Type: application/json' | jq .

# 3. 触发评估器
docker compose exec -T worker /app/antclaw-worker --once=alert_evaluator

# 4. 检查 notifications 表
docker compose exec -T postgres psql -U antclaw -d antclaw -c \
  "SELECT id, user_id, channel, title, created_at FROM notifications
   WHERE channel='alert' ORDER BY id DESC LIMIT 5;"

# 5. SSE（当前 API：`/sse/jobs`、`/sse/audit`；按 user channel 的通用 `/sse?channels=` 为规划）
curl -N http://localhost:8082/sse/jobs
# 若经 admin nginx 反代：curl -N http://localhost:8081/sse/jobs
```

---

## 10. 完成判定

- [ ] proto 编译通过；
- [ ] 5 个 RPC 工作正常；
- [ ] 配额限制有效；
- [ ] worker 评估器周期运行，命中条件后写 notifications + SSE；
- [ ] 防抖（cooldown）生效，重复评估不二次触发；
- [ ] 软删除（deleted_at）后告警立即停止。

## 11. 实施记录

<!-- -->
