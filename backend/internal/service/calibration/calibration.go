// Package calibration 提供 Platt 与 Isotonic 概率校准。
//
// 用法：用历史回测的 (raw_score, outcome) 对训练 Calibrator，再用 Predict 把新 raw 分映射到 [0,1] 校准概率。
package calibration

import (
	"errors"
	"math"
	"sort"
)

// Calibrator 通用接口。
type Calibrator interface {
	Type() string
	Fit(scores []float64, outcomes []bool) error
	Predict(score float64) float64
	Brier(scores []float64, outcomes []bool) float64
}

// Brier 通用 Brier score：mean((p - y)^2)，p 是 Predict 输出，y ∈ {0,1}。
func brier(c Calibrator, scores []float64, outcomes []bool) float64 {
	n := len(scores)
	if n == 0 || len(outcomes) != n {
		return math.NaN()
	}
	var s float64
	for i, x := range scores {
		p := c.Predict(x)
		y := 0.0
		if outcomes[i] {
			y = 1.0
		}
		d := p - y
		s += d * d
	}
	return s / float64(n)
}

// ============================================================
// Platt scaling: P(y=1|x) = 1 / (1 + exp(A*x + B))
// 用拟牛顿一维线搜索最大化对数似然（A,B 两参数 → Nelder-Mead 2D）。
// ============================================================

type Platt struct {
	A, B    float64
	NSample int
}

func NewPlatt() Calibrator { return &Platt{} }

func (p *Platt) Type() string { return "platt" }

func (p *Platt) Fit(scores []float64, outcomes []bool) error {
	n := len(scores)
	if n != len(outcomes) || n < 5 {
		return errors.New("calibration: insufficient samples for Platt")
	}
	// 用经验 prior 初始化 B（截距）：B = log((Nneg+1)/(Npos+1))。A 初始 -1（典型分类器）。
	pos, neg := 0, 0
	for _, y := range outcomes {
		if y {
			pos++
		} else {
			neg++
		}
	}
	if pos == 0 || neg == 0 {
		return errors.New("calibration: outcomes must contain both classes")
	}
	yPlus := float64(pos+1) / float64(pos+2)
	yMinus := 1.0 / float64(neg+2)
	negLL := func(theta [2]float64) float64 {
		A, B := theta[0], theta[1]
		var ll float64
		for i, x := range scores {
			t := yMinus
			if outcomes[i] {
				t = yPlus
			}
			fx := A*x + B
			// 数值稳定的 log(1+exp(fx))
			ll += t*fx + softplus(-fx) + (1-t)*0 // = t*fx + softplus(-fx) for cross-entropy
		}
		return ll
	}
	x0 := [2]float64{-1.0, math.Log(float64(neg+1) / float64(pos+1))}
	best, _ := nelderMead2(negLL, x0, 1000, 1e-9)
	p.A = best[0]
	p.B = best[1]
	p.NSample = n
	return nil
}

func (p *Platt) Predict(score float64) float64 {
	z := p.A*score + p.B
	return 1.0 / (1.0 + math.Exp(z))
}

func (p *Platt) Brier(scores []float64, outcomes []bool) float64 { return brier(p, scores, outcomes) }

func softplus(x float64) float64 {
	if x > 0 {
		return x + math.Log1p(math.Exp(-x))
	}
	return math.Log1p(math.Exp(x))
}

// ============================================================
// Isotonic Regression: PAV (Pool Adjacent Violators)
// 把 (score, y) 对按 score 排序，单调递增地拟合 P(y=1|score)。
// ============================================================

type Isotonic struct {
	X []float64 // 升序 score
	P []float64 // 与 X 等长的单调递增拟合概率
	N int
}

func NewIsotonic() Calibrator { return &Isotonic{} }

func (i *Isotonic) Type() string { return "isotonic" }

type pair struct {
	x float64
	y float64
}

