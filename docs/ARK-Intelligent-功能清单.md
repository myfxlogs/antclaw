# ARK Intelligent 项目功能清单

> 本文档基于 `Emulator/ark-intelligent` 参考仓库源码系统梳理，按模块完整列出当前已实现的全部功能。所有 Telegram 命令均来自 `internal/adapter/telegram/handler.go` 等注册点；各服务能力对应 `internal/service/*` 下的实现文件。

---

## 一、项目定位与总体架构

- **项目定位**：面向外汇/宏观交易者的**机构级宏观智能 Telegram Bot**，聚合 COT 持仓、FRED 宏观、经济日历、技术分析、期权/波动率、链上/DeFi、订单流等多策略信号。
- **语言与运行时**：Go 1.22+，通过 Docker Compose 部署，内置健康检查端点（`:8080`）。
- **存储**：嵌入式 BadgerDB（零外部依赖）。
- **消息层**：Telegram Bot API（长轮询）。
- **AI 层**：Google Gemini / Anthropic Claude，带缓存与限流；服务不可用时回退到模板输出。
- **架构**：六边形（端口与适配器）
  - **Telegram 适配器层**：命令解析、中间件、键盘、聊天机器人模式；
  - **服务层**：COT / FRED / 新闻 / 价格 / AI / 策略 / 回测 等；
  - **领域层**：实体、值对象、领域事件（`internal/domain`）；
  - **存储适配器层**：BadgerDB 各仓储（`internal/adapter/storage`）。

---

## 二、Telegram 机器人功能

### 2.1 基础命令

| 命令 | 功能 |
|------|------|
| `/start` | 首次启动，触发新手引导 |
| `/help` | 查看全部命令 |
| `/onboarding` | 重新进入新手引导流程 |
| `/settings` | 个人设置面板 |
| `/status` | 系统/用户状态 |
| `/membership` | 会员等级与升级信息 |
| `/clear` | 清空聊天历史 |
| `/history`、`/h` | 查看历史记录 |
| `/pin`、`/unpin`、`/pins` | 置顶命令管理 |
| `/feedback` | 提交反馈（`handler_feedback.go`）|

### 2.2 COT（CFTC 持仓分析）

- `/cot`（别名 `/c`、`/ce`）：COT 持仓摘要（全币种或单对）。
- 支持净变化、百分位排名、信号检测、置信度分级。
- 周度刷新；数据发布后自动广播（`scheduler.broadcastCOTRelease`）。
- 提供 COT 对比视图（`handler_cot_compare.go`）及键盘下钻（回调 `cot:`）。

### 2.3 经济日历与新闻影响

- `/calendar`、`/cal`：未来高影响经济事件列表，支持筛选 (`cal:filter:`) 与翻页 (`cal:nav:`)。
- `/impact`：事件历史影响评分与累计惊喜度（`handler_impact_cmd.go`）。
- 新闻模块 (`internal/service/news`)：MQL5 日历抓取、修订检测、Actual vs Consensus 惊喜打分、Fed RSS、影响自举 (`impact_bootstrap.go`)。

### 2.4 宏观 & 央行

| 命令 | 功能 |
|------|------|
| `/macro`、`/m` | FRED 宏观仪表盘（收益率/劳动力/通胀）|
| `/ecb` | 欧央行货币政策仪表盘（SDW）|
| `/snb` | 瑞央行资产负债/FX 干预代理 |
| `/leading` | OECD 综合领先指标 |
| `/eurostat`、`/eu` | 欧盟经济数据（Eurostat）|
| `/bis`、`/cbrates` | BIS 央行政策利率 + 信贷缺口 + REER |
| `/tedge`、`/globalm` | TradingEconomics 全球宏观仪表盘 |
| `/swaps` | DTCC 外汇互换机构流 |
| `/13f` | SEC EDGAR 13F 机构持仓 |
| `/treasury` | 美国国债拍卖结果 |

实现对应 `internal/service/macro/*`（ECB / OECD / Eurostat / SNB / DTCC / TE / Treasury 客户端）、`internal/service/bis`（cbpol / creditgap / reer）、`internal/service/worldbank`、`internal/service/imf`、`internal/service/sec`、`internal/service/treasury`、`internal/service/fed`（FedWatch 概率）。

