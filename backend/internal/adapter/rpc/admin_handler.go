package rpc

import (
	"context"
	"errors"
	"strings"

	"connectrpc.com/connect"
	adminv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
	"github.com/antclaw/antclaw/internal/service/admin"
)

// AdminHandler implements antclawv1connect.AdminServiceHandler.
type AdminHandler struct {
	svc *admin.Service
}

// NewAdminHandler creates a new AdminHandler.
func NewAdminHandler(svc *admin.Service) *AdminHandler {
	return &AdminHandler{svc: svc}
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

var _ antclawv1connect.AdminServiceHandler = (*AdminHandler)(nil)
