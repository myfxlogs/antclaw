package rpc

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	v1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
	"github.com/antclaw/antclaw/internal/infra/apiclient"
	"github.com/antclaw/antclaw/internal/infra/apiclient/secedgar"
)

// SECHandler 通过 SEC EDGAR 薄客户端提供 filings 列表。
type SECHandler struct {
	cli *secedgar.Client
}

func NewSECHandler() *SECHandler {
	src := apiclient.NewSource("secedgar", apiclient.Options{Timeout: 15 * time.Second})
	return &SECHandler{cli: secedgar.NewClient(src, "")}
}

func (h *SECHandler) ListFilings(ctx context.Context, req *connect.Request[v1.ListFilingsRequest]) (*connect.Response[v1.ListFilingsResponse], error) {
	if req.Msg.Cik == "" {
		return connect.NewResponse(&v1.ListFilingsResponse{}), nil
	}
	subs, err := h.cli.GetSubmissions(ctx, req.Msg.Cik)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, err)
	}
	out := &v1.ListFilingsResponse{}
	rec := subs.Filings.Recent
	limit := int(req.Msg.Limit)
	if limit <= 0 {
		limit = 50
	}
	for i := range rec.AccessionNumber {
		if req.Msg.FormType != "" && rec.Form[i] != req.Msg.FormType {
			continue
		}
		if len(out.Items) >= limit {
			break
		}
		filed, _ := time.Parse("2006-01-02", rec.FilingDate[i])
		out.Items = append(out.Items, &v1.SECFiling{
			AccessionNumber: rec.AccessionNumber[i],
			FormType:        rec.Form[i],
			FiledAt:         timestamppb.New(filed),
			CompanyName:     subs.Name,
			Url:             fmt.Sprintf("https://www.sec.gov/Archives/edgar/data/%s/%s/%s", subs.CIK, rec.AccessionNumber[i], rec.PrimaryDocument[i]),
		})
	}
	return connect.NewResponse(out), nil
}
func (h *SECHandler) GetFiling(ctx context.Context, req *connect.Request[v1.GetFilingRequest]) (*connect.Response[v1.GetFilingResponse], error) {
	return connect.NewResponse(&v1.GetFilingResponse{}), nil
}
func (h *SECHandler) GetAnalysis(ctx context.Context, req *connect.Request[v1.SECServiceGetAnalysisRequest]) (*connect.Response[v1.SECServiceGetAnalysisResponse], error) {
	return connect.NewResponse(&v1.SECServiceGetAnalysisResponse{Sentiment: "neutral"}), nil
}

var _ antclawv1connect.SECServiceHandler = (*SECHandler)(nil)