### 2.5 价格 & 行情

- `/price`、`/p`：日线价格上下文。
- `/levels`、`/l`：支撑/阻力与头寸规模建议。
- `/market`：跨资产行情总览（Firecrawl + Finviz）。
- `/session`：伦敦/纽约/东京交易时段行为分析。
- `/scenario`：蒙特卡洛价格路径情景生成。
- `/regime`：多资产 HMM 状态仪表盘。
- `/seasonal`：季节性模式。
- 多源价格回退链：TwelveData → AlphaVantage → Yahoo → CoinGecko → CryptoCompare（`internal/service/price`、`internal/service/marketdata/*`）。
- 内置 GARCH、Hurst 指数、相关性、背离检测、日内/周频聚合。

### 2.6 波动率 & 期权

- `/vix`：CBOE 波动率指数套件（含 VIX 期限结构、MOVE、跨市场波动率）。
- `/gex`：Gamma 曝险 (`internal/service/gex`)。
- `/ivol`：隐含波动率曲面。
- `/skew`：偏度分析；内含 SKEW/VIX 尾部风险告警 (`skew_vix_alert.go`)。
- Deribit DVOL 集成 (`internal/service/dvol`)、情绪模块中的 DVOL 桥接。

### 2.7 策略、信号与回测

| 命令 | 功能 |
|------|------|
| `/bias`、`/b` | 基于 COT + 宏观的方向偏好 |
| `/rank`、`/r` | 货币强弱排名（COT + 宏观汇流）|
| `/xfactors` | 跨因子汇流分析 |
| `/radar` | 统一 Alpha 雷达仪表盘 |
| `/intensity` | 信号强度（原 `/heat`）|
| `/transition` | 宏观状态转换追踪 |
| `/cryptoalpha` | 加密资产专用宏观信号 |
| `/signal` | 统一方向信号（COT+CTA+Quant+情绪+季节性）|
| `/setalert` | 按币对订阅 COT 告警 |
| `/backtest`、`/bt`、`/bta` | 信号回测 |
| `/accuracy` | 信号准确率统计 |
| `/report` | 综合报告 |
| `/quant`、`/q`、`/qe` | 量化信号 |
| `/qbacktest`、`/quantbacktest` | 量化回测（状态缓存 + 交互）|
| `/quantbt` | QuantBT 引擎回测 |
| `/vpbt` | 成交量轮廓回测 |
| `/cta`、`/ca` | CTA 趋势跟踪分析 |
| `/ctabt` | CTA 回测 |
| `/briefing`、`/br` | 每日宏观简报 |
| `/outlook`、`/out`、`/of` | AI 驱动的每周宏观展望 |

回测与评估能力（`internal/service/backtest`）：

- Walk-forward 回测、Bootstrap、蒙特卡洛、Excursion（MFE/MAE）分析；
- 成本模型（滑点/点差/佣金）；
- 衰减分析、去重、基线对比、因子分解；
- Logistic / Platt / Isotonic 置信度校准；
- 每个宏观状态分别统计命中率、Sharpe、Sortino、最大回撤、盈亏比。

策略引擎 (`internal/service/strategy`)：多因子汇流引擎 + Risk Parity 头寸分配器。

因子库 (`internal/service/factors`)：动量、低波动、趋势质量、Carry 调整、残差反转、拥挤度、资金流背离、组合画像。

### 2.8 技术分析（TA）

服务模块 `internal/service/ta` 提供：

- 经典指标：RSI、MACD、布林带、ATR、Fibonacci、Ichimoku、Supertrend；
- Elliott 波浪（独立模块 `internal/service/elliott`，含校验器、投影器、ZigZag）；
- ICT 方法（`internal/service/ict`）：结构、Swing、Order Block、FVG、流动性；
- Wyckoff 方法（`internal/service/wyckoff`）：阶段、分类器、事件、摘要；
- AMT（Auction Market Theory）日型、开盘/收盘分类、轮动、迁移；
- 订单流 / 成交量侧（`internal/service/orderflow`）：Delta、吸收、POC。

相关命令：`/elliott`、`/wyckoff`、`/ict`、`/auction`、`/vp`、`/orderflow`、`/flows`、`/intermarket`、`/tedge`。

