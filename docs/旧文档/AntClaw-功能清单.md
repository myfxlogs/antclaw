# AntClaw · 功能清单（基线）

> 本文以 `docs/ARK-Intelligent-功能清单.md` 为**功能参照**（仅对齐业务能力与语义，**不涉及代码复用**；参见《任务分解与 AI 助手约束》宪章 11），**剔除** Telegram Bot 专属章节（本期 Bot 不交付，见《Bot 接入规范》），将每项能力在 AntClaw **全新实现**为 **Proto Service + RPC** 与 **前端模块**。本文是功能完整性的基线：**AntClaw 上线后任一能力缺失均视为回退**（决策：功能零回退）。

## 一、使用方式

- 每一项能力包含：
  - **能力名**：旧清单原名；
  - **Proto Service.Method**：新 RPC 契约入口；
  - **流事件（如有）**：`stream.proto` 中对应 `StreamChannel`；
  - **前端模块**：`frontend/web/src/features/<module>`（移动端默认对齐，占位除外）；
  - **本期状态**：`MVP` / `占位（后期）` / `后期`。
- 详细字段以 `proto/antclaw/v1/*.proto` 为准；本文不重复字段定义（见《领域模型》）。

## 二、身份与会话

| 能力 | Proto | 流 | 前端模块 | 状态 |
|---|---|---|---|---|
| 注册（邮箱+密码，免激活） | `AuthService.Register` | — | `auth` | MVP |
| 登录 / 刷新 / 登出 | `AuthService.Login/Refresh/Logout` | — | `auth` | MVP |
| 密码找回 | `AuthService.RequestPasswordReset/ResetPassword` | — | `auth` | MVP |
| 邮箱验证（可选） | `AuthService.VerifyEmail` | — | `auth` | MVP |
| 获取个人资料 | `UserService.GetMe` | — | `settings` | MVP |
| 偏好设置（locale/timezone/通知） | `UserService.UpdateSettings` | — | `settings` | MVP |
| 会话管理（列出/吊销） | `UserService.ListSessions/RevokeSession/RevokeAllOtherSessions` | — | `settings` | MVP |
| BYOK 配置与探针 | `UserService.SetAiKey/GetAiKeyMeta` | — | `settings` | MVP |
| 订阅层级 | `UserService.GetMembership` | — | `settings` | MVP（计费延后） |
| 新用户引导 | `UserService.StartOnboarding` | — | `dashboard` | MVP |
| 历史浏览 | `UserService.GetHistory/ClearHistory` | — | `settings` | MVP |
| 收藏（Pin） | `UserService.ListPins/Pin/Unpin` | — | 跨模块 | MVP |
| 用户反馈 | `UserService.SubmitFeedback` | — | `settings` | MVP |
| 2FA（TOTP） | `UserService.EnableTotp/DisableTotp` | — | `settings` | **延后** |

## 三、管理员

| 能力 | Proto | 状态 |
|---|---|---|
| 用户查询/改角色/封禁/解封/强制登出/重置密码 | `AdminService.ListUsers/GetUser/SetRole/Ban/Unban/ForceLogout/ResetPassword` | MVP |
| 任务调度与手动触发 | `AdminService.ListJobs/GetJob/RunJob` | MVP |
| 审计日志查询 | `AdminService.ListAuditLogs` | MVP |
| Webhook 投递日志 | `AdminService.ListWebhookDeliveries` | MVP |
| 数据源健康与降级 | `AdminService.UpdateDatasourceHealth` | MVP |
| 手动数据清理 | `AdminService.CleanupData` | MVP |
| i18n 资源编辑与热更新 | `AdminService.UpdateI18nString/ReloadI18n` | MVP |
| 反馈工单管理 | `AdminService.ListFeedback/UpdateFeedback` | MVP |
| 2FA 全局强制开关 | `AdminService.UpdateSecurityPolicy` | **延后** |

## 四、COT（商品期货持仓）

