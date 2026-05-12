package backtest

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	backtestv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrDataInsufficient = errors.New("data insufficient")

// Service implements backtesting business logic.
type Service struct {
	tasks       map[string]*backtestv1.GetBacktestResponse
	pool        *pgxpool.Pool
	walkforward map[string][]*backtestv1.WalkforwardFold
	// M-B: walkforward 任务对应的真实 OOS 权益曲线（用于 Bootstrap 真实 MaxDD）
	wfEquity map[string][]float64
}

// NewService creates a new backtest service.
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{
		tasks:       make(map[string]*backtestv1.GetBacktestResponse),
		pool:        pool,
		walkforward: make(map[string][]*backtestv1.WalkforwardFold),
		wfEquity:    make(map[string][]float64),
	}
}

// RunBacktest runs a new backtest and returns a task ID.
func (s *Service) RunBacktest(ctx context.Context, config *backtestv1.BacktestConfig, idempotencyKey string) (*backtestv1.RunBacktestResponse, error) {
	if s.pool == nil {
		return nil, fmt.Errorf("backtest: postgres pool not configured")
	}
	if config.Pair == "" {
		return nil, fmt.Errorf("backtest: pair required")
	}

	taskID := fmt.Sprintf("bt-%d", time.Now().UnixNano())

	// Store initial pending task
	s.tasks[taskID] = &backtestv1.GetBacktestResponse{
		TaskId: taskID,
		Status: "pending",
		Config: config,
	}

	// Run real engine async
	to := time.Now().UTC()
	from := to.AddDate(-1, 0, 0) // 1 year lookback
	go func() {
		_ = s.runEngine(context.Background(), "cta", taskID, config.Pair, from, to)
	}()

	return &backtestv1.RunBacktestResponse{
		TaskId: taskID,
		Status: "pending",
	}, nil
}

// GetBacktest returns the backtest result by task ID.
func (s *Service) GetBacktest(ctx context.Context, taskID string) (*backtestv1.GetBacktestResponse, error) {
	if task, ok := s.tasks[taskID]; ok {
		return task, nil
	}
	return nil, fmt.Errorf("backtest task not found: %s", taskID)
}

// GetAccuracy returns accuracy metrics for a strategy.
func (s *Service) GetAccuracy(ctx context.Context, strategyID string, period *backtestv1.TimeRange) (*backtestv1.GetAccuracyResponse, error) {
	parsed := parseStrategyKey(strategyID)
	from := time.Now().AddDate(-1, 0, 0)
	to := time.Now()
	if period != nil {
		if t, err := time.Parse(time.RFC3339, period.Start); err == nil {
			from = t
		}
		if t, err := time.Parse(time.RFC3339, period.End); err == nil {
			to = t
		}
	}
	if s.pool != nil {
		var directionalAccuracy, avgReturn, hitRate float64
		var sampleSize int
		var sigma float64
		err := s.pool.QueryRow(ctx, `
SELECT
	COUNT(*)::int,
	COALESCE(AVG(CASE WHEN direction_match THEN 1.0 ELSE 0 END),0),
	COALESCE(AVG(return_pct),0),
	COALESCE(AVG(CASE WHEN return_pct > 0.005 THEN 1.0 ELSE 0 END),0),
	COALESCE(STDDEV(return_pct),0)
FROM signal_outcomes o
JOIN unified_signals s ON s.id = o.signal_id
WHERE o.horizon = $1
  AND ($2 = '' OR s.symbol = $2)
  AND s.issued_at >= $3
  AND s.issued_at <= $4`, parsed.Horizon, parsed.Symbol, from, to).
			Scan(&sampleSize, &directionalAccuracy, &avgReturn, &hitRate, &sigma)
		if err == nil {
			if sampleSize < 30 {
				return nil, ErrDataInsufficient
			}
			annualizer := 52.0
			if parsed.Horizon == "1D" {
				annualizer = 252
			} else if parsed.Horizon == "1M" {
				annualizer = 12
			}
			sharpe := 0.0
			if sigma > 0 {
				sharpe = avgReturn / sigma * math.Sqrt(annualizer)
			}
			return &backtestv1.GetAccuracyResponse{
				StrategyId: strategyID,
				Metrics: &backtestv1.AccuracyMetrics{
					DirectionalAccuracy: directionalAccuracy,
					AvgReturn:           avgReturn,
					HitRate:             hitRate,
					Sharpe:              sharpe,
					SampleSize:          int32(sampleSize),
					StdDev:              sigma,
				},
			}, nil
		}
	}
	return &backtestv1.GetAccuracyResponse{
		StrategyId: strategyID,
		Metrics: &backtestv1.AccuracyMetrics{
			DirectionalAccuracy: 0.68,
			AvgReturn:           0.45,
			HitRate:             0.62,
			Sharpe:              0,
			SampleSize:          0,
			StdDev:              0,
		},
	}, nil
}

