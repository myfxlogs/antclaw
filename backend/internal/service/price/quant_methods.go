package price

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	pricev1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/internal/service/quant"
)

// getRegimeHMM 用对数收益拟合 GaussianHMM 并输出当前状态。
//
// 状态命名：按各状态高斯均值排序，最低 -> "bear" / "risk_off"，
// 最高 -> "bull" / "risk_on"，中间 -> "neutral"。
// confidence = 当前状态的后验概率（>=0.5 才有信息量）。
func (s *Service) getRegimeHMM(ctx context.Context, pair, timeframe string, nStates int) (*pricev1.GetRegimeResponse, error) {
	if nStates < 2 {
		nStates = 2
	}
	// 至少需要 nStates*30 + 50 根用于 GARCH/HMM 联合容差。
	bars, err := s.loadBars(ctx, pair, timeframe, 600)
	if err != nil {
		return nil, err
	}
	if len(bars) < nStates*30 {
		return nil, fmt.Errorf("price regime hmm: need >= %d bars, got %d", nStates*30, len(bars))
	}
	closes := make([]float64, len(bars))
	for i, b := range bars {
		closes[i] = b.Close
	}
	rets := quant.LogReturns(closes)
	if len(rets) == 0 {
		return nil, fmt.Errorf("price regime hmm: empty returns")
	}
	// seed 用样本数派生，保证可复现且非全零。
	seed := int64(len(rets))
	model, err := quant.FitGaussianHMM(rets, nStates, seed, 200)
	if err != nil {
		return nil, fmt.Errorf("price regime hmm fit: %w", err)
	}
	post, err := model.Posterior(rets)
	if err != nil {
		return nil, fmt.Errorf("price regime hmm posterior: %w", err)
	}
	// 找最大后验状态
	curState, conf := 0, 0.0
	for i, p := range post {
		if p > conf {
			conf = p
			curState = i
		}
	}
	// 状态命名：按 mu 升序映射。
	order := make([]int, nStates)
	for i := 0; i < nStates; i++ {
		order[i] = i
	}
	for i := 0; i < nStates; i++ {
		for j := i + 1; j < nStates; j++ {
			if model.Mu[order[j]] < model.Mu[order[i]] {
				order[i], order[j] = order[j], order[i]
			}
		}
	}
	rank := -1
	for i, idx := range order {
		if idx == curState {
			rank = i
			break
		}
	}
	regime := stateLabel(rank, nStates)

	// 最近 5 期状态序列（去重保序）。
	tail := rets
	if len(tail) > 100 {
		tail = tail[len(tail)-100:]
	}
	path, _ := model.Decode(tail)
	recent := make([]string, 0, 5)
	seen := map[int]bool{}
	for i := len(path) - 1; i >= 0 && len(recent) < 5; i-- {
		if seen[path[i]] {
			continue
		}
		seen[path[i]] = true
		// 找其在 order 中的排名
		r := -1
		for j, idx := range order {
			if idx == path[i] {
				r = j
				break
			}
		}
		recent = append(recent, stateLabel(r, nStates))
	}
	return &pricev1.GetRegimeResponse{
		Regime: &pricev1.MarketRegime{
			Regime:     regime,
			Confidence: conf,
			Since:      time.Now().Add(-time.Duration(len(rets)/4) * time.Hour).Format(time.RFC3339),
		},
		RecentRegimes: recent,
		EngineUsed:    "hmm",
	}, nil
}

func stateLabel(rank, nStates int) string {
	switch nStates {
	case 2:
		if rank == 0 {
			return "risk_off"
		}
		return "risk_on"
	case 3:
		switch rank {
		case 0:
			return "bear"
		case 1:
			return "neutral"
		default:
			return "bull"
		}
	}
	return fmt.Sprintf("state_%d", rank)
}

// GetVolatility GARCH(1,1) 拟合 + 1 步预测；返回年化标准差序列。
func (s *Service) GetVolatility(ctx context.Context, pair, timeframe string, lookback int32) (*pricev1.GetVolatilityResponse, error) {
	n := int(lookback)
	if n <= 0 {
		n = 500
	}
	if n > 5000 {
		n = 5000
	}
	bars, err := s.loadBars(ctx, pair, timeframe, n)
	if err != nil {
		return nil, err
	}
	if len(bars) < 80 {
		return nil, fmt.Errorf("price volatility: need >= 80 bars, got %d", len(bars))
	}
	closes := make([]float64, len(bars))
	for i, b := range bars {
		closes[i] = b.Close
	}
	rets := quant.LogReturns(closes)
	params, condVar, err := quant.FitGARCH(rets)
	if err != nil {
		return nil, fmt.Errorf("price volatility GARCH: %w", err)
	}
	annu := quant.AnnualizationFactor(timeframe)
	scale := func(varPerBar float64) float64 {
		if varPerBar < 0 {
			return 0
		}
		return sqrt(varPerBar) * sqrt(annu)
	}
	// 序列时间戳与 condVar 长度对齐到 rets 长度（rets 比 bars 少 1）。
	out := make([]*pricev1.VolatilityPoint, 0, len(condVar))
	offset := len(bars) - len(condVar)
	for i, v := range condVar {
		ts := bars[offset+i].Timestamp.UTC().Format(time.RFC3339)
		out = append(out, &pricev1.VolatilityPoint{
			Timestamp:      ts,
			ConditionalVol: scale(v),
		})
	}
	last := condVar[len(condVar)-1]
	lastRet := rets[len(rets)-1]
	next := quant.ForecastGARCH(params, lastRet, last)
	uncondVar := params.UnconditionalVar()
	return &pricev1.GetVolatilityResponse{
		Pair:              pair,
		Omega:             params.Omega,
		Alpha:             params.Alpha,
		Beta:              params.Beta,
		Persistence:       params.Persistence(),
		UnconditionalVol:  scale(uncondVar),
		NextStepForecast:  scale(next),
		Series:            out,
	}, nil
}

