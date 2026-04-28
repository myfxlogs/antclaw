package factors

func trendScore(closes []float64) float64 {
	if len(closes) < 50 {
		return 0
	}
	ema20 := ema(closes, 20)
	ema50 := ema(closes, 50)
	if ema20 > ema50 {
		return 1
	}
	if ema20 < ema50 {
		return -1
	}
	return 0
}

func ema(closes []float64, period int) float64 {
	if len(closes) < period {
		return 0
	}
	k := 2.0 / float64(period+1)
	v := closes[0]
	for i := 1; i < len(closes); i++ {
		v = closes[i]*k + v*(1-k)
	}
	return v
}
