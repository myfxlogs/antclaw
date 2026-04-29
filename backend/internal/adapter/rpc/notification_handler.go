package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	v1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/internal/adapter/storage/postgres/db"
	"github.com/antclaw/antclaw/internal/auth"
	"github.com/antclaw/antclaw/internal/notify"
)

// NotificationHandler 实现 NotificationService。所有方法均要求登录态。
type NotificationHandler struct {
	svc *notify.Service
	q   *db.Queries
}

// NewNotificationHandler 注入 notify.Service 与 db.Queries（用于读偏好）。
func NewNotificationHandler(svc *notify.Service, q *db.Queries) *NotificationHandler {
	return &NotificationHandler{svc: svc, q: q}
}

func (h *NotificationHandler) currentUser(ctx context.Context) (uuid.UUID, error) {
	uidStr, ok := auth.UserIDFromContext(ctx)
	if !ok || uidStr == "" {
		return uuid.Nil, connect.NewError(connect.CodeUnauthenticated, errors.New("login required"))
	}
	uid, err := uuid.Parse(uidStr)
	if err != nil {
		return uuid.Nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid user id"))
	}
	return uid, nil
}

func (h *NotificationHandler) ListUnread(ctx context.Context, req *connect.Request[v1.ListUnreadRequest]) (*connect.Response[v1.ListUnreadResponse], error) {
	uid, err := h.currentUser(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := h.svc.GetUnread(ctx, uid, req.Msg.GetLimit())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&v1.ListUnreadResponse{Items: toProtoList(rows)}), nil
}

func (h *NotificationHandler) ListHistory(ctx context.Context, req *connect.Request[v1.ListHistoryRequest]) (*connect.Response[v1.ListHistoryResponse], error) {
	uid, err := h.currentUser(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := h.svc.GetHistory(ctx, uid, req.Msg.GetLimit())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&v1.ListHistoryResponse{Items: toProtoList(rows)}), nil
}

func (h *NotificationHandler) UnreadCount(ctx context.Context, _ *connect.Request[v1.UnreadCountRequest]) (*connect.Response[v1.UnreadCountResponse], error) {
	uid, err := h.currentUser(ctx)
	if err != nil {
		return nil, err
	}
	n, err := h.svc.CountUnread(ctx, uid)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&v1.UnreadCountResponse{Count: n}), nil
}

func (h *NotificationHandler) MarkRead(ctx context.Context, req *connect.Request[v1.MarkReadRequest]) (*connect.Response[v1.MarkReadResponse], error) {
	uid, err := h.currentUser(ctx)
	if err != nil {
		return nil, err
	}
	id, err := uuid.Parse(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid id"))
	}
	if err := h.svc.MarkAsRead(ctx, uid, id); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&v1.MarkReadResponse{}), nil
}

func (h *NotificationHandler) MarkAllRead(ctx context.Context, _ *connect.Request[v1.MarkAllReadRequest]) (*connect.Response[v1.MarkAllReadResponse], error) {
	uid, err := h.currentUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.svc.MarkAllRead(ctx, uid); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&v1.MarkAllReadResponse{}), nil
}

func (h *NotificationHandler) GetPrefs(ctx context.Context, _ *connect.Request[v1.GetPrefsRequest]) (*connect.Response[v1.GetPrefsResponse], error) {
	uid, err := h.currentUser(ctx)
	if err != nil {
		return nil, err
	}
	p, err := h.q.GetUserNotificationPrefs(ctx, uid)
	if err != nil {
		// 未配置 → 返回内置默认。
		return connect.NewResponse(&v1.GetPrefsResponse{Prefs: &v1.NotificationPrefs{
			EnabledTypes: []string{"alert", "signal", "system", "digest"},
			MinSeverity:  "low",
			QuietStart:   "00:00",
			QuietEnd:     "00:00",
			Timezone:     "UTC",
			PushEnabled:  true,
			EmailEnabled: false,
		}}), nil
	}
	return connect.NewResponse(&v1.GetPrefsResponse{Prefs: prefsToProto(p)}), nil
}

