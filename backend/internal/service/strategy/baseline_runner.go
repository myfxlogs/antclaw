// 真数据 Baseline Runner：从 quant_strategy_perf 读取最近一条历史绩效作为基线指标。
// 没有历史记录时返回 status=skipped 与解释字段，绝不编造随机数据。
package strategy

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BaselineRunner 基于 PostgreSQL `quant_strategy_perf` 表的真实历史绩效。
type BaselineRunner struct {
	pool *pgxpool.Pool
}

// NewBaselineRunner 构造基线 Runner；pool 为空时所有请求返回 skipped。
func NewBaselineRunner(pool *pgxpool.Pool) *BaselineRunner {
	return &BaselineRunner{pool: pool}
}

// Run 查询数据库里最近一条 strategy/symbol 绩效，返回真实指标。
func (r *BaselineRunner) Run(ctx context.Context, s Strategy) (RunResult, error) {
	started := time.Now().UTC()
	if r.pool == nil {
		return RunResult{
			RunID: uuid.New(), StrategyID: s.ID,
			StartedAt: started, FinishedAt: time.Now().UTC(),
			Status:       "skipped",
			ErrorMessage: "no postgres pool configured",
		}, nil
	}

	var (
		asof    time.Time
		sharpe  sql.NullFloat64
		sortino sql.NullFloat64
		dd      sql.NullFloat64
		win     sql.NullFloat64
		trades  sql.NullInt32
	)
	row := r.pool.QueryRow(ctx, `
		SELECT asof_date, sharpe, sortino, drawdown, win_rate, sample_trades
		  FROM quant_strategy_perf
		 WHERE symbol = $1 AND strategy = $2
		 ORDER BY asof_date DESC
		 LIMIT 1`, s.Symbol, s.Kind)
	err := row.Scan(&asof, &sharpe, &sortino, &dd, &win, &trades)
	finished := time.Now().UTC()
	if errors.Is(err, pgx.ErrNoRows) {
		return RunResult{
			RunID: uuid.New(), StrategyID: s.ID,
			StartedAt: started, FinishedAt: finished,
			Status:       "skipped",
			ErrorMessage: "no historical perf in quant_strategy_perf for this symbol/strategy",
		}, nil
	}
	if err != nil {
		return RunResult{
			RunID: uuid.New(), StrategyID: s.ID,
			StartedAt: started, FinishedAt: finished,
			Status:       "error",
			ErrorMessage: err.Error(),
		}, err
	}

	metrics := map[string]any{
		"asof_date":     asof.Format("2006-01-02"),
		"sharpe":        nullFloat(sharpe),
		"sortino":       nullFloat(sortino),
		"max_drawdown":  nullFloat(dd),
		"win_rate":      nullFloat(win),
		"sample_trades": nullInt(trades),
		"source":        "quant_strategy_perf",
	}
	return RunResult{
		RunID: uuid.New(), StrategyID: s.ID,
		StartedAt: started, FinishedAt: finished,
		Status:  "success",
		Metrics: metrics,
	}, nil
}

func nullFloat(v sql.NullFloat64) any {
	if v.Valid {
		return v.Float64
	}
	return nil
}

func nullInt(v sql.NullInt32) any {
	if v.Valid {
		return v.Int32
	}
	return nil
}
