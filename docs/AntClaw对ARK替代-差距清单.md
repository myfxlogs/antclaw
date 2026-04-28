# AntClaw 对 ARK Intelligent 的全面替代 · 差距清单

> 目标：AntClaw 100% 替代 `Emulator/ark-intelligent`，业务能力**零回退**。
> 范围：业务能力 + 数据源接入 + 出口协议合规性。
> 出口协议硬约束：**所有对外接口必须使用 Connect-RPC / gRPC / SSE**，禁用轮询、WebSocket、REST。
> 入口（第三方 API 抓取）不受此约束。
> 数据源粒度方案：**统一抽象（datasource + apiclient + Job）为主干 + 单源薄 client（infra/apiclient/<vendor>）为局部增强**；业务派生能力归属 `service/<capability>`。Telegram Bot 适配层不再实现，由 Connect-RPC + 前端模块承载等价能力。

---

## 0. 出口协议合规性差距（必须先治理）

| 出口路径 | 现状 | 是否合规 | 处置 |
|---|---|---|---|
| `/antclaw.v1.*` Connect 路径 | Connect-RPC | ✅ | 维持 |
| `/sse/jobs`、`/sse/audit` | SSE | ✅ | 维持 |
| `/health` | 普通 REST GET | ❌ | 迁入 `SystemService.Healthz`（Connect Unary）。容器健康探针改为 `grpc_health_probe` 或对 Connect 端点的 `POST` |
| `/`（根路径回显） | 普通 REST GET | ❌ | 删除；如保留欢迎页，则定为静态资源由 admin 前端容器托管，不在 API 容器暴露 |
| `frontend/admin/src/lib/crypto.ts` 中的 `fetch()` | 自定义 REST POST envelope | ❌ | envelope 通过 Connect RPC（如 `CryptoService.PostEnvelope`）投递；保留 RSA-OAEP+AES-GCM+HMAC 加密语义 |
| `fetch('${API_BASE_URL}/health')` 健康探活 | REST | ❌ | 换为 Connect 客户端调用 `SystemService.Healthz` |

**验收标准**：`grep -rE "fetch\\(|HandleFunc|gorilla/websocket" frontend backend/cmd backend/internal` 仅命中 Connect/SSE handler 与 connect-web 内部实现。

---

## 1. 业务能力差距矩阵

> 状态定义：✅ 已具备且对齐；🟡 部分实现需补强；❌ 缺失。
> 「Proto」列若标 *新增*，则需新建 `proto/antclaw/v1/<file>.proto`。

### 1.1 信号与策略层

| 能力（ARK 来源） | AntClaw 现状 | 状态 | Proto Service.Method | SSE 通道 | 前端模块 |
|---|---|---|---|---|---|
| Unified Signal Engine（`service/analysis`） | `service/signals/unified_compute.go` | 🟡 需对齐 ARK 的多因子合流（Platt scaling/isotonic 校准） | `SignalsService.ComputeUnified` *扩展* | `signals.unified` | `features/signals` |
| 多因子置信度校准（Platt / isotonic） | 缺失 | ❌ | `SignalsService.CalibrateConfidence` *新增* | — | `features/signals/calibration` |
| Regime Overlay Engine（`service/regime/overlay_engine`） | `service/signals/transition.go` 仅迁移矩阵 | 🟡 缺 overlay 决策 | `SignalsService.GetRegimeOverlay` *新增* | `signals.regime` | `features/signals/regime` |
| 策略合流引擎（`service/strategy`） | `service/strategy/*` | ✅ | 维持 | — | `features/strategy` |
| Risk Parity Sizing | `service/strategy/risk_parity.go`、`backtest/sizing` | ✅ | 维持 | — | `features/strategy/sizing` |

### 1.2 回测与评估

