package intermarket

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/antclaw/antclaw/internal/domain/shared"
)

// Engine tracks intermarket correlations and ratios
type Engine struct {
	priceProvider PriceProvider
}

// PriceProvider provides price data
type PriceProvider interface {
	GetPrices(ctx context.Context, symbol string, days int) ([]float64, error)
}

// NewEngine creates a new intermarket engine
func NewEngine(provider PriceProvider) *Engine {
	return &Engine{priceProvider: provider}
}

// IntermarketSignal represents intermarket analysis output
type IntermarketSignal struct {
	Timestamp    time.Time
	Pairs        []PairSignal
	RiskOnOff    string // "RISK_ON", "RISK_OFF", "NEUTRAL"
	Anomalies    []PairSignal
}

// PairSignal represents analysis for a pair
type PairSignal struct {
	Pair          CorrelationPair
	Correlation30 float64
	Correlation90 float64
	Ratio         float64
	Status        string // "NORMAL", "EXTREME", "DIVERGING"
	DirectionHint shared.Direction
}

// CorrelationPair defines a pair of assets to track
type CorrelationPair struct {
	Name        string
	Asset1      string
	Asset2      string
	ExpectedCorr float64 // Expected correlation (positive or negative)
}

// ClassicPairs defines standard intermarket pairs
var ClassicPairs = []CorrelationPair{
	{Name: "DXY_EURUSD", Asset1: "DXY", Asset2: "EURUSD", ExpectedCorr: -0.8},   // Inverse
	{Name: "DXY_GOLD", Asset1: "DXY", Asset2: "XAUUSD", ExpectedCorr: -0.7},     // Inverse
	{Name: "SPX_VIX", Asset1: "SPX", Asset2: "VIX", ExpectedCorr: -0.85},        // Inverse
	{Name: "SPX_TNX", Asset1: "SPX", Asset2: "TNX", ExpectedCorr: -0.6},          // Risk on/off
	{Name: "OIL_CAD", Asset1: "USOIL", Asset2: "USDCAD", ExpectedCorr: 0.7},      // Positive
	{Name: "AUD_GOLD", Asset1: "AUDUSD", Asset2: "XAUUSD", ExpectedCorr: 0.6},   // Positive
	{Name: "JPY_RISK", Asset1: "USDJPY", Asset2: "SPX", ExpectedCorr: -0.5},    // JPY safe haven
}

// Analyze performs intermarket analysis
func (e *Engine) Analyze(ctx context.Context) (*IntermarketSignal, error) {
	signal := &IntermarketSignal{
		Timestamp: time.Now(),
		Pairs:     make([]PairSignal, 0),
		Anomalies: make([]PairSignal, 0),
	}

	riskOnSignals := 0
	riskOffSignals := 0

	for _, pair := range ClassicPairs {
		pairSignal, err := e.analyzePair(ctx, pair)
		if err != nil {
			continue
		}

		signal.Pairs = append(signal.Pairs, *pairSignal)

		// Check for anomalies
		if pairSignal.Status == "EXTREME" || pairSignal.Status == "DIVERGING" {
			signal.Anomalies = append(signal.Anomalies, *pairSignal)
		}

		// Risk on/off detection
		if pair.Name == "SPX_VIX" || pair.Name == "SPX_TNX" {
			if pairSignal.DirectionHint == shared.DirectionLong {
				riskOnSignals++
			} else if pairSignal.DirectionHint == shared.DirectionShort {
				riskOffSignals++
			}
		}
	}

	// Determine overall risk regime
	if riskOnSignals > riskOffSignals+1 {
		signal.RiskOnOff = "RISK_ON"
	} else if riskOffSignals > riskOnSignals+1 {
		signal.RiskOnOff = "RISK_OFF"
	} else {
		signal.RiskOnOff = "NEUTRAL"
	}

	return signal, nil
}

// analyzePair analyzes a single pair
func (e *Engine) analyzePair(ctx context.Context, pair CorrelationPair) (*PairSignal, error) {
	prices1, err := e.priceProvider.GetPrices(ctx, pair.Asset1, 120)
	if err != nil {
		return nil, err
	}

	prices2, err := e.priceProvider.GetPrices(ctx, pair.Asset2, 120)
	if err != nil {
		return nil, err
	}

	if len(prices1) < 90 || len(prices2) < 90 {
		return nil, fmt.Errorf("insufficient data")
	}

	// Calculate correlations
	corr30 := correlation(prices1[len(prices1)-30:], prices2[len(prices2)-30:])
	corr90 := correlation(prices1[len(prices1)-90:], prices2[len(prices2)-90:])

	// Calculate ratio
	ratio := prices1[len(prices1)-1] / prices2[len(prices2)-1]

	signal := &PairSignal{
		Pair:          pair,
		Correlation30: corr30,
		Correlation90: corr90,
		Ratio:         ratio,
	}

	// Determine status
	signal.Status = e.determineStatus(corr30, corr90, pair.ExpectedCorr)
	signal.DirectionHint = e.inferDirection(corr30, pair.ExpectedCorr)

	return signal, nil
}

// determineStatus determines pair status
func (e *Engine) determineStatus(corr30, corr90, expected float64) string {
	// Check if correlation has broken down
	if math.Abs(corr30) < math.Abs(expected)*0.5 {
		return "DIVERGING"
	}

	// Check for extreme correlation
	if math.Abs(corr30) > 0.95 {
		return "EXTREME"
	}

	return "NORMAL"
}

// inferDirection infers directional hint from correlation
func (e *Engine) inferDirection(corr float64, expected float64) shared.Direction {
	// If correlation is as expected and positive -> aligned movement
	if corr*expected > 0 {
		if corr > 0.5 {
			return shared.DirectionLong
		}
	} else {
		// Inverted correlation -> inverse relationship
		if corr < -0.5 {
			return shared.DirectionShort
		}
	}

	return shared.DirectionNeutral
}

// correlation calculates Pearson correlation
func correlation(x, y []float64) float64 {
	if len(x) != len(y) || len(x) < 2 {
		return 0
	}

	meanX, meanY := mean(x), mean(y)

	var sumXY, sumX2, sumY2 float64
	for i := 0; i < len(x); i++ {
		dx := x[i] - meanX
		dy := y[i] - meanY
		sumXY += dx * dy
		sumX2 += dx * dx
		sumY2 += dy * dy
	}

	if sumX2 == 0 || sumY2 == 0 {
		return 0
	}

	return sumXY / math.Sqrt(sumX2*sumY2)
}

// mean calculates arithmetic mean
func mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}
