# 07 · 回测引擎 QuantBT（异步）

> 把 `backtest/service.go:141 RunQuantBt` 从空壳改造为完整异步回测引擎。是文档 08/09/10 的基础。

---

## 1. 目标

实现一个支持以下能力的异步回测引擎：

- 多策略 / 多标的同时回测
- 多种入场/出场规则（信号驱动 + 止盈/止损/时间止）
- 滑点 / 点差 / 佣金成本模型（与文档 10 共用）
- 完整 trade ledger + equity curve + 指标
- 任务持久化到 `backtest_jobs` / `backtest_results`
- 取消、进度查询、超时

---

## 2. 调度模型

```
┌─────────┐  Submit   ┌──────────────┐
│ RPC API │──────────▶│ backtest_jobs│  status=queued
└─────────┘           └──────┬───────┘
                             │ Redis Stream "backtest:queue"
                             ▼
                      ┌──────────────┐
                      │ Worker Pool  │  (N=4 concurrent)
                      │ executor.go  │
                      └──────┬───────┘
                             │
                             ▼
                      ┌──────────────┐
                      │ Engine       │ status=running
                      │ Run()        │
                      └──────┬───────┘
                             │
                             ▼
                      ┌──────────────┐
                      │ Result Store │ status=done/failed
                      └──────────────┘
```

**取消**：用户调用 `CancelJob(job_id)` → 写 Redis key `backtest:cancel:{job_id}`，Engine 在每个 bar 检查。

---

## 3. 数据契约

### 3.1 RPC 请求扩展

修改 `proto/antclaw/v1/backtest.proto`：

```proto
message QuantBtConfig {
  string strategy = 1;            // momentum_12_3 / mean_reversion_5d / breakout_donchian_20 / ...
  repeated string symbols = 2;
  string from_date = 3;           // YYYY-MM-DD
  string to_date = 4;
  string timeframe = 5;           // 1d/4h/1h
  map<string,double> params = 6;  // 策略参数
  CostModel cost_model = 7;
  RiskConfig risk = 8;
}

message CostModel {
  double slippage_bps = 1;        // 单边
  double spread_bps  = 2;
  double commission_per_trade = 3; // USD/合约
}

message RiskConfig {
  double initial_capital = 1;     // 默认 100000
  double risk_per_trade  = 2;     // 默认 0.01 = 1% 资金
  double max_concurrent_positions = 3; // 默认 3
  double max_drawdown_stop = 4;   // 触发后停止回测，默认 0.3
}

message RunQuantBtResponse {
  string task_id = 1;
  string status = 2;
}

// 新增查询接口
service BacktestService {
  rpc RunQuantBt(RunQuantBtRequest) returns (RunQuantBtResponse);
  rpc GetJob(GetJobRequest) returns (GetJobResponse);
  rpc CancelJob(CancelJobRequest) returns (CancelJobResponse);
  rpc ListJobs(ListJobsRequest) returns (ListJobsResponse);
  rpc GetResult(GetResultRequest) returns (GetResultResponse);
  // ...
}

message GetResultResponse {
  string job_id = 1;
  BacktestSummary summary = 2;
  repeated EquityPoint equity_curve = 3;
  repeated TradeRecord trades = 4;
  AdvancedMetrics advanced = 5;
}

message BacktestSummary {
  double total_return = 1;
  double cagr = 2;
  double sharpe = 3;
  double sortino = 4;
  double max_drawdown = 5;
  double calmar = 6;
  int32  num_trades = 7;
  double win_rate = 8;
  double profit_factor = 9;
  double avg_win = 10;
  double avg_loss = 11;
}

message TradeRecord {
  string symbol = 1;
  string side = 2;          // long/short
  string entry_time = 3;
  double entry_price = 4;
  string exit_time = 5;
  double exit_price = 6;
  double pnl_usd = 7;
  double pnl_pct = 8;
  double mfe_pct = 9;       // max favorable excursion
  double mae_pct = 10;      // max adverse excursion
  string exit_reason = 11;  // tp/sl/signal/time
}

message EquityPoint { string time = 1; double equity = 2; double drawdown = 3; }
```

---

## 4. 引擎核心

文件：`backend/internal/service/backtest/engine.go`

