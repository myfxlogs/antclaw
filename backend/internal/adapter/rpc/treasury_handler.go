package rpc

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	v1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
	"github.com/antclaw/antclaw/internal/infra/apiclient"
	"github.com/antclaw/antclaw/internal/infra/apiclient/ustreasury"
)

// TreasuryHandler 通过 home.treasury.gov 薄客户端提供美债收益率曲线。
type TreasuryHandler struct {
	cli *ustreasury.Client
}

func NewTreasuryHandler() *TreasuryHandler {
	src := apiclient.NewSource("ustreasury", apiclient.Options{Timeout: 30 * time.Second})
	return &TreasuryHandler{cli: ustreasury.NewClient(src)}
}

func (h *TreasuryHandler) GetCurve(ctx context.Context, req *connect.Request[v1.GetCurveRequest]) (*connect.Response[v1.GetCurveResponse], error) {
	year := time.Now().UTC().Year()
	if req.Msg.Date != nil {
		year = req.Msg.Date.AsTime().Year()
	}
	rows, err := h.cli.FetchYearXML(ctx, year)
	if err != nil || len(rows) == 0 {
		return connect.NewResponse(&v1.GetCurveResponse{Date: req.Msg.Date}), nil
	}
	last := rows[len(rows)-1]
	out := &v1.GetCurveResponse{Date: timestamppb.New(last.Date)}
	add := func(t string, v float64) {
		if v != 0 {
			out.Points = append(out.Points, &v1.YieldPoint{Tenor: t, Yield: v})
		}
	}
	add("1M", last.Y1M)
	add("2M", last.Y2M)
	add("3M", last.Y3M)
	add("6M", last.Y6M)
	add("1Y", last.Y1Y)
	add("2Y", last.Y2Y)
	add("3Y", last.Y3Y)
	add("5Y", last.Y5Y)
	add("7Y", last.Y7Y)
	add("10Y", last.Y10Y)
	add("20Y", last.Y20Y)
	add("30Y", last.Y30Y)
	return connect.NewResponse(out), nil
}

func (h *TreasuryHandler) GetAnalysis(ctx context.Context, req *connect.Request[v1.TreasuryServiceGetAnalysisRequest]) (*connect.Response[v1.TreasuryServiceGetAnalysisResponse], error) {
	rows, err := h.cli.FetchYearXML(ctx, time.Now().UTC().Year())
	if err != nil || len(rows) == 0 {
		return connect.NewResponse(&v1.TreasuryServiceGetAnalysisResponse{Regime: "normal"}), nil
	}
	last := rows[len(rows)-1]
	c2s10s := last.Y10Y - last.Y2Y
	c3m10y := last.Y10Y - last.Y3M
	regime := "normal"
	switch {
	case c2s10s < 0 && c3m10y < 0:
		regime = "inverted"
	case c2s10s > 1.0:
		regime = "steepening"
	case c2s10s < 0.2:
		regime = "flattening"
	}
	return connect.NewResponse(&v1.TreasuryServiceGetAnalysisResponse{
		Curve_2S10S: c2s10s, Curve_3M10Y: c3m10y, Regime: regime,
	}), nil
}

var _ antclawv1connect.TreasuryServiceHandler = (*TreasuryHandler)(nil)
