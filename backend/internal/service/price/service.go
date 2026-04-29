// Package price 实现 PriceService：从 price_daily / price_intraday 表读取真实 OHLC，
// 不再使用合成 bar 或 randFloat 占位。无数据时返回明确错误，由上层降级。
package price

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	pricev1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Service 价格服务。pool 为 nil 时仅 GetSession 等无数据依赖方法可用，其他方法返回错误。
type Service struct {
	pool *pgxpool.Pool
}

// Bar OHLCV 内部结构。
type Bar struct {
	Timestamp time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    int64
}

// NewService 构造价格服务（兼容旧调用，pool=nil）。
func NewService() *Service { return &Service{} }

// NewServiceWithPool 推荐用法：注入 pgxpool 以读取真实价格。
func NewServiceWithPool(pool *pgxpool.Pool) *Service { return &Service{pool: pool} }

// timeframeToInterval 把 RPC timeframe 字符串映射到 (table, sql interval, fallback bar interval)。
// 仅支持小时与日内两类；周线/月线由日线下游聚合，未来可改成 continuous aggregate 视图。
func timeframeToInterval(timeframe string) (table, intervalCol string, barDur time.Duration, ok bool) {
	switch strings.ToLower(strings.TrimSpace(timeframe)) {
	case "", "1d", "d", "daily":
		return "price_daily", "", 24 * time.Hour, true
	case "1h", "h", "hourly":
		return "price_intraday", "1h", time.Hour, true
	case "4h":
		return "price_intraday", "4h", 4 * time.Hour, true
	case "15m":
		return "price_intraday", "15m", 15 * time.Minute, true
	case "5m":
		return "price_intraday", "5m", 5 * time.Minute, true
	case "1m":
		return "price_intraday", "1m", time.Minute, true
	}
	return "", "", 0, false
}

// loadBars 从 price_daily / price_intraday 拉最近 N 条 OHLC，时间升序。
func (s *Service) loadBars(ctx context.Context, pair, timeframe string, count int) ([]Bar, error) {
	if s.pool == nil {
		return nil, fmt.Errorf("price: postgres pool not configured")
	}
	if count <= 0 || count > 5000 {
		count = 500
	}
	table, intervalCol, _, ok := timeframeToInterval(timeframe)
	if !ok {
		return nil, fmt.Errorf("price: unsupported timeframe %q", timeframe)
	}

	var (
		rows interface{ Close() }
		err  error
	)
	bars := make([]Bar, 0, count)
	if table == "price_daily" {
		r, e := s.pool.Query(ctx, `
			SELECT time, open, high, low, close, COALESCE(volume,0)
			  FROM price_daily
			 WHERE symbol = $1 AND close IS NOT NULL AND close > 0
			 ORDER BY time DESC LIMIT $2`, pair, count)
		if e != nil {
			return nil, e
		}
		rows = r
		for r.Next() {
			var b Bar
			if err = r.Scan(&b.Timestamp, &b.Open, &b.High, &b.Low, &b.Close, &b.Volume); err != nil {
				r.Close()
				return nil, err
			}
			bars = append(bars, b)
		}
	} else {
		r, e := s.pool.Query(ctx, `
			SELECT time, open, high, low, close, COALESCE(volume,0)
			  FROM price_intraday
			 WHERE symbol = $1 AND interval = $2 AND close IS NOT NULL AND close > 0
			 ORDER BY time DESC LIMIT $3`, pair, intervalCol, count)
		if e != nil {
			return nil, e
		}
		rows = r
		for r.Next() {
			var b Bar
			if err = r.Scan(&b.Timestamp, &b.Open, &b.High, &b.Low, &b.Close, &b.Volume); err != nil {
				r.Close()
				return nil, err
			}
			bars = append(bars, b)
		}
	}
	rows.Close()
	// 反转为升序
	for i, j := 0, len(bars)-1; i < j; i, j = i+1, j-1 {
		bars[i], bars[j] = bars[j], bars[i]
	}
	return bars, nil
}

