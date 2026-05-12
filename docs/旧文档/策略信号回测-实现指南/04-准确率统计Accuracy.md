# 04 · 准确率统计 Accuracy

> 替换 `backtest/service.go:130` 的硬编码假数据。基于 `signal_outcomes` 表统计真实准确率。

---

## 1. 目标

1. Worker 周期任务：扫描 `unified_signals`，对每条信号在 1D/1W/2W/1M 后评估 `signal_outcomes`；
2. `GetAccuracy(strategy_id, period)` 返回真实统计：
   - DirectionalAccuracy（方向准确率）
   - HitRate（达到 >0.5% 收益的比例）
   - AvgReturn
   - Sharpe / Sortino
   - 分 horizon、分 symbol、分 recommendation 切片

---

## 2. 评估器（Worker）

文件：`backend/cmd/antclaw-worker/outcome_evaluator.go`

调度：每小时一次。

```
horizons = [
    ("1D", 1*24h), ("1W", 7*24h), ("2W", 14*24h), ("1M", 30*24h),
]

for h in horizons:
    rows = SELECT s.id, s.symbol, s.issued_at, s.recommendation, s.unified_score
           FROM unified_signals s
           LEFT JOIN signal_outcomes o ON o.signal_id=s.id AND o.horizon=h.label
           WHERE o.signal_id IS NULL                                  -- 尚未评估
             AND s.issued_at <= now() - h.duration                    -- 已经过 horizon

    for row in rows:
        bars = price_intraday or price_daily 取 issued_at 与 issued_at+h.duration
              的 close
        if 缺数据: continue (下次再试)
        ret = (close_after - close_at) / close_at
        direction_match = sign(ret) == direction_of(recommendation)
        INSERT INTO signal_outcomes (signal_id, horizon, return_pct, direction_match, evaluated_at)
        VALUES ($1, $2, $3, $4, now())
```

`direction_of(rec)`：
```go
LONG / STRONG_LONG  → +1
NEUTRAL             →  0  (此时 direction_match 取 |ret| < threshold)
SHORT / STRONG_SHORT → -1
```

---

## 3. GetAccuracy 实现

修改 `backend/internal/service/backtest/service.go`：

```go
func (s *Service) GetAccuracy(ctx context.Context, strategyID string, period *backtestv1.TimeRange) (*backtestv1.GetAccuracyResponse, error) {
    horizon := "1W"   // 默认；如 strategyID 含 horizon 后缀解析
    symbol  := ""     // strategyID 形如 "unified:EURUSD:1W" 解析

    parsed := parseStrategyKey(strategyID)
    horizon = parsed.Horizon; symbol = parsed.Symbol

    from, to := periodRange(period)

    stats := s.signalRepo.GetOutcomeStats(ctx, "unified", symbol, horizon, from, to)
    if stats.SampleSize < 30 {
        return nil, ErrDataInsufficient
    }
    return &backtestv1.GetAccuracyResponse{
        StrategyId: strategyID,
        Metrics: &backtestv1.AccuracyMetrics{
            DirectionalAccuracy: stats.DirectionalAccuracy,
            AvgReturn:           stats.AvgReturn,
            HitRate:             stats.HitRate,
        },
    }, nil
}
```

`parseStrategyKey` 规则：
```
"unified:EURUSD:1W" → {Type:"unified", Symbol:"EURUSD", Horizon:"1W"}
"unified:*:1W"      → {Symbol:""(全部), Horizon:"1W"}
"unified:EURUSD"    → {Horizon:"1W"} (默认)
```

---

## 4. SignalRepo.GetOutcomeStats 实现

```sql
SELECT
  COUNT(*)                                                 AS sample_size,
  AVG(CASE WHEN direction_match THEN 1.0 ELSE 0.0 END)     AS dir_acc,
  AVG(return_pct)                                          AS avg_return,
  AVG(CASE WHEN return_pct > 0.005 THEN 1.0 ELSE 0.0 END)  AS hit_rate,
  STDDEV(return_pct)                                       AS sigma
FROM signal_outcomes o
JOIN unified_signals s ON s.id = o.signal_id
WHERE o.horizon = $1
  AND ($2 = '' OR s.symbol = $2)
  AND s.issued_at >= $3
  AND s.issued_at <= $4
```

`Sharpe` 后处理：`avg_return / sigma * sqrt(252 / horizon_days)`（1D=252, 1W=52, 1M=12）。

`Sortino` 需额外查询负收益方差，单独 query：
```sql
SELECT STDDEV(return_pct) FROM signal_outcomes o JOIN unified_signals s ON s.id=o.signal_id
WHERE o.horizon=$1 AND ($2=''OR s.symbol=$2)
  AND o.return_pct < 0 AND s.issued_at BETWEEN $3 AND $4
```

---

## 5. 扩展 proto（可选，避免破坏）

`AccuracyMetrics` 已有 3 字段。建议增加：
```proto
message AccuracyMetrics {
  double directional_accuracy = 1;
  double avg_return = 2;
  double hit_rate = 3;
  double sharpe   = 4;   // 新增
  double sortino  = 5;
  int32  sample_size = 6;
  double std_dev = 7;
}
```

---

## 6. 修改清单

| 文件 | 动作 |
|------|------|
| `backend/cmd/antclaw-worker/outcome_evaluator.go` | 新建 |
| `backend/cmd/antclaw-worker/main.go` | 注册 outcome_evaluator |
| `backend/internal/adapter/storage/postgres/signal_repo.go` | 实现 GetOutcomeStats / SaveOutcome |
| `backend/internal/service/backtest/service.go` | 重写 GetAccuracy |
| `backend/internal/service/backtest/keyparse.go` | 新建 parseStrategyKey |
| `backend/internal/service/backtest/keyparse_test.go` | 新建 |
| `proto/antclaw/v1/backtest.proto` | 扩展 AccuracyMetrics |

---

## 7. 验证

```bash
# 触发一次评估
docker compose exec -T worker /app/antclaw-worker --once=outcome_evaluator

# 检查 outcomes
docker compose exec -T postgres psql -U antclaw -d antclaw -c \
  "SELECT horizon, COUNT(*), AVG(return_pct), AVG(direction_match::int)
   FROM signal_outcomes GROUP BY horizon;"

# RPC
curl -s http://localhost:8082/antclaw.v1.BacktestService/GetAccuracy \
  -d '{"strategy_id":"unified:EURUSD:1W"}' \
  -H 'Content-Type: application/json' | jq .
```

预期：返回真实统计；样本不足时返回 FailedPrecondition。

## 8. 实施记录

<!-- -->
