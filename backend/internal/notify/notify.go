// Package notify 是 antclaw 的统一通知发送中枢。
//
// 推送链路（业务侧 → Service.Send → 去重/偏好/落库 → Redis Pub/Sub → SSE）：
//
//	业务侧 ──→ Service.Send
//	              │
//	              ├─ 1) Redis SETNX dedup_key（命中则丢弃）
//	              ├─ 2) 加载 user_notification_prefs（无则用默认值）
//	              ├─ 3) 持久化到 notifications 表（始终落库，便于回溯）
//	              └─ 4) 若 prefs 允许（类型/严重度/静默期），
//	                   发布到 Redis Pub/Sub 频道 user:{userID}:notifications
//	                   —— /sse/notifications 端点订阅这个频道实时推给浏览器。
//
// 设计取舍：
//   - 持久化与实时投递解耦：即使用户离线、SSE 断开，重新打开仍能从 ListUnread 看到
//   - 用 Pub/Sub 而非 Stream：每用户一个 channel，无消费组复杂度，断线重连不补发
//     （历史记录通过 ListUnread 拉取一次即可）
//   - 调用方完全不需要知道传输细节，只需要构造 Notification
package notify

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"

	notifyv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"

	"github.com/antclaw/antclaw/internal/adapter/storage/postgres/db"
)

// Notification 业务侧只填这几个字段。其它由 Service 自动补默认值。
type Notification struct {
	UserID   uuid.UUID
	Type     string // 投递通道：in_app / email / push（默认 in_app）
	Category string // 业务分类：alert / signal / system / digest（默认 system）
	Title    string
	Body     string
	Data     map[string]string
	Severity string // low | normal | high | critical（默认 normal）
	DedupKey string // 选填；非空则在 Redis 上做 SETNX TTL=DedupTTL 去重
	DedupTTL time.Duration
}

const userChannelPrefix = "user:" // 频道：user:{uuid}:notifications

// Service 是通知中枢；线程安全。
type Service struct {
	q      *db.Queries
	rdb    *redis.Client
	defTTL time.Duration
}

// NewService 构造 Service。queries 必填；rdb 选填（nil 则不做实时推送，仅落库）。
func NewService(q *db.Queries, rdb *redis.Client) *Service {
	return &Service{q: q, rdb: rdb, defTTL: 10 * time.Minute}
}

// UserChannel 返回某用户的 SSE 推送频道名（外部组件复用此约定）。
func UserChannel(userID string) string {
	return userChannelPrefix + userID + ":notifications"
}

// severityRank 用于 min_severity 比较。
var severityRank = map[string]int{"low": 0, "normal": 1, "high": 2, "critical": 3}

// Send 发送一条通知。返回错误仅当落库失败；dedup 命中、prefs 阻塞均视为成功（静默丢弃）。
func (s *Service) Send(ctx context.Context, n *Notification) error {
	if n == nil || n.UserID == uuid.Nil || n.Title == "" {
		return errors.New("notify: invalid notification")
	}
	n.Type = strings.ToLower(strings.TrimSpace(n.Type))
	if n.Type == "" {
		n.Type = "in_app"
	}
	n.Category = strings.ToLower(strings.TrimSpace(n.Category))
	if n.Category == "" {
		n.Category = "system"
	}
	n.Severity = strings.ToLower(strings.TrimSpace(n.Severity))
	if _, ok := severityRank[n.Severity]; !ok {
		n.Severity = "normal"
	}

	// 1) 去重
	if n.DedupKey != "" && s.rdb != nil {
		ttl := n.DedupTTL
		if ttl <= 0 {
			ttl = s.defTTL
		}
		ok, err := s.rdb.SetNX(ctx, "notify:dedup:"+n.DedupKey, "1", ttl).Result()
		if err == nil && !ok {
			return nil // 命中去重，安静丢弃
		}
	}

	// 2) 拉用户偏好（缺失走默认 → 全开）
	prefs, hasPrefs, _ := s.loadPrefs(ctx, n.UserID)

	// 3) 落库（始终）
	dataJSON, _ := json.Marshal(n.Data)
	priority := n.Severity
	dedup := nullable(n.DedupKey)
	if _, err := s.q.CreateNotification(ctx, db.CreateNotificationParams{
		UserID:   n.UserID,
		Type:     n.Type,
		Category: n.Category,
		Title:    n.Title,
		Body:     n.Body,
		Data:     dataJSON,
		Priority: &priority,
		Severity: n.Severity,
		DedupKey: dedup,
	}); err != nil {
		return fmt.Errorf("notify: persist: %w", err)
	}

	// 4) 实时推送（受偏好约束）
	if !s.shouldDeliverLive(prefs, hasPrefs, n) || s.rdb == nil {
		return nil
	}
	pb := &notifyv1.Notification{
		Category:  n.Category,
		Type:      n.Type,
		Title:     n.Title,
		Body:      n.Body,
		Severity:  n.Severity,
		Data:      n.Data,
		CreatedAt: time.Now().Unix(),
	}
	protoBytes, err := proto.Marshal(pb)
	if err != nil {
		return fmt.Errorf("notify: proto marshal: %w", err)
	}
	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(protoBytes)))
	base64.StdEncoding.Encode(encoded, protoBytes)
	_ = s.rdb.Publish(ctx, UserChannel(n.UserID.String()), string(encoded)).Err()
	return nil
}

