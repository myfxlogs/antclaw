package defillama

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/antclaw/antclaw/internal/infra/apiclient"
)

// Client 封装 DefiLlama 公开 API（无需 Key）。
type Client struct {
	src  apiclient.Source
	base string
}

func NewClient(src apiclient.Source) *Client {
	return &Client{src: src, base: "https://api.llama.fi"}
}

type Protocol struct {
	Slug     string  `json:"slug"`
	Name     string  `json:"name"`
	Category string  `json:"category"`
	TVL      float64 `json:"tvl"`
	Change1d float64 `json:"change_1d"`
	Change7d float64 `json:"change_7d"`
	Chain    string  `json:"chain"`
}

func (c *Client) ListProtocols(ctx context.Context) ([]Protocol, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, c.base+"/protocols", nil)
	resp, err := c.src.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("defillama http %d", resp.StatusCode)
	}
	var out []Protocol
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

type TVLPoint struct {
	Date int64   `json:"date"`
	TVL  float64 `json:"totalLiquidityUSD"`
}

type ProtocolDetail struct {
	TVL []TVLPoint `json:"tvl"`
}

func (c *Client) GetProtocolTVL(ctx context.Context, slug string) (*ProtocolDetail, error) {
	url := fmt.Sprintf("%s/protocol/%s", c.base, slug)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := c.src.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("defillama http %d", resp.StatusCode)
	}
	var out ProtocolDetail
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}
