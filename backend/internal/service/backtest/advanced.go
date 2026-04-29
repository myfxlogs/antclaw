// Package backtest M-B 高级回测增量：
//
//   - 成本模型（点差 / 滑点 / 佣金）
//   - 交易级 MFE / MAE 提取
//   - Monte Carlo 价格路径（基于 GARCH 残差）
//   - 改进的 Bootstrap：从 OOS equity curve 计算真实 MaxDD
//   - 状态分层指标（HMM 状态由 quant 包提供）
package backtest

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/antclaw/antclaw/internal/service/quant"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CostModel 交易成本（以基点 bp 为单位；1bp = 0.01%）。
type CostModel struct {
	SpreadBps     float64
	SlippageBps   float64
	CommissionBps float64
}

// PerSide 单边总成本（小数表示）。
func (c CostModel) PerSide() float64 {
	return (c.SpreadBps + c.SlippageBps + c.CommissionBps) / 10000.0
}

// DefaultCost 主流货币对默认 1bp 点差 + 0.5bp 滑点 + 0.5bp 佣金 → 单边 2bp。
func DefaultCost() CostModel {
	return CostModel{SpreadBps: 1, SlippageBps: 0.5, CommissionBps: 0.5}
}

// AdvTrade 单条交易内部表示。
type AdvTrade struct {
	Seq      int
	OpenedAt time.Time
	ClosedAt time.Time
	Side     string // long / short
	Entry    float64
	Exit     float64
	PnL      float64 // 包含成本后
	PnLPct   float64
	MFE      float64
	MAE      float64
	Cost     float64
	Regime   string
}

// extractAdvTrades 给定 closes / signals（同长，signals[i] ∈ {-1,0,1}），构造交易序列。
//
// 规则：连续相同方向视为同一笔交易，在方向切换或结束时平仓。
// 计算 MFE/MAE：开仓后 high/low 区间内最大有利/不利偏离（以 entry 价为基准的小数）。
// 由于本回测当前用日线 close，无 high/low 增量精度，使用 close 区间近似——MFE = max(close-entry)/entry（多）。
func extractAdvTrades(times []time.Time, closes []float64, signals []int, cost CostModel, regimeForBar []string) []AdvTrade {
	n := len(closes)
	if n < 2 || len(signals) != n-1 || len(times) != n {
		return nil
	}
	trades := make([]AdvTrade, 0)
	cur := 0
	openIdx := -1
	for i := 0; i < n-1; i++ {
		s := signals[i]
		if s != cur {
			// 平仓 cur
			if cur != 0 && openIdx >= 0 {
				closeIdx := i
				t := makeAdvTrade(len(trades)+1, times, closes, openIdx, closeIdx, cur, cost, regimeForBar)
				trades = append(trades, t)
			}
			cur = s
			if s != 0 {
				openIdx = i
			} else {
				openIdx = -1
			}
		}
	}
	// 收尾
	if cur != 0 && openIdx >= 0 {
		t := makeAdvTrade(len(trades)+1, times, closes, openIdx, n-1, cur, cost, regimeForBar)
		trades = append(trades, t)
	}
	return trades
}

func makeAdvTrade(seq int, times []time.Time, closes []float64, openIdx, closeIdx int, side int, cost CostModel, regimeForBar []string) AdvTrade {
	entry := closes[openIdx]
	exit := closes[closeIdx]
	signMul := 1.0
	if side < 0 {
		signMul = -1.0
	}
	gross := signMul * (exit - entry) / entry
	totalCost := 2 * cost.PerSide() // 进出各一次
	net := gross - totalCost
	// MFE / MAE
	mfe, mae := 0.0, 0.0
	for k := openIdx + 1; k <= closeIdx; k++ {
		ret := signMul * (closes[k] - entry) / entry
		if ret > mfe {
			mfe = ret
		}
		if ret < mae {
			mae = ret
		}
	}
	regime := ""
	if openIdx >= 0 && openIdx < len(regimeForBar) {
		regime = regimeForBar[openIdx]
	}
	sideStr := "long"
	if side < 0 {
		sideStr = "short"
	}
	return AdvTrade{
		Seq: seq, OpenedAt: times[openIdx], ClosedAt: times[closeIdx],
		Side: sideStr, Entry: entry, Exit: exit,
		PnL: net * 1.0, PnLPct: net * 100,
		MFE: mfe, MAE: mae, Cost: totalCost, Regime: regime,
	}
}

