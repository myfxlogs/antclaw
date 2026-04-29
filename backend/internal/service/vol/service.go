// Package vol 提供波动率指数（VIX/DVOL/MOVE）与隐含波动率派生指标。
//
// 数据源：
//   - VIX：CBOE 公开历史 JSON（cdn.cboe.com）
//   - DVOL：Deribit /public/get_volatility_index_data（BTC/ETH 等加密资产）
//   - GEX/IVol/Skew/SkewVixAlert：基于 Deribit BookSummaries（BTC/ETH 期权）
//   - MOVE：暂无免费实时端点，返回 unavailable 错误，由上层降级
//
// 所有方法均无随机数；空数据时返回错误，由调用方决定是否优雅降级。
package vol

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	volv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/internal/infra/apiclient"
	"github.com/antclaw/antclaw/internal/infra/apiclient/cboe"
	"github.com/antclaw/antclaw/internal/infra/apiclient/deribit"
	"github.com/antclaw/antclaw/internal/infra/apiclient/firecrawl"
)

// ErrUnavailable 数据源暂时不可用（如 MOVE 无免费 API）。
var ErrUnavailable = errors.New("vol: upstream unavailable")

// Service 真数据波动率服务。
type Service struct {
	cboe    *cboe.Client
	deribit *deribit.Client
	fire    *firecrawl.Client // 可选；为 nil 时 GetMove 返回 unavailable

	cacheMu sync.RWMutex
	vixHist []cboe.VIXSnapshot
	vixAt   time.Time
	dvol    map[string][]deribit.DVOLPoint
	dvolAt  map[string]time.Time
	move    *firecrawl.MoveSnapshot
	moveAt  time.Time
}

const cacheTTL = 5 * time.Minute

// NewService 用默认 apiclient.Source 构造客户端。
func NewService() *Service {
	return NewServiceWith(
		cboe.NewClient(apiclient.NewSource("cboe", apiclient.Options{Timeout: 30 * time.Second})),
		deribit.NewClient(apiclient.NewSource("deribit", apiclient.Options{Timeout: 30 * time.Second})),
	)
}

// NewServiceWith 允许注入测试 client。
func NewServiceWith(cb *cboe.Client, db *deribit.Client) *Service {
	return &Service{
		cboe: cb, deribit: db,
		dvol:   map[string][]deribit.DVOLPoint{},
		dvolAt: map[string]time.Time{},
	}
}

// WithFirecrawl 注入 firecrawl client，用于 GetMove 抓 yardeni MOVE 报告。
func (s *Service) WithFirecrawl(fc *firecrawl.Client) *Service {
	s.fire = fc
	return s
}

// loadVIX 拉取并缓存 VIX 历史。
func (s *Service) loadVIX(ctx context.Context) ([]cboe.VIXSnapshot, error) {
	s.cacheMu.RLock()
	if len(s.vixHist) > 0 && time.Since(s.vixAt) < cacheTTL {
		out := s.vixHist
		s.cacheMu.RUnlock()
		return out, nil
	}
	s.cacheMu.RUnlock()
	hist, err := s.cboe.FetchVIXHistory(ctx)
	if err != nil {
		return nil, err
	}
	if len(hist) == 0 {
		return nil, ErrUnavailable
	}
	s.cacheMu.Lock()
	s.vixHist = hist
	s.vixAt = time.Now()
	s.cacheMu.Unlock()
	return hist, nil
}

// loadDVOL 拉取并缓存指定资产的 Deribit DVOL（默认 30 天日线）。
func (s *Service) loadDVOL(ctx context.Context, asset string) ([]deribit.DVOLPoint, error) {
	asset = strings.ToUpper(strings.TrimSpace(asset))
	if asset == "" {
		asset = "BTC"
	}
	s.cacheMu.RLock()
	pts, ok := s.dvol[asset]
	at := s.dvolAt[asset]
	s.cacheMu.RUnlock()
	if ok && len(pts) > 0 && time.Since(at) < cacheTTL {
		return pts, nil
	}
	pts, err := s.deribit.GetVolatilityIndexData(ctx, asset, 86400, 0, 0)
	if err != nil {
		return nil, err
	}
	if len(pts) == 0 {
		return nil, ErrUnavailable
	}
	s.cacheMu.Lock()
	s.dvol[asset] = pts
	s.dvolAt[asset] = time.Now()
	s.cacheMu.Unlock()
	return pts, nil
}

