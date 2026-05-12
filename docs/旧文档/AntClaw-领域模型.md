# AntClaw · 领域模型

> 本文是各子域的**聚合根、值对象、领域事件、不变式**的单一真相。代码实现阶段的 `internal/domain/*`、`proto/*` 字段、`db/migrations/*` 均以本文为依据。
>
> 与 `AntClaw-重构解决方案.md` 的配合方式：方案文档定"做什么"，领域模型文档定"怎么建模"，任务分解文档定"怎么交付"。

## 一、建模原则

1. **六边形架构**：领域层 (`internal/domain`) **纯净**，不依赖任何基础设施包（不 import `pgx / go-redis / connect`）。
2. **充血模型优先**：业务规则写在聚合根方法，而非 service 的过程式代码。
3. **不变式就地保护**：任何违反不变式的方法返回错误，不允许事后"修正数据"。
4. **值对象不可变**：构造后字段私有，仅暴露 getter；变更返回新实例。
5. **时间全 UTC**：领域内的 `time.Time` 一律 UTC；进入边界前由 adapter 转换。
6. **金额全 `Money{Amount decimal.Decimal, Currency string}`**：禁止裸 `float64`。
7. **ID 类型化**：每个聚合根有自己的 ID 类型（`UserID`、`StrategyID`…），避免"any ID"混用。

---

## 二、通用值对象

### `Money`

```
Money {
  Amount   decimal.Decimal   // numeric(20,10)
  Currency string            // ISO 4217，如 "USD"、"EUR"；"XBT" 视为扩展
}

不变式：
- Amount 不能为 NaN / Inf
- Currency 长度 ∈ [3, 6]，全大写
```

### `Instrument`

```
Instrument {
  Symbol string     // "EURUSD"、"XAUUSD"、"BTCUSD"
  Venue  string     // "OANDA"、"BINANCE"、"FRED"、"CBOE"
  Kind   InstrumentKind  // FX | CRYPTO | EQUITY | COMMODITY | INDEX | MACRO
}

不变式：
- (Symbol, Venue) 唯一
- Kind 与 Venue 可配性由白名单校验
```

### `TimeRange`

```
TimeRange { Start time.Time; End time.Time }
不变式：
- Start <= End
- 两端均为 UTC
```

### `Locale`

```
Locale string  // BCP-47，如 "zh-CN"、"en-US"、"zh-TW"

不变式：
- 必须在平台支持列表内（首发：zh-CN、en-US）
- 未知 locale 由协商层兼底为 audience 默认 locale
```

### `Timezone`

```
Timezone string  // IANA tz，如 "Asia/Shanghai"、"UTC"
不变式：时区字符串必须能被 time.LoadLocation 解析
```

### `Percent`

```
Percent decimal.Decimal   // 存小数，0.1234 表示 12.34%
不变式：| value | <= 10000（兜底防异常）
```

---

## 三、子域列表（一览）

| 子域 | 聚合根 | 简述 |
|------|--------|------|
| 身份 | `User` | 账户、角色、locale、timezone |
| 会话 | `Session` | 刷新 token、设备指纹 |
| 鉴权 | （无聚合根） | 策略服务，跨聚合 |
| 通知 | `Notification`、`PushToken` | 站内信、推送 |
| 订阅 | `Subscription` | 用户对告警/行情的偏好 |
| COT | `CotReport` | 周度 COT 数据 |
| 宏观 | `MacroSeries`、`NewsImpact` | FRED/OECD/BIS + 新闻事件冲击 |
| 日历 | `EconomicEvent` | 经济日历 |
| 行情 | `PriceBar`、`PriceTick` | 多频次行情 |
| 波动 | `VolSurface` | IV/HV |
| 信号 | `Signal` | 策略产出的建仓信号 |
| 回测 | `BacktestRun`（后期执行） | 运行实例、指标 |
| 策略 | `UserStrategy` | 用户上传/生成的策略 |
| 情绪 | `SentimentSnapshot` | 社媒/机构情绪 |
| AI | （无聚合根） | 叙事服务；用量 `AiUsage` 作为事件流 |
| 告警 | `Alert` | 规则 + 触发 |
| 审计 | `AuditLog` | append-only |
| i18n | `I18nString` | `(key, locale)` |
| Bot | `BotBinding` | 预留 |
| 反馈 | `Feedback` | 用户反馈 |
| 任务 | `Task` | 异步任务，含 backtest |

