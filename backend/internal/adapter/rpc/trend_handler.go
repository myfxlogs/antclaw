// Package rpc provides the Trend Connect-RPC handler (P1).
package rpc

import (
	"context"

	"connectrpc.com/connect"
	alfqv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
	"github.com/antclaw/antclaw/internal/service/trend"
)

// TrendHandler implements TrendServiceHandler.
type TrendHandler struct {
	svc *trend.Service
}

// NewTrendHandler creates a new TrendHandler.
func NewTrendHandler(svc *trend.Service) *TrendHandler {
	return &TrendHandler{svc: svc}
}

// ListTrendingTopics returns trending topics.
func (h *TrendHandler) ListTrendingTopics(ctx context.Context, req *connect.Request[alfqv1.ListTrendingTopicsRequest]) (*connect.Response[alfqv1.ListTrendingTopicsResponse], error) {
	resp, err := h.svc.ListTrendingTopics(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

// ListHotSymbols returns hot symbols.
func (h *TrendHandler) ListHotSymbols(ctx context.Context, req *connect.Request[alfqv1.ListHotSymbolsRequest]) (*connect.Response[alfqv1.ListHotSymbolsResponse], error) {
	resp, err := h.svc.ListHotSymbols(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

var _ antclawv1connect.TrendServiceHandler = (*TrendHandler)(nil)
