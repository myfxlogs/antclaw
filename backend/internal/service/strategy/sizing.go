package strategy

import (
	"fmt"
	"math"
)

// ATRSizer calculates ATR-based position sizing
type ATRSizer struct{}

// NewATRSizer creates a new ATR sizer
func NewATRSizer() *ATRSizer {
	return &ATRSizer{}
}

// SizingInput contains sizing parameters
type SizingInput struct {
	AccountBalance float64
	RiskPercent    float64   // Max risk per trade (e.g., 0.02 = 2%)
	ATR            float64   // 14-period ATR
	ATRMultiplier  float64   // Stop distance in ATRs (typically 1.5-2)
	EntryPrice     float64
}

// SizingResult contains sizing output
type SizingResult struct {
	Shares           float64
	PositionSize     float64
	StopDistance     float64
	RiskAmount       float64
	RiskPercent      float64
	PositionPercent  float64 // Position size as % of account
}

// Calculate computes position size based on ATR
func (s *ATRSizer) Calculate(input SizingInput) SizingResult {
	// Calculate stop distance in price terms
	stopDistance := input.ATR * input.ATRMultiplier

	// Calculate risk amount in currency
	riskAmount := input.AccountBalance * input.RiskPercent

	// Calculate number of shares
	// riskAmount = shares * stopDistance
	// shares = riskAmount / stopDistance
	shares := riskAmount / stopDistance

	// Calculate position size in currency
	positionSize := shares * input.EntryPrice

	// Calculate position as % of account
	positionPercent := positionSize / input.AccountBalance

	return SizingResult{
		Shares:          shares,
		PositionSize:    positionSize,
		StopDistance:    stopDistance,
		RiskAmount:      riskAmount,
		RiskPercent:     input.RiskPercent,
		PositionPercent: positionPercent,
	}
}

// CalculateATR calculates ATR from price bars
func CalculateATR(closes, highs, lows []float64, period int) float64 {
	if len(closes) < period+1 {
		return 0
	}

	var trs []float64

	for i := len(closes) - period; i < len(closes); i++ {
		if i == 0 {
			continue
		}

		// True Range = max(high-low, |high-prevClose|, |low-prevClose|)
		tr1 := highs[i] - lows[i]
		tr2 := math.Abs(highs[i] - closes[i-1])
		tr3 := math.Abs(lows[i] - closes[i-1])

		tr := math.Max(tr1, math.Max(tr2, tr3))
		trs = append(trs, tr)
	}

	if len(trs) == 0 {
		return 0
	}

	// Calculate average
	var sum float64
	for _, tr := range trs {
		sum += tr
	}

	return sum / float64(len(trs))
}

// SizingRecommendation provides sizing recommendations
type SizingRecommendation struct {
	Conservative SizingResult
	Moderate     SizingResult
	Aggressive   SizingResult
}

// Recommend provides multiple sizing options
func (s *ATRSizer) Recommend(input SizingInput) SizingRecommendation {
	// Conservative: 1.5 ATR stop, 1% risk
	conservative := input
	conservative.ATRMultiplier = 1.5
	conservative.RiskPercent = 0.01

	// Moderate: 2.0 ATR stop, 2% risk
	moderate := input
	moderate.ATRMultiplier = 2.0
	moderate.RiskPercent = 0.02

	// Aggressive: 1.5 ATR stop, 3% risk
	aggressive := input
	aggressive.ATRMultiplier = 1.5
	aggressive.RiskPercent = 0.03

	return SizingRecommendation{
		Conservative: s.Calculate(conservative),
		Moderate:     s.Calculate(moderate),
		Aggressive:   s.Calculate(aggressive),
	}
}

// ValidateSizing checks if sizing is reasonable
func ValidateSizing(result SizingResult, accountBalance float64) error {
	// Max 25% of account in one position
	if result.PositionPercent > 0.25 {
		return fmt.Errorf("position size %.1f%% exceeds 25%% limit", result.PositionPercent*100)
	}

	// Max 3% risk per trade
	if result.RiskPercent > 0.03 {
		return fmt.Errorf("risk %.1f%% exceeds 3%% limit", result.RiskPercent*100)
	}

	return nil
}
