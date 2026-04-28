package signals

import "strings"

var Categories = map[string][]string{
	"majors":  {"EURUSD", "GBPUSD", "USDJPY", "USDCHF", "AUDUSD", "USDCAD", "NZDUSD"},
	"crosses": {"EURGBP", "EURJPY", "GBPJPY", "EURAUD", "AUDJPY", "CHFJPY", "EURCHF"},
	"crypto":  {"BTCUSDT", "ETHUSDT", "SOLUSDT", "BNBUSDT", "XRPUSDT"},
	"indices": {"SPX", "NDX", "DJI", "RUT", "DAX"},
}

func SymbolsByCategory(category string) []string {
	if strings.EqualFold(category, "all") || strings.TrimSpace(category) == "" {
		var all []string
		for _, symbols := range Categories {
			all = append(all, symbols...)
		}
		return all
	}
	return append([]string(nil), Categories[strings.ToLower(category)]...)
}
