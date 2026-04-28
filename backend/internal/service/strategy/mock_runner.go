// Package strategy provides a mock backtest runner for development.
package strategy

import (
	"context"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

// MockRunner simulates backtest execution with realistic latency.
type MockRunner struct {
	minLatency time.Duration
	maxLatency time.Duration
}

// NewMockRunner creates a new mock runner with 1-2s latency.
func NewMockRunner() *MockRunner {
	return &MockRunner{
		minLatency: 1 * time.Second,
		maxLatency: 2 * time.Second,
	}
}

// Run simulates a backtest run.
func (r *MockRunner) Run(ctx context.Context, s Strategy) (RunResult, error) {
	started := time.Now().UTC()

	// Simulate work duration
	duration := r.minLatency + time.Duration(rand.Int63n(int64(r.maxLatency-r.minLatency)))
	select {
	case <-time.After(duration):
	case <-ctx.Done():
		return RunResult{}, ctx.Err()
	}

	finished := time.Now().UTC()

	// Generate realistic mock metrics
	metrics := map[string]any{
		"total_return":  round(rand.Float64()*0.4-0.1, 4),   // -10% to +30%
		"sharpe":        round(rand.Float64()*2.5+0.5, 2),   // 0.5 to 3.0
		"max_drawdown":  round(-rand.Float64()*0.25, 4),     // 0 to -25%
		"win_rate":      round(0.4+rand.Float64()*0.4, 2),   // 40% to 80%
		"trades":        rand.Intn(100) + 10,                // 10 to 110 trades
		"avg_trade_pnl": round(rand.Float64()*200-50, 2),
		"profit_factor": round(rand.Float64()*3+1, 2),
	}

	return RunResult{
		RunID:       uuid.New(),
		StrategyID:  s.ID,
		StartedAt:   started,
		FinishedAt:  finished,
		Status:      "success",
		Metrics:     metrics,
		Mock:        true,
	}, nil
}

func round(f float64, prec int) float64 {
	shift := 1.0
	for i := 0; i < prec; i++ {
		shift *= 10
	}
	return float64(int64(f*shift+0.5)) / shift
}