// percentile 在历史样本里返回 v 落点的百分位（0-100）。
func percentile(v float64, hist []float64) float64 {
	if len(hist) == 0 {
		return 0
	}
	c := 0
	for _, h := range hist {
		if h <= v {
			c++
		}
	}
	return float64(c) / float64(len(hist)) * 100
}

// regimeFromVIX 按业界惯用阈值分类。
func regimeFromVIX(v float64) string {
	switch {
	case v < 13:
		return "low"
	case v < 20:
		return "normal"
	case v < 30:
		return "high"
	default:
		return "extreme"
	}
}

// GetVix 返回真实 VIX：当前值 + 30 日序列 + 百分位 + regime + 期限结构（用 30 日变化代理）。
func (s *Service) GetVix(ctx context.Context) (*volv1.GetVixResponse, error) {
	hist, err := s.loadVIX(ctx)
	if err != nil {
		return nil, err
	}
	last := hist[len(hist)-1]
	closes30 := lastN(hist, 30)
	pct := percentile(last.Close, closes30)
	regime := regimeFromVIX(last.Close)
	// 期限结构代理：当前 - 30 日均值。
	mean30 := mean(closes30)
	term := 0.0
	if mean30 > 0 {
		term = (last.Close - mean30) / mean30
	}
	out := &volv1.GetVixResponse{
		Vix: &volv1.VixData{
			Timestamp:      last.Date.Format(time.RFC3339),
			Spot:           last.Close,
			TermStructure:  term,
			Percentile_30D: pct,
			Regime:         regime,
		},
	}
	for _, p := range tailN(hist, 30) {
		out.History = append(out.History, &volv1.VixData{
			Timestamp:      p.Date.Format(time.RFC3339),
			Spot:           p.Close,
			TermStructure:  term,
			Percentile_30D: pct,
			Regime:         regimeFromVIX(p.Close),
		})
	}
	return out, nil
}

// GetMove 通过 firecrawl 抓 yardeni 公开 MOVE 报告。fire=nil 或 IsAvailable=false 时返回 unavailable。
// 缓存 6 小时，避免重复扫描 PDF。
func (s *Service) GetMove(ctx context.Context) (*volv1.GetMoveResponse, error) {
	if s.fire == nil || !s.fire.IsAvailable() {
		return nil, fmt.Errorf("MOVE index: %w (firecrawl client unavailable)", ErrUnavailable)
	}
	s.cacheMu.RLock()
	if s.move != nil && time.Since(s.moveAt) < 6*time.Hour {
		out := s.move
		s.cacheMu.RUnlock()
		return &volv1.GetMoveResponse{Move: &volv1.MoveData{
			Timestamp: out.Date.UTC().Format(time.RFC3339),
			Value:     out.Value,
			Trend:     out.Trend,
		}}, nil
	}
	s.cacheMu.RUnlock()
	snap, err := s.fire.FetchMove(ctx)
	if err != nil {
		return nil, fmt.Errorf("MOVE fetch: %w", err)
	}
	if snap == nil || snap.Value <= 0 {
		return nil, fmt.Errorf("MOVE: empty extraction result: %w", ErrUnavailable)
	}
	s.cacheMu.Lock()
	s.move = snap
	s.moveAt = time.Now()
	s.cacheMu.Unlock()
	return &volv1.GetMoveResponse{Move: &volv1.MoveData{
		Timestamp: snap.Date.UTC().Format(time.RFC3339),
		Value:     snap.Value,
		Trend:     snap.Trend,
	}}, nil
}

