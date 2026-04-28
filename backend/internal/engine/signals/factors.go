package signals

import "time"

// FactorValue represents a calculated signal factor.
type FactorValue struct {
	Name       string
	Value      float64 // Normalized -1 to 1
	Weight     float64
	Confidence float64
}

// MarketData holds price and volatility data for signal calculation.
type MarketData struct {
	Pair          string
	CurrentPrice  float64
	Change24h     float64
	ChangePct24h  float64
	Bars          []PriceBar
	VIX           float64
	ATR           float64 // Average True Range
	Timestamp     time.Time
}

// PriceBar represents OHLCV data.
type PriceBar struct {
	Timestamp time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    int64
}

// DataProvider interface for market data access.
type DataProvider interface {
	GetMarketData(pair string) (*MarketData, error)
	GetPairsByCategory(category string) []string
}

// SignalResult represents the output of signal calculation.
type SignalResult struct {
	Direction   string // bullish, bearish, neutral
	Confidence  float64
	Strength    float64 // 0-1 scale
	Factors     []FactorValue
	Timestamp   time.Time
}

// RadarPoint represents a point in the radar view.
type RadarPoint struct {
	Pair      string
	X         float64 // Trend strength: -1 to 1
	Y         float64 // Momentum: -1 to 1
	Quadrant  string  // bull, bear, ranging, reversal
	Strength  float64
}

// BriefingSection represents a section in market briefing.
type BriefingSection struct {
	Title     string
	Content   string
	Priority  string // high, medium, low
	Pairs     []string
}

// CalculateTrendFactor calculates trend direction and strength from price bars.
func CalculateTrendFactor(bars []PriceBar) FactorValue {
	if len(bars) < 10 {
		return FactorValue{Name: "trend", Value: 0, Weight: 0.4, Confidence: 0.3}
	}
	
	// Simple linear regression on closes
	n := float64(len(bars))
	sumX, sumY, sumXY, sumX2 := 0.0, 0.0, 0.0, 0.0
	
	for i, bar := range bars {
		x := float64(i)
		y := bar.Close
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}
	
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	
	// Normalize slope to -1 to 1 range (assuming max 2% change per bar)
	normalized := slope / (bars[len(bars)-1].Close * 0.02)
	if normalized > 1 {
		normalized = 1
	} else if normalized < -1 {
		normalized = -1
	}
	
	confidence := 0.5 + 0.5*min(float64(len(bars))/50.0, 1.0)
	
	return FactorValue{
		Name:       "trend",
		Value:      normalized,
		Weight:     0.4,
		Confidence: confidence,
	}
}

// CalculateMomentumFactor calculates momentum from recent price changes.
func CalculateMomentumFactor(bars []PriceBar) FactorValue {
	if len(bars) < 5 {
		return FactorValue{Name: "momentum", Value: 0, Weight: 0.35, Confidence: 0.3}
	}
	
	// Rate of change over last 5 bars
	recent := bars[len(bars)-1].Close
	previous := bars[len(bars)-5].Close
	
	roc := (recent - previous) / previous
	
	// Normalize to -1 to 1 (assuming max 5% move)
	normalized := roc / 0.05
	if normalized > 1 {
		normalized = 1
	} else if normalized < -1 {
		normalized = -1
	}
	
	return FactorValue{
		Name:       "momentum",
		Value:      normalized,
		Weight:     0.35,
		Confidence: 0.7,
	}
}

// CalculateVolatilityFactor calculates volatility regime factor.
func CalculateVolatilityFactor(vix float64) FactorValue {
	// VIX interpretation: low < 15, normal 15-25, high > 25
	var normalized float64
	switch {
	case vix < 15:
		normalized = 0.3 // Low vol, slightly bullish for carry
	case vix < 25:
		normalized = 0 // Neutral
	case vix < 35:
		normalized = -0.3 // Elevated vol, caution
	default:
		normalized = -0.6 // Extreme vol, risk-off
	}
	
	confidence := 0.6 + 0.3*(min(vix/50.0, 1.0))
	
	return FactorValue{
		Name:       "volatility",
		Value:      normalized,
		Weight:     0.25,
		Confidence: confidence,
	}
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// CompositeSignal aggregates multiple factors into a unified signal.
func CompositeSignal(factors []FactorValue) SignalResult {
	if len(factors) == 0 {
		return SignalResult{Direction: "neutral", Confidence: 0, Strength: 0, Factors: []FactorValue{}}
	}
	
	weightedSum := 0.0
	totalWeight := 0.0
	weightedConfidence := 0.0
	
	for _, f := range factors {
		weightedSum += f.Value * f.Weight
		totalWeight += f.Weight
		weightedConfidence += f.Confidence * f.Weight
	}
	
	if totalWeight == 0 {
		return SignalResult{Direction: "neutral", Confidence: 0, Strength: 0, Factors: factors}
	}
	
	composite := weightedSum / totalWeight
	confidence := weightedConfidence / totalWeight
	
	// Determine direction
	direction := "neutral"
	if composite > 0.2 {
		direction = "bullish"
	} else if composite < -0.2 {
		direction = "bearish"
	}
	
	// Strength is absolute value of composite, scaled to 0-1
	strength := composite
	if strength < 0 {
		strength = -strength
	}
	
	return SignalResult{
		Direction:   direction,
		Confidence:  confidence,
		Strength:    strength,
		Factors:     factors,
		Timestamp:   time.Now(),
	}
}