| 能力 | 现状 | 状态 | Proto | 备注 |
|---|---|---|---|---|
| 历史回测（baseline/costs） | `service/backtest/*` | ✅ | `BacktestService.*` | 已有 |
| **Walk-Forward 回测**（D 类） | 缺失 | ❌ | `BacktestService.RunWalkForward` *新增* | 滚动窗口训练/验证 |
| 性能指标（Sharpe/Sortino/MaxDD/HitRate） | `service/backtest/baseline.go` | 🟡 需补 Sortino、Calmar、PerRegime | 扩展现有 | — |
| Bootstrap 显著性检验（`service/backtest/bootstrap`） | 缺失 | ❌ | `BacktestService.BootstrapSignificance` *新增* | — |
| 成本模型（`service/backtest/costs`） | 缺失 | ❌ | 现有 RPC 加 `CostModel` 入参 | 滑点/手续费/借贷利率 |
| 信号 PnL 归因（`service/backtest/audit`） | `service/backtest/attribution.go` | 🟡 需对齐 ARK 字段集 | 扩展现有 | — |
| Daily Trend Filter | 缺失 | ❌ | `BacktestService.ApplyTrendFilter` *新增* | — |

### 1.3 COT / 持仓

| 能力 | 现状 | 状态 |
|---|---|---|
| COT 抓取/分析/置信度 | `service/cot/*` | ✅ |
| **Category Z-Score**（`service/cot/category_zscore`） | 缺失 | ❌ |
| **Confluence Score**（`service/cot/confluence_score`） | `service/signals/cot_mapping.go` 部分 | 🟡 需独立模块 |

### 1.4 经济日历 / 新闻 / 影响

| 能力 | 现状 | 状态 |
|---|---|---|
| 日历抓取（MQL5） | `service/calendar/*` | ✅ |
| 实际值/惊喜评分 | `service/calendar` 部分 | 🟡 ARK 的 `news/surprise.go` 历史影响权重模型缺失 |
| **Fed RSS**（`service/news/fed_rss.go`） | 缺失 | ❌ |
| **影响自举**（`service/news/impact_bootstrap.go`） | 缺失 | ❌ |
| **影响记录器**（`impact_recorder.go`） | 缺失 | ❌ |
| 日历广播 SSE | `stream:jobs_events` 复用 | 🟡 需独立 `stream:calendar_events` |

新增 Proto：扩展 `CalendarService.GetSurpriseHistory / GetImpactScores`。

### 1.5 宏观（FRED + 多源）

| 能力 | ARK 包 | 现状 | 状态 |
|---|---|---|---|
| FRED 系列抓取 | `service/fred/fetcher` | `service/macro` | ✅ |
| **FRED 告警**（regime change） | `fred/alerts.go` | 缺失 | ❌ |
| **FRED Carry Monitor** | `fred/carry_monitor.go` | 缺失 | ❌ |
| **FRED 复合指标** | `fred/composites.go` | 缺失 | ❌ |
| **DTCC 客户端** | `macro/dtcc_client.go` | 缺失 | ❌ |
| **ECB 客户端** | `macro/ecb_client.go` | 缺失 | ❌ |
| **Eurostat 客户端** | `macro/eurostat_client.go` | 缺失 | ❌ |
| **OECD 客户端** | `macro/oecd_client.go` | 缺失 | ❌ |
| **SNB 客户端** | `macro/snb_client.go` | 缺失 | ❌ |
| **TradingEconomics 客户端** | `macro/tradingeconomics_client.go` | 缺失 | ❌ |
| **US Treasury 客户端** | `macro/treasury_client.go` + `service/treasury/analyzer.go` | 缺失 | ❌ |
| **BIS（CBPOL/Credit Gap/REER）** | `service/bis/*` | 缺失 | ❌ |
| **IMF WEO** | `service/imf/weo.go` | 缺失 | ❌ |
| **World Bank** | `service/worldbank/client.go` | 缺失 | ❌ |
| **FedWatch（CME 概率）** | `service/fed/fedwatch.go` | 缺失 | ❌ |

新增 Proto：
- `MacroService.GetFREDAlerts / GetCarryMonitor / GetComposites`（扩展现有）
- `MacroExtrasService.GetBISIndicators / GetIMFWeo / GetWorldBankSeries / GetECBSeries / GetEurostatSeries / GetOECDSeries / GetSNBSeries / GetTradingEconomicsSeries / GetDTCCData` *新建一组或拆分*
- `TreasuryService.GetCurve / GetAnalysis` *新建*
- `FedWatchService.GetFOMCProbabilities` *新建*

