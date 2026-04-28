# AntClaw 项目重构解决方案（v2）

> 本文档基于 `docs/ARK-Intelligent-功能清单.md` 的功能边界，给出将项目全面重写为 **AntClaw** 的落地方案。相较 v1，本版采用 **Connect（Connect-RPC）+ Protobuf** 作为统一 API 协议，**gRPC** 对服务端/移动端原生暴露，**SSE** 作为浏览器单工推送通道；**PostgreSQL + Redis** 作为永久化与缓存层；新增**用户系统**、**管理员前端系统**与**完整 i18n 体系**；前后端目录分离，业务由**自有 Web 与移动 App** 承载。旧 Telegram Bot 本期**不迁移**，仅在架构中**预留可插拔 Bot 接入层接口**（未来可统一接入 Telegram / WeChat / Feishu 等）。

---

## 一、总体目标

1. **全面重写（硬约束）**：AntClaw 为**全新工程**，与 `ark-intelligent` 无代码继承关系。ARK-Intelligent 仅作为**功能参照项目（reference-only）**，用于对齐功能清单与业务语义；**严禁**搬运、粘贴、逐行翻译其任何 Go / TS / SQL / YAML / shell 代码。AntClaw 采用自有代码风格、命名规范、模块划分；仓库、模块路径、镜像、配置、日志、命令全部以 `antclaw` 命名；代码中不得出现 `ark-intelligent` / `ARK Intelligent` 字样。
2. **主通道**：
   - **对外 API**：基于 Connect-RPC + Protobuf；同一份 `.proto` 同时生成 **Connect**（Web/移动）、**gRPC**（服务端/移动原生）与服务端桩。
   - **业务前端**：自有 Web 与移动 App，作为用户主入口，取代旧 Telegram Bot 交互。
   - **Bot 接入层（预留）**：定义统一的 Bot 端口接口（`BotAdapter`）与消息模型，本期不实现任何具体 Bot；未来可平滑接入 Telegram / WeChat / Feishu 等，复用同一套账号、订阅、告警与权限体系。
3. **协议分工**：
   - **Connect RPC**（HTTP/1.1 + HTTP/2 + gRPC-Web 兼容）：Web/移动统一调用入口。
   - **gRPC**：服务端到服务端、移动端原生 SDK 可选。
   - **SSE（单工推送）**：浏览器端订阅告警/行情/任务进度，规避 WebSocket。
4. **功能零回退**：覆盖原清单全部 11 大类能力。
5. **用户体系**：邮箱 + 密码，用户名默认等于邮箱，密码使用 **Argon2id** 存储；JWT（访问 + 刷新）会话；RBAC。
6. **管理员前端**：独立 Admin Web 控制台，覆盖用户/权限/配额/数据源/任务/告警审计。
7. **存储**：**PostgreSQL**（持久化）+ **Redis**（缓存、限流、任务队列、SSE 扇出、Pub/Sub）。
8. **客户端**：Web + Admin（React + Vite + shadcn/ui + Tailwind）；移动 App分两条线：**React Native + Expo**（首发）、**ArkTS / HarmonyOS**（后期）；移动端整体排在后期。
9. **国际化**：全链路多语言；**用户端兼底 `en-US`，管理端兼底 `zh-CN`**。
10. **持久化**：PostgreSQL 16 + **TimescaleDB 扩展**（时间序列 hypertable + 压缩 + continuous aggregates）；首发**单实例**，读写分离仅在硬化阶段按需开启；Redis 7 单实例 + AOF/RDB 持久化；**MinIO** 本地对象存储，后期切 **Cloudflare R2**；审计日志同步双写至 MinIO **WORM / Object Lock** bucket 兜底。
11. **鉴权**：邮箱+密码（Argon2id），**免激活**注册即用；邮箱仅用于注册通知与密码找回；JWT 签名 **EdDSA (Ed25519)**。
12. **通知**：终态主通道为站内信 + 移动推送；**过渡期**（移动端未交付）由 **Web SSE + 站内信** 承担日常告警，不允许回落邮件；Webhook 仅内部/管理员。
13. **AI BYOK（严格隔离）**：用户自配 Gemini/Claude 密钥，仅自用、不跨用户共享；平台不提供共享密钥回退。
14. **运维栈**：Caddy 自动 TLS；OTel + Jaeger/Tempo + Prometheus + Grafana + **Sentry SaaS**，全默认开启。
15. **保留期**：业务/历史/回测/审计 **永久**（手动清理）；审计 **append-only**；调试/性能日志 30 天。
16. **工程准则**：所有计算在后端；前端密码输入框**明文显示**；「先分析，再实现」；任何脚本（含生成物）超 800 行必拆，Markdown 除外。
17. **一键部署**：单机单套 Docker Compose 为终态；MVP 服务 `caddy / postgres / redis / minio / antclaw-api / antclaw-worker / antclaw-web / antclaw-admin / prometheus / grafana / jaeger`；`postgres-replica`、`antclaw-backtest-runner`、`antclaw-mt-gateway`（mt4/mt5 等价层）均为硬化/后期可选 profile。
18. **实时性 SLA**：`alert`/`signal` 端到端 ≤ 500ms（P95）；`price_tick` ≤ 300ms；`task_progress` ≤ 1s；取消与恢复使用 `Last-Event-ID` 游标重连；SSE 满队则丢最旧、保证最新。
19. **任务隔离（预留）**：回测/策略执行本期 **不投产**。已预留契约：独立 `antclaw-backtest-runner` 服务位、Redis Streams 消费组、`tasks` 表、结果 URI 字段；后期启用时配 `cpu_quota`/`mem_limit`、容器硬化与 `context.Cancel`。
20. **流事件契约**：所有实时事件（Web SSE、移动 gRPC server-streaming）**共用同一份 `stream.proto`**，代码生成一次，传输层差异仅在编解码封装。
21. **策略执行沙箱（后期实现，本期仅预留结构）**：MVP 仅落实：`user_strategies` 表、`strategy.proto` + `mt4.proto` + `mt5.proto` 文件骨架、`StrategyService` / `MT4Service` / `MT5Service` RPC 契约返回 `UNIMPLEMENTED`、前端页面隐藏或显示「后期提供」。后期再实施：Starlark 嵌入引擎、MQL→DSL AI 转译、fuel/壁钟/内存硬限、`antclaw-backtest-runner` 容器硬化、租户隔离；同期提供 `antclaw-mt-gateway`（基于 `mt4.proto`/`mt5.proto`）以替代 MetaTrader 客户端，**不再采用 Wine+MT terminal 路线**。

---

## 二、命名与目录规范

### 2.1 命名基线（AntClaw 全新定义）

> 本表为 AntClaw **立项同时确定的命名基线**，并非从旧仓库映射转换。右列为 **唯一合法值**；任何出现带 `ark` / `ARK` 的标识符、文件名、环境变量、镜像名、日志字段皆为违规，CI 零容忍。

| 维度 | AntClaw 基线值 |
|----|----|
| 品牌小写/路径 | `antclaw` |
| 品牌大写/展示 | `AntClaw` |
| Go module | `github.com/antclaw/antclaw` |
| Docker 镜像名 | `antclaw/<component>:<tag>` |
| 容器/服务名 | `antclaw-*`（如 `antclaw-api`、`antclaw-worker`）|
| 配置环境变量前缀 | `ANTCLAW_*` |
| 数据目录 | `./data/antclaw` |
| 日志字段 | `service=antclaw` |

### 2.2 关键决策一览（已拍板）

