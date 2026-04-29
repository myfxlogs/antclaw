package rpc

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	backtestv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
	"github.com/antclaw/antclaw/internal/service/backtest"
)

// BacktestHandler implements antclawv1connect.BacktestServiceHandler.
type BacktestHandler struct {
	svc *backtest.Service
}

// NewBacktestHandler creates a new BacktestHandler.
func NewBacktestHandler(svc *backtest.Service) *BacktestHandler {
	return &BacktestHandler{svc: svc}
}

func (h *BacktestHandler) RunBacktest(ctx context.Context, req *connect.Request[backtestv1.RunBacktestRequest]) (*connect.Response[backtestv1.RunBacktestResponse], error) {
	resp, err := h.svc.RunBacktest(ctx, req.Msg.Config, req.Msg.IdempotencyKey)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *BacktestHandler) GetBacktest(ctx context.Context, req *connect.Request[backtestv1.GetBacktestRequest]) (*connect.Response[backtestv1.GetBacktestResponse], error) {
	resp, err := h.svc.GetBacktest(ctx, req.Msg.TaskId)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *BacktestHandler) GetAccuracy(ctx context.Context, req *connect.Request[backtestv1.GetAccuracyRequest]) (*connect.Response[backtestv1.GetAccuracyResponse], error) {
	resp, err := h.svc.GetAccuracy(ctx, req.Msg.StrategyId, req.Msg.Period)
	if err != nil {
		if errors.Is(err, backtest.ErrDataInsufficient) {
			return nil, connect.NewError(connect.CodeFailedPrecondition, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *BacktestHandler) RunQuantBt(ctx context.Context, req *connect.Request[backtestv1.RunQuantBtRequest]) (*connect.Response[backtestv1.RunQuantBtResponse], error) {
	resp, err := h.svc.RunQuantBt(ctx, req.Msg.Config)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *BacktestHandler) RunVpBt(ctx context.Context, req *connect.Request[backtestv1.RunVpBtRequest]) (*connect.Response[backtestv1.RunVpBtResponse], error) {
	resp, err := h.svc.RunVpBt(ctx, req.Msg.Config)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *BacktestHandler) RunCtaBt(ctx context.Context, req *connect.Request[backtestv1.RunCtaBtRequest]) (*connect.Response[backtestv1.RunCtaBtResponse], error) {
	resp, err := h.svc.RunCtaBt(ctx, req.Msg.Config)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *BacktestHandler) RunWalkforward(ctx context.Context, req *connect.Request[backtestv1.RunWalkforwardRequest]) (*connect.Response[backtestv1.RunWalkforwardResponse], error) {
	resp, err := h.svc.RunWalkforward(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *BacktestHandler) GetWalkforwardResult(ctx context.Context, req *connect.Request[backtestv1.GetWalkforwardResultRequest]) (*connect.Response[backtestv1.GetWalkforwardResultResponse], error) {
	resp, err := h.svc.GetWalkforwardResult(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *BacktestHandler) RunBootstrap(ctx context.Context, req *connect.Request[backtestv1.RunBootstrapRequest]) (*connect.Response[backtestv1.RunBootstrapResponse], error) {
	resp, err := h.svc.RunBootstrap(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

// M-B 增量
func (h *BacktestHandler) RunMonteCarlo(ctx context.Context, req *connect.Request[backtestv1.RunMonteCarloRequest]) (*connect.Response[backtestv1.RunMonteCarloResponse], error) {
	resp, err := h.svc.RunMonteCarlo(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *BacktestHandler) GetTrades(ctx context.Context, req *connect.Request[backtestv1.GetTradesRequest]) (*connect.Response[backtestv1.GetTradesResponse], error) {
	resp, err := h.svc.GetTrades(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *BacktestHandler) GetMetricsByRegime(ctx context.Context, req *connect.Request[backtestv1.GetMetricsByRegimeRequest]) (*connect.Response[backtestv1.GetMetricsByRegimeResponse], error) {
	resp, err := h.svc.GetMetricsByRegime(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

var _ antclawv1connect.BacktestServiceHandler = (*BacktestHandler)(nil)
