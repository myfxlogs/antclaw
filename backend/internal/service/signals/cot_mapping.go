package signals

import "strings"

var SymbolToCOTContract = map[string]string{
	"EURUSD": "099741", "GBPUSD": "096742", "USDJPY": "097741",
	"USDCHF": "092741", "AUDUSD": "232741", "USDCAD": "090741",
	"NZDUSD": "112741", "BTCUSDT": "133741", "ETHUSDT": "146021",
	"GOLD": "088691", "XAUUSD": "088691", "WTI": "067651",
	"SILVER": "084691", "SPX": "13874A", "NDX": "209742",
}

func ResolveCOTCode(symbol string) (string, bool) {
	code, ok := SymbolToCOTContract[strings.ToUpper(symbol)]
	return code, ok
}