| # | 主题 | 决策 |
|---|------|------|
| 1 | 移动框架 | RN+Expo 首发；ArkTS HarmonyOS 后期；移动端排后期 |
| 2 | JWT | EdDSA (Ed25519) |
| 3 | 网关 | Caddy（自动 TLS） |
| 4 | 任务队列 | Redis Streams（不引入 asynq） |
| 5 | 可观测 | OTel + Jaeger/Tempo + Prometheus + Grafana，全默认 |
| 6 | 错误上报 | Sentry SaaS（前+后端） |
| 7 | 邮件 | 仅注册通知与密码找回 |
| 8 | 对象存储 | MinIO→后期 Cloudflare R2 |
| 9 | 分层 | free / premium / admin |
| 10 | 注册 | 完全开放 |
| 11 | 邮箱验证 | 免激活 |
| 12 | 2FA | 延后；管理端预留开关 |
| 13 | API Key | 内部使用；AI 能力 BYOK，用户间严格隔离，无平台共享 |
| 14 | Webhook | 内部/admin；用户用站内信 |
| 15 | 计费 | 延后 |
| 16 | 保留 | 业务永久，手动清理；日志 30 天 |
| 17 | BadgerDB 迁移 | 执行一次性导入 |
| 18 | 多语言业务字段 | 启用 `title_i18n JSONB`；默认语言=设备系统语言 |
| 19 | 时区默认 | 按浏览器/设备 |
| 20 | 部署 | 单机 Compose 为终态 |
| 21 | PostgreSQL | 首发单实例 + TimescaleDB；读写分离仅硬化期可选 |
| 22 | Redis | 单实例+Docker+持久化 |
| 23 | 环境 | 单套 compose |
| 24 | 域名 | 同域不同路径 |
| 25 | 设计系统 | shadcn/ui + Tailwind |
| 26 | 移动分发 | App Store + Google Play |
| 27 | 合规 | 无 GDPR/出境/备案 |
| 28 | 审计 | append-only 永久；调试/性能 30 天；事件同步写入 MinIO WORM bucket 兜底 |
| 29 | 时间序列 | TimescaleDB hypertable + 压缩；price/signal/fred 等按 `(symbol, ts)` 分块 |
| 30 | 实时性 SLA | alert/signal ≤ 500ms；price_tick ≤ 300ms；task_progress ≤ 1s |
| 31 | 流事件 schema | 全客户端共享 `stream.proto`，只编译一次 |
| 32 | 回测运行 | 预留服务位 + 契约；MVP 不投产，后期由 `antclaw-backtest-runner` 承担 |
| 33 | AI 用量 | `ai_usage` 表，前端展示 token 开销；不入性能日志 |
| 34 | 策略沙箱 | **后期实现**；MVP 仅预留表、proto、RPC（返回 `UNIMPLEMENTED`） |
| 35 | 沙箱资源 | 后期参数：fuel 1e8、回测 30s / 优化 60s、内存 256MB；MVP 不生效 |
| 36 | MT 等价层 | 后期 `mt4.proto` / `mt5.proto` + `antclaw-mt-gateway` 替代 MetaTrader；不走 Wine+MT |

**强制工程准则**：

- 所有业务计算在**后端**；前端仅渲染与交互，不含策略/汇总/排序/过滤业务逻辑。
- 前端所有密码输入框**明文显示**（不使用 `type=password` 黑点遮罩）；仅传输层 TLS + 服务端 Argon2id 存储。
- 「先分析，再实现」；拒绝「快速可用」式摆烂；任何脚本（含生成物）超 800 行必须拆分，Markdown 除外。

### 2.3 前后端分离的仓库布局（单仓 monorepo）

```
antclaw/
├── backend/                              # Go 后端
│   ├── cmd/
│   │   ├── antclaw-api/                  # Connect + gRPC + SSE 主进程
│   │   └── antclaw-worker/               # 调度 / 数据管线 / 异步任务
│   │   # 注：antclaw-bot 进程本期不交付，仅保留端口接口（见 ports/bot.go）
│   ├── internal/
│   │   ├── app/                          # 装配 / DI
│   │   ├── config/                       # ANTCLAW_* 配置
│   │   ├── domain/                       # 领域模型
│   │   ├── ports/                        # 端口接口
│   │   ├── service/                      # 业务服务实现（按聚合拆分，全新编写）
│   │   ├── scheduler/                    # 后台任务
│   │   ├── auth/                         # 用户/鉴权/Argon2id/Ed25519 JWT
│   │   ├── notify/                       # 站内信 + 移动推送(FCM/APNs/HMS) + 邮件
│   │   ├── byok/                         # 用户自有 AI 密钥管理 (AES-GCM)
│   │   ├── i18n/                         # 消息目录、locale 解析、格式化
│   │   └── adapter/
│   │       ├── rpc/                      # Connect + gRPC Handler
│   │       ├── sse/                      # SSE 推送
│   │       ├── bot/                      # Bot 接入层（预留：接口 + 空实现）
│   │       ├── sandbox/                  # 策略沙箱引擎（预留：MVP 仅空实现，后期 Starlark）
│   │       │   ├── starlark/             # Starlark 解释器封装（后期启用）
│   │       │   ├── builtin/                # 白名单 built-in 函数（bar/ind/buy/sell/close_all/position/log/param）
│   │       │   └── validator/              # AST 白名单校验器（后期启用）
│   │       ├── storage/
│   │       │   ├── postgres/             # pgx + sqlc；主库写 + 只读副本池
│   │       │   ├── redis/                # 缓存/限流/Streams(任务队列+扇出)
│   │       │   └── objectstore/          # MinIO S3；未来可切 R2
│   │       └── webhook/                  # 出站 Webhook（仅内部/管理员）
│   │
│   ├── cmd/antclaw-backtest-runner/      # 回测独立运行进程（隔离）
│   │   ├── main.go                       # runner 入口（MVP 仅打印启动日志后退出）
│   │   └── runner/                       # 沙箱执行器（后期启用）
│   │       ├── engine.go                 # Starlark 引擎初始化
│   │       ├── fuel.go                   # 指令燃料计数与超时控制
│   │       ├── builtin.go                # 白名单 built-in 注册
│   │       └── task.go                   # Redis Streams 任务消费与结果上报
│   ├── pkg/                              # 复用基础设施
│   └── db/
│       ├── migrations/                   # goose 迁移
│       └── queries/                      # sqlc SQL
├── proto/                                # API 契约单一真相
│   └── antclaw/v1/
│       ├── common.proto                  # 通用类型（Locale、Money、TimeRange…）
│       ├── auth.proto
│       ├── user.proto
│       ├── admin.proto
│       ├── strategy.proto             # 策略与回测（MVP RPC 返 UNIMPLEMENTED）
│       ├── mt4.proto                  # MT4 等价层（后期）
│       ├── mt5.proto                  # MT5 等价层（后期）
│       ├── cot.proto
│       ├── calendar.proto
│       ├── macro.proto
│       ├── price.proto
│       ├── vol.proto
│       ├── signals.proto
│       ├── backtest.proto
│       ├── ta.proto
│       ├── sentiment.proto
│       ├── ai.proto
│       ├── alerts.proto
│       └── stream.proto                  # SSE 事件 schema
├── gen/                                  # buf 生成产物（Go/TS/Dart/Kotlin/Swift）
├── frontend/
│   ├── web/                              # 用户 Web（React+Vite+TS+shadcn/ui+Tailwind）
│   ├── admin/                            # 管理员 Web（同栈）
│   ├── mobile-rn/                        # 移动首发：React Native + Expo
│   ├── mobile-arkts/                     # HarmonyOS ArkTS（后期）
│   └── packages/
│       ├── ui/                           # shadcn/ui 组件与 Design Tokens
│       └── i18n/                         # 共享翻译资源（zh-CN、en-US…）
├── deploy/
│   ├── docker-compose.yaml
│   ├── Dockerfile.backend
│   ├── Dockerfile.web
│   └── Dockerfile.admin
├── scripts/
├── buf.yaml · buf.gen.yaml               # buf 契约与代码生成
└── docs/                                 # 中文文档
```

---

## 三、API 协议设计（Connect + gRPC + SSE）

### 3.1 选型原因

- **Connect-RPC**：单一 `.proto` 同时支持 Connect（HTTP/1.1 JSON/二进制）、gRPC、gRPC-Web，浏览器可直连，无需边车；TypeScript/Swift/Kotlin/Dart SDK 生成完善。
- **gRPC**：为服务端到服务端与移动原生提供高性能二进制通道。
- **SSE**：浏览器天然支持，单向推送（告警、进度、行情），避免 WebSocket 在代理/防火墙下的兼容问题；移动端若需双工可直接走 gRPC server-streaming。
- **禁用 WebSocket**：按要求移除。

### 3.2 统一规约

- **契约源**：`proto/antclaw/v1/*.proto`，由 `buf` 管理 lint/breaking/生成。
- **包名**：`antclaw.v1`；服务命名 `XxxService`；方法动词化。
- **错误模型**：Connect `code` + `google.rpc.Status` 明细；业务错误走 `ErrorDetail`。
- **鉴权**：Header `Authorization: Bearer <JWT>`；服务端到服务端用长期 API Key（`X-AntClaw-Key`）。
- **语言**：`Accept-Language` 头 + 用户资料 `locale` 字段协商，响应头回 `Content-Language`；错误码与可本地化文本键 `message_key` 同时返回。
- **限流**：Redis 令牌桶，按 `user_id` / `api_key` / `ip` 维度。
- **幂等**：写方法接受 `idempotency_key` 字段。
- **追踪**：`traceparent` 透传；OTel 导出。
- **SDK 产物**：`gen/go`、`gen/ts`、`gen/dart`、`gen/kotlin`、`gen/swift`。

### 3.3 服务清单

> 下表「对应命令」列仅作为能力对照（源自旧 Bot 清单），便于未来 Bot 接入层映射；本期不实现任何 Bot。

