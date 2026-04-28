package postgres

import (
	"context"
	"time"

	"github.com/antclaw/antclaw/internal/service/signals"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FlowPgProvider struct{ pool *pgxpool.Pool }

func NewFlowPgProvider(pool *pgxpool.Pool) *FlowPgProvider { return &FlowPgProvider{pool: pool} }

func (p *FlowPgProvider) GetDivergence(ctx context.Context, pairA, pairB string, since time.Time) ([]signals.FlowDivergence, error) {
	rows, err := p.pool.Query(ctx, `SELECT time, pair_a, pair_b, corr, baseline_mean, baseline_std, z_score, lead_lag FROM flow_divergence_history WHERE pair_a=$1 AND pair_b=$2 AND time >= $3 ORDER BY time DESC`, pairA, pairB, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (signals.FlowDivergence, error) {
		var d signals.FlowDivergence
		err := row.Scan(&d.Time, &d.PairA, &d.PairB, &d.Corr, &d.BaselineMean, &d.BaselineStd, &d.ZScore, &d.LeadLag)
		return d, err
	})
}

func (p *FlowPgProvider) GetTopDivergent(ctx context.Context, since time.Time, limit int) ([]signals.FlowDivergence, error) {
	rows, err := p.pool.Query(ctx, `SELECT time, pair_a, pair_b, corr, baseline_mean, baseline_std, z_score, lead_lag FROM flow_divergence_history WHERE time >= $1 ORDER BY ABS(z_score) DESC LIMIT $2`, since, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (signals.FlowDivergence, error) {
		var d signals.FlowDivergence
		err := row.Scan(&d.Time, &d.PairA, &d.PairB, &d.Corr, &d.BaselineMean, &d.BaselineStd, &d.ZScore, &d.LeadLag)
		return d, err
	})
}
