package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func evaluateAlerts(ctx context.Context, pool *pgxpool.Pool, logger *slog.Logger) error {
	rows, err := pool.Query(ctx, `SELECT id,user_id,alert_type,symbol,params::text,COALESCE(last_fired_at,'1970-01-01'::timestamptz),cooldown_seconds
FROM user_signal_alerts WHERE enabled=true AND deleted_at IS NULL`)
	if err != nil {
		return fmt.Errorf("alert evaluator query: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var userID, alertType, symbol, params string
		var lastFired time.Time
		var cooldown int32
		if err := rows.Scan(&id, &userID, &alertType, &symbol, &params, &lastFired, &cooldown); err != nil {
			continue
		}
		if time.Since(lastFired) < time.Duration(cooldown)*time.Second {
			continue
		}
		if !alertTriggered(ctx, pool, alertType, symbol, params) {
			continue
		}
		payload, _ := json.Marshal(map[string]any{
			"type": "alert.fired", "alert_id": id, "alert_type": alertType, "symbol": symbol, "fired_at": time.Now().UTC().Format(time.RFC3339),
		})
		_, _ = pool.Exec(ctx, `INSERT INTO notifications(user_id,type,title,body,data,priority) VALUES ($1::uuid,'in_app',$2,$3,$4::jsonb,'high')`,
			userID, "策略告警触发", "您的 "+symbol+" 告警已触发", string(payload))
		_ = globalRedis.Raw().Publish(ctx, "user:"+userID+":alerts", string(payload)).Err()
		_, _ = pool.Exec(ctx, `UPDATE user_signal_alerts SET last_fired_at=NOW(),updated_at=NOW() WHERE id=$1`, id)
	}
	return nil
}

func alertTriggered(ctx context.Context, pool *pgxpool.Pool, alertType, symbol, paramsJSON string) bool {
	switch alertType {
	case "cot_extreme":
		var p struct {
			COTIndexMin float64 `json:"cot_index_min"`
			COTIndexMax float64 `json:"cot_index_max"`
		}
		_ = json.Unmarshal([]byte(paramsJSON), &p)
		var idx float64
		if err := pool.QueryRow(ctx, `SELECT cot_index FROM cot_analyses WHERE contract_code=(SELECT contract_code FROM cot_analyses ORDER BY report_date DESC LIMIT 1) ORDER BY report_date DESC LIMIT 1`).Scan(&idx); err != nil {
			return false
		}
		return idx >= p.COTIndexMin || idx <= p.COTIndexMax
	case "signal_flip":
		var prev, curr string
		if err := pool.QueryRow(ctx, `SELECT recommendation FROM unified_signals WHERE symbol=$1 ORDER BY issued_at DESC OFFSET 1 LIMIT 1`, symbol).Scan(&prev); err != nil {
			return false
		}
		if err := pool.QueryRow(ctx, `SELECT recommendation FROM unified_signals WHERE symbol=$1 ORDER BY issued_at DESC LIMIT 1`, symbol).Scan(&curr); err != nil {
			return false
		}
		return prev != curr
	case "regime_change":
		var c int
		_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM regime_transitions WHERE symbol=$1 AND time >= NOW()-INTERVAL '1 day'`, symbol).Scan(&c)
		return c > 0
	case "price_threshold":
		var p struct {
			Above *float64 `json:"above"`
			Below *float64 `json:"below"`
		}
		_ = json.Unmarshal([]byte(paramsJSON), &p)
		var close float64
		if err := pool.QueryRow(ctx, `SELECT close FROM price_daily WHERE symbol=$1 ORDER BY time DESC LIMIT 1`, symbol).Scan(&close); err != nil {
			return false
		}
		if p.Above != nil && close > *p.Above {
			return true
		}
		if p.Below != nil && close < *p.Below {
			return true
		}
		return false
	}
	return false
}