### 2.9 情绪、链上与 DeFi

- `/sentiment`、`/s`：综合情绪（CBOE、MyFxBook、OpenInsider 等）。
- `/onchain`：CoinMetrics / 区块链客户端交易所资金流指标。
- `/defi`：DefiLlama TVL、DEX 成交、稳定币供应健康度。
- `/carry`：Carry 交易监测与平仓预警（FRED 端 `carry_monitor.go`）。

### 2.10 管理员与运营

- `/users`、`/setrole`、`/ban`、`/unban`：用户与权限管理，且带二次确认 (`adm_cf:`)。
- 会员分层 (`handler_onboarding.go`、`format_briefing.go`)：Free → 仅 USD/高影响；Premium → 全部告警。
- 告警冷却 / AI 冷却 / 每用户配额 (`checkAICooldown`、`checkAICooldownDynamic`)。
- 反馈系统、深链 (`deeplink.go`)。

### 2.11 交互增强

- 回调路由：`cot:`、`alert:`、`set:`、`alertmgr:`、`cal:filter:`、`cal:nav:`、`out:`、`cmd:`、`onboard:`、`tutorial:`、`macro:`、`imp:`、`nav:`、`help:`、`gex:`、`ivol:`、`skew:`、`setalert:`、`share:`、`adm_cf:`、`briefing:`、`hist:`、`sentiment:`、`view:`、`alpha:`、`cta:`、`ctabt:`、`quant:`、`qbacktest:`、`quantbt:`、`vpbt:`、`ict:`、`wck:`。
- 消息分片器 (`api_chunk_tracker.go`) 处理 Telegram 4096 字符限制。
- 紧凑/完整视图切换 (`cbViewToggle`)。
- 短别名：`/c /m /b /r /s /p /l /q /bt /out /cal /br /h`。
- 组合别名：`/ce`=`/cot EUR`、`/qe`=`/quant EUR`、`/bta`=`/backtest all`、`/of`=`/outlook fred` 等。

---

## 三、AI 层

位于 `internal/service/ai`：

- 多 Provider：Gemini、Claude（`gemini.go`、`claude.go`、`claude_analyzer.go`）。
- `interpreter.go`、`cached_interpreter.go`：响应缓存，命中后跳过调用。
- `ai_ratelimit.go`：限流（默认 `AI_MAX_RPM=15`）。
- `context_builder.go`、`prompts.go`：统一提示词与上下文拼装。
- `unified_outlook.go`：统一周度展望生成。
- `memory_store.go`、`tool_executor.go`、`tools.go`：工具调用（记忆、文件搜索、API 调用）。
- `chat_service.go`：非命令消息 → 聊天模式，支持上下文 & 分块。
- AI 服务不可用时降级为模板输出（"graceful template fallback"）。

---

## 四、数据管线与后台调度

`internal/scheduler/scheduler.go` 注册的后台任务：

| 任务 | 说明 |
|------|------|
| `jobCOTFetch` | 按周抓取 CFTC Socrata COT，驱动分析与广播 |
| `jobWeeklyOutlook` | 每周日 18:00 WIB 生成并推送周度展望 |
| `jobFREDAlerts` | 每小时拉取 FRED，宏观状态变化 → 广播 |
| `checkSKEWVIXAlert` | 尾部风险状态转变告警（TASK-208）|
| `jobCarryAlerts` | 每 4 小时检查 Carry 平仓 |
| `jobPriceFetch` | 周度价格 |
| `jobDailyPriceFetch` | 日线 OHLCV |
| `jobIntradayPriceFetch` | 15m 聚合至更高周期 |
| `jobSignalEval` | 回测评估未决信号 |
| `jobRetentionCleanup` | 每日 03:00 WIB 数据保留期清理 |
| `ImpactBootstrapper` | 启动时一次性回填历史事件影响 |
| `scheduler_briefing.go` | 每日简报推送 |
| `scheduler_pair_alerts.go` | 按币对告警 |
| `scheduler_regime.go` | 宏观状态广播 |
| `scheduler_skew_vix.go` | SKEW/VIX 尾部风险定时检查 |
| `alert_gate.go` | 告警闸门：按用户偏好/免打扰/分层过滤并记账 |

