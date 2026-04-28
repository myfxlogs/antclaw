package backtest

import (
	"math"
	"testing"
	"time"
)

func TestSharpeBasic(t *testing.T) {
	rets := []float64{0.001, 0.002, -0.001, 0.003, 0.0005, -0.0008, 0.0015}
	got := sharpe(rets)
	if math.IsNaN(got) || math.IsInf(got, 0) {
		t.Fatalf("sharpe NaN/Inf: %v", got)
	}
	if got <= 0 {
		t.Fatalf("expected positive sharpe for positively skewed series, got %v", got)
	}
}

func TestSharpeInsufficient(t *testing.T) {
	if v := sharpe([]float64{0.01}); v != 0 {
		t.Fatalf("expected 0 for insufficient sample, got %v", v)
	}
}

func TestSMACrossoverReturnsLength(t *testing.T) {
	closes := make([]float64, 200)
	for i := range closes {
		closes[i] = 100 + float64(i)*0.5
	}
	rets := smaCrossoverReturns(closes, 5, 30)
	if len(rets) == 0 {
		t.Fatal("expected non-empty returns on trending series")
	}
}

func TestSMACrossoverInvalidParams(t *testing.T) {
	closes := make([]float64, 100)
	if r := smaCrossoverReturns(closes, 30, 5); r != nil {
		t.Fatal("short>=long must return nil")
	}
	if r := smaCrossoverReturns(closes[:10], 5, 30); r != nil {
		t.Fatal("insufficient bars must return nil")
	}
}

func TestRunWalkforwardSMAUptrend(t *testing.T) {
	// 构造一个 600 根线、整体上升趋势的合成日线序列
	now := time.Now().UTC()
	ps := &priceSeries{Times: make([]time.Time, 600), Closes: make([]float64, 600)}
	for i := 0; i < 600; i++ {
		ps.Times[i] = now.AddDate(0, 0, -600+i)
		ps.Closes[i] = 100 + float64(i)*0.3 + math.Sin(float64(i)/10.0)
	}
	folds, err := runWalkforwardSMA(ps, 5, 0.7)
	if err != nil {
		t.Fatalf("walkforward error: %v", err)
	}
	if len(folds) != 5 {
		t.Fatalf("expected 5 folds, got %d", len(folds))
	}
	for _, f := range folds {
		if f.BestShort >= f.BestLong {
			t.Fatalf("invalid params short=%d long=%d", f.BestShort, f.BestLong)
		}
		if f.TestFrom.Before(f.TrainFrom) {
			t.Fatal("test must follow train")
		}
	}
}

func TestRunWalkforwardSMAInsufficient(t *testing.T) {
	ps := &priceSeries{Closes: make([]float64, 50), Times: make([]time.Time, 50)}
	if _, err := runWalkforwardSMA(ps, 5, 0.7); err == nil {
		t.Fatal("expected error on insufficient bars")
	}
}

func TestBootstrapPercentilesOrdered(t *testing.T) {
	v := []float64{0.5, 0.7, 0.9, 1.1, 1.3, 1.5}
	p5, p50, p95 := bootstrapPercentiles(v, 1000, 42)
	if !(p5 <= p50 && p50 <= p95) {
		t.Fatalf("percentile order violated: p5=%v p50=%v p95=%v", p5, p50, p95)
	}
	if p5 < 0.4 || p95 > 1.6 {
		t.Fatalf("percentiles outside reasonable range: p5=%v p95=%v", p5, p95)
	}
}

func TestBootstrapPercentilesEmpty(t *testing.T) {
	p5, p50, p95 := bootstrapPercentiles(nil, 100, 1)
	if p5 != 0 || p50 != 0 || p95 != 0 {
		t.Fatal("empty input must yield zeros")
	}
}
