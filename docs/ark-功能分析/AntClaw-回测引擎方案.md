# AntClaw 回测引擎方案

> **版本**：v1.0  
> **对应 ark-intelligent 模块**：`internal/service/backtest/` (37 files), `quantbt/`, `vpbt/`  
> **对应 AntClaw Proto**：`BacktestService`

---

## 一、ark-intelligent 方法分析

### 1.1 核心职责
回测引擎是 ark-intelligent **最复杂的模块**（37 个文件），实现了从单信号回测到 walk-forward 优化、蒙特卡洛、组合分析、归因、信号校准等机构级能力。

### 1.2 模块矩阵

#### A. 核心评估（evaluator + stats）

| 文件 | 职责 |
|------|------|
| `evaluator.go` | 主回测器：取过去信号，每个信号在 1W/2W/4W 后查价格，计算 win/loss/return |
| `stats.go` | 统计指标：win rate, avg win/loss, Sharpe, Sortino, Calmar, max DD, profit factor, expectancy |
| `dedup.go` | 信号去重（同合约 7 天内同方向只保留首发）|
| `costs.go` | 交易成本：Default spreads bps（EUR 1bp、JPY 1bp、NZD 3bp...），round-trip = 2× spread |

#### B. 高级回测

| 文件 | 职责 |
|------|------|
| `walkforward.go` | 滚动训练-测试窗口（in-sample 优化 → out-of-sample 验证）|
| `walkforward_optimizer.go` | **权重自动调优**：滚动调整子系统权重以最大化 OOS Sharpe |
| `walkforward_multi.go` | 多 symbol 同时 walk-forward |
| `bootstrap.go` | Bootstrap 重采样：从历史信号中重采样 N 次估计置信区间 |
| `montecarlo.go` | MC 模拟：路径乱序 + 收益重采样估计统计分布 |
| `portfolio.go` | 组合层面：等权周度信号 → 等价组合曲线 |

#### C. 校准与归因

| 文件 | 职责 |
|------|------|
| `logistic_calibration.go` | **Logistic 回归**校准信号 → 真实概率（输入: signal score 输出: P(win)）|
| `factor_decomposition.go` | 收益归因到 COT/CTA/Quant/Sentiment 各因子的贡献 |
| `decay.go` | 信号衰减分析：每天后 win rate / avg return 的变化曲线 |
| `timing.go` | 信号最佳持仓时长（按 signal type 找峰值 Sharpe 的 horizon）|

#### D. 风险管理

| 文件 | 职责 |
|------|------|
| `ruin.go` | Risk of Ruin：根据 win rate / avg win/loss / max risk 计算破产概率 |
| `excursion.go` | MAE / MFE：最大不利偏移 / 最大有利偏移分布 |
| `matrix.go` | signalType × regime 矩阵：每种组合下的胜率 |

#### E. 战略分析

| 文件 | 职责 |
|------|------|
| `strategy_composer.go` | **TASK-139** 多策略组合：把 N 个独立策略权重组合，最大化夏普 |
| `weights.go` | 信号权重的多种组合算法（等权、Sharpe-weighted、Kelly、Risk parity）|
| `baseline.go` | 与随机基线对比（MC 验证策略真实优势）|
| `report.go` | 周度报告生成 |
| `audit.go` / `audit_*test.go` | 强一致性回测审计（防止 lookahead bias）|

#### F. 辅助

| 文件 | 职责 |
|------|------|
| `daily_trend_filter.go` | COT 信号 + 日线趋势对齐过滤器 |
| `trend_filter_stats.go` | trend filter 命中率统计 |
| `regime_backfill.go` | 历史 regime 回填，用于按状态切片回测 |
| `smart_money.go` | smart money（杠杆基金）持仓变化与后续价格相关 |

#### G. QuantBT / VPBT

- `quantbt/`：调用 GARCH/HMM/Hurst 模型的回测专用版本（量化模型回测）
- `vpbt/`：Volume Profile 策略的专用回测（POC、VA 触碰策略）

---

## 二、AntClaw 设计方案

### 2.1 架构

```
BacktestService (Proto, 已存在)
  ├── service/backtest/core         (evaluator, stats, costs, dedup)
  ├── service/backtest/walkforward  (walkforward + optimizer)
  ├── service/backtest/calibration  (logistic regression)
  ├── service/backtest/montecarlo   (MC + bootstrap)
  ├── service/backtest/decomp       (factor decomposition)
  ├── service/backtest/risk         (ruin, excursion)
  ├── service/backtest/portfolio    (multi-strategy composer)
  ├── service/backtest/quant        (GARCH/HMM 回测)
  └── service/backtest/vp           (Volume Profile 回测)
      ↓
  infra/postgres/{signal_repo,backtest_repo,price_repo}.go
  infra/postgres/backtest_results.go  (持久化大型结果)
  worker/backtest/runner.go            (异步回测任务)
```

