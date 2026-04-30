// Package main 中 regime overlay 定时任务：对一组核心标的滚动调用
// regime.Service.Compute，把结果落到 regime_overlay_history & regime_transitions。
//
// 这个文件是 trigger.go / main.go 与 regime 服务之间的薄包装；
// regime.Service 内部的 Compute() 已经处理了 overlay_history 写库逻辑，
// 这里附加：transition 检测（与上一条记录比较 unified_label 是否变化）。
package main

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/antclaw/antclaw/internal/service/regime"
)

// regimeTargets 主动触发计算的标的清单。
var regimeTargets = []struct {
	symbol, timeframe, contract string
}{
	{"EURUSD", "D", "EUR"},
	{"GBPUSD", "D", "GBP"},
	{"USDJPY", "D", "JPY"},
	{"XAUUSD", "D", ""},
	{"SP500", "D", ""},
}

func computeRegimeOverlay(ctx context.Context, db *pgxpool.Pool, logger *slog.Logger) error {
	svc := regime.NewService(db)
	ok, fail, transitions := 0, 0, 0
	for _, t := range regimeTargets {
		var prevLabel string
		var prevScore float64
		_ = db.QueryRow(ctx, `
			SELECT unified_label, COALESCE(unified_score,0)
			  FROM regime_overlay_history
			 WHERE symbol=$1 AND timeframe=$2
			 ORDER BY time DESC LIMIT 1`, t.symbol, t.timeframe).Scan(&prevLabel, &prevScore)

		res, err := svc.Compute(ctx, t.symbol, t.timeframe, t.contract)
		if err != nil {
			fail++
			logger.Warn("regime compute failed", "symbol", t.symbol, "tf", t.timeframe, "error", err)
			continue
		}
		ok++

		if prevLabel != "" && prevLabel != res.UnifiedLabel {
			severity := classifyTransitionSeverity(prevLabel, res.UnifiedLabel)
			if _, err := db.Exec(ctx, `
				INSERT INTO regime_transitions(time, symbol, timeframe, from_label, to_label, from_score, to_score, severity)
				VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
				res.ComputedAt, t.symbol, t.timeframe, prevLabel, res.UnifiedLabel, prevScore, res.UnifiedScore, severity); err == nil {
				transitions++
			}
		}
	}
	logger.Info("regime-overlay done", "ok", ok, "fail", fail, "transitions", transitions)
	return nil
}

func classifyTransitionSeverity(from, to string) string {
	rank := map[string]int{
		"STRONG_BEAR": -2, "BEAR": -1, "NEUTRAL": 0, "BULL": 1, "STRONG_BULL": 2,
	}
	d := rank[to] - rank[from]
	switch {
	case d >= 2 || d <= -2:
		return "MAJOR"
	case d == 1 || d == -1:
		return "MINOR"
	default:
		return "INFO"
	}
}