### 1.6 价格 / 市场数据

| 能力 | 现状 | 状态 |
|---|---|---|
| 多源价格聚合 + 降级链 | `service/price` | 🟡 当前降级链硬编码，需读 datasource 优先级 |
| 4H/Daily 多周期 | `service/price` | 🟡 需补 `daily_context` / `intraday_sync` |
| 相关性矩阵（`price/correlation.go`） | 缺失 | ❌ |
| 季节性模式 | 缺失 | ❌ |

新增 Proto：`PriceService.GetCorrelationMatrix / GetSeasonality` *扩展*。

### 1.7 波动率 / 期权（D 类核心缺口）

| 能力（ARK） | 现状 | 状态 |
|---|---|---|
| **GEX 计算器**（`gex/calculator.go`） | 缺失 | ❌ |
| **GEX Engine**（场景汇总） | 缺失 | ❌ |
| **IV Surface**（`gex/iv_surface.go`） | 缺失 | ❌ |
| **Skew/IV-Skew Alert**（`gex/skew.go` + `vix/skew_vix_alert.go`） | 缺失 | ❌ |
| VIX/MOVE 抓取 | `service/vol/service.go` 部分 | 🟡 需补 MOVE、cross_vol、term structure |
| Deribit DVOL | 缺失（仅有 datasource seed） | ❌ |
| 波动率套件（`vix/vol_suite`） | 缺失 | ❌ |

新增 Proto：`OptionsService.GetGEX / GetIVSurface / GetSkew / GetIVAlerts` *新建*；`VolService.GetMOVE / GetCrossVol / GetTermStructure` *扩展*。

### 1.8 链上 / DeFi

| 能力 | 现状 | 状态 |
|---|---|---|
| 链上抓取（CoinGecko） | `worker/collector_onchain.go` | 🟡 records=0，需修复 + 改为 `service/onchain` |
| 链上分析器（`onchain/analyzer.go`） | 缺失 | ❌ |
| Coinmetrics 客户端 | 缺失 | ❌ |
| Blockchain.com 客户端 | 缺失 | ❌ |
| **DeFi 分析**（`service/defi/analyzer.go` + Defillama TVL） | 缺失 | ❌ |

新增 Proto：`OnchainService.GetMetrics / GetAnalysis` *新建*；`DeFiService.GetTVL / GetProtocolStats / GetAnalysis` *新建*。

### 1.9 情绪 / 资金流

| 能力 | ARK | 现状 | 状态 |
|---|---|---|---|
| Fear & Greed | — | `worker/collector_sentiment.go` | ✅ |
| **CBOE Put/Call**（`sentiment/cboe.go`） | 缺失 | ❌ |
| **MyFXBook 散户持仓**（`sentiment/myfxbook.go`） | 缺失 | ❌ |
| **OpenInsider 内幕交易**（`sentiment/openinsider.go`） | 缺失 | ❌ |
| **DVOL 整合**（`sentiment/dvol_integration.go`） | 缺失 | ❌ |
| **CryptoCompare social/sentiment** | 缺失 | ❌ |
| **Finviz 多空比** | 缺失 | ❌ |

新增 Proto：`SentimentService.GetCBOEPutCall / GetMyFXBookPositions / GetInsiderTrades / GetCryptoSocial / GetFinvizMetrics` *扩展*。

### 1.10 SEC / 监管

| 能力 | ARK | 现状 |
|---|---|---|
| **SEC EDGAR 客户端 + 解析 + 分析**（`service/sec/*`） | 缺失 | ❌ |

新增 Proto：`SECService.ListFilings / GetFiling / GetAnalysis` *新建*。

### 1.11 技术分析 / 价格行为

