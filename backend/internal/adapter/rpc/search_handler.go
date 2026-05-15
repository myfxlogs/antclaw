// Package rpc provides the Search Connect-RPC handler (P1).
package rpc

import (
	"context"

	"connectrpc.com/connect"
	alfqv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
	"github.com/antclaw/antclaw/internal/service/search"
)

// SearchHandler implements SearchServiceHandler.
type SearchHandler struct {
	svc *search.Service
}

// NewSearchHandler creates a new SearchHandler.
func NewSearchHandler(svc *search.Service) *SearchHandler {
	return &SearchHandler{svc: svc}
}

// Search performs a cross-entity search.
func (h *SearchHandler) Search(ctx context.Context, req *connect.Request[alfqv1.SearchRequest]) (*connect.Response[alfqv1.SearchResponse], error) {
	resp, err := h.svc.Search(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

var _ antclawv1connect.SearchServiceHandler = (*SearchHandler)(nil)
