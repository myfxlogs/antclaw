package postgres

import (
	"context"
	"time"

	"github.com/antclaw/antclaw/internal/service/signals"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PricePgProvider struct{ pool *pgxpool.Pool }

func NewPricePgProvider(pool *pgxpool.Pool) *PricePgProvider { return &PricePgProvider{pool: pool} }

func (p *PricePgProvider) GetDailyBars(ctx context.Context, symbol string, from, to time.Time) ([]signals.Bar, error) {
	rows, err := p.pool.Query(ctx, `SELECT time, symbol, open, high, low, close, volume, source FROM price_daily WHERE symbol=$1 AND time >= $2 AND time <= $3 ORDER BY time ASC`, symbol, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (signals.Bar, error) {
		var b signals.Bar
		err := row.Scan(&b.Time, &b.Symbol, &b.Open, &b.High, &b.Low, &b.Close, &b.Volume, &b.Source)
		return b, err
	})
}

func (p *PricePgProvider) GetIntradayBars(ctx context.Context, symbol, interval string, from, to time.Time) ([]signals.Bar, error) {
	rows, err := p.pool.Query(ctx, `SELECT time, symbol, open, high, low, close, volume, source FROM price_intraday WHERE symbol=$1 AND interval=$2 AND time >= $3 AND time <= $4 ORDER BY time ASC`, symbol, interval, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (signals.Bar, error) {
		var b signals.Bar
		err := row.Scan(&b.Time, &b.Symbol, &b.Open, &b.High, &b.Low, &b.Close, &b.Volume, &b.Source)
		return b, err
	})
}

func (p *PricePgProvider) GetLatestPrice(ctx context.Context, symbol string) (*signals.Bar, error) {
	var b signals.Bar
	err := p.pool.QueryRow(ctx, `SELECT time, symbol, open, high, low, close, volume, source FROM price_daily WHERE symbol=$1 ORDER BY time DESC LIMIT 1`, symbol).
		Scan(&b.Time, &b.Symbol, &b.Open, &b.High, &b.Low, &b.Close, &b.Volume, &b.Source)
	if err != nil {
		return nil, err
	}
	return &b, nil
}
