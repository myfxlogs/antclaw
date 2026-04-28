package sizing

func VolTargetScale(targetVol, realizedVol float64) float64 {
	if realizedVol <= 0 {
		return 1
	}
	scale := targetVol / realizedVol
	if scale < 0.1 {
		return 0.1
	}
	if scale > 3 {
		return 3
	}
	return scale
}
