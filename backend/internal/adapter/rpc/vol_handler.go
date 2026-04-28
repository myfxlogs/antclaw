package rpc

import (
	"context"

	"connectrpc.com/connect"
	"github.com/antclaw/antclaw/internal/service/vol"
	volv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
)

// VolHandler implements antclawv1connect.VolServiceHandler.
type VolHandler struct {
	svc *vol.Service
}

func NewVolHandler(svc *vol.Service) *VolHandler {
	return &VolHandler{svc: svc}
}

func (h *VolHandler) GetVix(ctx context.Context, req *connect.Request[volv1.GetVixRequest]) (*connect.Response[volv1.GetVixResponse], error) {
	resp, err := h.svc.GetVix(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *VolHandler) GetMove(ctx context.Context, req *connect.Request[volv1.GetMoveRequest]) (*connect.Response[volv1.GetMoveResponse], error) {
	resp, err := h.svc.GetMove(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *VolHandler) GetDvol(ctx context.Context, req *connect.Request[volv1.GetDvolRequest]) (*connect.Response[volv1.GetDvolResponse], error) {
	resp, err := h.svc.GetDvol(ctx, req.Msg.Asset)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *VolHandler) GetGex(ctx context.Context, req *connect.Request[volv1.GetGexRequest]) (*connect.Response[volv1.GetGexResponse], error) {
	resp, err := h.svc.GetGex(ctx, req.Msg.Pair)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *VolHandler) GetIvol(ctx context.Context, req *connect.Request[volv1.GetIvolRequest]) (*connect.Response[volv1.GetIvolResponse], error) {
	resp, err := h.svc.GetIvol(ctx, req.Msg.Pair, req.Msg.Expiry)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *VolHandler) GetSkew(ctx context.Context, req *connect.Request[volv1.GetSkewRequest]) (*connect.Response[volv1.GetSkewResponse], error) {
	resp, err := h.svc.GetSkew(ctx, req.Msg.Pair)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *VolHandler) GetSkewVixAlert(ctx context.Context, req *connect.Request[volv1.GetSkewVixAlertRequest]) (*connect.Response[volv1.GetSkewVixAlertResponse], error) {
	resp, err := h.svc.GetSkewVixAlert(ctx, req.Msg.Pair)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

var _ antclawv1connect.VolServiceHandler = (*VolHandler)(nil)
