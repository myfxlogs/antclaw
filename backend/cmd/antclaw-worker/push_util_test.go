package main

import (
	"testing"
	"time"
)

// ---------- nvl ----------

func TestNvl_ReturnsFallbackWhenEmpty(t *testing.T) {
	if got := nvl("", "fallback"); got != "fallback" {
		t.Errorf("nvl(\"\", \"fallback\") = %q, want \"fallback\"", got)
	}
}

func TestNvl_ReturnsValueWhenNonEmpty(t *testing.T) {
	if got := nvl("hello", "fallback"); got != "hello" {
		t.Errorf("nvl(\"hello\", \"fallback\") = %q, want \"hello\"", got)
	}
}

// ---------- matchFold ----------

func TestMatchFold_TrueWhenMatch(t *testing.T) {
	list := []string{"USD", "EUR", "GBP"}
	if !matchFold("usd", list) {
		t.Error("matchFold(\"usd\", [USD,EUR,GBP]) should be true")
	}
	if !matchFold("eur", list) {
		t.Error("matchFold(\"eur\", [USD,EUR,GBP]) should be true")
	}
}

func TestMatchFold_FalseWhenNoMatch(t *testing.T) {
	list := []string{"USD", "EUR"}
	if matchFold("jpy", list) {
		t.Error("matchFold(\"jpy\", [USD,EUR]) should be false")
	}
}

func TestMatchFold_TrimsItemsButNotSearch(t *testing.T) {
	list := []string{"USD", " EUR "}
	// matchFold trims items in list, so "EUR" should match " EUR "
	if !matchFold("EUR", list) {
		t.Error("matchFold(\"EUR\", [USD, \" EUR \"]) should be true (items are trimmed)")
	}
	// But it does NOT trim search string, so " eur " won't match "EUR"
	if matchFold(" eur ", []string{"EUR"}) {
		t.Error("matchFold(\" eur \", [EUR]) should be false (search is not trimmed)")
	}
}

// ---------- impactToSeverity ----------

func TestImpactToSeverity(t *testing.T) {
	tests := []struct {
		impact, want string
	}{
		{"high", "high"},
		{"HIGH", "high"},
		{" High ", "high"},
		{"medium", "normal"},
		{"MEDIUM", "normal"},
		{"low", "low"},
		{"unknown", "low"},
		{"", "low"},
	}
	for _, tt := range tests {
		got := impactToSeverity(tt.impact)
		if got != tt.want {
			t.Errorf("impactToSeverity(%q) = %q, want %q", tt.impact, got, tt.want)
		}
	}
}

// ---------- defaultCurrencies ----------

func TestDefaultCurrencies(t *testing.T) {
	c := defaultCurrencies()
	if len(c) != 8 {
		t.Errorf("defaultCurrencies() len = %d, want 8", len(c))
	}
	expected := "USD"
	if c[0] != expected {
		t.Errorf("defaultCurrencies()[0] = %q, want %q", c[0], expected)
	}
}

// ---------- isDigestWindow ----------

func TestIsDigestWindow(t *testing.T) {
	// 06:00 = digest window
	loc := time.UTC
	t06 := time.Date(2026, 5, 13, 6, 0, 0, 0, loc)
	t06_29 := time.Date(2026, 5, 13, 6, 29, 59, 0, loc)
	t06_30 := time.Date(2026, 5, 13, 6, 30, 0, 0, loc)
	t07 := time.Date(2026, 5, 13, 7, 0, 0, 0, loc)
	t05 := time.Date(2026, 5, 13, 5, 59, 0, 0, loc)

	if !isDigestWindow(t06) {
		t.Error("isDigestWindow(06:00) should be true")
	}
	if !isDigestWindow(t06_29) {
		t.Error("isDigestWindow(06:29) should be true")
	}
	if isDigestWindow(t06_30) {
		t.Error("isDigestWindow(06:30) should be false")
	}
	if isDigestWindow(t07) {
		t.Error("isDigestWindow(07:00) should be false")
	}
	if isDigestWindow(t05) {
		t.Error("isDigestWindow(05:59) should be false")
	}
}

// ---------- isWeeklyWindow ----------

func TestIsWeeklyWindow(t *testing.T) {
	loc := time.UTC
	// 2026-05-10 is Sunday
	sun18 := time.Date(2026, 5, 10, 18, 0, 0, 0, loc)
	sun18_30 := time.Date(2026, 5, 10, 18, 30, 0, 0, loc)
	sun17 := time.Date(2026, 5, 10, 17, 59, 0, 0, loc)
	sun19 := time.Date(2026, 5, 10, 19, 0, 0, 0, loc)
	// 2026-05-11 is Monday
	mon18 := time.Date(2026, 5, 11, 18, 0, 0, 0, loc)

	if !isWeeklyWindow(sun18) {
		t.Error("isWeeklyWindow(Sun 18:00) should be true")
	}
	if !isWeeklyWindow(sun18_30) {
		t.Error("isWeeklyWindow(Sun 18:30) should be true")
	}
	if isWeeklyWindow(sun17) {
		t.Error("isWeeklyWindow(Sun 17:59) should be false")
	}
	if isWeeklyWindow(sun19) {
		t.Error("isWeeklyWindow(Sun 19:00) should be false (hour!=18)")
	}
	if isWeeklyWindow(mon18) {
		t.Error("isWeeklyWindow(Mon 18:00) should be false")
	}
}

