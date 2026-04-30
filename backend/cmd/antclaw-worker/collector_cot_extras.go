// Package main 中 COT 相关派生采集：
//   - cot-signal-outcomes : 把过去强 COT 信号的实际未来收益落到 cot_signal_outcomes，
//     供 cot-calibration 计算 Platt 校准参数。
//   - data-snapshot       : 把 cot_analyses 的 cot_index 时间序列镜像到 data_snapshots
//     (source='cot', series_id='COT_<CCY>_INDEX')，让"通用数据快照 (其他来源)"
//     语料项不再为空，并便于统一时间序列查询。
//
// 设计原则：仅向上游历史数据派生，不依赖外部网络。
package main

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// cotContractToForex 把 COT 合约代码映射到 forex 标的，并指明方向是否需要翻转。
//
//	"long EUR" → "long EURUSD"          (flip=false)
//	"long JPY" → "short USDJPY"         (flip=true)
//
// USD 计价货币（CAD/CHF/JPY）的 COT 多头意味着对应 forex 报价的下跌。
var cotContractToForex = map[string]struct {
	symbol string
	flip   bool
}{
	"EUR": {"EURUSD", false},
	"GBP": {"GBPUSD", false},
	"NZD": {"NZDUSD", false},
	"JPY": {"USDJPY", true},
	"CAD": {"USDCAD", true},
	"CHF": {"USDCHF", true},
}

// collectCOTSignalOutcomes 扫描 cot_analyses 中至少 30 日前发出的强信号
// （|zscore| ≥ 1.5），用 price_daily 的未来 7/14/28 日收盘价计算收益，
// 并以此判定 win，写入 cot_signal_outcomes。
//
// signal_id 由 BIGSERIAL 自增；通过 (signal_type, contract_code, issued_at)
// 三元组的 NOT EXISTS 实现幂等。
func collectCOTSignalOutcomes(ctx context.Context, db *pgxpool.Pool, logger *slog.Logger) error {
	// 注：cot_analyses 上游 zscore/cot_index 当前可能未填，因此回退按 direction 字段挑选信号；
	// "oversold" 视为 LONG（反转看多），"overbought" 视为 SHORT。
	rows, err := db.Query(ctx, `
		SELECT report_date, contract_code, COALESCE(direction,''), COALESCE(zscore,0)
		  FROM cot_analyses
		 WHERE COALESCE(direction,'') <> ''
		   AND report_date <= NOW() - INTERVAL '30 days'
		 ORDER BY report_date ASC`)
	if err != nil {
		return err
	}
	defer rows.Close()

	type row struct {
		issuedAt  time.Time
		contract  string
		direction string
		zscore    float64
	}
	var batch []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.issuedAt, &r.contract, &r.direction, &r.zscore); err != nil {
			logger.Warn("cot-signal-outcome scan failed", "error", err)
			continue
		}
		batch = append(batch, r)
	}

	inserted := 0
	for _, r := range batch {
		fx, ok := cotContractToForex[strings.ToUpper(r.contract)]
		if !ok {
			continue
		}
		dir := strings.ToUpper(r.direction)
		switch dir {
		case "LONG", "OVERSOLD":
			dir = "LONG" // 反转思路：oversold → 后续上涨
		case "SHORT", "OVERBOUGHT":
			dir = "SHORT"
		default:
			if r.zscore > 0 {
				dir = "LONG"
			} else {
				dir = "SHORT"
			}
		}

		// 起点价：issued_at 当日最早 close
		var p0, p7, p14, p28 float64
		if err := db.QueryRow(ctx, `
			SELECT close FROM price_daily
			 WHERE symbol=$1 AND time >= $2::date
			 ORDER BY time ASC LIMIT 1`, fx.symbol, r.issuedAt).Scan(&p0); err != nil || p0 == 0 {
			continue
		}
		_ = db.QueryRow(ctx, `SELECT close FROM price_daily WHERE symbol=$1 AND time >= $2::date + INTERVAL '7 days' ORDER BY time ASC LIMIT 1`, fx.symbol, r.issuedAt).Scan(&p7)
		_ = db.QueryRow(ctx, `SELECT close FROM price_daily WHERE symbol=$1 AND time >= $2::date + INTERVAL '14 days' ORDER BY time ASC LIMIT 1`, fx.symbol, r.issuedAt).Scan(&p14)
		_ = db.QueryRow(ctx, `SELECT close FROM price_daily WHERE symbol=$1 AND time >= $2::date + INTERVAL '28 days' ORDER BY time ASC LIMIT 1`, fx.symbol, r.issuedAt).Scan(&p28)

		ret := func(end float64) float64 {
			if end == 0 || p0 == 0 {
				return 0
			}
			r := (end - p0) / p0
			if fx.flip {
				r = -r // COT 在 USD 计价货币上的多头 = 该 forex 报价下跌
			}
			return r
		}
		r1w := ret(p7)
		r2w := ret(p14)
		r4w := ret(p28)
		var win bool
		switch {
		case dir == "LONG" && r4w > 0, dir == "SHORT" && r4w < 0:
			win = true
		}
		conf := r.zscore
		if conf < 0 {
			conf = -conf
		}

		tag, err := db.Exec(ctx, `
			INSERT INTO cot_signal_outcomes
			  (signal_type, contract_code, issued_at, raw_confidence, return_1w, return_2w, return_4w, win, evaluated_at)
			SELECT 'cot_extreme'::varchar, $1::varchar, $2::timestamptz, $3::float8, $4::float8, $5::float8, $6::float8, $7::boolean, NOW()
			 WHERE NOT EXISTS (
			   SELECT 1 FROM cot_signal_outcomes
			    WHERE signal_type='cot_extreme' AND contract_code=$1::varchar AND issued_at=$2::timestamptz)`,
			r.contract, r.issuedAt, conf, r1w, r2w, r4w, win)
		if err != nil {
			logger.Warn("cot-signal-outcome insert failed", "contract", r.contract, "error", err)
			continue
		}
		if tag.RowsAffected() > 0 {
			inserted++
		}
	}
	logger.Info("cot-signal-outcomes inserted", "candidates", len(batch), "rows", inserted)
	return nil
}

// snapshotCOTIndexToDataSnapshots 把每周 cot_analyses.cot_index 镜像到 data_snapshots
// (source='cot')，使统一时间序列视图能覆盖 COT 维度。
//
// time = report_date 转 UTC midnight；series_id = 'COT_<CCY>_INDEX'。
func snapshotCOTIndexToDataSnapshots(ctx context.Context, db *pgxpool.Pool, logger *slog.Logger) error {
	tag, err := db.Exec(ctx, `
		INSERT INTO data_snapshots (time, source, series_id, value_numeric, value_text, raw_json, fetched_at)
		SELECT (a.report_date::timestamptz),
		       'cot',
		       'COT_' || a.contract_code || '_INDEX',
		       a.cot_index::float8,
		       NULL,
		       jsonb_build_object('direction', a.direction, 'zscore', a.zscore, 'percentile', a.percentile),
		       NOW()
		  FROM cot_analyses a
		 WHERE a.cot_index IS NOT NULL
		ON CONFLICT (time, source, series_id) DO UPDATE
		   SET value_numeric = EXCLUDED.value_numeric,
		       raw_json      = EXCLUDED.raw_json,
		       fetched_at    = EXCLUDED.fetched_at
		 WHERE data_snapshots.value_numeric IS DISTINCT FROM EXCLUDED.value_numeric`)
	if err != nil {
		return err
	}
	logger.Info("data-snapshot (cot) upserted", "rows", tag.RowsAffected())
	return nil
}