**架构特点**：
- 回测 = **独立后台 worker** 任务，不阻塞 API
- 大结果（如 walk-forward 全曲线）存 PostgreSQL JSONB；小结果走 Redis
- 支持 **streaming 进度**（SSE）

### 2.2 核心接口

```go
type BacktestRunner interface {
    Run(ctx, req BacktestRequest) (jobID string, err error)
    GetStatus(ctx, jobID string) (*BacktestStatus, error)
    GetResult(ctx, jobID string) (*BacktestResult, error)
    Cancel(ctx, jobID string) error
}

type Evaluator interface {
    Evaluate(ctx, signals []Signal, horizons []time.Duration) (*EvalResult, error)
}

type Calibrator interface {
    FitLogistic(ctx, signals []SignalWithOutcome) (*LogisticParams, error)
    Calibrate(rawScore float64, params LogisticParams) float64
}

type WalkForwardOptimizer interface {
    Optimize(ctx, params WFParams) (*WFResult, error)
}

type RiskAnalyzer interface {
    RiskOfRuin(stats Stats) (*RuinResult, error)
    ExcursionStats(ctx, signals []Signal) (*ExcursionStats, error)
}
```

### 2.3 Schema

```sql
-- 回测任务（异步）
CREATE TABLE backtest_jobs (
    job_id        UUID PRIMARY KEY,
    user_id       VARCHAR(64),
    type          VARCHAR(32),    -- 'evaluator','walkforward','montecarlo','composer'
    request       JSONB,
    status        VARCHAR(16),     -- 'queued','running','done','failed','canceled'
    progress      DOUBLE PRECISION,
    created_at    TIMESTAMPTZ,
    started_at    TIMESTAMPTZ,
    completed_at  TIMESTAMPTZ,
    error         TEXT
);

-- 回测结果
CREATE TABLE backtest_results (
    job_id        UUID PRIMARY KEY REFERENCES backtest_jobs(job_id),
    summary       JSONB,            -- 关键指标（win rate, sharpe, ...）
    equity_curve  JSONB,            -- [{time, equity}]
    trades        JSONB,
    detailed      JSONB,            -- 全量明细
    created_at    TIMESTAMPTZ DEFAULT NOW()
);

-- 信号校准参数（持久化 logistic 回归结果）
CREATE TABLE signal_calibration (
    signal_type   VARCHAR(64) PRIMARY KEY,
    logistic_a    DOUBLE PRECISION,
    logistic_b    DOUBLE PRECISION,
    sample_size   INT,
    win_rate      DOUBLE PRECISION,
    avg_return    DOUBLE PRECISION,
    updated_at    TIMESTAMPTZ
);

-- Walk-forward 优化历史
CREATE TABLE walkforward_history (
    id            BIGSERIAL PRIMARY KEY,
    fold_idx      INT,
    train_from    DATE, train_to DATE,
    test_from     DATE, test_to DATE,
    optimal_weights JSONB,
    in_sample_sharpe DOUBLE PRECISION,
    oos_sharpe       DOUBLE PRECISION,
    created_at    TIMESTAMPTZ
);
```

### 2.4 Redis

| Key | TTL | 内容 |
|-----|-----|------|
| `bt:job:{job_id}:progress` | 1h | 实时进度 |
| `bt:job:{job_id}:logs` | 1h | 日志流 |
| `cache:bt:summary:{user}` | 10m | 用户最近回测摘要 |
| `pubsub:bt:job:{job_id}` | - | SSE 推送 |

### 2.5 Worker 调度

```
[API] User → submit job → Postgres queue
[Worker] poll queue → run → write result → publish SSE
[Worker pool] 可水平扩展
```

- 长任务（walk-forward 全市场）可分片到多个 worker
- 失败重试：transient 失败（DB 错误）3 次重试；逻辑错误立即 fail
- 资源隔离：每个 worker 限 CPU/Memory

### 2.6 优化与提升

| 维度 | ark-intelligent | AntClaw |
|------|----------------|---------|
| 调用方式 | 同步 (Telegram cmd) | 异步 worker + SSE 进度 |
| 持久化 | 内存结果 + 截图 | 完整 JSONB + 可分享 |
| 并行度 | 单进程 | Worker 池水平扩展 |
| 校准 | 信号回测时算 | 独立日任务 + DB 持久化 |
| 防 look-ahead | audit_test 验证 | 同左 + DB schema 强约束（report_date 严格小于 evaluated_at）|
| 用户隔离 | 单 bot | 多用户，按 user_id 配额 |
| 大结果传输 | 文本输出 | JSONB + 前端流式渲染 |

---

## 三、参考文件

- ark-intelligent：`internal/service/backtest/*.go`（37 文件），`quantbt/`, `vpbt/`
- AntClaw proto：`proto/antclaw/v1/backtest.proto`
- AntClaw service：`backend/internal/service/backtest/`
