package rpc

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/antclaw/antclaw/internal/auth"
	"github.com/antclaw/antclaw/internal/service/user"
	userv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
)

// UserHandler implements antclawv1connect.UserServiceHandler.
type UserHandler struct {
	svc *user.Service
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(svc *user.Service) *UserHandler {
	return &UserHandler{svc: svc}
}

// userID extracts authenticated userID from context.
func userID(ctx context.Context) (string, error) {
	id, ok := auth.UserIDFromContext(ctx)
	if !ok || id == "" {
		return "", fmt.Errorf("unauthenticated")
	}
	return id, nil
}

func (h *UserHandler) GetMe(ctx context.Context, req *connect.Request[userv1.GetMeRequest]) (*connect.Response[userv1.GetMeResponse], error) {
	id, err := userID(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}
	resp, err := h.svc.GetMe(ctx, id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *UserHandler) UpdateSettings(ctx context.Context, req *connect.Request[userv1.UpdateSettingsRequest]) (*connect.Response[userv1.UpdateSettingsResponse], error) {
	id, err := userID(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}
	resp, err := h.svc.UpdateSettings(ctx, id, req.Msg.DisplayName, req.Msg.Locale, req.Msg.Timezone)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *UserHandler) GetMembership(ctx context.Context, req *connect.Request[userv1.GetMembershipRequest]) (*connect.Response[userv1.GetMembershipResponse], error) {
	id, err := userID(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}
	resp, err := h.svc.GetMembership(ctx, id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *UserHandler) StartOnboarding(ctx context.Context, req *connect.Request[userv1.StartOnboardingRequest]) (*connect.Response[userv1.StartOnboardingResponse], error) {
	resp, err := h.svc.StartOnboarding(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *UserHandler) GetHistory(ctx context.Context, req *connect.Request[userv1.GetHistoryRequest]) (*connect.Response[userv1.GetHistoryResponse], error) {
	id, err := userID(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}
	resp, err := h.svc.GetHistory(ctx, id, req.Msg.Cursor, req.Msg.PageSize)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *UserHandler) ClearHistory(ctx context.Context, req *connect.Request[userv1.ClearHistoryRequest]) (*connect.Response[userv1.ClearHistoryResponse], error) {
	id, err := userID(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}
	resp, err := h.svc.ClearHistory(ctx, id, req.Msg.All, req.Msg.Types)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *UserHandler) ListPins(ctx context.Context, req *connect.Request[userv1.ListPinsRequest]) (*connect.Response[userv1.ListPinsResponse], error) {
	id, err := userID(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}
	resp, err := h.svc.ListPins(ctx, id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *UserHandler) Pin(ctx context.Context, req *connect.Request[userv1.PinRequest]) (*connect.Response[userv1.PinResponse], error) {
	id, err := userID(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}
	resp, err := h.svc.Pin(ctx, id, req.Msg.ItemId, req.Msg.ItemType, req.Msg.Title)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *UserHandler) Unpin(ctx context.Context, req *connect.Request[userv1.UnpinRequest]) (*connect.Response[userv1.UnpinResponse], error) {
	id, err := userID(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}
	resp, err := h.svc.Unpin(ctx, id, req.Msg.PinId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *UserHandler) SubmitFeedback(ctx context.Context, req *connect.Request[userv1.SubmitFeedbackRequest]) (*connect.Response[userv1.SubmitFeedbackResponse], error) {
	resp, err := h.svc.SubmitFeedback(ctx, req.Msg.Category, req.Msg.Content, req.Msg.Contact)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *UserHandler) SetAiKey(ctx context.Context, req *connect.Request[userv1.SetAiKeyRequest]) (*connect.Response[userv1.SetAiKeyResponse], error) {
	id, err := userID(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}
	resp, err := h.svc.SetAiKey(ctx, id, req.Msg.Provider, req.Msg.ApiKey)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

var _ antclawv1connect.UserServiceHandler = (*UserHandler)(nil)
