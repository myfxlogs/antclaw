package backtest

import "strings"

type strategyKey struct {
	Type    string
	Symbol  string
	Horizon string
}

func parseStrategyKey(key string) strategyKey {
	out := strategyKey{Type: "unified", Horizon: "1W"}
	parts := strings.Split(strings.TrimSpace(key), ":")
	if len(parts) >= 1 && parts[0] != "" {
		out.Type = strings.ToLower(parts[0])
	}
	if len(parts) >= 2 && parts[1] != "" && parts[1] != "*" {
		out.Symbol = strings.ToUpper(parts[1])
	}
	if len(parts) >= 3 && parts[2] != "" {
		out.Horizon = strings.ToUpper(parts[2])
	}
	return out
}
