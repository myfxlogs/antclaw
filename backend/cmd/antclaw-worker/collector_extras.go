// 扩展数据采集
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ========== 1. 分时价格采集 ==========

var intradayPairs = []struct {
	Yahoo, Symbol string
}{
	{"EURUSD=X", "EURUSD"}, {"GBPUSD=X", "GBPUSD"}, {"USDJPY=X", "USDJPY"},
	{"^VIX", "VIX"}, {"^GSPC", "SP500"}, {"GC=F", "XAUUSD"}, {"CL=F", "CRUDE"},
}

func collectIntraday(ctx context.Context, dbpool *pgxpool.Pool, logger *slog.Logger) error {
	logger.Info("Starting intraday collection: 5-minute candles")
	total := 0
	for _, p := range intradayPairs {
		bars, err := fetchYahoo(ctx, p.Yahoo, "5d", "5m")
		if err != nil {
			logger.Warn("Intraday fetch failed", "symbol", p.Symbol, "error", err)
			continue
		}
		count := 0
		for _, b := range bars {
			_, err := dbpool.Exec(ctx, `
				INSERT INTO price_intraday (time, symbol, interval, open, high, low, close, volume, source)
				VALUES ($1, $2, '5m', $3, $4, $5, $6, $7, 'yahoo')
				ON CONFLICT (time, symbol, interval) DO UPDATE SET
				  close = EXCLUDED.close, volume = EXCLUDED.volume`,
				b.Time, p.Symbol, b.Open, b.High, b.Low, b.Close, b.Volume)
			if err == nil {
				count++
			}
		}
		total += count
		logger.Info("Intraday synced", "symbol", p.Symbol, "bars", count)
		time.Sleep(500 * time.Millisecond)
	}
	logger.Info("Intraday collection completed", "total", total)
	return nil
}

// ========== 2. DeFi采集 ==========

