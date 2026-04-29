// Package alphavantage 极简 AlphaVantage 客户端。
// 文档：https://www.alphavantage.co/documentation/
//
// 暴露 FetchOHLC（FX_DAILY / FX_INTRADAY）。
package alphavantage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/antclaw/antclaw/internal/infra/apiclient"
)

type Client struct {
	src    apiclient.Source
	apiKey string
	base   string
}

func NewClient(src apiclient.Source, apiKey string) *Client {
	return &Client{src: src, apiKey: strings.TrimSpace(apiKey), base: "https://www.alphavantage.co/query"}
}

func (c *Client) Available() bool { return c.apiKey != "" }
func (c *Client) Name() string    { return "alphavantage" }

type Bar struct {
	Time   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume int64
}

// FetchOHLC FX 与 Crypto 都返回相同结构。symbol = 6 位 FX (EURUSD) 或 BTCUSD。
// timeframe: "1d" 用 FX_DAILY；其他用 FX_INTRADAY。
func (c *Client) FetchOHLC(ctx context.Context, symbol, timeframe string, outputSize int) ([]Bar, error) {
	if !c.Available() {
		return nil, fmt.Errorf("alphavantage: api key not configured")
	}
	sym := strings.ToUpper(strings.TrimSpace(symbol))
	if len(sym) != 6 {
		return nil, fmt.Errorf("alphavantage: only 6-char FX symbols supported, got %q", sym)
	}
	from, to := sym[:3], sym[3:]
	tf := strings.ToLower(strings.TrimSpace(timeframe))
	q := url.Values{}
	q.Set("apikey", c.apiKey)
	q.Set("from_symbol", from)
	q.Set("to_symbol", to)
	q.Set("outputsize", "compact")
	if outputSize > 100 {
		q.Set("outputsize", "full")
	}
	var rootKey string
	switch tf {
	case "", "1d", "d", "daily":
		q.Set("function", "FX_DAILY")
		rootKey = "Time Series FX (Daily)"
	case "1h":
		q.Set("function", "FX_INTRADAY")
		q.Set("interval", "60min")
		rootKey = "Time Series FX (60min)"
	case "5m":
		q.Set("function", "FX_INTRADAY")
		q.Set("interval", "5min")
		rootKey = "Time Series FX (5min)"
	default:
		return nil, fmt.Errorf("alphavantage: unsupported timeframe %q", tf)
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, c.base+"?"+q.Encode(), nil)
	resp, err := c.src.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var raw map[string]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	if msg, ok := raw["Error Message"]; ok {
		return nil, fmt.Errorf("alphavantage: %s", string(msg))
	}
	tsData, ok := raw[rootKey]
	if !ok {
		// 命中速率限制时返回 "Note" 字段
		if note, ok := raw["Note"]; ok {
			return nil, fmt.Errorf("alphavantage: %s", string(note))
		}
		return nil, fmt.Errorf("alphavantage: missing key %q in response", rootKey)
	}
	var series map[string]map[string]string
	if err := json.Unmarshal(tsData, &series); err != nil {
		return nil, err
	}
	bars := make([]Bar, 0, len(series))
	layout := "2006-01-02"
	if tf != "1d" && tf != "d" && tf != "daily" && tf != "" {
		layout = "2006-01-02 15:04:05"
	}
	for ts, row := range series {
		t, err := time.Parse(layout, ts)
		if err != nil {
			continue
		}
		o, _ := strconv.ParseFloat(row["1. open"], 64)
		h, _ := strconv.ParseFloat(row["2. high"], 64)
		l, _ := strconv.ParseFloat(row["3. low"], 64)
		cl, _ := strconv.ParseFloat(row["4. close"], 64)
		if cl <= 0 {
			continue
		}
		bars = append(bars, Bar{Time: t.UTC(), Open: o, High: h, Low: l, Close: cl})
	}
	// 按时间升序排序
	for i := 0; i < len(bars); i++ {
		for j := i + 1; j < len(bars); j++ {
			if bars[j].Time.Before(bars[i].Time) {
				bars[i], bars[j] = bars[j], bars[i]
			}
		}
	}
	return bars, nil
}