---

## 四、身份子域

### 聚合根 `User`

```
User {
  ID                UserID
  Email             string         // UNIQUE, lowercased
  Username          string         // UNIQUE, default = email
  PasswordHash      []byte         // Argon2id
  PasswordAlgo      string         // "argon2id:<m=...,t=...,p=...>"
  Roles             []Role         // { "free" | "premium" | "admin" }
  Status            UserStatus     // ACTIVE | LOCKED | DELETED
  Locale            Locale
  Timezone          Timezone
  LastSeenTimezone  Timezone       // 浏览器最后一次上报
  TotpSecret        []byte?        // 预留（加密存）
  TotpEnabled       bool
  CreatedAt         time.Time
  UpdatedAt         time.Time
  Version           int64          // optimistic lock
}
```

不变式：

1. `Email` 唯一且小写；
2. `Roles` 非空；`admin` 不能与 `free` 共存；
3. 密码 hash 非空；
4. `Status=DELETED` 后不再可登录。

方法：

- `Register(cmd RegisterCmd) (*User, error)` 新建，免激活直接 `ACTIVE`；
- `ChangePassword(old, new string) error` 校验+重设；
- `Promote(role Role)` / `Demote(role Role)`；
- `Lock(reason string)` / `Unlock()`；
- `UpdateProfile(locale Locale, tz Timezone)`；
- `RevokeAllSessions()` 仅增 `Version`。

领域事件：

- `UserRegistered`、`UserLocked`、`UserPromoted`、`UserPasswordChanged`、`UserAllSessionsRevoked`。

### 聚合根 `Session`

```
Session {
  ID                 SessionID
  UserID             UserID
  RefreshTokenHash   []byte
  UserAgent          string
  IP                 netip.Addr
  ExpiresAt          time.Time
  RevokedAt          time.Time?
}
```

不变式：

- `ExpiresAt > CreatedAt`；
- `RevokedAt != nil` 的 session 不可刷新。

方法：`Refresh()`、`Revoke()`。

---

## 五、通知子域

### 聚合根 `Notification`

```
Notification {
  ID         NotificationID
  UserID     UserID
  Category   NotificationCategory  // ALERT | SIGNAL | SYSTEM | AI | FEEDBACK_REPLY
  TitleKey   string                // i18n key
  BodyKey    string
  Payload    map[string]any        // 语言无关数据，前端/模板渲染
  ReadAt     time.Time?
  CreatedAt  time.Time
}
```

不变式：

- `TitleKey` 与 `BodyKey` 必须已在 `i18n_strings` 注册；
- `Payload` 只含原始值、不含已本地化字符串。

方法：`MarkRead()`、`Snooze(until time.Time)`（可选扩展）。

领域事件：`NotificationCreated`、`NotificationRead`。

### 聚合根 `PushToken`

```
PushToken {
  ID         PushTokenID
  UserID     UserID
  Platform   PushPlatform // FCM | APNS | HMS
  Token      string
  UpdatedAt  time.Time
}
```

不变式：同一 `(UserID, Platform, Token)` 唯一。

---

## 六、订阅子域

### 聚合根 `Subscription`

```
Subscription {
  ID        SubscriptionID
  UserID    UserID
  Channel   string           // e.g. "alert.cot.release" / "price_tick:EURUSD"
  Filter    SubscriptionFilter  // jsonb：impact>=2, symbols=[…]
  CreatedAt time.Time
}
```

不变式：`Channel` 必须匹配 `stream.proto` 中定义的枚举字面量。

---

## 七、行情与时间序列

### 值对象 `PriceBar`

```
PriceBar {
  Instrument Instrument
  Interval   Interval   // M1 | M5 | M15 | H1 | H4 | D1 | W1
  Ts         time.Time
  Open, High, Low, Close Money
  Volume     decimal.Decimal
}
```

不变式：

- `Low <= min(Open, Close, High)` 且 `High >= max(Open, Close, Low)`；
- `Volume >= 0`；
- `Ts` 对齐 Interval 边界（例如 M15 必须能被 15 分钟整除）。