| Proto Service | 方法（节选） | 对应命令（历史对照） |
|---------------|--------------|--------|
| `AuthService` | `Register` `Login` `Refresh` `Logout` `RequestPasswordReset` `ResetPassword` `VerifyEmail` | — |
| `UserService` | `GetMe` `UpdateSettings` `GetMembership` `StartOnboarding` `GetHistory` `ClearHistory` `ListPins` `Pin` `Unpin` `SubmitFeedback` | `/start /settings /membership /onboarding /history /clear /pin /unpin /pins /feedback` |
| `AdminService` | `ListUsers` `SetRole` `Ban` `Unban` `RunJob` `ListJobs` `ListAuditLogs` `ListWebhookDeliveries` | `/users /setrole /ban /unban` |
| `CotService` | `GetSummary` `Compare` `GetSignals` `GetHistory` `SubscribePairAlert` | `/cot /ce /setalert` |
| `CalendarService` | `ListEvents` `GetEvent` `GetImpact` `GetImpactHistory` | `/calendar /cal /impact` |
| `MacroService` | `GetFred` `GetEcb` `GetSnb` `GetOecdLeading` `GetEurostat` `GetBis` `GetTradingEconomics` `GetDtccSwaps` `GetSec13F` `GetTreasuryAuctions` `GetFedWatch` `GetWorldBank` `GetImfWeo` | `/macro /ecb /snb /leading /eurostat /bis /cbrates /tedge /globalm /swaps /13f /treasury` |
| `PriceService` | `GetPrice` `GetLevels` `GetMarketOverview` `GetSession` `RunScenario` `GetRegime` `GetSeasonal` | `/price /levels /market /session /scenario /regime /seasonal` |
| `VolService` | `GetVix` `GetMove` `GetDvol` `GetGex` `GetIvol` `GetSkew` `GetSkewVixAlert` | `/vix /gex /ivol /skew` |
| `SignalService` | `GetBias` `GetRank` `GetXFactors` `GetRadar` `GetIntensity` `GetTransition` `GetCryptoAlpha` `GetUnified` `GetQuant` `GetCta` `GetBriefing` `GetOutlook` | `/bias /rank /xfactors /radar /intensity /transition /cryptoalpha /signal /quant /cta /briefing /outlook` |
| `BacktestService` | `RunBacktest` `GetBacktest` `GetAccuracy` `RunQuantBt` `RunVpBt` `RunCtaBt` | `/backtest /bt /bta /accuracy /quantbt /vpbt /ctabt` |
| `TaService` | `GetIndicators` `GetElliott` `GetWyckoff` `GetIct` `GetAmt` `GetOrderflow` `GetVolumeProfile` `GetIntermarket` | `/elliott /wyckoff /ict /auction /vp /orderflow /flows /intermarket` |
| `SentimentService` | `GetSentiment` `GetOnchain` `GetDefiHealth` `GetCarryMonitor` | `/sentiment /onchain /defi /carry` |
| `AiService` | `Chat`（server-stream）`Interpret` `Outlook` | 聊天模式 + AI 叙事 |
| `AlertService` | `ListSubscriptions` `Subscribe` `Unsubscribe` `RegisterWebhook` `ListWebhooks` | `/setalert` + 广播 |
| `StreamService` | SSE：`GET /sse/v1/stream?channels=...&token=...` | 告警/简报/进度推送 |

### 3.4 SSE 通道

- 端点：`GET /sse/v1/stream?channels=cot,alerts,briefing,skew_vix,carry,regime,tasks&token=<JWT>`
- 事件帧：`event: <channel>` + `data: <protojson payload>` + `id: <monotonic>`；客户端用 `Last-Event-ID` 断点续传。
- 扇出：Redis `PUBLISH` / Streams → SSE 网关 → 订阅用户。
- 任务进度：`POST BacktestService.RunBacktest` 返回 `task_id`，客户端订阅 `tasks:<id>` 频道接收阶段事件与最终结果。

---

## 四、用户系统

### 4.1 注册与登录

- **注册字段**：`email`、`password`、可选 `display_name`。
- **用户名**：默认 `username = email`；可后续修改为任意唯一串。
- **密码哈希**：**Argon2id**，参数 `memory=64MB, iterations=3, parallelism=2, saltLen=16, keyLen=32`（可由 `ANTCLAW_ARGON2_*` 覆盖）；使用 `golang.org/x/crypto/argon2`。
- **密码策略**：长度 ≥ 10，至少包含两类字符；zxcvbn 评分 ≥ 2。
- **邮箱验证**：注册后发送一次性令牌邮件，激活前仅能访问受限范围。
- **找回密码**：邮件发送一次性短时 token（15 分钟）。
- **登录**：返回 `access_token`（15 分钟）+ `refresh_token`（30 天，存 Redis，可吊销）。
- **JWT**：HS256 或 EdDSA；载荷 `sub, email, roles, ver`；`ver` 与用户记录版本绑定，便于全局失效。
- **会话**：每设备一条 `sessions` 记录，记录 UA/IP/最近活跃。
- **二要素（MVP 可选）**：TOTP（RFC 6238），管理员强制开启。
- **风控**：登录失败 Redis 计数 + 阶梯退避；异常地登录邮件提醒。

### 4.2 角色与权限（RBAC）

| 角色 | 说明 |
|------|------|
| `admin` | 全部管理权限 |
| `premium` | 全部业务 API + 全告警 |
| `free` | 业务 API 受限 + 仅 USD/高影响告警 |

配额与冷却（AI RPM、告警冷却、每用户配额）作为 Redis 中间件统一实现。

### 4.3 注册与邮箱策略

- **完全开放注册**；**免邮箱激活**，注册后立即可用。
- 邮箱用途：注册成功通知、密码找回、系统级重要通知；**不承担日常告警**（告警走站内信+移动推送）。
- **密码明文输入**：前端登录/注册/修改密码表单一律明文显示，不使用黑点遮罩；仅依赖 TLS 传输与 Argon2id 存储保障安全。

### 4.4 2FA（本期延后）

- `users` 预留 `totp_secret` + `totp_enabled`；
- 管理端预留全局开关「强制普通用户启用 2FA」，本期不投产；
- 管理员账户本期**不做** 2FA。

### 4.5 AI BYOK（Bring Your Own Key，严格隔离 + 用量）

- **用户只用自己的密钥**：在设置中配置自有 Gemini/Claude API Key；服务端以 `AES-GCM` 加密后存 `user_ai_keys`，主密钥 `ANTCLAW_BYOK_MASTER_KEY`（支持轮换）。
- **不提供平台共享密钥**、**不允许跨用户共享**；未配置或密钥失败时，AI 调用直接返回 `key_missing` / `key_invalid` 错误码，前端引导至设置页填入。绝无平台公用回退。
- 运行时强制校验：`user_ai_keys` 按 `user_id + provider` 堆出；调用时按当前用户上下文读取，严禁用户 A 的请求命中用户 B 的密钥。
- Admin 可配置：每用户每日调用配额、密钥轮换周期提示、异常行为监控；**不提供**「开启平台共享密钥」开关。
- **用量与费用可见**：每次调用写入 `ai_usage`（含 `prompt_tokens`、`completion_tokens`、估算 `cost_cents`）；前端设置页展示本日/本月总耗。密钥健康：`worker` 每日静默调用 `models.list` 探针对私有 key 验证，失败仅通知本人、不落持久日志。
- **主密钥版本**：`ANTCLAW_BYOK_MASTER_KEY` 以 `v<n>:<base64>` 前缀存放；密文头记录版本，应支持多版本解密与轮换重写。

---

## 五、管理员前端系统

独立 SPA `frontend/admin/`，仅 `admin` 角色可访问。模块：

- **用户管理**：搜索、分页、改角色、封禁/解封、强制登出、重置密码。
- **权限与配额**：角色模板、AI 每用户调用额度、告警冷却、分层白名单；**BYOK 严格隔离**，不包含平台共享密钥开关。
- **2FA 开关**（预留）：全局是否强制 2FA；本期不启用。
- **通知中心**：站内信模板、移动推送密钥、邮件 SMTP 配置与模板。
- **任务与调度**：查看/手动触发后台任务，查看最近运行。
- **数据源健康**：各外部源探针状态、错误率、降级标记。
- **审计日志**：登录/写操作/管理操作全量审计；表为 append-only，不提供删除 UI。
- **数据清理**：按类型对业务/历史/回测提供手动清理入口（非定时）。
- **对象存储浏览**：MinIO / R2 bucket 查看，回测导出、截图、附件预览。
- **i18n 资源**：`i18n_strings` 编辑、导入/导出 `json`、缺失键报表、完成度看板。
- **反馈工单**：查看用户反馈。
- **系统指标/追踪**：内嵌 Grafana；链路跳转 Jaeger/Tempo。

技术栈：React 18 + Vite + TypeScript + TanStack Router/Query + shadcn/ui + Tailwind + Connect-Query 客户端。

---

## 六、前端（用户 Web 与移动 App）

### 6.1 用户 Web `frontend/web/`

- React 18 + Vite + TypeScript + TanStack Router/Query + Tailwind + shadcn/ui + ECharts/Recharts。
- 模块：Dashboard、COT、Calendar、Macro、Price、Vol、Signals、Backtest、TA、Sentiment、AI Chat、Alerts、Settings。
- 与后端通信：`@connectrpc/connect-web`；推送通过 `EventSource` 订阅 SSE。
- 认证：JWT 存 HttpOnly Cookie（CSRF double-submit）。

