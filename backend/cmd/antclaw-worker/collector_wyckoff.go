// Package main 中 Wyckoff 形态检测：扫描 price_daily 最近 250 根日线，识别
// Spring / Upthrust / Sign of Strength (SOS) / Sign of Weakness (SOW) 四类事件。
//
// 检测规则（教科书简化版）：
//   - Spring   : 当前低点 < 过去 N 日最低，且收盘 > 过去 N 日最低（假突破回收，看多）
//   - Upthrust : 当前高点 > 过去 N 日最高，且收盘 < 过去 N 日最高（假突破回落，看空）
//   - SOS      : 收盘穿上 N 日均线，且 close-low > 1.5*ATR（爆发上行）
//   - SOW      : 收盘跌破 N 日均线，且 high-close > 1.5*ATR（爆发下行）
//
// 仅保留最近 60 个交易日内未入库事件。
package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const wyckoffLookback = 10

func detectWyckoff(ctx context.Context, db *pgxpool.Pool, logger *slog.Logger) error {
	symbols, err := listDailySymbols(ctx, db)
	if err != nil {
		return err
	}
	inserted := 0
	for _, sym := range symbols {
		bars, err := loadDailyBars(ctx, db, sym, 250)
		if err != nil || len(bars) < wyckoffLookback+5 {
			continue
		}
		atr := simpleATR(bars, 14)
		ma := movingAvg(bars, wyckoffLookback)
		for i := wyckoffLookback; i < len(bars); i++ {
			b := bars[i]
			window := bars[i-wyckoffLookback : i]
			lo, hi := windowMinMax(window)
			ev := classifyWyckoff(b, ma[i], atr[i], lo, hi)
			if ev == "" {
				continue
			}
			conf := math.Min(1.0, math.Abs(b.close-ma[i])/(2*atr[i]+1e-9))
			raw, _ := json.Marshal(map[string]any{"ma": ma[i], "atr": atr[i], "lo": lo, "hi": hi})
			tag, err := db.Exec(ctx, `
				INSERT INTO wyckoff_events(symbol, timeframe, event_name, bar_time, price, volume, confidence, raw)
				SELECT $1::varchar,'D',$2::varchar,$3::timestamptz,$4::float8,$5::float8,$6::float8,$7::jsonb
				 WHERE NOT EXISTS (
				   SELECT 1 FROM wyckoff_events
				    WHERE symbol=$1::varchar AND timeframe='D' AND event_name=$2::varchar AND bar_time=$3::timestamptz)`,
				sym, ev, b.time, b.close, b.volume, conf, string(raw))
			if err != nil {
				logger.Warn("wyckoff insert failed", "symbol", sym, "event", ev, "error", err)
				continue
			}
			if tag.RowsAffected() > 0 {
				inserted++
			}
		}
	}
	logger.Info("wyckoff-events inserted", "rows", inserted, "symbols", len(symbols))
	return nil
}

// ---------- helpers ----------

type ohlc struct {
	time                   time.Time
	open, high, low, close float64
	volume                 float64
}

func listDailySymbols(ctx context.Context, db *pgxpool.Pool) ([]string, error) {
	rows, err := db.Query(ctx, `SELECT DISTINCT symbol FROM price_daily WHERE time >= NOW() - INTERVAL '12 months'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err == nil {
			out = append(out, s)
		}
	}
	return out, nil
}

func loadDailyBars(ctx context.Context, db *pgxpool.Pool, symbol string, n int) ([]ohlc, error) {
	// 同一交易日可能存在多条数据（如多源采集），按 date 聚合后再取最近 N 天，避免重复扰乱回看窗口。
	rows, err := db.Query(ctx, `
		WITH dedup AS (
		  SELECT (time::date)::timestamptz AS d,
		         (array_agg(open  ORDER BY time))[1]                             AS open,
		         max(high)                                                       AS high,
		         min(low)                                                        AS low,
		         (array_agg(close ORDER BY time DESC))[1]                        AS close,
		         sum(COALESCE(volume,0))                                         AS volume
		    FROM price_daily WHERE symbol=$1
		   GROUP BY time::date
		)
		SELECT d, open, high, low, close, volume FROM dedup
		 ORDER BY d DESC LIMIT $2`, symbol, n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var rev []ohlc
	for rows.Next() {
		var b ohlc
		if err := rows.Scan(&b.time, &b.open, &b.high, &b.low, &b.close, &b.volume); err == nil {
			rev = append(rev, b)
		}
	}
	// 翻转为升序
	out := make([]ohlc, len(rev))
	for i, b := range rev {
		out[len(rev)-1-i] = b
	}
	return out, nil
}

func simpleATR(bars []ohlc, period int) []float64 {
	out := make([]float64, len(bars))
	if len(bars) == 0 {
		return out
	}
	for i := 1; i < len(bars); i++ {
		tr := math.Max(bars[i].high-bars[i].low, math.Max(
			math.Abs(bars[i].high-bars[i-1].close),
			math.Abs(bars[i].low-bars[i-1].close)))
		if i < period {
			out[i] = (out[i-1]*float64(i-1) + tr) / float64(i)
		} else {
			out[i] = (out[i-1]*float64(period-1) + tr) / float64(period)
		}
	}
	return out
}

func movingAvg(bars []ohlc, period int) []float64 {
	out := make([]float64, len(bars))
	var sum float64
	for i, b := range bars {
		sum += b.close
		if i >= period {
			sum -= bars[i-period].close
			out[i] = sum / float64(period)
		} else if i == period-1 {
			out[i] = sum / float64(period)
		}
	}
	return out
}

func windowMinMax(bars []ohlc) (lo, hi float64) {
	lo, hi = math.Inf(1), math.Inf(-1)
	for _, b := range bars {
		if b.low < lo {
			lo = b.low
		}
		if b.high > hi {
			hi = b.high
		}
	}
	return
}

func classifyWyckoff(b ohlc, ma, atr, prevLo, prevHi float64) string {
	if atr <= 0 || ma <= 0 {
		return ""
	}
	// SPRING: 当前低点穿下 prevLo 0.05*atr 以内即视为试探，收盘回到 prevLo 之上。
	if b.low < prevLo+0.05*atr && b.close > prevLo {
		return "SPRING"
	}
	if b.high > prevHi-0.05*atr && b.close < prevHi {
		return "UPTHRUST"
	}
	body := b.close - b.open
	if b.close > ma && body > 0.3*atr {
		return "SOS"
	}
	if b.close < ma && -body > 0.3*atr {
		return "SOW"
	}
	return ""
}
