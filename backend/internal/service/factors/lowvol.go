package factors

import "math"

func lowVolScore(returns []float64) float64 {
	if len(returns) < 60 {
		return 0
	}
	sub := returns[len(returns)-60:]
	var sum float64
	for _, v := range sub {
		sum += v
	}
	mean := sum / float64(len(sub))
	var ss float64
	for _, v := range sub {
		d := v - mean
		ss += d * d
	}
	vol := math.Sqrt(ss/float64(len(sub))) * math.Sqrt(252)
	return -vol
}
