package backtest

import (
	"math/rand"
	"sort"
)

// bootstrapPercentiles 基于一组 OOS sharpe 序列做 bootstrap 重采样，
// 返回 5/50/95 分位数。iter 为重采样轮次，seed 为 0 时按时间种子。
func bootstrapPercentiles(values []float64, iter int, seed uint64) (p5, p50, p95 float64) {
	if len(values) == 0 || iter <= 0 {
		return 0, 0, 0
	}
	src := rand.NewSource(int64(seed))
	if seed == 0 {
		src = rand.NewSource(1) // 固定种子保证可复现
	}
	rng := rand.New(src)
	means := make([]float64, 0, iter)
	n := len(values)
	for i := 0; i < iter; i++ {
		sum := 0.0
		for j := 0; j < n; j++ {
			sum += values[rng.Intn(n)]
		}
		means = append(means, sum/float64(n))
	}
	sort.Float64s(means)
	pick := func(p float64) float64 {
		idx := int(float64(len(means)-1) * p)
		return means[idx]
	}
	return pick(0.05), pick(0.50), pick(0.95)
}
