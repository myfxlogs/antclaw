package firecrawl

import (
	"context"
	"encoding/json"
	"time"
)

// MoveSnapshot ICE BofA MOVE 指数最新值。
type MoveSnapshot struct {
	Date  time.Time
	Value float64
	Trend string // rising/falling/stable
}

// FetchMove 通过 firecrawl 抓 yardeni.com/Bond Volatility 报告页中的 MOVE 当前值与近期趋势。
// 公开 PDF/HTML 报告地址：https://www.yardeni.com/pub/bond_volatility_report.pdf （或对应 HTML 页）
// firecrawl 会自动解析 PDF 并提取数值。
func (c *Client) FetchMove(ctx context.Context) (*MoveSnapshot, error) {
	schema := json.RawMessage(`{
        "type":"object",
        "properties":{
            "date":{"type":"string"},
            "value":{"type":"number"},
            "trend":{"type":"string"}
        }
    }`)
	prompt := "Extract the latest MOVE Index (ICE BofAML US Bond Market Option Volatility Estimate Index): the most recent value (number), date (YYYY-MM-DD), and short trend label (rising / falling / stable)."
	var raw struct {
		Date  string  `json:"date"`
		Value float64 `json:"value"`
		Trend string  `json:"trend"`
	}
	// yardeni 公开页面（不需要登录）
	url := "https://www.yardeni.com/pub/sp500move.pdf"
	if err := c.ScrapeJSON(ctx, url, prompt, schema, 6000, &raw); err != nil {
		return nil, err
	}
	t, _ := time.Parse("2006-01-02", raw.Date)
	if t.IsZero() {
		t = time.Now().UTC()
	}
	return &MoveSnapshot{Date: t, Value: raw.Value, Trend: raw.Trend}, nil
}
