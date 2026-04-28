package rpc

import (
"context"

"connectrpc.com/connect"
"github.com/antclaw/antclaw/internal/service/sentiment"
sentimentv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
)

// SentimentHandler implements antclawv1connect.SentimentServiceHandler.
type SentimentHandler struct {
	svc *sentiment.Service
}

// NewSentimentHandler creates a new SentimentHandler.
func NewSentimentHandler(svc *sentiment.Service) *SentimentHandler {
	return &SentimentHandler{svc: svc}
}

func (h *SentimentHandler) GetSentiment(ctx context.Context, req *connect.Request[sentimentv1.GetSentimentRequest]) (*connect.Response[sentimentv1.GetSentimentResponse], error) {
	resp, err := h.svc.GetSentiment(ctx, req.Msg.Asset)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *SentimentHandler) GetOnchain(ctx context.Context, req *connect.Request[sentimentv1.GetOnchainRequest]) (*connect.Response[sentimentv1.GetOnchainResponse], error) {
	resp, err := h.svc.GetOnchain(ctx, req.Msg.Asset)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *SentimentHandler) GetDefiHealth(ctx context.Context, req *connect.Request[sentimentv1.GetDefiHealthRequest]) (*connect.Response[sentimentv1.GetDefiHealthResponse], error) {
	resp, err := h.svc.GetDefiHealth(ctx, req.Msg.Chain)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *SentimentHandler) GetCarryMonitor(ctx context.Context, req *connect.Request[sentimentv1.GetCarryMonitorRequest]) (*connect.Response[sentimentv1.GetCarryMonitorResponse], error) {
	resp, err := h.svc.GetCarryMonitor(ctx, req.Msg.Category)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

var _ antclawv1connect.SentimentServiceHandler = (*SentimentHandler)(nil)
