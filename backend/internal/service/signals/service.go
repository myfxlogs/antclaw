package signals

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"strings"
	"time"
)

var ErrDataInsufficient = errors.New("data insufficient")

type Deps struct {
	Price  PriceProvider
	COT    COTProvider
	Regime MacroRegimeProvider
	Factor FactorProvider
	Vol    VolProvider
	Flow   FlowProvider
	Signal SignalRepo
	Logger *slog.Logger
}

type Service struct{ deps Deps }

func NewService(d Deps) *Service { return &Service{deps: d} }

type BiasData struct {
	Pair       string
	Direction  string
	Confidence float64
	Timeframe  string
}

func (s *Service) GetBias(ctx context.Context, pair, tf string) (*BiasData, error) {
	code, ok := ResolveCOTCode(pair)
	var cot *COTAnalysis
	if ok {
		cot, _ = s.deps.COT.GetLatestAnalysis(ctx, code)
	}
	regime, _ := s.deps.Regime.GetCurrent(ctx, strings.ToUpper(pair), tf)
	fb, _ := s.deps.Factor.GetSymbolFactors(ctx, strings.ToUpper(pair), time.Now())
	if cot == nil && regime == nil && fb == nil {
		return nil, fmt.Errorf("%w: missing cot/regime/factor for %s", ErrDataInsufficient, pair)
	}

	var score float64
	switch {
	case cot != nil && regime != nil && fb != nil:
		score = 0.4*cot.SentimentScore + 0.4*regime.UnifiedScore + 0.2*tanh(fb.Momentum)
	case cot != nil && regime != nil:
		score = 0.5*cot.SentimentScore + 0.5*regime.UnifiedScore
	case regime != nil && fb != nil:
		score = 0.7*regime.UnifiedScore + 0.3*tanh(fb.Momentum)
	case cot != nil && fb != nil:
		score = 0.7*cot.SentimentScore + 0.3*tanh(fb.Momentum)
	case regime != nil:
		score = regime.UnifiedScore
	case cot != nil:
		score = cot.SentimentScore
	default:
		score = tanh(fb.Momentum)
	}

	regimeUncertainty := 0.5
	if regime != nil {
		regimeUncertainty = 1 - clamp(regime.HMMConfidence, 0, 1)
	}
	conf := clamp(math.Abs(score)*(1-regimeUncertainty), 0.05, 0.95)
	return &BiasData{Pair: strings.ToUpper(pair), Direction: directionLabel(score), Confidence: conf, Timeframe: tf}, nil
}

func (s *Service) GetRank(ctx context.Context, category string) ([]RankItem, error) {
	ranking, err := s.deps.Factor.GetRanking(ctx, category, time.Now())
	if err != nil {
		return nil, err
	}
	if ranking == nil || len(ranking.Items) < 3 {
		return nil, fmt.Errorf("%w: category=%s has <3 items", ErrDataInsufficient, category)
	}
	items := append([]RankItem(nil), ranking.Items...)
	sort.Slice(items, func(i, j int) bool { return items[i].NormScore > items[j].NormScore })
	for i := range items {
		items[i].Rank = i + 1
	}
	return items, nil
}

type XFactorView struct {
	Name        string
	Weight      float64
	Direction   string
	Description string
	Value       float64
}

func (s *Service) GetXFactors(ctx context.Context, pair string) ([]XFactorView, string, error) {
	fb, err := s.deps.Factor.GetSymbolFactors(ctx, strings.ToUpper(pair), time.Now())
	if err != nil {
		return nil, "", err
	}
	if fb == nil {
		return nil, "", fmt.Errorf("%w: no factor breakdown for %s", ErrDataInsufficient, pair)
	}
	weights, err := s.deps.Signal.GetActiveWeights(ctx)
	if err != nil {
		return nil, "", err
	}
	values := []struct {
		name string
		val  float64
	}{
		{"Momentum", fb.Momentum}, {"LowVol", fb.LowVol}, {"Trend", fb.Trend},
		{"Carry", fb.Carry}, {"Crowding", fb.Crowding}, {"Residual", fb.Residual},
	}
	out := make([]XFactorView, 0, 6)
	var sum, sumW float64
	for _, v := range values {
		dir := "negative"
		if v.val >= 0 {
			dir = "positive"
		}
		w := weights[strings.ToLower(v.name)]
		sum += w * v.val
		sumW += w
		out = append(out, XFactorView{Name: v.name, Weight: w, Direction: dir, Description: factorDescription(v.name, v.val), Value: v.val})
	}
	if sumW == 0 {
		return out, "neutral", nil
	}
	composite := sum / sumW
	switch {
	case composite > 0.3:
		return out, "long", nil
	case composite < -0.3:
		return out, "short", nil
	default:
		return out, "neutral", nil
	}
}

type RadarPoint struct {
	Pair     string
	X        float64
	Y        float64
	Quadrant string
	Strength float64
}