// GetPrice 返回 pair 最近若干根 K 线 + 24h 涨跌（基于实际 K 线计算）。
func (s *Service) GetPrice(ctx context.Context, pair, timeframe string, count int32) (*pricev1.GetPriceResponse, error) {
	if pair == "" {
		return nil, fmt.Errorf("pair required")
	}
	c := int(count)
	if c <= 0 {
		c = 100
	}
	bars, err := s.loadBars(ctx, pair, timeframe, c)
	if err != nil {
		return nil, err
	}
	if len(bars) == 0 {
		return nil, fmt.Errorf("price: no bars for %s/%s", pair, timeframe)
	}
	last := bars[len(bars)-1]
	first := bars[0]
	// 24h 变化：选取最近时点 24 小时之前最近一根 bar 作为基准。
	cutoff := last.Timestamp.Add(-24 * time.Hour)
	prev := first
	for i := len(bars) - 1; i >= 0; i-- {
		if !bars[i].Timestamp.After(cutoff) {
			prev = bars[i]
			break
		}
	}
	change := last.Close - prev.Close
	pct := 0.0
	if prev.Close > 0 {
		pct = change / prev.Close * 100
	}
	pb := make([]*pricev1.PriceBar, 0, len(bars))
	for _, b := range bars {
		pb = append(pb, &pricev1.PriceBar{
			Timestamp: b.Timestamp.UTC().Format(time.RFC3339),
			Open:      fmt.Sprintf("%.5f", b.Open),
			High:      fmt.Sprintf("%.5f", b.High),
			Low:       fmt.Sprintf("%.5f", b.Low),
			Close:     fmt.Sprintf("%.5f", b.Close),
			Volume:    b.Volume,
		})
	}
	return &pricev1.GetPriceResponse{
		Pair:          pair,
		Current:       fmt.Sprintf("%.5f", last.Close),
		Change_24H:    fmt.Sprintf("%.5f", change),
		ChangePct_24H: fmt.Sprintf("%.2f%%", pct),
		Bars:          pb,
	}, nil
}

// GetLevels 计算经典 pivot/支撑阻力（基于最后一根真 K 线）。
func (s *Service) GetLevels(ctx context.Context, pair, timeframe string) (*pricev1.GetLevelsResponse, error) {
	if pair == "" {
		return nil, fmt.Errorf("pair required")
	}
	bars, err := s.loadBars(ctx, pair, timeframe, 1)
	if err != nil {
		return nil, err
	}
	if len(bars) == 0 {
		return nil, fmt.Errorf("price: no bars for %s", pair)
	}
	b := bars[len(bars)-1]
	pivot := (b.High + b.Low + b.Close) / 3
	r1 := 2*pivot - b.Low
	r2 := pivot + (b.High - b.Low)
	s1 := 2*pivot - b.High
	s2 := pivot - (b.High - b.Low)
	return &pricev1.GetLevelsResponse{
		Pair: pair,
		Levels: []*pricev1.PriceLevel{
			{Price: fmt.Sprintf("%.5f", r2), Type: "resistance", Strength: 0.8},
			{Price: fmt.Sprintf("%.5f", r1), Type: "resistance", Strength: 0.6},
			{Price: fmt.Sprintf("%.5f", pivot), Type: "pivot", Strength: 1.0},
			{Price: fmt.Sprintf("%.5f", s1), Type: "support", Strength: 0.6},
			{Price: fmt.Sprintf("%.5f", s2), Type: "support", Strength: 0.8},
		},
	}, nil
}

// GetMarketOverview 返回主流货币对的最近 close（取自 price_daily 最新一根）。
func (s *Service) GetMarketOverview(ctx context.Context, category string) (*pricev1.GetMarketOverviewResponse, error) {
	if s.pool == nil {
		return nil, fmt.Errorf("price: postgres pool not configured")
	}
	pairs := []string{"EURUSD", "GBPUSD", "USDJPY", "USDCHF", "AUDUSD", "USDCAD", "NZDUSD", "XAUUSD"}
	items := make([]*pricev1.MarketOverviewItem, 0, len(pairs))
	for _, p := range pairs {
		var close float64
		err := s.pool.QueryRow(ctx,
			`SELECT close FROM price_daily WHERE symbol=$1 AND close IS NOT NULL ORDER BY time DESC LIMIT 1`, p).Scan(&close)
		if err != nil {
			continue
		}
		items = append(items, &pricev1.MarketOverviewItem{Pair: p, Price: fmt.Sprintf("%.5f", close)})
	}
	return &pricev1.GetMarketOverviewResponse{Items: items}, nil
}

// GetSession 计算外汇 4 大时段开闭与简化波动率指数（基于 EURUSD 1h 真值）。
func (s *Service) GetSession(ctx context.Context, pair string) (*pricev1.GetSessionResponse, error) {
	now := time.Now().UTC()
	// 用 EURUSD 最近 100 根 1h K 线估算各时段标准差（标准化为 % 振幅 * 100）。
	volByHour := s.estimateHourlyVol(ctx, "EURUSD")
	mk := func(name string, start, end int) *pricev1.SessionInfo {
		return &pricev1.SessionInfo{
			Session:         name,
			IsOpen:          isSessionOpen(now, start, end),
			OpensAt:         fmt.Sprintf("%02d:00", start),
			ClosesAt:        fmt.Sprintf("%02d:00", end),
			VolatilityIndex: volByHour[start],
		}
	}
	sessions := []*pricev1.SessionInfo{
		mk("sydney", 22, 7),
		mk("tokyo", 0, 9),
		mk("london", 8, 17),
		mk("new_york", 13, 22),
	}
	current := ""
	for _, s := range sessions {
		if s.IsOpen {
			current = s.Session
			break
		}
	}
	return &pricev1.GetSessionResponse{Pair: pair, Sessions: sessions, CurrentSession: current}, nil
}