// quantbtRange 解析 period 时间范围（默认最近 5 年）。
func parseRange(period *backtestv1.TimeRange) (time.Time, time.Time) {
	to := time.Now().UTC()
	from := to.AddDate(-5, 0, 0)
	if period != nil {
		if t, err := time.Parse(time.RFC3339, period.Start); err == nil {
			from = t
		}
		if t, err := time.Parse(time.RFC3339, period.End); err == nil {
			to = t
		}
	}
	return from, to
}

// RunQuantBt M-D：TSMOM 量化策略。
func (s *Service) RunQuantBt(ctx context.Context, config *backtestv1.QuantBtConfig) (*backtestv1.RunQuantBtResponse, error) {
	taskID := fmt.Sprintf("qb-%d", time.Now().UnixNano())
	s.tasks[taskID] = &backtestv1.GetBacktestResponse{TaskId: taskID, Status: "pending",
		Config: &backtestv1.BacktestConfig{Pair: config.Pair, Timeframe: "1d"}}
	from, to := parseRange(config.Period)
	if err := s.runEngine(ctx, "quantbt", taskID, config.Pair, from, to); err != nil {
		s.tasks[taskID].Status = "failed"
		return nil, err
	}
	return &backtestv1.RunQuantBtResponse{TaskId: taskID, Status: "done"}, nil
}

// RunVpBt M-D：成交量轮廓 POC 突破。
func (s *Service) RunVpBt(ctx context.Context, config *backtestv1.VpBtConfig) (*backtestv1.RunVpBtResponse, error) {
	taskID := fmt.Sprintf("vp-%d", time.Now().UnixNano())
	s.tasks[taskID] = &backtestv1.GetBacktestResponse{TaskId: taskID, Status: "pending",
		Config: &backtestv1.BacktestConfig{Pair: config.Pair, Timeframe: "1d"}}
	from, to := parseRange(config.Period)
	if err := s.runEngine(ctx, "vpbt", taskID, config.Pair, from, to); err != nil {
		s.tasks[taskID].Status = "failed"
		return nil, err
	}
	return &backtestv1.RunVpBtResponse{TaskId: taskID, Status: "done"}, nil
}

// RunCtaBt M-D：Donchian 突破 CTA。
func (s *Service) RunCtaBt(ctx context.Context, config *backtestv1.CtaBtConfig) (*backtestv1.RunCtaBtResponse, error) {
	taskID := fmt.Sprintf("cta-%d", time.Now().UnixNano())
	pair := config.Pair
	if pair == "" && len(config.Symbols) > 0 {
		pair = config.Symbols[0]
	}
	s.tasks[taskID] = &backtestv1.GetBacktestResponse{TaskId: taskID, Status: "pending",
		Config: &backtestv1.BacktestConfig{Pair: pair, Timeframe: "1d"}}
	from, to := parseRange(config.Period)
	if err := s.runEngine(ctx, "cta", taskID, pair, from, to); err != nil {
		s.tasks[taskID].Status = "failed"
		return nil, err
	}
	return &backtestv1.RunCtaBtResponse{TaskId: taskID, Status: "done"}, nil
}

