// 三套独立回测引擎：quantbt（TSMOM）/ vpbt（成交量轮廓）/ cta（Donchian 突破）。
//
// 共同输入：从 price_daily 读 OHLC（vpbt 需要 volume）。
// 共同输出：写入 backtest_trades，并在内存 tasks 中保存 BacktestMetrics。
package backtest

import (
	"context"
	"fmt"
	"math"
	"time"

	backtestv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/jackc/pgx/v5/pgxpool"
)

// barsFull 完整 OHLCV 序列，用于 VP 引擎。
type barsFull struct {
	Times  []time.Time
	Open   []float64
	High   []float64
	Low    []float64
	Close  []float64
	Volume []float64
}

// loadDailyOHLCV 拉 OHLCV（VP 引擎使用）。
func loadDailyOHLCV(ctx context.Context, pool *pgxpool.Pool, symbol string, from, to time.Time) (*barsFull, error) {
	rows, err := pool.Query(ctx, `
		SELECT time, open, high, low, close, COALESCE(volume,0)
		  FROM price_daily WHERE symbol=$1 AND time BETWEEN $2 AND $3 AND close > 0
		 ORDER BY time ASC`, symbol, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := &barsFull{}
	for rows.Next() {
		var t time.Time
		var o, h, l, c, v float64
		if err := rows.Scan(&t, &o, &h, &l, &c, &v); err != nil {
			return nil, err
		}
		out.Times = append(out.Times, t)
		out.Open = append(out.Open, o)
		out.High = append(out.High, h)
		out.Low = append(out.Low, l)
		out.Close = append(out.Close, c)
		out.Volume = append(out.Volume, v)
	}
	return out, nil
}

// ============================================================
// quantbt — TSMOM (Time-Series Momentum)
// 信号：sign(close[i] - close[i - lookback])，月度调仓近似（每 21 根 bar 重算）。
// ============================================================
func quantbtSignals(closes []float64, lookback, rebalance int) []int {
	n := len(closes)
	if n <= lookback || rebalance <= 0 {
		return nil
	}
	sigs := make([]int, n-1)
	cur := 0
	for i := lookback; i < n-1; i++ {
		if (i-lookback)%rebalance == 0 {
			diff := closes[i] - closes[i-lookback]
			switch {
			case diff > 0:
				cur = 1
			case diff < 0:
				cur = -1
			default:
				cur = 0
			}
		}
		sigs[i] = cur
	}
	return sigs
}

// ============================================================
// vpbt — Volume Profile breakout
// POC = volume 最高的价格分箱中点；价格站上 POC → 多；跌破 → 空。
// 滚动窗 60 根，箱数 20。
// ============================================================
func vpbtSignals(b *barsFull, window, bins int) []int {
	n := len(b.Close)
	if n <= window+1 {
		return nil
	}
	sigs := make([]int, n-1)
	for i := window; i < n-1; i++ {
		from := i - window
		to := i
		minP, maxP := b.Low[from], b.High[from]
		for k := from; k < to; k++ {
			if b.Low[k] < minP {
				minP = b.Low[k]
			}
			if b.High[k] > maxP {
				maxP = b.High[k]
			}
		}
		if maxP <= minP {
			continue
		}
		binW := (maxP - minP) / float64(bins)
		volByBin := make([]float64, bins)
		for k := from; k < to; k++ {
			mid := (b.High[k] + b.Low[k]) / 2
			idx := int((mid - minP) / binW)
			if idx >= bins {
				idx = bins - 1
			}
			if idx < 0 {
				idx = 0
			}
			volByBin[idx] += b.Volume[k]
		}
		// POC 中点
		pocIdx := 0
		for k := 1; k < bins; k++ {
			if volByBin[k] > volByBin[pocIdx] {
				pocIdx = k
			}
		}
		poc := minP + (float64(pocIdx)+0.5)*binW
		switch {
		case b.Close[i] > poc:
			sigs[i] = 1
		case b.Close[i] < poc:
			sigs[i] = -1
		}
	}
	return sigs
}

// ============================================================
// cta — Donchian breakout (N 周期高/低)
// 收盘破 N 周期高 → 多；破 N 周期低 → 空。
// ============================================================
func ctaSignals(closes []float64, n int) []int {
	N := len(closes)
	if N <= n+1 {
		return nil
	}
	sigs := make([]int, N-1)
	cur := 0
	for i := n; i < N-1; i++ {
		hi := closes[i-n]
		lo := closes[i-n]
		for k := i - n + 1; k <= i; k++ {
			if closes[k] > hi {
				hi = closes[k]
			}
			if closes[k] < lo {
				lo = closes[k]
			}
		}
		switch {
		case closes[i] >= hi-1e-12:
			cur = 1
		case closes[i] <= lo+1e-12:
			cur = -1
		}
		sigs[i] = cur
	}
	return sigs
}

// ============================================================
// runEngine 通用入口
// ============================================================

func (s *Service) runEngine(ctx context.Context, engine, taskID, pair string, from, to time.Time) error {
	if s.pool == nil {
		return fmt.Errorf("backtest engine: pool not configured")
	}
	bf, err := loadDailyOHLCV(ctx, s.pool, pair, from, to)
	if err != nil {
		return err
	}
	if len(bf.Close) < 80 {
		return fmt.Errorf("backtest engine %s: insufficient bars (%d)", engine, len(bf.Close))
	}
	var sigs []int
	switch engine {
	case "quantbt":
		sigs = quantbtSignals(bf.Close, 252/4, 21) // 3 个月动量 + 月调仓
	case "vpbt":
		sigs = vpbtSignals(bf, 60, 20)
	case "cta":
		sigs = ctaSignals(bf.Close, 20) // 20 日 Donchian
	default:
		return fmt.Errorf("unknown engine: %s", engine)
	}
	regimeForBar := make([]string, len(bf.Close))
	for i := range regimeForBar {
		regimeForBar[i] = ""
	}
	trades := extractAdvTrades(bf.Times, bf.Close, sigs, DefaultCost(), regimeForBar)
	// 重新编号
	for i := range trades {
		trades[i].Seq = i + 1
	}
	_ = SaveTrades(ctx, s.pool, taskID, trades)
	// 计算指标
	metrics := buildMetricsFromTrades(trades)
	s.tasks[taskID] = &backtestv1.GetBacktestResponse{
		TaskId: taskID, Status: "done",
		Config:  &backtestv1.BacktestConfig{Pair: pair},
		Metrics: metrics,
	}
	return nil
}

func buildMetricsFromTrades(trades []AdvTrade) *backtestv1.BacktestMetrics {
	if len(trades) == 0 {
		return &backtestv1.BacktestMetrics{}
	}
	var totalReturn, sumPos, sumNeg float64
	wins := 0
	pnls := make([]float64, len(trades))
	for i, t := range trades {
		pnls[i] = t.PnL
		totalReturn += t.PnL
		if t.PnL > 0 {
			sumPos += t.PnL
			wins++
		} else if t.PnL < 0 {
			sumNeg += -t.PnL
		}
	}
	mean := totalReturn / float64(len(trades))
	var v float64
	for _, p := range pnls {
		v += (p - mean) * (p - mean)
	}
	std := math.Sqrt(v / float64(len(pnls)))
	sharpe := 0.0
	if std > 0 {
		sharpe = mean / std * math.Sqrt(252)
	}
	pf := 0.0
	if sumNeg > 0 {
		pf = sumPos / sumNeg
	}
	return &backtestv1.BacktestMetrics{
		TotalReturn:    fmt.Sprintf("%.4f", totalReturn),
		TotalReturnPct: fmt.Sprintf("%.2f%%", totalReturn*100),
		SharpeRatio:    sharpe,
		MaxDrawdown:    computeMaxDrawdown(pnls),
		WinRate:        float64(wins) / float64(len(trades)),
		TotalTrades:    int32(len(trades)),
		ProfitFactor:   pf,
	}
}
