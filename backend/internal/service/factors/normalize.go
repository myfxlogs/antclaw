package factors

func normalizeToScore(raw map[string]float64) map[string]float64 {
	out := make(map[string]float64, len(raw))
	minV, maxV := 0.0, 0.0
	first := true
	for _, v := range raw {
		if first {
			minV, maxV = v, v
			first = false
			continue
		}
		if v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
	}
	span := maxV - minV
	if span == 0 {
		for k := range raw {
			out[k] = 50
		}
		return out
	}
	for k, v := range raw {
		out[k] = (v - minV) / span * 100.0
	}
	return out
}
