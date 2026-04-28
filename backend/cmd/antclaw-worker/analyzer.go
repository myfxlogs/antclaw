// 派生数据分析: cot_analyses / macro_regime_history / flow_divergence / volume_profiles
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ========== 1. COT分析 (从cot_records计算COT Index/Z-score) ==========

func analyzeCOT(ctx context.Context, dbpool *pgxpool.Pool, logger *slog.Logger) error {
	logger.Info("Analyzing COT Index, Z-score, percentile")

	// 按合约分组查询最近3年数据
	rows, err := dbpool.Query(ctx, `
		SELECT report_date, contract_code, 
		       COALESCE(noncomm_long,0) - COALESCE(noncomm_short,0) AS net
		FROM cot_records
		WHERE report_date >= NOW() - INTERVAL '3 years'
		ORDER BY contract_code, report_date`)
	if err != nil {
		return fmt.Errorf("COT analysis query: %w", err)
	}
	defer rows.Close()

	type cotRow struct {
		Date time.Time
		Code string
		Net  int64
	}
	grouped := map[string][]cotRow{}
	for rows.Next() {
		var r cotRow
		if err := rows.Scan(&r.Date, &r.Code, &r.Net); err != nil {
			continue
		}
		grouped[r.Code] = append(grouped[r.Code], r)
	}

	count := 0
	for code, series := range grouped {
		if len(series) < 4 {
			continue
		}
		// 提取净头寸序列计算min/max/mean/std
		nets := make([]int64, len(series))
		for i, s := range series {
			nets[i] = s.Net
		}

		for i, cur := range series {
			window := nets[:i+1]
			minV, maxV, mean, std := stats(window)
			rng := maxV - minV
			cotIdx := 0.0
			if rng > 0 {
				cotIdx = float64(cur.Net-minV) / float64(rng) * 100
			}
			zscore := 0.0
			if std > 0 {
				zscore = (float64(cur.Net) - mean) / std
			}
			percentile := percentileRank(window, cur.Net)

			direction := "neutral"
			if cotIdx > 75 {
				direction = "overbought"
			} else if cotIdx < 25 {
				direction = "oversold"
			} else if cur.Net > 0 {
				direction = "long"
			} else if cur.Net < 0 {
				direction = "short"
			}

			var wow int64
			if i > 0 {
				wow = cur.Net - series[i-1].Net
			}

			_, err := dbpool.Exec(ctx, `
				INSERT INTO cot_analyses 
				(report_date, contract_code, net_position, cot_index, direction, sentiment_score, wow_change, zscore, percentile)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
				ON CONFLICT (report_date, contract_code) DO UPDATE SET
				  net_position = EXCLUDED.net_position,
				  cot_index = EXCLUDED.cot_index,
				  direction = EXCLUDED.direction,
				  zscore = EXCLUDED.zscore,
				  percentile = EXCLUDED.percentile`,
				cur.Date, code, cur.Net, cotIdx, direction,
				cotIdx/100.0, wow, zscore, percentile)
			if err == nil {
				count++
			}
		}
		logger.Info("COT analyzed", "contract", code, "rows", len(series))
	}
	logger.Info("COT analysis completed", "total", count)
	return nil
}

// ========== 2. 宏观状态分类 (从FRED数据推断Risk-On/Off) ==========

func analyzeMacroRegime(ctx context.Context, dbpool *pgxpool.Pool, logger *slog.Logger) error {
	logger.Info("Analyzing macro regime classification")

	// 查询关键指标最近1年
	rows, err := dbpool.Query(ctx, `
		SELECT time, series_id, value_numeric
		FROM data_snapshots
		WHERE series_id IN ('T10Y2Y', 'T10YIE', 'UNRATE', 'TB3MS', 'GDP')
		  AND time >= NOW() - INTERVAL '1 year'
		  AND value_numeric IS NOT NULL
		ORDER BY time`)
	if err != nil {
		return fmt.Errorf("macro regime query: %w", err)
	}
	defer rows.Close()

	// 按日期聚合
	byDate := map[time.Time]map[string]float64{}
	for rows.Next() {
		var t time.Time
		var sid string
		var v float64
		if err := rows.Scan(&t, &sid, &v); err != nil {
			continue
		}
		day := t.UTC().Truncate(24 * time.Hour)
		if byDate[day] == nil {
			byDate[day] = map[string]float64{}
		}
		byDate[day][sid] = v
	}

	count := 0
	for day, vals := range byDate {
		score := 0.0
		// 收益率曲线 (T10Y2Y): 正值=经济扩张,负值=衰退信号
		if v, ok := vals["T10Y2Y"]; ok {
			score += v * 10 // 权重
		}
		// 通胀预期 (T10YIE): 适度通胀=risk-on
		if v, ok := vals["T10YIE"]; ok {
			if v > 2.0 && v < 3.5 {
				score += 5
			} else if v > 4.0 {
				score -= 5
			}
		}
		// 失业率 (UNRATE): 低=强经济
		if v, ok := vals["UNRATE"]; ok {
			score += (5.0 - v) * 2
		}

		regime := "neutral"
		switch {
		case score > 10:
			regime = "risk_on"
		case score > 5:
			regime = "mild_risk_on"
		case score < -10:
			regime = "risk_off"
		case score < -5:
			regime = "mild_risk_off"
		}

		details, _ := json.Marshal(vals)
		_, err := dbpool.Exec(ctx, `
			INSERT INTO macro_regime_history (time, regime, score, details)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (time) DO UPDATE SET
			  regime = EXCLUDED.regime,
			  score = EXCLUDED.score,
			  details = EXCLUDED.details`,
			day, regime, score, details)
		if err == nil {
			count++
		}
	}
	logger.Info("Macro regime analysis completed", "records", count)
	return nil
}

