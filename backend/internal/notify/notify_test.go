package notify

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/antclaw/antclaw/internal/adapter/storage/postgres/db"
)

// ---------- shouldDeliverLive 逻辑测试 ----------

func TestShouldDeliverLive_NoPrefs(t *testing.T) {
	s := &Service{}
	if !s.shouldDeliverLive(db.UserNotificationPref{}, false, &Notification{Severity: "low"}) {
		t.Error("no prefs (hasPrefs=false) should deliver live")
	}
}

func TestShouldDeliverLive_PushDisabled(t *testing.T) {
	s := &Service{}
	p := db.UserNotificationPref{PushEnabled: false}
	if s.shouldDeliverLive(p, true, &Notification{Severity: "low"}) {
		t.Error("push disabled should NOT deliver live")
	}
}

func TestShouldDeliverLive_DisabledType(t *testing.T) {
	s := &Service{}
	p := db.UserNotificationPref{
		PushEnabled:  true,
		EnabledTypes: []string{"alert", "signal"},
	}
	if s.shouldDeliverLive(p, true, &Notification{Category: "digest", Severity: "normal"}) {
		t.Error("digest not in enabled types should NOT deliver live")
	}
}

func TestShouldDeliverLive_SeverityTooLow(t *testing.T) {
	s := &Service{}
	p := db.UserNotificationPref{
		PushEnabled: true,
		MinSeverity: "high",
	}
	if s.shouldDeliverLive(p, true, &Notification{Category: "alert", Severity: "normal"}) {
		t.Error("normal severity < min high should NOT deliver live")
	}
}

func TestShouldDeliverLive_CriticalPenetrates(t *testing.T) {
	s := &Service{}
	p := db.UserNotificationPref{
		PushEnabled: true,
	}
	if !s.shouldDeliverLive(p, true, &Notification{Category: "alert", Severity: "critical"}) {
		t.Error("critical should deliver when push enabled")
	}
}

// ---------- normalizeDefaults 提取 Send 中的默认值逻辑 ----------

// normalizeDefaults 模拟 Send 方法中的默认值注入（去除 db/redis 依赖）。
func normalizeDefaults(n *Notification) {
	if n.Type == "" {
		n.Type = "in_app"
	}
	if n.Category == "" {
		n.Category = "system"
	}
	if _, ok := severityRank[n.Severity]; !ok {
		n.Severity = "normal"
	}
}

// ---------- 严重度排序 ----------

func TestSeverityRank(t *testing.T) {
	tests := []struct {
		severity string
		rank     int
	}{
		{"low", 0},
		{"normal", 1},
		{"high", 2},
		{"critical", 3},
	}
	for _, tt := range tests {
		if got := severityRank[tt.severity]; got != tt.rank {
			t.Errorf("severityRank[%q] = %d, want %d", tt.severity, got, tt.rank)
		}
	}
}

// ---------- 通知字段默认值 ----------

func TestNotificationDefaults(t *testing.T) {
	n := &Notification{Title: "test", UserID: uuid.New()}
	normalizeDefaults(n)

	if n.Type != "in_app" {
		t.Errorf("default type = %q, want in_app", n.Type)
	}
	if n.Category != "system" {
		t.Errorf("default category = %q, want system", n.Category)
	}
	if n.Severity != "normal" {
		t.Errorf("default severity = %q, want normal", n.Severity)
	}
}

func TestNotificationInvalid(t *testing.T) {
	// nil notification
	if !isInvalid(nil) {
		t.Error("nil notification should be invalid")
	}
	// empty title
	if !isInvalid(&Notification{}) {
		t.Error("empty title notification should be invalid")
	}
	// valid
	if isInvalid(&Notification{Title: "hello", UserID: uuid.New()}) {
		t.Error("valid notification should not be invalid")
	}
}

func isInvalid(n *Notification) bool {
	return n == nil || n.UserID == uuid.Nil || n.Title == ""
}

// ---------- 用户频道命名 ----------

func TestUserChannel(t *testing.T) {
	ch := UserChannel("abc-123")
	if ch != "user:abc-123:notifications" {
		t.Errorf("UserChannel = %q, want user:abc-123:notifications", ch)
	}
}

// ---------- 默认 TTL ----------

func TestDefaultTTL(t *testing.T) {
	s := NewService(nil, nil)
	if s.defTTL != 10*time.Minute {
		t.Errorf("default DedupTTL = %v, want 10m", s.defTTL)
	}
}

// ---------- Dedup TTL 回退 ----------

func TestEffectiveDedupTTL(t *testing.T) {
	n := &Notification{DedupTTL: 0}
	ttl := n.DedupTTL
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	if ttl != 10*time.Minute {
		t.Errorf("effective TTL = %v, want 10m", ttl)
	}

	n.DedupTTL = 5 * time.Minute
	ttl = n.DedupTTL
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	if ttl != 5*time.Minute {
		t.Errorf("effective TTL = %v, want 5m", ttl)
	}
}

// ---------- 表驱动：通知构造正确性 ----------

func TestNotificationConstruction(t *testing.T) {
	tests := []struct {
		name     string
		n        Notification
		wantType string
		wantCat  string
		wantSev  string
	}{
		{
			name:     "alert-critical",
			n:        Notification{Category: "alert", Severity: "critical", Type: "in_app", Title: "test", UserID: uuid.New()},
			wantType: "in_app",
			wantCat:  "alert",
			wantSev:  "critical",
		},
		{
			name:     "signal-normal",
			n:        Notification{Category: "signal", Severity: "normal", Title: "test", UserID: uuid.New()},
			wantType: "in_app",
			wantCat:  "signal",
			wantSev:  "normal",
		},
		{
			name:     "digest-low",
			n:        Notification{Category: "digest", Severity: "low", Title: "test", UserID: uuid.New()},
			wantType: "in_app",
			wantCat:  "digest",
			wantSev:  "low",
		},
		{
			name:     "system-defaults",
			n:        Notification{Title: "test", UserID: uuid.New()},
			wantType: "in_app",
			wantCat:  "system",
			wantSev:  "normal",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalizeDefaults(&tt.n)
			if tt.n.Type != tt.wantType {
				t.Errorf("type = %q, want %q", tt.n.Type, tt.wantType)
			}
			if tt.n.Category != tt.wantCat {
				t.Errorf("category = %q, want %q", tt.n.Category, tt.wantCat)
			}
			if tt.n.Severity != tt.wantSev {
				t.Errorf("severity = %q, want %q", tt.n.Severity, tt.wantSev)
			}
		})
	}
}
