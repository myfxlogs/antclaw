package score

import (
	"math"
)

// MomentumScore calculates multi-window momentum score
type MomentumScore struct {
	Periods []int // Windows in days, e.g., [20, 60, 120, 240]
}

// NewMomentumScore creates momentum scorer with default periods
func NewMomentumScore() *MomentumScore {
	return &MomentumScore{
		Periods: []int{20, 60, 120, 240}, // 1M, 3M, 6M, 12M
	}
}

// Calculate computes momentum score (-1 to 1)
// Skips most recent 1M to avoid short-term reversal
func (m *MomentumScore) Calculate(closes []float64) float64 {
	if len(closes) < 240 {
		return 0
	}

	// Skip recent 20 days (1M)
	offset := 20
	effectiveLen := len(closes) - offset

	if effectiveLen < 220 {
		return 0
	}

	current := closes[len(closes)-1-offset]

	weights := []float64{0.4, 0.3, 0.2, 0.1}
	var weightedScore float64

	for i, period := range m.Periods {
		if effectiveLen < period {
			continue
		}

		past := closes[len(closes)-1-offset-period]
		if past == 0 {
			continue
		}

		// Calculate return
		ret := (current - past) / past

		// Normalize to -1 to 1 range (annualized returns typically +/- 50%)
		score := math.Tanh(ret * 2) // Tanh normalizes to -1 to 1

		weightedScore += score * weights[i]
	}

	return weightedScore
}

// CalculateWithCarry calculates momentum adjusted by carry
func (m *MomentumScore) CalculateWithCarry(closes []float64, carryRate float64) float64 {
	momentum := m.Calculate(closes)

	// Adjust momentum by carry (positive carry boosts score)
	// Typical FX carry: +/- 5%
	carryAdjustment := carryRate * 10 // Scale to comparable magnitude

	return momentum + carryAdjustment*0.15
}