所有任务：Panic 恢复、5 分钟 per-job 超时、优雅关机（≤10s）。

---

## 五、存储层（BadgerDB 仓储）

位于 `internal/adapter/storage`：

- `cot_repo.go`、`event_repo.go`、`price_repo.go`、`daily_price_repo.go`、`intraday_repo.go`、`signal_repo.go`、`news_repo.go`、`user_repo.go`、`prefs_repo.go`、`fred_repo.go`、`impact_repo.go`、`feedback_repo.go`、`memory_repo.go`、`conversation_repo.go`、`cache_repo.go`。
- `retention.go`：TTL / 保留期策略；
- `badger.go`：原子写、事务封装。

端口层 (`internal/ports`)：`ai.go`、`ai_cache.go`、`chat.go`、`conversation.go`、`fetcher.go`、`fred.go`、`messenger.go`、`news.go`、`price.go`、`repository.go`、`user.go`。

领域层 (`internal/domain`)：信号、COT、宏观、新闻、事件、价格（日/盘中）、用户、偏好、反馈、相关性、Carry 监测、汇率差、报告、风险、复合物、合约、AI 缓存。

---

## 六、可复用基础设施（`pkg/`）

| 包 | 作用 |
|----|------|
| `httpclient` | 带超时/重试的 HTTP 客户端 |
| `retry` | 指数退避重试 |
| `circuitbreaker` | 熔断器 |
| `saferun` | Panic-safe goroutine 运行器 |
| `logger` | zerolog 结构化 JSON 日志 |
| `metrics` | 指标（抓取成功率、信号准确率、AI 延迟等）|
| `validate` | 外部数据校验 |
| `timeutil` | WIB 时区与工具方法 |
| `fmtutil`、`format` | Telegram/数字/百分比格式化 |
| `mathutil` | 数学工具（Z-score、分位数等）|
| `errs` | 错误类型与封装 |

---

## 七、数据源清单

- **CFTC Socrata API**：COT 报告（周度，周五）。
- **MQL5 Economic Calendar**：实时经济事件与 Actual。
- **FRED（圣路易斯联储）**：收益率、劳动力、通胀等宏观序列。
- **BIS**：cbpol（央行政策利率）、creditgap、reer。
- **OECD**：综合领先指标。
- **ECB SDW**：欧央行货币政策数据。
- **Eurostat**：欧盟经济。
- **SNB**：瑞央行资产负债与外汇干预代理。
- **DTCC**：FX 互换机构流。
- **TradingEconomics**：全球宏观。
- **Treasury (US)**：国债拍卖。
- **SEC EDGAR**：13F 机构持仓。
- **WorldBank / IMF WEO**：全球指标。
- **CBOE / Deribit**：VIX、SKEW、DVOL、MOVE。
- **Bybit**：订单簿与微观结构（可选）。
- **价格：** TwelveData → AlphaVantage → Yahoo → CoinGecko → CryptoCompare 回退链。
- **链上：** CoinMetrics、区块链节点客户端。
- **DeFi：** DefiLlama。
- **行情抓取：** Finviz（通过 Firecrawl）。
- **AI：** Google Gemini、Anthropic Claude。

---

## 八、配置与运维

- `.env` 配置：`BOT_TOKEN`、`CHAT_ID` 必填；`GEMINI_API_KEY`、`CLAUDE_API_KEY`、`FRED_API_KEY`、`GEMINI_MODEL`、`DATA_DIR`、`AI_CACHE_TTL`、`AI_MAX_RPM`、`LOG_LEVEL` 等可选。
- 健康检查端点：`GET :8080/health`（`internal/health`）。
- Docker：`Dockerfile`、`docker-compose.yml` 一键部署。
- CI / Lint：`.golangci.yml`、`validate_syntax.sh`。
- 脚本 (`scripts/`)：
  - `feature_audit.sh`、`run_audit.sh`、`continuous_audit.sh` / `continuous_audit_v2.sh`、`rotating_audit.sh`、`sequential_audit.sh`、`verify_and_advance.sh`：持续自审；
  - `toggle_auto_commit.sh`：自动提交开关；
  - `backtest_chart.py`、`cta_chart.py`、`vpbt_chart.py`、`vpbt_engine.py`、`vp_engine.py`、`quant_engine.py`：回测/图表/量化引擎 Python 工具。
