package signals

import "sort"

var transitionStates = []string{"STRONG_BULL", "BULL", "NEUTRAL", "BEAR", "STRONG_BEAR"}

func transitionProbabilities(history []RegimeSnapshot, current string) []RegimeTransition {
	if len(history) < 2 {
		return nil
	}
	counts := map[string]map[string]int{}
	for _, s := range transitionStates {
		counts[s] = map[string]int{}
	}
	for i := 0; i < len(history)-1; i++ {
		from := history[i].UnifiedLabel
		to := history[i+1].UnifiedLabel
		if counts[from] == nil {
			counts[from] = map[string]int{}
		}
		counts[from][to]++
	}
	row := counts[current]
	if row == nil {
		row = map[string]int{}
	}
	total := 0
	for _, v := range row {
		total += v
	}
	if total == 0 {
		total = len(transitionStates)
		for _, s := range transitionStates {
			row[s] = 1
		}
	}
	out := make([]RegimeTransition, 0, len(transitionStates))
	for _, to := range transitionStates {
		p := float64(row[to]) / float64(total)
		out = append(out, RegimeTransition{FromLabel: current, ToLabel: to, Severity: "INFO", ToScore: p})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ToScore > out[j].ToScore })
	return out
}
