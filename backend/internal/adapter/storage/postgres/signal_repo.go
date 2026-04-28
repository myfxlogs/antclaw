package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/antclaw/antclaw/internal/service/signals"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SignalRepoPg struct{ pool *pgxpool.Pool }

func NewSignalRepoPg(pool *pgxpool.Pool) *SignalRepoPg { return &SignalRepoPg{pool: pool} }

func (r *SignalRepoPg) SaveUnified(ctx context.Context, sig signals.UnifiedSignalRecord) (int64, error) {
	components, _ := json.Marshal(sig.Components)
	weights, _ := json.Marshal(sig.WeightsUsed)
	var id int64
	err := r.pool.QueryRow(ctx, `
INSERT INTO unified_signals (symbol, issued_at, recommendation, unified_score, confidence, components, missing_subsys, weights_used)
VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8::jsonb)
RETURNING id`, sig.Symbol, sig.IssuedAt, sig.Recommendation, sig.UnifiedScore, sig.Confidence, components, sig.MissingSubsys, weights).Scan(&id)
	return id, err
}

func (r *SignalRepoPg) GetByID(ctx context.Context, id int64) (*signals.UnifiedSignalRecord, error) {
	rows, err := r.querySignals(ctx, `SELECT id, symbol, issued_at, recommendation, unified_score, confidence, components, missing_subsys, weights_used FROM unified_signals WHERE id=$1 LIMIT 1`, id)
	if err != nil || len(rows) == 0 {
		return nil, err
	}
	return &rows[0], nil
}

func (r *SignalRepoPg) GetRecentBySymbol(ctx context.Context, symbol string, limit int) ([]signals.UnifiedSignalRecord, error) {
	return r.querySignals(ctx, `SELECT id, symbol, issued_at, recommendation, unified_score, confidence, components, missing_subsys, weights_used FROM unified_signals WHERE symbol=$1 ORDER BY issued_at DESC LIMIT $2`, symbol, limit)
}

func (r *SignalRepoPg) querySignals(ctx context.Context, sql string, args ...any) ([]signals.UnifiedSignalRecord, error) {
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (signals.UnifiedSignalRecord, error) {
		var rec signals.UnifiedSignalRecord
		var componentsRaw []byte
		var weightsRaw []byte
		err := row.Scan(&rec.ID, &rec.Symbol, &rec.IssuedAt, &rec.Recommendation, &rec.UnifiedScore, &rec.Confidence, &componentsRaw, &rec.MissingSubsys, &weightsRaw)
		if err != nil {
			return rec, err
		}
		_ = json.Unmarshal(componentsRaw, &rec.Components)
		_ = json.Unmarshal(weightsRaw, &rec.WeightsUsed)
		return rec, nil
	})
}

func (r *SignalRepoPg) SaveOutcome(ctx context.Context, outcome signals.SignalOutcome) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO signal_outcomes (signal_id, horizon, return_pct, direction_match, evaluated_at) VALUES ($1,$2,$3,$4,$5) ON CONFLICT (signal_id, horizon) DO UPDATE SET return_pct=EXCLUDED.return_pct,direction_match=EXCLUDED.direction_match,evaluated_at=EXCLUDED.evaluated_at`, outcome.SignalID, outcome.Horizon, outcome.ReturnPct, outcome.DirectionMatch, outcome.EvaluatedAt)
	return err
}

func (r *SignalRepoPg) GetOutcomeStats(ctx context.Context, signalType, symbol, horizon string, since time.Time) (*signals.OutcomeStats, error) {
	var out signals.OutcomeStats
	err := r.pool.QueryRow(ctx, `
SELECT COUNT(*)::int,
       COALESCE(AVG(CASE WHEN so.direction_match THEN 1.0 ELSE 0 END),0),
       COALESCE(AVG(so.return_pct),0),
       COALESCE(AVG(CASE WHEN so.return_pct > 0 THEN 1.0 ELSE 0 END),0),
       COALESCE(AVG(so.return_pct)/NULLIF(STDDEV_POP(so.return_pct),0),0),
       COALESCE(STDDEV_POP(so.return_pct),0)
FROM signal_outcomes so
JOIN unified_signals us ON us.id = so.signal_id
WHERE us.symbol=$1 AND so.horizon=$2 AND so.evaluated_at >= $3`, symbol, horizon, since).
		Scan(&out.SampleSize, &out.DirectionalAccuracy, &out.AvgReturn, &out.HitRate, &out.Sharpe, &out.StdDev)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *SignalRepoPg) GetActiveWeights(ctx context.Context) (map[string]float64, error) {
	var raw []byte
	err := r.pool.QueryRow(ctx, `SELECT weights FROM signal_weight_config WHERE is_active = TRUE ORDER BY updated_at DESC LIMIT 1`).Scan(&raw)
	if err != nil {
		if err == pgx.ErrNoRows {
			return map[string]float64{
				"cot": 1, "macro": 1, "factor": 1, "flow": 1, "vol": 1, "season": 1,
				"momentum": 1, "lowvol": 1, "trend": 1, "carry": 1, "crowding": 1, "residual": 1,
			}, nil
		}
		return nil, err
	}
	out := map[string]float64{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}