func (s *Service) GetRadar(ctx context.Context, category string) ([]RadarPoint, error) {
	symbols := SymbolsByCategory(category)
	points := make([]RadarPoint, 0, len(symbols))
	for _, sym := range symbols {
		fb, _ := s.deps.Factor.GetSymbolFactors(ctx, sym, time.Now())
		regime, _ := s.deps.Regime.GetCurrent(ctx, sym, "1d")
		if fb == nil || regime == nil {
			continue
		}
		p := RadarPoint{
			Pair: sym, X: fb.Composite, Y: regime.UnifiedScore,
			Quadrant: classify(fb.Composite, regime.UnifiedScore),
			Strength: math.Abs(regime.UnifiedScore) * clamp(fb.Composite/100, 0, 1),
		}
		points = append(points, p)
	}
	if len(points) < 5 {
		return nil, fmt.Errorf("%w: category=%s has only %d points", ErrDataInsufficient, category, len(points))
	}
	return points, nil
}

func (s *Service) GetIntensity(ctx context.Context, pair string) (float64, string, error) {
	recent, err := s.deps.Signal.GetRecentBySymbol(ctx, strings.ToUpper(pair), 100)
	if err != nil {
		return 0, "", err
	}
	if len(recent) < 20 {
		return 0, "", fmt.Errorf("%w: need >=20 recent signals, got %d", ErrDataInsufficient, len(recent))
	}
	scores := make([]float64, 0, len(recent))
	for _, rec := range recent {
		scores = append(scores, math.Abs(rec.UnifiedScore))
	}
	current := scores[0]
	pct := percentile(current, scores)
	return current, intensityLabel(pct), nil
}

func (s *Service) GetUnified(ctx context.Context, pair string) (*UnifiedSignalRecord, error) {
	recent, err := s.deps.Signal.GetRecentBySymbol(ctx, strings.ToUpper(pair), 1)
	if err != nil {
		return nil, err
	}
	if len(recent) > 0 {
		return &recent[0], nil
	}
	return s.ComputeUnified(ctx, strings.ToUpper(pair))
}

func (s *Service) GetTransition(ctx context.Context, pair, currentState string) ([]RegimeTransition, error) {
	history, err := s.deps.Regime.GetHistory(ctx, strings.ToUpper(pair), "1d", 730)
	if err != nil {
		return nil, err
	}
	if len(history) < 60 {
		return nil, fmt.Errorf("%w: not enough history for %s", ErrDataInsufficient, pair)
	}
	cur := strings.ToUpper(strings.TrimSpace(currentState))
	if cur == "" {
		cur = history[0].UnifiedLabel
	}
	out := transitionProbabilities(history, cur)
	if len(out) == 0 {
		return nil, fmt.Errorf("%w: no transition probabilities for %s", ErrDataInsufficient, pair)
	}
	return out, nil
}

type CryptoAlphaSignal struct {
	Asset      string
	SignalType string
	Confidence float64
	Timeframe  string
}

func (s *Service) GetCryptoAlpha(ctx context.Context, assetFilter string) ([]CryptoAlphaSignal, error) {
	assets := []string{"BTCUSDT", "ETHUSDT", "SOLUSDT"}
	out := make([]CryptoAlphaSignal, 0, len(assets))
	for _, a := range assets {
		if assetFilter != "" && !strings.Contains(strings.ToUpper(a), strings.ToUpper(assetFilter)) {
			continue
		}
		u, err := s.GetUnified(ctx, a)
		if err != nil {
			continue
		}
		signalType := "neutral"
		if u.UnifiedScore > 0.2 {
			signalType = "accumulation"
		} else if u.UnifiedScore < -0.2 {
			signalType = "distribution"
		}
		out = append(out, CryptoAlphaSignal{Asset: a, SignalType: signalType, Confidence: u.Confidence, Timeframe: "1d"})
	}
	return out, nil
}

type QuantSignal struct {
	Pair     string
	Strategy string
	Signal   string
	Sharpe   float64
	Drawdown float64
}

func (s *Service) GetQuant(ctx context.Context, pair string) ([]QuantSignal, error) {
	u, err := s.GetUnified(ctx, pair)
	if err != nil {
		return nil, err
	}
	sig := "flat"
	if u.UnifiedScore > 0.2 {
		sig = "long"
	} else if u.UnifiedScore < -0.2 {
		sig = "short"
	}
	return []QuantSignal{{Pair: pair, Strategy: "unified_momentum", Signal: sig, Sharpe: u.Confidence * 2, Drawdown: 1 - u.Confidence}}, nil
}

type CtaSignal struct {
	Pair     string
	Trend    string
	Momentum float64
	Regime   string
}

func (s *Service) GetCta(ctx context.Context, pair string) (*CtaSignal, error) {
	regime, err := s.deps.Regime.GetCurrent(ctx, strings.ToUpper(pair), "1d")
	if err != nil {
		return nil, err
	}
	fb, err := s.deps.Factor.GetSymbolFactors(ctx, strings.ToUpper(pair), time.Now())
	if err != nil {
		return nil, err
	}
	trend := "neutral"
	if regime.UnifiedScore > 0.2 {
		trend = "bullish"
	} else if regime.UnifiedScore < -0.2 {
		trend = "bearish"
	}
	return &CtaSignal{Pair: strings.ToUpper(pair), Trend: trend, Momentum: fb.Momentum, Regime: regime.UnifiedLabel}, nil
}
