// Package main 中 COT 信号校准：对每个 contract_code（视作 signal_type）
// 拟合 Platt-scale (a,b) 与样本胜率，写入 cot_calibration（PK=signal_type，UPSERT）。
//
// 简化算法：
//   - net = comm_long - comm_short
//   - mean,std 取近 24 个月样本
//   - platt_a = 1/std, platt_b = -mean/std （即 score = a*net + b 即为 z 值）
//   - win_rate：极性一致样本占比（占位实现，待真实交易结果上线替换）
package main

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func calibrateCOT(ctx context.Context, db *pgxpool.Pool, logger *slog.Logger) error {
	rows, err := db.Query(ctx, `
		SELECT contract_code, report_date,
		       (COALESCE(comm_long,0)-COALESCE(comm_short,0))::float8 AS commercial_net
		  FROM cot_records
		 WHERE report_date >= NOW() - INTERVAL '24 months'
		 ORDER BY contract_code, report_date`)
	if err != nil {
		return fmt.Errorf("cot-calibration query: %w", err)
	}
	defer rows.Close()

	groups := map[string][]float64{}
	for rows.Next() {
		var code string
		var d time.Time
		var net float64
		if err := rows.Scan(&code, &d, &net); err != nil {
			continue
		}
		groups[code] = append(groups[code], net)
	}

	upsert := 0
	for code, nets := range groups {
		if len(nets) < 12 {
			continue
		}
		mean, std := meanStd(nets)
		if std < 1e-9 {
			continue
		}
		a := 1 / std
		b := -mean / std
		hits := 0
		for _, v := range nets {
			z := a*v + b
			if (v >= mean && z >= 0) || (v < mean && z < 0) {
				hits++
			}
		}
		win := float64(hits) / float64(len(nets))
		_, err := db.Exec(ctx, `
			INSERT INTO cot_calibration(signal_type, platt_a, platt_b, win_rate, sample_size, updated_at)
			VALUES ($1,$2,$3,$4,$5,NOW())
			ON CONFLICT (signal_type) DO UPDATE
			   SET platt_a=EXCLUDED.platt_a, platt_b=EXCLUDED.platt_b,
			       win_rate=EXCLUDED.win_rate, sample_size=EXCLUDED.sample_size,
			       updated_at=NOW()`, code, a, b, win, len(nets))
		if err == nil {
			upsert++
		}
	}
	logger.Info("cot-calibration upserted", "rows", upsert, "groups", len(groups))
	return nil
}

func meanStd(xs []float64) (float64, float64) {
	if len(xs) == 0 {
		return 0, 0
	}
	var sum float64
	for _, v := range xs {
		sum += v
	}
	mean := sum / float64(len(xs))
	var sq float64
	for _, v := range xs {
		sq += (v - mean) * (v - mean)
	}
	return mean, math.Sqrt(sq / float64(len(xs)))
}
