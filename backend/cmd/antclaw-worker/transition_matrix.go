package main

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func buildTransitionMatrix(ctx context.Context, pool *pgxpool.Pool, logger *slog.Logger) error {
	symbols := []string{"EURUSD", "GBPUSD", "USDJPY", "USDCHF", "AUDUSD", "USDCAD", "NZDUSD", "BTCUSDT", "ETHUSDT"}
	for _, symbol := range symbols {
		rows, err := pool.Query(ctx, `SELECT time, unified_label FROM regime_overlay_history WHERE symbol=$1 AND timeframe='1d' ORDER BY time ASC LIMIT 730`, symbol)
		if err != nil {
			logger.Warn("transition matrix query failed", "symbol", symbol, "error", err)
			continue
		}
		var labels []string
		for rows.Next() {
			var t time.Time
			var l string
			if err := rows.Scan(&t, &l); err == nil && l != "" {
				labels = append(labels, strings.ToUpper(l))
			}
		}
		rows.Close()
		if len(labels) < 60 {
			continue
		}
		states := []string{"STRONG_BULL", "BULL", "NEUTRAL", "BEAR", "STRONG_BEAR"}
		counts := map[string]map[string]int{}
		for _, s := range states {
			counts[s] = map[string]int{}
		}
		for i := 0; i < len(labels)-1; i++ {
			counts[labels[i]][labels[i+1]]++
		}
		for _, from := range states {
			total := 0
			for _, to := range states {
				total += counts[from][to]
			}
			if total == 0 {
				total = len(states)
			}
			for _, to := range states {
				num := counts[from][to]
				if num == 0 && counts[from][to] == 0 {
					num = 1
				}
				prob := float64(num) / float64(total)
				_, _ = pool.Exec(ctx, `INSERT INTO regime_transition_matrix(asof_date,symbol,timeframe,from_label,to_label,probability,sample_size)
VALUES (CURRENT_DATE,$1,'1d',$2,$3,$4,$5)
ON CONFLICT (asof_date,symbol,timeframe,from_label,to_label) DO UPDATE SET probability=EXCLUDED.probability,sample_size=EXCLUDED.sample_size`,
					symbol, from, to, prob, total)
			}
		}
	}
	return nil
}