- 测试规范：`go test ./... -race -cover`；核心覆盖率 ≥80%，适配器 ≥60%。

---

## 九、核心设计原则（已落地）

- **数据完整性优先**：外部数据先校验后持久化；原子写事务；回退链。
- **信号质量优先**：Platt / Isotonic 校准；新策略先回测；按宏观状态独立追踪绩效。
- **防御式编程**：指数退避重试、TTL 缓存、超时熔断器。
- **可观测性**：结构化日志、关键指标、健康检查、拉取成功率监控。
- **模块化与可测性**：六边形架构、手工 DI、核心逻辑单测。
- **优雅降级**：AI / 数据源缺席时回退模板与次级源。

---

## 十、已实现模块一览（源码导航，相对 `Emulator/ark-intelligent/`）

```
cmd/bot                              # 程序入口
internal/adapter/telegram            # Telegram 适配器（命令、回调、格式化、键盘、深链）
internal/adapter/storage             # BadgerDB 各仓储 + 保留期
internal/config                      # 配置加载
internal/domain                      # 领域模型
internal/health                      # 健康检查
internal/ports                       # 接口（六边形端口）
internal/scheduler                   # 后台调度
internal/service/ai                  # Gemini/Claude + 缓存 + 限流 + 工具调用
internal/service/analysis            # 统一信号引擎
internal/service/backtest            # 回测、校准、成本、MC、审计
internal/service/bis                 # BIS 统计
internal/service/cot                 # COT 抓取/分析/信号/汇流/季节性/状态
internal/service/defi                # DeFi 健康
internal/service/dvol                # Deribit DVOL
internal/service/elliott             # 波浪理论
internal/service/factors             # 多因子库
internal/service/fed                 # FedWatch
internal/service/fred                # FRED + 宏观状态 + Carry + 利差
internal/service/gex                 # Gamma / IV Surface / Skew
internal/service/ict                 # ICT
internal/service/imf                 # IMF WEO
internal/service/intermarket         # 跨市场相关
internal/service/macro               # ECB/OECD/Eurostat/SNB/DTCC/TE/Treasury
internal/service/marketdata/*        # 多行情源客户端
internal/service/microstructure      # 市场微观结构
internal/service/news                # 日历/惊喜/Fed RSS/影响
internal/service/onchain             # CoinMetrics + 区块链
internal/service/orderflow           # Delta/吸收/POC
internal/service/price               # 价格聚合、GARCH、Hurst、HMM、相关/背离、EIA
internal/service/quantbt             # 量化回测引擎
internal/service/regime              # 状态叠加引擎
internal/service/sec                 # 13F
internal/service/sentiment           # 情绪（CBOE/MyFxBook/OpenInsider/DVOL 桥接）
internal/service/strategy            # 策略引擎 + Risk Parity
internal/service/ta                  # 经典指标 + AMT + 回测 + Ichimoku/Fib/...
internal/service/treasury            # US Treasury
internal/service/vix                 # VIX / VolSuite / SKEW-VIX 告警 / MOVE
internal/service/vpbt                # 成交量轮廓回测
internal/service/worldbank           # World Bank
internal/service/wyckoff             # Wyckoff
pkg/*                                # 复用基础设施
scripts/*                            # 运维与量化脚本
```

---

## 十一、测试与审计

- 单元测试覆盖：分析器、信号生成、校准、因子、调度器、格式化、错误路径、回调覆盖。
- 审计测试：`audit.go`、`audit_deep_test.go`、`audit_pass4_test.go`（backtest、price、ta、cot、fred 等多处）。
- 连续自审脚本：`continuous_audit*.sh`、`feature_audit.sh`、`rotating_audit.sh`、`sequential_audit.sh`、`verify_and_advance.sh`。
- 语法验证：`validate_syntax.sh`。

---

> **结论**：ARK Intelligent 已覆盖**数据采集 → 领域建模 → 多策略/多因子分析 → 回测与校准 → AI 叙事 → Telegram 交互 → 告警广播 → 运维与自审**的完整闭环，形成一个可独立部署、可观测、可扩展的机构级宏观智能 Bot 平台。
