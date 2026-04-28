package signals

import (
	"fmt"
	"math"
	"sort"
)

func tanh(v float64) float64 { return math.Tanh(v) }

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func sigmoid(x float64) float64 { return 1.0 / (1.0 + math.Exp(-x)) }

func directionLabel(score float64) string {
	switch {
	case score > 0.2:
		return "bullish"
	case score < -0.2:
		return "bearish"
	default:
		return "neutral"
	}
}

func RecommendationToDirection(rec string) string {
	switch rec {
	case "STRONG_LONG", "LONG":
		return "bullish"
	case "STRONG_SHORT", "SHORT":
		return "bearish"
	default:
		return "neutral"
	}
}

func unifiedRecommendation(score float64) string {
	switch {
	case score >= 0.6:
		return "STRONG_LONG"
	case score >= 0.2:
		return "LONG"
	case score <= -0.6:
		return "STRONG_SHORT"
	case score <= -0.2:
		return "SHORT"
	default:
		return "NEUTRAL"
	}
}

func classify(x, y float64) string {
	isHigh := x >= 60
	isBull := y > 0.1
	switch {
	case isHigh && isBull:
		return "high_bull"
	case isHigh && !isBull:
		return "high_bear"
	case !isHigh && isBull:
		return "low_bull"
	default:
		return "low_bear"
	}
}

func intensityLabel(pct float64) string {
	switch {
	case pct < 30:
		return "weak"
	case pct < 60:
		return "moderate"
	case pct < 85:
		return "strong"
	default:
		return "extreme"
	}
}

func percentile(value float64, sample []float64) float64 {
	if len(sample) == 0 {
		return 0
	}
	cp := append([]float64(nil), sample...)
	sort.Float64s(cp)
	idx := sort.SearchFloat64s(cp, value)
	return float64(idx+1) / float64(len(cp)) * 100
}

func factorDescription(name string, value float64) string {
	strength := "weak"
	if math.Abs(value) > 1 {
		strength = "strong"
	}
	switch name {
	case "Momentum":
		return fmt.Sprintf("12W momentum %s (%.2f)", strength, value)
	case "LowVol":
		return fmt.Sprintf("Volatility %s vs peers (%.2f)", strength, value)
	case "Trend":
		return fmt.Sprintf("Trend quality %s (%.2f)", strength, value)
	case "Carry":
		return fmt.Sprintf("Carry spread %s (%.2f)", strength, value)
	case "Crowding":
		return fmt.Sprintf("Crowding pressure %s (%.2f)", strength, value)
	case "Residual":
		return fmt.Sprintf("Residual reversal %s (%.2f)", strength, value)
	default:
		return fmt.Sprintf("%s %.2f", name, value)
	}
}
