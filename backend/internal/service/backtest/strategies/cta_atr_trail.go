package strategies

func CTAATRStop(entry, atr, k float64, side string) float64 {
	if side == "long" {
		return entry - atr*k
	}
	return entry + atr*k
}
