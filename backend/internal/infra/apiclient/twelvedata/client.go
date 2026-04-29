// Package twelvedata 提供 TwelveData REST API 的极简客户端。
// 文档：https://twelvedata.com/docs
//
// 仅暴露 FetchOHLC，签名与其他 vendor 子包对齐，便于 marketdata.Aggregator 串联。
package twelvedata

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
	return &Client{src: src, apiKey: strings.TrimSpace(apiKey), base: "https://api.twelvedata.com"}
}

func (c *Client) Available() bool { return c.apiKey != "" }
func (c *Client) Name() string    { return "twelvedata" }

// Bar 通用 OHLC 结构。
type Bar struct {
	Time   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume int64
}

type tsResp struct {
	Values []struct {
		Datetime string `json:"datetime"`
		Open     string `json:"open"`
		High     string `json:"high"`
		Low      string `json:"low"`
		Close    string `json:"close"`
		Volume   string `json:"volume"`
	} `json:"values"`
	Status string `json:"status"`
	Code   int    `json:"code"`
	Msg    string `json:"message"`
}

// FetchOHLC 拉取 symbol 的最近 outputSize 根 K 线。
// timeframe: "1day"/"1h"/"4h"/"15min"/"1min"。
// 注意：FX symbol 形如 "EUR/USD"；输入 "EURUSD" 时自动转换。
func (c *Client) FetchOHLC(ctx context.Context, symbol, timeframe string, outputSize int) ([]Bar, error) {
	if !c.Available() {
		return nil, fmt.Errorf("twelvedata: api key not configured")
	}
	if outputSize <= 0 || outputSize > 5000 {
		outputSize = 500
	}
	tf, ok := mapTimeframe(timeframe)
	if !ok {
		return nil, fmt.Errorf("twelvedata: unsupported timeframe %q", timeframe)
	}
	q := url.Values{}
	q.Set("symbol", normalizeSymbol(symbol))
	q.Set("interval", tf)
	q.Set("outputsize", strconv.Itoa(outputSize))
	q.Set("apikey", c.apiKey)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, c.base+"/time_series?"+q.Encode(), nil)
	resp, err := c.src.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var ts tsResp
	if err := json.NewDecoder(resp.Body).Decode(&ts); err != nil {
		return nil, err
	}
	if ts.Status == "error" {
		return nil, fmt.Errorf("twelvedata: %s", ts.Msg)
	}
	out := make([]Bar, 0, len(ts.Values))
	for i := len(ts.Values) - 1; i >= 0; i-- { // 反转为升序
		v := ts.Values[i]
		t, err := time.Parse("2006-01-02 15:04:05", v.Datetime)
		if err != nil {
			t, err = time.Parse("2006-01-02", v.Datetime)
			if err != nil {
				continue
			}
		}
		o, _ := strconv.ParseFloat(v.Open, 64)
		h, _ := strconv.ParseFloat(v.High, 64)
		l, _ := strconv.ParseFloat(v.Low, 64)
		cl, _ := strconv.ParseFloat(v.Close, 64)
		vol, _ := strconv.ParseFloat(v.Volume, 64)
		if cl <= 0 {
			continue
		}
		out = append(out, Bar{Time: t.UTC(), Open: o, High: h, Low: l, Close: cl, Volume: int64(vol)})
	}
	return out, nil
}

// normalizeSymbol "EURUSD" -> "EUR/USD"；含 "/" 时原样返回。
func normalizeSymbol(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	if strings.Contains(s, "/") {
		return s
	}
	if len(s) == 6 {
		return s[:3] + "/" + s[3:]
	}
	return s
}

func mapTimeframe(tf string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(tf)) {
	case "", "1d", "d", "daily":
		return "1day", true
	case "1h", "h":
		return "1h", true
	case "4h":
		return "4h", true
	case "15m":
		return "15min", true
	case "5m":
		return "5min", true
	case "1m":
		return "1min", true
	}
	return "", false
}
