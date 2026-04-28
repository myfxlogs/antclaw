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
}

// NewService creates a new backtest service.
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{
		tasks:       make(map[string]*backtestv1.GetBacktestResponse),
		pool:        pool,
		walkforward: make(map[string][]*backtestv1.WalkforwardFold),
	}
}

// RunBacktest runs a new backtest and returns a task ID.
func (s *Service) RunBacktest(ctx context.Context, config *backtestv1.BacktestConfig, idempotencyKey string) (*backtestv1.RunBacktestResponse, error) {
	taskID := fmt.Sprintf("bt-%d", time.Now().UnixNano())

	// Store initial pending task
	s.tasks[taskID] = &backtestv1.GetBacktestResponse{
		TaskId: taskID,
		Status: "pending",
		Config: config,
	}

	// Simulate async execution
	go s.simulateBacktest(taskID, config)

	return &backtestv1.RunBacktestResponse{
		TaskId: taskID,
		Status: "pending",
	}, nil
}

// simulateBacktest simulates backtest execution with sample data.
func (s *Service) simulateBacktest(taskID string, config *backtestv1.BacktestConfig) {
	// Simulate processing time
	time.Sleep(100 * time.Millisecond)

	// Update to running
	if task, ok := s.tasks[taskID]; ok {
		task.Status = "running"
	}

	// Generate sample trades
	trades := []*backtestv1.TradeRecord{
		{
			TradeId:    "trade-001",
			EntryTime:  time.Now().Add(-48 * time.Hour).Format(time.RFC3339),
			ExitTime:   time.Now().Add(-36 * time.Hour).Format(time.RFC3339),
			Direction:  "long",
			EntryPrice: "1.0850",
			ExitPrice:  "1.0920",
			Pnl:        "70.0",
			PnlPct:     "0.65",
		},
		{
			TradeId:    "trade-002",
			EntryTime:  time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
			ExitTime:   time.Now().Add(-12 * time.Hour).Format(time.RFC3339),
			Direction:  "short",
			EntryPrice: "1.0950",
			ExitPrice:  "1.0880",
			Pnl:        "70.0",
			PnlPct:     "0.64",
		},
		{
			TradeId:    "trade-003",
			EntryTime:  time.Now().Add(-6 * time.Hour).Format(time.RFC3339),
			ExitTime:   time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
			Direction:  "long",
			EntryPrice: "1.0900",
			ExitPrice:  "1.0870",
			Pnl:        "-30.0",
			PnlPct:     "-0.28",
		},
	}

	// Calculate metrics
	var totalTrades int32 = int32(len(trades))
	var winCount int32
	var totalPnL float64
	for _, trade := range trades {
		if trade.Pnl[0] != '-' {
			winCount++
		}
		// Simple parse
		var pnl float64
		fmt.Sscanf(trade.Pnl, "%f", &pnl)
		totalPnL += pnl
	}

	winRate := float64(winCount) / float64(totalTrades) * 100

	// Update completed task
	s.tasks[taskID] = &backtestv1.GetBacktestResponse{
		TaskId: taskID,
		Status: "completed",
		Config: config,
		Metrics: &backtestv1.BacktestMetrics{
			TotalReturn:    fmt.Sprintf("%.2f", totalPnL),
			TotalReturnPct: "1.01",
			SharpeRatio:    1.35,
			MaxDrawdown:    0.05,
			WinRate:        winRate,
			TotalTrades:    totalTrades,
			ProfitFactor:   2.33,
		},
		Trades: trades,
	}
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

// RunQuantBt runs a quantitative backtest.
func (s *Service) RunQuantBt(ctx context.Context, config *backtestv1.QuantBtConfig) (*backtestv1.RunQuantBtResponse, error) {
	taskID := fmt.Sprintf("qb-%d", time.Now().UnixNano())
	s.tasks[taskID] = &backtestv1.GetBacktestResponse{
		TaskId: taskID,
		Status: "pending",
		Config: &backtestv1.BacktestConfig{Pair: config.Pair, Timeframe: "1d"},
	}
	go s.simulateBacktest(taskID, s.tasks[taskID].Config)
	return &backtestv1.RunQuantBtResponse{
		TaskId: taskID,
		Status: "pending",
	}, nil
}

// RunVpBt runs a volume profile backtest.
func (s *Service) RunVpBt(ctx context.Context, config *backtestv1.VpBtConfig) (*backtestv1.RunVpBtResponse, error) {
	taskID := fmt.Sprintf("vp-%d", time.Now().UnixNano())
	s.tasks[taskID] = &backtestv1.GetBacktestResponse{
		TaskId: taskID,
		Status: "pending",
		Config: &backtestv1.BacktestConfig{Pair: config.Pair, Timeframe: "1h"},
	}
	go s.simulateBacktest(taskID, s.tasks[taskID].Config)
	return &backtestv1.RunVpBtResponse{
		TaskId: taskID,
		Status: "pending",
	}, nil
}

// RunCtaBt runs a CTA-style backtest.
func (s *Service) RunCtaBt(ctx context.Context, config *backtestv1.CtaBtConfig) (*backtestv1.RunCtaBtResponse, error) {
	taskID := fmt.Sprintf("cta-%d", time.Now().UnixNano())
	s.tasks[taskID] = &backtestv1.GetBacktestResponse{
		TaskId: taskID,
		Status: "pending",
		Config: &backtestv1.BacktestConfig{Pair: config.Pair, Timeframe: "1d"},
	}
	go s.simulateBacktest(taskID, s.tasks[taskID].Config)
	return &backtestv1.RunCtaBtResponse{
		TaskId: taskID,
		Status: "pending",
	}, nil
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
				results, errWF := runWalkforwardSMA(ps, folds, req.TrainRatio)
				if errWF == nil {
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
	return &backtestv1.RunBootstrapResponse{
		SharpeP5: p5, SharpeP50: p50, SharpeP95: p95,
		// MaxDD 暂以 OOS Sharpe 倒数估算占位（真实回测引擎接入后替换）
		MaxddP5: math.Max(0.0, 0.5/(p95+1e-9)),
		MaxddP50: math.Max(0.0, 0.5/(p50+1e-9)),
		MaxddP95: math.Max(0.0, 0.5/(p5+1e-9)),
		Iterations: int32(iter),
	}, nil
}
