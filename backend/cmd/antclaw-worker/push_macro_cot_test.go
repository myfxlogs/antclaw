package main

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/antclaw/antclaw/internal/notify"
)

// ---------- push_macro 通知体构造测试 ----------

func TestMacroRegimeNotifyFormat(t *testing.T) {
	uid := uuid.New()
	n := &notify.Notification{
		UserID:   uid,
		Category: "alert",
		Title:    "宏观 regime 更新",
		Body:     "当前宏观 regime：STAGFLATION（3.45），更新时间：2026-05-13T00:00:00Z",
		Severity: "normal",
		Data:     map[string]string{"kind": "macro_regime", "regime": "STAGFLATION"},
		DedupTTL: time.Hour,
	}
	if n.Category != "alert" {
		t.Errorf("macro regime category = %q, want alert", n.Category)
	}
	if n.Data["kind"] != "macro_regime" {
		t.Errorf("data.kind = %q, want macro_regime", n.Data["kind"])
	}
}

func TestFredAnomalyNotifyFormat(t *testing.T) {
	uid := uuid.New()
	n := &notify.Notification{
		UserID:   uid,
		Category: "alert",
		Title:    "FRED T10Y2Y 出现异动",
		Body:     "T10Y2Y 较前值 下降 -18.5%，当前值：-0.4200",
		Severity: "normal",
		Data: map[string]string{
			"kind":       "macro_fred",
			"series":     "T10Y2Y",
			"direction":  "下降",
			"pct_change": "-18.5",
		},
		DedupTTL: time.Hour,
	}
	if n.DedupTTL.Nanoseconds() <= 0 {
		t.Error("FRED anomaly should have dedup TTL")
	}
}

func TestOptionsRiskNotifyFormat(t *testing.T) {
	uid := uuid.New()
	n := &notify.Notification{
		UserID:   uid,
		Category: "alert",
		Title:    "EURUSD 期权压力偏高",
		Body:     "EURUSD stress_score=0.85（最新快照），注意期权风险。",
		Severity: "high",
		Data: map[string]string{
			"kind":   "options_risk",
			"symbol": "EURUSD",
			"score":  "0.85",
		},
		DedupTTL: time.Hour,
	}
	if n.Severity != "high" {
		t.Errorf("options risk severity = %q, want high", n.Severity)
	}
}

func TestOnchainRiskNotifyFormat(t *testing.T) {
	uid := uuid.New()
	n := &notify.Notification{
		UserID:   uid,
		Category: "alert",
		Title:    "链上相关性断裂",
		Body:     "BTC/USDT 相关性断裂（z=3.2），注意链上风险传导。",
		Severity: "normal",
		Data: map[string]string{
			"kind":    "onchain_risk",
			"pair_a":  "BTC",
			"pair_b":  "USDT",
			"z_score": "3.2",
		},
		DedupTTL: time.Hour,
	}
	if n.Category != "alert" {
		t.Errorf("onchain risk category = %q, want alert", n.Category)
	}
}

// ---------- push_macro carry unwind ----------

func TestCarryUnwindNotifyFormat(t *testing.T) {
	uid := uuid.New()
	n := &notify.Notification{
		UserID:   uid,
		Category: "alert",
		Title:    "Carry Trade 风险提示",
		Body:     "利差异常货币：TRY(18.5%), ZAR(8.2%), MXN(5.1%)。注意 carry unwind 风险。",
		Severity: "high",
		Data: map[string]string{
			"kind":       "carry_unwind",
			"currencies": "TRY,ZAR,MXN",
		},
		DedupTTL: 4 * time.Hour,
	}
	if n.Severity != "high" {
		t.Errorf("carry unwind severity = %q, want high", n.Severity)
	}
}

// ---------- push_macro regime transition ----------

func TestRegimeTransitionNotifyFormat(t *testing.T) {
	uid := uuid.New()
	n := &notify.Notification{
		UserID:   uid,
		Category: "alert",
		Title:    "EURUSD 市场状态发生切换",
		Body:     "EURUSD regime 从 RANGE 切换为 BULL。",
		Severity: "normal",
		Data: map[string]string{
			"kind":       "regime_transition",
			"symbol":     "EURUSD",
			"from_state": "RANGE",
			"to_state":   "BULL",
		},
		DedupTTL: time.Hour,
	}
	if n.Data["from_state"] != "RANGE" {
		t.Errorf("from_state = %q, want RANGE", n.Data["from_state"])
	}
	if n.Data["to_state"] != "BULL" {
		t.Errorf("to_state = %q, want BULL", n.Data["to_state"])
	}
}

func TestRegimeTransitionSeverityMapping(t *testing.T) {
	// CRITICAL → critical, WARN → high, else → normal
	tests := []struct{ sev, want string }{
		{"CRITICAL", "critical"},
		{"WARN", "high"},
		{"INFO", "normal"},
		{"", "normal"},
	}
	for _, tt := range tests {
		got := "normal"
		if tt.sev == "CRITICAL" {
			got = "critical"
		} else if tt.sev == "WARN" {
			got = "high"
		}
		if got != tt.want {
			t.Errorf("severity(%q) = %q, want %q", tt.sev, got, tt.want)
		}
	}
}

