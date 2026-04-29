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
	"github.com/antclaw/antclaw/internal/infra/apiclient/cboe"
	"github.com/antclaw/antclaw/internal/infra/apiclient/firecrawl"
)

// SentimentExtrasHandler 扩展情绪指标。
// CBOE Put/Call 免 Key；MyFXBook / Finviz 通过 Firecrawl JSON 抽取，需 FIRECRAWL_API_KEY。
// Insider / CryptoSocial 仍为骨架，待专门 vendor 接入后再代。
type SentimentExtrasHandler struct {
	cboe *cboe.Client
	fc   *firecrawl.Client
}

func NewSentimentExtrasHandler() *SentimentExtrasHandler {
	cboeSrc := apiclient.NewSource("cboe", apiclient.Options{Timeout: 30 * time.Second})
	fcSrc := apiclient.NewSource("firecrawl", apiclient.Options{Timeout: 60 * time.Second})
	return &SentimentExtrasHandler{
		cboe: cboe.NewClient(cboeSrc),
		fc:   firecrawl.NewClient(fcSrc),
	}
}

// NewSentimentExtrasHandlerWithResolver 优先用数据库中的 firecrawl 密钥构造。
func NewSentimentExtrasHandlerWithResolver(r SecretReader) *SentimentExtrasHandler {
	cboeSrc := apiclient.NewSource("cboe", apiclient.Options{Timeout: 30 * time.Second})
	fcSrc := apiclient.NewSource("firecrawl", apiclient.Options{Timeout: 60 * time.Second})
	key := ""
	if r != nil {
		key = r.GetSecret("firecrawl")
	}
	return &SentimentExtrasHandler{
		cboe: cboe.NewClient(cboeSrc),
		fc:   firecrawl.NewClientWithKey(fcSrc, key),
	}
}

func (h *SentimentExtrasHandler) GetCBOEPutCall(ctx context.Context, req *connect.Request[v1.GetCBOEPutCallRequest]) (*connect.Response[v1.GetCBOEPutCallResponse], error) {
	d, err := h.cboe.FetchLatest(ctx)
	if err != nil {
		// 数据源不可用时返回空响应（200），避免阻塞前端
		return connect.NewResponse(&v1.GetCBOEPutCallResponse{}), nil
	}
	return connect.NewResponse(&v1.GetCBOEPutCallResponse{
		Date:     timestamppb.New(d.Date),
		TotalPc:  d.TotalPC,
		EquityPc: d.EquityPC,
		IndexPc:  d.IndexPC,
	}), nil
}

func (h *SentimentExtrasHandler) GetMyFXBookPositions(ctx context.Context, req *connect.Request[v1.GetMyFXBookPositionsRequest]) (*connect.Response[v1.GetMyFXBookPositionsResponse], error) {
	symbol := strings.ToUpper(strings.TrimSpace(req.Msg.Symbol))
	if symbol == "" {
		symbol = "EURUSD"
	}
	resp := &v1.GetMyFXBookPositionsResponse{Symbol: symbol}
	if !h.fc.IsAvailable() {
		return connect.NewResponse(resp), nil // 未配置 Key 优雅降级
	}
	snap, err := h.fc.FetchMyFXBook(ctx)
	if err != nil {
		return connect.NewResponse(resp), nil // 抽取失败不阻塞前端
	}
	for _, p := range snap.Pairs {
		if p.Symbol == symbol {
			resp.LongPct = p.LongPct
			resp.ShortPct = p.ShortPct
			break
		}
	}
	return connect.NewResponse(resp), nil
}

// GetInsiderTrades 通过 firecrawl 抓取 OpenInsider 公开数据；未配置 Key 时优雅降级返回空。
func (h *SentimentExtrasHandler) GetInsiderTrades(ctx context.Context, req *connect.Request[v1.GetInsiderTradesRequest]) (*connect.Response[v1.GetInsiderTradesResponse], error) {
	resp := &v1.GetInsiderTradesResponse{}
	if !h.fc.IsAvailable() {
		return connect.NewResponse(resp), nil
	}
	limit := int(req.Msg.Limit)
	trades, err := h.fc.FetchOpenInsiderTrades(ctx, req.Msg.Ticker, limit)
	if err != nil {
		return connect.NewResponse(resp), nil
	}
	for _, t := range trades {
		resp.Items = append(resp.Items, &v1.InsiderTrade{
			Ticker:   t.Ticker,
			Insider:  t.Insider,
			Title:    t.Title,
			Action:   t.Action,
			Date:     timestamppb.New(t.Date),
			Price:    t.Price,
			Quantity: t.Quantity,
		})
	}
	return connect.NewResponse(resp), nil
}

// GetCryptoSocial 通过 firecrawl 抓 LunarCrush 公开页面。
func (h *SentimentExtrasHandler) GetCryptoSocial(ctx context.Context, req *connect.Request[v1.GetCryptoSocialRequest]) (*connect.Response[v1.GetCryptoSocialResponse], error) {
	resp := &v1.GetCryptoSocialResponse{}
	if !h.fc.IsAvailable() {
		return connect.NewResponse(resp), nil
	}
	snap, err := h.fc.FetchCryptoSocial(ctx, req.Msg.Asset)
	if err != nil {
		return connect.NewResponse(resp), nil
	}
	resp.Date = timestamppb.New(snap.Date)
	resp.TwitterFollowersGrowth = snap.TwitterFollowersGrowth
	resp.RedditSubscribersGrowth = snap.RedditSubscribersGrowth
	resp.SentimentScore = snap.SentimentScore
	return connect.NewResponse(resp), nil
}

func (h *SentimentExtrasHandler) GetFinvizMetrics(ctx context.Context, req *connect.Request[v1.GetFinvizMetricsRequest]) (*connect.Response[v1.GetFinvizMetricsResponse], error) {
	resp := &v1.GetFinvizMetricsResponse{}
	if !h.fc.IsAvailable() {
		return connect.NewResponse(resp), nil
	}
	q, err := h.fc.FetchFinvizQuote(ctx, req.Msg.Ticker)
	if err != nil {
		return connect.NewResponse(resp), nil
	}
	resp.ShortRatio = q.ShortRatio
	resp.ShortPctFloat = q.ShortPctFloat
	resp.InstOwnPct = q.InstOwnPct
	return connect.NewResponse(resp), nil
}

var _ antclawv1connect.SentimentExtrasServiceHandler = (*SentimentExtrasHandler)(nil)