### 值对象 `PriceTick`

```
PriceTick {
  Instrument Instrument
  Ts         time.Time     // 纳秒级
  Bid, Ask   Money
}
```

不变式：`Bid.Currency == Ask.Currency`，`Bid.Amount <= Ask.Amount`。

### 聚合 `VolSurface`

```
VolSurface {
  Instrument Instrument
  Ts         time.Time
  Tenors     []Tenor         // 1W, 1M, 3M, ...
  Strikes    []decimal.Decimal
  Matrix     [][]decimal.Decimal  // [tenor][strike] -> iv
}
```

不变式：矩阵维度匹配；IV 非负；按 `(Instrument, Ts)` 唯一。

---

## 八、COT / 宏观 / 新闻

### 聚合根 `CotReport`

```
CotReport {
  Instrument    Instrument       // 归一到交易工具（EURUSD ↔ 6E）
  ReportDate    civil.Date       // 周五
  Noncomm, Comm, NonReportable CotPositions
  OpenInterest  int64
  Source        string
}

CotPositions { Long int64; Short int64 }
```

不变式：

- `ReportDate` 必须为周二（CFTC 标称日）；
- `Long/Short >= 0`。

### 聚合根 `MacroSeries`

```
MacroSeries {
  ID       MacroSeriesID
  Source   string         // "FRED" | "OECD" | "BIS" | "ECB"
  Code     string         // "DGS10"、"CPIAUCSL"
  Title    string
  Freq     Frequency      // D | W | M | Q | A
  Unit     string
  Title_i18n map[Locale]string
}
```

### 值对象 `MacroObservation`

```
MacroObservation {
  SeriesID MacroSeriesID
  Ts       time.Time
  Value    decimal.Decimal?   // 允许 NULL（缺失值）
}
```

不变式：`(SeriesID, Ts)` 唯一。

### 聚合根 `EconomicEvent`

```
EconomicEvent {
  ID         EventID
  Ts         time.Time
  Country    string        // ISO 3166-1 alpha-2
  Title_i18n map[Locale]string
  Impact     ImpactLevel   // LOW | MEDIUM | HIGH
  Actual     decimal.Decimal?
  Forecast   decimal.Decimal?
  Previous   decimal.Decimal?
  Source     string
}
```

不变式：`Impact` 必在枚举内；`Actual` 公布前为 NULL。

### 聚合根 `NewsImpact`

```
NewsImpact {
  EventID EventID
  Instrument Instrument
  SurpriseZ  decimal.Decimal   // 意外 z 分数
  Moves      map[Horizon]decimal.Decimal   // 1min, 5min, 1h 回报
}
```

---

## 九、信号与回测

### 聚合根 `Signal`

```
Signal {
  ID           SignalID
  UserID       UserID?       // nil 表示平台公共信号
  StrategyID   StrategyID?   // 来自用户策略时非空
  Instrument   Instrument
  Ts           time.Time
  Direction    Direction     // LONG | SHORT | FLAT
  Strength     decimal.Decimal  // [0,1]
  Reason       map[string]any  // 语言无关
  Expires      time.Time?
}
```

不变式：

- `Strength ∈ [0, 1]`；
- `Direction = FLAT` 时 `Strength = 0`。

### 聚合根 `UserStrategy`

```
UserStrategy {
  ID             StrategyID
  UserID         UserID
  Name           string
  SourceKind     StrategySourceKind   // MQL4 | MQL5 | DSL | NATURAL
  SourceBlobURI  string               // MinIO 对象
  DSLAst         *DSLAst              // 后期执行器唯一依据；MVP 可为 nil
  DSLSource      string
  Status         StrategyStatus       // DRAFT | READY | INVALID
  Error          *StrategyError
  CreatedAt      time.Time
  UpdatedAt      time.Time
}
```

不变式：

- `Status=READY` 要求 `DSLAst != nil`；
- `Status=INVALID` 要求 `Error != nil`。

> MVP 范围内只允许 `Status=DRAFT`；转译与执行属后期 LP-B。

### 聚合根 `BacktestRun`（**MVP 仅数据结构，不执行**）

