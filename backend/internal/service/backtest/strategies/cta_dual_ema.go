package strategies

func CTADualEMA(fast, slow float64) string {
	if fast > slow {
		return "long"
	}
	if fast < slow {
		return "short"
	}
	return "flat"
}
