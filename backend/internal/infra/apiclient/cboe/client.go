package cboe

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/antclaw/antclaw/internal/infra/apiclient"
)

// jsonDecoder 把字节数组包装成可重用的 decode 闭包，便于拆分错误处理。
func jsonDecoder(b []byte) func(any) error {
	return json.NewDecoder(strings.NewReader(string(b))).Decode
}

// Client 封装 CBOE 公开 CSV：daily index put/call，free。
type Client struct {
	src  apiclient.Source
	base string
}

func NewClient(src apiclient.Source) *Client {
	return &Client{src: src, base: "https://cdn.cboe.com/api/global/us_indices/daily_market_statistics"}
}

type PutCallDaily struct {
	Date     time.Time
	TotalPC  float64
	EquityPC float64
	IndexPC  float64
}

// VIXSnapshot CBOE 公开历史曲线接口的轻量快照。
type VIXSnapshot struct {
	Date  time.Time
	Close float64
}

// FetchVIXHistory 拉取 CBOE 公开历史 VIX，返回升序日级序列。
// 端点：https://cdn.cboe.com/api/global/delayed_quotes/charts/historical/^VIX.json
func (c *Client) FetchVIXHistory(ctx context.Context) ([]VIXSnapshot, error) {
	const url = "https://cdn.cboe.com/api/global/delayed_quotes/charts/historical/_VIX.json"
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 AntClaw/1.0")
	resp, err := c.src.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cboe vix http %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// 端点返回 { "data": [{"date":"YYYY-MM-DD","close":"17.24",...}, ...] }
	// 注意 close 是字符串。
	type point struct {
		Date  string `json:"date"`
		Close string `json:"close"`
	}
	var raw struct {
		Data []point `json:"data"`
	}
	dec := jsonDecoder(body)
	if err := dec(&raw); err != nil {
		return nil, fmt.Errorf("cboe vix decode: %w", err)
	}
	out := make([]VIXSnapshot, 0, len(raw.Data))
	for _, p := range raw.Data {
		t, err := time.Parse("2006-01-02", p.Date)
		if err != nil {
			continue
		}
		c, err := strconv.ParseFloat(p.Close, 64)
		if err != nil || c <= 0 {
			continue
		}
		out = append(out, VIXSnapshot{Date: t, Close: c})
	}
	return out, nil
}

// FetchLatest 返回最近一日 put/call 数据。
// 数据源使用 https://www.cboe.com/us/options/market_statistics/daily/?dt=YYYY-MM-DD 历史。
// 简化：以 cboe.com 公开汇总 CSV 端点，若不可用则返回 ErrUnavailable。
func (c *Client) FetchLatest(ctx context.Context) (*PutCallDaily, error) {
	url := "https://www.cboe.com/us/options/market_statistics/historical_data/csv/all/"
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := c.src.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cboe http %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	r := csv.NewReader(strings.NewReader(string(body)))
	rows, err := r.ReadAll()
	if err != nil || len(rows) < 2 {
		return nil, fmt.Errorf("cboe csv parse failed")
	}
	last := rows[len(rows)-1]
	if len(last) < 4 {
		return nil, fmt.Errorf("cboe row too short")
	}
	date, err := time.Parse("2006-01-02", strings.TrimSpace(last[0]))
	if err != nil {
		return nil, err
	}
	tot, _ := strconv.ParseFloat(last[1], 64)
	eq, _ := strconv.ParseFloat(last[2], 64)
	idx, _ := strconv.ParseFloat(last[3], 64)
	return &PutCallDaily{Date: date, TotalPC: tot, EquityPC: eq, IndexPC: idx}, nil
}
