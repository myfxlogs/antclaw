package rpc

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	signalsv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
	"github.com/antclaw/antclaw/internal/service/calibration"
	signalssvc "github.com/antclaw/antclaw/internal/service/signals"
)

type SignalsHandler struct {
	svc   *signalssvc.Service
	calib *calibration.Store // M-C: 校准模型存储；nil 时校准 RPC 返错。
}

func NewSignalsHandler(svc *signalssvc.Service) *SignalsHandler {
	return &SignalsHandler{svc: svc}
}

// WithCalibration 注入校准模型存储。
func (h *SignalsHandler) WithCalibration(s *calibration.Store) *SignalsHandler {
	h.calib = s
	return h
}

func asConnectErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, signalssvc.ErrDataInsufficient) {
		return connect.NewError(connect.CodeFailedPrecondition, err)
	}
	return connect.NewError(connect.CodeInternal, err)
}

func (h *SignalsHandler) GetBias(ctx context.Context, req *connect.Request[signalsv1.GetBiasRequest]) (*connect.Response[signalsv1.GetBiasResponse], error) {
	bias, err := h.svc.GetBias(ctx, req.Msg.Pair, req.Msg.Timeframe)
	if err != nil {
		return nil, asConnectErr(err)
	}
	return connect.NewResponse(&signalsv1.GetBiasResponse{Biases: []*signalsv1.BiasData{{
		Pair: bias.Pair, Direction: bias.Direction, Confidence: bias.Confidence, Timeframe: bias.Timeframe,
	}}}), nil
}

func (h *SignalsHandler) GetRank(ctx context.Context, req *connect.Request[signalsv1.GetRankRequest]) (*connect.Response[signalsv1.GetRankResponse], error) {
	items, err := h.svc.GetRank(ctx, req.Msg.Category)
	if err != nil {
		return nil, asConnectErr(err)
	}
	out := make([]*signalsv1.RankItem, 0, len(items))
	for _, it := range items {
		out = append(out, &signalsv1.RankItem{Pair: it.Symbol, Rank: int32(it.Rank), Score: it.NormScore, Trend: it.Trend})
	}
	return connect.NewResponse(&signalsv1.GetRankResponse{Rankings: out}), nil
}

func (h *SignalsHandler) GetXFactors(ctx context.Context, req *connect.Request[signalsv1.GetXFactorsRequest]) (*connect.Response[signalsv1.GetXFactorsResponse], error) {
	factors, signal, err := h.svc.GetXFactors(ctx, req.Msg.Pair)
	if err != nil {
		return nil, asConnectErr(err)
	}
	out := make([]*signalsv1.XFactor, 0, len(factors))
	for _, f := range factors {
		out = append(out, &signalsv1.XFactor{Name: f.Name, Weight: f.Weight, Direction: f.Direction, Description: f.Description})
	}
	return connect.NewResponse(&signalsv1.GetXFactorsResponse{Pair: req.Msg.Pair, Factors: out, CompositeSignal: signal}), nil
}

func (h *SignalsHandler) GetRadar(ctx context.Context, req *connect.Request[signalsv1.GetRadarRequest]) (*connect.Response[signalsv1.GetRadarResponse], error) {
	points, err := h.svc.GetRadar(ctx, req.Msg.Category)
	if err != nil {
		return nil, asConnectErr(err)
	}
	out := make([]*signalsv1.RadarDataPoint, 0, len(points))
	for _, p := range points {
		out = append(out, &signalsv1.RadarDataPoint{Pair: p.Pair, X: p.X, Y: p.Y, Quadrant: p.Quadrant, Strength: p.Strength})
	}
	return connect.NewResponse(&signalsv1.GetRadarResponse{Points: out}), nil
}

func (h *SignalsHandler) GetIntensity(ctx context.Context, req *connect.Request[signalsv1.GetIntensityRequest]) (*connect.Response[signalsv1.GetIntensityResponse], error) {
	value, label, err := h.svc.GetIntensity(ctx, req.Msg.Pair)
	if err != nil {
		return nil, asConnectErr(err)
	}
	return connect.NewResponse(&signalsv1.GetIntensityResponse{
		Intensity: &signalsv1.IntensityData{Pair: req.Msg.Pair, Intensity: value, StrengthLabel: label, Percentile_30D: value * 100},
	}), nil
}

