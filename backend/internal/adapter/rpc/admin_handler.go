package rpc

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	adminv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
	"github.com/antclaw/antclaw/internal/auth"
	"github.com/antclaw/antclaw/internal/notify"
	"github.com/antclaw/antclaw/internal/service/admin"
	"github.com/antclaw/antclaw/internal/service/presence"
)

// AdminHandler implements antclawv1connect.AdminServiceHandler.
type AdminHandler struct {
	svc      *admin.Service
	notify   *notify.Service
	presence *presence.Tracker
	pg       *pgxpool.Pool
}

// NewAdminHandler creates a new AdminHandler.
func NewAdminHandler(svc *admin.Service, ns *notify.Service, pt *presence.Tracker, pg *pgxpool.Pool) *AdminHandler {
	return &AdminHandler{svc: svc, notify: ns, presence: pt, pg: pg}
}

func (h *AdminHandler) ListUsers(ctx context.Context, req *connect.Request[adminv1.ListUsersRequest]) (*connect.Response[adminv1.ListUsersResponse], error) {
	resp, err := h.svc.ListUsers(ctx, req.Msg.Cursor, req.Msg.PageSize, req.Msg.EmailFilter, req.Msg.RoleFilter, req.Msg.BannedOnly)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *AdminHandler) SetRole(ctx context.Context, req *connect.Request[adminv1.SetRoleRequest]) (*connect.Response[adminv1.SetRoleResponse], error) {
	resp, err := h.svc.SetRole(ctx, req.Msg.UserId, req.Msg.Roles)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *AdminHandler) Ban(ctx context.Context, req *connect.Request[adminv1.BanRequest]) (*connect.Response[adminv1.BanResponse], error) {
	resp, err := h.svc.Ban(ctx, req.Msg.UserId, req.Msg.Reason, req.Msg.ExpiresAt)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *AdminHandler) Unban(ctx context.Context, req *connect.Request[adminv1.UnbanRequest]) (*connect.Response[adminv1.UnbanResponse], error) {
	resp, err := h.svc.Unban(ctx, req.Msg.UserId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *AdminHandler) RunJob(ctx context.Context, req *connect.Request[adminv1.RunJobRequest]) (*connect.Response[adminv1.RunJobResponse], error) {
	resp, err := h.svc.RunJob(ctx, req.Msg.JobName, req.Msg.Params)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *AdminHandler) ListJobs(ctx context.Context, req *connect.Request[adminv1.ListJobsRequest]) (*connect.Response[adminv1.ListJobsResponse], error) {
	resp, err := h.svc.ListJobs(ctx, req.Msg.StatusFilter)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *AdminHandler) SetJobEnabled(ctx context.Context, req *connect.Request[adminv1.SetJobEnabledRequest]) (*connect.Response[adminv1.SetJobEnabledResponse], error) {
	if req.Msg.GetJobId() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("job_id required"))
	}
	resp, err := h.svc.SetJobEnabled(ctx, req.Msg.GetJobId(), req.Msg.GetEnabled())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *AdminHandler) ListAuditLogs(ctx context.Context, req *connect.Request[adminv1.ListAuditLogsRequest]) (*connect.Response[adminv1.ListAuditLogsResponse], error) {
	resp, err := h.svc.ListAuditLogs(ctx, req.Msg.Cursor, req.Msg.PageSize, req.Msg.UserIdFilter, req.Msg.ActionFilter)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *AdminHandler) ListWebhookDeliveries(ctx context.Context, req *connect.Request[adminv1.ListWebhookDeliveriesRequest]) (*connect.Response[adminv1.ListWebhookDeliveriesResponse], error) {
	resp, err := h.svc.ListWebhookDeliveries(ctx, req.Msg.Cursor, req.Msg.PageSize, req.Msg.WebhookIdFilter)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *AdminHandler) ForceLogout(ctx context.Context, req *connect.Request[adminv1.ForceLogoutRequest]) (*connect.Response[adminv1.ForceLogoutResponse], error) {
	resp, err := h.svc.ForceLogout(ctx, req.Msg.UserId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

// SetUserCodeID 管理员为用户分配/修改数字 ID（5-10 位，避开 4/7，不以 0 开头）。
// code_id 留空触发系统自动生成；非空时要求满足格式且全局唯一。
func (h *AdminHandler) SetUserCodeID(ctx context.Context, req *connect.Request[adminv1.SetUserCodeIDRequest]) (*connect.Response[adminv1.SetUserCodeIDResponse], error) {
	if req.Msg.GetUserId() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("user_id required"))
	}
	resp, err := h.svc.SetUserCodeID(ctx, req.Msg.GetUserId(), req.Msg.GetCodeId())
	if err != nil {
		switch {
		case errors.Is(err, admin.ErrCodeIDTaken):
			return nil, connect.NewError(connect.CodeAlreadyExists, err)
		case errors.Is(err, auth.ErrInvalidCodeID):
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		case strings.Contains(err.Error(), "invalid user_id"):
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *AdminHandler) AdminResetUserPassword(ctx context.Context, req *connect.Request[adminv1.AdminResetUserPasswordRequest]) (*connect.Response[adminv1.AdminResetUserPasswordResponse], error) {
	if req.Msg.GetUserId() == "" || req.Msg.GetNewPassword() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("user_id and new_password required"))
	}
	resp, err := h.svc.AdminResetUserPassword(ctx, req.Msg.GetUserId(), req.Msg.GetNewPassword())
	if err != nil {
		if errors.Is(err, admin.ErrPasswordPolicy) {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		if strings.Contains(err.Error(), "invalid user_id") {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *AdminHandler) resolveTargets(req *connect.Request[adminv1.SendPushRequest]) []uuid.UUID {
	if ids := req.Msg.GetTargetUserIds(); len(ids) > 0 {
		return parseUUIDs(ids)
	}
	// 空 = 全部在线
	online := h.presence.List()
	out := make([]uuid.UUID, 0, len(online))
	for _, u := range online {
		if uid, err := uuid.Parse(u.UserID); err == nil {
			out = append(out, uid)
		}
	}
	return out
}

func parseUUIDs(ids []string) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if uid, err := uuid.Parse(id); err == nil {
			out = append(out, uid)
		}
	}
	return out
}

func (h *AdminHandler) SendPush(ctx context.Context, req *connect.Request[adminv1.SendPushRequest]) (*connect.Response[adminv1.SendPushResponse], error) {
	adminID, _ := auth.UserIDFromContext(ctx)
	title := strings.TrimSpace(req.Msg.GetTitle())
	if title == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("title required"))
	}
	sev := defaultStr(strings.ToLower(strings.TrimSpace(req.Msg.GetSeverity())), "normal")
	cat := defaultStr(strings.ToLower(strings.TrimSpace(req.Msg.GetCategory())), "system")

	targets := h.resolveTargets(req)
	if len(targets) == 0 {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("没有可推送的目标用户"))
	}

	sent := 0
	dedup := fmt.Sprintf("manual:%s:%d", adminID, time.Now().Unix())
	for _, uid := range targets {
		if h.notify.Send(ctx, &notify.Notification{
			UserID: uid, Category: cat, Title: title, Body: req.Msg.GetBody(),
			Severity: sev, DedupKey: dedup,
			Data: map[string]string{"kind": "manual_push", "admin_id": adminID},
		}) == nil {
			sent++
		}
	}

	tids := make([]string, len(targets))
	for i, u := range targets {
		tids[i] = u.String()
	}
	var logID string
	adminUID, _ := uuid.Parse(adminID)
	_ = h.pg.QueryRow(ctx,
		`INSERT INTO manual_push_log (title,body,severity,category,target_count,sent_count,admin_user_id,target_user_ids)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`,
		title, req.Msg.GetBody(), sev, cat, len(targets), sent, adminUID, tids,
	).Scan(&logID)

	return connect.NewResponse(&adminv1.SendPushResponse{
		SentCount: int32(sent), OnlineCount: int32(len(targets)), PushLogId: logID,
	}), nil
}

