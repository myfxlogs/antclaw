package score

import (
	"math"
)

// CarryScore calculates carry-adjusted score
type CarryScore struct{}

// NewCarryScore creates carry scorer
func NewCarryScore() *CarryScore {
	return &CarryScore{}
}

// Calculate computes carry score (-1 to 1)
// For FX: interest rate differential
// For crypto: funding rate
func (c *CarryScore) Calculate(carryRate float64) float64 {
	// Normalize carry: typical range +/- 5% annually
	// Scale to -1 to 1
	return math.Tanh(carryRate * 10)
}

// CarryAdjustedMomentum combines momentum with carry
type CarryAdjustedMomentum struct {
	momentumScorer *MomentumScore
	carryScorer    *CarryScore
	Alpha          float64 // Weight for carry adjustment (default 0.15)
}

// NewCarryAdjustedMomentum creates carry-adjusted momentum scorer
func NewCarryAdjustedMomentum() *CarryAdjustedMomentum {
	return &CarryAdjustedMomentum{
		momentumScorer: NewMomentumScore(),
		carryScorer:    NewCarryScore(),
		Alpha:          0.15,
	}
}

// Calculate computes carry-adjusted momentum
func (c *CarryAdjustedMomentum) Calculate(closes []float64, carryRate float64) float64 {
	momentum := c.momentumScorer.Calculate(closes)
	carry := c.carryScorer.Calculate(carryRate)

	// Combine: momentum + alpha * normalized(carry)
	return momentum + c.Alpha*carry
}