| 能力 | ARK | 现状 | 状态 |
|---|---|---|---|
| RSI/MACD/BB/ATR/Fib/Ichimoku/Supertrend | 通用 TA | `service/ta/service.go` | 🟡 需对齐到完整指标列表 |
| **AMT Close**（开收口） | `ta/amt_close.go` | 缺失 | ❌ |
| **AMT Day Type** | `ta/amt_daytype.go` | 缺失 | ❌ |
| **AMT Migration** | `ta/amt_migration.go` | 缺失 | ❌ |
| **AMT Opening** | `ta/amt_opening.go` | 缺失 | ❌ |
| **AMT Rotation** | `ta/amt_rotation.go` | 缺失 | ❌ |
| Volume Profile（`vpbt/engine.go`） | `worker/transition_matrix.go`+volume placeholder | 🟡 需独立 `service/volume_profile` |
| Wyckoff 完整四件套（classifier/phase/events/summary） | `service/wyckoff/engine.go` 较薄 | 🟡 行数 272 vs ARK 916，需补 phase 判别、event 检测 |
| ICT（FVG/Liquidity/OrderBlock/Structure/Swing） | `service/ict/engine.go` 341 行 | 🟡 vs ARK 622 行，需补 OrderBlock 完整逻辑、Liquidity sweep |
| Elliott（zigzag/projector/validator） | `service/elliott/engine.go` 180 行 | 🟡 vs ARK 多文件，需补 zigzag 与 projector |
| Microstructure | `service/microstructure/engine.go` 124 行 | 🟡 vs ARK 完整测试与多策略 |
| Orderflow（Delta/POC/Absorption） | `service/orderflow/engine.go` 204 行 | 🟡 vs ARK 4 文件，需补 Absorption / POC |

新增/扩展 Proto：
- `TAService.GetAMT*` *扩展*
- `MarketProfileService.GetVolumeProfile` *新建* 或并入 `MicrostructureService`

### 1.12 AI / 解读

| 能力 | ARK | 现状 | 状态 |
|---|---|---|---|
| Gemini / Claude 客户端 | `service/ai` | `service/ai`、`service/systemai` | ✅ |
| **Cached Interpreter** | `ai/cached_interpreter.go` | `service/ai/summarizer.go` 部分 | 🟡 需 TTL 缓存 + 失效策略 |
| **Chat Service Blocks** | `ai/chat_service.go` | 缺失 | ❌ |
| **Context Builder** | `ai/context_builder.go` | 缺失 | ❌ |
| **AI 限流** | `ai/ai_ratelimit.go` | 部分 | 🟡 |
| **Prompt 模板与降级** | `ai/interpreter.go` | 部分 | 🟡 |

新增 Proto：`AIService.Chat / GetInterpretation / BuildContext` *扩展*。

### 1.13 间市场 / 因子

| 能力 | ARK | 现状 | 状态 |
|---|---|---|---|
| Intermarket 引擎 | `intermarket/engine.go` | `service/intermarket/engine.go` 214 行 | ✅ 需对齐字段 |
| 因子（momentum/carry/lowvol/crowding/flow_divergence/residual_reversal） | `factors/*` | `service/factors/*` 主体齐 | ✅ |
| **Carry Adjusted**（`carry_adjusted.go`） | `factors/carry.go` 5 行占位 | 🟡 |
| **Profile Builder**（`factors/profile_builder.go`） | 缺失 | ❌ |

---

## 2. 数据源（Vendor 薄 client）差距

> 落位：每个数据源一个独立子包于 `backend/internal/infra/apiclient/<vendor>/`，仅包含协议层（鉴权、限流、字段映射、错误码）。业务派生能力住对应 `service/<capability>/`。

