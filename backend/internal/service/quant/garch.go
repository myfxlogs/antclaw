// Package quant 提供 M-A 量化基础原语：GARCH(1,1) / Hurst / Gaussian HMM /
// 滚动相关 / 背离检测。所有函数为纯函数，不依赖外部 IO，便于单测与回测复用。
package quant

import (
	"errors"
	"math"
)

// ErrNonConvergent 似然优化未达到收敛阈值。
var ErrNonConvergent = errors.New("quant: GARCH MLE did not converge")

// ErrInsufficientData 输入序列长度不足以拟合模型。
var ErrInsufficientData = errors.New("quant: insufficient data for fitting")

// GARCHParams GARCH(1,1) 参数：sigma^2_t = omega + alpha*r^2_{t-1} + beta*sigma^2_{t-1}.
// 约束：omega>0, alpha>=0, beta>=0, alpha+beta<1。
type GARCHParams struct {
	Omega float64
	Alpha float64
	Beta  float64
}

// Persistence alpha+beta，越接近 1 表示波动率聚簇越强。
func (p GARCHParams) Persistence() float64 { return p.Alpha + p.Beta }

// UnconditionalVar 无条件方差 omega/(1-alpha-beta)。
func (p GARCHParams) UnconditionalVar() float64 {
	d := 1 - p.Alpha - p.Beta
	if d <= 0 {
		return math.NaN()
	}
	return p.Omega / d
}

// FitGARCH 用最大似然 + 坐标下降数值优化拟合 GARCH(1,1)。
//
// 实现要点：
//  1. 起始点：omega = var(returns)*0.05；alpha=0.05；beta=0.9（对金融日收益的常见经验初值）。
//  2. 用 Nelder-Mead 在 (omega, alpha, beta) 三维做对数似然最大化；约束通过参数变换实现：
//     omega = exp(theta0)，alpha = sigmoid(theta1)*0.5，beta = sigmoid(theta2)*0.99。
//  3. 收敛判定：相邻两次迭代对数似然差 < 1e-8 或步数 > 500。
//
// 返回拟合参数与对应的条件方差序列（长度 = len(returns)）。
//
// returns 必须为 0 均值的 log returns（建议先减样本均值）；长度 < 50 返回 ErrInsufficientData。
func FitGARCH(returns []float64) (*GARCHParams, []float64, error) {
	n := len(returns)
	if n < 50 {
		return nil, nil, ErrInsufficientData
	}
	mean := 0.0
	for _, r := range returns {
		mean += r
	}
	mean /= float64(n)
	demeaned := make([]float64, n)
	for i, r := range returns {
		demeaned[i] = r - mean
	}
	variance := 0.0
	for _, r := range demeaned {
		variance += r * r
	}
	variance /= float64(n)
	if variance <= 0 {
		return nil, nil, ErrInsufficientData
	}

	// 目标函数：负对数似然（最小化）。theta -> (omega, alpha, beta)。
	negLL := func(theta [3]float64) float64 {
		omega := math.Exp(theta[0])
		alpha := sigmoid(theta[1]) * 0.5
		beta := sigmoid(theta[2]) * 0.99
		// 约束 alpha+beta < 0.999；越界给极大惩罚。
		if alpha+beta >= 0.999 {
			return 1e18
		}
		// 初始方差：用样本方差。
		s2 := variance
		ll := 0.0
		for _, r := range demeaned {
			if s2 <= 1e-12 {
				return 1e18
			}
			ll += -0.5 * (math.Log(2*math.Pi*s2) + r*r/s2)
			s2 = omega + alpha*r*r + beta*s2
		}
		return -ll
	}

	// Nelder-Mead 简易实现（3 维 -> 4 顶点）。
	x0 := [3]float64{
		math.Log(variance * 0.05),
		invSigmoid(0.05 / 0.5),
		invSigmoid(0.9 / 0.99),
	}
	best, ok := nelderMead3(negLL, x0, 500, 1e-8)
	if !ok {
		return nil, nil, ErrNonConvergent
	}
	params := &GARCHParams{
		Omega: math.Exp(best[0]),
		Alpha: sigmoid(best[1]) * 0.5,
		Beta:  sigmoid(best[2]) * 0.99,
	}
	if !(params.Persistence() < 1) {
		return nil, nil, ErrNonConvergent
	}
	// 还原条件方差序列。
	condVar := make([]float64, n)
	s2 := variance
	for i, r := range demeaned {
		condVar[i] = s2
		s2 = params.Omega + params.Alpha*r*r + params.Beta*s2
	}
	return params, condVar, nil
}

// ForecastGARCH 给定参数与最近一次收益、最近一次条件方差，返回下一步条件方差。
func ForecastGARCH(p *GARCHParams, lastReturn, lastVar float64) float64 {
	return p.Omega + p.Alpha*lastReturn*lastReturn + p.Beta*lastVar
}

