// 价格采集器
package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	apiclient "github.com/antclaw/antclaw/internal/infra/apiclient"
	yhoo "github.com/antclaw/antclaw/internal/infra/apiclient/yahoo"
)

var pricePairs = []struct {
	Yahoo  string
	Symbol string
}{
	{"EURUSD=X", "EURUSD"}, {"GBPUSD=X", "GBPUSD"}, {"USDJPY=X", "USDJPY"},
	{"AUDUSD=X", "AUDUSD"}, {"USDCAD=X", "USDCAD"}, {"USDCHF=X", "USDCHF"},
	{"NZDUSD=X", "NZDUSD"}, {"GC=F", "XAUUSD"}, {"SI=F", "XAGUSD"},
	{"CL=F", "CRUDE"}, {"^VIX", "VIX"}, {"^GSPC", "SP500"},
	{"^DJI", "DJIA"}, {"^IXIC", "NASDAQ"},
}

func collectPrices(ctx context.Context, dbpool *pgxpool.Pool, logger *slog.Logger) error {
	logger.Info("Starting price collection from Yahoo Finance")

	totalInserted := 0
	for _, pair := range pricePairs {
		bars, err := fetchYahoo(ctx, pair.Yahoo, "1y", "1d")
		if err != nil || len(bars) == 0 {
			logger.Warn("Yahoo daily fetch failed, trying Stooq", "symbol", pair.Symbol)
			bars, err = fetchStooqDaily(ctx, strings.ToLower(strings.TrimSuffix(pair.Yahoo, "=X")))
			if err != nil || len(bars) == 0 {
				logger.Warn("Price fetch failed for all sources", "symbol", pair.Symbol)
				continue
			}
		}

		inserted := savePriceBars(ctx, dbpool, pair.Symbol, "yahoo", bars)
		totalInserted += inserted
		logger.Info("Price synced", "symbol", pair.Symbol, "bars", inserted)
		time.Sleep(500 * time.Millisecond)
	}

	logger.Info("Price collection completed", "total_inserted", totalInserted)
	return nil
}

type priceBar struct {
	Time   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
}

// fetchYahoo 通用Yahoo chart获取（使用 apiclient.Source + yahoo 薄客户端）
func fetchYahoo(ctx context.Context, symbol, rangeStr, interval string) ([]priceBar, error) {
	src := apiclient.NewSource("yahoo", apiclient.Options{Timeout: 15 * time.Second})
	cli := yhoo.NewClient(src)
	data, err := cli.GetChart(ctx, symbol, rangeStr, interval)
	if err != nil {
		return nil, err
	}
	if len(data.Chart.Result) == 0 || len(data.Chart.Result[0].Indicators.Quote) == 0 {
		return nil, fmt.Errorf("empty yahoo response")
	}
	r := data.Chart.Result[0]
	q := r.Indicators.Quote[0]
	bars := make([]priceBar, 0, len(r.Timestamp))
	for i, ts := range r.Timestamp {
		if i >= len(q.Close) || q.Close[i] == 0 {
			continue
		}
		bars = append(bars, priceBar{Time: time.Unix(ts, 0), Open: q.Open[i], High: q.High[i], Low: q.Low[i], Close: q.Close[i], Volume: q.Volume[i]})
	}
	return bars, nil
}

func fetchStooqDaily(ctx context.Context, symbol string) ([]priceBar, error) {
	url := fmt.Sprintf("https://stooq.com/q/d/l/?s=%s&i=d", symbol)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("stooq status %d", resp.StatusCode)
	}

	reader := csv.NewReader(resp.Body)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("stooq no data")
	}

	var bars []priceBar
	for i, row := range records {
		if i == 0 || len(row) < 5 {
			continue
		}
		t, err := time.Parse("2006-01-02", row[0])
		if err != nil {
			continue
		}
		open, _ := strconv.ParseFloat(row[1], 64)
		high, _ := strconv.ParseFloat(row[2], 64)
		low, _ := strconv.ParseFloat(row[3], 64)
		cls, _ := strconv.ParseFloat(row[4], 64)
		var vol float64
		if len(row) > 5 {
			vol, _ = strconv.ParseFloat(row[5], 64)
		}
		if cls == 0 {
			continue
		}
		bars = append(bars, priceBar{Time: t, Open: open, High: high, Low: low, Close: cls, Volume: vol})
	}
	return bars, nil
}

func savePriceBars(ctx context.Context, dbpool *pgxpool.Pool, symbol, source string, bars []priceBar) int {
	count := 0
	for _, b := range bars {
		_, err := dbpool.Exec(ctx, `
			INSERT INTO price_daily (time, symbol, open, high, low, close, volume, source)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (time, symbol) DO UPDATE SET
			  open = EXCLUDED.open, high = EXCLUDED.high, low = EXCLUDED.low,
			  close = EXCLUDED.close, volume = EXCLUDED.volume, source = EXCLUDED.source`,
			b.Time, symbol, b.Open, b.High, b.Low, b.Close, b.Volume, source)
		if err == nil {
			count++
		}
	}
	return count
}

func parseInt64(s string) int64 {
	if s = strings.TrimSpace(s); s == "" {
		return 0
	}
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}
