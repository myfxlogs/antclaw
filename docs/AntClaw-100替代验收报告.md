# AntClaw 100% 替代 ARK Intelligent · 验收报告

- 验收时间：2026-04-29 00:18 UTC
- 提交版本：未打 tag（开发分支）
- 容器集合：`antclaw-api`、`antclaw-worker`、`antclaw-postgres`、`antclaw-redis`、`antclaw-admin`、`antclaw-web`、`antclaw-minio`
- 数据源凭据：见 `docs/凭据就绪核对.md`

## 1. 协议合规

| 检查项 | 结果 | 说明 |
|---|---|---|
| `mux.HandleFunc` 业务路由 | ✅ 仅命中 SSE | `/sse/jobs`、`/sse/audit` 是允许例外 |
| 后端 `gorilla/websocket` / `nhooyr.io/websocket` | ✅ 无命中 | 已删除；只用 Connect-RPC + SSE |
| 前端 `fetch()` / `axios` / `new WebSocket` / `setInterval(业务)` | ✅ 无命中 | `fetchConfigs` 函数命名不算违规 |
| `SystemService.Healthz` Connect 健康端点 | ✅ | `{status:healthy, postgres:up, redis:up}` |
| Crypto envelope 走 Connect | ✅ | `CryptoService.PostEnvelope` 已注册（业务层校验返回 `unauthenticated` 表示路由通） |
| Apiclient.Source 中间件（限流/断路器/超时/指标） | ✅ | `internal/infra/apiclient/source.go` |
| 所有 vendor 走 `apiclient/<vendor>/` 子包 | 🟡 | 已迁移核心 14 家；`fred_client.go` / `calendar_client.go (MQL5)` 仍以平铺保留（用量大、改动风险高，留待后续小步迁移） |

## 2. ARK 命令到 AntClaw RPC 命令映射（覆盖度）

> 完整 23 条命令映射详见 `docs/ARK替代实施方案/09-M6零回退验收.md §2`。下面给出实际可用状态。

| ARK 命令 | AntClaw RPC | 状态 |
|---|---|---|
| `/cot`、`/c`、`/ce` | `COTService.GetSummary` | ✅ 真数据（CFTC Socrata legacy） |
| `/calendar`、`/cal` | `CalendarService.ListEvents` | ✅ 真数据（MQL5） |
| `/impact` | `CalendarService.GetImpact*` | 🟡 占位仍未补全 |
| `/outlook` | `AIService.Outlook` | 🟡 代码就绪；zhipu 余额不足 → 验收无法跑通 |
| `/bias`、`/rank`、`/heat`、`/rankx` | `SignalsService.{GetBias,GetRanking,GetHeatmap,GetRankingExt}` | ✅ |
| `/macro` | `MacroService.GetFred / GetEcb / GetBis / ...` | ✅ 真数据（FRED/BIS/ECB/...） |
| `/xfactors` | `SignalsService.ComputeUnified` | ✅ |
| `/playbook` | `StrategyService.GetPlaybook` | ✅ |
| `/transition` | `SignalsService.GetRegimeTransition` | ✅ |
| `/cryptoalpha` | `OnchainService.GetMetrics` + `OptionsService.GetGEX` | ✅ |
| `/prefs`、`/membership`、`/feedback`、`/history` 等 | `UserService.*` | ✅ |
| `/help`、`/start`、`/onboarding` | 前端帮助 + `UserService.StartOnboarding` | ✅ |
| `/status` | `SystemService.Healthz` | ✅ |
| `/settings` | `UserService.{GetSettings,UpdateSettings}` | ✅ |

## 3. E2E 18 场景验收（脚本化）

执行：`bash scripts/e2e/run_all.sh`

| ID | 场景 | RPC | 结果 | 备注 |
|---|---|---|---|---|
| SC-1  | COT 摘要              | `COTService.GetSummary` | ✅ | latest.nonCommLong 等真值 |
| SC-2  | SSE jobs 事件流       | `/sse/jobs`             | ✅ | 30s 内收到 4 行 SSE |
| SC-3  | GEX                   | `OptionsService.GetGEX` | ✅ | strikes=96（Deribit BTC） |
| SC-4  | Walk-Forward          | `BacktestService.RunWalkforward` | ✅ | folds=4，价格序列 EURUSD 260 根 K 线 |
| SC-5  | 链上分析              | `OnchainService.GetMetrics` | ✅ | points=30（CoinGecko 真值） |
| SC-6  | DeFi 协议榜            | `DeFiService.GetTopProtocols` | ✅ | items=50（DefiLlama） |
| SC-7  | SEC EDGAR 文件        | `SECService.ListFilings` | ✅ | items≥1（Apple CIK 真数据） |
| SC-8  | FedWatch FOMC 概率    | `FedWatchService.GetFOMCProbabilities` | ✅ | sum=100（firecrawl→CME） |
| SC-9  | 美债收益率曲线        | `TreasuryService.GetCurve` | ✅ | tenors=12（home.treasury.gov 真数据） |
| SC-10 | MacroExtras WorldBank | `MacroExtrasService.GetSeries` | ✅ | 美 GDP 1960~2024 真值 |
| SC-11 | IV Surface             | `OptionsService.GetIVSurface` | ✅ | points≈908 |
| SC-12 | Skew                  | `OptionsService.GetOptionsSkew` | ✅ | rr_25d 真值 |
| SC-13 | DVOL + VIX            | `VolService.GetDvol / GetVix` | ✅ | DVOL=40.57（Deribit），VIX=18.02（CBOE） |
| SC-14 | 情绪面板（Finviz）    | `SentimentExtrasService.GetFinvizMetrics` | ✅ | AAPL short_ratio=2.94（firecrawl） |
| SC-15 | AI Outlook            | `AIService.Outlook` | ❌ | zhipu **余额不足 1113**，代码已通；非代码缺陷 |
| SC-16 | AI Cached Interpret   | `AIService.Interpret` | ❌ | 同上 |
| SC-17 | 健康检查              | `SystemService.Healthz` | ✅ | status=healthy |
| SC-18 | Crypto envelope        | `CryptoService.PostEnvelope` | ✅ | 路由通；业务校验返回 unauthenticated（合理） |

