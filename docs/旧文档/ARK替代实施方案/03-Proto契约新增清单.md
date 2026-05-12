# 03 · Proto 契约新增清单

> 本文给出所有新增 / 扩展的 `.proto` 文件骨架，作为生成 Go/TS 桩的依据。
> 每个 Service 一个 proto 文件，方法签名稳定后即可执行 `buf generate`。
> 所有 RPC 默认 Unary；`Stream` 通道由 SSE handler 承载，proto 中**不**使用 `stream` 关键字。

---

## 1. 命名与版本

- 所有新文件位于 `proto/antclaw/v1/<service>.proto`
- `package antclaw.v1`；`option go_package = "github.com/antclaw/antclaw/gen/go/antclaw/v1;antclawv1";`
- 字段命名 `snake_case`；枚举 `SCREAMING_SNAKE_CASE`
- 时间字段统一 `google.protobuf.Timestamp`
- 数值字段单位写在字段注释中

### 1.1 Proto 与源码的 800 行上限

按总章程，**Proto 定义与源代码均不得超过 800 行**。

> 修订（M1 收尾决策）：800 行硬上限**仅作用于源代码（含 `.proto`）**，不包含 buf 生成的机械产物（`gen/` 目录下的 `*.pb.go` / `*.connect.go` / `*_pb.ts` / `*_connect.ts`）。理由：生成产物是 proto 编译器机械输出，体积取决于消息数量与字段位数，无可读性/可维护性概念；强制压缩生成产物会把 proto 切分成不必要的碎片文件，损害契约清晰度。

经验阈值（供 proto 拆分决策）：

| 单 proto 文件 | message + RPC 体量 | 处置 |
|---|---|---|
| 较少 | < 200 行 | ✅ 保持 |
| 中等 | 200–500 行 | � 评估子能力切分 |
| 偏多 | > 500 行 | 🟡 建议按子能力拆分 |
| 单文件超 800 行 | — | ❌ 必须拆分（lint 阻断）|

**拆分策略（保持向后兼容的前提下）**：

1. **按子能力拆 proto**：例如 `macro.proto` → `macro_dashboard.proto` + `macro_alerts.proto` + `macro_carry.proto`，Service 可以保持不变或拆分为多个 Service
2. **独立 Service**：例如把 `signals.proto` 里的 Regime/Calibration 新增 RPC 拆出到新的 `signals_regime.proto`
3. **共享 message 下沉到 `common.proto`**：避免在多个 proto 中重复定义

### 1.2 存量治理结论（M1 收尾）

经现状审计：**全部源 `.proto` 文件 ≤ 285 行（最大 `macro.proto` 285 行），无需拆分**。

历史曾列出 15 个超 800 行的 `.pb.go`（`macro.pb.go` 2384 行等），均为 `gen/` 下机械产物，已按 §1.1 修订决策从 lint 审计中排除。

落地动作（已完成）：

- `scripts/lint-filesize.sh` 排除 `gen/` 目录，仅审计 `backend/`、`frontend/`、`proto/`、`scripts/`、`deploy/` 等源代码区
- `.github/workflows/ci.yml` 新增 `protocol-lint` job，串行执行 `lint-protocol.sh`（禁 REST/WS）+ `lint-filesize.sh`（800 行硬上限），双闸门强制启用
- 验证结果：`OK: 后端协议合规` / `OK: 所有受检文件 ≤ 800 行`

---

## 2. 新增 Service 蓝图

### 2.1 SystemService（M1）

**文件**：`proto/antclaw/v1/system.proto`

```proto
syntax = "proto3";
package antclaw.v1;

option go_package = "github.com/antclaw/antclaw/gen/go/antclaw/v1;antclawv1";

import "google/protobuf/timestamp.proto";

service SystemService {
  rpc Healthz(HealthzRequest) returns (HealthzResponse);
  rpc Readyz(ReadyzRequest) returns (ReadyzResponse);
  rpc Info(InfoRequest) returns (InfoResponse);
}

message HealthzRequest {}
message HealthzResponse {
  string status = 1;                 // healthy | degraded | unhealthy
  map<string, ComponentHealth> components = 2;
  google.protobuf.Timestamp checked_at = 3;
}
message ComponentHealth {
  string name = 1;
  string status = 2;                 // up | degraded | down
  string detail = 3;
}

message ReadyzRequest {}
message ReadyzResponse { bool ready = 1; }

message InfoRequest {}
message InfoResponse {
  string version = 1;
  string git_commit = 2;
  google.protobuf.Timestamp built_at = 3;
}
```