func (h *SignalsHandler) GetUnified(ctx context.Context, req *connect.Request[signalsv1.GetUnifiedRequest]) (*connect.Response[signalsv1.GetUnifiedResponse], error) {
	sig, err := h.svc.GetUnified(ctx, req.Msg.Pair)
	if err != nil {
		return nil, asConnectErr(err)
	}
	return connect.NewResponse(&signalsv1.GetUnifiedResponse{
		Signal: &signalsv1.UnifiedSignal{
			Pair:                sig.Symbol,
			Direction:           signalssvc.RecommendationToDirection(sig.Recommendation),
			Confidence:          sig.Confidence,
			ContributingFactors: signalssvc.Top3Components(sig.Components),
		},
	}), nil
}

func (h *SignalsHandler) GetTransition(ctx context.Context, req *connect.Request[signalsv1.GetTransitionRequest]) (*connect.Response[signalsv1.GetTransitionResponse], error) {
	rows, err := h.svc.GetTransition(ctx, req.Msg.Pair, req.Msg.CurrentState)
	if err != nil {
		return nil, asConnectErr(err)
	}
	out := make([]*signalsv1.TransitionProb, 0, len(rows))
	for _, r := range rows {
		out = append(out, &signalsv1.TransitionProb{FromState: r.FromLabel, ToState: r.ToLabel, Probability: r.ToScore})
	}
	return connect.NewResponse(&signalsv1.GetTransitionResponse{Pair: req.Msg.Pair, Transitions: out}), nil
}
func (h *SignalsHandler) GetCryptoAlpha(ctx context.Context, req *connect.Request[signalsv1.GetCryptoAlphaRequest]) (*connect.Response[signalsv1.GetCryptoAlphaResponse], error) {
	rows, err := h.svc.GetCryptoAlpha(ctx, req.Msg.AssetFilter)
	if err != nil {
		return nil, asConnectErr(err)
	}
	out := make([]*signalsv1.CryptoAlphaSignal, 0, len(rows))
	for _, r := range rows {
		out = append(out, &signalsv1.CryptoAlphaSignal{Asset: r.Asset, SignalType: r.SignalType, Confidence: r.Confidence, Timeframe: r.Timeframe})
	}
	return connect.NewResponse(&signalsv1.GetCryptoAlphaResponse{Signals: out}), nil
}
func (h *SignalsHandler) GetQuant(ctx context.Context, req *connect.Request[signalsv1.GetQuantRequest]) (*connect.Response[signalsv1.GetQuantResponse], error) {
	rows, err := h.svc.GetQuant(ctx, req.Msg.Pair)
	if err != nil {
		return nil, asConnectErr(err)
	}
	out := make([]*signalsv1.QuantSignal, 0, len(rows))
	for _, r := range rows {
		out = append(out, &signalsv1.QuantSignal{Pair: r.Pair, Strategy: r.Strategy, Signal: r.Signal, Sharpe: r.Sharpe, Drawdown: r.Drawdown})
	}
	return connect.NewResponse(&signalsv1.GetQuantResponse{Signals: out}), nil
}
func (h *SignalsHandler) GetCta(ctx context.Context, req *connect.Request[signalsv1.GetCtaRequest]) (*connect.Response[signalsv1.GetCtaResponse], error) {
	sig, err := h.svc.GetCta(ctx, req.Msg.Pair)
	if err != nil {
		return nil, asConnectErr(err)
	}
	return connect.NewResponse(&signalsv1.GetCtaResponse{Signal: &signalsv1.CtaSignal{Pair: sig.Pair, Trend: sig.Trend, Momentum: sig.Momentum, Regime: sig.Regime}}), nil
}
func (h *SignalsHandler) GetBriefing(ctx context.Context, req *connect.Request[signalsv1.GetBriefingRequest]) (*connect.Response[signalsv1.GetBriefingResponse], error) {
	sections := []*signalsv1.BriefingSection{}
	if bias, err := h.svc.GetBias(ctx, "EURUSD", "1d"); err == nil {
		sections = append(sections, &signalsv1.BriefingSection{
			Title:    "方向偏好",
			Content:  "EURUSD 当前方向为 " + bias.Direction,
			Priority: "high",
		})
	}
	if points, err := h.svc.GetRadar(ctx, "majors"); err == nil && len(points) > 0 {
		sections = append(sections, &signalsv1.BriefingSection{
			Title:    "市场结构",
			Content:  "主要货币对的因子-状态分布已更新，建议关注高象限标的。",
			Priority: "medium",
		})
	}
	if len(sections) == 0 {
		sections = append(sections, &signalsv1.BriefingSection{
			Title:    "系统提示",
			Content:  "当前可用数据不足，建议等待下一轮分析任务完成。",
			Priority: "low",
		})
	}
	return connect.NewResponse(&signalsv1.GetBriefingResponse{Sections: sections, GeneratedAt: "auto"}), nil
}
func (h *SignalsHandler) GetOutlook(ctx context.Context, req *connect.Request[signalsv1.GetOutlookRequest]) (*connect.Response[signalsv1.GetOutlookResponse], error) {
	sig, err := h.svc.GetUnified(ctx, req.Msg.Pair)
	if err != nil {
		return nil, asConnectErr(err)
	}
	horizon := req.Msg.Horizon
	if horizon == "" {
		horizon = "medium"
	}
	outlooks := []*signalsv1.OutlookData{{
		Pair: req.Msg.Pair, Outlook: signalssvc.RecommendationToDirection(sig.Recommendation), Confidence: sig.Confidence, Horizon: horizon,
	}}
	if horizon == "medium" || horizon == "long" {
		conf := sig.Confidence * 0.9
		if horizon == "long" {
			conf = sig.Confidence * 0.8
		}
		outlooks = append(outlooks, &signalsv1.OutlookData{
			Pair: req.Msg.Pair, Outlook: signalssvc.RecommendationToDirection(sig.Recommendation), Confidence: conf, Horizon: "long",
		})
	}
	return connect.NewResponse(&signalsv1.GetOutlookResponse{Outlooks: outlooks}), nil
}

