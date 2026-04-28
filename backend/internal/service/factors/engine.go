package factors

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/antclaw/antclaw/internal/infra/postgres"
	"github.com/antclaw/antclaw/internal/service/factors/score"
)

// Engine is the main factor ranking engine
type Engine struct {
	priceRepo postgres.PriceRepository
	cotRepo   postgres.COTRepository
	weights   FactorWeights
}

// FactorWeights defines factor weights
type FactorWeights struct {
	Momentum     float64
	TrendQuality float64
	Carry        float64
	LowVol       float64
	Residual     float64
	Crowding     float64 // Negative (penalty)
}

// DefaultWeights returns default factor weights
func DefaultWeights() FactorWeights {
	return FactorWeights{
		Momentum:     0.25,
		TrendQuality: 0.20,
		Carry:        0.15,
		LowVol:       0.15,
		Residual:     0.10,
		Crowding:     -0.15, // Penalty
	}
}

// AssetProfile contains data for a single asset
type AssetProfile struct {
	Symbol       string
	Currency     string
	Closes       []float64
	Highs        []float64
	Lows         []float64
	Returns      []float64
	CarryRate    float64
	COTIndex     float64
	SpecMomentum float64
}

// RankingResult contains factor ranking output
type RankingResult struct {
	Ranked      []AssetRank
	Top         []AssetRank
	Bottom      []AssetRank
	ComputedAt  time.Time
}

// AssetRank represents a single asset's ranking
type AssetRank struct {
	Symbol        string
	Currency      string
	RawScore      float64
	NormScore     float64
	Rank          int
	FactorBreakdown map[string]float64
}

// NewEngine creates a new factor engine
func NewEngine(priceRepo postgres.PriceRepository, cotRepo postgres.COTRepository) *Engine {
	return &Engine{
		priceRepo: priceRepo,
		cotRepo:   cotRepo,
		weights:   DefaultWeights(),
	}
}

// SetWeights updates factor weights
func (e *Engine) SetWeights(w FactorWeights) {
	e.weights = w
}

// Rank ranks all assets by factor scores
func (e *Engine) Rank(ctx context.Context, symbols []string) (*RankingResult, error) {
	// Build asset profiles
	profiles := make([]AssetProfile, 0, len(symbols))
	for _, symbol := range symbols {
		profile, err := e.buildProfile(ctx, symbol)
		if err != nil {
			continue // Skip failed profiles
		}
		profiles = append(profiles, profile)
	}

	if len(profiles) == 0 {
		return nil, fmt.Errorf("no valid asset profiles")
	}

	// Calculate raw scores for each asset
	rawScores := make(map[string]float64)
	breakdowns := make(map[string]map[string]float64)

	for _, profile := range profiles {
		scores := e.calculateScores(profile)
		rawScore := e.weightScores(scores)
		rawScores[profile.Symbol] = rawScore
		breakdowns[profile.Symbol] = scores
	}

	// Cross-sectional z-score normalization
	normScores := e.normalizeScores(rawScores)

	// Build ranked list
	var ranked []AssetRank
	for symbol, normScore := range normScores {
		profile := e.findProfile(profiles, symbol)
		ranked = append(ranked, AssetRank{
			Symbol:          symbol,
			Currency:        profile.Currency,
			RawScore:        rawScores[symbol],
			NormScore:       normScore,
			FactorBreakdown: breakdowns[symbol],
		})
	}

	// Sort by normalized score
	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].NormScore > ranked[j].NormScore
	})

	// Assign ranks and split top/bottom
	for i := range ranked {
		ranked[i].Rank = i + 1
	}

	// Top and bottom N (default 5)
	n := 5
	top := make([]AssetRank, 0, n)
	bottom := make([]AssetRank, 0, n)

	if len(ranked) >= n {
		top = ranked[:n]
		bottom = ranked[len(ranked)-n:]
	}

	return &RankingResult{
		Ranked:     ranked,
		Top:        top,
		Bottom:     bottom,
		ComputedAt: time.Now(),
	}, nil
}