### 2.2 CryptoService 扩展（M1）

**文件**：扩展 `proto/antclaw/v1/crypto.proto`

```proto
service CryptoService {
  rpc GetCryptoPublicKey(GetCryptoPublicKeyRequest) returns (GetCryptoPublicKeyResponse);
  rpc PostEnvelope(PostEnvelopeRequest) returns (PostEnvelopeResponse);  // 新增
}

message PostEnvelopeRequest {
  string body_b64 = 1;
  string ts = 2;
  string nonce = 3;
  string sig = 4;
  string target_path = 5;        // 例如 "antclaw.v1.UserService/UpdateSettings"
  string target_body_b64 = 6;
}
message PostEnvelopeResponse { string body_b64 = 1; }
```

### 2.3 OptionsService（M2）

**文件**：`proto/antclaw/v1/options.proto`

```proto
syntax = "proto3";
package antclaw.v1;
option go_package = "github.com/antclaw/antclaw/gen/go/antclaw/v1;antclawv1";

service OptionsService {
  rpc GetGEX(GetGEXRequest) returns (GetGEXResponse);
  rpc GetIVSurface(GetIVSurfaceRequest) returns (GetIVSurfaceResponse);
  rpc GetSkew(GetSkewRequest) returns (GetSkewResponse);
  rpc GetIVAlerts(GetIVAlertsRequest) returns (GetIVAlertsResponse);
}

message GetGEXRequest {
  string asset = 1;             // 例如 "BTC", "ETH", "SPX"
  string venue = 2;             // deribit | cme
}
message GetGEXResponse {
  double total_gex = 1;         // 单位：$/1% spot move
  double zero_gamma = 2;        // 现货价格附近的 zero-gamma 水平
  repeated GEXBucket strikes = 3;
}
message GEXBucket {
  double strike = 1;
  double call_gex = 2;
  double put_gex = 3;
  double net_gex = 4;
}

message GetIVSurfaceRequest { string asset = 1; string venue = 2; }
message GetIVSurfaceResponse {
  repeated IVPoint points = 1;
}
message IVPoint {
  double strike = 1;
  int32 dte = 2;                // days to expiry
  double iv = 3;                // 0.32 = 32%
  string option_type = 4;       // call | put
}

message GetSkewRequest { string asset = 1; }
message GetSkewResponse {
  double rr_25d = 1;            // 25-delta risk reversal
  double bf_25d = 2;            // 25-delta butterfly
  double atm_iv = 3;
  double skew_z = 4;            // 历史百分位
}

message GetIVAlertsRequest { string asset = 1; }
message GetIVAlertsResponse { repeated IVAlert alerts = 1; }
message IVAlert {
  string kind = 1;              // skew_extreme | iv_spike | term_inversion
  string severity = 2;          // info | warn | critical
  string message = 3;
}
```

### 2.4 BacktestService 扩展（M2）

**文件**：扩展 `proto/antclaw/v1/backtest.proto`

新增方法：

