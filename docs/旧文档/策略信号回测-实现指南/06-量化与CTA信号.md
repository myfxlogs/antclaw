# 06 · 量化信号 与 CTA 信号（quant / cta）

> 替换 `signals_handler.go:176-182` 两个 placeholder。基于 `price_daily/price_intraday` 实时计算策略信号 + 历史 Sharpe/MaxDD。

---

## 1. GetQuant — 量化策略信号

### 1.1 目标
对给定 `pair` 返回 N 个量化策略的当前信号 + 历史绩效统计：

| Strategy | 类型 |
|----------|-----|
| `momentum_12_3` | 12-1 月动量减反转 |
| `mean_reversion_5d` | 5 日 RSI 均值回归 |
| `breakout_donchian_20` | 唐奇安 20 日通道突破 |
| `low_vol_60d` | 60 日低波动持有 |
| `pair_carry` | （仅 FX）利差 |

### 1.2 算法（每个策略一个纯函数）

文件：`backend/internal/service/quant/strategies.go`

```go
type StrategyResult struct {
    Name     string
    Signal   string  // long/short/flat
    Strength float64 // 0-1
}

func MomentumTwelveMinusOne(bars []Bar) StrategyResult {
    if len(bars) < 252+30: return ZeroResult
    ret_12_1 := bars[len-21].close / bars[len-252].close - 1
    last_1m  := bars[len-1].close / bars[len-21].close - 1
    score    := ret_12_1 - last_1m
    sig := "flat"
    if score > 0.05 { sig = "long" }
    if score < -0.05 { sig = "short" }
    return StrategyResult{"momentum_12_3", sig, math.Abs(score)/0.2}
}

func MeanReversion5d(bars []Bar) StrategyResult { /* RSI(14) <30 long, >70 short */ }
func BreakoutDonchian20(bars []Bar) StrategyResult { /* 收盘破20日高=long, 破20日低=short */ }
func LowVol60d(bars []Bar) StrategyResult { /* 60d annualized vol percentile <30 → long */ }
func PairCarry(symbol string, ratesProv RatesProvider) StrategyResult { /* fx interest差 */ }
```

### 1.3 历史 Sharpe / MaxDD 缓存

新建表 `quant_strategy_perf`：
```sql
CREATE TABLE IF NOT EXISTS quant_strategy_perf (
    asof_date DATE NOT NULL,
    symbol    VARCHAR(32) NOT NULL,
    strategy  VARCHAR(64) NOT NULL,
    sharpe    DOUBLE PRECISION,
    sortino   DOUBLE PRECISION,
    drawdown  DOUBLE PRECISION,
    win_rate  DOUBLE PRECISION,
    sample_trades INT,
    PRIMARY KEY (asof_date, symbol, strategy)
);
```

Worker 每日 03:00 计算最近 2 年滚动 Sharpe/MaxDD（用回测引擎，文档 07 提供）。

### 1.4 GetQuant 流程
```
bars = PriceProvider.GetDailyBars(pair, now-2y, now)
strategies = [MomentumTwelveMinusOne, MeanReversion5d, BreakoutDonchian20, LowVol60d]
if isFX(pair): strategies += [PairCarry]

results = []
for strat in strategies:
    sr = strat(bars)
    perf = SELECT sharpe, drawdown FROM quant_strategy_perf
           WHERE asof_date = (SELECT MAX(asof_date) FROM quant_strategy_perf
                              WHERE symbol=$1 AND strategy=$2)
    results.append(QuantSignal{
        Pair: pair, Strategy: sr.Name,
        Signal: sr.Signal, Sharpe: perf.Sharpe, Drawdown: perf.Drawdown,
    })
```

---

## 2. GetCta — CTA 趋势信号

### 2.1 目标
返回当前趋势方向 + 动量值 + 波动率制度判定。

### 2.2 算法
```
bars = PriceProvider.GetDailyBars(pair, now-1y, now)
adx = computeADX(bars, 14)       # 趋势强度
ema20 = ema(bars, 20).last
ema50 = ema(bars, 50).last
ema200 = ema(bars, 200).last
atr14  = atr(bars, 14).last

trend = "bullish"  if ema20 > ema50 > ema200 else
        "bearish"  if ema20 < ema50 < ema200 else "sideways"

momentum = (close - ema50) / atr14   # 用 ATR 归一

regime = "trending" if adx > 25 else "ranging"
         if vol_60d_percentile > 80: regime = "volatile"
```

### 2.3 输出
```go
&CtaSignal{
    Pair: pair,
    Trend: trend,            // bullish/bearish/sideways
    Momentum: momentum,      // -3..+3 (ATR 单位)
    Regime: regime,          // trending/ranging/volatile
}
```

### 2.4 数据不足
- bars < 200 → DATA_INSUFFICIENT

---

## 3. 公共指标库

文件：`backend/internal/service/ta/indicators.go`（如已部分存在则补全）

```go
func EMA(bars []Bar, period int) []float64
func SMA(bars []Bar, period int) []float64
func RSI(bars []Bar, period int) []float64
func ATR(bars []Bar, period int) []float64
func ADX(bars []Bar, period int) []float64
func DonchianChannel(bars []Bar, period int) (upper, lower []float64)
func RollingStdDev(values []float64, period int) []float64
func Percentile(values []float64, p float64) float64
```

每个函数必须有单测（黄金对比 + 边界）。

---

## 4. 修改清单

| 文件 | 动作 |
|------|------|
| `backend/internal/service/quant/service.go` | 新建 |
| `backend/internal/service/quant/strategies.go` | 新建 |
| `backend/internal/service/quant/strategies_test.go` | 新建 |
| `backend/internal/service/cta/service.go` | 新建 |
| `backend/internal/service/cta/service_test.go` | 新建 |
| `backend/internal/service/ta/indicators.go` | 新建 / 补全 |
| `backend/internal/service/ta/indicators_test.go` | 新建 |
| `backend/internal/adapter/storage/postgres/ensure_schema.go` | 新增 `quant_strategy_perf` |
| `backend/cmd/antclaw-worker/quant_perf.go` | 每日计算 perf |
| `backend/internal/service/signals/service.go` | `GetQuant` / `GetCta` 委托 |

---

## 5. 验证

```bash
# 触发 perf 任务
docker compose exec -T worker /app/antclaw-worker --once=quant_perf

docker compose exec -T postgres psql -U antclaw -d antclaw -c \
  "SELECT symbol, strategy, sharpe, drawdown FROM quant_strategy_perf
   WHERE asof_date=CURRENT_DATE LIMIT 10;"

# RPC
curl -s http://localhost:8082/antclaw.v1.SignalsService/GetQuant \
  -d '{"pair":"EURUSD"}' -H 'Content-Type: application/json' | jq .
curl -s http://localhost:8082/antclaw.v1.SignalsService/GetCta \
  -d '{"pair":"EURUSD"}' -H 'Content-Type: application/json' | jq .
```

预期：
- GetQuant 返回 4-5 条不同 strategy 的 signal + 真实 sharpe/drawdown
- GetCta 返回当前 trend/momentum/regime

## 6. 实施记录

<!-- -->
