package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MacroObservation represents a FRED observation
type MacroObservation struct {
	Time         time.Time `db:"time"`
	Source       string    `db:"source"`
	SeriesID     string    `db:"series_id"`
	ValueNumeric *float64  `db:"value_numeric"`
	ValueText    *string   `db:"value_text"`
	RawJSON      []byte    `db:"raw_json"`
	FetchedAt    time.Time `db:"fetched_at"`
}

// RegimeSnapshot represents a macro regime classification
type RegimeSnapshot struct {
	Time    time.Time `db:"time"`
	Regime  string    `db:"regime"`
	Score   float64   `db:"score"`
	Details []byte    `db:"details"`
}

// MacroRepository defines macro data operations
type MacroRepository interface {
	SaveObservations(ctx context.Context, obs []MacroObservation) (int, error)
	GetLatest(ctx context.Context, seriesID string) (*MacroObservation, error)
	GetHistory(ctx context.Context, seriesID string, from, to time.Time) ([]MacroObservation, error)
	GetRegimeHistory(ctx context.Context, from, to time.Time) ([]RegimeSnapshot, error)
	SaveRegime(ctx context.Context, snapshot RegimeSnapshot) error
}

type macroRepo struct {
	pool *pgxpool.Pool
}

// NewMacroRepository creates a new macro repository
func NewMacroRepository(pool *pgxpool.Pool) MacroRepository {
	return &macroRepo{pool: pool}
}

// SaveObservations inserts or updates macro observations
func (r *macroRepo) SaveObservations(ctx context.Context, obs []MacroObservation) (int, error) {
	if len(obs) == 0 {
		return 0, nil
	}

	batch := &pgx.Batch{}
	for _, o := range obs {
		batch.Queue(`
			INSERT INTO data_snapshots (time, source, series_id, value_numeric, value_text, raw_json, fetched_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (time, source, series_id) DO UPDATE SET
				value_numeric = EXCLUDED.value_numeric,
				value_text = EXCLUDED.value_text,
				raw_json = EXCLUDED.raw_json,
				fetched_at = EXCLUDED.fetched_at
			WHERE data_snapshots.value_numeric IS DISTINCT FROM EXCLUDED.value_numeric
		`, o.Time, o.Source, o.SeriesID, o.ValueNumeric, o.ValueText, o.RawJSON, o.FetchedAt)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	var count int
	for i := 0; i < len(obs); i++ {
		if _, err := br.Exec(); err != nil {
			return count, fmt.Errorf("save observation %d failed: %w", i, err)
		}
		count++
	}
	return count, br.Close()
}

// GetLatest retrieves the latest observation for a series
func (r *macroRepo) GetLatest(ctx context.Context, seriesID string) (*MacroObservation, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT time, source, series_id, value_numeric, value_text, raw_json, fetched_at
		FROM data_snapshots
		WHERE series_id = $1 AND source = 'fred'
		ORDER BY time DESC
		LIMIT 1
	`, seriesID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results, err := pgx.CollectRows(rows, pgx.RowToStructByName[MacroObservation])
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}
	return &results[0], nil
}

// GetHistory retrieves historical observations for a series
func (r *macroRepo) GetHistory(ctx context.Context, seriesID string, from, to time.Time) ([]MacroObservation, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT time, source, series_id, value_numeric, value_text, raw_json, fetched_at
		FROM data_snapshots
		WHERE series_id = $1 AND source = 'fred' AND time >= $2 AND time <= $3
		ORDER BY time DESC
	`, seriesID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[MacroObservation])
}

// GetRegimeHistory retrieves regime history
func (r *macroRepo) GetRegimeHistory(ctx context.Context, from, to time.Time) ([]RegimeSnapshot, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT time, regime, score, details
		FROM macro_regime_history
		WHERE time >= $1 AND time <= $2
		ORDER BY time DESC
	`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[RegimeSnapshot])
}

// SaveRegime saves a regime snapshot
func (r *macroRepo) SaveRegime(ctx context.Context, snapshot RegimeSnapshot) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO macro_regime_history (time, regime, score, details)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (time) DO UPDATE SET
			regime = EXCLUDED.regime,
			score = EXCLUDED.score,
			details = EXCLUDED.details
	`, snapshot.Time, snapshot.Regime, snapshot.Score, snapshot.Details)
	return err
}
