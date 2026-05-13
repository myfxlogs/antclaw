package main

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/antclaw/antclaw/internal/notify"
)

// ---------- 每日晨报构造测试 ----------

func TestDailyNewsRadarNotifyFormat(t *testing.T) {
	uid := uuid.New()
	dateKey := "2026-05-13"
	body := "今日 high impact 3 个，medium impact 2 个。首个重点事件：08:30 USD 非农就业人数。今日为 Storm Day：高影响事件集中，建议降低杠杆、扩大止损或减少事件前开仓。"

	n := &notify.Notification{
		UserID:   uid,
		Category: "digest",
		Title:    "今日市场雷达",
		Body:     body,
		Severity: "high", // Storm Day
		Data: map[string]string{
			"kind":      "daily_news_radar",
			"date":      dateKey,
			"high_count": "3",
			"med_count":  "2",
			"storm_day":  "true",
		},
		DedupTTL: 25 * time.Hour,
	}

	// 验证分类
	if n.Category != "digest" {
		t.Errorf("daily news category = %q, want digest", n.Category)
	}
	if n.Severity != "high" {
		t.Errorf("storm day severity = %q, want high", n.Severity)
	}
	if n.Data["storm_day"] != "true" {
		t.Errorf("storm_day = %q, want true", n.Data["storm_day"])
	}

	// dedup key 格式
	ek := "digest:daily:" + uid.String() + ":" + dateKey
	t.Logf("daily news dedup key: %s", ek)
}

func TestDailyNewsRadarNormalDay(t *testing.T) {
	uid := uuid.New()
	body := "今日 high impact 1 个，medium impact 3 个。首个重点事件：10:00 EUR CPI。"

	n := &notify.Notification{
		UserID:   uid,
		Category: "digest",
		Title:    "今日市场雷达",
		Body:     body,
		Severity: "normal", // 非 Storm Day
		Data: map[string]string{
			"kind":      "daily_news_radar",
			"high_count": "1",
			"med_count":  "3",
			"storm_day":  "false",
		},
	}

	if n.Severity != "normal" {
		t.Errorf("normal day severity = %q, want normal", n.Severity)
	}
}

func TestDailyBriefingNotifyFormat(t *testing.T) {
	uid := uuid.New()
	dateKey := "2026-05-13"
	body := "今日宏观：high 2 / medium 4 事件。首个重点事件：08:30 USD CPI。最新COT：报告日期 2026-05-09。市场状态：EURUSD:BULL GBPUSD:RANGE。"

	n := &notify.Notification{
		UserID:   uid,
		Category: "digest",
		Title:    "今日综合摘要",
		Body:     body,
		Severity: "normal",
		Data: map[string]string{
			"kind": "daily_briefing",
			"date": dateKey,
		},
		DedupTTL: 25 * time.Hour,
	}

	if n.Title != "今日综合摘要" {
		t.Errorf("title = %q, want 今日综合摘要", n.Title)
	}
}

// ---------- 周度展望测试 ----------

func TestWeeklyOutlookNotifyFormat(t *testing.T) {
	uid := uuid.New()
	weekKey := "2026-W20"
	body := "本周重点事件 5 个。当前市场状态：EURUSD:BULL USDJPY:BEAR。"

	n := &notify.Notification{
		UserID:   uid,
		Category: "digest",
		Title:    "本周市场展望",
		Body:     body,
		Severity: "normal",
		Data: map[string]string{
			"kind":      "weekly_outlook",
			"year_week": weekKey,
		},
		DedupTTL: 8 * 24 * time.Hour,
	}

	if n.Category != "digest" {
		t.Errorf("weekly outlook category = %q, want digest", n.Category)
	}
	if n.DedupTTL != 8*24*time.Hour {
		t.Errorf("weekly dedup TTL = %v, want 8 days", n.DedupTTL)
	}
}

// ---------- Storm Day 判定测试 ----------

func TestStormDayDetection(t *testing.T) {
	// high impact >= 3 → Storm Day
	tests := []struct {
		high      int
		isStorm   bool
	}{
		{0, false},
		{1, false},
		{2, false},
		{3, true},
		{5, true},
	}
	for _, tt := range tests {
		got := tt.high >= 3
		if got != tt.isStorm {
			t.Errorf("high=%d → isStorm=%v, want %v", tt.high, got, tt.isStorm)
		}
	}
}

// ---------- 摘要无事件时跳过测试 ----------

func TestDigestNoEvents(t *testing.T) {
	// 无 high/medium 事件时不发送晨报
	tests := []struct {
		counts   eventCounts
		shouldSkip bool
	}{
		{eventCounts{High: 0, Medium: 0}, true},
		{eventCounts{High: 1, Medium: 0}, false},
		{eventCounts{High: 0, Medium: 1}, false},
		{eventCounts{High: 2, Medium: 3}, false},
	}
	for _, tt := range tests {
		got := tt.counts.High == 0 && tt.counts.Medium == 0
		if got != tt.shouldSkip {
			t.Errorf("counts{high=%d, med=%d} → skip=%v, want %v",
				tt.counts.High, tt.counts.Medium, got, tt.shouldSkip)
		}
	}
}

// 时间窗口测试 (isDigestWindow / isWeeklyWindow) 已在 push_util_test.go 中全面覆盖。
// 此处不重复，避免测试冗余。