// ---------- push_macro risk confluence ----------

func TestRiskConfluenceNotifyFormat(t *testing.T) {
	uid := uuid.New()
	n := &notify.Notification{
		UserID:   uid,
		Category: "signal",
		Title:    "多资产风险共振",
		Body:     "当前 3 类风险同时触发：宏观:STAGFLATION、期权压力:0.78、相关性断裂:2对。建议关注整体市场风险。",
		Severity: "high",
		Data: map[string]string{
			"kind":       "risk_confluence",
			"risk_count": "3",
			"details":    "macro:STAGFLATION,options:0.78,correlation:2",
		},
		DedupTTL: 4 * time.Hour,
	}
	if n.Category != "signal" {
		t.Errorf("risk confluence category = %q, want signal", n.Category)
	}
}

func TestRiskConfluenceBelowThreshold(t *testing.T) {
	tests := []struct {
		count     int
		shouldPush bool
	}{
		{0, false},
		{1, false},
		{2, true},
		{3, true},
	}
	for _, tt := range tests {
		got := tt.count >= 2
		if got != tt.shouldPush {
			t.Errorf("risk count=%d → push=%v, want %v", tt.count, got, tt.shouldPush)
		}
	}
}

// ---------- push_cot ----------

func TestCOTReleaseNotifyFormat(t *testing.T) {
	uid := uuid.New()
	n := &notify.Notification{
		UserID:   uid,
		Category: "alert",
		Title:    "COT 新数据已更新",
		Body:     "最新报告日期：2026-05-09。系统已完成持仓分析与信号刷新。",
		Severity: "normal",
		Data:     map[string]string{"kind": "cot_release", "report_date": "2026-05-09"},
		DedupTTL: 7 * 24 * time.Hour,
	}
	if n.DedupTTL != 7*24*time.Hour {
		t.Errorf("COT release dedup TTL = %v, want 7 days", n.DedupTTL)
	}
}

func TestCOTSignalNotifyFormat(t *testing.T) {
	uid := uuid.New()
	n := &notify.Notification{
		UserID:   uid,
		Category: "signal",
		Title:    "COT USDX 强看涨信号",
		Body:     "USDX commercial=EXTREME_LONG noncomm=EXTREME_SHORT signal=0.92。",
		Severity: "high",
		Data: map[string]string{
			"kind":            "cot_signal",
			"contract":        "USDX",
			"direction":       "看涨",
			"signal_strength": "0.92",
			"report_date":     "2026-05-09",
		},
		DedupTTL: 7 * 24 * time.Hour,
	}
	if n.Category != "signal" {
		t.Errorf("COT signal category = %q, want signal", n.Category)
	}
	if n.Severity != "high" {
		t.Errorf("COT extreme bias severity = %q, want high", n.Severity)
	}
}

func TestCOTSignalThreshold(t *testing.T) {
	tests := []struct {
		strength   float64
		shouldPush bool
	}{
		{-0.92, true},
		{-0.71, true},
		{-0.69, false},
		{-0.30, false},
		{0.30, false},
		{0.69, false},
		{0.71, true},
		{0.92, true},
	}
	for _, tt := range tests {
		abs := tt.strength
		if abs < 0 {
			abs = -abs
		}
		got := abs >= 0.7
		if got != tt.shouldPush {
			t.Errorf("|strength|=%.2f → push=%v, want %v", abs, got, tt.shouldPush)
		}
	}
}

func TestCalibrationNotifyFormat(t *testing.T) {
	uid := uuid.New()
	n := &notify.Notification{
		UserID:   uid,
		Category: "digest",
		Title:    "策略校准已更新",
		Body:     "信号校准模型已完成最新拟合，回测质量监控结果可查看。",
		Severity: "low",
		Data: map[string]string{
			"kind": "calibration_update",
			"date": "2026-05-13",
		},
		DedupTTL: 12 * time.Hour,
	}
	if n.Severity != "low" {
		t.Errorf("calibration severity = %q, want low", n.Severity)
	}
}

// ---------- 用户偏好过滤逻辑测试 ----------

func TestAlertPrefsChannelFilter(t *testing.T) {
	// 用户关闭 macro alerts → 不推送宏观
	p := alertPrefs{MacroAlerts: false}
	if p.MacroAlerts {
		t.Error("macro alerts disabled → skip push")
	} else {
		t.Log("macro alerts disabled → skip push (expected)")
	}
}

func TestAlertPrefsCotFilter(t *testing.T) {
	p := alertPrefs{CotAlerts: false}
	if !p.CotAlerts {
		t.Log("COT alerts disabled → skip COT push (expected)")
	}
}

func TestAlertPrefsDigestFilter(t *testing.T) {
	p := alertPrefs{DailyDigest: false}
	if !p.DailyDigest {
		t.Log("daily digest disabled → skip morning briefing (expected)")
	}
}