```go
type Engine struct {
    price PriceProvider
    cost  CostModel
    risk  RiskConfig
    log   *slog.Logger
    cancelCheck func() bool
}

type Position struct {
    Symbol   string
    Side     string
    Qty      float64
    EntryPrice float64
    EntryTime  time.Time
    StopPrice  float64
    TargetPrice float64
}

type Engine_Run_Output struct {
    Trades      []TradeRecord
    Equity      []EquityPoint
    Summary     BacktestSummary
    Advanced    AdvancedMetrics
}

func (e *Engine) Run(ctx context.Context, cfg QuantBtConfig, strat StrategyAdapter) (*Engine_Run_Output, error) {
    bars := loadBars(cfg)            // map[symbol][]Bar，按时间对齐
    if not enough: return ErrDataInsufficient

    capital := cfg.RiskConfig.InitialCapital
    var openPositions []Position
    var trades []TradeRecord
    var equity []EquityPoint

    for ts := range alignedTimeline(bars) {
        if e.cancelCheck() { return nil, ErrCanceled }

        // 1. 更新 MFE/MAE
        for p in openPositions: updateExcursion(p, bars[p.Symbol][ts])

        // 2. 检查止盈/止损/时间止
        for p in openPositions:
            if shouldExit(p, bars, e.cost):
                trades.append(closePosition(p, ...))
                capital += pnl

        // 3. 调用策略获取新信号
        signals := strat.OnBar(ts, bars)

        // 4. 仓位分配（risk parity / fixed risk）
        for sig in signals:
            if len(openPositions) >= cfg.RiskConfig.MaxConcurrent: break
            qty := sizing.FixedRisk(capital, cfg.RiskConfig.RiskPerTrade,
                                    sig.StopDistance)
            entryPrice := bars[sig.Symbol][ts].Close * (1 + e.cost.SlippageBps/10000)
            openPositions.append(Position{...})
            capital -= entryPrice * qty * e.cost.Commission

        // 5. 记权益
        unrealized := sum(mtm(p, bars[p.Symbol][ts]) for p in openPositions)
        equity.append({Time: ts, Equity: capital+unrealized, ...})

        // 6. 全局风控：MaxDrawdown 触发停损
        if drawdown(equity) > cfg.RiskConfig.MaxDrawdownStop:
            forceCloseAll(); break
    }

    // 收尾：强制平仓所有未平仓位
    for p in openPositions: trades.append(closePosition(p, last_bar))

    return composeOutput(trades, equity)
}
```

### 4.1 StrategyAdapter 接口

```go
type StrategyAdapter interface {
    Name() string
    InitParams(params map[string]float64) error
    OnBar(ctx context.Context, ts time.Time, bars map[string][]Bar) []TradeSignal
}

type TradeSignal struct {
    Symbol     string
    Side       string  // long/short
    StopDistance float64  // 入场价的百分比
    TargetDistance float64
    TimeStop time.Duration  // 0 表示无
    Reason   string         // 用于日志
}
```

策略实现位于 `backend/internal/service/backtest/strategies/`，每个策略一个文件：
```
strategies/momentum_12_3.go
strategies/mean_reversion_5d.go
strategies/breakout_donchian_20.go
strategies/low_vol_60d.go
```

策略注册：
```go
func init() {
    Registry.Register("momentum_12_3", func() StrategyAdapter { return &MomentumStrategy{} })
    ...
}
```

---

## 5. 异步执行

### 5.1 提交（API 进程）

```go
func (s *Service) RunQuantBt(ctx context.Context, cfg *QuantBtConfig) (*RunQuantBtResponse, error) {
    if !Registry.Has(cfg.Strategy):
        return nil, errors.New("unknown strategy")
    if !validateRange(cfg): ...

    job := BacktestJob{
        JobID: uuid.NewString(),
        UserID: userIDFrom(ctx),
        Type: "quantbt",
        Request: marshal(cfg),
        Status: "queued",
        CreatedAt: time.Now(),
    }
    s.repo.Insert(ctx, job)
    s.queue.Push(ctx, "backtest:queue", job.JobID)

    return &RunQuantBtResponse{TaskID: job.JobID, Status: "queued"}, nil
}
```

### 5.2 执行（Worker 进程）

文件：`backend/cmd/antclaw-worker/backtest_runner.go`

