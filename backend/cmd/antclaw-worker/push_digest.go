// 客户端智能推送：每日摘要与周度展望。
//
// 使用 push_util.go 中的 scanUsers / sendIfNotPushed / countEventsInRange / dayRange / weekRange。
package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/antclaw/antclaw/internal/notify"
)

// ---------- 每日晨报 + 综合 Briefing ----------

func pushDailyDigest(env *pushEnv) func(context.Context, *pushEnv) {
	return func(ctx context.Context, e *pushEnv) {
		now := time.Now().UTC()
		sentNews, sentBriefing := 0, 0
		e.scanUsers(ctx, func(uid uuid.UUID, p alertPrefs) bool {
			if !p.DailyDigest {
				return false
			}
			local := now.In(loadLocationOrDefault(p.Timezone))
			if !isDigestWindow(local) {
				return false
			}
			dateKey := local.Format("2006-01-02")
			if e.sendDailyNewsRadar(ctx, uid, dateKey, p) {
				sentNews++
			}
			if e.sendDailyBriefing(ctx, uid, dateKey, p) {
				sentBriefing++
			}
			return false
		})
		e.log.Info("digest: daily scan complete", "sent_news", sentNews, "sent_briefing", sentBriefing)
	}
}

func (e *pushEnv) sendDailyNewsRadar(ctx context.Context, userID uuid.UUID, dateKey string, p alertPrefs) bool {
	start, end := dayRange(p)
	counts, first := e.countEventsInRange(ctx, start, end, p.Currencies)
	if counts.High == 0 && counts.Medium == 0 {
		return false
	}

	body := fmt.Sprintf("今日 high impact %d 个，medium impact %d 个。", counts.High, counts.Medium)
	if first != "" {
		body += " " + first
	}
	isStorm := counts.High >= 3
	sev := "normal"
	if isStorm {
		body += " 今日为 Storm Day：高影响事件集中，建议降低杠杆、扩大止损或减少事件前开仓。"
		sev = "high"
	}

	return e.sendIfNotPushed(ctx, userID, fmt.Sprintf("digest:daily:%s:%s", userID, dateKey), "daily_news", &notify.Notification{
		UserID: userID, Category: "digest", Title: "今日市场雷达", Body: body, Severity: sev,
		Data:     map[string]string{"kind": "daily_news_radar", "date": dateKey, "high_count": fmt.Sprint(counts.High), "med_count": fmt.Sprint(counts.Medium), "storm_day": fmt.Sprint(isStorm)},
		DedupTTL: 25 * time.Hour,
	})
}

func (e *pushEnv) sendDailyBriefing(ctx context.Context, userID uuid.UUID, dateKey string, p alertPrefs) bool {
	body := e.buildBriefingBody(ctx, p)
	return e.sendIfNotPushed(ctx, userID, fmt.Sprintf("digest:briefing:%s:%s", userID, dateKey), "daily_briefing", &notify.Notification{
		UserID: userID, Category: "digest", Title: "今日综合摘要", Body: body, Severity: "normal",
		Data:     map[string]string{"kind": "daily_briefing", "date": dateKey},
		DedupTTL: 25 * time.Hour,
	})
}

func (e *pushEnv) buildBriefingBody(ctx context.Context, p alertPrefs) string {
	var parts []string
	start, end := dayRange(p)
	counts, first := e.countEventsInRange(ctx, start, end, p.Currencies)
	if counts.High+counts.Medium > 0 {
		parts = append(parts, fmt.Sprintf("今日宏观：high %d / medium %d 事件", counts.High, counts.Medium))
		if first != "" {
			parts = append(parts, first)
		}
	}
	if cot := e.latestCOTDate(ctx); cot != "" {
		parts = append(parts, "最新COT："+cot)
	}
	if reg := e.latestRegimeStatus(ctx); reg != "" {
		parts = append(parts, "市场状态："+reg)
	}
	if len(parts) == 0 {
		return "今日暂无特殊摘要。"
	}
	return strings.Join(parts, "。") + "。"
}

// ---------- 周度展望 ----------

func pushWeeklyOutlook(env *pushEnv) func(context.Context, *pushEnv) {
	return func(ctx context.Context, e *pushEnv) {
		now := time.Now().UTC()
		sent := 0
		e.scanUsers(ctx, func(uid uuid.UUID, p alertPrefs) bool {
			if !p.WeeklyDigest {
				return false
			}
			local := now.In(loadLocationOrDefault(p.Timezone))
			if !isWeeklyWindow(local) {
				return false
			}
			year, week := local.ISOWeek()
			weekKey := fmt.Sprintf("%d-W%02d", year, week)
			body := e.buildWeeklyBody(ctx, p)
			if e.sendIfNotPushed(ctx, uid, fmt.Sprintf("digest:weekly:%s:%s", uid, weekKey), "weekly_outlook", &notify.Notification{
				UserID: uid, Category: "digest", Title: "本周市场展望", Body: body, Severity: "normal",
				Data:     map[string]string{"kind": "weekly_outlook", "year_week": weekKey},
				DedupTTL: 8 * 24 * time.Hour,
			}) {
				sent++
			}
			return false
		})
		e.log.Info("weekly: outlook scan complete", "sent", sent)
	}
}

func (e *pushEnv) buildWeeklyBody(ctx context.Context, p alertPrefs) string {
	var parts []string
	mon, sun := weekRange(p)
	counts, _ := e.countEventsInRange(ctx, mon, sun, p.Currencies)
	parts = append(parts, fmt.Sprintf("本周重点事件 %d 个", counts.High))
	if reg := e.latestRegimeStatus(ctx); reg != "" {
		parts = append(parts, "当前市场状态："+reg)
	}
	if len(parts) == 0 {
		return "本周暂无特别展望。"
	}
	return strings.Join(parts, "。") + "。"
}
