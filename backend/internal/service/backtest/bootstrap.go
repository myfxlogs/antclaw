package backtest

import (
	"math"
	"math/rand"
	"sort"
)

// sortFloats 升序原地排序（导出给本包内其他文件使用）。
func sortFloats(x []float64) { sort.Float64s(x) }

// bootstrapMaxDDs block-bootstrap 估计 MaxDD 分布。
//
// 块长按 N^(1/3) 启发式选择，保留近邻自相关。
// 返回 iter 个 MaxDD（小数），由调用方排序取分位。
func bootstrapMaxDDs(returns []float64, iter int, seed uint64) []float64 {
	n := len(returns)
	if n < 5 || iter <= 0 {
		return nil
	}
	src := rand.NewSource(int64(seed))
	if seed == 0 {
		src = rand.NewSource(1)
	}
	rng := rand.New(src)
	blockLen := int(math.Cbrt(float64(n)))
	if blockLen < 2 {
		blockLen = 2
	}
	out := make([]float64, 0, iter)
	for i := 0; i < iter; i++ {
		// 拼接成 n 长度的样本
		sample := make([]float64, 0, n)
		for len(sample) < n {
			start := rng.Intn(n)
			for k := 0; k < blockLen && len(sample) < n; k++ {
				sample = append(sample, returns[(start+k)%n])
			}
		}
		// equity curve & maxDD
		eq, peak, maxDD := 1.0, 1.0, 0.0
		for _, r := range sample {
			eq *= 1 + r
			if eq > peak {
				peak = eq
			}
			dd := (peak - eq) / peak
			if dd > maxDD {
				maxDD = dd
			}
		}
		out = append(out, maxDD)
	}
	return out
}

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
