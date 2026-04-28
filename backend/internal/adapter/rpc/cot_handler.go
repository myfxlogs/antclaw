package rpc

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	"github.com/antclaw/antclaw/internal/service/cot"
	cotv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
)

// COTHandler implements antclawv1connect.COTServiceHandler.
type COTHandler struct {
	svc *cot.Service
}

func NewCOTHandler(svc *cot.Service) *COTHandler {
	return &COTHandler{svc: svc}
}

func (h *COTHandler) GetSummary(ctx context.Context, req *connect.Request[cotv1.GetSummaryRequest]) (*connect.Response[cotv1.GetSummaryResponse], error) {
	resp, err := h.svc.GetSummary(ctx, req.Msg.Pair)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *COTHandler) Compare(ctx context.Context, req *connect.Request[cotv1.CompareRequest]) (*connect.Response[cotv1.CompareResponse], error) {
	resp, err := h.svc.Compare(ctx, req.Msg.Pair, req.Msg.DateA, req.Msg.DateB)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *COTHandler) GetSignals(ctx context.Context, req *connect.Request[cotv1.GetSignalsRequest]) (*connect.Response[cotv1.GetSignalsResponse], error) {
	resp, err := h.svc.GetSignals(ctx, req.Msg.PairFilter)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *COTHandler) GetHistory(ctx context.Context, req *connect.Request[cotv1.COTServiceGetHistoryRequest]) (*connect.Response[cotv1.COTServiceGetHistoryResponse], error) {
	// Default limit
	limit := int32(10)

	resp, err := h.svc.GetHistory(ctx, req.Msg.Pair, limit)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *COTHandler) SubscribePairAlert(ctx context.Context, req *connect.Request[cotv1.SubscribePairAlertRequest]) (*connect.Response[cotv1.SubscribePairAlertResponse], error) {
	resp, err := h.svc.SubscribePairAlert(ctx, req.Msg.Pair, req.Msg.Threshold)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	return connect.NewResponse(resp), nil
}

var _ antclawv1connect.COTServiceHandler = (*COTHandler)(nil)

// RegisterCOTService registers the COT service with the mux.
func RegisterCOTService(mux *http.ServeMux, handler antclawv1connect.COTServiceHandler) {
	mux.Handle(antclawv1connect.NewCOTServiceHandler(handler))
}