// LogReturns 把价格序列转成对数收益。len(returns) = len(price)-1。
func LogReturns(price []float64) []float64 {
	if len(price) < 2 {
		return nil
	}
	out := make([]float64, len(price)-1)
	for i := 1; i < len(price); i++ {
		if price[i-1] <= 0 || price[i] <= 0 {
			continue
		}
		out[i-1] = math.Log(price[i] / price[i-1])
	}
	return out
}

// AnnualizationFactor 给 timeframe 的年化因子（每年的 K 线数）。
// 默认 1d=252；其他主流周期亦给出常用值。
func AnnualizationFactor(timeframe string) float64 {
	switch timeframe {
	case "1m":
		return 252 * 24 * 60
	case "5m":
		return 252 * 24 * 12
	case "15m":
		return 252 * 24 * 4
	case "1h":
		return 252 * 24
	case "4h":
		return 252 * 6
	case "", "1d", "d", "daily":
		return 252
	case "1w", "w", "weekly":
		return 52
	}
	return 252
}

// sigmoid / invSigmoid 用作参数无约束化。
func sigmoid(x float64) float64 { return 1.0 / (1.0 + math.Exp(-x)) }
func invSigmoid(p float64) float64 {
	if p <= 0 {
		p = 1e-9
	} else if p >= 1 {
		p = 1 - 1e-9
	}
	return math.Log(p / (1 - p))
}

// nmPoint Nelder-Mead 单纯形顶点。
type nmPoint struct {
	x [3]float64
	v float64
}

// nelderMead3 三维 Nelder-Mead，返回最优点与是否收敛。
// 参数：f 目标函数（最小化）、x0 初值、maxIter、tol。
// 这是一个紧凑、稳定的实现；金融 GARCH 一般 < 200 步可收敛。
func nelderMead3(f func([3]float64) float64, x0 [3]float64, maxIter int, tol float64) ([3]float64, bool) {
	const (
		alpha = 1.0
		gamma = 2.0
		rho   = 0.5
		sigma = 0.5
		step  = 0.5
	)
	simplex := make([]nmPoint, 4)
	simplex[0] = nmPoint{x0, f(x0)}
	for i := 0; i < 3; i++ {
		x := x0
		x[i] += step
		simplex[i+1] = nmPoint{x, f(x)}
	}
	for iter := 0; iter < maxIter; iter++ {
		// 排序：最小在 0，最大在末。
		sort4(simplex)
		// 收敛：极差小于 tol。
		if math.Abs(simplex[3].v-simplex[0].v) < tol {
			return simplex[0].x, true
		}
		// 计算重心（不含最差点）。
		var c [3]float64
		for i := 0; i < 3; i++ {
			c[0] += simplex[i].x[0]
			c[1] += simplex[i].x[1]
			c[2] += simplex[i].x[2]
		}
		c[0] /= 3
		c[1] /= 3
		c[2] /= 3
		// 反射
		xr := [3]float64{
			c[0] + alpha*(c[0]-simplex[3].x[0]),
			c[1] + alpha*(c[1]-simplex[3].x[1]),
			c[2] + alpha*(c[2]-simplex[3].x[2]),
		}
		fr := f(xr)
		if fr < simplex[2].v && fr >= simplex[0].v {
			simplex[3] = nmPoint{xr, fr}
			continue
		}
		if fr < simplex[0].v {
			// 扩张
			xe := [3]float64{
				c[0] + gamma*(xr[0]-c[0]),
				c[1] + gamma*(xr[1]-c[1]),
				c[2] + gamma*(xr[2]-c[2]),
			}
			fe := f(xe)
			if fe < fr {
				simplex[3] = nmPoint{xe, fe}
			} else {
				simplex[3] = nmPoint{xr, fr}
			}
			continue
		}
		// 收缩
		xc := [3]float64{
			c[0] + rho*(simplex[3].x[0]-c[0]),
			c[1] + rho*(simplex[3].x[1]-c[1]),
			c[2] + rho*(simplex[3].x[2]-c[2]),
		}
		fc := f(xc)
		if fc < simplex[3].v {
			simplex[3] = nmPoint{xc, fc}
			continue
		}
		// 整体收缩
		x0p := simplex[0].x
		for i := 1; i < 4; i++ {
			simplex[i].x[0] = x0p[0] + sigma*(simplex[i].x[0]-x0p[0])
			simplex[i].x[1] = x0p[1] + sigma*(simplex[i].x[1]-x0p[1])
			simplex[i].x[2] = x0p[2] + sigma*(simplex[i].x[2]-x0p[2])
			simplex[i].v = f(simplex[i].x)
		}
	}
	sort4(simplex)
	return simplex[0].x, math.Abs(simplex[3].v-simplex[0].v) < tol*100
}

// sort4 4 点冒泡排序，按 v 升序。
func sort4(a []nmPoint) {
	for i := 0; i < 4; i++ {
		for j := i + 1; j < 4; j++ {
			if a[j].v < a[i].v {
				a[i], a[j] = a[j], a[i]
			}
		}
	}
}
