package rpc

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	v1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
	"github.com/antclaw/antclaw/internal/infra/apiclient"
	"github.com/antclaw/antclaw/internal/infra/apiclient/firecrawl"
)

// FedWatchHandler CME FedWatch 概率，通过 Firecrawl JSON 抽取实现。
// 未配置 FIRECRAWL_API_KEY 时返回空概率列表（HTTP 200，避免阻塞前端）。
type FedWatchHandler struct {
	fc *firecrawl.Client
}

func NewFedWatchHandler() *FedWatchHandler {
	src := apiclient.NewSource("firecrawl", apiclient.Options{Timeout: 60 * time.Second})
	return &FedWatchHandler{fc: firecrawl.NewClient(src)}
}

func (h *FedWatchHandler) GetFOMCProbabilities(ctx context.Context, req *connect.Request[v1.GetFOMCProbabilitiesRequest]) (*connect.Response[v1.GetFOMCProbabilitiesResponse], error) {
	resp := &v1.GetFOMCProbabilitiesResponse{MeetingDate: req.Msg.MeetingDate}
	if !h.fc.IsAvailable() {
		return connect.NewResponse(resp), nil
	}
	snap, err := h.fc.FetchFedWatch(ctx)
	if err != nil {
		return connect.NewResponse(resp), nil
	}
	if t, perr := time.Parse("2006-01-02", snap.NextMeetingDate); perr == nil {
		resp.MeetingDate = timestamppb.New(t)
	}
	// CME FedWatch 默认以"是否变动 25/50bp"形式给出概率；映射到 RateProbability：
	// rate_low/high 用 bps 偏移（hold=0、cut25=-25、cut50=-50、hike25=+25），probability 为 %。
	type pair struct {
		low, high, prob float64
	}
	for _, p := range []pair{
		{0, 0, snap.HoldProbability},
		{-25, -25, snap.Cut25Probability},
		{-50, -50, snap.Cut50Probability},
		{25, 25, snap.Hike25Probability},
	} {
		if p.prob <= 0 {
			continue
		}
		resp.Probabilities = append(resp.Probabilities, &v1.RateProbability{
			RateLow: p.low, RateHigh: p.high, Probability: p.prob,
		})
	}
	return connect.NewResponse(resp), nil
}

var _ antclawv1connect.FedWatchServiceHandler = (*FedWatchHandler)(nil)