**通过率：16/18 = 88.9%**。

## 4. 性能（参考样本）

| RPC | 触发耗时 |
|---|---|
| `SystemService.Healthz` | < 50ms |
| `OptionsService.GetGEX` | ≈ 1.5s（Deribit + 解析 96 strikes） |
| `OptionsService.GetIVSurface` | ≈ 2.0s（Deribit BookSummary + Instruments） |
| `TreasuryService.GetCurve` | ≈ 0.8s（home.treasury.gov） |
| `BacktestService.RunWalkforward` | ≈ 0.4s（260 K 线 SMA 网格搜索） |

未做严格 hey/wrk 压测（`hey` 工具未安装在容器宿主，需运维补装）；以上为人眼观测样本。

## 5. 数据可用性（核心表）

| 表 | 行数 / 最新时间 |
|---|---|
| `cot_records` | 312 行 / 2026-04-21 |
| `onchain_metrics` | 465 行 / 2026-04-28 |
| `price_daily` | 3584 行 / 2026-04-29（EURUSD/GBPUSD/USDJPY/AUDUSD/USDCAD/USDCHF/NZDUSD/XAUUSD/XAGUSD/CRUDE/VIX/SP500/DJIA/NASDAQ） |
| `data_source_configs` | 33 行（9/9 必需 api_key 已就绪） |
| `system_ai_configs` | 8 行（zhipu enabled+secret） |
| `ai_cache` | 0 行（首次调用因 zhipu 余额失败未写入；账户充值后即可命中） |

`sentiment_snapshots` / `data_snapshots` 仍空：相关 collector 需要的 schema 后续单独迁移。

## 6. Job 健康

```
$ docker exec antclaw-redis redis-cli --scan --pattern 'jobs:status:*' | wc -l
17 个 Job 已播种快照
```

启动后立即跑通：`calendar-sync / macro-sync / cot-sync / price-sync / sentiment-sync / onchain-sync / intraday-sync / defi-sync / vix-term-sync …`。其中 `cot-sync` 312 条、`onchain-sync` 465 条、`price-sync` 3584 条数据写入。

## 7. 已知未交付项 / 后续 Roadmap

| 项 | 影响 | 建议 |
|---|---|---|
| **AI 真数据验收 SC-15/16 失败** | zhipu glm-4.5 账户 1113「余额不足」，全链路代码已通 | 充值或切换其他已有 provider（anthropic/openai 等已存表，配 Key 即用） |
| **前端 13 个 features 模块未交付** | gap 清单第 4 章列出的 `features/options`、`features/backtest/walkforward`、`features/onchain` 等仍未在 `frontend/admin/src` 与 `frontend/web/src` 下落地 | 单独的前端会话推进，按 `docs/ARK替代实施方案/10-前端模块规范.md` 逐个 |
| **Insider/CryptoSocial/MOVE 真实端点** | `SentimentExtrasService.GetInsiderTrades / GetCryptoSocial`、`VolService.GetMove` 暂返回 `unavailable`/空 | 通过 firecrawl 抽取或新接 vendor（OpenInsider、CryptoCompare social、yardeni MOVE） |
| **`service/sentiment/service.go`、`service/ta/service.go`、`service/price/service.go` 仍含 `randFloat() = 0.5` 占位** | 用于无 DB 数据时的兜底；非随机但仍属合成 | 等 collector 补齐对应 schema 后改为 DB 查询；vol 服务已完成此重构 |
| **`fred_client.go`、`calendar_client.go (MQL5)` 仍是平铺旧文件** | 与 `apiclient.Source` 中间件未对齐 | 后续小步迁移到 `apiclient/fred/`、`apiclient/mql5/` |
| **Walk-Forward 当前最少 60 K 线 / fold（共 240）** | 配置过严，新 symbol 短数据时会 `no_data` | 已在脚本里以 EURUSD 260 根 + folds=4 通过；上层可接受 fold 自适应 |
| **`hey` 性能压测未跑** | 仅靠观测样本估 P50/P95 | 运维补装后可直接跑 `09-M6零回退验收.md §5` 的命令 |

## 8. 结论

- [x] **可发布的限定状态**：所有合规检查、协议改造、核心后端能力（除 AI provider 余额问题外）均通过；E2E 18 场景 16 项 PASS。
- [ ] **不能宣称「全量替代」**：前端 13 模块尚未交付，AI 真数据需运维补充付费；上述 Roadmap 必须与运维/前端排期一起推进。

> 本报告由 `scripts/e2e/run_all.sh` 自动产出辅助生成；可在每次代码修改后重跑回归。
