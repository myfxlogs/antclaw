package quant

import (
	"errors"
	"math"
	"math/rand"
)

// GaussianHMM 高斯发射 HMM：N 状态，每状态一个 (mu, sigma) 高斯。
type GaussianHMM struct {
	N     int         // 状态数
	A     [][]float64 // 转移矩阵 [N][N]
	Pi    []float64   // 初始分布 [N]
	Mu    []float64   // 各状态高斯均值
	Sigma []float64   // 各状态高斯标准差（>0）
	LogL  float64     // 训练终止时对数似然
}

// FitGaussianHMM Baum-Welch 训练。返回拟合后的模型。
//
// 参数：
//   - obs：一维观测序列（如对数收益）
//   - nStates：状态数（推荐 2 或 3）
//   - seed：随机种子（用于初值）
//   - maxIter：最大迭代次数
//
// 收敛条件：相邻迭代对数似然增量 < 1e-6 或达到 maxIter。
//
// 失败场景：
//   - len(obs) < nStates*30：ErrInsufficientData
//   - sigma 收敛到 ~0（退化簇）：ErrNonConvergent
func FitGaussianHMM(obs []float64, nStates int, seed int64, maxIter int) (*GaussianHMM, error) {
	if nStates < 2 {
		return nil, errors.New("quant: HMM nStates must be >= 2")
	}
	T := len(obs)
	if T < nStates*30 {
		return nil, ErrInsufficientData
	}
	if maxIter <= 0 {
		maxIter = 200
	}
	rng := rand.New(rand.NewSource(seed))

	// ---------- 初始化 ----------
	// Mu: 用观测的分位点；Sigma: 用样本标准差；A 初始为均匀；Pi 初始为均匀。
	mu := make([]float64, nStates)
	sigma := make([]float64, nStates)
	mean, sd := meanStd(obs)
	for i := 0; i < nStates; i++ {
		// 在 mean 周围按状态序号偏移；保留少量随机扰动以避免对称局部极小。
		offset := (float64(i) - float64(nStates-1)/2) * sd
		mu[i] = mean + offset + rng.NormFloat64()*sd*0.05
		sigma[i] = sd
	}
	A := make([][]float64, nStates)
	for i := range A {
		A[i] = make([]float64, nStates)
		for j := range A[i] {
			if i == j {
				A[i][j] = 0.7
			} else {
				A[i][j] = 0.3 / float64(nStates-1)
			}
		}
	}
	pi := make([]float64, nStates)
	for i := range pi {
		pi[i] = 1.0 / float64(nStates)
	}

	prevLL := math.Inf(-1)
	for it := 0; it < maxIter; it++ {
		// ---------- E-step ----------
		// 发射概率矩阵 B[T][N]
		B := make([][]float64, T)
		for t := 0; t < T; t++ {
			B[t] = make([]float64, nStates)
			for i := 0; i < nStates; i++ {
				B[t][i] = gaussianPDF(obs[t], mu[i], sigma[i])
			}
		}
		// 前向 alpha[T][N] + scale c[T]（数值稳定）
		alpha := make([][]float64, T)
		c := make([]float64, T)
		alpha[0] = make([]float64, nStates)
		var s0 float64
		for i := 0; i < nStates; i++ {
			alpha[0][i] = pi[i] * B[0][i]
			s0 += alpha[0][i]
		}
		if s0 == 0 {
			return nil, ErrNonConvergent
		}
		c[0] = 1 / s0
		for i := 0; i < nStates; i++ {
			alpha[0][i] *= c[0]
		}
		for t := 1; t < T; t++ {
			alpha[t] = make([]float64, nStates)
			var st float64
			for j := 0; j < nStates; j++ {
				var sum float64
				for i := 0; i < nStates; i++ {
					sum += alpha[t-1][i] * A[i][j]
				}
				alpha[t][j] = sum * B[t][j]
				st += alpha[t][j]
			}
			if st == 0 {
				return nil, ErrNonConvergent
			}
			c[t] = 1 / st
			for j := 0; j < nStates; j++ {
				alpha[t][j] *= c[t]
			}
		}
		// 对数似然 = -sum(log c[t])
		ll := 0.0
		for t := 0; t < T; t++ {
			ll -= math.Log(c[t])
		}
		// 后向 beta[T][N]
		beta := make([][]float64, T)
		beta[T-1] = make([]float64, nStates)
		for i := 0; i < nStates; i++ {
			beta[T-1][i] = c[T-1]
		}
		for t := T - 2; t >= 0; t-- {
			beta[t] = make([]float64, nStates)
			for i := 0; i < nStates; i++ {
				var sum float64
				for j := 0; j < nStates; j++ {
					sum += A[i][j] * B[t+1][j] * beta[t+1][j]
				}
				beta[t][i] = sum * c[t]
			}
		}
		// gamma[T][N], xi sum [N][N]
		gamma := make([][]float64, T)
		xi := make([][]float64, nStates)
		for i := range xi {
			xi[i] = make([]float64, nStates)
		}
		for t := 0; t < T; t++ {
			gamma[t] = make([]float64, nStates)
			var z float64
			for i := 0; i < nStates; i++ {
				gamma[t][i] = alpha[t][i] * beta[t][i] / c[t]
				z += gamma[t][i]
			}
			if z > 0 {
				for i := 0; i < nStates; i++ {
					gamma[t][i] /= z
				}
			}
			if t < T-1 {
				var zxi float64
				tmp := make([][]float64, nStates)
				for i := 0; i < nStates; i++ {
					tmp[i] = make([]float64, nStates)
					for j := 0; j < nStates; j++ {
						tmp[i][j] = alpha[t][i] * A[i][j] * B[t+1][j] * beta[t+1][j]
						zxi += tmp[i][j]
					}
				}
				if zxi > 0 {
					for i := 0; i < nStates; i++ {
						for j := 0; j < nStates; j++ {
							xi[i][j] += tmp[i][j] / zxi
						}
					}
				}
			}
		}

		// ---------- M-step ----------
		// pi
		for i := 0; i < nStates; i++ {
			pi[i] = gamma[0][i]
		}
		// A
		for i := 0; i < nStates; i++ {
			var rowSum float64
			for t := 0; t < T-1; t++ {
				rowSum += gamma[t][i]
			}
			if rowSum == 0 {
				continue
			}
			for j := 0; j < nStates; j++ {
				A[i][j] = xi[i][j] / rowSum
			}
		}
		// mu, sigma
		newMu := make([]float64, nStates)
		newSig := make([]float64, nStates)
		for i := 0; i < nStates; i++ {
			var num, den float64
			for t := 0; t < T; t++ {
				num += gamma[t][i] * obs[t]
				den += gamma[t][i]
			}
			if den == 0 {
				return nil, ErrNonConvergent
			}
			newMu[i] = num / den
			var v float64
			for t := 0; t < T; t++ {
				d := obs[t] - newMu[i]
				v += gamma[t][i] * d * d
			}
			newSig[i] = math.Sqrt(v/den) + 1e-9
		}
		mu, sigma = newMu, newSig

		// 收敛检测
		if math.Abs(ll-prevLL) < 1e-6 {
			return &GaussianHMM{N: nStates, A: A, Pi: pi, Mu: mu, Sigma: sigma, LogL: ll}, nil
		}
		prevLL = ll
	}
	return &GaussianHMM{N: nStates, A: A, Pi: pi, Mu: mu, Sigma: sigma, LogL: prevLL}, nil
}