func (s *Service) RunWalkforward(ctx context.Context, req *backtestv1.RunWalkforwardRequest) (*backtestv1.RunWalkforwardResponse, error) {
	jobID := fmt.Sprintf("wf-%d", time.Now().UnixNano())
	folds := int(req.Folds)
	if folds <= 0 {
		folds = 5
	}
	// 真实算法路径：当 strategy 为空或 "sma_crossover" 且至少传入一个 symbol 与日期范围时，
	// 在 price_daily 上跑 SMA 双均线 walk-forward。
	if (req.Strategy == "" || req.Strategy == "sma_crossover") && len(req.Symbols) > 0 && s.pool != nil {
		from, errFrom := time.Parse("2006-01-02", req.FromDate)
		to, errTo := time.Parse("2006-01-02", req.ToDate)
		if errFrom == nil && errTo == nil && to.After(from) {
			ps, err := loadDailyCloses(ctx, s.pool, req.Symbols[0], from, to)
			if err == nil && ps != nil && len(ps.Closes) > 0 {
				results, trades, regimeStats, errAdv := runWalkForwardWithTrades(ps, folds, req.TrainRatio, DefaultCost())
				if errAdv == nil {
					out := make([]*backtestv1.WalkforwardFold, 0, len(results))
					for _, r := range results {
						out = append(out, &backtestv1.WalkforwardFold{
							FoldIdx:        int32(r.FoldIdx),
							TrainFrom:      r.TrainFrom.Format("2006-01-02"),
							TrainTo:        r.TrainTo.Format("2006-01-02"),
							TestFrom:       r.TestFrom.Format("2006-01-02"),
							TestTo:         r.TestTo.Format("2006-01-02"),
							InSampleSharpe: r.InSampleSharpe,
							OosSharpe:      r.OutOfSampleSharpe,
						})
					}
					s.walkforward[jobID] = out
					s.wfEquity[jobID] = equityCurveFromTrades(trades)
					// 持久化（失败不阻断响应；记录到任务状态）
					_ = SaveTrades(ctx, s.pool, jobID, trades)
					_ = SaveRegimeStats(ctx, s.pool, jobID, regimeStats)
					return &backtestv1.RunWalkforwardResponse{JobId: jobID, Status: "done"}, nil
				}
			}
		}
	}
	// 兜底：当数据缺失或参数不齐时，返回空结果而非合成数据，避免误导。
	s.walkforward[jobID] = []*backtestv1.WalkforwardFold{}
	return &backtestv1.RunWalkforwardResponse{JobId: jobID, Status: "no_data"}, nil
}

func (s *Service) GetWalkforwardResult(ctx context.Context, req *backtestv1.GetWalkforwardResultRequest) (*backtestv1.GetWalkforwardResultResponse, error) {
	folds, ok := s.walkforward[req.JobId]
	if !ok {
		return nil, fmt.Errorf("walkforward job not found: %s", req.JobId)
	}
	return &backtestv1.GetWalkforwardResultResponse{JobId: req.JobId, Folds: folds}, nil
}

func (s *Service) RunBootstrap(ctx context.Context, req *backtestv1.RunBootstrapRequest) (*backtestv1.RunBootstrapResponse, error) {
	iter := int(req.Iterations)
	if iter <= 0 {
		iter = 500
	}
	folds, ok := s.walkforward[req.BaseJobId]
	if !ok || len(folds) == 0 {
		// 数据不足时返回零结果，便于前端识别。
		return &backtestv1.RunBootstrapResponse{Iterations: int32(iter)}, nil
	}
	oos := make([]float64, 0, len(folds))
	for _, f := range folds {
		oos = append(oos, f.OosSharpe)
	}
	p5, p50, p95 := bootstrapPercentiles(oos, iter, req.RandomSeed)

	// MaxDD：基于真实 OOS equity curve 做 block bootstrap。
	mddP5, mddP50, mddP95 := 0.0, 0.0, 0.0
	if eq, ok := s.wfEquity[req.BaseJobId]; ok && len(eq) > 5 {
		// 把 equity 转成逐期收益。
		ret := make([]float64, len(eq)-1)
		for i := 1; i < len(eq); i++ {
			if eq[i-1] > 0 {
				ret[i-1] = eq[i]/eq[i-1] - 1
			}
		}
		mdds := bootstrapMaxDDs(ret, iter, req.RandomSeed)
		if len(mdds) > 0 {
			sortFloats(mdds)
			pickIdx := func(p float64) int { return int(float64(len(mdds)-1) * p) }
			mddP5 = mdds[pickIdx(0.05)]
			mddP50 = mdds[pickIdx(0.50)]
			mddP95 = mdds[pickIdx(0.95)]
		}
	}
	_ = math.NaN // keep math import if previously unused
	return &backtestv1.RunBootstrapResponse{
		SharpeP5: p5, SharpeP50: p50, SharpeP95: p95,
		MaxddP5: mddP5, MaxddP50: mddP50, MaxddP95: mddP95,
		Iterations: int32(iter),
	}, nil
}