| 能力 | Proto | 流 | 前端 | 状态 |
|---|---|---|---|---|
| 持仓总览 | `CotService.GetSummary` | — | `cot` | MVP |
| 对比（多品种） | `CotService.Compare` | — | `cot` | MVP |
| 衍生信号 | `CotService.GetSignals` | `CHANNEL_COT` | `cot` | MVP |
| 历史 | `CotService.GetHistory` | — | `cot` | MVP |
| 对某品种订阅告警 | `AlertService.Subscribe(channel=alerts, filter.cot.*)` | `CHANNEL_ALERTS` | `alerts` | MVP |

## 五、财经日历

| 能力 | Proto | 前端 | 状态 |
|---|---|---|---|
| 列表 | `CalendarService.ListEvents` | `calendar` | MVP |
| 详情 | `CalendarService.GetEvent` | `calendar` | MVP |
| 影响评估 | `CalendarService.GetImpact` | `calendar` | MVP |
| 影响历史 | `CalendarService.GetImpactHistory` | `calendar` | MVP |

## 六、宏观（Macro）

| 能力 | Proto | 前端 | 状态 |
|---|---|---|---|
| FRED | `MacroService.GetFred` | `macro` | MVP |
| ECB | `MacroService.GetEcb` | `macro` | MVP |
| SNB | `MacroService.GetSnb` | `macro` | MVP |
| OECD 领先指标 | `MacroService.GetOecdLeading` | `macro` | MVP |
| Eurostat | `MacroService.GetEurostat` | `macro` | MVP |
| BIS | `MacroService.GetBis` | `macro` | MVP |
| TradingEconomics | `MacroService.GetTradingEconomics` | `macro` | MVP |
| DTCC 掉期 | `MacroService.GetDtccSwaps` | `macro` | MVP |
| SEC 13F | `MacroService.GetSec13F` | `macro` | MVP |
| 国债拍卖 | `MacroService.GetTreasuryAuctions` | `macro` | MVP |
| FedWatch | `MacroService.GetFedWatch` | `macro` | MVP |
| 世界银行 | `MacroService.GetWorldBank` | `macro` | MVP |
| IMF WEO | `MacroService.GetImfWeo` | `macro` | MVP |
| 制度切换告警 | 订阅 `CHANNEL_REGIME` | `alerts` | MVP |

## 七、价格与市场

| 能力 | Proto | 流 | 前端 | 状态 |
|---|---|---|---|---|
| 报价 | `PriceService.GetPrice` | `CHANNEL_PRICE_TICKS` | `price` | MVP |
| 关键位 | `PriceService.GetLevels` | — | `price` | MVP |
| 市场总览 | `PriceService.GetMarketOverview` | — | `dashboard` / `price` | MVP |
| 交易时段 | `PriceService.GetSession` | — | `price` | MVP |
| 情景推演 | `PriceService.RunScenario` | — | `price` | MVP |
| 市场制度 | `PriceService.GetRegime` | `CHANNEL_REGIME` | `price` | MVP |
| 季节性 | `PriceService.GetSeasonal` | — | `price` | MVP |

## 八、波动率

| 能力 | Proto | 流 | 前端 | 状态 |
|---|---|---|---|---|
| VIX | `VolService.GetVix` | — | `vol` | MVP |
| MOVE | `VolService.GetMove` | — | `vol` | MVP |
| DVOL | `VolService.GetDvol` | — | `vol` | MVP |
| GEX | `VolService.GetGex` | — | `vol` | MVP |
| IV | `VolService.GetIvol` | — | `vol` | MVP |
| 偏度 | `VolService.GetSkew` | — | `vol` | MVP |
| 偏度/VIX 告警 | `VolService.GetSkewVixAlert` | `CHANNEL_SKEW_VIX` | `alerts` | MVP |

## 九、信号

