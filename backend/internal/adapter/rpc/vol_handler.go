package rpc

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	volv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
	"github.com/antclaw/antclaw/internal/service/vol"
)

// volErr 把 service 层错误映射到合适的 connect code。
// vol.ErrUnavailable 表示上游数据源（firecrawl / Deribit 等）暂不可用，应回 Unavailable(503)
// 而非 Internal(500)，让前端可据此显示"暂未配置 / 服务降级"提示。
func volErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, vol.ErrUnavailable) {
		return connect.NewError(connect.CodeUnavailable, err)
	}
	return connect.NewError(connect.CodeInternal, err)
}

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
		return nil, volErr(err)
	}
	return connect.NewResponse(resp), nil
}

func (h *VolHandler) GetMove(ctx context.Context, req *connect.Request[volv1.GetMoveRequest]) (*connect.Response[volv1.GetMoveResponse], error) {
	resp, err := h.svc.GetMove(ctx)
	if err != nil {
		return nil, volErr(err)
	}
	return connect.NewResponse(resp), nil
}

func (h *VolHandler) GetDvol(ctx context.Context, req *connect.Request[volv1.GetDvolRequest]) (*connect.Response[volv1.GetDvolResponse], error) {
	resp, err := h.svc.GetDvol(ctx, req.Msg.Asset)
	if err != nil {
		return nil, volErr(err)
	}
	return connect.NewResponse(resp), nil
}

func (h *VolHandler) GetGex(ctx context.Context, req *connect.Request[volv1.GetGexRequest]) (*connect.Response[volv1.GetGexResponse], error) {
	resp, err := h.svc.GetGex(ctx, req.Msg.Pair)
	if err != nil {
		return nil, volErr(err)
	}
	return connect.NewResponse(resp), nil
}

func (h *VolHandler) GetIvol(ctx context.Context, req *connect.Request[volv1.GetIvolRequest]) (*connect.Response[volv1.GetIvolResponse], error) {
	resp, err := h.svc.GetIvol(ctx, req.Msg.Pair, req.Msg.Expiry)
	if err != nil {
		return nil, volErr(err)
	}
	return connect.NewResponse(resp), nil
}

func (h *VolHandler) GetSkew(ctx context.Context, req *connect.Request[volv1.GetSkewRequest]) (*connect.Response[volv1.GetSkewResponse], error) {
	resp, err := h.svc.GetSkew(ctx, req.Msg.Pair)
	if err != nil {
		return nil, volErr(err)
	}
	return connect.NewResponse(resp), nil
}

func (h *VolHandler) GetSkewVixAlert(ctx context.Context, req *connect.Request[volv1.GetSkewVixAlertRequest]) (*connect.Response[volv1.GetSkewVixAlertResponse], error) {
	resp, err := h.svc.GetSkewVixAlert(ctx, req.Msg.Pair)
	if err != nil {
		return nil, volErr(err)
	}
	return connect.NewResponse(resp), nil
}

var _ antclawv1connect.VolServiceHandler = (*VolHandler)(nil)