```go
func runBacktestWorker(ctx context.Context, deps Deps) {
    pool := workerpool.New(4)   // 并发 4 个回测
    for {
        msg := redisStream.Read("backtest:queue")
        pool.Submit(func() { executeJob(ctx, msg.JobID, deps) })
    }
}

func executeJob(ctx, jobID, deps) {
    deps.repo.UpdateStatus(jobID, "running", started_at=now)

    cancelCh := watchCancel(jobID)
    engine := backtest.NewEngine(deps.price, ..., cancelCh)
    cfg := loadJob(jobID).Request
    strat := Registry.New(cfg.Strategy)
    strat.InitParams(cfg.Params)

    out, err := engine.Run(ctx, cfg, strat)
    if err != nil:
        deps.repo.UpdateFailed(jobID, err.Error()); return
    deps.repo.SaveResult(jobID, out)
    deps.repo.UpdateStatus(jobID, "done", completed_at=now)
}
```

### 5.3 进度上报

每处理 5% 进度更新一次 `backtest_jobs.progress`，前端轮询。

---

## 6. 修改清单

| 文件 | 动作 |
|------|------|
| `proto/antclaw/v1/backtest.proto` | 扩展（§3） |
| `backend/internal/service/backtest/engine.go` | 新建 |
| `backend/internal/service/backtest/types.go` | 新建 |
| `backend/internal/service/backtest/registry.go` | 新建 |
| `backend/internal/service/backtest/strategies/*.go` | 新建（4-5 个） |
| `backend/internal/service/backtest/sizing/fixed_risk.go` | 新建 |
| `backend/internal/service/backtest/exits.go` | 新建（止盈止损时间止）|
| `backend/internal/service/backtest/cost.go` | 新建（复用文档 10）|
| `backend/internal/service/backtest/service.go` | 重写 RunQuantBt + GetJob/CancelJob/ListJobs/GetResult |
| `backend/internal/adapter/storage/postgres/backtest_repo.go` | 新建 |
| `backend/cmd/antclaw-worker/backtest_runner.go` | 新建 |
| `backend/cmd/antclaw-worker/main.go` | 启动 backtest worker pool |
| `backend/internal/adapter/rpc/backtest_handler.go` | 扩展 4 个新 RPC |
| `backend/internal/service/backtest/engine_test.go` | 新建（含金标准对比） |

---

## 7. 单测策略

1. **engine_test.go**：用 fixture bars + Mock StrategyAdapter（强制 long），断言：
   - 入场价 = bar.Close * (1 + slippage_bps/10000)
   - 止损触发 → exit_reason="sl"
   - 时间止触发
   - MaxDrawdown 触发停损
2. **strategies/*_test.go**：每个策略喂入历史数据，断言信号方向。
3. **integration**：用 dockertest + 真实 Postgres，端到端 RunQuantBt → 轮询 GetJob → GetResult。

---

## 8. 验证

```bash
# 1. 提交回测
JOB=$(curl -s http://localhost:8082/antclaw.v1.BacktestService/RunQuantBt \
  -H 'Content-Type: application/json' \
  -d '{"config":{
        "strategy":"momentum_12_3",
        "symbols":["EURUSD"],
        "from_date":"2022-01-01","to_date":"2025-01-01",
        "timeframe":"1d",
        "cost_model":{"slippage_bps":2,"spread_bps":1,"commission_per_trade":2},
        "risk":{"initial_capital":100000,"risk_per_trade":0.01}
      }}' | jq -r .task_id)

# 2. 轮询直到 done
for i in $(seq 1 60); do
  STATUS=$(curl -s http://localhost:8082/antclaw.v1.BacktestService/GetJob \
    -d "{\"job_id\":\"$JOB\"}" -H 'Content-Type: application/json' | jq -r .job.status)
  echo "$i status=$STATUS"
  [ "$STATUS" = "done" ] && break
  [ "$STATUS" = "failed" ] && exit 1
  sleep 2
done

# 3. 拉取结果
curl -s http://localhost:8082/antclaw.v1.BacktestService/GetResult \
  -d "{\"job_id\":\"$JOB\"}" -H 'Content-Type: application/json' \
  | jq '.summary, (.trades | length), (.equity_curve | length)'
```

预期：
- summary 包含 sharpe/maxdd 等真实数值
- trades > 0
- equity_curve 长度 ≈ 交易日数

---

## 9. 完成判定

- [ ] 5 个策略 adapter 全部实现且单测通过
- [ ] engine_test 覆盖率 ≥ 85%
- [ ] 提交→执行→拉取结果链路通畅
- [ ] 取消任务能在 ≤ 1 个 bar 内停止
- [ ] 进度更新可见
- [ ] 4 个并发任务不互相阻塞

## 10. 实施记录

<!-- -->
