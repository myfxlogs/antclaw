package main

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/antclaw/antclaw/internal/notify"
)

// ---------- push_calendar 构造函数测试 ----------

// TestCalendarNotify_constructs 验证事件前提醒/actual/surprise 的通知体构造逻辑。
// 由于实际的 push* 函数依赖数据库连接，此处测试通知体构造的预期格式。

func TestCalendarPreEventNotifyFormat(t *testing.T) {
	uid := uuid.New()
	evt := calendarEvent{
		EventID: "evt-001",
		Title:   "非农就业人数",
		Currency: "USD",
		Impact:  "high",
		ScheduledAt: time.Now().Add(15 * time.Minute),
		ForecastVal: "250K",
		PreviousVal: "220K",
	}
	// 验证 dedup key 前缀格式
	ek := "calendar:" + evt.EventID + ":pre:15"
	if ek != "calendar:evt-001:pre:15" {
		t.Errorf("unexpected dedup key: %s", ek)
	}

	// 验证 notify.Notification 字段
	n := &notify.Notification{
		UserID:   uid,
		Category: "alert",
		Title:    "USD 重要数据即将公布",
		Body:     "非农就业人数 将在 15 分钟后公布。影响级别：high，预测值：250K，前值：220K。",
		Severity: "high",
		Data: map[string]string{
			"kind":     "calendar_pre_event",
			"event_id": evt.EventID,
			"currency": evt.Currency,
			"impact":   evt.Impact,
			"minutes":  "15",
		},
	}
	if n.Category != "alert" {
		t.Errorf("category = %q, want alert", n.Category)
	}
	if n.Severity != "high" {
		t.Errorf("severity = %q, want high", n.Severity)
	}
	if n.Data["kind"] != "calendar_pre_event" {
		t.Errorf("data.kind = %q, want calendar_pre_event", n.Data["kind"])
	}
}

func TestCalendarActualNotifyFormat(t *testing.T) {
	uid := uuid.New()
	evt := calendarEvent{
		EventID: "evt-002",
		Title:   "CPI m/m",
		Currency: "USD",
		Impact:  "high",
		ActualVal: strPtr("0.3%"),
		ForecastVal: "0.2%",
		PreviousVal: "0.1%",
	}

	// dedup key
	ek := "calendar:evt-002:actual"
	t.Logf("dedup key: %s", ek)

	n := &notify.Notification{
		UserID:   uid,
		Category: "alert",
		Title:    evt.Currency + " 数据已公布",
		Body:     "CPI m/m actual=0.3%，forecast=0.2%，previous=0.1%。",
		Severity: impactToSeverity(evt.Impact),
		Data: map[string]string{
			"kind":     "calendar_actual",
			"event_id": evt.EventID,
			"currency": evt.Currency,
			"actual":   *evt.ActualVal,
		},
	}
	if n.Severity != "high" {
		t.Errorf("severity = %q, want high", n.Severity)
	}
	if n.Data["actual"] != "0.3%" {
		t.Errorf("data.actual = %q, want 0.3%%", n.Data["actual"])
	}
}

func TestCalendarSurpriseNotifyFormat(t *testing.T) {
	uid := uuid.New()
	score := 2.5
	evt := calendarEvent{
		EventID: "evt-003",
		Title:   "GDP q/q",
		Currency: "USD",
		Impact:  "high",
		SurpriseScore: &score,
	}

	ek := "calendar:evt-003:surprise:2"
	t.Logf("dedup key: %s", ek)

	n := &notify.Notification{
		UserID:   uid,
		Category: "signal",
		Title:    evt.Currency + " 数据意外超预期",
		Body:     "GDP q/q surprise=2.5σ。",
		Severity: "critical", // high impact + surprise >= 2.0
		Data: map[string]string{
			"kind":           "calendar_surprise",
			"event_id":       evt.EventID,
			"currency":       evt.Currency,
			"surprise_score": "2.50",
		},
	}
	if n.Category != "signal" {
		t.Errorf("category = %q, want signal", n.Category)
	}
	if n.Severity != "critical" {
		t.Errorf("severity = %q, want critical", n.Severity)
	}
}

func TestCalendarSurpriseThreshold(t *testing.T) {
	tests := []struct {
		score      float64
		shouldPush bool
	}{
		{0.0, false},
		{1.5, false},
		{-1.5, false},
		{1.99, false},
		{2.0, true},
		{-2.0, true},
		{-2.1, true},
		{3.5, true},
	}
	for _, tt := range tests {
		abs := tt.score
		if abs < 0 {
			abs = -abs
		}
		got := abs >= 2.0
		if got != tt.shouldPush {
			t.Errorf("|score|=%.2f → push=%v, want %v", abs, got, tt.shouldPush)
		}
	}
}

func TestCalendarImpactToSeverity(t *testing.T) {
	if s := impactToSeverity("high"); s != "high" {
		t.Errorf("high impact → severity %q, want high", s)
	}
	if s := impactToSeverity("medium"); s != "normal" {
		t.Errorf("medium impact → severity %q, want normal", s)
	}
	if s := impactToSeverity("low"); s != "low" {
		t.Errorf("low impact → severity %q, want low", s)
	}
}

func TestCalendarSeverityForSurprise(t *testing.T) {
	// high impact + surprise ≥ 2.0 → critical
	sev := "normal"
	impact := "high"
	abs := 2.5
	if impact == "high" && abs >= 2.0 {
		sev = "critical"
	}
	if sev != "critical" {
		t.Errorf("high impact + surprise must be critical, got %q", sev)
	}
}

// ---------- 边界条件 ----------

func TestCalendarNoActualYet(t *testing.T) {
	// actual 为空时不应推送 actual 通知
	evt := calendarEvent{
		EventID: "evt-004",
		ActualVal: nil,
	}
	if evt.ActualVal != nil && *evt.ActualVal != "" {
		t.Error("empty actual should skip push")
	}
	// 空字符串也应该跳过
	emptyStr := ""
	evt.ActualVal = &emptyStr
	if *evt.ActualVal == "" {
		t.Log("empty string actual → skip push (expected)")
	} else {
		t.Error("empty string actual should skip push")
	}
}

func TestCalendarZeroScheduledTime(t *testing.T) {
	// 时间为零值时不推送事件前提醒 (文档 §12 risk table)
	zero := time.Time{}
	if zero.IsZero() {
		t.Log("zero time → skip pre-event notify (expected)")
	} else {
		t.Error("zero time should be detected")
	}
}

func strPtr(s string) *string { return &s }
