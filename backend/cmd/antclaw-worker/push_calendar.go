// 客户端智能推送：经济日历事件检测。
//
// 频率：每 1 分钟（由 push_scheduler.go 的 t1m ticker 调度）
//
// 使用 push_util.go 中的 scanUsers / sendIfNotPushed / countEventsInRange / alertPrefs。
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/antclaw/antclaw/internal/notify"
)

// ---------- 调度入口 ----------

func pushCalendar(env *pushEnv) func(context.Context, *pushEnv) {
	return func(ctx context.Context, e *pushEnv) {
		started := time.Now()
		now := time.Now().UTC()
		windowEnd := now.Add(120 * time.Minute)

		events, err := e.scanCalendarEvents(ctx, now, windowEnd)
		if err != nil {
			e.log.Error("calendar: scan events failed", "error", err)
			return
		}
		e.log.Debug("calendar: events in window", "count", len(events))

		sentPre, sentActual, sentSurprise := 0, 0, 0
		e.scanUsers(ctx, func(uid uuid.UUID, p alertPrefs) bool {
			for _, evt := range events {
				sentPre += e.tryPreEventNotify(ctx, uid, evt, p)
				if e.tryActualNotify(ctx, uid, evt, p) {
					sentActual++
				}
				if e.trySurpriseNotify(ctx, uid, evt, p) {
					sentSurprise++
				}
			}
			return false // 不中断扫描
		})

		e.log.Info("calendar: scan complete",
			"events", len(events),
			"sent_pre", sentPre, "sent_actual", sentActual, "sent_surprise", sentSurprise,
			"elapsed", time.Since(started).Round(time.Millisecond),
		)
	}
}

// ---------- 事件扫描 ----------

type calendarEvent struct {
	EventID       string
	Title         string
	Currency      string
	Impact        string
	ScheduledAt   time.Time
	ForecastVal   string
	PreviousVal   string
	ActualVal     *string
	SurpriseScore *float64
	SurpriseLabel *string
}

func (e *pushEnv) scanCalendarEvents(ctx context.Context, now, end time.Time) ([]calendarEvent, error) {
	rows, err := e.pool.Query(ctx, `
		SELECT event_id, title, COALESCE(currency,''), COALESCE(impact,'low'),
		       scheduled_at,
		       COALESCE(forecast_value,''), COALESCE(previous_value,''),
		       actual_value, surprise_score, surprise_label
		  FROM calendar_events
		 WHERE scheduled_at BETWEEN $1 AND $2
		 ORDER BY scheduled_at`, now, end)
	if err != nil {
		return nil, fmt.Errorf("scan calendar_events: %w", err)
	}
	defer rows.Close()

	var out []calendarEvent
	for rows.Next() {
		var evt calendarEvent
		if err := rows.Scan(
			&evt.EventID, &evt.Title, &evt.Currency, &evt.Impact,
			&evt.ScheduledAt,
			&evt.ForecastVal, &evt.PreviousVal,
			&evt.ActualVal, &evt.SurpriseScore, &evt.SurpriseLabel,
		); err != nil {
			return nil, fmt.Errorf("scan calendar row: %w", err)
		}
		out = append(out, evt)
	}
	return out, rows.Err()
}

// ---------- 事件前提醒 ----------

func (e *pushEnv) tryPreEventNotify(ctx context.Context, userID uuid.UUID, evt calendarEvent, p alertPrefs) int {
	if !p.matchesCurrency(evt.Currency) || !p.matchesImpact(evt.Impact) {
		return 0
	}
	now := time.Now().UTC()
	sent := 0
	for _, mins := range p.ReminderMinutes {
		target := evt.ScheduledAt.Add(-time.Duration(mins) * time.Minute)
		if now.Before(target.Add(-30*time.Second)) || now.After(target.Add(30*time.Second)) {
			continue
		}
		ek := fmt.Sprintf("calendar:%s:pre:%d", evt.EventID, mins)
		ok := e.sendIfNotPushed(ctx, userID, ek, "calendar_pre", &notify.Notification{
			UserID:   userID,
			Category: "alert",
			Title:    fmt.Sprintf("%s 重要数据即将公布", evt.Currency),
			Body:     fmt.Sprintf("%s 将在 %d 分钟后公布。影响级别：%s，预测值：%s，前值：%s。", evt.Title, mins, evt.Impact, nvl(evt.ForecastVal, "—"), nvl(evt.PreviousVal, "—")),
			Severity: impactToSeverity(evt.Impact),
			Data:     map[string]string{"kind": "calendar_pre_event", "event_id": evt.EventID, "currency": evt.Currency, "impact": evt.Impact, "minutes": fmt.Sprintf("%d", mins)},
		})
		if ok {
			sent++
		}
	}
	return sent
}

// ---------- actual 公布提醒 ----------

func (e *pushEnv) tryActualNotify(ctx context.Context, userID uuid.UUID, evt calendarEvent, p alertPrefs) bool {
	if !p.matchesCurrency(evt.Currency) || !p.matchesImpact(evt.Impact) {
		return false
	}
	if evt.ActualVal == nil || *evt.ActualVal == "" {
		return false
	}
	ek := "calendar:" + evt.EventID + ":actual"
	return e.sendIfNotPushed(ctx, userID, ek, "calendar_actual", &notify.Notification{
		UserID:   userID,
		Category: "alert",
		Title:    fmt.Sprintf("%s 数据已公布", evt.Currency),
		Body:     fmt.Sprintf("%s actual=%s，forecast=%s，previous=%s。", evt.Title, *evt.ActualVal, nvl(evt.ForecastVal, "—"), nvl(evt.PreviousVal, "—")),
		Severity: impactToSeverity(evt.Impact),
		Data:     map[string]string{"kind": "calendar_actual", "event_id": evt.EventID, "currency": evt.Currency, "actual": *evt.ActualVal},
	})
}

// ---------- surprise 高阈值 ----------

func (e *pushEnv) trySurpriseNotify(ctx context.Context, userID uuid.UUID, evt calendarEvent, p alertPrefs) bool {
	if !p.matchesCurrency(evt.Currency) || !p.matchesImpact(evt.Impact) {
		return false
	}
	if evt.SurpriseScore == nil {
		return false
	}
	abs := *evt.SurpriseScore
	if abs < 0 {
		abs = -abs
	}
	if abs < 2.0 {
		return false
	}

	ek := fmt.Sprintf("calendar:%s:surprise:%.0f", evt.EventID, abs)
	sev := "normal"
	if evt.Impact == "high" {
		sev = "critical"
	}
	return e.sendIfNotPushed(ctx, userID, ek, "surprise", &notify.Notification{
		UserID:   userID,
		Category: "signal",
		Title:    fmt.Sprintf("%s 数据意外超预期", evt.Currency),
		Body:     fmt.Sprintf("%s surprise=%.1fσ。", evt.Title, *evt.SurpriseScore),
		Severity: sev,
		Data:     map[string]string{"kind": "calendar_surprise", "event_id": evt.EventID, "currency": evt.Currency, "surprise_score": fmt.Sprintf("%.2f", *evt.SurpriseScore)},
	})
}