// estimateHourlyVol 把 1h K 线按小时桶聚合，计算每小时 (high-low)/close 平均 *100，作为波动指数。
// 没有数据时返回空 map，VolatilityIndex 字段会是 0。
func (s *Service) estimateHourlyVol(ctx context.Context, pair string) map[int]float64 {
	out := map[int]float64{}
	if s.pool == nil {
		return out
	}
	rows, err := s.pool.Query(ctx, `
		SELECT time, high, low, close
		  FROM price_intraday
		 WHERE symbol = $1 AND interval = '1h'
		   AND time > NOW() - INTERVAL '30 days'
		   AND close > 0`, pair)
	if err != nil {
		return out
	}
	defer rows.Close()
	type acc struct{ sum, n float64 }
	buckets := map[int]*acc{}
	for rows.Next() {
		var t time.Time
		var h, l, c float64
		if err := rows.Scan(&t, &h, &l, &c); err != nil {
			continue
		}
		if c <= 0 {
			continue
		}
		hour := t.UTC().Hour()
		a, ok := buckets[hour]
		if !ok {
			a = &acc{}
			buckets[hour] = a
		}
		a.sum += (h - l) / c * 100
		a.n++
	}
	for h, a := range buckets {
		if a.n > 0 {
			out[h] = a.sum / a.n
		}
	}
	return out
}

func isSessionOpen(now time.Time, startHour, endHour int) bool {
	hour := now.Hour()
	if startHour < endHour {
		return hour >= startHour && hour < endHour
	}
	return hour >= startHour || hour < endHour
}

// RunScenario 基于真实 ATR 估算冲击场景。shock_pct 默认取最近 14 根 ATR / Close 的均值（百分比）。
func (s *Service) RunScenario(ctx context.Context, pair string, params map[string]string) (*pricev1.RunScenarioResponse, error) {
	bars, err := s.loadBars(ctx, pair, "1d", 30)
	if err != nil {
		return nil, err
	}
	if len(bars) < 15 {
		return nil, fmt.Errorf("price scenario: insufficient bars for %s", pair)
	}
	last := bars[len(bars)-1]
	atr := computeATR(bars, 14)
	shock := atr / last.Close * 100
	if v, err := strconv.ParseFloat(params["shock_pct"], 64); err == nil && v > 0 {
		shock = v
	}
	shockAmount := last.Close * (shock / 100)
	results := []*pricev1.ScenarioResult{
		{ScenarioName: "bullish_breakout", Outcome: fmt.Sprintf("Price reaches %.5f", last.Close+shockAmount*2), Probability: 0.20},
		{ScenarioName: "moderate_rally", Outcome: fmt.Sprintf("Price reaches %.5f", last.Close+shockAmount), Probability: 0.30},
		{ScenarioName: "sideways", Outcome: fmt.Sprintf("Price range %.5f - %.5f", last.Close-shockAmount*0.5, last.Close+shockAmount*0.5), Probability: 0.30},
		{ScenarioName: "moderate_decline", Outcome: fmt.Sprintf("Price reaches %.5f", last.Close-shockAmount), Probability: 0.15},
		{ScenarioName: "bearish_breakdown", Outcome: fmt.Sprintf("Price reaches %.5f", last.Close-shockAmount*2), Probability: 0.05},
	}
	return &pricev1.RunScenarioResponse{Results: results}, nil
}

// computeATR Wilder 平滑 ATR。
func computeATR(bars []Bar, period int) float64 {
	if len(bars) < 2 {
		return 0
	}
	if period <= 0 {
		period = 14
	}
	trs := make([]float64, 0, len(bars)-1)
	for i := 1; i < len(bars); i++ {
		c, p := bars[i], bars[i-1]
		tr := math.Max(c.High-c.Low, math.Max(math.Abs(c.High-p.Close), math.Abs(c.Low-p.Close)))
		trs = append(trs, tr)
	}
	if len(trs) < period {
		// 简单平均兜底
		s := 0.0
		for _, t := range trs {
			s += t
		}
		return s / float64(len(trs))
	}
	// Wilder 平滑
	atr := 0.0
	for i := 0; i < period; i++ {
		atr += trs[i]
	}
	atr /= float64(period)
	for i := period; i < len(trs); i++ {
		atr = (atr*float64(period-1) + trs[i]) / float64(period)
	}
	return atr
}