// runWalkForwardWithTrades 在已加载的价格序列上做 walk-forward + 提取交易明细 + 状态分层。
// 返回 fold 结果 + 完整 trade 列表 + 每根 bar 的 HMM 状态映射。
func runWalkForwardWithTrades(ps *priceSeries, folds int, trainRatio float64, cost CostModel) (
	[]wfFoldResult, []AdvTrade, map[string]*AdvRegimeStats, error,
) {
	results, err := runWalkforwardSMA(ps, folds, trainRatio)
	if err != nil {
		return nil, nil, nil, err
	}
	// 估计每根 bar 的 HMM 状态（用整段对数收益）。
	regimeForBar := make([]string, len(ps.Closes))
	for i := range regimeForBar {
		regimeForBar[i] = "unknown"
	}
	rets := quant.LogReturns(ps.Closes)
	if len(rets) >= 100 {
		hmm, herr := quant.FitGaussianHMM(rets, 2, 7, 200)
		if herr == nil {
			path, derr := hmm.Decode(rets)
			if derr == nil {
				// 状态命名按 mu 升序。
				low := 0
				if hmm.Mu[1] < hmm.Mu[0] {
					low = 1
				}
				labels := []string{"risk_on", "risk_on"}
				labels[low] = "risk_off"
				// path 长度 = len(rets) = len(closes)-1；对齐到 closes（首根置 unknown）。
				for i, st := range path {
					regimeForBar[i+1] = labels[st]
				}
			}
		}
	}
	// 用 OOS 段重建信号 → 提取交易。
	allTrades := []AdvTrade{}
	for _, r := range results {
		// OOS bar 区间
		from, to := indexOf(ps.Times, r.TestFrom), indexOf(ps.Times, r.TestTo)
		if from < 0 || to < 0 || to <= from {
			continue
		}
		seg := ps.Closes[from : to+1]
		segTimes := ps.Times[from : to+1]
		segRegime := regimeForBar[from : to+1]
		signals := smaSignals(seg, r.BestShort, r.BestLong)
		t := extractAdvTrades(segTimes, seg, signals, cost, segRegime)
		// 重新编号
		for i := range t {
			t[i].Seq = len(allTrades) + i + 1
		}
		allTrades = append(allTrades, t...)
	}
	regimeStats := computeRegimeStats(allTrades)
	return results, allTrades, regimeStats, nil
}

// smaSignals 双 SMA 信号 (1/0)；多头-only。返回长度 n-1。
func smaSignals(closes []float64, short, long int) []int {
	n := len(closes)
	if n <= long+1 {
		return nil
	}
	sigs := make([]int, n-1)
	for i := long; i < n-1; i++ {
		ssum, lsum := 0.0, 0.0
		for j := i - short + 1; j <= i; j++ {
			ssum += closes[j]
		}
		for j := i - long + 1; j <= i; j++ {
			lsum += closes[j]
		}
		if ssum/float64(short) > lsum/float64(long) {
			sigs[i] = 1
		}
	}
	return sigs
}

func indexOf(times []time.Time, target time.Time) int {
	for i, t := range times {
		if t.Equal(target) {
			return i
		}
	}
	return -1
}

// AdvRegimeStats 单状态分层指标。
type AdvRegimeStats struct {
	N           int
	Sharpe      float64
	Sortino     float64
	MaxDrawdown float64
	WinRate     float64
}

// computeRegimeStats 按 trade.Regime 分组，计算 Sharpe / Sortino / MaxDD / WinRate。
func computeRegimeStats(trades []AdvTrade) map[string]*AdvRegimeStats {
	groups := map[string][]float64{}
	for _, t := range trades {
		groups[t.Regime] = append(groups[t.Regime], t.PnL)
	}
	out := make(map[string]*AdvRegimeStats, len(groups))
	for r, pnls := range groups {
		stats := &AdvRegimeStats{N: len(pnls)}
		if len(pnls) == 0 {
			out[r] = stats
			continue
		}
		var mean, sumPos float64
		var wins int
		for _, p := range pnls {
			mean += p
			if p > 0 {
				sumPos++
				wins++
			}
		}
		_ = sumPos
		mean /= float64(len(pnls))
		var v, vneg float64
		for _, p := range pnls {
			d := p - mean
			v += d * d
			if p < 0 {
				vneg += p * p
			}
		}
		std := math.Sqrt(v / float64(len(pnls)))
		stdNeg := math.Sqrt(vneg / float64(len(pnls)))
		if std > 0 {
			stats.Sharpe = mean / std * math.Sqrt(252)
		}
		if stdNeg > 0 {
			stats.Sortino = mean / stdNeg * math.Sqrt(252)
		}
		stats.WinRate = float64(wins) / float64(len(pnls))
		// MaxDrawdown 基于复利 equity curve
		stats.MaxDrawdown = computeMaxDrawdown(pnls)
		out[r] = stats
	}
	return out
}

// computeMaxDrawdown 给定逐笔 PnL（小数），返回最大回撤（正数）。
func computeMaxDrawdown(pnls []float64) float64 {
	eq := 1.0
	peak := 1.0
	maxDD := 0.0
	for _, p := range pnls {
		eq *= 1 + p
		if eq > peak {
			peak = eq
		}
		dd := (peak - eq) / peak
		if dd > maxDD {
			maxDD = dd
		}
	}
	return maxDD
}

// equityCurveFromTrades 累计权益曲线（初始 1.0）。
func equityCurveFromTrades(trades []AdvTrade) []float64 {
	out := make([]float64, len(trades)+1)
	out[0] = 1.0
	for i, t := range trades {
		out[i+1] = out[i] * (1 + t.PnL)
	}
	return out
}