// GetDvol 返回指定资产的 Deribit DVOL 真值与最近趋势。
func (s *Service) GetDvol(ctx context.Context, asset string) (*volv1.GetDvolResponse, error) {
	pts, err := s.loadDVOL(ctx, asset)
	if err != nil {
		return nil, err
	}
	last := pts[len(pts)-1]
	return &volv1.GetDvolResponse{
		Dvol: &volv1.DvolData{
			Timestamp: time.UnixMilli(last.Timestamp).UTC().Format(time.RFC3339),
			Value:     last.Close,
			Asset:     strings.ToUpper(strings.TrimSpace(asset)),
		},
	}, nil
}

// GetGex 用 Deribit BookSummaries 估算 BTC/ETH 期权 GEX，flip 取最近 net=0 跨越点。
// 非加密 pair（如 EURUSD）返回 unavailable。
func (s *Service) GetGex(ctx context.Context, pair string) (*volv1.GetGexResponse, error) {
	currency, ok := pairToCurrency(pair)
	if !ok {
		return nil, fmt.Errorf("vol gex: pair %q not supported (only BTC/ETH/SOL)", pair)
	}
	books, err := s.deribit.GetBookSummaries(ctx, currency)
	if err != nil {
		return nil, err
	}
	if len(books) == 0 {
		return nil, ErrUnavailable
	}
	type bucket struct{ call, put float64 }
	bk := map[float64]*bucket{}
	var spot float64
	for _, b := range books {
		if b.UnderlyingPrice > 0 && spot == 0 {
			spot = b.UnderlyingPrice
		}
		strike, opt, ok := splitDeribitName(b.InstrumentName)
		if !ok || b.MarkIV <= 0 || b.OpenInterest <= 0 {
			continue
		}
		gex := b.OpenInterest * b.MarkIV * b.MarkIV * 1e-4
		v, exists := bk[strike]
		if !exists {
			v = &bucket{}
			bk[strike] = v
		}
		switch opt {
		case "C":
			v.call += gex
		case "P":
			v.put += gex
		}
	}
	type kv struct {
		strike float64
		net    float64
	}
	rows := make([]kv, 0, len(bk))
	var net float64
	for k, v := range bk {
		n := v.call - v.put
		net += n
		rows = append(rows, kv{k, n})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].strike < rows[j].strike })
	// flip = 最接近 spot 的 strike 处累计 net 跨越 0 的点
	flip := spot
	cum := 0.0
	for _, r := range rows {
		cum += r.net
		if cum >= 0 {
			flip = r.strike
			break
		}
	}
	wall := "neutral"
	if net > 0 {
		wall = "call_wall"
	} else if net < 0 {
		wall = "put_wall"
	}
	return &volv1.GetGexResponse{Gex: &volv1.GexData{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Pair:      strings.ToUpper(pair),
		NetGex:    net,
		FlipPoint: flip,
		Wall:      wall,
	}}, nil
}

// GetIvol 返回 Deribit 期权 IV 表面（按 strike 升序，附 Black-Scholes Greeks）。
func (s *Service) GetIvol(ctx context.Context, pair, expiry string) (*volv1.GetIvolResponse, error) {
	currency, ok := pairToCurrency(pair)
	if !ok {
		return nil, fmt.Errorf("vol ivol: pair %q not supported", pair)
	}
	books, err := s.deribit.GetBookSummaries(ctx, currency)
	if err != nil {
		return nil, err
	}
	if len(books) == 0 {
		return nil, ErrUnavailable
	}
	insts, err := s.deribit.GetInstruments(ctx, currency)
	if err != nil {
		return nil, err
	}
	expByName := map[string]int64{}
	strikeByName := map[string]float64{}
	for _, ins := range insts {
		expByName[ins.InstrumentName] = ins.ExpirationTs
		strikeByName[ins.InstrumentName] = ins.Strike
	}
	type ivRow struct {
		strike float64
		iv     float64
		dte    float64
		isCall bool
	}
	var rows []ivRow
	var spot, atmIV float64
	var atmDist = math.MaxFloat64
	now := time.Now()
	for _, b := range books {
		if b.MarkIV <= 0 || b.UnderlyingPrice <= 0 {
			continue
		}
		strike, opt, ok := splitDeribitName(b.InstrumentName)
		if !ok {
			continue
		}
		if spot == 0 {
			spot = b.UnderlyingPrice
		}
		exp := expByName[b.InstrumentName]
		dte := math.Max(1, time.UnixMilli(exp).Sub(now).Hours()/24)
		if d := math.Abs(strike - spot); d < atmDist {
			atmDist = d
			atmIV = b.MarkIV
		}
		rows = append(rows, ivRow{strike: strike, iv: b.MarkIV / 100, dte: dte, isCall: opt == "C"})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].strike < rows[j].strike })
	out := &volv1.GetIvolResponse{Pair: strings.ToUpper(pair), AtmIv: atmIV / 100}
	for _, r := range rows {
		tau := r.dte / 365.0
		out.Surface = append(out.Surface, &volv1.IvolPoint{
			Strike: r.strike,
			Iv:     r.iv,
			Delta:  bsDelta(r.strike, spot, r.iv, tau, r.isCall),
			Gamma:  bsGamma(r.strike, spot, r.iv, tau),
			Theta:  bsTheta(r.strike, spot, r.iv, tau),
			Vega:   bsVega(r.strike, spot, r.iv, tau),
		})
	}
	return out, nil
}

