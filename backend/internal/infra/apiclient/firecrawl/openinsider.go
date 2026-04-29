package firecrawl

import (
	"context"
	"encoding/json"
	"strings"
	"time"
)

// OpenInsiderTrade 单条内幕交易记录。
type OpenInsiderTrade struct {
	Ticker   string
	Insider  string
	Title    string
	Action   string // P=Purchase, S=Sale
	Date     time.Time
	Price    float64
	Quantity int64
}

// FetchOpenInsiderTrades 抓 openinsider.com/<ticker> 的最近 N 条内幕交易表，结构化输出。
func (c *Client) FetchOpenInsiderTrades(ctx context.Context, ticker string, limit int) ([]OpenInsiderTrade, error) {
	t := strings.ToUpper(strings.TrimSpace(ticker))
	if t == "" {
		t = "AAPL"
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	schema := json.RawMessage(`{
        "type":"object",
        "properties":{
          "trades":{
            "type":"array",
            "items":{
              "type":"object",
              "properties":{
                "insider":{"type":"string"},
                "title":{"type":"string"},
                "action":{"type":"string"},
                "date":{"type":"string"},
                "price":{"type":"number"},
                "quantity":{"type":"number"}
              }
            }
          }
        }
    }`)
	prompt := "Extract the recent insider trade rows from this OpenInsider page. For each row return insider name, title, action (P/S/A), trade date (YYYY-MM-DD), price (number), quantity (number). Return at most " + atoi(limit) + " rows."
	var raw struct {
		Trades []struct {
			Insider, Title, Action, Date string
			Price                        float64
			Quantity                     float64
		} `json:"trades"`
	}
	url := "http://openinsider.com/screener?s=" + t
	if err := c.ScrapeJSON(ctx, url, prompt, schema, 8000, &raw); err != nil {
		return nil, err
	}
	out := make([]OpenInsiderTrade, 0, len(raw.Trades))
	for _, r := range raw.Trades {
		date, _ := time.Parse("2006-01-02", strings.TrimSpace(r.Date))
		out = append(out, OpenInsiderTrade{
			Ticker:   t,
			Insider:  r.Insider,
			Title:    r.Title,
			Action:   r.Action,
			Date:     date,
			Price:    r.Price,
			Quantity: int64(r.Quantity),
		})
	}
	return out, nil
}

// atoi 把 int 转 string；避免 import strconv 仅为此一处。
func atoi(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	buf := make([]byte, 0, 8)
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	if neg {
		buf = append([]byte{'-'}, buf...)
	}
	return string(buf)
}