// GetRegime 判定市场状态。
//
// engine：
//   - "" / "adx"：经典 ADX 判定（trending/volatile/ranging/calm）
//   - "hmm"：用 quant.GaussianHMM 拟合对数收益，输出状态序号 + Viterbi 后验置信度
//
// engine="hmm" 时，若样本不足 / 收敛失败，自动回退到 ADX 实现并在响应 EngineUsed
// 字段中标注 "adx_fallback"。
func (s *Service) GetRegime(ctx context.Context, pair, timeframe, engine string, nStates int32) (*pricev1.GetRegimeResponse, error) {
	if engine == "hmm" {
		resp, err := s.getRegimeHMM(ctx, pair, timeframe, int(nStates))
		if err == nil {
			return resp, nil
		}
		// 失败回退到 ADX；err 仅作日志参考。
	}
	resp, err := s.getRegimeADX(ctx, pair, timeframe)
	if err != nil {
		return nil, err
	}
	if engine == "hmm" {
		resp.EngineUsed = "adx_fallback"
	} else {
		resp.EngineUsed = "adx"
	}
	return resp, nil
}

func (s *Service) getRegimeADX(ctx context.Context, pair, timeframe string) (*pricev1.GetRegimeResponse, error) {
	bars, err := s.loadBars(ctx, pair, timeframe, 60)
	if err != nil {
		return nil, err
	}
	if len(bars) < 15 {
		return &pricev1.GetRegimeResponse{
			Regime: &pricev1.MarketRegime{Regime: "unknown", Confidence: 0, Since: time.Now().Format(time.RFC3339)},
		}, nil
	}
	plusDM, minusDM, trSum := 0.0, 0.0, 0.0
	for i := 1; i < len(bars) && i < 15; i++ {
		c, p := bars[i], bars[i-1]
		tr := math.Max(c.High-c.Low, math.Max(math.Abs(c.High-p.Close), math.Abs(c.Low-p.Close)))
		trSum += tr
		up := c.High - p.High
		down := p.Low - c.Low
		if up > down && up > 0 {
			plusDM += up
		}
		if down > up && down > 0 {
			minusDM += down
		}
	}
	var diPlus, diMinus, adx float64
	if trSum > 0 {
		diPlus = plusDM / trSum * 100
		diMinus = minusDM / trSum * 100
		if diPlus+diMinus > 0 {
			adx = math.Abs(diPlus-diMinus) / (diPlus + diMinus) * 100
		}
	}
	regime := "ranging"
	conf := 0.5
	switch {
	case adx > 25 && diPlus > diMinus:
		regime, conf = "trending", adx/100
	case adx > 25:
		regime, conf = "volatile", adx/100
	case adx < 15:
		regime, conf = "calm", (25-adx)/25
	}
	return &pricev1.GetRegimeResponse{
		Regime:        &pricev1.MarketRegime{Regime: regime, Confidence: conf, Since: time.Now().Add(-24 * time.Hour).Format(time.RFC3339)},
		RecentRegimes: []string{"ranging", regime},
	}, nil
}

// GetSeasonal 用真历史日线按月聚合：avg_return = 月内日收益均值；win_rate = 正收益占比。
func (s *Service) GetSeasonal(ctx context.Context, pair string, years int32) (*pricev1.GetSeasonalResponse, error) {
	if s.pool == nil {
		return nil, fmt.Errorf("price seasonal: postgres pool not configured")
	}
	if years <= 0 {
		years = 5
	}
	rows, err := s.pool.Query(ctx, `
		SELECT time, close
		  FROM price_daily
		 WHERE symbol = $1 AND close > 0
		   AND time > NOW() - ($2 || ' years')::INTERVAL
		 ORDER BY time ASC`, pair, fmt.Sprintf("%d", years))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	type rec struct{ retSum, winN, total float64 }
	by := map[int]*rec{}
	var prev float64
	first := true
	var prevMonth time.Month
	for rows.Next() {
		var t time.Time
		var c float64
		if err := rows.Scan(&t, &c); err != nil {
			continue
		}
		if first {
			prev, prevMonth, first = c, t.Month(), false
			continue
		}
		ret := (c - prev) / prev
		mIdx := int(t.Month()) - 1
		r, ok := by[mIdx]
		if !ok {
			r = &rec{}
			by[mIdx] = r
		}
		r.retSum += ret
		r.total++
		if ret > 0 {
			r.winN++
		}
		prev = c
		prevMonth = t.Month()
	}
	_ = prevMonth
	months := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	out := make([]*pricev1.SeasonalDataPoint, 0, 12)
	for i, name := range months {
		r, ok := by[i]
		if !ok || r.total == 0 {
			out = append(out, &pricev1.SeasonalDataPoint{Month: name})
			continue
		}
		out = append(out, &pricev1.SeasonalDataPoint{
			Month:     name,
			AvgReturn: r.retSum / r.total * 100,
			WinRate:   r.winN / r.total,
		})
	}
	return &pricev1.GetSeasonalResponse{Pair: pair, Data: out}, nil
}