```proto
rpc RunWalkForward(RunWalkForwardRequest) returns (RunWalkForwardResponse);
rpc BootstrapSignificance(BootstrapSignificanceRequest) returns (BootstrapSignificanceResponse);
rpc ApplyTrendFilter(ApplyTrendFilterRequest) returns (ApplyTrendFilterResponse);

message RunWalkForwardRequest {
  string strategy_id = 1;
  google.protobuf.Timestamp start = 2;
  google.protobuf.Timestamp end = 3;
  int32 train_days = 4;
  int32 test_days = 5;
  CostModel cost = 6;
}
message RunWalkForwardResponse {
  repeated WalkForwardWindow windows = 1;
  PerfMetrics aggregated = 2;
}
message WalkForwardWindow {
  google.protobuf.Timestamp train_start = 1;
  google.protobuf.Timestamp train_end = 2;
  google.protobuf.Timestamp test_start = 3;
  google.protobuf.Timestamp test_end = 4;
  PerfMetrics in_sample = 5;
  PerfMetrics out_sample = 6;
}
message PerfMetrics {
  double sharpe = 1;
  double sortino = 2;
  double calmar = 3;
  double max_dd = 4;
  double hit_rate = 5;
  double avg_pnl = 6;
  int32 trades = 7;
}
message CostModel {
  double commission_bps = 1;
  double slippage_bps = 2;
  double borrow_bps = 3;
}

message BootstrapSignificanceRequest {
  string strategy_id = 1;
  int32 iterations = 2;
}
message BootstrapSignificanceResponse {
  double p_value = 1;
  double mean_sharpe = 2;
  double std_sharpe = 3;
}

message ApplyTrendFilterRequest {
  string strategy_id = 1;
  string filter = 2;            // ema200 | sma50 | adx14
}
message ApplyTrendFilterResponse {
  PerfMetrics before = 1;
  PerfMetrics after = 2;
}
```

### 2.5 OnchainService（M4）

**文件**：`proto/antclaw/v1/onchain.proto`

```proto
service OnchainService {
  rpc GetMetrics(GetOnchainMetricsRequest) returns (GetOnchainMetricsResponse);
  rpc GetAnalysis(GetOnchainAnalysisRequest) returns (GetOnchainAnalysisResponse);
}

message GetOnchainMetricsRequest {
  string asset = 1;
  google.protobuf.Timestamp start = 2;
  google.protobuf.Timestamp end = 3;
}
message GetOnchainMetricsResponse {
  repeated OnchainPoint points = 1;
}
message OnchainPoint {
  google.protobuf.Timestamp time = 1;
  double active_addresses = 2;
  double tx_count = 3;
  double exchange_netflow = 4;
  double mvrv = 5;
  double sopr = 6;
}

message GetOnchainAnalysisRequest { string asset = 1; }
message GetOnchainAnalysisResponse {
  string regime = 1;            // accumulation | distribution | euphoria | capitulation
  double confidence = 2;
  string narrative = 3;
}
```

### 2.6 DeFiService（M4）

**文件**：`proto/antclaw/v1/defi.proto`

```proto
service DeFiService {
  rpc GetTopProtocols(GetTopProtocolsRequest) returns (GetTopProtocolsResponse);
  rpc GetProtocolTVL(GetProtocolTVLRequest) returns (GetProtocolTVLResponse);
  rpc GetAnalysis(GetDeFiAnalysisRequest) returns (GetDeFiAnalysisResponse);
}

message GetTopProtocolsRequest { int32 limit = 1; string chain = 2; }
message GetTopProtocolsResponse { repeated DeFiProtocol items = 1; }
message DeFiProtocol {
  string slug = 1;
  string name = 2;
  string category = 3;
  double tvl_usd = 4;
  double change_1d = 5;
  double change_7d = 6;
}
message GetProtocolTVLRequest { string slug = 1; }
message GetProtocolTVLResponse { repeated TVLPoint points = 1; }
message TVLPoint { google.protobuf.Timestamp time = 1; double tvl_usd = 2; }
message GetDeFiAnalysisRequest { string chain = 1; }
message GetDeFiAnalysisResponse {
  double total_tvl = 1;
  double tvl_change_7d = 2;
  string regime = 3;
  string narrative = 4;
}
```

### 2.7 SECService（M4）

**文件**：`proto/antclaw/v1/sec.proto`

```proto
service SECService {
  rpc ListFilings(ListFilingsRequest) returns (ListFilingsResponse);
  rpc GetFiling(GetFilingRequest) returns (GetFilingResponse);
  rpc GetAnalysis(GetSECAnalysisRequest) returns (GetSECAnalysisResponse);
}

message ListFilingsRequest {
  string cik = 1;
  string form_type = 2;          // 10-K | 10-Q | 8-K | 13F
  int32 limit = 3;
}
message ListFilingsResponse { repeated SECFiling items = 1; }
message SECFiling {
  string accession_number = 1;
  string form_type = 2;
  google.protobuf.Timestamp filed_at = 3;
  string company_name = 4;
  string url = 5;
}
message GetFilingRequest { string accession_number = 1; }
message GetFilingResponse {
  SECFiling filing = 1;
  string raw_text_excerpt = 2;
}
message GetSECAnalysisRequest { string ticker = 1; }
message GetSECAnalysisResponse {
  string sentiment = 1;          // bullish | neutral | bearish
  double risk_score = 2;
  string highlights = 3;
}
```

