// Package main 中微观结构快照（基于 price_intraday 派生的代理指标）：
//
// 缺乏真实订单簿数据时，以下列代理逐 5 分钟生成一行：
//   - spread_bps   : (high-low)/close * 10000  // 短窗高低价幅作为 bps 价差代理
//   - obi_top10    : (close-open)/(high-low+ε)  // 价差中收盘相对位置 ∈ [-1,1]
//   - bid_depth    : volume * (1 - position)    // close 越接近 high → 卖压释放小
//   - ask_depth    : volume * position
//   - stress_score : |obi_top10| + spread_bps/100  // 综合应激度
//
// 仅写入最近 24h 内尚未入库的样本，避免每次重算全表。
package main

import (
	"context"
	"fmt"
	"log/slog"
	"math"

	"github.com/jackc/pgx/v5/pgxpool"
)

func collectMicroSnapshots(ctx context.Context, db *pgxpool.Pool, logger *slog.Logger) error {
	rows, err := db.Query(ctx, `
		SELECT time, symbol, open, high, low, close, COALESCE(volume,0)
		  FROM price_intraday
		 WHERE interval='5m' AND time >= NOW() - INTERVAL '24 hours'
		   AND NOT EXISTS (
		     SELECT 1 FROM micro_snapshots m
		      WHERE m.time=price_intraday.time AND m.symbol=price_intraday.symbol)`)
	if err != nil {
		return fmt.Errorf("micro-snapshot scan: %w", err)
	}
	defer rows.Close()

	type bar struct {
		time   any
		symbol string
		o, h, l, c, v float64
	}
	var batch []bar
	for rows.Next() {
		var b bar
		if err := rows.Scan(&b.time, &b.symbol, &b.o, &b.h, &b.l, &b.c, &b.v); err == nil {
			batch = append(batch, b)
		}
	}

	inserted := 0
	for _, b := range batch {
		if b.c <= 0 || b.h <= b.l {
			continue
		}
		spreadBps := (b.h - b.l) / b.c * 10000
		rng := b.h - b.l
		pos := (b.c - b.l) / rng // ∈ [0,1]
		obi := 2*pos - 1         // ∈ [-1,1]
		bidDepth := b.v * (1 - pos)
		askDepth := b.v * pos
		stress := math.Abs(obi) + spreadBps/100
		_, err := db.Exec(ctx, `
			INSERT INTO micro_snapshots(time, symbol, obi_top10, spread_bps, bid_depth, ask_depth, stress_score)
			VALUES ($1,$2,$3,$4,$5,$6,$7)
			ON CONFLICT (time, symbol) DO NOTHING`,
			b.time, b.symbol, obi, spreadBps, bidDepth, askDepth, stress)
		if err == nil {
			inserted++
		}
	}
	logger.Info("micro-snapshot inserted", "rows", inserted, "candidates", len(batch))
	return nil
}