// GetHurst 用对数收益估计 Hurst 指数。
func (s *Service) GetHurst(ctx context.Context, pair, timeframe string, lookback int32) (*pricev1.GetHurstResponse, error) {
	n := int(lookback)
	if n <= 0 {
		n = 500
	}
	if n > 5000 {
		n = 5000
	}
	bars, err := s.loadBars(ctx, pair, timeframe, n)
	if err != nil {
		return nil, err
	}
	if len(bars) < 80 {
		return nil, fmt.Errorf("price hurst: need >= 80 bars, got %d", len(bars))
	}
	closes := make([]float64, len(bars))
	for i, b := range bars {
		closes[i] = b.Close
	}
	rets := quant.LogReturns(closes)
	res, err := quant.HurstRS(rets)
	if err != nil {
		return nil, fmt.Errorf("price hurst: %w", err)
	}
	return &pricev1.GetHurstResponse{
		Pair:           pair,
		Hurst:          res.H,
		Interpretation: res.Interpretation,
		SampleSize:     int32(res.SampleSize),
	}, nil
}

// GetCorrelations 滚动相关矩阵。assets 为空时使用默认 8 主流货币对。
func (s *Service) GetCorrelations(ctx context.Context, assets []string, timeframe string, window int32) (*pricev1.GetCorrelationsResponse, error) {
	if len(assets) == 0 {
		assets = []string{"EURUSD", "GBPUSD", "USDJPY", "USDCHF", "AUDUSD", "USDCAD", "NZDUSD", "XAUUSD"}
	}
	w := int(window)
	if w <= 0 {
		w = 30
	}
	// 拉取每个资产至少 w+5 根 K 线的收益序列。
	series := make([][]float64, 0, len(assets))
	usable := make([]string, 0, len(assets))
	for _, a := range assets {
		bars, err := s.loadBars(ctx, a, timeframe, w+50)
		if err != nil || len(bars) < w+1 {
			continue
		}
		closes := make([]float64, len(bars))
		for i, b := range bars {
			closes[i] = b.Close
		}
		series = append(series, quant.LogReturns(closes))
		usable = append(usable, a)
	}
	if len(usable) < 2 {
		return nil, fmt.Errorf("price corr: need >= 2 assets with data, got %d", len(usable))
	}
	matrix := quant.RollingCorrelationMatrix(series, w)
	cells := make([]*pricev1.CorrelationCell, 0, len(usable)*len(usable))
	for i, a := range usable {
		for j, b := range usable {
			cells = append(cells, &pricev1.CorrelationCell{AssetA: a, AssetB: b, Value: matrix[i][j]})
		}
	}
	return &pricev1.GetCorrelationsResponse{
		Assets: usable,
		Window: int32(w),
		Matrix: cells,
	}, nil
}

// GetDivergences 在指定 lookback 内检测 RSI / OBV / MACD 三类背离。
func (s *Service) GetDivergences(ctx context.Context, pair, timeframe string, lookback int32, indicators []string) (*pricev1.GetDivergencesResponse, error) {
	n := int(lookback)
	if n <= 0 {
		n = 200
	}
	bars, err := s.loadBars(ctx, pair, timeframe, n+30)
	if err != nil {
		return nil, err
	}
	if len(bars) < 50 {
		return nil, fmt.Errorf("price divergences: need >= 50 bars, got %d", len(bars))
	}
	closes := make([]float64, len(bars))
	vols := make([]int64, len(bars))
	for i, b := range bars {
		closes[i] = b.Close
		vols[i] = b.Volume
	}
	if len(indicators) == 0 {
		indicators = []string{"rsi", "obv", "macd"}
	}
	want := map[string]bool{}
	for _, ind := range indicators {
		want[strings.ToLower(strings.TrimSpace(ind))] = true
	}
	var events []*pricev1.DivergenceEvent
	collect := func(ind string, series []float64) {
		ds := quant.FindDivergences(closes, series, ind, 5, n)
		for _, d := range ds {
			ts := bars[d.Index].Timestamp.UTC().Format(time.RFC3339)
			events = append(events, &pricev1.DivergenceEvent{
				Indicator:      ind,
				Kind:           string(d.Kind),
				DetectedAt:     ts,
				PricePivot:     d.PricePivot,
				IndicatorPivot: d.IndicatorPivot,
				Note:           d.Note,
			})
		}
	}
	if want["rsi"] {
		collect("rsi", quant.RSI(closes, 14))
	}
	if want["obv"] {
		collect("obv", quant.OBV(closes, vols))
	}
	if want["macd"] {
		collect("macd", quant.MACDLine(closes, 12, 26))
	}
	return &pricev1.GetDivergencesResponse{Pair: pair, Events: events}, nil
}

func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	return math.Sqrt(x)
}