### 6.2 移动 App（后期交付）

- **首发**：`frontend/mobile-rn/` — React Native + Expo；分发计划 App Store + Google Play。
- **HarmonyOS 版**：`frontend/mobile-arkts/` — ArkTS；后期立项。
- 通信：`@connectrpc/connect` 客户端；推送通道用 gRPC server-streaming 替代 SSE（更省电）。
- 原生推送：FCM / APNs / HarmonyOS Push；**站内信 + 推送为用户通知主通道**。
- **排期**：移动端整体在 Web 与后端稳定后开工。

### 6.3 设计系统

- **shadcn/ui + Tailwind** 终选；Design Tokens（颜色/间距/字号/阴影）集中在 `frontend/packages/ui`，Web/Admin 共享，移动端映射同 token。

### 6.4 前后端职责红线

- **前端零业务计算**：仅调用 RPC、渲染响应、维护 UI 状态；任何业务指标/策略/过滤/排序/聚合由后端 RPC 返回结果或提供专用方法。
- **密码明文**：所有密码输入框不使用黑点遮罩，统一配「显示/隐藏」以外的纯明文展示。

---

## 七、持久化与缓存（PostgreSQL + Redis）

### 7.1 PostgreSQL + TimescaleDB（首发单实例）

- **拓扑**：`postgres` 单实例，镜像 `timescale/timescaledb:2.x-pg16`；安装后 `CREATE EXTENSION IF NOT EXISTS timescaledb, pgcrypto, pg_trgm`。
- **读写分离**（硬化期可选）：提供 `postgres-replica` 流复制 profile，应用代码预留 `db.WithReader/WithWriter` 双连接池抽象（首发下两者指向同一实例），开启副本时只改 DSN。
- 驱动：`pgx/v5` + `sqlc`；迁移：`goose`。
- **时间序列 hypertable**：`price_tick`、`price_minute`、`price_daily`、`signals`、`fred_series`、`cot_records`、`news_impact`、`backtest_equity`、`ai_usage` 等按 `SELECT create_hypertable('<t>', 'ts', chunk_time_interval => INTERVAL '7 days')` 建块；主键 `(symbol, ts)` 或 `(user_id, ts)`；历史块启动 `compression`（> 30 天）；`continuous_aggregate` 预计算日/小时 OHLC。
- **金融精度**：价格/金额字段统一用 `numeric(20,10)`，禁止 `float/double`；Money 值对象 = `{amount:numeric, currency:text}`。
- 关键表：
  - `users(id, email UNIQUE, username UNIQUE, password_hash, password_algo, roles, status, locale, timezone, totp_secret NULL, totp_enabled BOOL, created_at, updated_at, version)`
  - `sessions(id, user_id, refresh_token_hash, ua, ip, expires_at, revoked_at)`
  - `bot_bindings(user_id, platform, external_id, bound_at, UNIQUE(platform, external_id))` 预留
  - `api_keys(id, user_id, name, hash, scopes, last_used_at, revoked_at)` **仅内部服务**
  - `user_ai_keys(user_id, provider, ciphertext, nonce, updated_at, PRIMARY KEY(user_id, provider))` BYOK
  - `subscriptions(id, user_id, channel, filter_jsonb, created_at)`
  - `notifications(id, user_id, category, title_key, body_key, payload_jsonb, read_at, created_at)` 站内信
  - `push_tokens(id, user_id, platform, token, updated_at)` FCM/APNs/HMS
  - `webhooks(id, owner_type, owner_id, url, secret, active, created_at)` / `webhook_deliveries(...)` 仅内部/admin
  - `audit_logs(id, actor_id, action, target, meta_jsonb, prev_hash, hash, at)` **append-only**；建表后收紧 DELETE/UPDATE 权限，行级 hash 形成链式可校验
  - `feedback(id, user_id, content, at)`
  - `tasks(id, user_id, kind, status, params_jsonb, result_jsonb_uri, error, created_at, updated_at)`；重型结果仅存 `result_jsonb_uri`（MinIO）
  - `backtest_runs(id, user_id, strategy_id, params_jsonb, artifact_uri, metrics_jsonb, status, created_at)`、`backtest_equity(run_id, ts, equity numeric(20,10))` hypertable
  - `ai_usage(user_id, provider, model, prompt_tokens, completion_tokens, cost_cents numeric(20,10), ts)` hypertable
  - `user_strategies(id, user_id, name, source_kind ENUM('mql4','mql5','dsl','natural'), source_blob_uri, dsl_ast_jsonb, dsl_source TEXT, status ENUM('draft','ready','invalid'), error_jsonb NULL, created_at, updated_at)`：MQL/DSL/AI 策略统一落库。`source_blob_uri` 指向 MinIO 中的原始源码；`dsl_ast_jsonb` 为执行时唯一依据。
  - `i18n_strings(key, locale, text, updated_at, PRIMARY KEY(key, locale))`；业务 `title_i18n JSONB` 建 `GIN` 索引 + `CHECK (title_i18n ? 'en-US')` 确保默认 locale 必存
  - 业务数据表（COT、Price、FRED、新闻、影响、信号、回测结果等）均为 hypertable；`users` 添加 `last_seen_timezone` 字段，浏览器离线时用此渲染告警。
- **审计 append-only 硬约束**：
  1. 权限：应用角色 `REVOKE UPDATE, DELETE, TRUNCATE ON audit_logs`；
  2. 触发器：`BEFORE UPDATE OR DELETE OR TRUNCATE ON audit_logs FOR EACH STATEMENT EXECUTE FUNCTION audit_immutable()`，函数 `RAISE EXCEPTION`；
  3. 内容链：`INSERT` 触发器自动计算 `hash = sha256(prev_hash || canonical_json(row))`；
  4. **双写 WORM**：`audit` 服务同步写入 MinIO `audit-worm` bucket，开启 `object lock = COMPLIANCE`。
- 保留：业务/历史/回测/审计 **永久**；仅管理员手动清理；调试/性能日志不入 PG。
- 备份：每日 `pg_dump` + 归档 WAL 到 MinIO；副本启用后改为流复制 + 夜备份。

### 7.2 Redis（单实例 + 持久化）

- 部署：Docker 单实例；`appendonly yes`、`appendfsync everysec` + RDB 快照；卷 `redisdata`。
- 用途：缓存、限流（令牌桶 Lua）、登录风控、刷新 token 吊销、**Streams 任务队列**（`XADD/XREADGROUP/XACK`，不引入 asynq）、SSE 扇出、站内信实时推送。
- 驱动：`redis/go-redis/v9`。

### 7.3 对象存储（MinIO → Cloudflare R2）

- 本期：`minio` 容器（S3 兼容）存放回测导出、截图、文件附件、备份。
- 后期：业务提升后切 **Cloudflare R2**（仅端点与密钥更换，代码不变）；预留 `ANTCLAW_S3_ENDPOINT / ACCESS_KEY / SECRET_KEY / BUCKET / REGION`。

### 7.4 从 BadgerDB 迁移

- `backend/cmd/antclaw-migrate`：**首次部署必执行**，按领域将旧 Badger 数据落库 PG；幂等设计，重跑安全。

---

## 八、通知与站内信

- **终态主通道**：站内信（PG `notifications`）+ 移动推送（FCM/APNs/HMS）。
- **过渡期主通道**（移动端未交付）：Web SSE + 站内信；**不允许**用邮件作为日常告警替代通道。
- **交付流**：调度事件 → Redis Streams → `notify` 服务 →
  1. 写入 `notifications`（按用户 `locale` 本地化）；
  2. 若用户有 `push_tokens` 则推送原生通知；
  3. Web 在线同时经 SSE `notify.new` 频道收到微提示。
- **邮件**：仅出现在系统级事件（注册通知、密码找回、账户安全操作）；模板文件 `templates/mail/<locale>/*.html`。
- **Webhook**：仅 admin/内部使用；用户面不开放注册。
- **Bot**：未来作为可选订阅者，入口路由在 `adapter/bot/`。

---

## 九、后端架构（保留六边形）

```
       ┌──────────────────────────────────────────────────────────┐
 入站  │  Connect/HTTP(S)  ·  gRPC  ·  SSE   ·  Bot 接入层（预留）  │
       └──────────▲────────────▲──────────▲─────────────▲─────────┘
                  │ports       │          │             │
       ┌──────────┴────────────┴──────────┴─────────────┴─────────┐
 应用  │  应用服务：鉴权、配额、任务编排、缓存、事件总线            │
       └──────────▲────────────▲──────────▲─────────────▲─────────┘
                  │            │          │             │
       ┌──────────┴────────────┴──────────┴─────────────┴─────────┐
 领域  │  领域模型：信号/COT/宏观/价格/新闻/用户/订阅/…             │
       └──────────▲────────────▲──────────▲─────────────▲─────────┘
                  │ports       │          │             │
       ┌──────────┴────────────┴──────────┴─────────────┴─────────┐
 出站  │  PostgreSQL · Redis · 外部数据源 · AI(Gemini/Claude)      │
       └──────────────────────────────────────────────────────────┘
```