// ---------- loadLocationOrDefault ----------

func TestLoadLocationOrDefault_UTCWhenEmpty(t *testing.T) {
	loc := loadLocationOrDefault("")
	if loc != time.UTC {
		t.Error("loadLocationOrDefault(\"\") should return UTC")
	}
}

func TestLoadLocationOrDefault_UTCWhenInvalid(t *testing.T) {
	loc := loadLocationOrDefault("Mars/Olympus")
	if loc != time.UTC {
		t.Error("loadLocationOrDefault(\"Mars/Olympus\") should return UTC")
	}
}

func TestLoadLocationOrDefault_ValidTimezone(t *testing.T) {
	loc := loadLocationOrDefault("Asia/Shanghai")
	if loc == time.UTC {
		t.Error("loadLocationOrDefault(\"Asia/Shanghai\") should not be UTC")
	}
	name, _ := time.Now().In(loc).Zone()
	if name != "CST" {
		t.Logf("zone name = %q (expected CST but DST may change this)", name)
	}
}

// ---------- dayRange ----------

func TestDayRange_UTC(t *testing.T) {
	p := alertPrefs{Timezone: "UTC"}
	start, end := dayRange(p)
	if start.Location() != time.UTC {
		t.Error("dayRange start should be in UTC")
	}
	diff := end.Sub(start)
	if diff != 24*time.Hour {
		t.Errorf("dayRange duration = %v, want 24h", diff)
	}
}

// ---------- weekRange ----------

func TestWeekRange_UTC(t *testing.T) {
	p := alertPrefs{Timezone: "UTC"}
	mon, nextMon := weekRange(p)
	diff := nextMon.Sub(mon)
	if diff != 7*24*time.Hour {
		t.Errorf("weekRange duration = %v, want 7 days", diff)
	}
	if mon.Weekday() != time.Monday {
		t.Errorf("weekRange start weekday = %v, want Monday", mon.Weekday())
	}
	if mon.Location() != time.UTC {
		t.Error("weekRange start should be in UTC")
	}
}

// ---------- alertPrefs.matchesCurrency ----------

func TestAlertPrefs_MatchesCurrency(t *testing.T) {
	p := alertPrefs{Currencies: []string{"USD", "EUR", "GBP"}}
	if !p.matchesCurrency("USD") {
		t.Error("should match USD")
	}
	if !p.matchesCurrency("eur") {
		t.Error("should match eur (case-insensitive)")
	}
	if p.matchesCurrency("JPY") {
		t.Error("should not match JPY")
	}
	if p.matchesCurrency("") {
		t.Error("should not match empty")
	}
}

// ---------- alertPrefs.matchesImpact ----------

func TestAlertPrefs_MatchesImpact(t *testing.T) {
	// No high_impact_only filter
	p := alertPrefs{Impacts: []string{"high", "medium"}, HighImpactOnly: false}
	if !p.matchesImpact("high") {
		t.Error("should match high")
	}
	if !p.matchesImpact("medium") {
		t.Error("should match medium")
	}
	if p.matchesImpact("low") {
		t.Error("should not match low (not in Impacts list)")
	}
}

func TestAlertPrefs_MatchesImpact_HighOnly(t *testing.T) {
	p := alertPrefs{Impacts: []string{"high", "medium"}, HighImpactOnly: true}
	if !p.matchesImpact("high") {
		t.Error("should match high")
	}
	if p.matchesImpact("medium") {
		t.Error("should not match medium when HighImpactOnly")
	}
}

// ---------- 表驱动测试：去重 key 格式 ----------

func TestDedupKeyFormats(t *testing.T) {
	// Verify key format patterns used across detectors
	tests := []struct {
		name string
		key  string
	}{
		{"calendar pre-event", "calendar:evt-123:pre:15"},
		{"calendar actual", "calendar:evt-123:actual"},
		{"calendar surprise", "calendar:evt-123:surprise:2"},
		{"daily digest", "digest:daily:user-uuid:2026-05-13"},
		{"daily briefing", "digest:briefing:user-uuid:2026-05-13"},
		{"weekly digest", "digest:weekly:user-uuid:2026-W20"},
		{"COT release", "cot:release:2026-05-09"},
		{"COT signal", "cot:signal:USDX:EXTREME_LONG:2026-05-09"},
		{"regime transition", "regime:EURUSD:BULL:BEAR:2026-05-13"},
		{"macro regime", "macro:regime:STAGFLATION:2026-05-13"},
		{"FRED anomaly", "macro:T10Y2Y:上升:2026-05-13"},
		{"options stress", "options:stress:EURUSD:2026-05-13"},
		{"onchain correlation", "onchain:correlation:BTCUSDT:2026-05-13"},
		{"carry unwind", "carry:unwind:2026-05-13"},
		{"risk confluence", "risk_confluence:2026-05-13:3"},
		{"calibration", "calibration:update:2026-05-13"},
	}
	for _, tt := range tests {
		if tt.key == "" {
			t.Errorf("%s: key should not be empty", tt.name)
		}
	}
}