func (i *Isotonic) Fit(scores []float64, outcomes []bool) error {
	n := len(scores)
	if n != len(outcomes) || n < 5 {
		return errors.New("calibration: insufficient samples for Isotonic")
	}
	pairs := make([]pair, n)
	for k := 0; k < n; k++ {
		y := 0.0
		if outcomes[k] {
			y = 1.0
		}
		pairs[k] = pair{x: scores[k], y: y}
	}
	sort.Slice(pairs, func(a, b int) bool { return pairs[a].x < pairs[b].x })
	// PAV
	weights := make([]float64, 0, n)
	values := make([]float64, 0, n)
	xs := make([]float64, 0, n)
	for _, p := range pairs {
		values = append(values, p.y)
		weights = append(weights, 1.0)
		xs = append(xs, p.x)
		// 合并相邻违反单调性的块
		for len(values) >= 2 && values[len(values)-2] > values[len(values)-1] {
			w1, w2 := weights[len(weights)-2], weights[len(weights)-1]
			merged := (values[len(values)-2]*w1 + values[len(values)-1]*w2) / (w1 + w2)
			values = values[:len(values)-2]
			weights = weights[:len(weights)-2]
			xs = xs[:len(xs)-1] // 保留较小 x 作为代表点
			values = append(values, merged)
			weights = append(weights, w1+w2)
		}
	}
	// 展开为按原 pairs 顺序的预测；为简化 Predict，仅保留各块代表 (xs[k], values[k])。
	i.X = xs
	i.P = values
	i.N = n
	return nil
}

func (i *Isotonic) Predict(score float64) float64 {
	if i.N == 0 {
		return 0.5
	}
	// 二分查找最右端 x ≤ score 的块
	if score <= i.X[0] {
		return clip01(i.P[0])
	}
	if score >= i.X[len(i.X)-1] {
		return clip01(i.P[len(i.P)-1])
	}
	lo, hi := 0, len(i.X)-1
	for lo+1 < hi {
		m := (lo + hi) / 2
		if i.X[m] <= score {
			lo = m
		} else {
			hi = m
		}
	}
	return clip01(i.P[lo])
}

func (i *Isotonic) Brier(scores []float64, outcomes []bool) float64 { return brier(i, scores, outcomes) }

func clip01(p float64) float64 {
	if p < 0 {
		return 0
	}
	if p > 1 {
		return 1
	}
	return p
}

// ============================================================
// 二维 Nelder-Mead 优化器（仅用于 Platt MLE）
// ============================================================

type pt2 struct {
	x [2]float64
	v float64
}

func nelderMead2(f func([2]float64) float64, x0 [2]float64, maxIter int, tol float64) ([2]float64, bool) {
	const (
		alpha = 1.0
		gamma = 2.0
		rho   = 0.5
		sigma = 0.5
		step  = 0.5
	)
	simplex := []pt2{
		{x0, f(x0)},
		{[2]float64{x0[0] + step, x0[1]}, 0},
		{[2]float64{x0[0], x0[1] + step}, 0},
	}
	simplex[1].v = f(simplex[1].x)
	simplex[2].v = f(simplex[2].x)
	for iter := 0; iter < maxIter; iter++ {
		sort.Slice(simplex, func(i, j int) bool { return simplex[i].v < simplex[j].v })
		if math.Abs(simplex[2].v-simplex[0].v) < tol {
			return simplex[0].x, true
		}
		c := [2]float64{(simplex[0].x[0] + simplex[1].x[0]) / 2, (simplex[0].x[1] + simplex[1].x[1]) / 2}
		xr := [2]float64{c[0] + alpha*(c[0]-simplex[2].x[0]), c[1] + alpha*(c[1]-simplex[2].x[1])}
		fr := f(xr)
		if fr < simplex[1].v && fr >= simplex[0].v {
			simplex[2] = pt2{xr, fr}
			continue
		}
		if fr < simplex[0].v {
			xe := [2]float64{c[0] + gamma*(xr[0]-c[0]), c[1] + gamma*(xr[1]-c[1])}
			fe := f(xe)
			if fe < fr {
				simplex[2] = pt2{xe, fe}
			} else {
				simplex[2] = pt2{xr, fr}
			}
			continue
		}
		xc := [2]float64{c[0] + rho*(simplex[2].x[0]-c[0]), c[1] + rho*(simplex[2].x[1]-c[1])}
		fc := f(xc)
		if fc < simplex[2].v {
			simplex[2] = pt2{xc, fc}
			continue
		}
		// 整体收缩
		x0p := simplex[0].x
		for k := 1; k < 3; k++ {
			simplex[k].x[0] = x0p[0] + sigma*(simplex[k].x[0]-x0p[0])
			simplex[k].x[1] = x0p[1] + sigma*(simplex[k].x[1]-x0p[1])
			simplex[k].v = f(simplex[k].x)
		}
	}
	sort.Slice(simplex, func(i, j int) bool { return simplex[i].v < simplex[j].v })
	return simplex[0].x, false
}