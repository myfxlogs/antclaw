package signals

import (
	"fmt"
	"time"
)

// Engine is the core signal calculation engine.
type Engine struct {
	provider DataProvider
}

// NewEngine creates a new signal engine with the given data provider.
func NewEngine(provider DataProvider) *Engine {
	return &Engine{provider: provider}
}

// GetProvider returns the data provider.
func (e *Engine) GetProvider() DataProvider {
	return e.provider
}

// CalculateBias calculates directional bias for a pair.
func (e *Engine) CalculateBias(pair string, timeframe string) (*BiasResult, error) {
	data, err := e.provider.GetMarketData(pair)
	if err != nil {
		return nil, fmt.Errorf("failed to get market data: %w", err)
	}
	
	factors := []FactorValue{
		CalculateTrendFactor(data.Bars),
		CalculateMomentumFactor(data.Bars),
		CalculateVolatilityFactor(data.VIX),
	}
	
	signal := CompositeSignal(factors)
	
	return &BiasResult{
		Pair:       pair,
		Timeframe:  timeframe,
		Direction:  signal.Direction,
		Confidence: signal.Confidence,
		Strength:   signal.Strength,
		Factors:    factors,
		Timestamp:  time.Now(),
	}, nil
}

// CalculateUnified calculates unified signal for a pair.
func (e *Engine) CalculateUnified(pair string) (*UnifiedResult, error) {
	data, err := e.provider.GetMarketData(pair)
	if err != nil {
		return nil, fmt.Errorf("failed to get market data: %w", err)
	}
	
	factors := []FactorValue{
		CalculateTrendFactor(data.Bars),
		CalculateMomentumFactor(data.Bars),
		CalculateVolatilityFactor(data.VIX),
	}
	
	// Add price level factor
	priceFactor := calculatePriceLevelFactor(data)
	factors = append(factors, priceFactor)
	
	signal := CompositeSignal(factors)
	
	// Build contributing factors list
	contributing := []string{}
	for _, f := range factors {
		if abs(f.Value) > 0.3 {
			contributing = append(contributing, f.Name)
		}
	}
	
	return &UnifiedResult{
		Pair:                pair,
		Direction:           signal.Direction,
		Confidence:          signal.Confidence,
		Strength:            signal.Strength,
		ContributingFactors: contributing,
		Timestamp:           time.Now(),
	}, nil
}

// CalculateRadar generates radar view points for a category.
func (e *Engine) CalculateRadar(category string) ([]RadarPoint, error) {
	pairs := e.provider.GetPairsByCategory(category)
	
	var points []RadarPoint
	for _, pair := range pairs {
		data, err := e.provider.GetMarketData(pair)
		if err != nil {
			continue // Skip pairs with no data
		}
		
		trendFactor := CalculateTrendFactor(data.Bars)
		momentumFactor := CalculateMomentumFactor(data.Bars)
		
		point := RadarPoint{
			Pair:     pair,
			X:        trendFactor.Value,
			Y:        momentumFactor.Value,
			Strength: (abs(trendFactor.Value) + abs(momentumFactor.Value)) / 2,
		}
		
		// Determine quadrant
		point.Quadrant = determineQuadrant(point.X, point.Y)
		
		points = append(points, point)
	}
	
	return points, nil
}

// GenerateBriefing creates market briefing for a category.
func (e *Engine) GenerateBriefing(category string) (*BriefingResult, error) {
	pairs := e.provider.GetPairsByCategory(category)
	
	var sections []BriefingSection
	
	// Analyze each pair for significant signals
	bullishCount, bearishCount := 0, 0
	var strongSignals []string
	
	for _, pair := range pairs {
		result, err := e.CalculateUnified(pair)
		if err != nil {
			continue
		}
		
		switch result.Direction {
		case "bullish":
			bullishCount++
			if result.Confidence > 0.7 {
				strongSignals = append(strongSignals, pair)
			}
		case "bearish":
			bearishCount++
			if result.Confidence > 0.7 {
				strongSignals = append(strongSignals, pair)
			}
		}
	}
	
	total := len(pairs)
	if total == 0 {
		return &BriefingResult{Sections: sections, Timestamp: time.Now()}, nil
	}
	
	// Generate market sentiment section
	sentiment := "neutral"
	if bullishCount > bearishCount*2 {
		sentiment = "strongly_bullish"
	} else if bullishCount > bearishCount {
		sentiment = "bullish"
	} else if bearishCount > bullishCount*2 {
		sentiment = "strongly_bearish"
	} else if bearishCount > bullishCount {
		sentiment = "bearish"
	}
	
	sections = append(sections, BriefingSection{
		Title:    "Market Sentiment",
		Content:  fmt.Sprintf("%s: %d bullish, %d bearish out of %d pairs", sentiment, bullishCount, bearishCount, total),
		Priority: determineSentimentPriority(sentiment),
		Pairs:    pairs,
	})
	
	// Generate strong signals section
	if len(strongSignals) > 0 {
		sections = append(sections, BriefingSection{
			Title:    "High Confidence Signals",
			Content:  fmt.Sprintf("Pairs with >70%% confidence: %v", strongSignals),
			Priority: "high",
			Pairs:    strongSignals,
		})
	}
	
	// Generate volatility alert if needed
	avgVIX := e.calculateAvgVIX(pairs)
	if avgVIX > 30 {
		sections = append(sections, BriefingSection{
			Title:    "Volatility Alert",
			Content:  fmt.Sprintf("Average VIX at %.1f, elevated risk environment", avgVIX),
			Priority: "high",
			Pairs:    pairs,
		})
	}
	
	return &BriefingResult{
		Sections:  sections,
		Timestamp: time.Now(),
	}, nil
}

// Helper types

type BiasResult struct {
	Pair       string
	Timeframe  string
	Direction  string
	Confidence float64
	Strength   float64
	Factors    []FactorValue
	Timestamp  time.Time
}

type UnifiedResult struct {
	Pair                string
	Direction           string
	Confidence          float64
	Strength            float64
	ContributingFactors []string
	Timestamp           time.Time
}
type BriefingResult struct {
	Sections  []BriefingSection
	Timestamp time.Time
}

// Helper functions
func calculatePriceLevelFactor(data *MarketData) FactorValue {
	// Simple mean reversion factor based on 24h change
	change := data.ChangePct24h
	normalized := -change / 5.0 // Mean reversion: opposite of recent move
	
	if normalized > 1 {
		normalized = 1
	} else if normalized < -1 {
		normalized = -1
	}
	
	return FactorValue{
		Name:       "mean_reversion",
		Value:      normalized,
		Weight:     0.15,
		Confidence: 0.5,
	}
}

func determineQuadrant(x, y float64) string {
	// Check ranging first (values close to zero)
	if abs(x) < 0.3 && abs(y) < 0.3 {
		return "ranging"
	}
	if x > 0 && y > 0 {
		return "bull"
	}
	if x < 0 && y < 0 {
		return "bear"
	}
	return "reversal"
}

func determineSentimentPriority(sentiment string) string {
	switch sentiment {
	case "strongly_bullish", "strongly_bearish":
		return "high"
	case "bullish", "bearish":
		return "medium"
	default:
		return "low"
	}
}

func (e *Engine) calculateAvgVIX(pairs []string) float64 {
	total := 0.0
	count := 0
	for _, pair := range pairs {
		data, err := e.provider.GetMarketData(pair)
		if err != nil {
			continue
		}
		total += data.VIX
		count++
	}
	if count == 0 {
		return 0
	}
	return total / float64(count)
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
