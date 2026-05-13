// 客户端智能推送：公共类型与辅助函数。
//
// 所有 push_*.go 共享：
//   - alertPrefs 类型及其构造/过滤方法
//   - 用户分页扫描 scanUsers
//   - 事件计数查询 countEventsInRange
//   - 去重安全推送 sendIfNotPushed
//   - 时间工具函数
package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/antclaw/antclaw/internal/adapter/storage/postgres/db"
	"github.com/antclaw/antclaw/internal/notify"
)

// ---------- 用户偏好 ----------

// alertPrefs 是 ListUsersWithAlertPrefsRow 的简化视图，避免在各 push 文件中
// 直接依赖 db 层类型。
type alertPrefs struct {
	Currencies      []string
	Symbols         []string
	Impacts         []string
	ReminderMinutes []int32
	HighImpactOnly  bool
	DailyDigest     bool
	WeeklyDigest    bool
	CotAlerts       bool
	MacroAlerts     bool
	OptionsAlerts   bool
	OnchainAlerts   bool
	PushEnabled     bool
	MinSeverity     string
	QuietStart      string
	QuietEnd        string
	Timezone        string
}

func userAlertPrefsFromRow(r db.ListUsersWithAlertPrefsRow) alertPrefs {
	qs := func(t pgtype.Time) string {
		if !t.Valid {
			return "00:00"
		}
		secs := t.Microseconds / 1_000_000
		return fmt.Sprintf("%02d:%02d", secs/3600, (secs%3600)/60)
	}
	return alertPrefs{
		Currencies:      r.Currencies,
		Symbols:         r.Symbols,
		Impacts:         r.Impacts,
		ReminderMinutes: r.ReminderMinutes,
		HighImpactOnly:  r.HighImpactOnly,
		DailyDigest:     r.DailyDigestEnabled,
		WeeklyDigest:    r.WeeklyDigestEnabled,
		CotAlerts:       r.CotAlertsEnabled,
		MacroAlerts:     r.MacroAlertsEnabled,
		OptionsAlerts:   r.OptionsAlertsEnabled,
		OnchainAlerts:   r.OnchainAlertsEnabled,
		PushEnabled:     r.PushEnabled,
		MinSeverity:     r.MinSeverity,
		QuietStart:      qs(r.QuietStart),
		QuietEnd:        qs(r.QuietEnd),
		Timezone:        r.Timezone,
	}
}

func (p alertPrefs) matchesCurrency(c string) bool {
	return matchFold(strings.ToUpper(strings.TrimSpace(c)), p.Currencies)
}

func (p alertPrefs) matchesImpact(impact string) bool {
	impact = strings.ToLower(strings.TrimSpace(impact))
	if p.HighImpactOnly && impact != "high" {
		return false
	}
	return matchFold(impact, p.Impacts)
}

// ---------- 批量用户扫描 ----------

// scanUsers 遍历所有用户（分页 100），对每个用户执行 fn。
// fn 返回 false 时跳过该用户继续下一个。
func (e *pushEnv) scanUsers(ctx context.Context, fn func(uuid.UUID, alertPrefs) bool) {
	var cursor uuid.UUID
	for {
		users, err := e.q.ListUsersWithAlertPrefs(ctx, db.ListUsersWithAlertPrefsParams{
			ID:    cursor,
			Limit: 100,
		})
		if err != nil {
			e.log.Error("scanUsers: query failed", "error", err, "cursor", cursor)
			return
		}
		if len(users) == 0 {
			return
		}
		for _, u := range users {
			cursor = u.UserID
			if !fn(u.UserID, userAlertPrefsFromRow(u)) {
				continue
			}
		}
		if len(users) < 100 {
			return
		}
	}
}

// ---------- 去重安全推送 ----------

// sendIfNotPushed 先查 push_state 去重，再调 svc.Send，最后记录 push_state。
// 返回是否实际发送。
func (e *pushEnv) sendIfNotPushed(ctx context.Context, userID uuid.UUID, eventKey, pushType string, n *notify.Notification) bool {
	if e.alreadyPushed(ctx, userID, eventKey, pushType) {
		return false
	}
	n.DedupKey = eventKey
	if n.DedupTTL <= 0 {
		n.DedupTTL = 10 * time.Minute
	}
	if err := e.svc.Send(ctx, n); err != nil {
		e.log.Error("push: send failed", "user", userID, "type", pushType, "key", eventKey, "error", err)
		return false
	}
	e.recordPush(ctx, userID, eventKey, pushType)
	return true
}

func (e *pushEnv) alreadyPushed(ctx context.Context, userID uuid.UUID, eventKey, pushType string) bool {
	_, err := e.q.GetPushState(ctx, db.GetPushStateParams{
		UserID:   userID,
		EventKey: eventKey,
		PushType: pushType,
	})
	return err == nil
}

func (e *pushEnv) recordPush(ctx context.Context, userID uuid.UUID, eventKey, pushType string) {
	_, _ = e.q.UpsertPushState(ctx, db.UpsertPushStateParams{
		UserID:     userID,
		EventKey:   eventKey,
		PushType:   pushType,
		LastSentAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		PayloadHash: "",
	})
}

// ---------- 事件统计 ----------

