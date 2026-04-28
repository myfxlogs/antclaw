package factors

import "math"

func momentum12Minus1(closes []float64) float64 {
	if len(closes) < 273 {
		return 0
	}
	ret12_1 := closes[len(closes)-22]/closes[len(closes)-273] - 1
	ret1m := closes[len(closes)-1]/closes[len(closes)-22] - 1
	return ret12_1 - ret1m
}

func tanhScale(v float64) float64 {
	return math.Tanh(v)
}