| 能力 | Proto | 流 | 前端 | 状态 |
|---|---|---|---|---|
| 偏好 | `SignalService.GetBias` | — | `signals` | MVP |
| 排名 | `SignalService.GetRank` | — | `signals` | MVP |
| X 因子 | `SignalService.GetXFactors` | — | `signals` | MVP |
| 雷达 | `SignalService.GetRadar` | — | `signals` | MVP |
| 强度 | `SignalService.GetIntensity` | — | `signals` | MVP |
| 制度迁移 | `SignalService.GetTransition` | `CHANNEL_REGIME` | `signals` | MVP |
| Crypto Alpha | `SignalService.GetCryptoAlpha` | — | `signals` | MVP |
| 统一信号 | `SignalService.GetUnified` | `CHANNEL_SIGNALS` | `signals` | MVP |
| Quant | `SignalService.GetQuant` | — | `signals` | MVP |
| CTA | `SignalService.GetCta` | — | `signals` | MVP |
| 简报 | `SignalService.GetBriefing` | `CHANNEL_BRIEFING` | `dashboard` | MVP |
| 展望 | `SignalService.GetOutlook` | — | `dashboard` | MVP |

## 十、技术分析（TA）

| 能力 | Proto | 前端 | 状态 |
|---|---|---|---|
| 指标 | `TaService.GetIndicators` | `ta` | MVP |
| 艾略特浪 | `TaService.GetElliott` | `ta` | MVP |
| Wyckoff | `TaService.GetWyckoff` | `ta` | MVP |
| ICT | `TaService.GetIct` | `ta` | MVP |
| 拍卖市场理论 | `TaService.GetAmt` | `ta` | MVP |
| 订单流 | `TaService.GetOrderflow` | `ta` | MVP |
| 量能分布 | `TaService.GetVolumeProfile` | `ta` | MVP |
| 跨市场 | `TaService.GetIntermarket` | `ta` | MVP |

## 十一、情绪与链上

| 能力 | Proto | 流 | 前端 | 状态 |
|---|---|---|---|---|
| 情绪 | `SentimentService.GetSentiment` | — | `sentiment` | MVP |
| 链上 | `SentimentService.GetOnchain` | — | `sentiment` | MVP |
| DeFi 健康 | `SentimentService.GetDefiHealth` | — | `sentiment` | MVP |
| Carry 监测 | `SentimentService.GetCarryMonitor` | `CHANNEL_CARRY` | `sentiment` | MVP |

## 十二、AI

| 能力 | Proto | 流 | 前端 | 状态 |
|---|---|---|---|---|
| 聊天（server-stream） | `AiService.Chat` | gRPC stream | `ai` | MVP（BYOK） |
| 解读 | `AiService.Interpret` | — | 跨模块 | MVP（BYOK） |
| 展望 | `AiService.Outlook` | — | `dashboard` | MVP（BYOK） |
| 策略转译（MQL → DSL） | `AiService.TranslateStrategy` | — | `strategy` | **后期** |
| AI 用量看板 | `AiService.GetUsage` | — | `settings` | MVP |

## 十三、告警与订阅

| 能力 | Proto | 流 | 前端 | 状态 |
|---|---|---|---|---|
| 订阅列表 | `AlertService.ListSubscriptions` | — | `alerts` | MVP |
| 订阅/取消/更新 | `AlertService.Subscribe/Unsubscribe/UpdateSubscription` | — | `alerts` | MVP |
| 实时告警流 | — | `CHANNEL_ALERTS` | `alerts` | MVP |
| 管理员 Webhook | `AlertService.RegisterWebhook/ListWebhooks` | — | `admin` | MVP（仅管理员） |

## 十四、策略与回测（占位）

| 能力 | Proto | 前端 | 状态 |
|---|---|---|---|
| 运行回测 | `BacktestService.RunBacktest` | `backtest` | **占位**（RPC 返 `UNIMPLEMENTED`） |
| 查询回测 | `BacktestService.GetBacktest` | `backtest` | **占位** |
| 准确率 | `BacktestService.GetAccuracy` | `backtest` | **占位** |
| Quant/VP/CTA 回测 | `BacktestService.RunQuantBt/RunVpBt/RunCtaBt` | `backtest` | **占位** |
| 用户策略上传 | `StrategyService.*` | `strategy` | **占位** |
| MT4/MT5 等价层 | `MT4Service.*` / `MT5Service.*` | — | **占位**（后期 `antclaw-mt-gateway`） |

> 占位要求：proto 定义齐全；service 方法 `return codes.Unimplemented`；`cmd/antclaw-backtest-runner/` 目录骨架（`main.go` 仅打印日志后退出）；`internal/adapter/sandbox/` 目录骨架（空实现）；前端页面显示 `ui.placeholder.coming_later`。见《任务分解与 AI 助手约束》P6c。