```
BacktestRun {
  ID          RunID
  UserID      UserID
  StrategyID  StrategyID
  Params      map[string]any
  ArtifactURI string          // 工件在 MinIO 的根路径
  Metrics     BacktestMetrics
  Status      RunStatus       // QUEUED | RUNNING | SUCCEEDED | FAILED | CANCELED | RESOURCE_LIMIT
  CreatedAt   time.Time
  FinishedAt  time.Time?
}

BacktestMetrics {
  TotalReturn  decimal.Decimal
  MaxDrawdown  decimal.Decimal
  Sharpe       decimal.Decimal
  Sortino      decimal.Decimal
  Trades       int
  WinRate      decimal.Decimal
}
```

附属时间序列：`BacktestEquity(RunID, Ts, Equity Money)`（hypertable）。

---

## 十、AI 子域

### 值对象 `AiUsage`

```
AiUsage {
  UserID   UserID
  Provider string        // "gemini" | "claude"
  Model    string
  PromptTokens     int64
  CompletionTokens int64
  CostCents        decimal.Decimal
  Ts               time.Time
}
```

- 只记录一行即视为一次调用；写入走 hypertable。
- 不记录提示词内容（保护隐私）。

### 领域服务 `AiNarrativeService`（无聚合根）

方法签名：

- `Chat(userID UserID, prompt string, context ChatContext) (Stream, error)`
- `Interpret(userID UserID, signal Signal) (Narrative, error)`
- `Outlook(userID UserID, horizon Horizon) (Narrative, error)`
- `TranslateStrategy(userID UserID, src StrategySource) (DSLAst, []Warning, error)` **MVP 返回 `Unimplemented`**

规则：

- 所有方法必以调用方 `userID` 的 BYOK 密钥执行；
- 缓存 key 含 `(userID, prompt_hash, locale, model)`；
- 失败时 `Narrative.Degraded = true` 且提供模板回退。

---

## 十一、告警子域

### 聚合根 `Alert`

```
Alert {
  ID         AlertID
  UserID     UserID?     // nil = 系统级
  Kind       AlertKind   // COT_RELEASE | MACRO | SIGNAL | CUSTOM
  Condition  AlertCondition   // jsonb
  State      AlertState       // ARMED | TRIGGERED | SILENCED
  Cooldown   time.Duration
  LastFired  time.Time?
  CreatedAt  time.Time
}
```

不变式：

- `Cooldown > 0`；
- `TRIGGERED` 状态下不能再次触发，直到 `Cooldown` 满或手动 `Disarm`。

领域事件：`AlertArmed`、`AlertTriggered`、`AlertSilenced`、`AlertDisarmed`。

---

## 十二、审计

### 聚合根 `AuditLog`

```
AuditLog {
  ID        AuditID
  ActorID   UserID?        // nil = system
  Action    string         // "auth.login.success" / "admin.user.promote"
  Target    string         // "user:<id>" / "strategy:<id>"
  Meta      map[string]any
  PrevHash  []byte
  Hash      []byte          // sha256(prev_hash || canonical_json(row))
  At        time.Time
}
```

不变式（**SQL 层 + WORM 兜底**）：

- 插入后不可 `UPDATE` / `DELETE` / `TRUNCATE`；
- `Hash = SHA-256(PrevHash || canonical(Action, Target, ActorID, Meta, At))`；
- `PrevHash` = 前一条的 `Hash`；第一条 `PrevHash = 32*\x00`；
- 同步双写 MinIO `audit-worm` bucket（object lock = COMPLIANCE）。

---

## 十三、Bot 子域（预留）

### 聚合根 `BotBinding`

```
BotBinding {
  UserID      UserID
  Platform    BotPlatform  // TELEGRAM | WECHAT | FEISHU | SLACK | DISCORD | ...
  ExternalID  string
  BoundAt     time.Time
}
```

唯一约束：`(Platform, ExternalID)`。

### 端口接口（Go 视角）

```go
type BotAdapter interface {
    Platform() BotPlatform
    Send(ctx context.Context, external BotID, msg BotOutbound) error
    Subscribe(ctx context.Context, handler BotInboundHandler) error
}
```

本期不提供任何实现。

---

## 十四、i18n / 反馈 / 任务

### 值对象 `I18nString`

```
I18nString { Key string; Locale Locale; Text string; UpdatedAt time.Time }
```

