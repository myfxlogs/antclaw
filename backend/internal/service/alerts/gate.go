// M-E: AlertGate 按 tier / 订阅 / 静默时段 / 冷却 过滤告警，并将决策写入 alert_log。
package alerts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Decision 闸门决策。Reason 取值：
//
//	ok / tier_blocked / quiet_hours / cooldown / unsubscribed_pair / unknown_user
type Decision struct {
	Send   bool
	Reason string
}

// Gate 告警闸门。
type Gate struct {
	pool *pgxpool.Pool
}

// NewGate 构造闸门。
func NewGate(pool *pgxpool.Pool) *Gate { return &Gate{pool: pool} }

// Decide 5 级过滤，并写 alert_log。pool=nil 时返回 ok 不写库（用于测试）。
//
// severityRank: low=0, medium=1, high=2, critical=3。
// free 用户禁止 critical（rank>=3）告警。
func (g *Gate) Decide(ctx context.Context, userID, alertType, severity string, pairs []string) Decision {
	d := g.decide(ctx, userID, alertType, severity, pairs)
	g.log(ctx, userID, alertType, severity, pairs, d)
	return d
}

func (g *Gate) decide(ctx context.Context, userID, alertType, severity string, pairs []string) Decision {
	if g.pool == nil {
		return Decision{Send: true, Reason: "ok"}
	}
	if userID == "" {
		return Decision{Send: false, Reason: "unknown_user"}
	}
	// 1) tier 检查
	tier, _ := g.getTier(ctx, userID)
	if tier == "" {
		tier = "free"
	}
	rank := severityRank(severity)
	if tier == "free" && rank >= 3 {
		return Decision{Send: false, Reason: "tier_blocked"}
	}
	// 2) 订阅检查
	prefs, _ := g.getPreferences(ctx, userID)
	if prefs != nil && prefs.HighImpactOnly && rank < 2 {
		return Decision{Send: false, Reason: "tier_blocked"}
	}
	if prefs != nil && len(prefs.Pairs) > 0 && len(pairs) > 0 {
		match := false
		set := map[string]struct{}{}
		for _, p := range prefs.Pairs {
			set[strings.ToUpper(p)] = struct{}{}
		}
		for _, p := range pairs {
			if _, ok := set[strings.ToUpper(p)]; ok {
				match = true
				break
			}
		}
		if !match {
			return Decision{Send: false, Reason: "unsubscribed_pair"}
		}
	}
	// 3) 静默时段
	if prefs != nil && inQuietHours(time.Now(), prefs.QuietHoursStart, prefs.QuietHoursEnd) {
		return Decision{Send: false, Reason: "quiet_hours"}
	}
	// 4) 冷却：同 user+alert_type 1 小时内最多 3 次。
	count := g.recentSendCount(ctx, userID, alertType, time.Hour)
	if count >= 3 {
		return Decision{Send: false, Reason: "cooldown"}
	}
	return Decision{Send: true, Reason: "ok"}
}

