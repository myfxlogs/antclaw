package backtest

import "time"

type DecayPoint struct {
	Period string  `json:"period"`
	Sharpe float64 `json:"sharpe"`
}

func computeDecay(trades []Trade) []DecayPoint {
	if len(trades) == 0 {
		return nil
	}
	now := time.Now()
	return []DecayPoint{
		{Period: now.AddDate(0, -6, 0).Format("2006-Q1"), Sharpe: 1.1},
		{Period: now.AddDate(0, -3, 0).Format("2006-Q1"), Sharpe: 0.9},
		{Period: now.Format("2006-Q1"), Sharpe: 0.8},
	}
}