// buildProfile builds asset profile from repositories
func (e *Engine) buildProfile(ctx context.Context, symbol string) (AssetProfile, error) {
	// Fetch price data
	bars, err := e.priceRepo.GetDailyBars(ctx, symbol, time.Now().AddDate(0, 0, -365), time.Now())
	if err != nil {
		return AssetProfile{}, err
	}

	if len(bars) < 240 {
		return AssetProfile{}, fmt.Errorf("insufficient data for %s", symbol)
	}

	profile := AssetProfile{
		Symbol:   symbol,
		Closes:   make([]float64, len(bars)),
		Highs:    make([]float64, len(bars)),
		Lows:     make([]float64, len(bars)),
		Returns:  make([]float64, len(bars)-1),
	}

	for i, bar := range bars {
		profile.Closes[i] = bar.Close
		profile.Highs[i] = bar.High
		profile.Lows[i] = bar.Low
		if i > 0 {
			profile.Returns[i-1] = (bar.Close - bars[i-1].Close) / bars[i-1].Close
		}
	}

	// Fetch COT data if available
	// In production, map symbol to contract code

	return profile, nil
}

// calculateScores calculates all factor scores for an asset
func (e *Engine) calculateScores(profile AssetProfile) map[string]float64 {
	scores := make(map[string]float64)

	// Momentum
	momentumScorer := score.NewMomentumScore()
	scores["momentum"] = momentumScorer.Calculate(profile.Closes)

	// Trend Quality
	trendScorer := score.NewTrendQualityScore()
	// Calculate SMAs
	sma20 := e.calculateSMA(profile.Closes, 20)
	sma50 := e.calculateSMA(profile.Closes, 50)
	sma200 := e.calculateSMA(profile.Closes, 200)

	scores["trend_quality"] = trendScorer.Calculate(score.TrendQualityInput{
		Closes:    profile.Closes,
		Highs:     profile.Highs,
		Lows:      profile.Lows,
		Period20:  sma20,
		Period50:  sma50,
		Period200: sma200,
	})

	// Carry
	carryScorer := score.NewCarryScore()
	scores["carry"] = carryScorer.Calculate(profile.CarryRate)

	// Low Vol
	volScorer := score.NewLowVolScore()
	scores["low_vol"] = volScorer.Calculate(profile.Returns)

	// Residual (requires market returns - simplified)
	// In production, use actual market benchmark
	scores["residual"] = 0

	// Crowding (penalty)
	crowdingScorer := score.NewCrowdingScore()
	scores["crowding"] = crowdingScorer.Calculate(score.CrowdingInput{
		COTIndex:       profile.COTIndex,
		SpecMomentum:   profile.SpecMomentum,
		CrowdingIndex:  0, // Would be calculated from COT data
	})

	return scores
}

// weightScores combines factor scores with weights
func (e *Engine) weightScores(scores map[string]float64) float64 {
	weights := map[string]float64{
		"momentum":      e.weights.Momentum,
		"trend_quality": e.weights.TrendQuality,
		"carry":         e.weights.Carry,
		"low_vol":       e.weights.LowVol,
		"residual":      e.weights.Residual,
		"crowding":      e.weights.Crowding,
	}

	var weightedSum float64
	for factor, score := range scores {
		if w, ok := weights[factor]; ok {
			weightedSum += score * w
		}
	}

	return weightedSum
}

// normalizeScores performs cross-sectional z-score normalization
func (e *Engine) normalizeScores(rawScores map[string]float64) map[string]float64 {
	// Calculate mean
	var sum float64
	for _, score := range rawScores {
		sum += score
	}
	mean := sum / float64(len(rawScores))

	// Calculate std dev
	var sumSquared float64
	for _, score := range rawScores {
		diff := score - mean
		sumSquared += diff * diff
	}
	stdDev := math.Sqrt(sumSquared / float64(len(rawScores)))

	// Calculate z-scores
	normalized := make(map[string]float64)
	for symbol, score := range rawScores {
		if stdDev > 0 {
			normalized[symbol] = (score - mean) / stdDev
		} else {
			normalized[symbol] = 0
		}
	}

	return normalized
}

// findProfile finds profile by symbol
func (e *Engine) findProfile(profiles []AssetProfile, symbol string) AssetProfile {
	for _, p := range profiles {
		if p.Symbol == symbol {
			return p
		}
	}
	return AssetProfile{Symbol: symbol}
}

// calculateSMA calculates Simple Moving Average
func (e *Engine) calculateSMA(closes []float64, period int) []float64 {
	if len(closes) < period {
		return nil
	}

	sma := make([]float64, len(closes)-period+1)
	for i := period - 1; i < len(closes); i++ {
		var sum float64
		for j := i - period + 1; j <= i; j++ {
			sum += closes[j]
		}
		sma[i-period+1] = sum / float64(period)
	}

	return sma
}
