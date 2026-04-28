package rpc

import (
	"context"
	"strings"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	v1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
	"github.com/antclaw/antclaw/internal/infra/apiclient"
	"github.com/antclaw/antclaw/internal/infra/apiclient/bis"
	"github.com/antclaw/antclaw/internal/infra/apiclient/dtcc"
	"github.com/antclaw/antclaw/internal/infra/apiclient/ecb"
	"github.com/antclaw/antclaw/internal/infra/apiclient/eurostat"
	"github.com/antclaw/antclaw/internal/infra/apiclient/imf"
	"github.com/antclaw/antclaw/internal/infra/apiclient/oecd"
	"github.com/antclaw/antclaw/internal/infra/apiclient/snb"
	"github.com/antclaw/antclaw/internal/infra/apiclient/worldbank"
)

// MacroExtrasHandler 容纳 BIS/IMF/WB/ECB/Eurostat/OECD/SNB/DTCC 等宏观系列查询。
// 已接入（免 Key）：worldbank、imf、ecb、bis、eurostat、oecd、snb、dtcc。
// 待接入（付费）：trading economics（接入需 FIRECRAWL_API_KEY，参考 SentimentExtras 路径）。
type MacroExtrasHandler struct {
	wb       *worldbank.Client
	imf      *imf.Client
	ecb      *ecb.Client
	bis      *bis.Client
	eurostat *eurostat.Client
	oecd     *oecd.Client
	snb      *snb.Client
	dtcc     *dtcc.Client
}

func NewMacroExtrasHandler() *MacroExtrasHandler {
	mk := func(name string, t time.Duration) apiclient.Source {
		return apiclient.NewSource(name, apiclient.Options{Timeout: t})
	}
	return &MacroExtrasHandler{
		wb:       worldbank.NewClient(mk("worldbank", 30*time.Second)),
		imf:      imf.NewClient(mk("imf", 30*time.Second)),
		ecb:      ecb.NewClient(mk("ecb", 30*time.Second)),
		bis:      bis.NewClient(mk("bis", 60*time.Second)),
		eurostat: eurostat.NewClient(mk("eurostat", 30*time.Second)),
		oecd:     oecd.NewClient(mk("oecd", 60*time.Second)),
		snb:      snb.NewClient(mk("snb", 30*time.Second)),
		dtcc:     dtcc.NewClient(mk("dtcc", 60*time.Second)),
	}
}

func (h *MacroExtrasHandler) GetSeries(ctx context.Context, req *connect.Request[v1.MacroExtrasServiceGetSeriesRequest]) (*connect.Response[v1.MacroExtrasServiceGetSeriesResponse], error) {
	out := &v1.MacroExtrasServiceGetSeriesResponse{}
	switch strings.ToLower(req.Msg.Source) {
	case "worldbank", "wb":
		points, err := h.wb.GetSeries(ctx, req.Msg.SeriesId)
		if err != nil {
			return nil, connect.NewError(connect.CodeUnavailable, err)
		}
		appendMacroPoints(out, points)
		out.Frequency = "annual"
		out.Unit = "indicator-native"
	case "imf":
		points, err := h.imf.GetSeries(ctx, req.Msg.SeriesId)
		if err != nil {
			return nil, connect.NewError(connect.CodeUnavailable, err)
		}
		for _, p := range points {
			out.Points = append(out.Points, &v1.MacroPoint{Time: timestamppb.New(p.Date), Value: p.Value})
		}
		out.Frequency = "annual"
		out.Unit = "indicator-native"
	case "ecb":
		points, err := h.ecb.GetSeries(ctx, req.Msg.SeriesId)
		if err != nil {
			return nil, connect.NewError(connect.CodeUnavailable, err)
		}
		for _, p := range points {
			out.Points = append(out.Points, &v1.MacroPoint{Time: timestamppb.New(p.Date), Value: p.Value})
		}
		out.Frequency = "varies"
		out.Unit = "ECB-native"
	case "bis":
		points, err := h.bis.GetSeries(ctx, req.Msg.SeriesId)
		if err != nil {
			return nil, connect.NewError(connect.CodeUnavailable, err)
		}
		for _, p := range points {
			out.Points = append(out.Points, &v1.MacroPoint{Time: timestamppb.New(p.Date), Value: p.Value})
		}
		out.Frequency = "varies"
		out.Unit = "BIS-native"
	case "eurostat":
		points, err := h.eurostat.GetSeries(ctx, req.Msg.SeriesId)
		if err != nil {
			return nil, connect.NewError(connect.CodeUnavailable, err)
		}
		for _, p := range points {
			out.Points = append(out.Points, &v1.MacroPoint{Time: timestamppb.New(p.Date), Value: p.Value})
		}
		out.Frequency = "varies"
		out.Unit = "Eurostat-native"
	case "oecd":
		points, err := h.oecd.GetSeries(ctx, req.Msg.SeriesId)
		if err != nil {
			return nil, connect.NewError(connect.CodeUnavailable, err)
		}
		for _, p := range points {
			out.Points = append(out.Points, &v1.MacroPoint{Time: timestamppb.New(p.Date), Value: p.Value})
		}
		out.Frequency = "varies"
		out.Unit = "OECD-native"
	case "snb":
		points, err := h.snb.GetSeries(ctx, req.Msg.SeriesId)
		if err != nil {
			return nil, connect.NewError(connect.CodeUnavailable, err)
		}
		for _, p := range points {
			out.Points = append(out.Points, &v1.MacroPoint{Time: timestamppb.New(p.Date), Value: p.Value})
		}
		out.Frequency = "varies"
		out.Unit = "SNB-native"
	case "dtcc":
		// series_id 形如 "FOREX/2026-04-25" 或 "FOREX"（默认昨日）。
		asOf := time.Now().UTC().AddDate(0, 0, -1)
		seg := strings.SplitN(req.Msg.SeriesId, "/", 2)
		if len(seg) == 2 {
			if t, err := time.Parse("2006-01-02", seg[1]); err == nil {
				asOf = t
			}
		}
		aggs, err := h.dtcc.FetchByDate(ctx, asOf, nil)
		if err != nil {
			return nil, connect.NewError(connect.CodeUnavailable, err)
		}
		for _, a := range aggs {
			out.Points = append(out.Points, &v1.MacroPoint{
				Time:  timestamppb.New(a.AsOf),
				Value: a.TotalNotional,
				Label: a.Pair,
			})
		}
		out.Frequency = "daily"
		out.Unit = "USD-equivalent (millions)"
	default:
		// 其余源待实现，返回空集合不报错以保持 API 可用
	}
	return connect.NewResponse(out), nil
}

func appendMacroPoints(out *v1.MacroExtrasServiceGetSeriesResponse, points []worldbank.Point) {
	for _, p := range points {
		out.Points = append(out.Points, &v1.MacroPoint{
			Time:  timestamppb.New(p.Date),
			Value: p.Value,
		})
	}
}
func (h *MacroExtrasHandler) ListAvailableSeries(ctx context.Context, req *connect.Request[v1.ListAvailableSeriesRequest]) (*connect.Response[v1.ListAvailableSeriesResponse], error) {
	return connect.NewResponse(&v1.ListAvailableSeriesResponse{}), nil
}

var _ antclawv1connect.MacroExtrasServiceHandler = (*MacroExtrasHandler)(nil)
