package strategies

func CTADonchianBreakout(close, upper, lower float64) string {
	if close > upper {
		return "long"
	}
	if close < lower {
		return "short"
	}
	return "flat"
}
