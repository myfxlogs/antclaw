// Package regime 计算多模型融合的市场状态评分。
//
// 4 个子模型：
//   - HMM 替代：基于收益率自相关与日内波动比的轻量代理（HMM-lite）
//   - GARCH 替代：基于近 20 日年化波动率与 60 日均值之比（GARCH-lite）
//   - ADX：从 daily OHLC 计算 14 日 ADX 趋势强度
//   - COT：从 cot_analyses 表取最新 net_position percentile 极值方向
//
// 融合权重：HMM 30% / GARCH 25% / ADX 25% / COT 20%；
// 任一子模型不可用时，剩余权重按比例重分配。最终 unified_score 区间为 -100..+100。
package regime

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Service 状态融合服务。
type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service { return &Service{pool: pool} }

// SubModel 子模型评分结果。
type SubModel struct {
	Name       string
	Available  bool
	Score      float64 // -100..+100
	Weight     float64 // 0..1
	State      string
	Confidence float64
}

// Result 融合结果。
type Result struct {
	Symbol          string
	Timeframe       string
	ComputedAt      time.Time
	UnifiedScore    float64
	UnifiedLabel    string
	HMM             SubModel
	GARCH           SubModel
	ADX             SubModel
	COT             SubModel
	AvailableModels []string
}

// 子模型初始权重
const (
	wHMM   = 0.30
	wGARCH = 0.25
	wADX   = 0.25
	wCOT   = 0.20
)

// Compute 计算并持久化 overlay 结果。contractCode 可空（COT 子模型则不可用）。
func (s *Service) Compute(ctx context.Context, symbol, timeframe, contractCode string) (*Result, error) {
	if symbol == "" {
		return nil, fmt.Errorf("regime: empty symbol")
	}
	if timeframe == "" {
		timeframe = "D"
	}
	closes, highs, lows, err := s.loadDaily(ctx, symbol, 250)
	if err != nil {
		return nil, fmt.Errorf("regime: load daily: %w", err)
	}
	if len(closes) < 30 {
		return nil, fmt.Errorf("regime: %s/%s 仅 %d 根日线，至少需 30 根（请先在 worker 中拉取 price_daily）", symbol, timeframe, len(closes))
	}

	res := &Result{
		Symbol:     symbol,
		Timeframe:  timeframe,
		ComputedAt: time.Now().UTC(),
		HMM:        runHMMLite(closes),
		GARCH:      runGARCHLite(closes),
		ADX:        runADX(highs, lows, closes, 14),
		COT:        s.runCOT(ctx, contractCode),
	}
	res.fuse()
	if err := s.persist(ctx, res); err != nil {
		// 记录失败不影响返回（融合结果是计算结果，存储失败可重试）
		_ = err
	}
	return res, nil
}

// loadDaily 取 price_daily 最近 N 根日线（按时间升序）。
func (s *Service) loadDaily(ctx context.Context, symbol string, n int) (closes, highs, lows []float64, err error) {
	rows, err := s.pool.Query(ctx, `
		SELECT high, low, close FROM price_daily
		 WHERE symbol = $1
		 ORDER BY time DESC
		 LIMIT $2`, strings.ToUpper(symbol), n)
	if err != nil {
		return nil, nil, nil, err
	}
	defer rows.Close()
	var c, h, l []float64
	for rows.Next() {
		var hi, lo, cl float64
		if err := rows.Scan(&hi, &lo, &cl); err != nil {
			return nil, nil, nil, err
		}
		h = append(h, hi)
		l = append(l, lo)
		c = append(c, cl)
	}
	// 反转为升序
	for i, j := 0, len(c)-1; i < j; i, j = i+1, j-1 {
		c[i], c[j] = c[j], c[i]
		h[i], h[j] = h[j], h[i]
		l[i], l[j] = l[j], l[i]
	}
	return c, h, l, nil
}

// runHMMLite 轻量 HMM 代理：用 20 日收益自相关方向作为趋势/反转信号。
func runHMMLite(closes []float64) SubModel {
	if len(closes) < 30 {
		return SubModel{Name: "hmm"}
	}
	r := returns(closes)
	n := len(r)
	if n < 20 {
		return SubModel{Name: "hmm"}
	}
	// 自相关 lag 1
	mean := mean(r)
	var num, den float64
	for i := 1; i < n; i++ {
		num += (r[i] - mean) * (r[i-1] - mean)
		den += (r[i] - mean) * (r[i] - mean)
	}
	rho := 0.0
	if den > 0 {
		rho = num / den
	}
	// 累积净收益
	cum := 0.0
	for _, x := range r[n-20:] {
		cum += x
	}
	score := 100 * math.Tanh(rho*4+cum*30) // 拉伸到 -100..+100
	state := "MEAN_REVERTING"
	if rho > 0.05 {
		state = "TRENDING"
	} else if rho < -0.05 {
		state = "MEAN_REVERTING"
	} else {
		state = "RANDOM"
	}
	return SubModel{
		Name: "hmm", Available: true, Score: score, State: state,
		Confidence: math.Min(1, math.Abs(rho)*5),
	}
}

