// Package main 中 Walk-forward 回测：对每个 daily symbol 用 MA(20)/MA(50) 交叉
// 简单策略，按 6 个月训练 / 1 个月测试滚动 fold，计算 In-sample 与 OOS Sharpe，
// 写入 walkforward_history。
//
// 训练阶段优化的是"信号方向反转"开关（顺势 vs 反向）；OOS 阶段固定使用训练期更优方向。
package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"math"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	wfTrainDays = 180
	wfTestDays  = 30
	wfMinHist   = wfTrainDays + wfTestDays + 50
)

func runWalkforward(ctx context.Context, db *pgxpool.Pool, logger *slog.Logger) error {
	// 仅取 EURUSD 与 SP500 两类代表性标的，避免一次写入过多冗余 fold。
	symbols := []string{"EURUSD", "SP500"}
	inserted := 0
	for _, sym := range symbols {
		bars, err := loadDailyBars(ctx, db, sym, 600)
		if err != nil || len(bars) < wfMinHist {
			continue
		}
		folds := buildWalkforwardFolds(bars)
		for i, f := range folds {
			ret := maCrossReturns(f.train, 20, 50, true)
			retInv := maCrossReturns(f.train, 20, 50, false)
			invert := sumFloats(retInv) > sumFloats(ret)
			isReturns := ret
			if invert {
				isReturns = retInv
			}
			oosReturns := maCrossReturns(f.test, 20, 50, !invert)

			weights, _ := json.Marshal(map[string]any{
				"strategy": "ma_cross", "fast": 20, "slow": 50, "invert": invert,
			})
			_, err := db.Exec(ctx, `
				INSERT INTO walkforward_history(fold_idx, train_from, train_to, test_from, test_to,
				                                 optimal_weights, in_sample_sharpe, oos_sharpe, created_at)
				SELECT $1,$2,$3,$4,$5,$6::jsonb,$7,$8,NOW()
				 WHERE NOT EXISTS (
				   SELECT 1 FROM walkforward_history
				    WHERE fold_idx=$1 AND train_from=$2 AND test_from=$4)`,
				i, f.train[0].time, f.train[len(f.train)-1].time,
				f.test[0].time, f.test[len(f.test)-1].time,
				string(weights), sharpeAnnual(isReturns), sharpeAnnual(oosReturns))
			if err == nil {
				inserted++
			}
		}
	}
	logger.Info("walkforward inserted", "rows", inserted, "symbols", len(symbols))
	return nil
}

type wfFold struct {
	train, test []ohlc
}

func buildWalkforwardFolds(bars []ohlc) []wfFold {
	if len(bars) < wfMinHist {
		return nil
	}
	var out []wfFold
	for end := wfTrainDays + wfTestDays; end <= len(bars); end += wfTestDays {
		train := bars[end-wfTrainDays-wfTestDays : end-wfTestDays]
		test := bars[end-wfTestDays : end]
		out = append(out, wfFold{train: train, test: test})
	}
	return out
}

// maCrossReturns 计算 MA 快慢线交叉策略的逐日收益序列。
// follow=true 顺势：fast>slow → 多；fast<slow → 空。
// follow=false 反向：fast>slow → 空；fast<slow → 多。
func maCrossReturns(bars []ohlc, fast, slow int, follow bool) []float64 {
	if len(bars) < slow+1 {
		return nil
	}
	maF := movingAvg(bars, fast)
	maS := movingAvg(bars, slow)
	out := make([]float64, 0, len(bars)-slow)
	for i := slow; i < len(bars); i++ {
		if maF[i-1] == 0 || maS[i-1] == 0 {
			continue
		}
		dir := 0.0
		if maF[i-1] > maS[i-1] {
			dir = 1
		} else if maF[i-1] < maS[i-1] {
			dir = -1
		}
		if !follow {
			dir = -dir
		}
		ret := (bars[i].close - bars[i-1].close) / bars[i-1].close
		out = append(out, dir*ret)
	}
	return out
}

func sumFloats(xs []float64) float64 {
	var s float64
	for _, v := range xs {
		s += v
	}
	return s
}

// sharpeAnnual 年化 Sharpe（假设 252 交易日）。
func sharpeAnnual(rets []float64) float64 {
	if len(rets) < 2 {
		return 0
	}
	mean, std := meanStd(rets)
	if std < 1e-12 {
		return 0
	}
	return (mean / std) * math.Sqrt(252)
}

