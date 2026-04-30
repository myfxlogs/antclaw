// Package main 中 calendar 衍生数据采集：
//   - calendar-titles      : 兜底以原 title 写入英文行；后续可挂 LLM 翻译。
//   - calendar-surprise    : 历史 (actual − forecast) 的 σ 标准化样本。
//   - event-impact         : 高影响事件前后窗口（15m/30m/1h/4h）价格变化。
package main

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// collectCalendarTitles 把 calendar_events 中尚未存在 lang='en' 行的事件，
// 以原始 title 兜底写入；同步生成 source='mql5' 标记，便于后续 LLM 翻译只覆盖 source='llm' 的行。
func collectCalendarTitles(ctx context.Context, db *pgxpool.Pool, logger *slog.Logger) error {
	tag, err := db.Exec(ctx, `
		INSERT INTO calendar_event_titles (event_id, lang, title, source, confidence, fetched_at)
		SELECT e.event_id, 'en', e.title, 'mql5', 1.0, NOW()
		  FROM calendar_events e
		  LEFT JOIN calendar_event_titles t
		    ON t.event_id = e.event_id AND t.lang = 'en'
		 WHERE t.event_id IS NULL`)
	if err != nil {
		return fmt.Errorf("calendar-titles upsert: %w", err)
	}
	logger.Info("calendar-titles inserted", "rows", tag.RowsAffected())
	return nil
}

// collectCalendarSurprise 解析 actual/forecast 数值并写入 history 表。
// σ 在表内单点计算：用同 event_name+currency 的历史 |diff| stddev 作为分母。
func collectCalendarSurprise(ctx context.Context, db *pgxpool.Pool, logger *slog.Logger) error {
	rows, err := db.Query(ctx, `
		SELECT event_id, title, currency,
		       COALESCE(NULLIF(scheduled_at, '0001-01-01 00:00:00+00'::timestamptz), fetched_at) AS released_ts,
		       actual_value, forecast_value
		  FROM calendar_events
		 WHERE actual_value IS NOT NULL AND actual_value <> ''
		   AND forecast_value IS NOT NULL AND forecast_value <> ''`)
	if err != nil {
		return fmt.Errorf("calendar-surprise query: %w", err)
	}
	defer rows.Close()

	type ev struct {
		eventID, name, ccy   string
		ts                   time.Time
		actual, forecast     float64
	}
	var batch []ev
	for rows.Next() {
		var e ev
		var actual, forecast string
		if err := rows.Scan(&e.eventID, &e.name, &e.ccy, &e.ts, &actual, &forecast); err != nil {
			continue
		}
		a, ok1 := parseNumeric(actual)
		f, ok2 := parseNumeric(forecast)
		if !ok1 || !ok2 {
			continue
		}
		e.actual, e.forecast = a, f
		batch = append(batch, e)
	}

	inserted := 0
	for _, e := range batch {
		// 上插：以 (event_name,currency,released_at) 唯一去重。
		diff := e.actual - e.forecast
		var sigma float64
		_ = db.QueryRow(ctx, `
			SELECT COALESCE(STDDEV_SAMP(diff),0)
			  FROM calendar_surprise_history
			 WHERE event_name=$1 AND currency=$2`, e.name, e.ccy).Scan(&sigma)
		var z float64
		if sigma > 1e-9 {
			z = diff / sigma
		}
		tag, err := db.Exec(ctx, `
			INSERT INTO calendar_surprise_history(event_name,currency,released_at,actual_val,forecast_val,diff,sigma)
			SELECT $1::varchar,$2::varchar,$3::timestamptz,$4::float8,$5::float8,$6::float8,$7::float8
			 WHERE NOT EXISTS (
			   SELECT 1 FROM calendar_surprise_history
			    WHERE event_name=$1::varchar AND currency=$2::varchar AND released_at=$3::timestamptz)`,
			e.name, e.ccy, e.ts, e.actual, e.forecast, diff, z)
		if err != nil {
			logger.Warn("calendar-surprise insert failed", "name", e.name, "ccy", e.ccy, "error", err)
			continue
		}
		if tag.RowsAffected() > 0 {
			inserted++
		}
	}
	logger.Info("calendar-surprise inserted", "candidates", len(batch), "rows", inserted)
	return nil
}