// SendInApp 便捷方法：默认 in_app+system+normal。
func (s *Service) SendInApp(ctx context.Context, userID uuid.UUID, title, body string, data map[string]string) error {
	return s.Send(ctx, &Notification{UserID: userID, Title: title, Body: body, Data: data})
}

// MarkAsRead 标记某条已读（按用户绑定，避免越权）。
func (s *Service) MarkAsRead(ctx context.Context, userID, id uuid.UUID) error {
	return s.q.MarkNotificationRead(ctx, db.MarkNotificationReadParams{ID: id, UserID: userID})
}

// MarkAllRead 标记该用户所有未读为已读。
func (s *Service) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	return s.q.MarkAllRead(ctx, userID)
}

// GetUnread 返回未读列表（默认上限 100）。
func (s *Service) GetUnread(ctx context.Context, userID uuid.UUID, limit int32) ([]db.Notification, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	return s.q.GetUnreadNotifications(ctx, db.GetUnreadNotificationsParams{UserID: userID, Limit: limit})
}

// CountUnread 返回未读数。
func (s *Service) CountUnread(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.q.CountUnreadNotifications(ctx, userID)
}

// GetHistory 返回近 N 条历史。
func (s *Service) GetHistory(ctx context.Context, userID uuid.UUID, limit int32) ([]db.Notification, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	return s.q.GetNotificationHistory(ctx, db.GetNotificationHistoryParams{UserID: userID, Limit: limit})
}

func nullable(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// shouldDeliverLive 综合偏好/严重度/静默时段判断是否实时推送。
// 历史记录始终落库，因此即便此处返回 false，用户重新打开也能看到。
func (s *Service) shouldDeliverLive(p db.UserNotificationPref, hasPrefs bool, n *Notification) bool {
	if !hasPrefs {
		return true // 用户从未配置 → 全开
	}
	if !p.PushEnabled {
		return false
	}
	if len(p.EnabledTypes) > 0 && !contains(p.EnabledTypes, n.Category) {
		return false
	}
	if rank(n.Severity) < rank(p.MinSeverity) {
		return false
	}
	if inQuietHours(p, time.Now()) {
		return n.Severity == "critical" // 仅 critical 穿透静默
	}
	return true
}

func (s *Service) loadPrefs(ctx context.Context, userID uuid.UUID) (db.UserNotificationPref, bool, error) {
	p, err := s.q.GetUserNotificationPrefs(ctx, userID)
	if err != nil {
		return db.UserNotificationPref{}, false, err
	}
	return p, true, nil
}

func contains(arr []string, x string) bool {
	for _, v := range arr {
		if v == x {
			return true
		}
	}
	return false
}

func rank(s string) int {
	if v, ok := severityRank[strings.ToLower(s)]; ok {
		return v
	}
	return 1
}

// inQuietHours 当 quiet_start == quiet_end 视为不静默。跨午夜区间正确处理。
func inQuietHours(p db.UserNotificationPref, now time.Time) bool {
	if !p.QuietStart.Valid || !p.QuietEnd.Valid {
		return false
	}
	if p.QuietStart.Microseconds == p.QuietEnd.Microseconds {
		return false
	}
	loc, err := time.LoadLocation(p.Timezone)
	if err != nil {
		loc = time.UTC
	}
	local := now.In(loc)
	cur := pgTime(local)
	start := p.QuietStart.Microseconds
	end := p.QuietEnd.Microseconds
	if start < end {
		return cur >= start && cur < end
	}
	// 跨午夜（如 22:00 → 07:00）
	return cur >= start || cur < end
}

// pgTime 把 time.Time 当天的 wallclock 转为 pgtype.Time 的 microseconds（与 sqlc 保持一致）。
func pgTime(t time.Time) int64 {
	h, m, s := t.Clock()
	return int64(h)*3600_000_000 + int64(m)*60_000_000 + int64(s)*1_000_000 + int64(t.Nanosecond()/1000)
}

// 编译期保证 pgtype.Time 字段被使用（避免未引用 import）。
var _ = pgtype.Time{}
