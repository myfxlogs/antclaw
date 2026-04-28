package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DailyBar represents a daily price bar
type DailyBar struct {
	Time     time.Time `db:"time"`
	Symbol   string    `db:"symbol"`
	Open     float64   `db:"open"`
	High     float64   `db:"high"`
	Low      float64   `db:"low"`
	Close    float64   `db:"close"`
	Volume   float64   `db:"volume"`
	Source   string    `db:"source"`
}

// IntradayBar represents an intraday price bar
type IntradayBar struct {
	Time     time.Time `db:"time"`
	Symbol   string    `db:"symbol"`
	Interval string    `db:"interval"`
	Open     float64   `db:"open"`
	High     float64   `db:"high"`
	Low      float64   `db:"low"`
	Close    float64   `db:"close"`
	Volume   float64   `db:"volume"`
	Source   string    `db:"source"`
}

// PriceRepository defines price data operations
type PriceRepository interface {
	UpsertDailyBars(ctx context.Context, bars []DailyBar) error
	UpsertIntradayBars(ctx context.Context, bars []IntradayBar) error
	GetDailyBars(ctx context.Context, symbol string, from, to time.Time) ([]DailyBar, error)
	GetIntradayBars(ctx context.Context, symbol, interval string, from, to time.Time) ([]IntradayBar, error)
	GetLatest(ctx context.Context, symbol string) (*DailyBar, error)
	GetLatestIntraday(ctx context.Context, symbol, interval string) (*IntradayBar, error)
}

type priceRepo struct {
	pool *pgxpool.Pool
}

// NewPriceRepository creates a new price repository
func NewPriceRepository(pool *pgxpool.Pool) PriceRepository {
	return &priceRepo{pool: pool}
}

// UpsertDailyBars inserts or updates daily bars
func (r *priceRepo) UpsertDailyBars(ctx context.Context, bars []DailyBar) error {
	if len(bars) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, b := range bars {
		batch.Queue(`
			INSERT INTO price_daily (time, symbol, open, high, low, close, volume, source)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (time, symbol) DO UPDATE SET
				open = EXCLUDED.open,
				high = EXCLUDED.high,
				low = EXCLUDED.low,
				close = EXCLUDED.close,
				volume = EXCLUDED.volume,
				source = EXCLUDED.source
		`, b.Time, b.Symbol, b.Open, b.High, b.Low, b.Close, b.Volume, b.Source)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < len(bars); i++ {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("upsert daily bar %d failed: %w", i, err)
		}
	}
	return br.Close()
}

// UpsertIntradayBars inserts or updates intraday bars
func (r *priceRepo) UpsertIntradayBars(ctx context.Context, bars []IntradayBar) error {
	if len(bars) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, b := range bars {
		batch.Queue(`
			INSERT INTO price_intraday (time, symbol, interval, open, high, low, close, volume, source)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (time, symbol, interval) DO UPDATE SET
				open = EXCLUDED.open,
				high = EXCLUDED.high,
				low = EXCLUDED.low,
				close = EXCLUDED.close,
				volume = EXCLUDED.volume,
				source = EXCLUDED.source
		`, b.Time, b.Symbol, b.Interval, b.Open, b.High, b.Low, b.Close, b.Volume, b.Source)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < len(bars); i++ {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("upsert intraday bar %d failed: %w", i, err)
		}
	}
	return br.Close()
}

// GetDailyBars retrieves daily bars for a symbol in a date range
func (r *priceRepo) GetDailyBars(ctx context.Context, symbol string, from, to time.Time) ([]DailyBar, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT time, symbol, open, high, low, close, volume, source
		FROM price_daily
		WHERE symbol = $1 AND time >= $2 AND time <= $3
		ORDER BY time ASC
	`, symbol, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[DailyBar])
}

// GetIntradayBars retrieves intraday bars for a symbol in a date range
func (r *priceRepo) GetIntradayBars(ctx context.Context, symbol, interval string, from, to time.Time) ([]IntradayBar, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT time, symbol, interval, open, high, low, close, volume, source
		FROM price_intraday
		WHERE symbol = $1 AND interval = $2 AND time >= $3 AND time <= $4
		ORDER BY time ASC
	`, symbol, interval, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[IntradayBar])
}

// GetLatest retrieves the latest daily bar for a symbol
func (r *priceRepo) GetLatest(ctx context.Context, symbol string) (*DailyBar, error) {
	row, err := r.pool.Query(ctx, `
		SELECT time, symbol, open, high, low, close, volume, source
		FROM price_daily
		WHERE symbol = $1
		ORDER BY time DESC
		LIMIT 1
	`, symbol)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	results, err := pgx.CollectRows(row, pgx.RowToStructByName[DailyBar])
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}
	return &results[0], nil
}

// GetLatestIntraday retrieves the latest intraday bar for a symbol
func (r *priceRepo) GetLatestIntraday(ctx context.Context, symbol, interval string) (*IntradayBar, error) {
	row, err := r.pool.Query(ctx, `
		SELECT time, symbol, interval, open, high, low, close, volume, source
		FROM price_intraday
		WHERE symbol = $1 AND interval = $2
		ORDER BY time DESC
		LIMIT 1
	`, symbol, interval)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	results, err := pgx.CollectRows(row, pgx.RowToStructByName[IntradayBar])
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}
	return &results[0], nil
}
