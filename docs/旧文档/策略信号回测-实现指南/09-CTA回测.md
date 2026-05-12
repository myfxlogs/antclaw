# 09 · CTA 回测（/ctabt）

> ARK `/ctabt` 命令完全缺失。基于 CTA 趋势策略（多周期均线、突破、波动率自适应）在多标的上回测。

> **依赖**：先完成 `07-回测引擎QuantBT.md`。

---

## 1. 目标

提供一个 CTA 策略组合回测，支持以下 CTA 范式：

| 策略 | 描述 |
|------|------|
| `cta_dual_ema` | 双均线（EMA20/EMA50）金叉死叉 |
| `cta_donchian_breakout` | 唐奇安通道突破（55/20）|
| `cta_atr_trail` | ATR 跟踪止损 + 动量入场 |
| `cta_multi_tf` | 多周期共振（日线趋势 + 4h 入场） |

每个策略支持 vol-targeting：根据品种 GARCH 波动率动态调整仓位。

---

## 2. 关键设计

### 2.1 Vol-targeting

```
target_vol_annual = 0.15       # 默认 15% 年化
realized_vol = stddev(daily_returns last 60 bars) * sqrt(252)
position_scale = target_vol_annual / max(realized_vol, 0.05)
qty = base_qty * position_scale
```

### 2.2 多周期共振（cta_multi_tf）

```
daily_trend = ema20_d > ema50_d ? +1 : (ema20_d < ema50_d ? -1 : 0)
intraday_signal = donchian_breakout_4h(symbol)

if daily_trend > 0 && intraday_signal == "long":  enter long
if daily_trend < 0 && intraday_signal == "short": enter short
```

需要在 Engine 时同时加载多个 timeframe；`profile_loader` 模式扩展为 `barLoader(symbol, tf, ts)`。

---

## 3. RPC

```proto
message CtaBtConfig {
  string strategy = 1;          // cta_dual_ema / cta_donchian_breakout / cta_atr_trail / cta_multi_tf
  repeated string symbols = 2;
  string from_date = 3;
  string to_date = 4;
  string timeframe = 5;         // 主时序
  string secondary_timeframe = 6; // 仅 multi_tf 用
  double target_vol = 7;        // 默认 0.15
  map<string,double> params = 8;
  CostModel cost_model = 9;
  RiskConfig risk = 10;
}

service BacktestService {
  rpc RunCtaBt(RunCtaBtRequest) returns (RunCtaBtResponse);
}
```

`backtest_jobs.type = 'ctabt'`。

---

## 4. 修改清单

| 文件 | 动作 |
|------|------|
| `proto/antclaw/v1/backtest.proto` | 增 RunCtaBt + CtaBtConfig |
| `backend/internal/service/backtest/strategies/cta_dual_ema.go` | 新建 |
| `backend/internal/service/backtest/strategies/cta_donchian_breakout.go` | 新建 |
| `backend/internal/service/backtest/strategies/cta_atr_trail.go` | 新建 |
| `backend/internal/service/backtest/strategies/cta_multi_tf.go` | 新建 |
| `backend/internal/service/backtest/sizing/vol_target.go` | 新建 |
| `backend/internal/service/backtest/multi_tf_loader.go` | 新建 |
| `backend/internal/service/backtest/service.go` | 实现 RunCtaBt |
| `backend/internal/adapter/rpc/backtest_handler.go` | 新增 RunCtaBt handler |
| `backend/internal/service/backtest/strategies/cta_*_test.go` | 新建（单测）|

---

## 5. 验证

```bash
JOB=$(curl -s http://localhost:8082/antclaw.v1.BacktestService/RunCtaBt \
  -H 'Content-Type: application/json' \
  -d '{"config":{
       "strategy":"cta_dual_ema",
       "symbols":["EURUSD","GBPUSD","USDJPY"],
       "from_date":"2020-01-01","to_date":"2025-01-01",
       "timeframe":"1d",
       "target_vol":0.15,
       "params":{"fast":20,"slow":50},
       "cost_model":{"slippage_bps":2,"spread_bps":1,"commission_per_trade":2},
       "risk":{"initial_capital":100000,"risk_per_trade":0.01}}}' | jq -r .task_id)

# 轮询 + 拉结果
```

预期：
- 多标的同时进行；trades 在 3 个标的上分布
- vol-targeting 体现在 qty 上：低波动品种仓位更大
- multi_tf 日线 + 4h 数据均加载成功

---

## 6. 完成判定

- [ ] 4 个 CTA 策略 adapter 实现完成
- [ ] vol_target.go sizing 通过单测
- [ ] multi_tf_loader 通过单测
- [ ] 多标的 + 多周期回测无并发错误
- [ ] 结果中 metadata 含 `vol_target_scale_avg` 字段

## 7. 实施记录

<!-- -->
