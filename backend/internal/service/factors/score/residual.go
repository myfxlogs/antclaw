package score

import (
	"math"
)

// ResidualScore calculates residual reversal score
// Assets that have deviated significantly from market expected to mean-revert
type ResidualScore struct{}

// NewResidualScore creates residual scorer
func NewResidualScore() *ResidualScore {
	return &ResidualScore{}
}

// Calculate computes residual score (-1 to 1)
// OLS regression vs market, large residual = potential reversal
func (r *ResidualScore) Calculate(assetReturns, marketReturns []float64) float64 {
	if len(assetReturns) != len(marketReturns) || len(assetReturns) < 20 {
		return 0
	}

	// Calculate OLS regression: asset = alpha + beta * market
	beta, alpha := r.calculateOLS(assetReturns, marketReturns)

	// Calculate residuals for last period
	latestAsset := assetReturns[len(assetReturns)-1]
	latestMarket := marketReturns[len(marketReturns)-1]
	predicted := alpha + beta*latestMarket
	residual := latestAsset - predicted

	// Calculate historical residual std dev
	residuals := r.calculateResiduals(assetReturns, marketReturns, alpha, beta)
	resStdDev := r.stdDev(residuals, 0)

	if resStdDev == 0 {
		return 0
	}

	// Z-score of residual
	zScore := residual / resStdDev

	// Large positive residual -> negative score (expect mean reversion down)
	// Large negative residual -> positive score (expect mean reversion up)
	return -math.Tanh(zScore)
}

// calculateOLS calculates OLS regression coefficients
func (r *ResidualScore) calculateOLS(y, x []float64) (beta, alpha float64) {
	n := float64(len(y))
	var sumX, sumY, sumXY, sumX2 float64

	for i := 0; i < len(y); i++ {
		sumX += x[i]
		sumY += y[i]
		sumXY += x[i] * y[i]
		sumX2 += x[i] * x[i]
	}

	beta = (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	alpha = (sumY - beta*sumX) / n

	return beta, alpha
}

// calculateResiduals calculates all residuals
func (r *ResidualScore) calculateResiduals(assetReturns, marketReturns []float64, alpha, beta float64) []float64 {
	residuals := make([]float64, len(assetReturns))
	for i := 0; i < len(assetReturns); i++ {
		predicted := alpha + beta*marketReturns[i]
		residuals[i] = assetReturns[i] - predicted
	}
	return residuals
}

// stdDev calculates standard deviation
func (r *ResidualScore) stdDev(values []float64, mean float64) float64 {
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
