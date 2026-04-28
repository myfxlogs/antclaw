package cboe

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/antclaw/antclaw/internal/infra/apiclient"
)

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
