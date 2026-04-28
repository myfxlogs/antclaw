package rpc

import (
	"context"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	v1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
	"github.com/antclaw/antclaw/internal/service/regime"
)

// RegimeHandler 暴露 RegimeService 的 RPC，由 internal/service/regime 实现实际算法。
type RegimeHandler struct {
	svc *regime.Service
}

func NewRegimeHandler(svc *regime.Service) *RegimeHandler { return &RegimeHandler{svc: svc} }

func (h *RegimeHandler) GetOverlay(ctx context.Context, req *connect.Request[v1.GetOverlayRequest]) (*connect.Response[v1.GetOverlayResponse], error) {
	r, err := h.svc.Compute(ctx, req.Msg.Symbol, req.Msg.Timeframe, req.Msg.ContractCode)
	if err != nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, err)
	}
	resp := &v1.GetOverlayResponse{
		Symbol:          r.Symbol,
		Timeframe:       r.Timeframe,
		ComputedAt:      timestamppb.New(r.ComputedAt),
		UnifiedScore:    r.UnifiedScore,
		UnifiedLabel:    r.UnifiedLabel,
		Hmm:             toPB(r.HMM),
		Garch:           toPB(r.GARCH),
		Adx:             toPB(r.ADX),
		Cot:             toPB(r.COT),
		AvailableModels: r.AvailableModels,
	}
	return connect.NewResponse(resp), nil
}

func (h *RegimeHandler) ListRecent(ctx context.Context, req *connect.Request[v1.ListRecentRequest]) (*connect.Response[v1.ListRecentResponse], error) {
	items, err := h.svc.ListRecent(ctx, req.Msg.Symbol, req.Msg.Timeframe, int(req.Msg.Days))
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, err)
	}
	resp := &v1.ListRecentResponse{}
	for _, it := range items {
		resp.Items = append(resp.Items, &v1.OverlaySnapshot{
			Time:         timestamppb.New(it.Time),
			UnifiedScore: it.UnifiedScore,
			UnifiedLabel: it.UnifiedLabel,
		})
	}
	return connect.NewResponse(resp), nil
}

func toPB(s regime.SubModel) *v1.RegimeSubModel {
	return &v1.RegimeSubModel{
		Name:       s.Name,
		Available:  s.Available,
		Score:      s.Score,
		Weight:     s.Weight,
		State:      s.State,
		Confidence: s.Confidence,
	}
}

var _ antclawv1connect.RegimeServiceHandler = (*RegimeHandler)(nil)