| Vendor | ARK 路径 | AntClaw 现状 | 状态 |
|---|---|---|---|
| Bybit | `marketdata/bybit/client.go` | datasource seed only | ❌ 需建 `apiclient/bybit/` |
| Deribit（含 DVOL） | `marketdata/deribit/{client,dvol,types}.go` | datasource seed only | ❌ 需建 `apiclient/deribit/` |
| CoinGecko | `marketdata/coingecko/client.go` | `worker/collector_onchain.go` 直写 HTTP | 🟡 需抽出到 `apiclient/coingecko/` |
| CryptoCompare | `marketdata/cryptocompare/{client,analyzer,models}.go` | datasource seed only | ❌ |
| Defillama | `marketdata/defillama/client.go` | datasource seed only | ❌ |
| Finviz | `marketdata/finviz/client.go` | datasource seed only | ❌ |
| TwelveData / AlphaVantage / Yahoo | `service/price/aggregator.go` 内嵌 | `service/price` 部分 | 🟡 需拆为独立 `apiclient/twelvedata`、`apiclient/alphavantage`、`apiclient/yahoo` |
| CFTC Socrata | `service/cot/fetcher.go` | `worker/collector_cot.go` | 🟡 需抽出 |
| MQL5 | `service/news/fetcher.go` | `infra/apiclient/mql5.go` | ✅ |
| FRED | `service/fred/fetcher.go` | `infra/apiclient/fred.go` | ✅ |
| Coinmetrics | `onchain/coinmetrics.go` | 缺失 | ❌ |
| Blockchain.com | `onchain/blockchain_client.go` | 缺失 | ❌ |
| SEC EDGAR | `sec/client.go` | 缺失 | ❌ |
| US Treasury | `macro/treasury_client.go` | 缺失 | ❌ |
| ECB / Eurostat / OECD / SNB / IMF / World Bank / BIS / DTCC / TradingEconomics | `macro/*_client.go`、`bis/*`、`imf/weo.go`、`worldbank/client.go` | 缺失 | ❌ |

**横切要求**（全部 vendor 子包共同遵守）

- 实现统一 `apiclient.Source` 接口（暂未定义，需新增）：限流、重试、断路器、超时、密钥从 `datasource.Resolver` 取
- 统一指标埋点：`requests_total{vendor,endpoint,status}` / `latency_seconds`
- 统一错误码映射到 `errs.External`
- 单包行数硬上限 800 行，否则按端点拆分

---

## 3. 后台 / 调度 / 可观测性差距

| 能力 | ARK | 现状 | 状态 |
|---|---|---|---|
| Job 调度与启停 | `internal/scheduler` | `worker/main.go` + Redis | ✅ |
| Job 手动触发 | scheduler 内 | Redis Pub/Sub `jobs:trigger` | ✅ |
| Job 启动播种 | — | 已实现 | ✅ |
| Job SSE 实时推送 | — | `/sse/jobs` | ✅ |
| **Walk-Forward 任务** | — | 缺失 | ❌ |
| **影响自举（impact bootstrap）任务** | `news/scheduler.go` | 缺失 | ❌ |
| **告警评估器**（FRED regime/skew alert） | `fred/alerts.go` | `worker/alert_evaluator.go` 部分 | 🟡 需扩 FRED/Skew/CrossVol 告警 |
| 健康检查 | `:8080/health` REST | 同名 REST | ❌ 见 §0 |
| 指标暴露 | 项目内 metrics 包 | 部分 | 🟡 需统一 Prometheus 端点 + 通过 Connect 健康 RPC 输出业务指标摘要 |

---

## 4. 前端模块差距（Admin + Web）

> 现状盘点见 `docs/AntClaw-前端架构.md`；本表只列对应"业务能力差距"应新增/补强的前端模块。

| 业务能力 | 前端模块（建议路径） | 状态 |
|---|---|---|
| GEX / IV Surface / Skew | `features/options` | ❌ 新建 |
| VIX/MOVE/CrossVol/TermStructure | `features/vol` | 🟡 扩展 |
| Walk-Forward 回测 | `features/backtest/walkforward` | ❌ 新建 |
| 链上分析 | `features/onchain` | ❌ 新建 |
| DeFi TVL | `features/defi` | ❌ 新建 |
| AMT 日内类型 | `features/ta/amt` | ❌ 新建 |
| Volume Profile | `features/microstructure/vp` | ❌ 新建 |
| Regime Overlay | `features/signals/regime` | ❌ 新建 |
| FedWatch 概率 | `features/macro/fedwatch` | ❌ 新建 |
| 宏观多源（BIS/IMF/WB/ECB/Eurostat/OECD/SNB/TE/Treasury） | `features/macro/extras` | ❌ 新建 |
| SEC EDGAR | `features/sec` | ❌ 新建 |
| 情绪面板（CBOE/MyFXBook/Insider/DVOL/Social） | `features/sentiment` | 🟡 扩展 |
| AI 解读对话 | `features/ai/chat` | 🟡 扩展 |

