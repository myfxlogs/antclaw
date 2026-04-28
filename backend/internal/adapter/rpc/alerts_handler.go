package rpc

import (
	"context"
	"strings"

	"connectrpc.com/connect"
	alertsv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
	"github.com/antclaw/antclaw/internal/auth"
	"github.com/antclaw/antclaw/internal/service/alerts"
)

// AlertsHandler implements antclawv1connect.AlertServiceHandler.
type AlertsHandler struct {
	svc *alerts.Service
}

// NewAlertsHandler creates a new AlertsHandler.
func NewAlertsHandler(svc *alerts.Service) *AlertsHandler {
	return &AlertsHandler{svc: svc}
}

func userIDFromCtx(ctx context.Context) string {
	if uid, ok := auth.UserIDFromContext(ctx); ok && strings.TrimSpace(uid) != "" {
		return uid
	}
	return "admin-001"
}

func (h *AlertsHandler) ListSubscriptions(ctx context.Context, req *connect.Request[alertsv1.ListSubscriptionsRequest]) (*connect.Response[alertsv1.ListSubscriptionsResponse], error) {
	resp, err := h.svc.ListSubscriptions(ctx, req.Msg.AlertTypeFilter, req.Msg.ActiveOnly)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *AlertsHandler) Subscribe(ctx context.Context, req *connect.Request[alertsv1.SubscribeRequest]) (*connect.Response[alertsv1.SubscribeResponse], error) {
	resp, err := h.svc.Subscribe(ctx, req.Msg.AlertType, req.Msg.Pair, req.Msg.Condition, req.Msg.Threshold, req.Msg.NotificationMethod)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *AlertsHandler) Unsubscribe(ctx context.Context, req *connect.Request[alertsv1.UnsubscribeRequest]) (*connect.Response[alertsv1.UnsubscribeResponse], error) {
	resp, err := h.svc.Unsubscribe(ctx, req.Msg.SubscriptionId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *AlertsHandler) RegisterWebhook(ctx context.Context, req *connect.Request[alertsv1.RegisterWebhookRequest]) (*connect.Response[alertsv1.RegisterWebhookResponse], error) {
	resp, err := h.svc.RegisterWebhook(ctx, req.Msg.Url, req.Msg.Secret, req.Msg.EventTypes)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *AlertsHandler) ListWebhooks(ctx context.Context, req *connect.Request[alertsv1.ListWebhooksRequest]) (*connect.Response[alertsv1.ListWebhooksResponse], error) {
	resp, err := h.svc.ListWebhooks(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *AlertsHandler) CreateAlert(ctx context.Context, req *connect.Request[alertsv1.CreateAlertRequest]) (*connect.Response[alertsv1.CreateAlertResponse], error) {
	rule, err := h.svc.CreateAlert(ctx, userIDFromCtx(ctx), req.Msg.AlertType, req.Msg.Symbol, req.Msg.ParamsJson, req.Msg.CooldownSeconds)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	return connect.NewResponse(&alertsv1.CreateAlertResponse{Alert: rule}), nil
}

func (h *AlertsHandler) ListAlerts(ctx context.Context, req *connect.Request[alertsv1.ListAlertsRequest]) (*connect.Response[alertsv1.ListAlertsResponse], error) {
	rows, err := h.svc.ListAlerts(ctx, userIDFromCtx(ctx), req.Msg.AlertType)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&alertsv1.ListAlertsResponse{Alerts: rows}), nil
}

func (h *AlertsHandler) UpdateAlert(ctx context.Context, req *connect.Request[alertsv1.UpdateAlertRequest]) (*connect.Response[alertsv1.UpdateAlertResponse], error) {
	rule, err := h.svc.UpdateAlert(ctx, userIDFromCtx(ctx), req.Msg.Id, req.Msg.ParamsJson, req.Msg.CooldownSeconds)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	return connect.NewResponse(&alertsv1.UpdateAlertResponse{Alert: rule}), nil
}

func (h *AlertsHandler) DeleteAlert(ctx context.Context, req *connect.Request[alertsv1.DeleteAlertRequest]) (*connect.Response[alertsv1.DeleteAlertResponse], error) {
	if err := h.svc.DeleteAlert(ctx, userIDFromCtx(ctx), req.Msg.Id); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&alertsv1.DeleteAlertResponse{Ok: true}), nil
}

func (h *AlertsHandler) ToggleAlert(ctx context.Context, req *connect.Request[alertsv1.ToggleAlertRequest]) (*connect.Response[alertsv1.ToggleAlertResponse], error) {
	rule, err := h.svc.ToggleAlert(ctx, userIDFromCtx(ctx), req.Msg.Id, req.Msg.Enabled)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&alertsv1.ToggleAlertResponse{Alert: rule}), nil
}

var _ antclawv1connect.AlertServiceHandler = (*AlertsHandler)(nil)