func (h *AdminHandler) GetPushHistory(ctx context.Context, req *connect.Request[adminv1.GetPushHistoryRequest]) (*connect.Response[adminv1.GetPushHistoryResponse], error) {
	rows, err := h.pg.Query(ctx,
		`SELECT id,title,body,severity,target_count,sent_count,admin_user_id,created_at
		   FROM manual_push_log WHERE ($1='' OR created_at<$1::timestamptz)
		   ORDER BY created_at DESC LIMIT $2`,
		req.Msg.GetCursor(), clampPage32(req.Msg.GetPageSize())+1,
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	defer rows.Close()

	var entries []*adminv1.PushHistoryEntry
	var next string
	for rows.Next() {
		var e adminv1.PushHistoryEntry
		var t time.Time
		if rows.Scan(&e.Id, &e.Title, &e.Body, &e.Severity, &e.TargetCount, &e.SentCount, &e.AdminUserId, &t) != nil {
			continue
		}
		e.CreatedAt = t.Unix()
		entries = append(entries, &e)
	}
	return connect.NewResponse(&adminv1.GetPushHistoryResponse{Entries: entries, NextCursor: next}), nil
}

func clampPage32(n int32) int32 {
	if n <= 0 || n > 100 {
		return 50
	}
	return n
}

func defaultStr(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

var _ antclawv1connect.AdminServiceHandler = (*AdminHandler)(nil)