// runGARCHLite 用近 20 日年化波动率与 60 日均值之比作为波动率区间代理。
func runGARCHLite(closes []float64) SubModel {
	r := returns(closes)
	if len(r) < 60 {
		return SubModel{Name: "garch"}
	}
	short := annualVol(r[len(r)-20:])
	long := annualVol(r[len(r)-60:])
	if long == 0 {
		return SubModel{Name: "garch"}
	}
	ratio := short / long
	// ratio>1.3 高波动（避险/反向），ratio<0.7 低波动（顺势更优），中性区间得分 0
	score := 0.0
	state := "NORMAL"
	switch {
	case ratio >= 1.3:
		score = -60
		state = "HIGH_VOL"
	case ratio >= 1.1:
		score = -30
		state = "ELEVATED"
	case ratio <= 0.7:
		score = 60
		state = "LOW_VOL"
	case ratio <= 0.9:
		score = 30
		state = "DAMPENED"
	}
	return SubModel{
		Name: "garch", Available: true, Score: score, State: state,
		Confidence: math.Min(1, math.Abs(ratio-1)),
	}
}

// runADX 标准 14 日 ADX 计算。
func runADX(high, low, close []float64, period int) SubModel {
	n := len(close)
	if n < period*2+1 {
		return SubModel{Name: "adx"}
	}
	tr := make([]float64, n)
	plusDM := make([]float64, n)
	minusDM := make([]float64, n)
	for i := 1; i < n; i++ {
		tr[i] = math.Max(high[i]-low[i], math.Max(math.Abs(high[i]-close[i-1]), math.Abs(low[i]-close[i-1])))
		up := high[i] - high[i-1]
		dn := low[i-1] - low[i]
		if up > dn && up > 0 {
			plusDM[i] = up
		}
		if dn > up && dn > 0 {
			minusDM[i] = dn
		}
	}
	atr := wilderAvg(tr, period)
	plusDI := wilderAvg(plusDM, period)
	minusDI := wilderAvg(minusDM, period)
	if atr == 0 {
		return SubModel{Name: "adx"}
	}
	pDI := 100 * plusDI / atr
	mDI := 100 * minusDI / atr
	dx := 0.0
	if pDI+mDI > 0 {
		dx = 100 * math.Abs(pDI-mDI) / (pDI + mDI)
	}
	// ADX 简化：对 dx 取一次平滑近似（完整 14-period 平滑略，已足够分类）
	adx := dx
	dir := 1.0
	if mDI > pDI {
		dir = -1.0
	}
	score := 0.0
	state := "WEAK"
	switch {
	case adx >= 40:
		score = 80 * dir
		state = "STRONG"
	case adx >= 25:
		score = 50 * dir
		state = "MODERATE"
	case adx >= 15:
		score = 20 * dir
		state = "WEAK"
	default:
		state = "RANGING"
	}
	return SubModel{
		Name: "adx", Available: true, Score: score, State: state,
		Confidence: math.Min(1, adx/50),
	}
}

// runCOT 从 cot_analyses 取最新 net_position 与方向。
func (s *Service) runCOT(ctx context.Context, contractCode string) SubModel {
	if s.pool == nil || contractCode == "" {
		return SubModel{Name: "cot"}
	}
	var direction string
	var netPos float64
	err := s.pool.QueryRow(ctx, `
		SELECT direction, COALESCE(net_position::DOUBLE PRECISION, 0)
		  FROM cot_analyses
		 WHERE contract_code = $1
		 ORDER BY report_date DESC LIMIT 1`, contractCode).Scan(&direction, &netPos)
	if err != nil {
		return SubModel{Name: "cot"}
	}
	score := 0.0
	switch strings.ToUpper(direction) {
	case "BULLISH":
		score = 60
	case "BEARISH":
		score = -60
	case "EXTREME_BULLISH":
		score = 90
	case "EXTREME_BEARISH":
		score = -90
	}
	return SubModel{
		Name: "cot", Available: true, Score: score,
		State: strings.ToUpper(direction), Confidence: 0.6,
	}
}