## 十五、通知

| 能力 | 通道 | 状态 |
|---|---|---|
| 站内信 | `notifications` 表 + SSE/流推送 | MVP |
| 移动推送 | FCM / APNs / HMS | MVP（移动端上线后） |
| 邮件 | 仅注册通知 + 密码找回 + 安全通知 | MVP |
| Webhook（管理员） | 出站 | MVP |
| 日常告警走邮件 | — | **禁用**（决策 #14） |

## 十六、国际化

| 能力 | 说明 | 状态 |
|---|---|---|
| `zh-CN` / `en-US` 资源齐全 | 前后端一致 | MVP |
| locale 协商 + 兜底（用户端 en-US / 管理端 zh-CN） | — | MVP |
| 业务多语字段 `title_i18n JSONB` + `i18n_pick` | PG 侧函数 | MVP |
| `ja-JP` / `ko-KR` / `zh-TW` 预留 | 结构支持、资源未齐 | 预留 |

## 十七、观测与运维

| 能力 | 工具 | 状态 |
|---|---|---|
| 指标 | Prometheus + Grafana | MVP |
| 追踪 | OTel + Jaeger/Tempo | MVP |
| 错误上报 | Sentry SaaS（backend + frontend） | MVP |
| 审计 | `audit_logs`（append-only，哈希链） + MinIO WORM 双写 | MVP |
| 备份 | `pg_dump` 日备 + MinIO | MVP |
| 升级/回滚 | compose + 迁移 CLI | MVP |

## 十八、Bot 接入（预留）

| 能力 | 说明 | 状态 |
|---|---|---|
| `BotAdapter` 端口 + `bot_bindings` 表 | 仅接口与空实现 | MVP |
| Telegram / WeChat / Feishu 具体实现 | — | **后期** |

## 十九、与旧清单的差异说明

以下旧清单条目在 AntClaw 体系下发生**形态变化**或**落入占位**：

| 旧条目 | 变化 |
|---|---|
| Telegram Bot 全部命令（`/cot`、`/price`、`/setalert` 等） | 能力由 RPC + Web/移动承担；Bot 命令表作为未来接入参考保留 |
| Bot 权限与 RBAC | 统一走 `users.role` |
| Bot 冷却/配额 | 由《用户系统与鉴权》§七 统一配额替代 |
| 回测直接在 Bot 中运行 | 后期由 `antclaw-backtest-runner`（隔离）承担 |
| 策略以 MQL 形式执行 | 后期经 `AiService.TranslateStrategy` → Starlark DSL → runner 执行；**不使用** Wine + MT terminal |
| Telegram 下载附件 | 改为 `MinIO/R2` 预签名 URL；前端点开 |
| Bot i18n 简化 | 统一 ICU + `packages/i18n` |

## 二十、能力完备度验收

**上线前必过**：

- [ ] 本文 §二 ~ §十三、§十五 ~ §十七 全部条目：MVP 标记项 Proto + Handler + 前端模块齐全。
- [ ] §十四（占位项）：proto 编译通过且返回 `UNIMPLEMENTED`；前端显示「后期提供」。
- [ ] §十八 Bot 端口与空实现通过单测。
- [ ] 任一 MVP 条目能在 Web 前端完成一条端到端调用（RPC 调通、i18n 正确、SSE 推送可达）。

## 二十一、已决事项（2026-04-24）

- **FN1 · Pin 跨模块共用一张表**：**不拆分**；统一 `user_pins(user_id, ref_type, ref_id, created_at)`，`ref_type` 为各领域聚合标识（如 `cot_symbol`、`calendar_event`、`signal_id`…）；细节补充至《领域模型》。
- **FN2 · `GetHistory` 保留窗口**：**永久**；不限制最大条数；仅由 `UserService.ClearHistory` 提供用户主动清理入口。
- **FN3 · `TranslateStrategy` 指令集**：本期不拍板；**策略沙箱启动前定**；MVP 阶段 `AiService.TranslateStrategy` 仍返回 `UNIMPLEMENTED`。
