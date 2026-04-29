package quant

import (
	"errors"
	"math"
)

// HurstResult Hurst 指数估计结果。
type HurstResult struct {
	H              float64
	SampleSize     int
	Interpretation string // trending / mean_reverting / random_walk
}

// HurstRS Hurst 指数 R/S 法估计。
//
// 算法（Hurst 1951 / Mandelbrot）：
//  1. 把序列分成 m 个长度 n 的不重叠段；
//  2. 对每段 i：去均值得到 Y_t = X_t - mean；累积偏离 Z_t = sum_{k=1..t}(Y_k)；
//     R_i = max(Z) - min(Z)；S_i = std(X)；
//  3. R/S 段平均 = mean(R_i / S_i)；
//  4. 在不同 n 上做 log(R/S) ~ H * log(n) 的最小二乘拟合，斜率即为 H。
//
// 选 n 的网格：n in {16, 32, 64, 128, 256, 512, 1024} 的子集（取 ≤ N/2 的值）。
//
// 输入序列长度 < 64 返回 ErrInsufficientData。
// 输出 H ∈ [0, 1]：
//   - H > 0.55 → trending（持续性）
//   - H < 0.45 → mean_reverting（反持续）
//   - 否则 → random_walk
func HurstRS(series []float64) (*HurstResult, error) {
	N := len(series)
	if N < 64 {
		return nil, ErrInsufficientData
	}
	candidates := []int{16, 32, 64, 128, 256, 512, 1024, 2048}
	type sample struct {
		logN, logRS float64
	}
	var pts []sample
	for _, n := range candidates {
		if n > N/2 {
			break
		}
		m := N / n
		if m < 2 {
			continue
		}
		var sumRS, count float64
		for i := 0; i < m; i++ {
			seg := series[i*n : (i+1)*n]
			rs, ok := rescaledRange(seg)
			if !ok {
				continue
			}
			sumRS += rs
			count++
		}
		if count == 0 {
			continue
		}
		avg := sumRS / count
		if avg <= 0 {
			continue
		}
		pts = append(pts, sample{logN: math.Log(float64(n)), logRS: math.Log(avg)})
	}
	if len(pts) < 3 {
		return nil, errors.New("quant: Hurst: not enough valid scales")
	}
	// OLS 拟合 logRS = a + H * logN
	var sx, sy, sxx, sxy float64
	for _, p := range pts {
		sx += p.logN
		sy += p.logRS
		sxx += p.logN * p.logN
		sxy += p.logN * p.logRS
	}
	k := float64(len(pts))
	denom := k*sxx - sx*sx
	if denom == 0 {
		return nil, errors.New("quant: Hurst: degenerate regression")
	}
	H := (k*sxy - sx*sy) / denom
	if H < 0 {
		H = 0
	} else if H > 1 {
		H = 1
	}
	interp := "random_walk"
	if H > 0.55 {
		interp = "trending"
	} else if H < 0.45 {
		interp = "mean_reverting"
	}
	return &HurstResult{H: H, SampleSize: N, Interpretation: interp}, nil
}

// rescaledRange 单段 R/S。返回 (R/S, ok)；S=0（恒定序列）时 ok=false。
func rescaledRange(seg []float64) (float64, bool) {
	n := len(seg)
	if n < 2 {
		return 0, false
	}
	// mean
	var mu float64
	for _, v := range seg {
		mu += v
	}
	mu /= float64(n)
	// std (sample)
	var sd float64
	for _, v := range seg {
		d := v - mu
		sd += d * d
	}
	sd = math.Sqrt(sd / float64(n))
	if sd == 0 {
		return 0, false
	}
	// cumulative deviation Z_t
	var z, zMin, zMax float64
	zMin = math.Inf(1)
	zMax = math.Inf(-1)
	for _, v := range seg {
		z += v - mu
		if z < zMin {
			zMin = z
		}
		if z > zMax {
			zMax = z
		}
	}
	R := zMax - zMin
	return R / sd, true
}
