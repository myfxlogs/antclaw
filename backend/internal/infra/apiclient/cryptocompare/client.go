// Package cryptocompare 极简 CryptoCompare 客户端。
// 文档：https://min-api.cryptocompare.com/documentation
//
// 仅暴露 FetchOHLC（histoday / histohour / histominute）。
package cryptocompare

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
	apiKey string // 可选；空时仍可用公开端点
	base   string
}

func NewClient(src apiclient.Source, apiKey string) *Client {
	return &Client{src: src, apiKey: strings.TrimSpace(apiKey), base: "https://min-api.cryptocompare.com/data/v2"}
}

// Available CryptoCompare 公开端点无需 Key；Key 仅用于提升配额。
func (c *Client) Available() bool { return true }
func (c *Client) Name() string    { return "cryptocompare" }

type Bar struct {
	Time   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume int64
}

type histoResp struct {
	Response string `json:"Response"`
	Message  string `json:"Message"`
	Data     struct {
		Data []struct {
			Time   int64   `json:"time"`
			Open   float64 `json:"open"`
			High   float64 `json:"high"`
			Low    float64 `json:"low"`
			Close  float64 `json:"close"`
			Volume float64 `json:"volumefrom"`
		} `json:"Data"`
	} `json:"Data"`
}

// FetchOHLC symbol 形如 "BTCUSD" / "ETHUSDT"，自动拆分。outputSize 上限 2000。
func (c *Client) FetchOHLC(ctx context.Context, symbol, timeframe string, outputSize int) ([]Bar, error) {
	sym := strings.ToUpper(strings.TrimSpace(symbol))
	from, to, ok := splitCryptoSymbol(sym)
	if !ok {
		return nil, fmt.Errorf("cryptocompare: cannot split symbol %q", sym)
	}
	endpoint, ok := mapEndpoint(timeframe)
	if !ok {
		return nil, fmt.Errorf("cryptocompare: unsupported timeframe %q", timeframe)
	}
	if outputSize <= 0 || outputSize > 2000 {
		outputSize = 500
	}
	q := url.Values{}
	q.Set("fsym", from)
	q.Set("tsym", to)
	q.Set("limit", strconv.Itoa(outputSize))
	if c.apiKey != "" {
		q.Set("api_key", c.apiKey)
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, c.base+"/"+endpoint+"?"+q.Encode(), nil)
	resp, err := c.src.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var hr histoResp
	if err := json.NewDecoder(resp.Body).Decode(&hr); err != nil {
		return nil, err
	}
	if hr.Response == "Error" {
		return nil, fmt.Errorf("cryptocompare: %s", hr.Message)
	}
	out := make([]Bar, 0, len(hr.Data.Data))
	for _, b := range hr.Data.Data {
		if b.Close <= 0 {
			continue
		}
		out = append(out, Bar{
			Time:   time.Unix(b.Time, 0).UTC(),
			Open:   b.Open,
			High:   b.High,
			Low:    b.Low,
			Close:  b.Close,
			Volume: int64(b.Volume),
		})
	}
	return out, nil
}

// splitCryptoSymbol 优先匹配 4 字符法币尾缀（USDT/USDC/BUSD），再退到 3 字符（USD/EUR）。
func splitCryptoSymbol(s string) (string, string, bool) {
	for _, suf := range []string{"USDT", "USDC", "BUSD", "TUSD"} {
		if strings.HasSuffix(s, suf) && len(s) > len(suf) {
			return s[:len(s)-len(suf)], suf, true
		}
	}
	for _, suf := range []string{"USD", "EUR", "JPY", "BTC", "ETH"} {
		if strings.HasSuffix(s, suf) && len(s) > len(suf) {
			return s[:len(s)-len(suf)], suf, true
		}
	}
	return "", "", false
}

func mapEndpoint(tf string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(tf)) {
	case "", "1d", "d", "daily":
		return "histoday", true
	case "1h", "h", "hourly":
		return "histohour", true
	case "1m", "5m", "15m":
		return "histominute", true
	}
	return "", false
}