// collectEventImpact 对 24h 内已发生的高影响事件，按 (15m,30m,1h,4h) 取价格窗口，
// 用 EURUSD/GBPUSD/USDJPY 中匹配 currency 的代理标的（USD 事件 → 用 EURUSD 反向）。
func collectEventImpact(ctx context.Context, db *pgxpool.Pool, logger *slog.Logger) error {
	windows := []struct {
		name string
		dur  time.Duration
	}{{"15m", 15 * time.Minute}, {"30m", 30 * time.Minute}, {"1h", time.Hour}, {"4h", 4 * time.Hour}}

	rows, err := db.Query(ctx, `
		SELECT event_id, currency,
		       COALESCE(NULLIF(scheduled_at, '0001-01-01 00:00:00+00'::timestamptz), fetched_at) AS event_ts
		  FROM calendar_events
		 WHERE impact='high'
		   AND COALESCE(NULLIF(scheduled_at, '0001-01-01 00:00:00+00'::timestamptz), fetched_at) <= NOW() - INTERVAL '1 hour'
		   AND COALESCE(NULLIF(scheduled_at, '0001-01-01 00:00:00+00'::timestamptz), fetched_at) >= NOW() - INTERVAL '60 days'`)
	if err != nil {
		return fmt.Errorf("event-impact query: %w", err)
	}
	defer rows.Close()

	type evRow struct {
		id, ccy string
		ts      time.Time
	}
	var events []evRow
	for rows.Next() {
		var e evRow
		if err := rows.Scan(&e.id, &e.ccy, &e.ts); err == nil {
			events = append(events, e)
		}
	}

	inserted := 0
	for _, e := range events {
		symbol := pickProxySymbol(e.ccy)
		if symbol == "" {
			continue
		}
		for _, w := range windows {
			before, ok1 := nearestIntradayClose(ctx, db, symbol, e.ts.Add(-w.dur))
			after, ok2 := nearestIntradayClose(ctx, db, symbol, e.ts.Add(w.dur))
			if !ok1 || !ok2 || before == 0 {
				continue
			}
			pct := (after - before) / before * 100
			_, err := db.Exec(ctx, `
				INSERT INTO event_impact_records(event_id,"window",symbol,price_before,price_after,pct_change,recorded_at)
				VALUES ($1,$2,$3,$4,$5,$6,$7)
				ON CONFLICT DO NOTHING`,
				e.id, w.name, symbol, before, after, pct, e.ts)
			if err == nil {
				inserted++
			}
		}
	}
	logger.Info("event-impact inserted", "events", len(events), "rows", inserted)
	return nil
}

// ---------- helpers ----------

func parseNumeric(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "%")
	s = strings.ReplaceAll(s, ",", "")
	v, err := strconv.ParseFloat(s, 64)
	return v, err == nil
}

// pickProxySymbol 根据事件货币选择代理标的（必须存在于 price_intraday）。
func pickProxySymbol(ccy string) string {
	switch strings.ToUpper(ccy) {
	case "USD":
		return "EURUSD"
	case "EUR":
		return "EURUSD"
	case "GBP":
		return "GBPUSD"
	case "JPY":
		return "USDJPY"
	default:
		return ""
	}
}

func nearestIntradayClose(ctx context.Context, db *pgxpool.Pool, symbol string, target time.Time) (float64, bool) {
	var v float64
	err := db.QueryRow(ctx, `
		SELECT close FROM price_intraday
		 WHERE symbol=$1 AND interval='5m' AND time BETWEEN $2 AND $3
		 ORDER BY ABS(EXTRACT(EPOCH FROM (time-$2))) ASC LIMIT 1`,
		symbol, target, target.Add(2*time.Hour)).Scan(&v)
	if err != nil {
		return 0, false
	}
	return v, true
}
