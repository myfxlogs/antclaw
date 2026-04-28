package postgres

import (
	"context"

	"github.com/antclaw/antclaw/internal/service/signals"
	"github.com/jackc/pgx/v5/pgxpool"
)

type VolPgProvider struct{ pool *pgxpool.Pool }

func NewVolPgProvider(pool *pgxpool.Pool) *VolPgProvider { return &VolPgProvider{pool: pool} }

func (p *VolPgProvider) GetVIX(ctx context.Context) (float64, error) {
	var close float64
	err := p.pool.QueryRow(ctx, `SELECT close FROM price_daily WHERE symbol='VIX' ORDER BY time DESC LIMIT 1`).Scan(&close)
	return close, err
}

func (p *VolPgProvider) GetGARCH(ctx context.Context, symbol string) (*signals.VolRegimeData, error) {
	var out signals.VolRegimeData
	out.Symbol = symbol
	out.Regime = "MEDIUM"
	out.Annualized = 0.2
	out.Percentile = 50
	return &out, nil
}

func (p *VolPgProvider) GetIVSurface(ctx context.Context, underlying string) (*signals.IVSurface, error) {
	return &signals.IVSurface{
		Underlying:  underlying,
		AtTheMoney:  0.2,
		Skew25Delta: 0,
		TermStructure: []signals.TermPoint{
			{TenorDays: 7, IV: 0.18},
			{TenorDays: 30, IV: 0.2},
			{TenorDays: 90, IV: 0.22},
		},
	}, nil
}
