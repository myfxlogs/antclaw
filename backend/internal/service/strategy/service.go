// Package strategy provides CRUD operations for trading strategies.
package strategy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Runner executes backtest runs.
type Runner interface {
	Run(ctx context.Context, s Strategy) (RunResult, error)
}

// Service provides strategy management.
type Service struct {
	pool   *pgxpool.Pool
	runner Runner
}

// NewService creates a new strategy service.
func NewService(pool *pgxpool.Pool, runner Runner) *Service {
	return &Service{pool: pool, runner: runner}
}

// List returns paginated strategies (excluding soft-deleted).
func (s *Service) List(ctx context.Context, offset, limit int) ([]Strategy, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	var total int
	err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM strategies WHERE deleted_at IS NULL`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id, name, kind, symbol, timeframe, params, schedule_cron, enabled, status,
		       COALESCE(description, ''), created_at, updated_at, COALESCE(created_by, ''), COALESCE(updated_by, ''),
		       last_run_at, COALESCE(last_run_status, '')
		FROM strategies WHERE deleted_at IS NULL
		ORDER BY updated_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	return scanStrategies(rows), total, nil
}

// Get returns a single strategy by ID.
func (s *Service) Get(ctx context.Context, id uuid.UUID) (*Strategy, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, name, kind, symbol, timeframe, params, schedule_cron, enabled, status,
		       COALESCE(description, ''), created_at, updated_at, COALESCE(created_by, ''), COALESCE(updated_by, ''),
		       last_run_at, COALESCE(last_run_status, '')
		FROM strategies WHERE id = $1 AND deleted_at IS NULL`, id)
	return scanStrategyRow(row)
}

// Create adds a new strategy.
func (s *Service) Create(ctx context.Context, st *Strategy, by string) error {
	if err := validateStrategy(st); err != nil {
		return err
	}
	st.ID = uuid.New()
	st.CreatedAt = time.Now()
	st.UpdatedAt = st.CreatedAt
	st.CreatedBy = by
	st.UpdatedBy = by

	_, err := s.pool.Exec(ctx, `
		INSERT INTO strategies (id, name, kind, symbol, timeframe, params, schedule_cron, enabled, status,
		                      description, created_at, updated_at, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
		st.ID, st.Name, st.Kind, st.Symbol, st.Timeframe, st.Params, st.ScheduleCron,
		st.Enabled, st.Status, st.Description, st.CreatedAt, st.UpdatedAt, st.CreatedBy, st.UpdatedBy)
	return err
}

// Update modifies an existing strategy.
func (s *Service) Update(ctx context.Context, id uuid.UUID, st *Strategy, by string) error {
	if err := validateStrategy(st); err != nil {
		return err
	}
	st.UpdatedAt = time.Now()
	st.UpdatedBy = by

	result, err := s.pool.Exec(ctx, `
		UPDATE strategies SET name = $1, kind = $2, symbol = $3, timeframe = $4, params = $5,
		                      schedule_cron = $6, enabled = $7, status = $8, description = $9,
		                      updated_at = $10, updated_by = $11
		WHERE id = $12 AND deleted_at IS NULL`,
		st.Name, st.Kind, st.Symbol, st.Timeframe, st.Params, st.ScheduleCron,
		st.Enabled, st.Status, st.Description, st.UpdatedAt, st.UpdatedBy, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors.New("strategy not found")
	}
	return nil
}

// SoftDelete marks a strategy as deleted.
func (s *Service) SoftDelete(ctx context.Context, id uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE strategies SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	return err
}

// Enable sets enabled = true.
func (s *Service) Enable(ctx context.Context, id uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE strategies SET enabled = TRUE, status = 'active', updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	return err
}

// Disable sets enabled = false.
func (s *Service) Disable(ctx context.Context, id uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE strategies SET enabled = FALSE, status = 'draft', updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	return err
}

// Run executes a backtest run immediately.
func (s *Service) Run(ctx context.Context, id uuid.UUID) (*RunResult, error) {
	st, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	result, err := s.runner.Run(ctx, *st)
	if err != nil {
		return nil, err
	}

	// Update last run info
	_, _ = s.pool.Exec(ctx, `
		UPDATE strategies SET last_run_at = $1, last_run_status = $2 WHERE id = $3`,
		result.FinishedAt, result.Status, id)

	// Store run result
	_, _ = s.pool.Exec(ctx, `
		INSERT INTO backtest_results (strategy_id, params, started_at, finished_at, status, metrics, mock)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		result.StrategyID, st.Params, result.StartedAt, result.FinishedAt,
		result.Status, result.Metrics, result.Mock)

	return &result, nil
}

// ListRuns returns historical runs for a strategy.
func (s *Service) ListRuns(ctx context.Context, strategyID uuid.UUID, limit int) ([]RunResult, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, strategy_id, started_at, finished_at, status, metrics, mock
		FROM backtest_results WHERE strategy_id = $1
		ORDER BY finished_at DESC LIMIT $2`, strategyID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []RunResult
	for rows.Next() {
		var r RunResult
		err := rows.Scan(&r.RunID, &r.StrategyID, &r.StartedAt, &r.FinishedAt, &r.Status, &r.Metrics, &r.Mock)
		if err != nil {
			continue
		}
		results = append(results, r)
	}
	return results, nil
}

// validateStrategy checks required fields.
func validateStrategy(s *Strategy) error {
	if s.Name == "" {
		return errors.New("name is required")
	}
	if !IsValidKind(s.Kind) {
		return fmt.Errorf("invalid kind: %s", s.Kind)
	}
	if s.Symbol == "" {
		return errors.New("symbol is required")
	}
	if s.Timeframe == "" {
		return errors.New("timeframe is required")
	}
	if !IsValidCron(s.ScheduleCron) {
		return errors.New("invalid schedule_cron")
	}
	if s.Status != "" && !IsValidStatus(s.Status) {
		return fmt.Errorf("invalid status: %s", s.Status)
	}
	return nil
}

func scanStrategies(rows pgx.Rows) []Strategy {
	var strategies []Strategy
	for rows.Next() {
		var s Strategy
		var rawParams []byte
		err := rows.Scan(&s.ID, &s.Name, &s.Kind, &s.Symbol, &s.Timeframe, &rawParams, &s.ScheduleCron,
			&s.Enabled, &s.Status, &s.Description, &s.CreatedAt, &s.UpdatedAt, &s.CreatedBy, &s.UpdatedBy,
			&s.LastRunAt, &s.LastRunStatus)
		if err != nil {
			continue
		}
		if len(rawParams) > 0 {
			if err := json.Unmarshal(rawParams, &s.Params); err != nil {
				s.Params = map[string]any{}
			}
		} else {
			s.Params = map[string]any{}
		}
		strategies = append(strategies, s)
	}
	return strategies
}

func scanStrategyRow(row pgx.Row) (*Strategy, error) {
	var s Strategy
	var rawParams []byte
	err := row.Scan(&s.ID, &s.Name, &s.Kind, &s.Symbol, &s.Timeframe, &rawParams, &s.ScheduleCron,
		&s.Enabled, &s.Status, &s.Description, &s.CreatedAt, &s.UpdatedAt, &s.CreatedBy, &s.UpdatedBy,
		&s.LastRunAt, &s.LastRunStatus)
	if err != nil {
		return nil, err
	}
	if len(rawParams) > 0 {
		if err := json.Unmarshal(rawParams, &s.Params); err != nil {
			s.Params = map[string]any{}
		}
	} else {
		s.Params = map[string]any{}
	}
	return &s, nil
}
