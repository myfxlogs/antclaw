package calibration

import (
	"math"
	"math/rand"
	"testing"
)

// 生成有偏分类器：score 越高越倾向 outcome=true，但带噪声。
func synthetic(n int, seed int64) ([]float64, []bool) {
	rng := rand.New(rand.NewSource(seed))
	scores := make([]float64, n)
	outcomes := make([]bool, n)
	for i := 0; i < n; i++ {
		s := rng.NormFloat64()
		// p = sigmoid(2s - 0.5)
		p := 1.0 / (1.0 + math.Exp(-(2*s - 0.5)))
		scores[i] = s
		outcomes[i] = rng.Float64() < p
	}
	return scores, outcomes
}

func TestPlatt_BrierBetterThanRaw(t *testing.T) {
	scores, outcomes := synthetic(2000, 7)
	c := NewPlatt()
	if err := c.Fit(scores, outcomes); err != nil {
		t.Fatalf("Platt fit: %v", err)
	}
	br := c.Brier(scores, outcomes)
	if !(br < 0.25) {
		t.Errorf("Platt brier should beat naive 0.25, got %v", br)
	}
}

func TestIsotonic_BrierBetterThanPlatt(t *testing.T) {
	// Isotonic 在样本足够大时应至少不输给 Platt。
	scores, outcomes := synthetic(3000, 9)
	platt := NewPlatt()
	iso := NewIsotonic()
	_ = platt.Fit(scores, outcomes)
	_ = iso.Fit(scores, outcomes)
	bp := platt.Brier(scores, outcomes)
	bi := iso.Brier(scores, outcomes)
	// 允许少量浮动；只要 isotonic 不显著差就算 OK。
	if bi > bp+0.02 {
		t.Errorf("Isotonic brier %v worse than Platt %v", bi, bp)
	}
}

func TestPlatt_Monotonic(t *testing.T) {
	scores, outcomes := synthetic(1000, 3)
	c := NewPlatt()
	_ = c.Fit(scores, outcomes)
	// 拟合后 Predict 在 score 上应单调（A<0 时单增）。
	prev := c.Predict(-3)
	for x := -2.5; x <= 3; x += 0.25 {
		cur := c.Predict(x)
		if cur < prev-1e-9 {
			t.Errorf("Platt not monotonic at x=%v: %v -> %v", x, prev, cur)
		}
		prev = cur
	}
}