// ========== 3. 资金流向背离 (相关货币对对比) ==========

func analyzeFlowDivergence(ctx context.Context, dbpool *pgxpool.Pool, logger *slog.Logger) error {
	logger.Info("Analyzing flow divergence")

	pairs := []struct{ A, B string }{
		{"EURUSD", "GBPUSD"}, {"AUDUSD", "NZDUSD"},
		{"USDJPY", "USDCHF"}, {"XAUUSD", "XAGUSD"},
	}

	count := 0
	for _, p := range pairs {
		rows, err := dbpool.Query(ctx, `
			SELECT a.time, a.close, b.close
			FROM price_daily a
			JOIN price_daily b ON a.time = b.time
			WHERE a.symbol = $1 AND b.symbol = $2
			  AND a.time >= NOW() - INTERVAL '90 days'
			ORDER BY a.time`, p.A, p.B)
		if err != nil {
			continue
		}
		var closesA, closesB []float64
		var times []time.Time
		for rows.Next() {
			var t time.Time
			var a, b float64
			if err := rows.Scan(&t, &a, &b); err == nil {
				times = append(times, t)
				closesA = append(closesA, a)
				closesB = append(closesB, b)
			}
		}
		rows.Close()

		if len(closesA) < 30 {
			continue
		}
		// 计算日收益率
		retA := returns(closesA)
		retB := returns(closesB)

		// 滚动20日相关性+基线
		windowSize := 20
		baseline := pearson(retA, retB)
		_, _, _, std := statsF(retA) // 用回报率std作为baseline std

		for i := windowSize; i < len(retA); i++ {
			winA := retA[i-windowSize : i]
			winB := retB[i-windowSize : i]
			corr := pearson(winA, winB)
			zScore := 0.0
			if std > 0 {
				zScore = (corr - baseline) / std
			}
			_, err := dbpool.Exec(ctx, `
				INSERT INTO flow_divergence_history 
				(time, pair_a, pair_b, corr, baseline_mean, baseline_std, z_score, lead_lag)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				ON CONFLICT (time, pair_a, pair_b) DO UPDATE SET
				  corr = EXCLUDED.corr,
				  z_score = EXCLUDED.z_score`,
				times[i], p.A, p.B, corr, baseline, std, zScore, 0)
			if err == nil {
				count++
			}
		}
	}
	logger.Info("Flow divergence analysis completed", "records", count)
	return nil
}

// ========== 4. 成交量分布 (POC/VAH/VAL) ==========

