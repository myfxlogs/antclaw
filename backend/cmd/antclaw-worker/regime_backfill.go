// Package main 中 regime 历史回填：基于 price_daily 的简化分类器，
// 把每天的状态评分写入 regime_overlay_history（带历史时间戳），
// 然后扫描出 label 变化点写入 regime_transitions。
//
// 简化分类器（不替代 regime.Service.Compute 的多模型融合，仅用于补全 label 时间序列）：
//   - 取 MA(20)、MA(50)，计算 (close-MA50)/ATR14 作为强度因子 s。
//   - s ≥ +1.5 → STRONG_BULL；+0.5 → BULL；±0.5 内 → NEUTRAL；
//     -0.5 → BEAR；≤ -1.5 → STRONG_BEAR。
//
// 仅在首次启动且 regime_transitions 为空时执行一次，避免重复写入。
package main

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

func backfillRegimeHistory(ctx context.Context, db *pgxpool.Pool, logger *slog.Logger) error {
	var existing int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM regime_transitions`).Scan(&existing); err == nil && existing > 0 {
		logger.Info("regime backfill skipped (already populated)", "rows", existing)
		return nil
	}

	totalOverlay, totalTrans := 0, 0
	for _, t := range regimeTargets {
		bars, err := loadDailyBars(ctx, db, t.symbol, 600)
		if err != nil || len(bars) < 60 {
			continue
		}
		ma50 := movingAvg(bars, 50)
		atr := simpleATR(bars, 14)

		var prevLabel string
		var prevScore float64
		for i := 50; i < len(bars); i++ {
			if ma50[i] == 0 || atr[i] == 0 {
				continue
			}
			s := (bars[i].close - ma50[i]) / atr[i]
			label := mapStrengthToLabel(s)
			score := s * 50 // 大致映射 ±100 区间
			if score > 100 {
				score = 100
			}
			if score < -100 {
				score = -100
			}
			models := `["heuristic"]`
			_, err := db.Exec(ctx, `
				INSERT INTO regime_overlay_history(time, symbol, timeframe, unified_score, unified_label,
				                                    hmm_state, hmm_confidence, hmm_score,
				                                    garch_regime, garch_score,
				                                    adx_strength, adx_score,
				                                    cot_score, available_models)
				VALUES ($1,$2,$3,$4,$5,'',0,0,'',0,'',0,0,$6::jsonb)
				ON CONFLICT (time, symbol, timeframe) DO NOTHING`,
				bars[i].time, t.symbol, t.timeframe, score, label, models)
			if err == nil {
				totalOverlay++
			}
			// 比上一天 label 变化 → transition
			if i > 50 && prevLabel != "" && prevLabel != label {
				severity := classifyTransitionSeverity(prevLabel, label)
				if _, err := db.Exec(ctx, `
					INSERT INTO regime_transitions(time, symbol, timeframe, from_label, to_label, from_score, to_score, severity)
					VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
					bars[i].time, t.symbol, t.timeframe, prevLabel, label, prevScore, score, severity); err == nil {
					totalTrans++
				}
			}
			prevLabel = label
			prevScore = score
		}
	}
	logger.Info("regime backfill done", "overlay_rows", totalOverlay, "transitions", totalTrans)
	return nil
}

func mapStrengthToLabel(s float64) string {
	switch {
	case s >= 1.5:
		return "STRONG_BULL"
	case s >= 0.5:
		return "BULL"
	case s <= -1.5:
		return "STRONG_BEAR"
	case s <= -0.5:
		return "BEAR"
	default:
		return "NEUTRAL"
	}
}
