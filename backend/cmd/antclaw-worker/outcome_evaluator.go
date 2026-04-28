package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// evaluateSignalOutcomes backfills signal_outcomes for horizons.
func evaluateSignalOutcomes(ctx context.Context, pool *pgxpool.Pool, logger *slog.Logger) error {
	horizons := []struct {
		label string
		dur   time.Duration
	}{
		{"1D", 24 * time.Hour},
		{"1W", 7 * 24 * time.Hour},
		{"2W", 14 * 24 * time.Hour},
		{"1M", 30 * 24 * time.Hour},
	}
	for _, h := range horizons {
		_, err := pool.Exec(ctx, `
INSERT INTO signal_outcomes(signal_id, horizon, return_pct, direction_match, evaluated_at)
SELECT s.id, $1,
       COALESCE((p2.close - p1.close) / NULLIF(p1.close,0),0),
       CASE
           WHEN s.recommendation IN ('LONG','STRONG_LONG') THEN p2.close > p1.close
           WHEN s.recommendation IN ('SHORT','STRONG_SHORT') THEN p2.close < p1.close
           ELSE ABS(COALESCE((p2.close - p1.close) / NULLIF(p1.close,0),0)) < 0.002
       END,
       NOW()
FROM unified_signals s
JOIN LATERAL (
  SELECT close FROM price_daily WHERE symbol = s.symbol AND time >= s.issued_at ORDER BY time ASC LIMIT 1
) p1 ON TRUE
JOIN LATERAL (
  SELECT close FROM price_daily WHERE symbol = s.symbol AND time >= s.issued_at + $2::interval ORDER BY time ASC LIMIT 1
) p2 ON TRUE
LEFT JOIN signal_outcomes o ON o.signal_id=s.id AND o.horizon=$1
WHERE o.signal_id IS NULL AND s.issued_at <= NOW() - $2::interval
ON CONFLICT (signal_id, horizon) DO NOTHING`, h.label, h.dur.String())
		if err != nil {
			logger.Warn("outcome evaluator failed", "horizon", h.label, "error", err)
			return fmt.Errorf("outcome evaluator %s: %w", h.label, err)
		}
	}
	return nil
}