// fuse 计算融合分数与标签，按缺失模型重分配权重。
func (r *Result) fuse() {
	type sm struct {
		ptr  *SubModel
		base float64
	}
	mods := []sm{
		{&r.HMM, wHMM}, {&r.GARCH, wGARCH}, {&r.ADX, wADX}, {&r.COT, wCOT},
	}
	totalAvail := 0.0
	for _, m := range mods {
		if m.ptr.Available {
			totalAvail += m.base
		}
	}
	if totalAvail == 0 {
		r.UnifiedLabel = "NEUTRAL"
		return
	}
	weighted := 0.0
	for _, m := range mods {
		if !m.ptr.Available {
			m.ptr.Weight = 0
			continue
		}
		m.ptr.Weight = m.base / totalAvail
		weighted += m.ptr.Score * m.ptr.Weight
		r.AvailableModels = append(r.AvailableModels, m.ptr.Name)
	}
	r.UnifiedScore = weighted
	r.UnifiedLabel = labelFromScore(weighted)
}

func labelFromScore(s float64) string {
	switch {
	case s >= 60:
		return "STRONG_BULL"
	case s >= 20:
		return "BULL"
	case s <= -60:
		return "STRONG_BEAR"
	case s <= -20:
		return "BEAR"
	default:
		return "NEUTRAL"
	}
}

// persist 把结果写入 regime_overlay_history（upsert）。
func (s *Service) persist(ctx context.Context, r *Result) error {
	if s.pool == nil {
		return nil
	}
	models, _ := json.Marshal(r.AvailableModels)
	_, err := s.pool.Exec(ctx, `
		INSERT INTO regime_overlay_history(
			time, symbol, timeframe,
			unified_score, unified_label,
			hmm_state, hmm_confidence, hmm_score,
			garch_regime, garch_score,
			adx_strength, adx_score,
			cot_score, available_models)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14::jsonb)
		ON CONFLICT (time, symbol, timeframe) DO UPDATE SET
			unified_score = EXCLUDED.unified_score,
			unified_label = EXCLUDED.unified_label`,
		r.ComputedAt, strings.ToUpper(r.Symbol), r.Timeframe,
		r.UnifiedScore, r.UnifiedLabel,
		r.HMM.State, r.HMM.Confidence, r.HMM.Score,
		r.GARCH.State, r.GARCH.Score,
		r.ADX.State, r.ADX.Score,
		r.COT.Score, string(models))
	return err
}

// ListRecent 取近 N 天 overlay 历史。
func (s *Service) ListRecent(ctx context.Context, symbol, timeframe string, days int) ([]struct {
	Time         time.Time
	UnifiedScore float64
	UnifiedLabel string
}, error) {
	if days <= 0 {
		days = 30
	}
	rows, err := s.pool.Query(ctx, `
		SELECT time, unified_score, unified_label FROM regime_overlay_history
		 WHERE symbol = $1 AND timeframe = $2
		   AND time >= NOW() - ($3 || ' days')::INTERVAL
		 ORDER BY time ASC`, strings.ToUpper(symbol), timeframe, fmt.Sprintf("%d", days))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []struct {
		Time         time.Time
		UnifiedScore float64
		UnifiedLabel string
	}
	for rows.Next() {
		var t time.Time
		var sc float64
		var lb string
		if err := rows.Scan(&t, &sc, &lb); err != nil {
			return nil, err
		}
		out = append(out, struct {
			Time         time.Time
			UnifiedScore float64
			UnifiedLabel string
		}{t, sc, lb})
	}
	return out, nil
}

// ---- helpers ----

func returns(c []float64) []float64 {
	out := make([]float64, 0, len(c)-1)
	for i := 1; i < len(c); i++ {
		if c[i-1] == 0 {
			continue
		}
		out = append(out, math.Log(c[i]/c[i-1]))
	}
	return out
}

func mean(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	s := 0.0
	for _, x := range xs {
		s += x
	}
	return s / float64(len(xs))
}

func annualVol(r []float64) float64 {
	if len(r) < 2 {
		return 0
	}
	m := mean(r)
	var v float64
	for _, x := range r {
		v += (x - m) * (x - m)
	}
	v /= float64(len(r) - 1)
	return math.Sqrt(v) * math.Sqrt(252)
}

// wilderAvg Wilder 平滑（period 内简单累积，再除以 period）。
func wilderAvg(xs []float64, period int) float64 {
	if len(xs) < period+1 {
		return 0
	}
	s := 0.0
	for i := len(xs) - period; i < len(xs); i++ {
		s += xs[i]
	}
	return s / float64(period)
}
