package postgres

import (
	"context"
	"time"

	"github.com/antclaw/antclaw/internal/service/signals"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RegimePgProvider struct{ pool *pgxpool.Pool }

func NewRegimePgProvider(pool *pgxpool.Pool) *RegimePgProvider { return &RegimePgProvider{pool: pool} }

func (p *RegimePgProvider) GetCurrent(ctx context.Context, symbol, timeframe string) (*signals.RegimeSnapshot, error) {
	var r signals.RegimeSnapshot
	err := p.pool.QueryRow(ctx, `SELECT time, symbol, timeframe, unified_score, unified_label, hmm_state, hmm_confidence, garch_regime, vol_ratio, adx_strength, adx_value FROM regime_overlay_history WHERE symbol=$1 AND timeframe=$2 ORDER BY time DESC LIMIT 1`, symbol, timeframe).
		Scan(&r.Time, &r.Symbol, &r.Timeframe, &r.UnifiedScore, &r.UnifiedLabel, &r.HMMState, &r.HMMConfidence, &r.GARCHRegime, &r.VolRatio, &r.ADXStrength, &r.ADXValue)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (p *RegimePgProvider) GetHistory(ctx context.Context, symbol, timeframe string, lookback int) ([]signals.RegimeSnapshot, error) {
	rows, err := p.pool.Query(ctx, `SELECT time, symbol, timeframe, unified_score, unified_label, hmm_state, hmm_confidence, garch_regime, vol_ratio, adx_strength, adx_value FROM regime_overlay_history WHERE symbol=$1 AND timeframe=$2 ORDER BY time DESC LIMIT $3`, symbol, timeframe, lookback)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (signals.RegimeSnapshot, error) {
		var r signals.RegimeSnapshot
		err := row.Scan(&r.Time, &r.Symbol, &r.Timeframe, &r.UnifiedScore, &r.UnifiedLabel, &r.HMMState, &r.HMMConfidence, &r.GARCHRegime, &r.VolRatio, &r.ADXStrength, &r.ADXValue)
		return r, err
	})
}

func (p *RegimePgProvider) GetTransitions(ctx context.Context, symbol string, since time.Time) ([]signals.RegimeTransition, error) {
	rows, err := p.pool.Query(ctx, `SELECT time, symbol, timeframe, from_label, to_label, from_score, to_score, severity FROM regime_transitions WHERE symbol=$1 AND time >= $2 ORDER BY time DESC`, symbol, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (signals.RegimeTransition, error) {
		var r signals.RegimeTransition
		err := row.Scan(&r.Time, &r.Symbol, &r.Timeframe, &r.FromLabel, &r.ToLabel, &r.FromScore, &r.ToScore, &r.Severity)
		return r, err
	})
}