**强约束**：所有前端模块只能通过 `@connectrpc/connect-web` 调用 RPC，或 `EventSource` 订阅 SSE；**禁止 `fetch/axios/WebSocket/setInterval`**（lint 规则需固化）。

---

## 5. 优先级与里程碑（建议）

> 一切按"先打通端到端，再加深"的节奏。每个里程碑均要求：Proto → 后端服务 → Connect handler → 前端模块 → SSE（如适用）→ 单元/集成测试 → 文档（中文，存于 `docs/`）。

**M1 · 协议合规与基线（1 周）**
- §0 全部清理（去 REST、去 fetch）
- 修 onchain records=0
- 抽出 CoinGecko、CFTC 到 `apiclient/<vendor>`
- 引入 `apiclient.Source` 接口与中间件链

**M2 · D 类闭环（2~3 周）**
- GEX/IV Surface/Skew（含 Deribit 薄 client）
- Walk-Forward 回测 + Bootstrap 显著性
- Wyckoff/AMT/Volume Profile 补齐
- VIX/MOVE/CrossVol/TermStructure 扩展

**M3 · 宏观全谱（2~3 周）**
- BIS / IMF / World Bank / ECB / Eurostat / OECD / SNB / TE / DTCC / Treasury 薄 client
- FedWatch、FRED Alerts/Carry/Composites
- Treasury 分析器
- 影响自举任务（impact bootstrap）

**M4 · 链上 / DeFi / SEC / 情绪扩展（2 周）**
- Coinmetrics、Blockchain.com 薄 client + Onchain 分析
- Defillama TVL + DeFi 分析
- SEC EDGAR 全链路
- CBOE/MyFXBook/OpenInsider/CryptoCompare social/Finviz 情绪扩展

**M5 · AI 与编排（1~2 周）**
- Cached Interpreter / Context Builder / Chat Blocks
- 限流与降级模板对齐
- 全链路告警体系（FRED regime / Skew / CrossVol）

**M6 · 校验与零回退验收（1 周）**
- 用 ARK 端到端命令清单逐项 mapping，确认 AntClaw 已有等价 RPC
- 跑 ARK 测试用例的语义对照（hit rate ≥ ARK 历史值）
- 出 `AntClaw-100替代验收报告.md`

---

## 6. 验收门槛

- **协议**：仓库内仅存 Connect-RPC handler 与 SSE handler；前端无 `fetch/axios/WebSocket/setInterval`。
- **能力**：本文档中所有 ❌/🟡 项变 ✅，每项有：
  - Proto 定义提交并生成
  - 后端服务实现 + 单元测试 ≥80% 覆盖
  - Connect handler 注册到 `cmd/antclaw-api/main.go`
  - 前端模块可访问，状态/错误友好显示
  - 中文文档存于 `docs/`
- **数据**：`/datasources` 列出全部 vendor，密钥可热配；每条 vendor 健康度可观测。
- **回测**：Walk-Forward + Bootstrap 报告可在 UI 直接生成。

---

## 7. 风险与未决项

- **第三方 API 配额与稳定性**：BIS/SNB/DTCC/Eurostat 速率较低且字段易变，需要在 `apiclient/<vendor>` 内置缓存/快照表
- **AMT 与 Wyckoff 的"语义复刻"成本**：ARK 测试样本可作为基线；但 AntClaw 的数据存储模型不同（PG vs Badger），需要重新对齐数据形态
- **AI Provider 凭据**：BYOK 与系统密钥的优先级策略需在 `systemai` 与 `byok` 之间正式约定
- **Telegram 历史命令的 UI 翻译**：每个 `/cmd` 在前端都需要等价模块；交互（inline keyboard 等）需重新设计为 Connect 调用 + SSE 状态
- **/health 与 crypto envelope 的迁移**：会同时影响容器健康探针与现有登录加密流，需要安排灰度切换