`(Key, Locale)` 为主键。

### 聚合根 `Feedback`

```
Feedback { ID FeedbackID; UserID UserID; Content string; At time.Time }
```

### 聚合根 `Task`

```
Task {
  ID         TaskID
  UserID     UserID?
  Kind       string       // "backtest" | "ai_chat" | "ingest_cot" | ...
  Status     TaskStatus   // QUEUED | RUNNING | SUCCEEDED | FAILED | CANCELED
  Params     map[string]any
  ResultURI  string       // MinIO；大型结果存 URI
  Error      *TaskError
  CreatedAt  time.Time
  UpdatedAt  time.Time
}
```

---

## 十五、领域事件清单

| 事件 | 发生时机 | 订阅者 |
|------|----------|--------|
| `UserRegistered` | 注册成功 | `notify`（欢迎站内信 + 邮件）、`audit` |
| `UserPasswordChanged` | 改密成功 | `audit`、`notify` |
| `UserAllSessionsRevoked` | 管理员重置 | `session gc`、`audit` |
| `NotificationCreated` | 站内信写入 | `sse`、`push` |
| `AlertTriggered` | 规则命中 | `notify`、`webhook`（admin） |
| `BacktestSubmitted` | 后期 | `backtest-runner` |
| `BacktestCompleted` | 后期 | `notify`、前端轮询 |
| `StrategyTranslated` | 后期 | `notify`、前端刷新 |
| `AuditLogged` | 任意写操作 | `worm writer` |

事件总线统一走 **Redis Streams**，键名见附录 B.2。

---

## 十六、跨子域约束

1. 任何发自用户操作的数据写入 **必须**产生 `AuditLogged` 事件；
2. `Notification.TitleKey` / `BodyKey` 必须在 `I18nString` 存在 `en-US` 与 `zh-CN` 版本；
3. `Signal` 与 `Alert` 只能引用 `Instrument`、不得裸引用 `Symbol` 字符串；
4. BYOK 解密结果在领域内以 `[]byte` + `memguard` 传递，绝不持久化、绝不跨聚合传播；
5. `AuditLog.Meta` 不得包含密码、token、API key、完整的 AI 提示词。

---

## 十七、领域层目录映射

```
backend/internal/domain/
  identity/         # User, Session
  notify/           # Notification, PushToken
  subscription/
  market/           # PriceBar, PriceTick, VolSurface
  cot/
  macro/
  calendar/
  news/
  signal/
  strategy/         # UserStrategy, DSLAst
  backtest/         # BacktestRun, BacktestMetrics, BacktestEquity
  ai/
  alert/
  audit/
  i18n/
  bot/
  feedback/
  task/
  shared/           # Money, Instrument, TimeRange, Locale, Timezone, Percent
```

每个子域目录下：

- `<aggregate>.go`：聚合根 + 值对象；
- `events.go`：领域事件；
- `errors.go`：领域错误（`ErrEmailTaken`、`ErrInvariantBroken`...）；
- `service.go`：跨聚合的领域服务（若需要）；
- `<aggregate>_test.go`：聚合根不变式测试，覆盖率 ≥ 85%。

---

## 十八、与 Proto 的映射规则

- 领域类型 ↔ Proto 消息：字段名 snake_case，枚举带类型前缀（`DIRECTION_LONG` 不是 `LONG`）；
- `Money` → `Money { string amount = 1; string currency = 2; }`；传输用字符串表达 `decimal`，保留精度；
- `time.Time` → `google.protobuf.Timestamp`；
- `decimal.Decimal` → `string`（按 `numeric(20,10)` 渲染）；
- `map[Locale]string` → `map<string, string> title_i18n = N;`；
- Adapter 层 `internal/adapter/rpc/` 做领域 ↔ Proto 互转，领域层不依赖 proto。

---

## 十九、未列入本版本的主题

- Pine Script / cTrader 等脚本接入：后期 LP-B 扩展。
- 组合账户 / 多策略合并：后期。
- 真实下单流水：等 LP-C MT 等价层落地后另立子域 `trading`。

---

> **备注**：本文的字段与不变式是 `.proto` 与 `db/migrations/` 的**规范上游**。任何字段增删请回到本文修改后再下发实现。