- 单进程 `antclaw-api` 同端口暴露 Connect + gRPC（`h2c` + Connect Handler）与 SSE。
- 事件总线：Redis Streams + 消费者组；调度器产出事件 → `notify` 服务 → 站内信 + 移动推送 + SSE；Webhook/Bot 作为可选订阅者。

---

## 十、Bot 接入层（预留，本期不实现）

旧 Telegram Bot 的迁移成本高，本期**仅设计接口、不提供实现**，用自有 Web/移动 App 承载业务；未来可统一接入 Telegram、WeChat、Feishu、Slack、Discord 等。

### 9.1 端口接口（Go）

```go
// backend/internal/ports/bot.go
package ports

type BotPlatform string // "telegram" | "wechat" | "feishu" | "slack" | ...

type BotMessage struct {
    Platform   BotPlatform
    ExternalID string            // 平台原生消息 ID
    ChatID     string            // 会话/群 ID
    UserExtID  string            // 平台侧用户 ID
    UserID     string            // 已绑定的 AntClaw 用户 ID，可空
    Locale     string            // 平台提供的语言
    Text       string
    Command    string            // 规范化命令名，如 "cot"、"calendar"
    Args       []string
    Meta       map[string]string
}

type BotOutbound struct {
    ChatID   string
    Text     string            // 已本地化文本
    Keyboard any               // 平台差异化 payload
    Silent   bool
}

type BotAdapter interface {
    Platform() BotPlatform
    Start(ctx context.Context, in chan<- BotMessage) error
    Send(ctx context.Context, msg BotOutbound) error
    Close() error
}
```

### 9.2 账号绑定模型（预留表结构）

`bot_bindings(user_id, platform, external_id, bound_at, UNIQUE(platform, external_id))`，取代原先的 `telegram_bindings` 单表；任何未来 Bot 复用同一张表。

### 9.3 命令路由

- 所有 Bot 命令统一规范化为 `Command + Args`，经 `BotRouter` 调用**内部 gRPC**（与 Web/移动 App 等价通道），保证业务单一真相。
- 输出回流经 `BotAdapter.Send`，文本由 i18n 模块按用户 `locale` 渲染。

### 9.4 告警分发

事件总线订阅者新增可选 `BotDispatcher`，未来启用具体 Bot 时即可接收调度器事件，无需改动业务层。

---

## 十A、策略执行沙箱（本期仅预留结构，后期实现）

> 本章描述的是**后期目标态**。MVP 阶段仅交付：
> 
> - `user_strategies` 表（只存源码，不执行）；
> - `proto/antclaw/v1/strategy.proto`、`mt4.proto`、`mt5.proto` 骨架 + 全 RPC 返回 `UNIMPLEMENTED`；
> - 前端相关页面隐藏或展示「后期提供」提示；
> - **不启动** `antclaw-backtest-runner` / `antclaw-mt-gateway`。
> 
> 以下内容为后期实现的规范依据，不在 MVP 验收范围内。

### 10A.1 目的

用户以三种来源提交策略：粘贴 MQL4/MQL5 EA、自然语言 + AI 生成、手写 DSL。三者**统一归一为 `AntClaw-Strategy-DSL` AST**，再由平台引擎执行。平台**永不直接执行用户提交的 MQL 或任何图灵完备源码**。

### 10A.2 转译流水线

```
源来源
  ├─ MQL4/MQL5 文本──┐
  ├─ 自然语言 运─┼───►  AIService.TranslateStrategy → DSL 文本
  └─ 手写 DSL  ────┘                        │
                                                    ▼
                                        Starlark 解析器 → AST
                                                    │
                                                    ▼
                                          AST 白名单校验
                                                    │
                                                    ▼
                                        user_strategies.dsl_ast_jsonb
```

规则：

- MQL 转译由 AI 完成，结果**必经用户前端确认**（展示 DSL + 源对照 + 可能缺失的 API）；驳回或调整后才 `status=ready`。
- 任何不受支持的 MQL 特性（文件 IO、DLL、外网 http、线程、全局模式轮询）→ `status=invalid` + `error_jsonb.reason`，前端显示标红。
- 转译使用的是用户自有 BYOK AI 密钥，计入 `ai_usage`。

### 10A.3 AntClaw-Strategy-DSL（引擎 = Starlark 子集）

- **基底**：`go.starlark.net/starlark`；Starlark 本身不含 `import/io/os/threading`，绝对安全。
- **白名单 built-in**（唯一引入数据与动作的通道）：
  - `bar(i) -> {open, high, low, close, volume, ts}`（仅可访问当前步及之前，禁未来）
  - `ind(name, *args)` 线安：`EMA/SMA/RSI/ATR/MACD/BB` 等固定集
  - `buy(lot, sl?, tp?)`、`sell(lot, sl?, tp?)`、`close_all()`、`position() -> {...}`
  - `log(msg)`（仅写入本次运行的 `backtest_logs`，不入全局日志）
  - `param(name, default)`（读取用户传入的整数/浮点/布尔）
- **禁止**：全局导入、`getattr/setattr` 循环访存、自定义递归深度 > 32、`load()` 模块、任何 IO。
- **硬限**（初始值，可在 Admin 调）：

| 项 | 值 |
|----|----|
| `thread.SetMaxExecutionSteps` 指令 fuel | 1e8 |
| 单次回测 wall-clock | 30s |
| 单次优化 wall-clock | 60s |
| 解释器堆内存 | 256 MiB |
| 调用栈深度 | 256 |
| 单脚本大小 | 128 KiB |

超限都以 `BACKTEST_RESOURCE_LIMIT` 返回，结果标记 `aborted=true`。

### 10A.4 执行环境硬化（容器层）

`antclaw-backtest-runner` Compose 配置：

```yaml
antclaw-backtest-runner:
  image: antclaw/backtest-runner:<ver>
  read_only: true
  tmpfs: ["/tmp"]
  cap_drop: ["ALL"]
  security_opt: ["no-new-privileges:true", "seccomp=deploy/seccomp-strict.json"]
  networks: [internal]          # 仅访问 antclaw-api、redis、postgres；无外网
  deploy:
    resources:
      limits:  { cpus: "1.0", memory: 512M }
      reservations: { cpus: "0.25", memory: 128M }
  depends_on: [ redis, postgres ]
```

- `seccomp-strict.json` 基于 Docker default 再关闭 `unshare/mount/ptrace` 等。
- 运行进程不持有 JWT 私钥与其他用户数据。

### 10A.5 租户隔离

- runner 按 `user_id` 从 Redis Streams 消费任务；执行前仍通过 gRPC 回调 `antclaw-api` 拉取 **仅属于该用户的回测数据快照**（返回为执行器内的只读切片）。
- runner 进程生命周期内 **仅持有当前任务用户** 的 BYOK 解密结果（若需要 AI 调用），任务结束立即 `memguard.Destroy()`。
- 禁止 runner 直接操作 PG；所有数据读写走 `api` 的受限 RPC，由 `api` 做租户校验。

### 10A.6 观测指标

`backtest_runs_total{status}`、`backtest_duration_seconds`、`backtest_fuel_used`、`backtest_cancels_total`、`dsl_translate_total{result}`。

### 10A.7 MT4/MT5 等价层（后期）

代替 MetaTrader 客户端的路线不再采用 Wine + MT terminal，而是由平台自提供 `mt4.proto` / `mt5.proto`（行情订阅、下单、仓位、历史等方法），落地在 `antclaw-mt-gateway` 服务，上游对接实际经纪商 API。

- MVP：仅提交 `mt4.proto` / `mt5.proto` 骨架，不启动 `antclaw-mt-gateway` 容器；所有相关 RPC 返回 `UNIMPLEMENTED`。
- 后期：实现 `antclaw-mt-gateway`，按同样的容器硬化标准 (第 4 层) 运行；和 `antclaw-backtest-runner` 共享租户隔离原则。

### 10A.8 其他后期扩展

- Pine Script、cTrader cAlgo等同样走「子集转译 → DSL」模型，仍由 `antclaw-backtest-runner` 解释执行。

---

## 十一、安全

- **密码**：Argon2id；参数可配；校验恒定时间；**前端明文展示**仅 UI 约定，不影响后端策略。
- **JWT**：EdDSA（Ed25519）非对称；私钥仅 `antclaw-api` 持有，公钥可内网共享；短 TTL + 刷新；载荷 `aud∈{app,admin}`、`locale`、`ver`；`ver` 吊销全局 token。
- **BYOK 加密**：`AES-GCM` + `ANTCLAW_BYOK_MASTER_KEY`（`v<n>:<base64>`多版本）；密文头携版本标识；支持轮换时由 `worker` 后台分批重写。密文仅在调用瞬间解密，日志不落地；**不存在跨用户共享路径**，所有读取均默认按 `user_id = ctx.user_id` 约束。
- **CSRF**：Web 采用 Cookie + SameSite=Lax + `X-CSRF-Token` 双提交。
- **CORS**：按来源白名单；Admin 独立严格策略。
- **输入校验**：Proto schema + `protovalidate`。
- **速率限制**：Redis 令牌桶 + 登录阶梯退避。
- **Webhook 签名**：`X-AntClaw-Signature: sha256=HMAC(secret, ts + "." + body)`，带 `X-AntClaw-Timestamp` 防重放。
- **审计**：写/管理操作全量落 `audit_logs`。
- **密钥**：`.env` 或 Docker Secrets；生产 KMS/SOPS 可选。
- **依赖扫描**：`govulncheck` + `npm audit` + 镜像扫描（Trivy）。