// SaveTrades 把交易批量写入 backtest_trades 表。pool=nil 时跳过。
func SaveTrades(ctx context.Context, pool *pgxpool.Pool, jobID string, trades []AdvTrade) error {
	if pool == nil || len(trades) == 0 {
		return nil
	}
	for _, t := range trades {
		_, err := pool.Exec(ctx, `
			INSERT INTO backtest_trades(job_id, seq, opened_at, closed_at, side, entry, exit,
				pnl, pnl_pct, mfe, mae, cost, regime)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
			ON CONFLICT (job_id, seq) DO UPDATE SET
				opened_at=EXCLUDED.opened_at, closed_at=EXCLUDED.closed_at,
				side=EXCLUDED.side, entry=EXCLUDED.entry, exit=EXCLUDED.exit,
				pnl=EXCLUDED.pnl, pnl_pct=EXCLUDED.pnl_pct,
				mfe=EXCLUDED.mfe, mae=EXCLUDED.mae, cost=EXCLUDED.cost, regime=EXCLUDED.regime`,
			jobID, t.Seq, t.OpenedAt, t.ClosedAt, t.Side, t.Entry, t.Exit,
			t.PnL, t.PnLPct, t.MFE, t.MAE, t.Cost, t.Regime)
		if err != nil {
			return err
		}
	}
	return nil
}

// SaveRegimeStats 把状态分层指标批量写入 backtest_metrics_by_regime。
func SaveRegimeStats(ctx context.Context, pool *pgxpool.Pool, jobID string, m map[string]*AdvRegimeStats) error {
	if pool == nil || len(m) == 0 {
		return nil
	}
	for r, s := range m {
		_, err := pool.Exec(ctx, `
			INSERT INTO backtest_metrics_by_regime(job_id, regime, n_trades, sharpe, sortino, max_drawdown, win_rate)
			VALUES ($1,$2,$3,$4,$5,$6,$7)
			ON CONFLICT (job_id, regime) DO UPDATE SET
				n_trades=EXCLUDED.n_trades, sharpe=EXCLUDED.sharpe,
				sortino=EXCLUDED.sortino, max_drawdown=EXCLUDED.max_drawdown,
				win_rate=EXCLUDED.win_rate`,
			jobID, r, s.N, s.Sharpe, s.Sortino, s.MaxDrawdown, s.WinRate)
		if err != nil {
			return err
		}
	}
	return nil
}

// MonteCarlo 用 GARCH 残差 bootstrap 生成多条价格路径。
//
// 算法：
//  1. 从 closes 估计 GARCH(1,1)；得到条件方差序列 condVar。
//  2. 标准化残差 ε_t = r_t / σ_t；构成 i.i.d. 经验分布。
//  3. 用 GARCH 递推 σ²_{t+1} = ω + α r²_t + β σ²_t；
//     未来 r_{t+1} = ε* · σ_{t+1}，ε* 从经验分布有放回抽取。
//  4. 累计为价格 P_{t+1} = P_t · exp(r_{t+1})。
//
// 返回：每路径的终值，以及 p05/p50/p95 分位路径。
func MonteCarlo(closes []float64, paths, horizon int, seed int64) ([][]float64, *quant.GARCHParams, error) {
	if paths <= 0 || paths > 10000 {
		paths = 1000
	}
	if horizon <= 0 {
		horizon = 20
	}
	rets := quant.LogReturns(closes)
	if len(rets) < 80 {
		return nil, nil, fmt.Errorf("monte carlo: need >= 80 returns, got %d", len(rets))
	}
	params, condVar, err := quant.FitGARCH(rets)
	if err != nil {
		return nil, nil, fmt.Errorf("monte carlo GARCH: %w", err)
	}
	// 标准化残差
	resid := make([]float64, len(rets))
	for i, r := range rets {
		s := math.Sqrt(condVar[i])
		if s > 0 {
			resid[i] = r / s
		}
	}
	rng := rand.New(rand.NewSource(seed))
	last := closes[len(closes)-1]
	lastVar := condVar[len(condVar)-1]
	lastRet := rets[len(rets)-1]
	all := make([][]float64, paths)
	for p := 0; p < paths; p++ {
		path := make([]float64, horizon+1)
		path[0] = last
		s2 := lastVar
		r := lastRet
		for h := 1; h <= horizon; h++ {
			s2 = params.Omega + params.Alpha*r*r + params.Beta*s2
			eps := resid[rng.Intn(len(resid))]
			r = eps * math.Sqrt(s2)
			path[h] = path[h-1] * math.Exp(r)
		}
		all[p] = path
	}
	return all, params, nil
}

// QuantilePath 给定 N 条等长路径，逐时点取分位数；q ∈ [0,1]。
func QuantilePath(paths [][]float64, q float64) []float64 {
	if len(paths) == 0 {
		return nil
	}
	h := len(paths[0])
	out := make([]float64, h)
	col := make([]float64, len(paths))
	for j := 0; j < h; j++ {
		for i := range paths {
			col[i] = paths[i][j]
		}
		sort.Float64s(col)
		idx := int(float64(len(col)-1) * q)
		out[j] = col[idx]
	}
	return out
}