// GetSkew 真值：从 BookSummaries 以 ATM 为基准近似 25-delta RR/BF。
func (s *Service) GetSkew(ctx context.Context, pair string) (*volv1.GetSkewResponse, error) {
	currency, ok := pairToCurrency(pair)
	if !ok {
		return nil, fmt.Errorf("vol skew: pair %q not supported", pair)
	}
	books, err := s.deribit.GetBookSummaries(ctx, currency)
	if err != nil {
		return nil, err
	}
	if len(books) == 0 {
		return nil, ErrUnavailable
	}
	var callIV, putIV, atmIV float64
	var nC, nP, nA int
	for _, b := range books {
		if b.MarkIV <= 0 {
			continue
		}
		switch {
		case strings.HasSuffix(b.InstrumentName, "-C"):
			callIV += b.MarkIV
			nC++
		case strings.HasSuffix(b.InstrumentName, "-P"):
			putIV += b.MarkIV
			nP++
		}
		atmIV += b.MarkIV
		nA++
	}
	avgC := safeAvg(callIV, nC)
	avgP := safeAvg(putIV, nP)
	avgATM := safeAvg(atmIV, nA)
	rr := (avgC - avgP) / 100
	bf := ((avgC+avgP)/2 - avgATM) / 100
	term := "flat"
	switch {
	case rr > 1.0/100:
		term = "call_skew"
	case rr < -1.0/100:
		term = "put_skew"
	}
	return &volv1.GetSkewResponse{Skew: &volv1.SkewData{
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		Pair:          strings.ToUpper(pair),
		RiskReversal:  rr,
		Fly:           bf,
		TermStructure: term,
	}}, nil
}

// GetSkewVixAlert 把 GetSkew + 当前 VIX/DVOL 相结合给出告警。
func (s *Service) GetSkewVixAlert(ctx context.Context, pair string) (*volv1.GetSkewVixAlertResponse, error) {
	skew, err := s.GetSkew(ctx, pair)
	if err != nil {
		return nil, err
	}
	// VIX 不可达时退化为只看 DVOL（加密 pair）。
	var atmVol float64
	if hist, err := s.loadVIX(ctx); err == nil && len(hist) > 0 {
		atmVol = hist[len(hist)-1].Close
	}
	if atmVol == 0 {
		if dvol, err := s.GetDvol(ctx, strings.ToUpper(pair[:3])); err == nil && dvol.Dvol != nil {
			atmVol = dvol.Dvol.Value
		}
	}
	out := &volv1.GetSkewVixAlertResponse{}
	low := atmVol > 0 && atmVol < 20
	high := atmVol > 30
	if low && skew.Skew.RiskReversal < -2.0/100 {
		out.Alerts = append(out.Alerts, &volv1.SkewVixAlert{
			AlertId:    fmt.Sprintf("sv-%s-%d", pair, time.Now().Unix()),
			Timestamp:  time.Now().UTC().Format(time.RFC3339),
			Pair:       strings.ToUpper(pair),
			Signal:     "complacency_warning",
			Confidence: 0.75,
			Reason:     "Elevated put skew with low VIX/DVOL suggests complacency",
		})
	}
	if high && skew.Skew.RiskReversal > 2.0/100 {
		out.Alerts = append(out.Alerts, &volv1.SkewVixAlert{
			AlertId:    fmt.Sprintf("sv-%s-%d", pair, time.Now().Unix()),
			Timestamp:  time.Now().UTC().Format(time.RFC3339),
			Pair:       strings.ToUpper(pair),
			Signal:     "euphoria_warning",
			Confidence: 0.65,
			Reason:     "Elevated call skew with high VIX/DVOL suggests euphoria",
		})
	}
	return out, nil
}

