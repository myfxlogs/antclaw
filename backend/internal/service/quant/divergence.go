package quant

import "math"

// DivergenceKind 背离类型。
type DivergenceKind string

const (
	BullishDivergence DivergenceKind = "bullish"
	BearishDivergence DivergenceKind = "bearish"
)

// Divergence 单条背离事件。
type Divergence struct {
	Indicator      string         // rsi / obv / macd
	Kind           DivergenceKind // bullish / bearish
	Index          int            // 命中 K 线的右侧索引
	PricePivot     float64
	IndicatorPivot float64
	Note           string
}

// FindDivergences 在价格 close 与指标 ind 上找经典背离：
//   - bullish：价格创近窗低点，指标却抬高
//   - bearish：价格创近窗高点，指标却走低
//
// pivotWindow 默认 5。lookback 默认 60。
//
// 实现是经典“两个相邻枢轴点对比”法：
//  1. 找局部极值（左右各 pivotWindow 根更高/更低）
//  2. 取最近两个同类极值，做价格 vs 指标方向对比
//  3. 同向无背离；反向记一条
func FindDivergences(close, ind []float64, indicator string, pivotWindow, lookback int) []Divergence {
	if pivotWindow <= 0 {
		pivotWindow = 5
	}
	if lookback <= 0 {
		lookback = 60
	}
	if len(close) != len(ind) || len(close) < pivotWindow*2+5 {
		return nil
	}
	from := len(close) - lookback
	if from < pivotWindow {
		from = pivotWindow
	}
	var out []Divergence
	// 高点 / 低点列表。
	var highs, lows []int
	for i := from; i < len(close)-pivotWindow; i++ {
		if isLocalHigh(close, i, pivotWindow) {
			highs = append(highs, i)
		}
		if isLocalLow(close, i, pivotWindow) {
			lows = append(lows, i)
		}
	}
	// bearish: 最近两个高点。
	if len(highs) >= 2 {
		a := highs[len(highs)-2]
		b := highs[len(highs)-1]
		if close[b] > close[a] && ind[b] < ind[a] {
			out = append(out, Divergence{
				Indicator: indicator, Kind: BearishDivergence,
				Index: b, PricePivot: close[b], IndicatorPivot: ind[b],
				Note: "Higher price high with lower indicator high",
			})
		}
	}
	// bullish: 最近两个低点。
	if len(lows) >= 2 {
		a := lows[len(lows)-2]
		b := lows[len(lows)-1]
		if close[b] < close[a] && ind[b] > ind[a] {
			out = append(out, Divergence{
				Indicator: indicator, Kind: BullishDivergence,
				Index: b, PricePivot: close[b], IndicatorPivot: ind[b],
				Note: "Lower price low with higher indicator low",
			})
		}
	}
	return out
}

func isLocalHigh(x []float64, i, w int) bool {
	for k := 1; k <= w; k++ {
		if i-k < 0 || i+k >= len(x) {
			return false
		}
		if !(x[i] > x[i-k]) || !(x[i] > x[i+k]) {
			return false
		}
	}
	return true
}

func isLocalLow(x []float64, i, w int) bool {
	for k := 1; k <= w; k++ {
		if i-k < 0 || i+k >= len(x) {
			return false
		}
		if !(x[i] < x[i-k]) || !(x[i] < x[i+k]) {
			return false
		}
	}
	return true
}

// RSI 经典 14 周期 RSI；返回与 close 等长的 RSI 序列（前 period-1 个用 50 填充）。
func RSI(close []float64, period int) []float64 {
	n := len(close)
	if n < period+1 || period <= 1 {
		return nil
	}
	out := make([]float64, n)
	for i := 0; i < period; i++ {
		out[i] = 50
	}
	var gain, loss float64
	for i := 1; i <= period; i++ {
		ch := close[i] - close[i-1]
		if ch > 0 {
			gain += ch
		} else {
			loss += -ch
		}
	}
	avgG := gain / float64(period)
	avgL := loss / float64(period)
	out[period] = rsiFromAvg(avgG, avgL)
	for i := period + 1; i < n; i++ {
		ch := close[i] - close[i-1]
		g := math.Max(0, ch)
		l := math.Max(0, -ch)
		avgG = (avgG*float64(period-1) + g) / float64(period)
		avgL = (avgL*float64(period-1) + l) / float64(period)
		out[i] = rsiFromAvg(avgG, avgL)
	}
	return out
}

func rsiFromAvg(g, l float64) float64 {
	if l == 0 {
		return 100
	}
	rs := g / l
	return 100 - 100/(1+rs)
}

// OBV On-Balance Volume；与 close/volume 等长。
func OBV(close []float64, volume []int64) []float64 {
	n := len(close)
	if n != len(volume) || n == 0 {
		return nil
	}
	out := make([]float64, n)
	for i := 1; i < n; i++ {
		v := float64(volume[i])
		switch {
		case close[i] > close[i-1]:
			out[i] = out[i-1] + v
		case close[i] < close[i-1]:
			out[i] = out[i-1] - v
		default:
			out[i] = out[i-1]
		}
	}
	return out
}

// MACDLine 经典 (12,26,9)；返回 macd line（不含 signal）。
func MACDLine(close []float64, fast, slow int) []float64 {
	if fast <= 0 {
		fast = 12
	}
	if slow <= 0 {
		slow = 26
	}
	emaF := ema(close, fast)
	emaS := ema(close, slow)
	if emaF == nil || emaS == nil {
		return nil
	}
	out := make([]float64, len(close))
	for i := range out {
		out[i] = emaF[i] - emaS[i]
	}
	return out
}

func ema(x []float64, period int) []float64 {
	n := len(x)
	if n == 0 || period <= 1 {
		return nil
	}
	out := make([]float64, n)
	k := 2.0 / float64(period+1)
	out[0] = x[0]
	for i := 1; i < n; i++ {
		out[i] = x[i]*k + out[i-1]*(1-k)
	}
	return out
}
