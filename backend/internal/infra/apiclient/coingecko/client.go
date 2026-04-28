package coingecko

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/antclaw/antclaw/internal/infra/apiclient"
)

type Client struct {
	src apiclient.Source
	base string
}

func NewClient(src apiclient.Source) *Client {
	return &Client{src: src, base: "https://api.coingecko.com/api/v3"}
}

type MarketResp struct {
	Prices       [][]float64 `json:"prices"`
	MarketCaps   [][]float64 `json:"market_caps"`
	TotalVolumes [][]float64 `json:"total_volumes"`
}

func (c *Client) GetMarketChart(ctx context.Context, coinID, vsCurrency string, days int, interval string) (*MarketResp, error) {
	u, _ := url.Parse(fmt.Sprintf("%s/coins/%s/market_chart", c.base, coinID))
	q := u.Query()
	q.Set("vs_currency", vsCurrency)
	q.Set("days", fmt.Sprintf("%d", days))
	if interval != "" { q.Set("interval", interval) }
	u.RawQuery = q.Encode()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	req.Header.Set("Accept", "application/json")
	resp, err := c.src.Do(ctx, req)
	if err != nil { return nil, err }
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("coingecko http %d", resp.StatusCode)
	}
	var out MarketResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil { return nil, err }
	return &out, nil
}
