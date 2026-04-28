package rpc

import (
"context"

"connectrpc.com/connect"
"github.com/antclaw/antclaw/internal/service/ta"
tav1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
)

// TAHandler implements antclawv1connect.TAServiceHandler.
type TAHandler struct {
	svc *ta.Service
}

// NewTAHandler creates a new TAHandler.
func NewTAHandler(svc *ta.Service) *TAHandler {
	return &TAHandler{svc: svc}
}

func (h *TAHandler) GetIndicators(ctx context.Context, req *connect.Request[tav1.GetIndicatorsRequest]) (*connect.Response[tav1.GetIndicatorsResponse], error) {
	resp, err := h.svc.GetIndicators(ctx, req.Msg.Pair, req.Msg.Timeframe, req.Msg.Indicators)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *TAHandler) GetElliott(ctx context.Context, req *connect.Request[tav1.GetElliottRequest]) (*connect.Response[tav1.GetElliottResponse], error) {
	resp, err := h.svc.GetElliott(ctx, req.Msg.Pair, req.Msg.Timeframe)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *TAHandler) GetWyckoff(ctx context.Context, req *connect.Request[tav1.GetWyckoffRequest]) (*connect.Response[tav1.GetWyckoffResponse], error) {
	resp, err := h.svc.GetWyckoff(ctx, req.Msg.Pair)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *TAHandler) GetIct(ctx context.Context, req *connect.Request[tav1.GetIctRequest]) (*connect.Response[tav1.GetIctResponse], error) {
	resp, err := h.svc.GetIct(ctx, req.Msg.Pair, req.Msg.Timeframe)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *TAHandler) GetAmt(ctx context.Context, req *connect.Request[tav1.GetAmtRequest]) (*connect.Response[tav1.GetAmtResponse], error) {
	resp, err := h.svc.GetAmt(ctx, req.Msg.Pair, req.Msg.LookbackDays)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *TAHandler) GetOrderflow(ctx context.Context, req *connect.Request[tav1.GetOrderflowRequest]) (*connect.Response[tav1.GetOrderflowResponse], error) {
	resp, err := h.svc.GetOrderflow(ctx, req.Msg.Pair, req.Msg.Timeframe)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *TAHandler) GetVolumeProfile(ctx context.Context, req *connect.Request[tav1.GetVolumeProfileRequest]) (*connect.Response[tav1.GetVolumeProfileResponse], error) {
	resp, err := h.svc.GetVolumeProfile(ctx, req.Msg.Pair)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *TAHandler) GetIntermarket(ctx context.Context, req *connect.Request[tav1.GetIntermarketRequest]) (*connect.Response[tav1.GetIntermarketResponse], error) {
	resp, err := h.svc.GetIntermarket(ctx, req.Msg.Pair)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

var _ antclawv1connect.TAServiceHandler = (*TAHandler)(nil)