func (h *NotificationHandler) UpdatePrefs(ctx context.Context, req *connect.Request[v1.UpdatePrefsRequest]) (*connect.Response[v1.UpdatePrefsResponse], error) {
	uid, err := h.currentUser(ctx)
	if err != nil {
		return nil, err
	}
	in := req.Msg.GetPrefs()
	if in == nil {
		in = &v1.NotificationPrefs{}
	}
	qStart, err := parseHHMM(in.GetQuietStart())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("quiet_start: %w", err))
	}
	qEnd, err := parseHHMM(in.GetQuietEnd())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("quiet_end: %w", err))
	}
	tz := strings.TrimSpace(in.GetTimezone())
	if tz == "" {
		tz = "UTC"
	}
	if _, err := time.LoadLocation(tz); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("timezone: %w", err))
	}
	sev := strings.ToLower(strings.TrimSpace(in.GetMinSeverity()))
	if sev == "" {
		sev = "low"
	}
	types := normalizeTypes(in.GetEnabledTypes())
	out, err := h.q.UpsertUserNotificationPrefs(ctx, db.UpsertUserNotificationPrefsParams{
		UserID:       uid,
		EnabledTypes: types,
		MinSeverity:  sev,
		QuietStart:   qStart,
		QuietEnd:     qEnd,
		Timezone:     tz,
		PushEnabled:  in.GetPushEnabled(),
		EmailEnabled: in.GetEmailEnabled(),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&v1.UpdatePrefsResponse{Prefs: prefsToProto(out)}), nil
}

// ---------- 转换 ----------

func toProtoList(rows []db.Notification) []*v1.Notification {
	out := make([]*v1.Notification, 0, len(rows))
	for _, r := range rows {
		out = append(out, dbNotifToProto(r))
	}
	return out
}

func dbNotifToProto(r db.Notification) *v1.Notification {
	data := map[string]string{}
	if len(r.Data) > 0 {
		var m map[string]any
		if json.Unmarshal(r.Data, &m) == nil {
			for k, v := range m {
				if s, ok := v.(string); ok {
					data[k] = s
				} else {
					b, _ := json.Marshal(v)
					data[k] = string(b)
				}
			}
		}
	}
	created := int64(0)
	if r.CreatedAt.Valid {
		created = r.CreatedAt.Time.Unix()
	}
	read := int64(0)
	if r.ReadAt.Valid {
		read = r.ReadAt.Time.Unix()
	}
	return &v1.Notification{
		Id:        r.ID.String(),
		UserId:    r.UserID.String(),
		Type:      r.Type,
		Category:  r.Category,
		Severity:  r.Severity,
		Title:     r.Title,
		Body:      r.Body,
		Data:      data,
		IsRead:    r.IsRead != nil && *r.IsRead,
		CreatedAt: created,
		ReadAt:    read,
	}
}

func prefsToProto(p db.UserNotificationPref) *v1.NotificationPrefs {
	return &v1.NotificationPrefs{
		EnabledTypes: append([]string(nil), p.EnabledTypes...),
		MinSeverity:  p.MinSeverity,
		QuietStart:   formatHHMM(p.QuietStart),
		QuietEnd:     formatHHMM(p.QuietEnd),
		Timezone:     p.Timezone,
		PushEnabled:  p.PushEnabled,
		EmailEnabled: p.EmailEnabled,
	}
}

func parseHHMM(s string) (pgtype.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		s = "00:00"
	}
	t, err := time.Parse("15:04", s)
	if err != nil {
		return pgtype.Time{}, fmt.Errorf("expect HH:MM, got %q", s)
	}
	micros := int64(t.Hour())*3600_000_000 + int64(t.Minute())*60_000_000
	return pgtype.Time{Microseconds: micros, Valid: true}, nil
}

func formatHHMM(t pgtype.Time) string {
	if !t.Valid {
		return "00:00"
	}
	total := t.Microseconds / 1_000_000 // seconds
	h := total / 3600
	m := (total % 3600) / 60
	return fmt.Sprintf("%02d:%02d", h, m)
}

func normalizeTypes(in []string) []string {
	allowed := map[string]bool{"alert": true, "signal": true, "system": true, "digest": true}
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, x := range in {
		x = strings.ToLower(strings.TrimSpace(x))
		if !allowed[x] || seen[x] {
			continue
		}
		seen[x] = true
		out = append(out, x)
	}
	if len(out) == 0 {
		out = []string{"alert", "signal", "system", "digest"}
	}
	return out
}
