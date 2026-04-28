package firecrawl

import (
	"context"
	"encoding/json"
	"strings"
	"time"
)

// MyFXBookPair 单货币对零售持仓。
type MyFXBookPair struct {
	Symbol   string  `json:"symbol"`
	LongPct  float64 `json:"long_pct"`
	ShortPct float64 `json:"short_pct"`
	Signal   string  `json:"signal"`
}

// MyFXBookSnapshot 全部货币对快照。
type MyFXBookSnapshot struct {
	Pairs     []MyFXBookPair
	FetchedAt time.Time
}

// FetchMyFXBook 抓取 myfxbook.com/community/outlook，返回结构化结果。
// 当 FIRECRAWL_API_KEY 未设置或抓取失败时返回 nil, error，由调用方决定是否降级。
func (c *Client) FetchMyFXBook(ctx context.Context) (*MyFXBookSnapshot, error) {
	schema := json.RawMessage(`{
        "type":"object",
        "properties":{
          "pairs":{"type":"array","items":{
            "type":"object",
            "properties":{
              "symbol":{"type":"string"},
              "long_pct":{"type":"number"},
              "short_pct":{"type":"number"}
            }
          }}
        }
    }`)
	prompt := "Extract retail trader positioning from Myfxbook Community Outlook. " +
		"For each currency pair, return symbol (e.g. EURUSD), long percentage and short percentage."
	var raw struct {
		Pairs []struct {
			Symbol   string  `json:"symbol"`
			LongPct  float64 `json:"long_pct"`
			ShortPct float64 `json:"short_pct"`
		} `json:"pairs"`
	}
	if err := c.ScrapeJSON(ctx, "https://www.myfxbook.com/community/outlook", prompt, schema, 5000, &raw); err != nil {
		return nil, err
	}
	out := &MyFXBookSnapshot{FetchedAt: time.Now().UTC()}
	for _, p := range raw.Pairs {
		sym := strings.ToUpper(strings.TrimSpace(p.Symbol))
		if sym == "" {
			continue
		}
		out.Pairs = append(out.Pairs, MyFXBookPair{
			Symbol: sym, LongPct: p.LongPct, ShortPct: p.ShortPct,
			Signal: ClassifyRetailContrarian(p.LongPct),
		})
	}
	return out, nil
}

// ClassifyRetailContrarian 把零售多头占比映射为反向交易信号。
// ≤20% / ≤30% / ≥80% / ≥70% / 其余分别为 CONTRARIAN_BULLISH / LEAN_BULLISH / CONTRARIAN_BEARISH / LEAN_BEARISH / NEUTRAL。
func ClassifyRetailContrarian(longPct float64) string {
	switch {
	case longPct <= 20:
		return "CONTRARIAN_BULLISH"
	case longPct <= 30:
		return "LEAN_BULLISH"
	case longPct >= 80:
		return "CONTRARIAN_BEARISH"
	case longPct >= 70:
		return "LEAN_BEARISH"
	default:
		return "NEUTRAL"
	}
}