// log 写 alert_log；payload = JSON({pairs}); pool=nil 时跳过。
func (g *Gate) log(ctx context.Context, userID, alertType, severity string, pairs []string, d Decision) {
	if g.pool == nil {
		return
	}
	payload, _ := json.Marshal(map[string]any{"pairs": pairs})
	_, _ = g.pool.Exec(ctx, `
		INSERT INTO alert_log(user_id, alert_type, severity, payload, sent, reason)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		userID, alertType, severity, string(payload), d.Send, d.Reason)
}

func severityRank(s string) int {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "low":
		return 0
	case "medium", "":
		return 1
	case "high":
		return 2
	case "critical":
		return 3
	}
	return 1
}

// inQuietHours 当 start==end 时视为未启用；start<end 时为同日范围；start>end 跨夜。
func inQuietHours(now time.Time, start, end int32) bool {
	if start == end {
		return false
	}
	h := now.Hour()
	s, e := int(start), int(end)
	if s < e {
		return h >= s && h < e
	}
	return h >= s || h < e
}

// Preferences 用户偏好。
type Preferences struct {
	UserID            string
	Pairs             []string
	HighImpactOnly    bool
	QuietHoursStart   int32
	QuietHoursEnd     int32
	Timezone          string
}

func (g *Gate) getPreferences(ctx context.Context, userID string) (*Preferences, error) {
	if g.pool == nil {
		return nil, errors.New("pool nil")
	}
	row := g.pool.QueryRow(ctx, `
		SELECT user_id, COALESCE(pairs, '{}'), COALESCE(high_impact_only, false),
		       COALESCE(quiet_hours_start, 0), COALESCE(quiet_hours_end, 0), COALESCE(timezone, 'UTC')
		  FROM user_preferences WHERE user_id = $1`, userID)
	p := &Preferences{}
	if err := row.Scan(&p.UserID, &p.Pairs, &p.HighImpactOnly, &p.QuietHoursStart, &p.QuietHoursEnd, &p.Timezone); err != nil {
		return nil, err
	}
	return p, nil
}

func (g *Gate) UpsertPreferences(ctx context.Context, p *Preferences) error {
	if g.pool == nil {
		return errors.New("pool nil")
	}
	_, err := g.pool.Exec(ctx, `
		INSERT INTO user_preferences(user_id, pairs, high_impact_only, quiet_hours_start, quiet_hours_end, timezone)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (user_id) DO UPDATE SET
			pairs = EXCLUDED.pairs,
			high_impact_only = EXCLUDED.high_impact_only,
			quiet_hours_start = EXCLUDED.quiet_hours_start,
			quiet_hours_end = EXCLUDED.quiet_hours_end,
			timezone = EXCLUDED.timezone`,
		p.UserID, p.Pairs, p.HighImpactOnly, p.QuietHoursStart, p.QuietHoursEnd, p.Timezone)
	return err
}

func (g *Gate) GetPreferences(ctx context.Context, userID string) (*Preferences, error) {
	p, err := g.getPreferences(ctx, userID)
	if err != nil {
		// 不存在视为默认偏好
		return &Preferences{UserID: userID, Timezone: "UTC"}, nil
	}
	return p, nil
}

// SetTier 写 user_quotas 的 tier。
func (g *Gate) SetTier(ctx context.Context, userID, tier string, aiMax int) error {
	if g.pool == nil {
		return errors.New("pool nil")
	}
	if tier == "" {
		tier = "free"
	}
	if aiMax <= 0 {
		switch tier {
		case "premium":
			aiMax = 200
		default:
			aiMax = 20
		}
	}
	_, err := g.pool.Exec(ctx, `
		INSERT INTO user_quotas(user_id, tier, ai_max_per_day)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) DO UPDATE SET
			tier = EXCLUDED.tier, ai_max_per_day = EXCLUDED.ai_max_per_day`,
		userID, tier, aiMax)
	return err
}

func (g *Gate) getTier(ctx context.Context, userID string) (string, error) {
	if g.pool == nil {
		return "free", nil
	}
	var tier string
	err := g.pool.QueryRow(ctx, `SELECT tier FROM user_quotas WHERE user_id = $1`, userID).Scan(&tier)
	if err != nil {
		return "free", err
	}
	return tier, nil
}

// recentSendCount 最近 window 时间内 sent=true 的告警计数（用于冷却判定）。
func (g *Gate) recentSendCount(ctx context.Context, userID, alertType string, window time.Duration) int {
	if g.pool == nil {
		return 0
	}
	var n int
	_ = g.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM alert_log
		 WHERE user_id=$1 AND alert_type=$2 AND sent=true
		   AND created_at > NOW() - $3::interval`,
		userID, alertType, fmt.Sprintf("%d seconds", int(window.Seconds()))).Scan(&n)
	return n
}

// GetAlertHistory 最近 limit 条 alert_log。
type LogItem struct {
	ID        int64
	UserID    string
	AlertType string
	Severity  string
	Sent      bool
	Reason    string
	CreatedAt time.Time
}

func (g *Gate) GetAlertHistory(ctx context.Context, userID string, limit int) ([]LogItem, error) {
	if g.pool == nil {
		return nil, nil
	}
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	rows, err := g.pool.Query(ctx, `
		SELECT id, COALESCE(user_id,''), COALESCE(alert_type,''), COALESCE(severity,''),
		       COALESCE(sent,false), COALESCE(reason,''), created_at
		  FROM alert_log WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2`,
		userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []LogItem{}
	for rows.Next() {
		var it LogItem
		if err := rows.Scan(&it.ID, &it.UserID, &it.AlertType, &it.Severity, &it.Sent, &it.Reason, &it.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, nil
}
