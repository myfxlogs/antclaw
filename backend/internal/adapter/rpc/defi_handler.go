package rpc

import (
	"context"
	"sort"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	v1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
	"github.com/antclaw/antclaw/internal/infra/apiclient"
	"github.com/antclaw/antclaw/internal/infra/apiclient/defillama"
)

// DeFiHandler 通过 DefiLlama 薄客户端提供 DeFi 协议 TVL 数据。
type DeFiHandler struct {
	cli *defillama.Client
}

func NewDeFiHandler() *DeFiHandler {
	src := apiclient.NewSource("defillama", apiclient.Options{Timeout: 60 * time.Second})
	return &DeFiHandler{cli: defillama.NewClient(src)}
}

func (h *DeFiHandler) GetTopProtocols(ctx context.Context, req *connect.Request[v1.GetTopProtocolsRequest]) (*connect.Response[v1.GetTopProtocolsResponse], error) {
	all, err := h.cli.ListProtocols(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, err)
	}
	chain := req.Msg.Chain
	filtered := all[:0]
	for _, p := range all {
		if chain == "" || p.Chain == chain {
			filtered = append(filtered, p)
		}
	}
	sort.Slice(filtered, func(i, j int) bool { return filtered[i].TVL > filtered[j].TVL })
	limit := int(req.Msg.Limit)
	if limit <= 0 || limit > len(filtered) {
		limit = len(filtered)
	}
	out := &v1.GetTopProtocolsResponse{}
	for _, p := range filtered[:limit] {
		out.Items = append(out.Items, &v1.DeFiProtocol{
			Slug: p.Slug, Name: p.Name, Category: p.Category,
			TvlUsd: p.TVL, Change_1D: p.Change1d, Change_7D: p.Change7d,
		})
	}
	return connect.NewResponse(out), nil
}

func (h *DeFiHandler) GetProtocolTVL(ctx context.Context, req *connect.Request[v1.GetProtocolTVLRequest]) (*connect.Response[v1.GetProtocolTVLResponse], error) {
	d, err := h.cli.GetProtocolTVL(ctx, req.Msg.Slug)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, err)
	}
	out := &v1.GetProtocolTVLResponse{}
	for _, p := range d.TVL {
		out.Points = append(out.Points, &v1.TVLPoint{
			Time:   timestamppb.New(time.Unix(p.Date, 0).UTC()),
			TvlUsd: p.TVL,
		})
	}
	return connect.NewResponse(out), nil
}

func (h *DeFiHandler) GetAnalysis(ctx context.Context, req *connect.Request[v1.DeFiServiceGetAnalysisRequest]) (*connect.Response[v1.DeFiServiceGetAnalysisResponse], error) {
	all, err := h.cli.ListProtocols(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, err)
	}
	var total, weightedChg7 float64
	for _, p := range all {
		if req.Msg.Chain != "" && p.Chain != req.Msg.Chain {
			continue
		}
		total += p.TVL
		weightedChg7 += p.TVL * p.Change7d
	}
	change7d := 0.0
	if total > 0 {
		change7d = weightedChg7 / total
	}
	regime := "neutral"
	switch {
	case change7d > 5:
		regime = "expansion"
	case change7d < -5:
		regime = "contraction"
	}
	return connect.NewResponse(&v1.DeFiServiceGetAnalysisResponse{
		TotalTvl: total, TvlChange_7D: change7d, Regime: regime,
	}), nil
}

var _ antclawv1connect.DeFiServiceHandler = (*DeFiHandler)(nil)