// Decode Viterbi 解码：返回最可能状态序列。
func (m *GaussianHMM) Decode(obs []float64) ([]int, error) {
	T := len(obs)
	if T == 0 {
		return nil, errors.New("quant: empty observation")
	}
	N := m.N
	logA := make([][]float64, N)
	for i := range logA {
		logA[i] = make([]float64, N)
		for j := range logA[i] {
			logA[i][j] = math.Log(m.A[i][j] + 1e-300)
		}
	}
	logPi := make([]float64, N)
	for i := range logPi {
		logPi[i] = math.Log(m.Pi[i] + 1e-300)
	}
	logEmit := func(t, i int) float64 {
		return math.Log(gaussianPDF(obs[t], m.Mu[i], m.Sigma[i]) + 1e-300)
	}
	delta := make([][]float64, T)
	psi := make([][]int, T)
	delta[0] = make([]float64, N)
	psi[0] = make([]int, N)
	for i := 0; i < N; i++ {
		delta[0][i] = logPi[i] + logEmit(0, i)
	}
	for t := 1; t < T; t++ {
		delta[t] = make([]float64, N)
		psi[t] = make([]int, N)
		for j := 0; j < N; j++ {
			best := math.Inf(-1)
			bi := 0
			for i := 0; i < N; i++ {
				v := delta[t-1][i] + logA[i][j]
				if v > best {
					best = v
					bi = i
				}
			}
			delta[t][j] = best + logEmit(t, j)
			psi[t][j] = bi
		}
	}
	// 回溯
	path := make([]int, T)
	last := 0
	bestV := math.Inf(-1)
	for i := 0; i < N; i++ {
		if delta[T-1][i] > bestV {
			bestV = delta[T-1][i]
			last = i
		}
	}
	path[T-1] = last
	for t := T - 2; t >= 0; t-- {
		path[t] = psi[t+1][path[t+1]]
	}
	return path, nil
}

// Posterior 给出最末时点处于各状态的后验概率（归一化的 alpha[T-1]）。
func (m *GaussianHMM) Posterior(obs []float64) ([]float64, error) {
	T := len(obs)
	if T == 0 {
		return nil, errors.New("quant: empty observation")
	}
	N := m.N
	alpha := make([]float64, N)
	c := 0.0
	for i := 0; i < N; i++ {
		alpha[i] = m.Pi[i] * gaussianPDF(obs[0], m.Mu[i], m.Sigma[i])
		c += alpha[i]
	}
	if c == 0 {
		return nil, ErrNonConvergent
	}
	for i := 0; i < N; i++ {
		alpha[i] /= c
	}
	for t := 1; t < T; t++ {
		next := make([]float64, N)
		var s float64
		for j := 0; j < N; j++ {
			var sum float64
			for i := 0; i < N; i++ {
				sum += alpha[i] * m.A[i][j]
			}
			next[j] = sum * gaussianPDF(obs[t], m.Mu[j], m.Sigma[j])
			s += next[j]
		}
		if s == 0 {
			return nil, ErrNonConvergent
		}
		for j := 0; j < N; j++ {
			next[j] /= s
		}
		alpha = next
	}
	return alpha, nil
}

// gaussianPDF N(mu, sigma) 概率密度。
func gaussianPDF(x, mu, sigma float64) float64 {
	if sigma <= 0 {
		return 0
	}
	z := (x - mu) / sigma
	return math.Exp(-0.5*z*z) / (sigma * math.Sqrt(2*math.Pi))
}

// meanStd 样本均值与标准差。
func meanStd(x []float64) (float64, float64) {
	if len(x) == 0 {
		return 0, 0
	}
	var s float64
	for _, v := range x {
		s += v
	}
	m := s / float64(len(x))
	var v float64
	for _, vx := range x {
		d := vx - m
		v += d * d
	}
	return m, math.Sqrt(v / float64(len(x)))
}
