package score

import (
	"math"
)

// TrendQualityScore calculates trend quality score
type TrendQualityScore struct{}

// NewTrendQualityScore creates trend quality scorer
func NewTrendQualityScore() *TrendQualityScore {
	return &TrendQualityScore{}
}

// TrendQualityInput contains required data
type TrendQualityInput struct {
	Closes     []float64
	Highs      []float64
	Lows       []float64
	Period20   []float64 // 20-period SMA
	Period50   []float64 // 50-period SMA
	Period200  []float64 // 200-period SMA
}

// Calculate computes trend quality score (-1 to 1)
func (t *TrendQualityScore) Calculate(input TrendQualityInput) float64 {
	if len(input.Closes) < 50 {
		return 0
	}

	// 1. MA alignment score
	maScore := t.maAlignment(input)

	// 2. Linear regression R² score
	r2Score := t.regressionRSquare(input.Closes)

	// 3. Consecutive days score
	consecutiveScore := t.consecutiveDays(input.Closes)

	// Combine scores
	finalScore := maScore*0.4 + r2Score*0.4 + consecutiveScore*0.2

	return math.Max(-1, math.Min(1, finalScore))
}

// maAlignment checks if MAs are aligned (bullish: 20>50>200)
func (t *TrendQualityScore) maAlignment(input TrendQualityInput) float64 {
	if len(input.Period20) == 0 || len(input.Period50) == 0 || len(input.Period200) == 0 {
		return 0
	}

	ma20 := input.Period20[len(input.Period20)-1]
	ma50 := input.Period50[len(input.Period50)-1]
	ma200 := input.Period200[len(input.Period200)-1]

	// Bullish alignment
	if ma20 > ma50 && ma50 > ma200 {
		return 1.0
	}

	// Bearish alignment
	if ma20 < ma50 && ma50 < ma200 {
		return -1.0
	}

	// Mixed
	return 0.0
}

// regressionRSquare calculates R² of linear regression
func (t *TrendQualityScore) regressionRSquare(closes []float64) float64 {
	if len(closes) < 20 {
		return 0
	}

	// Use last 20 days
	data := closes[len(closes)-20:]

	// Calculate slope and intercept
	n := float64(len(data))
	var sumX, sumY, sumXY, sumX2 float64

	for i, y := range data {
		x := float64(i)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	intercept := (sumY - slope*sumX) / n

	// Calculate R²
	var ssTot, ssRes float64
	meanY := sumY / n

	for i, y := range data {
		x := float64(i)
		predicted := slope*x + intercept
		ssTot += (y - meanY) * (y - meanY)
		ssRes += (y - predicted) * (y - predicted)
	}

	if ssTot == 0 {
		return 0
	}

	r2 := 1 - ssRes/ssTot

	// Direction based on slope
	if slope < 0 {
		return -r2
	}
	return r2
}

// consecutiveDays scores consecutive up/down days
func (t *TrendQualityScore) consecutiveDays(closes []float64) float64 {
	if len(closes) < 10 {
		return 0
	}

	recent := closes[len(closes)-10:]
	var consecutive int
	var direction float64

	for i := 1; i < len(recent); i++ {
		if recent[i] > recent[i-1] {
			if direction >= 0 {
				consecutive++
				direction = 1
			} else {
				break
			}
		} else if recent[i] < recent[i-1] {
			if direction <= 0 {
				consecutive++
				direction = -1
			} else {
				break
			}
		}
	}

	// Normalize to -1 to 1
	return float64(consecutive) / 10 * direction
}
