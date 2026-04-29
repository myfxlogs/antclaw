package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/antclaw/antclaw/internal/adapter/storage/postgres/db"
	"github.com/antclaw/antclaw/internal/notify"
)

// evaluateAlerts 扫描用户告警，命中即通过 notify.Service 统一投递（持久化 + 实时 SSE）。
//
// 推送链路："逐用户匹配 → 发送"——传输层是 SSE，分发由 Redis Pub/Sub 完成。
func evaluateAlerts(ctx context.Context, pool *pgxpool.Pool, logger *slog.Logger) error {
	rows, err := pool.Query(ctx, `SELECT id,user_id,alert_type,symbol,params::text,COALESCE(last_fired_at,'1970-01-01'::timestamptz),cooldown_seconds
FROM user_signal_alerts WHERE enabled=true AND deleted_at IS NULL`)
	if err != nil {
		return fmt.Errorf("alert evaluator query: %w", err)
	}
	defer rows.Close()

	notifySvc := notify.NewService(db.New(pool), globalRedis.Raw())

	for rows.Next() {
		var id int64
		var userIDStr, alertType, symbol, params string
		var lastFired time.Time
		var cooldown int32
		if err := rows.Scan(&id, &userIDStr, &alertType, &symbol, &params, &lastFired, &cooldown); err != nil {
			continue
		}
		if time.Since(lastFired) < time.Duration(cooldown)*time.Second {
			continue
		}
		if !alertTriggered(ctx, pool, alertType, symbol, params) {
			continue
		}
		uid, err := uuid.Parse(userIDStr)
		if err != nil {
			logger.Warn("alert: invalid user_id", "id", id, "user_id", userIDStr)
			continue
		}
		// dedup_key：同一告警在 cooldown 周期内只发一次。
		dedup := fmt.Sprintf("alert:%d:%s", id, time.Now().UTC().Format("20060102T1504"))
		if err := notifySvc.Send(ctx, &notify.Notification{
			UserID:   uid,
			Category: "alert",
			Severity: "high",
			Title:    "策略告警触发",
			Body:     "您的 " + symbol + " 告警已触发",
			Data: map[string]string{
				"alert_id":   fmt.Sprintf("%d", id),
				"alert_type": alertType,
				"symbol":     symbol,
				"fired_at":   time.Now().UTC().Format(time.RFC3339),
			},
			DedupKey: dedup,
			DedupTTL: time.Duration(cooldown) * time.Second,
		}); err != nil {
			logger.Warn("alert: notify send failed", "id", id, "err", err)
		}
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