func analyzeVolumeProfile(ctx context.Context, dbpool *pgxpool.Pool, logger *slog.Logger) error {
	logger.Info("Analyzing volume profile (POC/VAH/VAL)")

	symbols := []string{"EURUSD", "GBPUSD", "USDJPY", "SP500", "VIX", "XAUUSD"}
	count := 0
	for _, sym := range symbols {
		rows, err := dbpool.Query(ctx, `
			SELECT time, high, low, close, volume
			FROM price_daily
			WHERE symbol = $1 AND time >= NOW() - INTERVAL '30 days'
			ORDER BY time`, sym)
		if err != nil {
			continue
		}
		type bar struct {
			t             time.Time
			h, l, c, v float64
		}
		var bars []bar
		for rows.Next() {
			var b bar
			if err := rows.Scan(&b.t, &b.h, &b.l, &b.c, &b.v); err == nil {
				bars = append(bars, b)
			}
		}
		rows.Close()
		if len(bars) < 5 {
			continue
		}

		// 构建价格桶 (20个bucket)
		hi := bars[0].h
		lo := bars[0].l
		for _, b := range bars {
			if b.h > hi {
				hi = b.h
			}
			if b.l < lo {
				lo = b.l
			}
		}
		if hi <= lo {
			continue
		}
		buckets := 20
		bucketSize := (hi - lo) / float64(buckets)
		volBuckets := make([]float64, buckets)
		priceBuckets := make([]float64, buckets)
		for i := 0; i < buckets; i++ {
			priceBuckets[i] = lo + bucketSize*(float64(i)+0.5)
		}
		for _, b := range bars {
			// 均匀分布该bar成交量到其price range
			idxLow := int((b.l - lo) / bucketSize)
			idxHigh := int((b.h - lo) / bucketSize)
			if idxLow < 0 {
				idxLow = 0
			}
			if idxHigh >= buckets {
				idxHigh = buckets - 1
			}
			span := idxHigh - idxLow + 1
			if span < 1 {
				span = 1
			}
			per := b.v / float64(span)
			for j := idxLow; j <= idxHigh; j++ {
				volBuckets[j] += per
			}
		}

		// POC = 最大成交量对应价格
		pocIdx := 0
		for i, v := range volBuckets {
			if v > volBuckets[pocIdx] {
				pocIdx = i
			}
		}
		poc := priceBuckets[pocIdx]

		// Value Area: 70%成交量区间
		totalVol := 0.0
		for _, v := range volBuckets {
			totalVol += v
		}
		target := totalVol * 0.7
		accumulated := volBuckets[pocIdx]
		lowIdx, highIdx := pocIdx, pocIdx
		for accumulated < target && (lowIdx > 0 || highIdx < buckets-1) {
			var upV, dnV float64
			if highIdx < buckets-1 {
				upV = volBuckets[highIdx+1]
			}
			if lowIdx > 0 {
				dnV = volBuckets[lowIdx-1]
			}
			if upV >= dnV && highIdx < buckets-1 {
				highIdx++
				accumulated += upV
			} else if lowIdx > 0 {
				lowIdx--
				accumulated += dnV
			} else {
				break
			}
		}
		vah := priceBuckets[highIdx]
		val := priceBuckets[lowIdx]

		profile, _ := json.Marshal(map[string]interface{}{
			"prices": priceBuckets,
			"volumes": volBuckets,
		})
		now := time.Now().UTC().Truncate(24 * time.Hour)
		_, err = dbpool.Exec(ctx, `
			INSERT INTO volume_profiles (time, symbol, period, poc, vah, val, profile)
			VALUES ($1, $2, '30d', $3, $4, $5, $6)
			ON CONFLICT (time, symbol, period) DO UPDATE SET
			  poc = EXCLUDED.poc, vah = EXCLUDED.vah, val = EXCLUDED.val,
			  profile = EXCLUDED.profile`,
			now, sym, poc, vah, val, profile)
		if err == nil {
			count++
			logger.Info("Volume profile", "symbol", sym, "poc", poc, "vah", vah, "val", val)
		}
	}
	logger.Info("Volume profile analysis completed", "records", count)
	return nil
}

// ========== 统计辅助函数 ==========

func stats(xs []int64) (min, max int64, mean, std float64) {
	if len(xs) == 0 {
		return
	}
	min, max = xs[0], xs[0]
	var sum float64
	for _, x := range xs {
		if x < min {
			min = x
		}
		if x > max {
			max = x
		}
		sum += float64(x)
	}
	mean = sum / float64(len(xs))
	var ss float64
	for _, x := range xs {
		d := float64(x) - mean
		ss += d * d
	}
	std = math.Sqrt(ss / float64(len(xs)))
	return
}

func statsF(xs []float64) (min, max, mean, std float64) {
	if len(xs) == 0 {
		return
	}
	min, max = xs[0], xs[0]
	var sum float64
	for _, x := range xs {
		if x < min {
			min = x
		}
		if x > max {
			max = x
		}
		sum += x
	}
	mean = sum / float64(len(xs))
	var ss float64
	for _, x := range xs {
		d := x - mean
		ss += d * d
	}
	std = math.Sqrt(ss / float64(len(xs)))
	return
}

func percentileRank(xs []int64, v int64) float64 {
	sorted := make([]int64, len(xs))
	copy(sorted, xs)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	idx := sort.Search(len(sorted), func(i int) bool { return sorted[i] >= v })
	return float64(idx) / float64(len(sorted)) * 100
}

func returns(closes []float64) []float64 {
	if len(closes) < 2 {
		return nil
	}
	r := make([]float64, len(closes)-1)
	for i := 1; i < len(closes); i++ {
		if closes[i-1] == 0 {
			continue
		}
		r[i-1] = (closes[i] - closes[i-1]) / closes[i-1]
	}
	return r
}

func pearson(a, b []float64) float64 {
	n := len(a)
	if n != len(b) || n == 0 {
		return 0
	}
	var sumA, sumB, sumAB, sumAA, sumBB float64
	for i := 0; i < n; i++ {
		sumA += a[i]
		sumB += b[i]
		sumAB += a[i] * b[i]
		sumAA += a[i] * a[i]
		sumBB += b[i] * b[i]
	}
	num := float64(n)*sumAB - sumA*sumB
	den := math.Sqrt((float64(n)*sumAA - sumA*sumA) * (float64(n)*sumBB - sumB*sumB))
	if den == 0 {
		return 0
	}
	return num / den
}