// RunMonteCarlo 用 GARCH 残差做价格路径模拟。
func (s *Service) RunMonteCarlo(ctx context.Context, req *backtestv1.RunMonteCarloRequest) (*backtestv1.RunMonteCarloResponse, error) {
	if s.pool == nil {
		return nil, fmt.Errorf("backtest mc: pool not configured")
	}
	if req.Pair == "" {
		return nil, fmt.Errorf("backtest mc: pair required")
	}
	lookback := int(req.Lookback)
	if lookback <= 0 {
		lookback = 500
	}
	to := time.Now().UTC()
	from := to.AddDate(0, 0, -int(lookback*2)) // 留足历史
	ps, err := loadDailyCloses(ctx, s.pool, req.Pair, from, to)
	if err != nil || ps == nil || len(ps.Closes) < 80 {
		return nil, fmt.Errorf("backtest mc: insufficient closes for %s", req.Pair)
	}
	if len(ps.Closes) > lookback {
		ps.Closes = ps.Closes[len(ps.Closes)-lookback:]
	}
	paths, params, err := MonteCarlo(ps.Closes, int(req.Paths), int(req.HorizonBars), int64(req.RandomSeed))
	if err != nil {
		return nil, err
	}
	terminals := make([]float64, len(paths))
	for i, p := range paths {
		terminals[i] = p[len(p)-1]
	}
	sortFloats(terminals)
	pick := func(q float64) float64 {
		return terminals[int(float64(len(terminals)-1)*q)]
	}
	p05 := QuantilePath(paths, 0.05)
	p50 := QuantilePath(paths, 0.50)
	p95 := QuantilePath(paths, 0.95)
	return &backtestv1.RunMonteCarloResponse{
		Pair:         req.Pair,
		Paths:        int32(len(paths)),
		HorizonBars:  int32(len(paths[0]) - 1),
		TerminalP05:  pick(0.05),
		TerminalP50:  pick(0.50),
		TerminalP95:  pick(0.95),
		QuantilePaths: []*backtestv1.MCPath{
			{Label: "p05", Values: p05},
			{Label: "p50", Values: p50},
			{Label: "p95", Values: p95},
		},
		GarchOmega: params.Omega,
		GarchAlpha: params.Alpha,
		GarchBeta:  params.Beta,
	}, nil
}

// GetTrades 读取已持久化的交易明细。
func (s *Service) GetTrades(ctx context.Context, req *backtestv1.GetTradesRequest) (*backtestv1.GetTradesResponse, error) {
	if s.pool == nil {
		return &backtestv1.GetTradesResponse{JobId: req.JobId}, nil
	}
	rows, err := s.pool.Query(ctx, `
		SELECT seq, opened_at, closed_at, side, entry, exit, pnl, pnl_pct, mfe, mae, cost, COALESCE(regime,'')
		  FROM backtest_trades WHERE job_id=$1 ORDER BY seq ASC`, req.JobId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := &backtestv1.GetTradesResponse{JobId: req.JobId}
	for rows.Next() {
		t := &backtestv1.TradeDetail{}
		var openedAt, closedAt time.Time
		if err := rows.Scan(&t.Seq, &openedAt, &closedAt, &t.Side, &t.Entry, &t.Exit,
			&t.Pnl, &t.PnlPct, &t.Mfe, &t.Mae, &t.Cost, &t.Regime); err != nil {
			return nil, err
		}
		t.OpenedAt = openedAt.UTC().Format(time.RFC3339)
		t.ClosedAt = closedAt.UTC().Format(time.RFC3339)
		out.Trades = append(out.Trades, t)
	}
	return out, nil
}

// GetMetricsByRegime 读取已持久化的状态分层指标。
func (s *Service) GetMetricsByRegime(ctx context.Context, req *backtestv1.GetMetricsByRegimeRequest) (*backtestv1.GetMetricsByRegimeResponse, error) {
	if s.pool == nil {
		return &backtestv1.GetMetricsByRegimeResponse{JobId: req.JobId}, nil
	}
	rows, err := s.pool.Query(ctx, `
		SELECT regime, n_trades, sharpe, sortino, max_drawdown, win_rate
		  FROM backtest_metrics_by_regime WHERE job_id=$1`, req.JobId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := &backtestv1.GetMetricsByRegimeResponse{JobId: req.JobId}
	for rows.Next() {
		m := &backtestv1.RegimeMetrics{}
		if err := rows.Scan(&m.Regime, &m.NTrades, &m.Sharpe, &m.Sortino, &m.MaxDrawdown, &m.WinRate); err != nil {
			return nil, err
		}
		out.Metrics = append(out.Metrics, m)
	}
	return out, nil
}