### 2.8 FedWatchService（M3）

**文件**：`proto/antclaw/v1/fedwatch.proto`

```proto
service FedWatchService {
  rpc GetFOMCProbabilities(GetFOMCProbabilitiesRequest) returns (GetFOMCProbabilitiesResponse);
}
message GetFOMCProbabilitiesRequest { google.protobuf.Timestamp meeting_date = 1; }
message GetFOMCProbabilitiesResponse {
  google.protobuf.Timestamp meeting_date = 1;
  repeated RateProbability probabilities = 2;
}
message RateProbability {
  double rate_low = 1;
  double rate_high = 2;
  double probability = 3;       // 0~1
}
```

### 2.9 MacroExtrasService（M3）

> 集中容纳 BIS / IMF / WB / ECB / Eurostat / OECD / SNB / TE / DTCC / Treasury 的查询。

**文件**：`proto/antclaw/v1/macro_extras.proto`

```proto
service MacroExtrasService {
  rpc GetSeries(GetMacroSeriesRequest) returns (GetMacroSeriesResponse);
  rpc ListAvailableSeries(ListAvailableSeriesRequest) returns (ListAvailableSeriesResponse);
}
message GetMacroSeriesRequest {
  string source = 1;             // bis | imf | worldbank | ecb | eurostat | oecd | snb | te | dtcc | treasury
  string series_id = 2;
  google.protobuf.Timestamp start = 3;
  google.protobuf.Timestamp end = 4;
}
message GetMacroSeriesResponse {
  repeated MacroPoint points = 1;
  string unit = 2;
  string frequency = 3;
}
message MacroPoint {
  google.protobuf.Timestamp time = 1;
  double value = 2;
}
message ListAvailableSeriesRequest { string source = 1; string keyword = 2; }
message ListAvailableSeriesResponse { repeated MacroSeriesMeta items = 1; }
message MacroSeriesMeta {
  string source = 1;
  string series_id = 2;
  string name = 3;
  string unit = 4;
  string frequency = 5;
}
```

### 2.10 TreasuryService（M3）

**文件**：`proto/antclaw/v1/treasury.proto`

```proto
service TreasuryService {
  rpc GetCurve(GetCurveRequest) returns (GetCurveResponse);
  rpc GetAnalysis(GetTreasuryAnalysisRequest) returns (GetTreasuryAnalysisResponse);
}
message GetCurveRequest { google.protobuf.Timestamp date = 1; }
message GetCurveResponse {
  google.protobuf.Timestamp date = 1;
  repeated YieldPoint points = 2;
}
message YieldPoint { string tenor = 1; double yield = 2; }   // tenor: "1M","3M","2Y","10Y" 等
message GetTreasuryAnalysisRequest {}
message GetTreasuryAnalysisResponse {
  double curve_2s10s = 1;
  double curve_3m10y = 2;
  string regime = 3;             // steepening | flattening | inverted | normal
}
```

### 2.11 SignalsService 扩展（Regime Overlay + Calibration）

**文件**：扩展 `proto/antclaw/v1/signals.proto`

```proto
rpc GetRegimeOverlay(GetRegimeOverlayRequest) returns (GetRegimeOverlayResponse);
rpc CalibrateConfidence(CalibrateConfidenceRequest) returns (CalibrateConfidenceResponse);

message GetRegimeOverlayRequest { string asset = 1; }
message GetRegimeOverlayResponse {
  string macro_regime = 1;        // expansion | slowdown | recession | recovery
  string vol_regime = 2;          // low | medium | high
  string liquidity_regime = 3;    // ample | tight
  double composite_score = 4;
  string narrative = 5;
}

message CalibrateConfidenceRequest {
  string strategy_id = 1;
  string method = 2;              // platt | isotonic
}
message CalibrateConfidenceResponse {
  double brier_before = 1;
  double brier_after = 2;
  bytes calibration_blob = 3;     // 序列化的校准模型，存于 calibration_models 表
}
```

