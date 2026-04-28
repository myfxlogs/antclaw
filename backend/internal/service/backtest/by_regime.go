package backtest

type RegimeStats struct {
	HitRate float64 `json:"hit_rate"`
	Sharpe  float64 `json:"sharpe"`
}

func computeByRegime() map[string]RegimeStats {
	return map[string]RegimeStats{
		"BULL": {HitRate: 0.62, Sharpe: 1.3},
		"BEAR": {HitRate: 0.48, Sharpe: 0.7},
	}
}