---

## 十二、国际化（i18n）

### 11.1 目标

- 首发语言：`zh-CN`、`en-US`；架构上支持任意 BCP-47 locale。
- 全链路本地化：界面、错误、邮件、告警、AI 叙事、PDF/图表导出。
- 单一真相：翻译资源放在 `frontend/packages/i18n/`，后端构建时按需嵌入。

### 11.2 Locale 协商

优先级：`用户资料 locale` → `Accept-Language` → `客户端默认 locale`。写入 JWT 载荷 `locale` + `audience`（`app` / `admin`）字段，避免每次读库。

**按客户端分别兜底**（关键差异）：

| 客户端 | `audience` | 默认 locale | 未知 locale 兜底 |
|--------|-----------|-------------|-------------------|
| 用户 Web / 移动 App | `app` | `en-US` | `en-US` |
| 管理员 Web | `admin` | `zh-CN` | `zh-CN` |

服务端根据 JWT 的 `audience`（或请求入口路径 `/app/*` vs `/admin/*`）选择对应回退链；匿名请求（登录前）按入口路径判定。

### 11.3 后端

- 翻译引擎：`go-i18n/v2`（`toml` 或 `json` 目录）；消息键采用稳定的点分命名，如 `error.auth.invalid_credentials`、`alert.cot.release.title`。
- 响应包含 `message_key` + 已渲染 `message`；前端可选择直接展示或二次本地化。
- 数字/日期/货币：`golang.org/x/text`（`message.NewPrinter`、`number`、`currency`、`plural`）。
- 时区：用户资料 `timezone`（IANA）；所有对外时间输出带时区或 UTC + 偏移；内部存储一律 UTC。
- 邮件模板（注册激活、找回密码、告警摘要）：`templates/mail/<locale>/*.html`，缺失时回退到 `en-US`。
- AI 叙事：`Chat/Interpret/Outlook` 请求带 `locale`；提示词模板按语言切换；输出缓存 key 包含 `locale`。
- 告警：调度事件载荷**语言无关**（结构化数据）；渲染层按订阅者 `locale` 本地化后推送到 SSE/邮件/Webhook/Bot。

### 11.4 前端

- 库：`react-i18next`（Web/Admin）、`i18next` + `expo-localization`（移动 RN）。
- 资源：`frontend/packages/i18n/locales/<locale>/<namespace>.json`（namespace：`common / auth / dashboard / cot / calendar / ... / errors`）。
- 按需加载（按 namespace + locale 分包）；构建时校验所有 key 在所有 locale 存在，缺失 CI 告警。
- 图表/表格：数字、百分比、货币、日期使用 `Intl.*`；RTL 预留能力（未来阿语等）。
- 切换：全局 `LocaleSwitcher` 组件；切换后写入 `PUT /user.settings.locale`。

### 11.5 数据层

- 用户表新增 `locale VARCHAR(16)`、`timezone VARCHAR(64)`。
- 可翻译的静态文案（如告警模板标题、帮助条目）落 `i18n_strings(key, locale, text, updated_at)`，支持管理员在 Admin 控制台热更。
- 业务实体（如新闻事件）可选多语言字段 `title_i18n JSONB`（`{"zh-CN":"...","en-US":"..."}`）。

### 11.6 管理员控制台

- 提供 i18n 资源管理页：查看/编辑 `i18n_strings`、导入/导出 `json`、缺失键报表、翻译完成度仪表盘。

### 11.7 质量保障

- CI 校验：`i18n-check`（key 完整性、未使用 key、占位符一致性）。
- 单元测试：关键消息键渲染快照。
- 默认回退链（按 `audience` 区分）：
  - **app（用户端）**：`zh-TW → zh-CN → en-US`、`en-GB → en-US`、未知 → `en-US`。
  - **admin（管理端）**：`en-US → zh-CN`、`zh-TW → zh-CN`、未知 → `zh-CN`。
- 邮件模板：面向用户的邮件（注册激活、告警推送）兜底 `en-US`；面向管理员的内部邮件（审计告警、系统通知）兜底 `zh-CN`。

---

## 十三、可观测性（全默认启用）

- 后端日志：`zerolog` 结构化 JSON；轮卷仅留 30 天。
- 前/后端错误：**Sentry SaaS** 默认开启（DSN 由 `.env`）。
- 指标：Prometheus + Grafana；关键：`rpc_requests_total`、`rpc_duration_seconds`、`sse_connections`、`stream_delivery_total{channel,result}`、`job_runs_total{name,result}`、`pg_pool_*{role=primary|replica}`、`redis_pool_*`、`ai_tokens_total{provider,byok}`、`objectstore_ops_total`。
- 追踪：OTel 全链路 → Jaeger（或 Tempo）；RPC/PG/Redis/对象存储自动埋点。
- 健康：**当前 API** 提供 `GET /health`（简单存活）；`/readyz` 及 PG replica 深度探针为规划项，落地后与本节对齐。

---

## 十三A、实时性 SLA 与流事件契约

### SLA（P95）

| 事件 | 端到端时延 | 限制 |
|------|------|------|
| `alert.*` / `signal.*` | ≤ 500ms | 丢包率 < 0.1% |
| `price_tick` | ≤ 300ms | 满队丢最旧 |
| `task_progress` | ≤ 1s | 每任务 ≥ 1Hz |
| `notify.new` | ≤ 1s | 广播成功率 > 99% |

### 连接模型

- 单实例 SSE 上限 **10K** 连接；客户端通过 `/sse/v1/events?topics=alert,signal,price_tick:EURUSD` 同时订阅多主题。
- 客户端断连重连携 `Last-Event-ID`，服务端基于 Redis Streams 游标补发（最多 5 分钟）。
- 背压：客户端缓冲满达阈值 → 丢弃最旧事件，保证最新；连续失败则断开并发送 `stream.closed{reason}`。

### `stream.proto` 统一 schema

- Web SSE 与移动 gRPC server-streaming **共享**同一 `stream.proto`；SSE 编码为 `event: <type>\n data: <base64(protobuf)>`。
- 事件类型固定集：`alert.triggered`、`signal.emitted`、`price_tick`、`task_progress`、`notify.new`、`stream.keepalive`、`stream.closed`。
- 事件载荷**语言无关**；本地化由订阅者根据 `user.locale` 自己渲染。

---

## 十四、Docker Compose 一键部署（终态）

```
caddy              # 入口网关，自动 TLS；统一路由：
                   #   /rpc/* /sse/* /api/*  → antclaw-api
                   #   /app/*                → antclaw-web
                   #   /admin/*              → antclaw-admin
                   #   /grafana/* /minio/*   → 内部服务（admin 白名单）
postgres           # timescale/timescaledb:2.x-pg16（含 TimescaleDB 扩展）
# postgres-replica # 硬化期可选 profile：replica，开启后才启用
redis              # redis:7-alpine，AOF+RDB
minio              # minio/minio，S3 兼容
antclaw-api        # Connect + gRPC + SSE
antclaw-worker     # 调度 + Redis Streams 消费（轻型任务）
# antclaw-backtest-runner  # 后期 profile：LP-A 回测执行启用后再加入
# antclaw-mt-gateway       # 后期 profile：LP-C MT 等价层启用后再加入
antclaw-web        # 用户 Web 静态
antclaw-admin      # 管理员 Web 静态
prometheus         # 指标
grafana            # 可视化
jaeger             # 追踪（或 tempo）
```

要点：

- `.env`：`ANTCLAW_*`、`POSTGRES_PRIMARY_*`、`POSTGRES_REPLICA_*`、`REDIS_*`、`MINIO_*`、`S3_*`、`CADDY_DOMAIN`、`JWT_ED25519_PRIVATE/PUBLIC`、`ARGON2_*`、`ANTCLAW_BYOK_MASTER_KEY`、`SENTRY_DSN_BACKEND/FRONTEND`。
- `depends_on: service_healthy` 串联；primary 就绪后 replica 再起。
- 持久卷：`pgdata`、`redisdata`、`miniodata`（含 `audit-worm` bucket，启用 object lock）、`caddydata`、`grafanadata`。
- **Caddy SSE 反代**：对 `/sse/*` 路由必须 `flush_interval -1`、`transport http { versions h2 1.1 }`，读超时 ≥ 1 小时；关闭响应缓冲。
- **域名与 Cookie**（同域不同路径）：Cookie 用 `__Secure-antclaw_app` + `Path=/app` 和 `__Secure-antclaw_admin` + `Path=/admin` 隔离；**禁用 `__Host-` 前缀**（与 `Path≠/` 互斥）。`/app` 下发者 JWT `aud=app`、`/admin` 下发者 `aud=admin`，兼底 locale 分别为 `en-US` / `zh-CN`。

