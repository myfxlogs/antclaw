// Package main 中期权数据采集（Deribit 公开 API，免鉴权）：
//   - gex-snapshot : Total GEX、Flip Level、Call Wall、Put Wall。
//   - iv-skew      : 5-point smile + skew slope + term slope。
//
// API：
//   GET https://www.deribit.com/api/v2/public/get_book_summary_by_currency?currency=BTC&kind=option
//   返回每个期权合约的 mark_iv / mark_price / open_interest / underlying_price。
//
// 计算简化（教科书）：
//   gamma_proxy ≈ exp(-d1²/2) / (S * σ * √(T-t))，d1 用 ln(S/K)/(σ√T) + 0.5σ√T
//   GEX_i = gamma_i * OI_i * sign  （call:+, put:-）
//   total_gex = Σ GEX_i / 1e9        （亿美元口径）
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const deribitBookURL = "https://www.deribit.com/api/v2/public/get_book_summary_by_currency"

type deribitOption struct {
	InstrumentName  string  `json:"instrument_name"`
	MarkIV          float64 `json:"mark_iv"`
	MarkPrice       float64 `json:"mark_price"`
	OpenInterest    float64 `json:"open_interest"`
	UnderlyingPrice float64 `json:"underlying_price"`
}

func fetchDeribitOptions(ctx context.Context, currency string) ([]deribitOption, error) {
	url := fmt.Sprintf("%s?currency=%s&kind=option", deribitBookURL, currency)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var raw struct {
		Result []deribitOption `json:"result"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("deribit decode: %w", err)
	}
	return raw.Result, nil
}

// instrumentInfo 解析 BTC-30JUN26-50000-C 风格命名。
type instrumentInfo struct {
	currency string
	expiry   time.Time
	strike   float64
	isCall   bool
}

func parseInstrument(name string) (instrumentInfo, bool) {
	parts := strings.Split(name, "-")
	if len(parts) != 4 {
		return instrumentInfo{}, false
	}
	exp, err := time.Parse("2Jan06", parts[1])
	if err != nil {
		return instrumentInfo{}, false
	}
	var strike float64
	fmt.Sscanf(parts[2], "%f", &strike)
	return instrumentInfo{
		currency: parts[0],
		expiry:   exp,
		strike:   strike,
		isCall:   strings.EqualFold(parts[3], "C"),
	}, true
}

// collectGEX 计算两类标的的 GEX 快照。
func collectGEX(ctx context.Context, db *pgxpool.Pool, logger *slog.Logger) error {
	now := time.Now().UTC()
	for _, ccy := range []string{"BTC", "ETH"} {
		opts, err := fetchDeribitOptions(ctx, ccy)
		if err != nil {
			logger.Warn("gex-snapshot fetch failed", "currency", ccy, "error", err)
			continue
		}
		spot, totalGEX, flip, callWall, putWall, levels := computeGEX(opts, now)
		if spot == 0 {
			continue
		}
		levelsJSON, _ := json.Marshal(levels)
		_, err = db.Exec(ctx, `
			INSERT INTO gex_snapshots(time, symbol, spot_price, total_gex, flip_level, max_call_wall, max_put_wall, levels)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8::jsonb)
			ON CONFLICT (time, symbol) DO UPDATE SET
			   spot_price=EXCLUDED.spot_price, total_gex=EXCLUDED.total_gex,
			   flip_level=EXCLUDED.flip_level, max_call_wall=EXCLUDED.max_call_wall,
			   max_put_wall=EXCLUDED.max_put_wall, levels=EXCLUDED.levels`,
			now, ccy, spot, totalGEX, flip, callWall, putWall, string(levelsJSON))
		if err != nil {
			logger.Warn("gex-snapshot insert failed", "currency", ccy, "error", err)
		}
	}
	logger.Info("gex-snapshot done")
	return nil
}

func computeGEX(opts []deribitOption, now time.Time) (spot, totalGEX, flip, callWall, putWall float64, levels []map[string]any) {
	if len(opts) == 0 {
		return
	}
	type strikeAgg struct {
		strike   float64
		callGEX  float64
		putGEX   float64
	}
	agg := map[float64]*strikeAgg{}

	for _, o := range opts {
		info, ok := parseInstrument(o.InstrumentName)
		if !ok || o.UnderlyingPrice <= 0 {
			continue
		}
		spot = o.UnderlyingPrice
		t := info.expiry.Sub(now).Hours() / (24 * 365)
		if t <= 0 || o.MarkIV <= 0 {
			continue
		}
		sigma := o.MarkIV / 100
		d1 := (math.Log(spot/info.strike))/(sigma*math.Sqrt(t)) + 0.5*sigma*math.Sqrt(t)
		gamma := math.Exp(-d1*d1/2) / (spot * sigma * math.Sqrt(t) * math.Sqrt(2*math.Pi))
		gex := gamma * o.OpenInterest * spot * spot * 0.01 // 1% 移动下的美元 gamma exposure
		s := agg[info.strike]
		if s == nil {
			s = &strikeAgg{strike: info.strike}
			agg[info.strike] = s
		}
		if info.isCall {
			s.callGEX += gex
		} else {
			s.putGEX -= gex
		}
	}

	// 排序 + 累计
	var strikes []*strikeAgg
	for _, s := range agg {
		strikes = append(strikes, s)
	}
	sort.Slice(strikes, func(i, j int) bool { return strikes[i].strike < strikes[j].strike })

	flipDelta := math.Inf(1)
	for _, s := range strikes {
		net := s.callGEX + s.putGEX
		totalGEX += net
		if math.Abs(s.strike-spot)/spot < 0.20 && math.Abs(net) < flipDelta {
			flip = s.strike
			flipDelta = math.Abs(net)
		}
		if s.callGEX > callWall {
			callWall = s.callGEX
		}
		if -s.putGEX > putWall {
			putWall = -s.putGEX
		}
		levels = append(levels, map[string]any{
			"strike": s.strike, "call_gex": s.callGEX, "put_gex": s.putGEX, "net": net,
		})
	}
	return
}

// collectIVSkew 计算 IV smile / skew slope / term slope，写入 iv_skew_history。
func collectIVSkew(ctx context.Context, db *pgxpool.Pool, logger *slog.Logger) error {
	now := time.Now().UTC()
	for _, ccy := range []string{"BTC", "ETH"} {
		opts, err := fetchDeribitOptions(ctx, ccy)
		if err != nil {
			logger.Warn("iv-skew fetch failed", "currency", ccy, "error", err)
			continue
		}
		spot, smile, skewSlope, termSlope, pcRatio := computeIVSkew(opts, now)
		if spot == 0 {
			continue
		}
		smileJSON, _ := json.Marshal(smile)
		_, err = db.Exec(ctx, `
			INSERT INTO iv_skew_history(time, symbol, pc_iv_ratio, skew_slope, smile, term_slope)
			VALUES ($1,$2,$3,$4,$5::jsonb,$6)
			ON CONFLICT (time, symbol) DO UPDATE SET
			   pc_iv_ratio=EXCLUDED.pc_iv_ratio, skew_slope=EXCLUDED.skew_slope,
			   smile=EXCLUDED.smile, term_slope=EXCLUDED.term_slope`,
			now, ccy, pcRatio, skewSlope, string(smileJSON), termSlope)
		if err != nil {
			logger.Warn("iv-skew insert failed", "currency", ccy, "error", err)
		}
	}
	logger.Info("iv-skew done")
	return nil
}

// computeIVSkew 用前两个最近到期日的 ATM IV 算 term_slope，
// 用 5-point moneyness {0.9,0.95,1.0,1.05,1.1} 拟合 smile，
// 用 (call IV − put IV) 同 strike 对比近似 risk reversal 的 skew slope，
// 用 OI 加权 put/call IV 平均比作为 pc_iv_ratio。
func computeIVSkew(opts []deribitOption, now time.Time) (spot float64, smile map[string]float64, skewSlope, termSlope, pcRatio float64) {
	smile = map[string]float64{}
	type pt struct {
		money    float64 // K/S
		iv       float64
		oi       float64
		expiryYr float64
		isCall   bool
	}
	var pts []pt
	var atmEarly, atmLate, atmEarlyT, atmLateT float64

	for _, o := range opts {
		info, ok := parseInstrument(o.InstrumentName)
		if !ok || o.UnderlyingPrice <= 0 || o.MarkIV <= 0 {
			continue
		}
		spot = o.UnderlyingPrice
		t := info.expiry.Sub(now).Hours() / (24 * 365)
		if t <= 0 {
			continue
		}
		money := info.strike / spot
		pts = append(pts, pt{money: money, iv: o.MarkIV, oi: o.OpenInterest, expiryYr: t, isCall: info.isCall})
		if money > 0.95 && money < 1.05 {
			if atmEarly == 0 || t < atmEarlyT {
				atmLate, atmLateT = atmEarly, atmEarlyT
				atmEarly, atmEarlyT = o.MarkIV, t
			} else if atmLate == 0 || (t < atmLateT && t > atmEarlyT) {
				atmLate, atmLateT = o.MarkIV, t
			}
		}
	}

	if len(pts) == 0 {
		return
	}

	// 5-point smile：选最近到期日且最贴近 moneyness 的样本
	bands := []struct {
		key    string
		target float64
	}{{"0.90", 0.9}, {"0.95", 0.95}, {"1.00", 1.0}, {"1.05", 1.05}, {"1.10", 1.1}}
	for _, b := range bands {
		best := math.Inf(1)
		var iv float64
		for _, p := range pts {
			d := math.Abs(p.money - b.target)
			if d < best && d < 0.05 {
				best = d
				iv = p.iv
			}
		}
		smile[b.key] = iv
	}

	// skew slope = (iv_0.9 − iv_1.1) / 0.2
	if smile["0.90"] > 0 && smile["1.10"] > 0 {
		skewSlope = (smile["0.90"] - smile["1.10"]) / 0.2
	}

	// term slope = (atm_late − atm_early) / (T_late − T_early)
	if atmLate > 0 && atmEarly > 0 && atmLateT > atmEarlyT {
		termSlope = (atmLate - atmEarly) / (atmLateT - atmEarlyT)
	}

	// pcRatio = OI 加权 put IV / call IV
	var sumP, sumC, oiP, oiC float64
	for _, p := range pts {
		if p.isCall {
			sumC += p.iv * p.oi
			oiC += p.oi
		} else {
			sumP += p.iv * p.oi
			oiP += p.oi
		}
	}
	if oiC > 0 && oiP > 0 {
		pcRatio = (sumP / oiP) / (sumC / oiC)
	}
	return
}