### 2.12 SentimentService 扩展（M4）

**文件**：扩展 `proto/antclaw/v1/sentiment.proto`

```proto
rpc GetCBOEPutCall(GetCBOEPutCallRequest) returns (GetCBOEPutCallResponse);
rpc GetMyFXBookPositions(GetMyFXBookPositionsRequest) returns (GetMyFXBookPositionsResponse);
rpc GetInsiderTrades(GetInsiderTradesRequest) returns (GetInsiderTradesResponse);
rpc GetCryptoSocial(GetCryptoSocialRequest) returns (GetCryptoSocialResponse);
rpc GetFinvizMetrics(GetFinvizMetricsRequest) returns (GetFinvizMetricsResponse);

message GetCBOEPutCallRequest {}
message GetCBOEPutCallResponse {
  google.protobuf.Timestamp date = 1;
  double total_pc = 2;
  double equity_pc = 3;
  double index_pc = 4;
}

message GetMyFXBookPositionsRequest { string symbol = 1; }
message GetMyFXBookPositionsResponse {
  string symbol = 1;
  double long_pct = 2;
  double short_pct = 3;
  int64 long_lots = 4;
  int64 short_lots = 5;
}

message GetInsiderTradesRequest { string ticker = 1; int32 limit = 2; }
message GetInsiderTradesResponse { repeated InsiderTrade items = 1; }
message InsiderTrade {
  string ticker = 1;
  string insider = 2;
  string title = 3;
  string action = 4;             // BUY | SELL
  google.protobuf.Timestamp date = 5;
  double price = 6;
  int64 quantity = 7;
}

message GetCryptoSocialRequest { string asset = 1; }
message GetCryptoSocialResponse {
  google.protobuf.Timestamp date = 1;
  double twitter_followers_growth = 2;
  double reddit_subscribers_growth = 3;
  double sentiment_score = 4;
}

message GetFinvizMetricsRequest { string ticker = 1; }
message GetFinvizMetricsResponse {
  double short_ratio = 1;
  double short_pct_float = 2;
  double inst_own_pct = 3;
}
```

### 2.13 TAService 扩展（AMT 模块）

**文件**：扩展 `proto/antclaw/v1/ta.proto`

```proto
rpc GetAMTOpening(GetAMTOpeningRequest) returns (GetAMTOpeningResponse);
rpc GetAMTClose(GetAMTCloseRequest) returns (GetAMTCloseResponse);
rpc GetAMTDayType(GetAMTDayTypeRequest) returns (GetAMTDayTypeResponse);
rpc GetAMTRotation(GetAMTRotationRequest) returns (GetAMTRotationResponse);
rpc GetAMTMigration(GetAMTMigrationRequest) returns (GetAMTMigrationResponse);

// 共用请求体
message AMTRequest {
  string symbol = 1;
  google.protobuf.Timestamp date = 2;
}
message GetAMTOpeningRequest { AMTRequest base = 1; }
message GetAMTCloseRequest   { AMTRequest base = 1; }
message GetAMTDayTypeRequest { AMTRequest base = 1; }
message GetAMTRotationRequest{ AMTRequest base = 1; }
message GetAMTMigrationRequest{AMTRequest base = 1; }

message GetAMTOpeningResponse { string opening_type = 1; double confidence = 2; }
message GetAMTCloseResponse   { string close_type = 1;   double confidence = 2; }
message GetAMTDayTypeResponse { string day_type = 1;     double confidence = 2; }
message GetAMTRotationResponse{ string rotation = 1;     double confidence = 2; }
message GetAMTMigrationResponse{string migration = 1;    double confidence = 2; }
```

### 2.14 VolService 扩展（VIX/MOVE/CrossVol）

**文件**：扩展 `proto/antclaw/v1/vol.proto`