---

## 十五、实施路线图

| 阶段 | 目标 | 交付 |
|------|------|------|
| P0 准备（1 天） | 新建 AntClaw 仓库 + 基线分支 `main`；落实命名基线（§2.1）与 CI lint（零容忍 `ark` 字样）| 空仓可构建 |
| P2 Monorepo 骨架（1 天） | `backend/` `frontend/` `proto/` 结构；`buf` 配置 | 目录就位 |
| P3 契约先行（2 天） | 按 §3.3 编写全部 `.proto`；生成多语言 SDK | `gen/*` 可用 |
| P4 存储与鉴权（3 天） | PG + TimescaleDB、hypertable、sqlc、Redis；用户系统、Argon2id、Ed25519 JWT、会话；审计触发器 + hash 链 | `AuthService` 就绪 |
| P4a BadgerDB 迁移（1 天） | `antclaw-migrate` 首次部署必跑；幂等 | PG 数据就位 |
| P5 业务服务实现（7–10 天） | 按领域逐服务**全新实现** `internal/service/*`（Connect+gRPC 双暴露） | 全部只读服务 |
| P6 异步与 SSE（3 天） | `stream.proto` 统一契约 + Redis Streams + SSE 网关；`Last-Event-ID` 游标重连 | 长任务可用 |
| P6b 审计 WORM（0.5 天） | 审计双写 MinIO `audit-worm` + object lock | 审计不可篡改 |
| P6c 策略/MT 结构预留（0.5 天） | `user_strategies` 表、`strategy.proto`/`mt4.proto`/`mt5.proto` 骨架、`StrategyService/MT4Service/MT5Service` RPC 契约（返回 `UNIMPLEMENTED`）；前端页面隐藏或置「后期提供」 | 代码路径打通 |
| — **后期项目**— | | |
| LP-A 回测执行 | `antclaw-backtest-runner` 容器池、cpu/mem、工件上 MinIO、取消语义 | 后期 |
| LP-B 策略沙箱 | Starlark + 白名单 built-in + fuel/超时/内存 + runner 硬化 + 租户隔离 + MQL→DSL AI 转译 | 后期 |
| LP-C MT 等价层 | `antclaw-mt-gateway` 实现 `mt4.proto`/`mt5.proto` 对接经纪商 API；代替 MetaTrader | 后期 |
| — **继续 MVP** — | | |
| P7 Bot 接口预留（0.5 天） | `BotAdapter` 端口、`bot_bindings` 表、命令路由骨架（空实现） | 可编译、可单测 |
| P8 i18n 基础设施（2 天） | go-i18n + react-i18next + 资源目录 + CI；邮件/告警模板双语；用户端兼底 en-US / 管理端 zh-CN | `zh-CN`、`en-US` 覆盖 |
| P9 通知与站内信（2 天） | `notifications` + SSE + 推送骨架（过渡期仅 SSE+站内信，不走邮件） | Web 可收站内信 |
| P9a AI 用量与密钥探针（0.5 天） | `ai_usage` 写入、前端用量展示、BYOK 每日探针 | 用量可视 |
| P10 前端 Web（6–8 天） | 用户 Web（含语言切换、密码明文输入）；策略/回测页面隐藏或「后期」框 | 可演示 |
| P11 Admin 前端（5 天） | 控制台全模块 + i18n + 数据清理 + BYOK 开关 | 可上线 |
| P12 对象存储接入（1 天） | MinIO SDK + 回测/附件上传下载；预留 R2 切换开关 | 可用 |
| P13 硬化（3 天） | 压测、可观测性、Sentry、Caddy TLS、文档 | 准生产就绪 |
| P14 移动 RN（5–7 天，**后期**） | React Native + Expo 首版 | App Store/Google Play 内测 |
| P15 ArkTS（后期） | HarmonyOS 版立项交付 | HarmonyOS 市场 |

---

## 十六、验收标准

1. 全仓 `grep -ir "ark[-_ ]intelligent" .` 结果为 0（归档文档除外）。
2. 原功能清单每项能力至少对应一个 RPC 方法与契约测试通过。
3. `docker compose up` 后：caddy/postgres/redis/minio/api/worker/web/admin/prometheus/grafana/jaeger 全部 healthy；TimescaleDB 扩展已装载；hypertable 已创建；`antclaw-backtest-runner`/`antclaw-mt-gateway` 不在 MVP profile 中启动。
4. Web 可注册（即注册即用，无需激活）、登录、密码表单**明文显示**、浏览数据、SSE 实时收到站内信；用户端切 `en-US`/`zh-CN` 正常，未知 locale 兼底 `en-US`。
5. 管理端未知 locale 兼底 `zh-CN`；可管理用户/角色/任务/审计/i18n/BYOK 开关/数据清理/对象存储。
6. JWT 签名为 Ed25519；`aud` 区分 `app` / `admin`。
7. BadgerDB → PG 一次性迁移脚本可重跑幂等、数据正确入库。
8. `BotAdapter` 端口存在、接口单测通过，但无具体 Bot 实现投产。
9. 移动端本轮可缺席；交付时 RN 首版能登录、收推送、按系统语言本地化。
10. 监控：Prometheus/Grafana/Jaeger/Sentry 都有数据；审计表 append-only，`UPDATE/DELETE` 被触发器与权限共同拒绝；`audit-worm` bucket 具备 object lock。
11. 实时性：压测下 `alert/signal` P95 ≤ 500ms，`price_tick` P95 ≤ 300ms；`Last-Event-ID` 重连不丢事件。
12. 策略/回测/MT 层：`user_strategies` 表存在；`StrategyService/MT4Service/MT5Service` 每个 RPC 返回 `UNIMPLEMENTED`；前端相关入口不暴露或明确标「后期」；`antclaw-backtest-runner` / `antclaw-mt-gateway` **不启动**。
13. BYOK：只有当前用户密钥能解密成功；主密钥版本轮换下旧密文可解密；`ai_usage` 记录准确。
14. 开发仓库中任何脚本（含生成物）无超过 800 行的文件（MD 除外）。
15. 文档齐备且以中文位于 `docs/`。

---

## 十七、风险与对策

| 风险 | 对策 |
|------|------|
| 外部数据源限流/改版 | 回退链 + 熔断 + 缓存；健康探针显式降级 |
| AI 成本与限流 | 用户自负（BYOK）；缓存、冷却、模板回退、响应 `degraded` 标记；无 Key 时返回 `key_missing` |
| Argon2id 参数与硬件 | 压测选型；参数可热更 |
| Connect/gRPC 在部分企业网络 | 默认 Connect over HTTPS（兼容最好）；SSE 作为浏览器推送兜底 |
| 移动端长连接耗电 | gRPC server-streaming + 原生推送（FCM/APNs）双通道策略 |
| 从 BadgerDB 到 PG 的数据断层 | 一次性迁移脚本 + 新部署默认空库 |
| i18n 资源漂移 | CI 校验 key 完整性/占位符一致性；`i18n_strings` 支持热更 |
| 未来 Bot 接入改动大 | 提前冻结 `BotAdapter` 端口契约；账号绑定走通用 `bot_bindings` 表 |
| 前端密码明文展示 | 纯 UI 选择；依靠 TLS + Argon2id 保障安全；可在 `/settings` 单独为小白用户提示 |
| BYOK 密钥泄露 | AES-GCM + 主密钥轮换；日志/追踪打码保护；解密仅在调用瞬间 |
| 审计永久存储增长 | 管理端手动清理；hypertable 按块压缩；必要时归档到 MinIO/R2 |
| 时间序列规模迅速膨胀 | TimescaleDB 压缩 + `continuous_aggregate`；策略按 `(symbol, ts)` 切块 |
| 回测占用主进程资源 | 独立 `antclaw-backtest-runner`；compose 限制 cpu/mem |
| SSE 在 Caddy/网关变高延迟 | `flush_interval -1`；客户端 `Last-Event-ID` 补发；单实例 10K 连接封顶 |

---

## 十八、后续文档计划

- `docs/AntClaw-功能清单.md`
- `docs/AntClaw-Proto参考.md`（由 `buf` + 脚本生成）
- `docs/AntClaw-部署指南.md`
- `docs/AntClaw-迁移指南.md`
- `docs/AntClaw-订阅与SSE.md`
- `docs/AntClaw-用户系统与鉴权.md`
- `docs/AntClaw-管理员控制台.md`
- `docs/AntClaw-前端架构.md`
- `docs/AntClaw-移动端架构.md`
- `docs/AntClaw-国际化规范.md`
- `docs/AntClaw-Bot接入规范.md`（端口契约、绑定、路由、未来实施手册）
- `docs/AntClaw-领域模型.md`（各子域聚合根、值对象、领域事件、不变式）
- `docs/AntClaw-测试策略.md`（单测/契约/金融 golden 测试/回测一致性）
- `docs/AntClaw-任务分解与AI助手约束.md`（代码实现分发用）

