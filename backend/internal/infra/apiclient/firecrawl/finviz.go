package firecrawl

import (
	"context"
	"encoding/json"
	"strings"
	"time"
)

// FinvizQuote 单 ticker 的 Finviz 快照。
type FinvizQuote struct {
	Ticker        string
	ShortRatio    float64
	ShortPctFloat float64
	InstOwnPct    float64
	FetchedAt     time.Time
}

// FetchFinvizQuote 抓取 finviz.com/quote.ashx?t=<ticker> 的 short 与 institutional 字段。
func (c *Client) FetchFinvizQuote(ctx context.Context, ticker string) (*FinvizQuote, error) {
	t := strings.ToUpper(strings.TrimSpace(ticker))
	if t == "" {
		t = "SPY"
	}
	schema := json.RawMessage(`{
        "type":"object",
        "properties":{
          "short_ratio":{"type":"number"},
          "short_pct_float":{"type":"number"},
          "inst_own_pct":{"type":"number"}
        }
    }`)
	prompt := "Extract short interest ratio, short percentage of float, and institutional ownership percentage from this Finviz quote page. Return numbers only (no % sign)."
	var raw struct {
		ShortRatio    float64 `json:"short_ratio"`
		ShortPctFloat float64 `json:"short_pct_float"`
		InstOwnPct    float64 `json:"inst_own_pct"`
	}
	url := "https://finviz.com/quote.ashx?t=" + t
	if err := c.ScrapeJSON(ctx, url, prompt, schema, 4000, &raw); err != nil {
		return nil, err
	}
	return &FinvizQuote{
		Ticker:        t,
		ShortRatio:    raw.ShortRatio,
		ShortPctFloat: raw.ShortPctFloat,
		InstOwnPct:    raw.InstOwnPct,
		FetchedAt:     time.Now().UTC(),
	}, nil
}