// =====================================================
// M-C 校准 RPC
// =====================================================

func (h *SignalsHandler) FitCalibration(ctx context.Context, req *connect.Request[signalsv1.FitCalibrationRequest]) (*connect.Response[signalsv1.FitCalibrationResponse], error) {
	if h.calib == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("calibration store not initialized"))
	}
	if req.Msg.ModelId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("model_id required"))
	}
	if len(req.Msg.Scores) != len(req.Msg.Outcomes) || len(req.Msg.Scores) < 5 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("need >=5 paired scores/outcomes"))
	}
	var c calibration.Calibrator
	switch req.Msg.Type {
	case "platt":
		c = calibration.NewPlatt()
	case "isotonic":
		c = calibration.NewIsotonic()
	default:
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("type must be platt or isotonic"))
	}
	if err := c.Fit(req.Msg.Scores, req.Msg.Outcomes); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	br := c.Brier(req.Msg.Scores, req.Msg.Outcomes)
	if err := h.calib.Save(ctx, req.Msg.ModelId, c, len(req.Msg.Scores), br); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&signalsv1.FitCalibrationResponse{
		ModelId: req.Msg.ModelId, Type: c.Type(),
		NSamples: int32(len(req.Msg.Scores)), Brier: br,
	}), nil
}

func (h *SignalsHandler) PredictCalibrated(ctx context.Context, req *connect.Request[signalsv1.PredictCalibratedRequest]) (*connect.Response[signalsv1.PredictCalibratedResponse], error) {
	if h.calib == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("calibration store not initialized"))
	}
	c, err := h.calib.Load(ctx, req.Msg.ModelId)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(&signalsv1.PredictCalibratedResponse{
		ModelId: req.Msg.ModelId, Calibrated: c.Predict(req.Msg.Score),
	}), nil
}

func (h *SignalsHandler) ListCalibrations(ctx context.Context, req *connect.Request[signalsv1.ListCalibrationsRequest]) (*connect.Response[signalsv1.ListCalibrationsResponse], error) {
	if h.calib == nil {
		return connect.NewResponse(&signalsv1.ListCalibrationsResponse{}), nil
	}
	rows, err := h.calib.List(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	out := &signalsv1.ListCalibrationsResponse{}
	for _, s := range rows {
		out.Items = append(out.Items, &signalsv1.CalibrationSummary{
			ModelId: s.ModelID, Type: s.Type, NSamples: int32(s.NSamples),
			Brier: s.Brier, FittedAt: s.FittedAt.Format("2006-01-02T15:04:05Z"),
		})
	}
	return connect.NewResponse(out), nil
}

var _ antclawv1connect.SignalsServiceHandler = (*SignalsHandler)(nil)