---

## 附录 A · 依赖白名单（仅允许使用）

### A.1 后端 Go

| 包 | 最低版本 | 用途 |
|------|------|------|
| `github.com/jackc/pgx/v5` | v5.6+ | PG 驱动 |
| `github.com/sqlc-dev/sqlc` | v1.27+ | CLI |
| `github.com/pressly/goose/v3` | v3.21+ | 迁移 |
| `github.com/redis/go-redis/v9` | v9.5+ | Redis |
| `connectrpc.com/connect` | v1.16+ | RPC |
| `google.golang.org/grpc` | v1.65+ | gRPC |
| `github.com/bufbuild/protovalidate-go` | v0.6+ | 校验 |
| `github.com/rs/zerolog` | v1.33+ | 日志 |
| `go.opentelemetry.io/otel` | v1.28+ | 追踪 |
| `github.com/getsentry/sentry-go` | v0.28+ | 错误上报 |
| `github.com/nicksnyder/go-i18n/v2` | v2.4+ | i18n |
| `github.com/minio/minio-go/v7` | v7.0.70+ | S3/MinIO |
| `go.starlark.net/starlark` | 2024+ | 策略 DSL 引擎 |
| `github.com/awnumar/memguard` | v0.22+ | BYOK 解密结果内存安全 |
| `crypto/ed25519`（标准库）| — | JWT 签名 |
| `github.com/hibiken/asynq` | **禁止** | 已判定不引入 |

### A.2 前端

| 包 | 最低版本 | 用途 |
|------|------|------|
| `react` | 18.3+ | UI |
| `vite` | 5.4+ | 构建 |
| `@connectrpc/connect-web` | 1.5+ | RPC |
| `@tanstack/react-router` | 1.x | 路由 |
| `@tanstack/react-query` | 5.x | 数据 |
| `tailwindcss` | 3.4+ | 样式 |
| shadcn/ui CLI | 最新 | 组件 |
| `react-i18next` | 14+ | i18n |
| `echarts` 或 `lightweight-charts` | 最新 | 图表 |
| `@sentry/react` | 7+ | 错误上报 |
| `redux` / `mobx` / `axios` | **禁止** | 统一 `fetch` + connect |

### A.3 移动（后期）

- `expo` 最新 LTS；`react-native` 配套版本；`@connectrpc/connect-web`；`expo-localization`。

**引入此表以外的任何运行时依赖需先更新本文档并由用户拍板。**

---

## 附录 B · 命名与错误码词汇表

### B.1 事件 channel〈`<domain>.<verb>`〉

`alert.triggered`、`signal.emitted`、`price_tick`、`task_progress`、`notify.new`、`stream.keepalive`、`stream.closed`。新增事件必在 `stream.proto` 注册枚举。

### B.2 Redis Streams 键名

- 任务：`stream:tasks:<kind>`（kind 例如 `backtest`、`ai_chat`、`ingest_cot`）
- 扇出：`stream:events:<channel>`
- 消费者组：`cg:<service>:<instance>`

### B.3 i18n key 命名

- 结构：`<domain>.<entity>.<attribute>`，全小写点分，例 `error.auth.invalid_credentials`、`alert.cot.release.title`。
- 占位符：ICU MessageFormat，不允许 `%s`、`{0}`等其他风格。

### B.4 错误码（RPC `status.code` = Proto `Code` 枚举）

| Code | 含义 |
|------|------|
| `AUTH_INVALID_CREDENTIALS` | 邮箱/密码错误 |
| `AUTH_TOKEN_EXPIRED` | JWT 过期 |
| `AUTH_FORBIDDEN` | 角色/配额不足 |
| `USER_NOT_FOUND` | 用户不存在 |
| `USER_LOCKED` | 封禁 |
| `RATE_LIMITED` | 限流 |
| `KEY_MISSING` | BYOK 未配 |
| `KEY_INVALID` | BYOK 失败 |
| `AI_UPSTREAM_ERROR` | provider 报错 |
| `AI_DEGRADED` | 回退模板 |
| `BACKTEST_CANCELED` | 用户中断 |
| `BACKTEST_RESOURCE_LIMIT` | 超 cpu/mem |
| `STREAM_BACKPRESSURE` | 背压丢包 |
| `VALIDATION_FAILED` | `protovalidate` |
| `CONFLICT` | 版本冲突 |
| `INTERNAL` | 其他 |

新增错误码必须先更新本表并在 `common.proto` 枚举注册。

---

## 附录 C · 测试策略硬规则

1. **单测覆盖率**：后端核心包（`domain`、`service`、`auth`、`byok`、`notify`）≥ 80%；其他 ≥ 60%。CI 硬门槛。
2. **契约测试**：每个 Proto 方法至少一个 `buf curl` 或 Connect 客户端的实体调用用例；破坏性变更必须 `buf breaking` 通过。
3. **金融 golden 测试**：指标计算（MA/信号/回测统计）提供固定输入与固定输出 `testdata/*.golden`；精度 `numeric(20,10)` 严格比对。
4. **回测↔实盘一致性**：每个策略必须有一个集成用例，证明回测引擎与实盘信号生成器调用同一 `SignalEvaluator`。
5. **流压测**：`k6` 模拟 1K 并发 SSE，P95 ≤ 500ms是验收硬指标。
6. **测试修辞禁令**：不得为通过测试删除/弱化 assertion；新增 bug 修复必须先添加回归测试。

---

## 附录 D · 精度与时区约定

- 金额/价格/收益率：`numeric(20,10)`，禁止 `float/double`；前端用 `decimal.js` 或后端渲染后的字符串。
- 比率/百分比：后端返回 `0.1234` 与 `display_percent: true`，前端只做格式化。
- 时间存储：全部 UTC `timestamptz`；对外输出带时区或 ISO-8601+Z。
- 客户端显示：默认用用户 `timezone`，无则浏览器 `Intl.DateTimeFormat().resolvedOptions().timeZone`，并写回 `users.last_seen_timezone`。

---

## 附录 E · AntClaw-Strategy-DSL 契约

### E.1 支持的 built-in API

| 函数 | 返回 | 说明 |
|------|------|------|
| `bar(i:int)` | dict | `i=0` 当前 K 线；i>0后回望；i<0 驳回 |
| `bars(n:int)` | list | 最近 n 根 K 线倒序 |
| `ind(name:str, *args)` | list/number | 白名单：`EMA/SMA/WMA/RSI/ATR/MACD/BB/STOCH/ADX` |
| `buy(lot, sl=None, tp=None)` | bool | 挂买单 |
| `sell(lot, sl=None, tp=None)` | bool | 挂卖单 |
| `close_all()` | None | 平掉当前仓 |
| `position()` | dict | 当前持仓快照 |
| `param(name:str, default)` | number/bool | 读异外参数 |
| `log(msg:str)` | None | 运行层日志 |
| `now()` | ts | 当前模拟时间 |

禁用：`print`、`load`、`open`、`getattr`、任何未白名单的 built-in。

### E.2 脚本入口

必须定义函数：

```python
def on_init(ctx):
    pass

def on_bar(ctx):
    # 每根 K 线调用一次
    pass

def on_deinit(ctx):
    pass
```

### E.3 硬限默认值

- fuel（1e8指令）；wall-clock（回测 30s / 优化 60s）；内存 256 MiB；栈深度 256；脚本 ≤ 128 KiB。
- Admin 可调，但上限 ≤ fuel 1e9 / wall-clock 300s / 内存 1 GiB。

### E.4 转译器接口

`AIService.TranslateStrategy(source_kind, source) -> { dsl, unsupported[], warnings[] }`，使用用户 BYOK 调用。驳回条件：`unsupported` 非空 → `status=invalid`。

---

> **结论**：AntClaw v2 以 **Connect + gRPC + SSE** 为统一 API 通道，由自有 **Web 与后期移动 App**（RN+Expo 首发、ArkTS 后期）承载业务；**PostgreSQL 16 + TimescaleDB（首发单实例，读写分离硬化期可选）+ Redis 单实例持久化 + MinIO（后期 R2）** 构成持久化与对象存储；**免激活用户系统（Argon2id）+ Ed25519 JWT + AI BYOK + 站内信/移动推送通知 + 全链路 i18n（用户端兼底 en-US，管理端兼底 zh-CN） + append-only 审计**；运维全栈默认启用 Caddy、OTel、Jaeger、Prometheus、Grafana、Sentry；所有业务计算由后端完成，前端密码明文显示；代码「先分析再实现」、脚本超 800 行必拆。对旧 Telegram Bot 仅保留可插拔 `BotAdapter` 接口，本期不实现；单机单套 Docker Compose 为终态部署形态。
