package quant

import (
	"math"
	"math/rand"
	"testing"
)

// generateGARCH 生成已知参数的 GARCH 仿真数据，用于回归测试。
func generateGARCH(n int, omega, alpha, beta float64, seed int64) []float64 {
	rng := rand.New(rand.NewSource(seed))
	r := make([]float64, n)
	s2 := omega / (1 - alpha - beta)
	for i := 0; i < n; i++ {
		s := math.Sqrt(s2)
		r[i] = rng.NormFloat64() * s
		s2 = omega + alpha*r[i]*r[i] + beta*s2
	}
	return r
}

func TestFitGARCH_RecoversParams(t *testing.T) {
	// 真值 omega=2e-6 alpha=0.08 beta=0.9 typical日频参数
	const om, al, be = 2e-6, 0.08, 0.9
	data := generateGARCH(2000, om, al, be, 42)
	p, condVar, err := FitGARCH(data)
	if err != nil {
		t.Fatalf("FitGARCH failed: %v", err)
	}
	if len(condVar) != len(data) {
		t.Fatalf("condVar length mismatch")
	}
	if !(p.Persistence() < 1) {
		t.Fatalf("persistence must be < 1, got %v", p.Persistence())
	}
	// 数值优化 + 有限样本：使用宽容差，但参数必须在合理量级。
	if p.Alpha < 0.01 || p.Alpha > 0.5 {
		t.Errorf("alpha out of plausible range: %v", p.Alpha)
	}
	if p.Beta < 0.5 || p.Beta > 0.99 {
		t.Errorf("beta out of plausible range: %v", p.Beta)
	}
	if p.Omega <= 0 {
		t.Errorf("omega must be positive: %v", p.Omega)
	}
}

func TestFitGARCH_InsufficientData(t *testing.T) {
	if _, _, err := FitGARCH([]float64{1, 2, 3}); err != ErrInsufficientData {
		t.Fatalf("expected ErrInsufficientData, got %v", err)
	}
}

func TestHurst_WhiteNoise(t *testing.T) {
	// 注：HurstRS 输入应是“增量序列”（如收益率），而非累积价格。
	// 对独立同分布白噪声，H 理论值 0.5。
	rng := rand.New(rand.NewSource(7))
	n := 2048
	x := make([]float64, n)
	for i := 0; i < n; i++ {
		x[i] = rng.NormFloat64()
	}
	res, err := HurstRS(x)
	if err != nil {
		t.Fatalf("HurstRS err: %v", err)
	}
	// 估计器有偏，允许 0.35–0.65。
	if res.H < 0.35 || res.H > 0.65 {
		t.Errorf("white noise H out of range: %v", res.H)
	}
}

func TestHurst_Trending(t *testing.T) {
	// 强趋势：纯线性 + 噪声
	rng := rand.New(rand.NewSource(11))
	n := 1024
	x := make([]float64, n)
	for i := 0; i < n; i++ {
		x[i] = float64(i)*0.1 + rng.NormFloat64()*0.05
	}
	res, _ := HurstRS(x)
	if res.H < 0.7 {
		t.Errorf("expected H>0.7 for trending, got %v", res.H)
	}
	if res.Interpretation != "trending" {
		t.Errorf("expected trending, got %v", res.Interpretation)
	}
}

func TestHMM_TwoStateGaussianRecovery(t *testing.T) {
	// 真值：状态 0 N(-0.5, 0.5)；状态 1 N(0.8, 0.3)；自相关高。
	rng := rand.New(rand.NewSource(99))
	n := 1500
	state := 0
	obs := make([]float64, n)
	for i := 0; i < n; i++ {
		// 90% 维持，10% 跳到对方
		if rng.Float64() < 0.1 {
			state = 1 - state
		}
		if state == 0 {
			obs[i] = rng.NormFloat64()*0.5 + (-0.5)
		} else {
			obs[i] = rng.NormFloat64()*0.3 + 0.8
		}
	}
	model, err := FitGaussianHMM(obs, 2, 7, 200)
	if err != nil {
		t.Fatalf("HMM fit failed: %v", err)
	}
	// 至少一个状态的 mu 应接近 -0.5，另一个接近 0.8（不区分状态编号）。
	mu := model.Mu
	near := func(a, b, tol float64) bool { return math.Abs(a-b) < tol }
	hit1 := near(mu[0], -0.5, 0.4) || near(mu[1], -0.5, 0.4)
	hit2 := near(mu[0], 0.8, 0.4) || near(mu[1], 0.8, 0.4)
	if !(hit1 && hit2) {
		t.Errorf("HMM did not recover means; got mu=%v", mu)
	}
	path, err := model.Decode(obs)
	if err != nil {
		t.Fatalf("Viterbi failed: %v", err)
	}
	if len(path) != n {
		t.Errorf("path length mismatch")
	}
}

func TestPearsonCorr(t *testing.T) {
	a := []float64{1, 2, 3, 4, 5}
	b := []float64{2, 4, 6, 8, 10}
	if c := PearsonCorr(a, b); math.Abs(c-1.0) > 1e-9 {
		t.Errorf("expected 1.0, got %v", c)
	}
	c := []float64{5, 4, 3, 2, 1}
	if v := PearsonCorr(a, c); math.Abs(v+1.0) > 1e-9 {
		t.Errorf("expected -1.0, got %v", v)
	}
}

func TestRollingCorrelationMatrix_Diagonal(t *testing.T) {
	s := [][]float64{{1, 2, 3, 4}, {2, 1, 3, 4}, {0, 0, 1, 1}}
	m := RollingCorrelationMatrix(s, 0)
	for i := range m {
		if math.Abs(m[i][i]-1.0) > 1e-9 {
			t.Errorf("diagonal[%d]=%v", i, m[i][i])
		}
	}
}

func TestRSI(t *testing.T) {
	// 单调递增价格 → RSI 接近 100
	x := make([]float64, 30)
	for i := range x {
		x[i] = float64(i + 1)
	}
	r := RSI(x, 14)
	if r[len(r)-1] < 99 {
		t.Errorf("monotonic up should yield RSI~100, got %v", r[len(r)-1])
	}
}

func TestFindDivergences_Bearish(t *testing.T) {
	// 价格创新高，RSI 没创新高 → bearish 背离
	close := make([]float64, 60)
	rsi := make([]float64, 60)
	for i := 0; i < 60; i++ {
		close[i] = float64(i)
		rsi[i] = 50
	}
	// 制造两个高点：i=20 与 i=50；价格更高，RSI 更低
	close[20] = 100
	close[50] = 110
	rsi[20] = 80
	rsi[50] = 70
	out := FindDivergences(close, rsi, "rsi", 5, 60)
	if len(out) == 0 {
		t.Fatalf("expected at least one divergence")
	}
	if out[0].Kind != BearishDivergence {
		t.Errorf("expected bearish, got %v", out[0].Kind)
	}
}
