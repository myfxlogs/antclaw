package quant

import "math"

// PearsonCorr 两个等长序列的 Pearson 相关系数。
// 序列方差为 0 时返回 0（无意义）。
func PearsonCorr(a, b []float64) float64 {
	n := len(a)
	if n != len(b) || n < 2 {
		return 0
	}
	var sa, sb float64
	for i := 0; i < n; i++ {
		sa += a[i]
		sb += b[i]
	}
	ma := sa / float64(n)
	mb := sb / float64(n)
	var num, va, vb float64
	for i := 0; i < n; i++ {
		da := a[i] - ma
		db := b[i] - mb
		num += da * db
		va += da * da
		vb += db * db
	}
	if va == 0 || vb == 0 {
		return 0
	}
	return num / math.Sqrt(va*vb)
}

// RollingCorrelationMatrix 给定 N 个等长序列，输出 N*N 的 Pearson 相关矩阵
// 使用最后 window 个观测；window <= 1 时回退到全长。返回值对角线必为 1.0。
func RollingCorrelationMatrix(series [][]float64, window int) [][]float64 {
	n := len(series)
	out := make([][]float64, n)
	for i := range out {
		out[i] = make([]float64, n)
		out[i][i] = 1.0
	}
	if n < 2 {
		return out
	}
	clip := func(s []float64) []float64 {
		if window <= 1 || window >= len(s) {
			return s
		}
		return s[len(s)-window:]
	}
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			a := clip(series[i])
			b := clip(series[j])
			// 对齐到较短长度。
			m := len(a)
			if len(b) < m {
				m = len(b)
			}
			if m < 2 {
				continue
			}
			c := PearsonCorr(a[len(a)-m:], b[len(b)-m:])
			out[i][j] = c
			out[j][i] = c
		}
	}
	return out
}