func collectDefi(ctx context.Context, dbpool *pgxpool.Pool, logger *slog.Logger) error {
	logger.Info("Starting DeFi collection from DefiLlama")

	// 历史TVL
	tvlURL := "https://api.llama.fi/v2/historicalChainTvl"
	req, _ := http.NewRequestWithContext(ctx, "GET", tvlURL, nil)
	resp, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	if err != nil {
		logger.Warn("DefiLlama fetch failed", "error", err)
		return fmt.Errorf("defillama fetch: %w", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	var tvlHistory []struct {
		Date int64   `json:"date"`
		TVL  float64 `json:"tvl"`
	}
	if err := json.Unmarshal(body, &tvlHistory); err != nil {
		logger.Warn("DefiLlama parse failed", "error", err)
		return fmt.Errorf("defillama parse: %w", err)
	}

	// 稳定币市值
	stableURL := "https://stablecoins.llama.fi/stablecoincharts/all?stablecoin=1"
	req2, _ := http.NewRequestWithContext(ctx, "GET", stableURL, nil)
	resp2, _ := (&http.Client{Timeout: 30 * time.Second}).Do(req2)
	var stableMC float64
	if resp2 != nil {
		b2, _ := io.ReadAll(resp2.Body)
		resp2.Body.Close()
		var stableData []struct {
			Date          int64 `json:"date"`
			TotalCirculating struct {
				PeggedUSD float64 `json:"peggedUSD"`
			} `json:"totalCirculating"`
		}
		if json.Unmarshal(b2, &stableData) == nil && len(stableData) > 0 {
			stableMC = stableData[len(stableData)-1].TotalCirculating.PeggedUSD
		}
	}

	// 采集最近90天
	count := 0
	total := len(tvlHistory)
	start := 0
	if total > 90 {
		start = total - 90
	}
	for i := start; i < total; i++ {
		cur := tvlHistory[i]
		t := time.Unix(cur.Date, 0)

		var chg24, chg7d float64
		if i > 0 && tvlHistory[i-1].TVL > 0 {
			chg24 = (cur.TVL - tvlHistory[i-1].TVL) / tvlHistory[i-1].TVL * 100
		}
		if i > 6 && tvlHistory[i-7].TVL > 0 {
			chg7d = (cur.TVL - tvlHistory[i-7].TVL) / tvlHistory[i-7].TVL * 100
		}
		raw, _ := json.Marshal(cur)
		_, err := dbpool.Exec(ctx, `
			INSERT INTO defi_snapshots (time, total_tvl, tvl_change_24h, tvl_change_7d, stablecoin_mc, raw)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (time) DO UPDATE SET
			  total_tvl = EXCLUDED.total_tvl,
			  tvl_change_24h = EXCLUDED.tvl_change_24h,
			  tvl_change_7d = EXCLUDED.tvl_change_7d,
			  stablecoin_mc = EXCLUDED.stablecoin_mc`,
			t, cur.TVL, chg24, chg7d, stableMC, raw)
		if err == nil {
			count++
		}
	}
	var latestTVL float64
	if total > 0 {
		latestTVL = tvlHistory[total-1].TVL
	}
	logger.Info("DeFi collection completed", "records", count, "latest_tvl", latestTVL)
	return nil
}

// ========== 3. VIX期限结构 (Yahoo多指数) ==========

func collectVIXTermStructure(ctx context.Context, dbpool *pgxpool.Pool, logger *slog.Logger) error {
	logger.Info("Starting VIX term structure collection")

	symbols := map[string]string{
		"spot":  "^VIX", "vix9d": "^VIX9D", "vix3m": "^VIX3M",
		"m1":    "^VIX", // 近月期货代理
		"skew":  "^SKEW",
		"move":  "^MOVE",
	}

	// 按日期收集所有指数
	dailyData := map[time.Time]map[string]float64{}
	for field, sym := range symbols {
		bars, err := fetchYahoo(ctx, sym, "3mo", "1d")
		if err != nil {
			logger.Debug("VIX index fetch failed", "field", field, "error", err)
			continue
		}
		for _, b := range bars {
			day := b.Time.UTC().Truncate(24 * time.Hour)
			if dailyData[day] == nil {
				dailyData[day] = map[string]float64{}
			}
			dailyData[day][field] = b.Close
		}
		time.Sleep(500 * time.Millisecond)
	}

	count := 0
	for day, vals := range dailyData {
		spot := vals["spot"]
		vix9d := vals["vix9d"]
		vix3m := vals["vix3m"]
		// contango: 近期 < 远期 (vix3m > spot)
		contango := vix3m > spot
		regime := "normal"
		if vix9d > spot*1.1 {
			regime = "near_fear"
		} else if spot > vix3m*1.1 {
			regime = "backwardation"
		} else if contango {
			regime = "contango"
		}
		raw, _ := json.Marshal(vals)
		_, err := dbpool.Exec(ctx, `
			INSERT INTO vix_term_structure (time, spot, m1, vix9d, vix3m, skew, move, contango, cross_regime, raw)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			ON CONFLICT (time) DO UPDATE SET
			  spot = EXCLUDED.spot, vix9d = EXCLUDED.vix9d, vix3m = EXCLUDED.vix3m,
			  contango = EXCLUDED.contango, cross_regime = EXCLUDED.cross_regime`,
			day, spot, vals["m1"], vix9d, vix3m, vals["skew"], vals["move"], contango, regime, raw)
		if err == nil {
			count++
		}
	}
	logger.Info("VIX term structure collection completed", "records", count)
	return nil
}

// ========== 4. DVOL采集 (Deribit) ==========

func collectDVOL(ctx context.Context, dbpool *pgxpool.Pool, logger *slog.Logger) error {
	logger.Info("Starting DVOL implied volatility collection from Deribit")

	currencies := []string{"BTC", "ETH"}
	count := 0
	for _, ccy := range currencies {
		// Deribit DVOL历史
		to := time.Now().Unix() * 1000
		from := time.Now().AddDate(0, -3, 0).Unix() * 1000
		url := fmt.Sprintf(
			"https://www.deribit.com/api/v2/public/get_volatility_index_data?currency=%s&resolution=1D&start_timestamp=%d&end_timestamp=%d",
			ccy, from, to)

		req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
		resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
		if err != nil {
			logger.Warn("DVOL fetch failed", "currency", ccy, "error", err)
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var data struct {
			Result struct {
				Data [][]float64 `json:"data"` // [timestamp, open, high, low, close]
			} `json:"result"`
		}
		if err := json.Unmarshal(body, &data); err != nil {
			logger.Warn("DVOL parse failed", "currency", ccy, "error", err)
			continue
		}

		series := data.Result.Data
		for i, row := range series {
			if len(row) < 5 {
				continue
			}
			t := time.Unix(int64(row[0])/1000, 0)
			closeIV := row[4]
			var change24 float64
			if i > 0 && series[i-1][4] > 0 {
				change24 = (closeIV - series[i-1][4]) / series[i-1][4] * 100
			}
			spike := change24 > 10

			_, err := dbpool.Exec(ctx, `
				INSERT INTO dvol_snapshots (time, currency, current_iv, change_24h_pct, spike)
				VALUES ($1, $2, $3, $4, $5)
				ON CONFLICT (time, currency) DO UPDATE SET
				  current_iv = EXCLUDED.current_iv,
				  change_24h_pct = EXCLUDED.change_24h_pct,
				  spike = EXCLUDED.spike`,
				t, ccy, closeIV, change24, spike)
			if err == nil {
				count++
			}
		}
		logger.Info("DVOL synced", "currency", ccy, "records", len(series))
	}
	logger.Info("DVOL collection completed", "total", count)
	return nil
}
