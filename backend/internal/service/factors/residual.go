package factors

func residualScore(series, peers []float64) float64 {
	if len(series) == 0 || len(peers) == 0 {
		return 0
	}
	n := len(series)
	if len(peers) < n {
		n = len(peers)
	}
	var sum float64
	for i := 0; i < n; i++ {
		sum += series[i] - peers[i]
	}
	return sum / float64(n)
}
