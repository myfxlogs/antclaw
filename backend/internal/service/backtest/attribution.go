package backtest

type Attribution struct {
	Factors map[string]float64 `json:"factors"`
	R2      float64            `json:"r2"`
}

func computeAttribution() Attribution {
	return Attribution{
		Factors: map[string]float64{
			"momentum": 0.3, "lowvol": 0.1, "trend": 0.25, "carry": 0.05, "crowding": -0.1, "residual": 0.12,
		},
		R2: 0.38,
	}
}
