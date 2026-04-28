package score

import (
	"math"
)

// LowVolScore calculates low volatility anomaly score (Sharpe proxy)
type LowVolScore struct{}

// NewLowVolScore creates low vol scorer
func NewLowVolScore() *LowVolScore {
	return &LowVolScore{}
}

// Calculate computes low volatility score (-1 to 1)
// Rewards low volatility + positive returns (Sharpe-like)
func (l *LowVolScore) Calculate(returns []float64) float64 {
	if len(returns) < 20 {
		return 0
	}

	// Calculate annualized return and volatility
	mean := l.mean(returns)
	stdDev := l.stdDev(returns, mean)

	if stdDev == 0 {
		return 0
	}

	// Annualize (assuming daily returns)
	annualReturn := mean * 252
	annualVol := stdDev * math.Sqrt(252)

	// Sharpe-like ratio (simplified, no risk-free rate)
	sharpeLike := annualReturn / annualVol

	// Normalize to -1 to 1
	return math.Tanh(sharpeLike)
}

// mean calculates arithmetic mean
func (l *LowVolScore) mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// stdDev calculates standard deviation
func (l *LowVolScore) stdDev(values []float64, mean float64) float64 {
	if len(values) < 2 {
		return 0
	}

	var sumSquared float64
	for _, v := range values {
		diff := v - mean
		sumSquared += diff * diff
	}

	variance := sumSquared / float64(len(values)-1)
	return math.Sqrt(variance)
}