// eventCounts 统计某时间段内的日历事件数量。
type eventCounts struct{ High, Medium int }

// countEventsInRange 查询 calendar_events 在 [start, end) 内的 impact 统计及首条摘要。
func (e *pushEnv) countEventsInRange(ctx context.Context, start, end time.Time, currencies []string) (eventCounts, string) {
	if len(currencies) == 0 {
		currencies = defaultCurrencies()
	}
	rows, err := e.pool.Query(ctx, `
		SELECT COALESCE(impact,''), COALESCE(title,''), COALESCE(currency,''), scheduled_at
		  FROM calendar_events
		 WHERE scheduled_at >= $1 AND scheduled_at < $2
		   AND impact IN ('high','medium')
		 ORDER BY scheduled_at
		 LIMIT 100`, start, end)
	if err != nil {
		return eventCounts{}, ""
	}
	defer rows.Close()

	var counts eventCounts
	var first string
	for rows.Next() {
		var impact, title, currency string
		var scheduled time.Time
		if err := rows.Scan(&impact, &title, &currency, &scheduled); err != nil {
			continue
		}
		if !matchFold(strings.ToUpper(currency), currencies) {
			continue
		}
		switch impact {
		case "high":
			counts.High++
			if first == "" {
				first = fmt.Sprintf("首个重点事件：%s %s %s", scheduled.Format("15:04"), currency, title)
			}
		case "medium":
			counts.Medium++
		}
	}
	return counts, first
}

// ---------- 聚合查询 ----------

// latestCOTDate 返回最新 COT 报告日期的格式化字符串。
func (e *pushEnv) latestCOTDate(ctx context.Context) string {
	var ts pgtype.Timestamptz
	if err := e.pool.QueryRow(ctx,
		`SELECT report_date FROM cot_analyses ORDER BY report_date DESC LIMIT 1`).Scan(&ts); err != nil || !ts.Valid {
		return ""
	}
	return "报告日期 " + ts.Time.Format("2006-01-02")
}

// latestRegimeStatus 返回最近 3 条 regime_transitions 的 symbol:regime 摘要。
func (e *pushEnv) latestRegimeStatus(ctx context.Context) string {
	rows, err := e.pool.Query(ctx,
		`SELECT symbol, to_label FROM regime_transitions ORDER BY time DESC LIMIT 3`)
	if err != nil {
		return ""
	}
	defer rows.Close()

	var parts []string
	for rows.Next() {
		var sym, regime string
		if err := rows.Scan(&sym, &regime); err != nil {
			continue
		}
		parts = append(parts, sym+":"+regime)
	}
	return strings.Join(parts, " ")
}

// latestCOTReportDate 返回最新 COT 报告时间。
func (e *pushEnv) latestCOTReportDate(ctx context.Context) (time.Time, error) {
	var ts time.Time
	err := e.pool.QueryRow(ctx,
		`SELECT report_date FROM cot_analyses ORDER BY report_date DESC LIMIT 1`).Scan(&ts)
	return ts, err
}

// ---------- 时间工具 ----------

// isDigestWindow 判断当前是否为 06:00-06:30（每日摘要窗口）。
func isDigestWindow(localTime time.Time) bool {
	return localTime.Hour() == 6 && localTime.Minute() < 30
}

// isWeeklyWindow 判断当前是否为周日 18:00-18:59（周度摘要窗口）。
func isWeeklyWindow(localTime time.Time) bool {
	return localTime.Weekday() == time.Sunday && localTime.Hour() == 18
}

// loadLocationOrDefault 安全加载时区，失败回退 UTC。
func loadLocationOrDefault(name string) *time.Location {
	if name == "" {
		return time.UTC
	}
	if loc, err := time.LoadLocation(name); err == nil {
		return loc
	}
	return time.UTC
}

// dayRange 返回用户本地时区当天的 UTC 区间。
func dayRange(prefs alertPrefs) (start, end time.Time) {
	tz := loadLocationOrDefault(prefs.Timezone)
	now := time.Now().In(tz)
	start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, tz).UTC()
	end = start.Add(24 * time.Hour)
	return
}

// weekRange 返回用户本地时区本周一的 00:00（UTC）到下周一 00:00（UTC）。
func weekRange(prefs alertPrefs) (monday, sunday time.Time) {
	tz := loadLocationOrDefault(prefs.Timezone)
	now := time.Now().In(tz)
	wd := now.Weekday()
	if wd == time.Sunday {
		wd = 7
	}
	monday = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, tz).
		Add(-time.Duration(wd-1) * 24 * time.Hour).UTC()
	sunday = monday.Add(7 * 24 * time.Hour)
	return
}

// ---------- 杂项 ----------

func nvl(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

func matchFold(s string, list []string) bool {
	for _, item := range list {
		if strings.EqualFold(s, strings.TrimSpace(item)) {
			return true
		}
	}
	return false
}

func defaultCurrencies() []string {
	return []string{"USD", "EUR", "GBP", "JPY", "CHF", "CAD", "AUD", "NZD"}
}

func impactToSeverity(impact string) string {
	switch strings.ToLower(strings.TrimSpace(impact)) {
	case "high":
		return "high"
	case "medium":
		return "normal"
	default:
		return "low"
	}
}