// ====== 工具函数 ======

// pairToCurrency 把内部 pair（BTCUSD/ETHUSD/SOLUSD）映射到 Deribit currency 代码。
func pairToCurrency(pair string) (string, bool) {
	p := strings.ToUpper(strings.TrimSpace(pair))
	switch {
	case strings.HasPrefix(p, "BTC"):
		return "BTC", true
	case strings.HasPrefix(p, "ETH"):
		return "ETH", true
	case strings.HasPrefix(p, "SOL"):
		return "SOL", true
	}
	return "", false
}

// splitDeribitName 拆解 "BTC-25APR25-65000-C" → strike, "C"/"P"。
func splitDeribitName(n string) (float64, string, bool) {
	parts := strings.Split(n, "-")
	if len(parts) < 4 {
		return 0, "", false
	}
	if parts[3] != "C" && parts[3] != "P" {
		return 0, "", false
	}
	var s float64
	for _, c := range parts[2] {
		if c < '0' || c > '9' {
			return 0, "", false
		}
		s = s*10 + float64(c-'0')
	}
	return s, parts[3], true
}

func lastN(hist []cboe.VIXSnapshot, n int) []float64 {
	if len(hist) < n {
		n = len(hist)
	}
	out := make([]float64, 0, n)
	for _, p := range hist[len(hist)-n:] {
		out = append(out, p.Close)
	}
	return out
}

func tailN(hist []cboe.VIXSnapshot, n int) []cboe.VIXSnapshot {
	if len(hist) < n {
		return hist
	}
	return hist[len(hist)-n:]
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

func safeAvg(sum float64, n int) float64 {
	if n == 0 {
		return 0
	}
	return sum / float64(n)
}

// ====== 简化 Black-Scholes Greeks ======

func bsDelta(strike, spot, vol, tau float64, isCall bool) float64 {
	if vol <= 0 || tau <= 0 || spot <= 0 {
		return 0
	}
	d1 := (math.Log(spot/strike) + 0.5*vol*vol*tau) / (vol * math.Sqrt(tau))
	if isCall {
		return 0.5 + 0.5*math.Erf(d1/math.Sqrt(2))
	}
	return -0.5 + 0.5*math.Erf(-d1/math.Sqrt(2))
}

func bsGamma(strike, spot, vol, tau float64) float64 {
	if vol <= 0 || tau <= 0 || spot <= 0 {
		return 0
	}
	d1 := (math.Log(spot/strike) + 0.5*vol*vol*tau) / (vol * math.Sqrt(tau))
	return math.Exp(-d1*d1/2) / (spot * vol * math.Sqrt(2*math.Pi*tau))
}

func bsTheta(strike, spot, vol, tau float64) float64 {
	if tau <= 0 {
		return 0
	}
	return -spot * vol / (2 * math.Sqrt(2*math.Pi*tau)) * 0.01
}

func bsVega(strike, spot, vol, tau float64) float64 {
	if vol <= 0 || tau <= 0 || spot <= 0 {
		return 0
	}
	d1 := (math.Log(spot/strike) + 0.5*vol*vol*tau) / (vol * math.Sqrt(tau))
	return spot * math.Sqrt(tau/(2*math.Pi)) * math.Exp(-d1*d1/2) * 0.01
}
