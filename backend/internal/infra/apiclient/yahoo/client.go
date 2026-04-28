package yahoo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/antclaw/antclaw/internal/infra/apiclient"
)

type Client struct {
	src  apiclient.Source
	base string
}

func NewClient(src apiclient.Source) *Client {
	return &Client{src: src, base: "https://query1.finance.yahoo.com"}
}

type ChartResp struct {
	Chart struct {
		Result []struct {
			Timestamp  []int64 `json:"timestamp"`
			Indicators struct {
				Quote []struct {
					Open   []float64 `json:"open"`
					High   []float64 `json:"high"`
					Low    []float64 `json:"low"`
					Close  []float64 `json:"close"`
					Volume []float64 `json:"volume"`
				} `json:"quote"`
			} `json:"indicators"`
		} `json:"result"`
	} `json:"chart"`
}

func (c *Client) GetChart(ctx context.Context, symbol, rangeStr, interval string) (*ChartResp, error) {
	url := fmt.Sprintf("%s/v8/finance/chart/%s?range=%s&interval=%s", c.base, symbol, rangeStr, interval)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36")
	resp, err := c.src.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("yahoo http %d", resp.StatusCode)
	}
	var out ChartResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}
