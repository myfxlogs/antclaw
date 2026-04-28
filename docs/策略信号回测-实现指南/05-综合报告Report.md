# 05 · 综合报告 Report

> ARK `/report` 命令完全缺失。提供给定标的的多维聚合分析报告（一站式）。

---

## 1. 目标

`GetReport(symbol, sections?)` 返回结构化报告，组合：
- 当前 Bias / Unified Signal
- COT 摘要
- 宏观状态 + 转换历史
- 因子分解（X-Factors）
- 信号准确率（accuracy 1W/1M）
- 资金流背离 top 3
- 波动率 / 期限结构
- AI 文本摘要（若用户启用）

---

## 2. RPC 设计

新建 `proto/antclaw/v1/report.proto`：

```proto
syntax = "proto3";
package antclaw.v1;
option go_package = "github.com/antclaw/antclaw/gen/go/antclaw/v1;antclawv1";

import "antclaw/v1/signals.proto";
import "antclaw/v1/cot.proto";

service ReportService {
  rpc GetReport(GetReportRequest) returns (GetReportResponse);
}

message GetReportRequest {
  string symbol = 1;
  repeated string sections = 2;   // 空=全部；可指定 bias/cot/regime/factor/accuracy/flow/vol/ai
  bool   with_ai_summary = 3;
}

message GetReportResponse {
  string symbol = 1;
  string generated_at = 2;
  ReportBias    bias    = 3;
  ReportCOT     cot     = 4;
  ReportRegime  regime  = 5;
  ReportFactor  factor  = 6;
  ReportAccuracy accuracy = 7;
  ReportFlow    flow    = 8;
  ReportVol     vol     = 9;
  string ai_summary = 10;          // 若请求开启
  repeated string missing_sections = 11;  // 数据不足的 section 名
}

message ReportBias {
  string direction = 1;
  double confidence = 2;
  string reasoning_short = 3;
}
message ReportCOT {
  double cot_index = 1;
  double zscore   = 2;
  string direction = 3;
  int64  net_position = 4;
  int64  wow_change = 5;
  string report_date = 6;
}
message ReportRegime {
  string label = 1;
  double unified_score = 2;
  double hmm_confidence = 3;
  repeated string recent_transitions = 4;  // ["NEUTRAL→BULL @ 2026-04-20", ...]
}
message ReportFactor {
  double composite = 1;
  map<string,double> breakdown = 2;
  int32 rank_in_category = 3;
  int32 category_size = 4;
}
message ReportAccuracy {
  double accuracy_1w = 1;
  double accuracy_1m = 2;
  double avg_return_1w = 3;
  double avg_return_1m = 4;
  int32  sample_size = 5;
}
message ReportFlow {
  repeated FlowDivergenceItem top = 1;
}
message FlowDivergenceItem { string pair_b=1; double z_score=2; int32 lead_lag=3; }
message ReportVol {
  double annualized = 1;
  string regime = 2;
  double atm_iv = 3;          // 若有期权
  double skew_25d = 4;
  double percentile_30d = 5;
}
```

---

## 3. Service 实现

文件：`backend/internal/service/report/service.go`

```go
type Service struct {
    signals signals.Service
    cot     COTProvider
    regime  MacroRegimeProvider
    factor  FactorProvider
    flow    FlowProvider
    vol     VolProvider
    backtest backtest.Service
    ai      ai.Client       // 可选；nil 表示不生成摘要
    log     *slog.Logger
}

func (s *Service) GetReport(ctx context.Context, in In) (*Out, error) {
    out := &Out{Symbol: in.Symbol, GeneratedAt: time.Now()}
    sections := defaultIfEmpty(in.Sections, allSections)

    // 并发拉取每个 section（errgroup）
    g, gctx := errgroup.WithContext(ctx)
    if want("bias", sections)   { g.Go(func() error { ... }) }
    if want("cot",  sections)   { g.Go(func() error { ... }) }
    ...
    if err := g.Wait(); err != nil { return nil, err }

    if in.WithAISummary && s.ai != nil {
        out.AISummary = s.composeAISummary(ctx, out)
    }
    return out, nil
}
```

每个 section 失败应记录到 `missing_sections`，**不**让整体失败。

---

## 4. AI 摘要 prompt 模板

文件：`backend/internal/service/report/ai_prompt.go`

```go
const reportSummaryPromptZH = `你是一名宏观交易分析师。请基于以下结构化数据，撰写不超过 200 字的中文交易简报，覆盖：方向偏好、关键驱动因素、主要风险与建议关注事件。语气专业，禁止编造未提供的数据。

【标的】{{.Symbol}}
【方向】{{.Bias.Direction}}（置信度 {{.Bias.Confidence | printf "%.2f"}}）
【COT】index={{.COT.COTIndex | printf "%.0f"}}, z={{.COT.ZScore | printf "%.2f"}}, 方向={{.COT.Direction}}
【宏观】{{.Regime.Label}} score={{.Regime.UnifiedScore | printf "%.2f"}}, hmm_conf={{.Regime.HMMConfidence | printf "%.2f"}}
【因子】composite={{.Factor.Composite | printf "%.0f"}}, 在类别中排名 {{.Factor.RankInCategory}}/{{.Factor.CategorySize}}
【近期信号准确率】1W={{.Accuracy.Accuracy1W | printf "%.0f%%"}}（n={{.Accuracy.SampleSize}}）
【波动】regime={{.Vol.Regime}}, percentile_30d={{.Vol.Percentile30D | printf "%.0f"}}
`
```

调用 `system_ai_configs` 中 `primary_for=summarizer` 的 provider。失败时返回 `""`，并将 `missing_sections += ["ai_summary"]`。

---

## 5. 修改清单

| 文件 | 动作 |
|------|------|
| `proto/antclaw/v1/report.proto` | 新建 |
| `backend/internal/service/report/service.go` | 新建 |
| `backend/internal/service/report/ai_prompt.go` | 新建 |
| `backend/internal/service/report/service_test.go` | 新建 |
| `backend/internal/adapter/rpc/report_handler.go` | 新建 |
| `backend/cmd/antclaw-api/main.go` | 注册 ReportHandler |

---

## 6. 验证

```bash
curl -s http://localhost:8082/antclaw.v1.ReportService/GetReport \
  -d '{"symbol":"EURUSD","with_ai_summary":true}' \
  -H 'Content-Type: application/json' | jq .

# 期望：bias/cot/regime/factor 各 section 都填充；ai_summary 含 100-200 字中文摘要
```

---

## 7. 完成判定

- [ ] 9 个 section 全部能正常返回（除非数据缺失）
- [ ] missing_sections 字段生效
- [ ] AI 摘要可关闭可启用
- [ ] 一次报告 P95 ≤ 1.5 秒（不含 AI），含 AI ≤ 5 秒

## 8. 实施记录

<!-- -->
