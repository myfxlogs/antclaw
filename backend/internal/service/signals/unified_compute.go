package signals

import (
	"context"
	"math"
	"sort"
	"time"
)

func (s *Service) ComputeUnified(ctx context.Context, pair string) (*UnifiedSignalRecord, error) {
	code, _ := ResolveCOTCode(pair)
	cot, _ := s.deps.COT.GetLatestAnalysis(ctx, code)
	regime, _ := s.deps.Regime.GetCurrent(ctx, pair, "1d")
	fb, _ := s.deps.Factor.GetSymbolFactors(ctx, pair, time.Now())
	weights, _ := s.deps.Signal.GetActiveWeights(ctx)
	flow, _ := s.deps.Flow.GetTopDivergent(ctx, time.Now().AddDate(0, 0, -7), 200)
	vix, _ := s.deps.Vol.GetVIX(ctx)

	components := map[string]float64{}
	if cot != nil {
		components["cot"] = cot.SentimentScore
	}
	if regime != nil {
		components["macro"] = regime.UnifiedScore
	}
	if fb != nil {
		components["factor"] = tanh(fb.Composite/100*2 - 1)
	}
	components["flow"] = flowScore(flow, pair)
	components["vol"] = volScore(vix)
	components["season"] = seasonalScore(time.Now())

	var sum, sumW float64
	for k, v := range components {
		w := weights[k]
		if w == 0 {
			w = 1
		}
		sum += w * v
		sumW += w
	}
	if sumW == 0 {
		sumW = 1
	}
	score := clamp(sum/sumW, -1, 1)
	completeness := float64(len(components)) / 6.0
	confidence := sigmoid(math.Abs(score)*4) * completeness

	record := UnifiedSignalRecord{
		Symbol:         pair,
		IssuedAt:       time.Now().UTC(),
		Recommendation: unifiedRecommendation(score),
		UnifiedScore:   score,
		Confidence:     confidence,
		Components:     components,
		WeightsUsed:    weights,
	}
	id, err := s.deps.Signal.SaveUnified(ctx, record)
	if err != nil {
		return nil, err
	}
	record.ID = id
	return &record, nil
}

func Top3Components(components map[string]float64) []string {
	type kv struct {
		k string
		v float64
	}
	all := make([]kv, 0, len(components))
	for k, v := range components {
		all = append(all, kv{k: k, v: math.Abs(v)})
	}
	sort.Slice(all, func(i, j int) bool { return all[i].v > all[j].v })
	n := 3
	if len(all) < n {
		n = len(all)
	}
	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, all[i].k)
	}
	return out
}

func flowScore(rows []FlowDivergence, pair string) float64 {
	for _, r := range rows {
		if r.PairA == pair || r.PairB == pair {
			return clamp(-r.ZScore/3.0, -1, 1)
		}
	}
	return 0
}

func volScore(vix float64) float64 {
	switch {
	case vix > 35:
		return -0.8
	case vix > 25:
		return -0.4
	case vix > 18:
		return -0.1
	default:
		return 0.2
	}
}

func seasonalScore(now time.Time) float64 {
	month := int(now.Month())
	switch month {
	case 11, 12:
		return 0.2
	case 5, 9:
		return -0.1
	default:
		return 0
	}
}
