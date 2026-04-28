package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BacktestJob represents a backtest job
type BacktestJob struct {
	JobID         int64     `db:"job_id"`
	StrategyName  string    `db:"strategy_name"`
	Status        string    `db:"status"`
	Parameters    []byte    `db:"parameters"`
	CreatedAt     time.Time `db:"created_at"`
	StartedAt     *time.Time `db:"started_at"`
	CompletedAt   *time.Time `db:"completed_at"`
}

// BacktestResult represents backtest results
type BacktestResult struct {
	ResultID     int64     `db:"result_id"`
	JobID        int64     `db:"job_id"`
	TotalReturn  float64   `db:"total_return"`
	SharpeRatio  float64   `db:"sharpe_ratio"`
	MaxDrawdown  float64   `db:"max_drawdown"`
	WinRate      float64   `db:"win_rate"`
	TotalTrades  int       `db:"total_trades"`
	ResultJSON   []byte    `db:"result_json"`
	CreatedAt    time.Time `db:"created_at"`
}

// BacktestRepository defines backtest data operations
type BacktestRepository interface {
	CreateJob(ctx context.Context, job *BacktestJob) (int64, error)
	UpdateJob(ctx context.Context, job *BacktestJob) error
	GetJob(ctx context.Context, jobID int64) (*BacktestJob, error)
	SaveResult(ctx context.Context, result *BacktestResult) error
	GetResult(ctx context.Context, jobID int64) (*BacktestResult, error)
}

type backtestRepo struct {
	pool *pgxpool.Pool
}

// NewBacktestRepository creates a new backtest repository
func NewBacktestRepository(pool *pgxpool.Pool) BacktestRepository {
	return &backtestRepo{pool: pool}
}

// CreateJob creates a new backtest job
func (r *backtestRepo) CreateJob(ctx context.Context, job *BacktestJob) (int64, error) {
	var jobID int64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO backtest_jobs (strategy_name, status, parameters, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING job_id
	`, job.StrategyName, job.Status, job.Parameters).Scan(&jobID)
	return jobID, err
}

// UpdateJob updates a backtest job
func (r *backtestRepo) UpdateJob(ctx context.Context, job *BacktestJob) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE backtest_jobs SET
			status = $1,
			started_at = $2,
			completed_at = $3
		WHERE job_id = $4
	`, job.Status, job.StartedAt, job.CompletedAt, job.JobID)
	return err
}

// GetJob retrieves a backtest job
func (r *backtestRepo) GetJob(ctx context.Context, jobID int64) (*BacktestJob, error) {
	row, err := r.pool.Query(ctx, `
		SELECT job_id, strategy_name, status, parameters, created_at, started_at, completed_at
		FROM backtest_jobs
		WHERE job_id = $1
	`, jobID)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	if !row.Next() {
		return nil, nil
	}

	var job BacktestJob
	err = row.Scan(&job.JobID, &job.StrategyName, &job.Status, &job.Parameters, &job.CreatedAt, &job.StartedAt, &job.CompletedAt)
	return &job, err
}

// SaveResult saves backtest results
func (r *backtestRepo) SaveResult(ctx context.Context, result *BacktestResult) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO backtest_results (job_id, total_return, sharpe_ratio, max_drawdown, win_rate, total_trades, result_json, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT (job_id) DO UPDATE SET
			total_return = EXCLUDED.total_return,
			sharpe_ratio = EXCLUDED.sharpe_ratio,
			max_drawdown = EXCLUDED.max_drawdown,
			win_rate = EXCLUDED.win_rate,
			total_trades = EXCLUDED.total_trades,
			result_json = EXCLUDED.result_json
	`, result.JobID, result.TotalReturn, result.SharpeRatio, result.MaxDrawdown, result.WinRate, result.TotalTrades, result.ResultJSON)
	return err
}

// GetResult retrieves backtest results
func (r *backtestRepo) GetResult(ctx context.Context, jobID int64) (*BacktestResult, error) {
	row, err := r.pool.Query(ctx, `
		SELECT result_id, job_id, total_return, sharpe_ratio, max_drawdown, win_rate, total_trades, result_json, created_at
		FROM backtest_results
		WHERE job_id = $1
	`, jobID)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	if !row.Next() {
		return nil, nil
	}

	var result BacktestResult
	err = row.Scan(&result.ResultID, &result.JobID, &result.TotalReturn, &result.SharpeRatio, &result.MaxDrawdown, &result.WinRate, &result.TotalTrades, &result.ResultJSON, &result.CreatedAt)
	return &result, err
}
