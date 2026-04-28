package rpc

import (
	"context"

	"connectrpc.com/connect"
	reportv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
	"github.com/antclaw/antclaw/internal/service/report"
)

type ReportHandler struct {
	svc *report.Service
}

func NewReportHandler(svc *report.Service) *ReportHandler {
	return &ReportHandler{svc: svc}
}

func (h *ReportHandler) GetReport(ctx context.Context, req *connect.Request[reportv1.GetReportRequest]) (*connect.Response[reportv1.GetReportResponse], error) {
	out, err := h.svc.GetReport(ctx, report.Request{
		Symbol:        req.Msg.Symbol,
		Sections:      req.Msg.Sections,
		WithAISummary: req.Msg.WithAiSummary,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	resp := &reportv1.GetReportResponse{
		Symbol:          out.Symbol,
		GeneratedAt:     out.GeneratedAt,
		Accuracy_1W:     out.Accuracy1W,
		Accuracy_1M:     out.Accuracy1M,
		MissingSections: out.MissingSections,
	}
	if out.Bias != nil {
		resp.Bias = &reportv1.ReportBias{
			Pair:       out.Bias.Pair,
			Direction:  out.Bias.Direction,
			Confidence: out.Bias.Confidence,
			Timeframe:  out.Bias.Timeframe,
		}
	}
	if out.Unified != nil {
		resp.Unified = &reportv1.ReportUnified{
			Symbol:         out.Unified.Symbol,
			IssuedAt:       out.Unified.IssuedAt.Format("2006-01-02T15:04:05Z07:00"),
			Recommendation: out.Unified.Recommendation,
			UnifiedScore:   out.Unified.UnifiedScore,
			Confidence:     out.Unified.Confidence,
		}
	}
	return connect.NewResponse(resp), nil
}

var _ antclawv1connect.ReportServiceHandler = (*ReportHandler)(nil)
