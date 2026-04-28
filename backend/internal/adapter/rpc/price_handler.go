package rpc

import (
	"context"

	"connectrpc.com/connect"
	"github.com/antclaw/antclaw/internal/service/price"
	pricev1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
)

// PriceHandler implements antclawv1connect.PriceServiceHandler.
type PriceHandler struct {
	svc *price.Service
}

func NewPriceHandler(svc *price.Service) *PriceHandler {
	return &PriceHandler{svc: svc}
}

func (h *PriceHandler) GetPrice(ctx context.Context, req *connect.Request[pricev1.GetPriceRequest]) (*connect.Response[pricev1.GetPriceResponse], error) {
	resp, err := h.svc.GetPrice(ctx, req.Msg.Pair, req.Msg.Timeframe, req.Msg.Count)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *PriceHandler) GetLevels(ctx context.Context, req *connect.Request[pricev1.GetLevelsRequest]) (*connect.Response[pricev1.GetLevelsResponse], error) {
	resp, err := h.svc.GetLevels(ctx, req.Msg.Pair, req.Msg.Timeframe)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *PriceHandler) GetMarketOverview(ctx context.Context, req *connect.Request[pricev1.GetMarketOverviewRequest]) (*connect.Response[pricev1.GetMarketOverviewResponse], error) {
	resp, err := h.svc.GetMarketOverview(ctx, req.Msg.Category)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *PriceHandler) GetSession(ctx context.Context, req *connect.Request[pricev1.GetSessionRequest]) (*connect.Response[pricev1.GetSessionResponse], error) {
	resp, err := h.svc.GetSession(ctx, req.Msg.Pair)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *PriceHandler) RunScenario(ctx context.Context, req *connect.Request[pricev1.RunScenarioRequest]) (*connect.Response[pricev1.RunScenarioResponse], error) {
	resp, err := h.svc.RunScenario(ctx, req.Msg.Pair, req.Msg.Params)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *PriceHandler) GetRegime(ctx context.Context, req *connect.Request[pricev1.GetRegimeRequest]) (*connect.Response[pricev1.GetRegimeResponse], error) {
	resp, err := h.svc.GetRegime(ctx, req.Msg.Pair, req.Msg.Timeframe)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *PriceHandler) GetSeasonal(ctx context.Context, req *connect.Request[pricev1.GetSeasonalRequest]) (*connect.Response[pricev1.GetSeasonalResponse], error) {
	resp, err := h.svc.GetSeasonal(ctx, req.Msg.Pair, req.Msg.Years)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

var _ antclawv1connect.PriceServiceHandler = (*PriceHandler)(nil)
