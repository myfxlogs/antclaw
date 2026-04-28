package backtest

type BaselineSummary struct {
	BuyHoldReturn float64 `json:"buyhold_return"`
	RandomReturn  float64 `json:"random_return"`
}

func computeBaselines() BaselineSummary {
	return BaselineSummary{BuyHoldReturn: 0.08, RandomReturn: 0.01}
}
