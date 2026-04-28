# 08 · 成交量轮廓回测 VpBT

> 替换 `backtest/service.go:151 RunVpBt` 空壳。基于 `volume_profiles` 表的 POC/VAH/VAL 实施反弹/突破策略回测。

> **依赖**：先完成 `07-回测引擎QuantBT.md`（共用 Engine / StrategyAdapter / 异步队列）。

---

## 1. 目标

VpBt 是 `RunQuantBt` 的特例：策略选项固定为基于 Volume Profile 的若干变体，但保留同样的：成本模型、风控、异步队列、结果格式。

---

## 2. 策略变体

| 名称 | 入场规则 | 止损/止盈 |
|------|---------|----------|
| `vp_poc_revert` | 价格触及 POC ± k*ATR 反向回归 | SL=对侧 VAH/VAL；TP=POC |
| `vp_va_breakout` | 价格突破 VAH（多头）或跌破 VAL（空头）| SL=回到 VAH/VAL 内；TP=N×ATR |
| `vp_imbalance_fade` | 大幅偏离价值区（价格 vs POC 距离 > p95）| SL=新极端；TP=POC |

---

## 3. 数据接入

### 3.1 已有表
- `volume_profiles(time, symbol, period, poc, vah, val, profile)`
- `price_intraday` 用作入场判断 + 模拟成交

### 3.2 时间对齐
- VpBt 默认以 1h 或 4h `price_intraday` 作回测主时序；
- POC/VAH/VAL 取该 bar 时间所在 30d 滚动 profile（从 `volume_profiles` 查询 `period='30d'` 最近一条 ≤ bar.time）。

### 3.3 数据加载优化
- 一次性加载整个回测区间 `volume_profiles` + `price_intraday`，按 `(symbol, time)` 双索引内存中查找。

---

## 4. 实现要点

文件：`backend/internal/service/backtest/strategies/vp_*.go`

```go
type VPStrategyBase struct {
    profileLoader func(symbol string, ts time.Time) *VolumeProfile
    atrPeriod     int
    distFactor    float64
}

func (s *VPPocRevert) OnBar(ctx, ts, bars) []TradeSignal {
    sym := "EURUSD"
    bar := bars[sym][last]
    profile := s.profileLoader(sym, ts)
    if profile == nil: return nil
    atr := atrCache.Get(sym, ts, s.atrPeriod)
    deviation := bar.Close - profile.POC
    if math.Abs(deviation) >= s.distFactor*atr:
        side := "short"; if deviation < 0 { side = "long" }
        stopDist := math.Abs(profile.VAH - profile.VAL) / bar.Close
        return []TradeSignal{{
            Symbol: sym, Side: side,
            StopDistance: stopDist,
            TargetDistance: math.Abs(deviation)/bar.Close,
            Reason: "vp_poc_revert",
        }}
}
```

---

## 5. 修改清单

| 文件 | 动作 |
|------|------|
| `backend/internal/service/backtest/strategies/vp_poc_revert.go` | 新建 |
| `backend/internal/service/backtest/strategies/vp_va_breakout.go` | 新建 |
| `backend/internal/service/backtest/strategies/vp_imbalance_fade.go` | 新建 |
| `backend/internal/service/backtest/profile_loader.go` | 新建（VP 数据缓存）|
| `backend/internal/service/backtest/service.go` | 实现 `RunVpBt`（复用 RunQuantBt 但策略限定为 vp_*）|
| `backend/internal/service/backtest/strategies/vp_*_test.go` | 新建 |

---

## 6. RPC

`RunVpBt` 与 `RunQuantBt` 共用 backtest_jobs 表（`type='vpbt'`）。请求 schema：

```proto
message VpBtConfig {
  string strategy = 1;     // vp_poc_revert / vp_va_breakout / vp_imbalance_fade
  repeated string symbols = 2;
  string from_date = 3;
  string to_date = 4;
  string timeframe = 5;    // 1h/4h
  int32  atr_period = 6;
  double dist_factor = 7;  // POC 距离阈值，ATR 倍数
  CostModel cost_model = 8;
  RiskConfig risk = 9;
}
```

---

## 7. 验证

```bash
JOB=$(curl -s http://localhost:8082/antclaw.v1.BacktestService/RunVpBt \
  -H 'Content-Type: application/json' \
  -d '{"config":{"strategy":"vp_poc_revert","symbols":["EURUSD"],
       "from_date":"2024-01-01","to_date":"2025-01-01",
       "timeframe":"1h","atr_period":14,"dist_factor":1.5,
       "cost_model":{"slippage_bps":2,"spread_bps":1,"commission_per_trade":1},
       "risk":{"initial_capital":50000,"risk_per_trade":0.01}}}' | jq -r .task_id)

# 轮询 + 拉结果（同 07）
```

预期：返回真实 trades + summary，trades 中 entry/exit 时间能映射回 volume_profiles 中的 POC/VAH/VAL。

## 8. 实施记录

<!-- -->
