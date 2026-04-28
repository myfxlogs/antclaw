package rpc

import (
	"context"
	"sort"
	"strings"
	"time"

	"connectrpc.com/connect"

	v1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
	"github.com/antclaw/antclaw/internal/infra/apiclient"
	"github.com/antclaw/antclaw/internal/infra/apiclient/deribit"
)

// OptionsHandler 期权服务，使用 Deribit 公开 API 提供 GEX/IV Surface/Skew/Alerts。
type OptionsHandler struct {
	cli *deribit.Client
}

func NewOptionsHandler() *OptionsHandler {
	src := apiclient.NewSource("deribit", apiclient.Options{Timeout: 30 * time.Second})
	return &OptionsHandler{cli: deribit.NewClient(src)}
}

// parseDeribitName 拆解 Deribit 合约名 "BTC-25APR25-65000-C" → strike, type
func parseDeribitName(n string) (strike float64, opt string, ok bool) {
	parts := strings.Split(n, "-")
	if len(parts) < 4 {
		return 0, "", false
	}
	switch parts[3] {
	case "C":
		opt = "call"
	case "P":
		opt = "put"
	default:
		return 0, "", false
	}
	// strike
	var s float64
	for _, c := range parts[2] {
		if c < '0' || c > '9' {
			return 0, "", false
		}
		s = s*10 + float64(c-'0')
	}
	return s, opt, true
}

func (h *OptionsHandler) GetGEX(ctx context.Context, req *connect.Request[v1.GetGEXRequest]) (*connect.Response[v1.GetGEXResponse], error) {
	asset := strings.ToUpper(req.Msg.Asset)
	if asset == "" {
		asset = "BTC"
	}
	books, err := h.cli.GetBookSummaries(ctx, asset)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, err)
	}
	bucketMap := map[float64]*v1.GEXBucket{}
	var totalGex, spot float64
	for _, b := range books {
		if b.UnderlyingPrice > 0 {
			spot = b.UnderlyingPrice
		}
		strike, opt, ok := parseDeribitName(b.InstrumentName)
		if !ok || b.MarkIV <= 0 || b.OpenInterest <= 0 {
			continue
		}
		// 简化 GEX 估算：OI * IV^2 * 1e-4，方向由期权类型决定。
		gex := b.OpenInterest * b.MarkIV * b.MarkIV * 1e-4
		bucket, exists := bucketMap[strike]
		if !exists {
			bucket = &v1.GEXBucket{Strike: strike}
			bucketMap[strike] = bucket
		}
		if opt == "call" {
			bucket.CallGex += gex
			bucket.NetGex += gex
			totalGex += gex
		} else {
			bucket.PutGex += gex
			bucket.NetGex -= gex
			totalGex -= gex
		}
	}
	out := &v1.GetGEXResponse{TotalGex: totalGex, ZeroGamma: spot}
	for _, b := range bucketMap {
		out.Strikes = append(out.Strikes, b)
	}
	sort.Slice(out.Strikes, func(i, j int) bool { return out.Strikes[i].Strike < out.Strikes[j].Strike })
	return connect.NewResponse(out), nil
}

func (h *OptionsHandler) GetIVSurface(ctx context.Context, req *connect.Request[v1.GetIVSurfaceRequest]) (*connect.Response[v1.GetIVSurfaceResponse], error) {
	asset := strings.ToUpper(req.Msg.Asset)
	if asset == "" {
		asset = "BTC"
	}
	insts, err := h.cli.GetInstruments(ctx, asset)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, err)
	}
	books, err := h.cli.GetBookSummaries(ctx, asset)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, err)
	}
	ivByName := map[string]float64{}
	for _, b := range books {
		ivByName[b.InstrumentName] = b.MarkIV / 100
	}
	now := time.Now().UTC()
	out := &v1.GetIVSurfaceResponse{}
	for _, ins := range insts {
		iv, ok := ivByName[ins.InstrumentName]
		if !ok || iv <= 0 {
			continue
		}
		dte := int32((time.UnixMilli(ins.ExpirationTs).UTC().Sub(now)).Hours() / 24)
		out.Points = append(out.Points, &v1.IVPoint{
			Strike: ins.Strike, Dte: dte, Iv: iv, OptionType: ins.OptionType,
		})
	}
	return connect.NewResponse(out), nil
}

func (h *OptionsHandler) GetOptionsSkew(ctx context.Context, req *connect.Request[v1.GetOptionsSkewRequest]) (*connect.Response[v1.GetOptionsSkewResponse], error) {
	asset := strings.ToUpper(req.Msg.Asset)
	if asset == "" {
		asset = "BTC"
	}
	books, err := h.cli.GetBookSummaries(ctx, asset)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, err)
	}
	var atm, callIV, putIV float64
	var nC, nP int
	for _, b := range books {
		if b.MarkIV <= 0 {
			continue
		}
		if strings.HasSuffix(b.InstrumentName, "-C") {
			callIV += b.MarkIV
			nC++
		} else if strings.HasSuffix(b.InstrumentName, "-P") {
			putIV += b.MarkIV
			nP++
		}
		atm += b.MarkIV
	}
	avgC, avgP := 0.0, 0.0
	if nC > 0 {
		avgC = callIV / float64(nC)
	}
	if nP > 0 {
		avgP = putIV / float64(nP)
	}
	out := &v1.GetOptionsSkewResponse{
		Rr_25D: (avgC - avgP) / 100,
		Bf_25D: ((avgC + avgP) / 2 - atm/float64(nC+nP)) / 100,
		AtmIv:  atm / float64(nC+nP) / 100,
	}
	return connect.NewResponse(out), nil
}

func (h *OptionsHandler) GetIVAlerts(ctx context.Context, req *connect.Request[v1.GetIVAlertsRequest]) (*connect.Response[v1.GetIVAlertsResponse], error) {
	return connect.NewResponse(&v1.GetIVAlertsResponse{}), nil
}

var _ antclawv1connect.OptionsServiceHandler = (*OptionsHandler)(nil)