```proto
rpc GetMOVE(GetMOVERequest) returns (GetMOVEResponse);
rpc GetCrossVol(GetCrossVolRequest) returns (GetCrossVolResponse);
rpc GetTermStructure(GetTermStructureRequest) returns (GetTermStructureResponse);

message GetMOVERequest { google.protobuf.Timestamp start = 1; google.protobuf.Timestamp end = 2; }
message GetMOVEResponse { repeated VolPoint points = 1; }
message VolPoint { google.protobuf.Timestamp time = 1; double value = 2; }

message GetCrossVolRequest {}
message GetCrossVolResponse {
  double vix = 1;
  double move = 2;
  double dvol_btc = 3;
  double dvol_eth = 4;
  double composite = 5;
}

message GetTermStructureRequest { string asset = 1; }
message GetTermStructureResponse {
  repeated TermPoint points = 1;
  string regime = 2;             // contango | backwardation | flat
}
message TermPoint { int32 dte = 1; double iv = 2; }
```

### 2.15 CalendarService 扩展（影响评分）

**文件**：扩展 `proto/antclaw/v1/calendar.proto`

```proto
rpc GetSurpriseHistory(GetSurpriseHistoryRequest) returns (GetSurpriseHistoryResponse);
rpc GetImpactScores(GetImpactScoresRequest) returns (GetImpactScoresResponse);

message GetSurpriseHistoryRequest { string event_id = 1; int32 lookback = 2; }
message GetSurpriseHistoryResponse { repeated SurprisePoint items = 1; }
message SurprisePoint {
  google.protobuf.Timestamp time = 1;
  double actual = 2;
  double consensus = 3;
  double surprise = 4;
  double impact_weight = 5;
}

message GetImpactScoresRequest { string event_id = 1; }
message GetImpactScoresResponse {
  string event_id = 1;
  double cumulative_surprise = 2;
  double historical_impact = 3;
  string narrative = 4;
}
```

### 2.16 AIService 扩展（Chat / Context Builder / 解读）

**文件**：扩展 `proto/antclaw/v1/ai.proto`

```proto
rpc Chat(ChatRequest) returns (ChatResponse);
rpc GetInterpretation(GetInterpretationRequest) returns (GetInterpretationResponse);
rpc BuildContext(BuildContextRequest) returns (BuildContextResponse);

message ChatRequest {
  string session_id = 1;
  repeated ChatMessage messages = 2;
  string model = 3;              // 可空，由后端选择
  bool prefer_byok = 4;
}
message ChatMessage { string role = 1; string content = 2; }
message ChatResponse { ChatMessage reply = 1; string used_model = 2; bool from_cache = 3; }

message GetInterpretationRequest {
  string subject = 1;            // signal_id | regime | event_id
  string subject_id = 2;
  string locale = 3;
}
message GetInterpretationResponse { string text = 1; bool from_cache = 2; }

message BuildContextRequest { string asset = 1; string scope = 2; }   // scope: "macro"|"crypto"|"forex"
message BuildContextResponse { string context_blob = 1; }            // 用于 Prompt 注入
```

---

## 3. 生成与注册

### 3.1 buf 生成

```bash
cd /opt/antclaw
buf generate
# 产物：
#   gen/go/antclaw/v1/<service>.pb.go
#   gen/go/antclaw/v1/antclawv1connect/<service>.connect.go
#   gen/ts/antclaw/v1/<service>_pb.ts
#   gen/ts/antclaw/v1/<service>_connect.ts
```

### 3.2 在 cmd/antclaw-api/main.go 注册

每个新 Service 都需添加：

```go
mux.Handle(antclawv1connect.NewSystemServiceHandler(systemHandler))
mux.Handle(antclawv1connect.NewOptionsServiceHandler(optionsHandler))
mux.Handle(antclawv1connect.NewOnchainServiceHandler(onchainHandler))
mux.Handle(antclawv1connect.NewDeFiServiceHandler(defiHandler))
mux.Handle(antclawv1connect.NewSECServiceHandler(secHandler))
mux.Handle(antclawv1connect.NewFedWatchServiceHandler(fedwatchHandler))
mux.Handle(antclawv1connect.NewMacroExtrasServiceHandler(macroExtrasHandler))
mux.Handle(antclawv1connect.NewTreasuryServiceHandler(treasuryHandler))
```

---

## 4. 完成判据

- [ ] 每个新 proto 文件存在并通过 `buf lint`
- [ ] `buf generate` 成功，Go/TS 桩提交
- [ ] 所有新 Service 在 `cmd/antclaw-api/main.go` 注册
- [ ] 至少一个 `curl` 端到端验证返回 `200 OK`（即使是空数据）
