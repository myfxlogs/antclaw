package strategies

func CTAMultiTF(dailyTrend, intradaySignal string) string {
	if dailyTrend == "bullish" && intradaySignal == "long" {
		return "long"
	}
	if dailyTrend == "bearish" && intradaySignal == "short" {
		return "short"
	}
	return "flat"
}
