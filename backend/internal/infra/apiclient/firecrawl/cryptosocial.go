package firecrawl

import (
	"context"
	"encoding/json"
	"strings"
	"time"
)

// CryptoSocialSnapshot 单个加密资产的社媒指标。
type CryptoSocialSnapshot struct {
	Asset                  string
	Date                   time.Time
	TwitterFollowersGrowth float64
	RedditSubscribersGrowth float64
	SentimentScore         float64 // -1..1
}

// FetchCryptoSocial 通过 firecrawl 抓 LunarCrush 公开的资产页（无登录），抽取社交增长与情绪分。
// 端点：https://lunarcrush.com/coins/<symbol>
func (c *Client) FetchCryptoSocial(ctx context.Context, asset string) (*CryptoSocialSnapshot, error) {
	a := strings.ToLower(strings.TrimSpace(asset))
	if a == "" {
		a = "btc"
	}
	schema := json.RawMessage(`{
        "type":"object",
        "properties":{
          "twitter_followers_growth":{"type":"number"},
          "reddit_subscribers_growth":{"type":"number"},
          "sentiment_score":{"type":"number"}
        }
    }`)
	prompt := "From this LunarCrush coin page, extract: twitter_followers_growth (% change last 24h, signed number), reddit_subscribers_growth (% 24h, signed), sentiment_score (-1..1, current galaxy/sentiment metric mapped to -1..1)."
	var raw struct {
		TwitterFollowersGrowth  float64 `json:"twitter_followers_growth"`
		RedditSubscribersGrowth float64 `json:"reddit_subscribers_growth"`
		SentimentScore          float64 `json:"sentiment_score"`
	}
	url := "https://lunarcrush.com/coins/" + a
	if err := c.ScrapeJSON(ctx, url, prompt, schema, 6000, &raw); err != nil {
		return nil, err
	}
	return &CryptoSocialSnapshot{
		Asset:                   strings.ToUpper(a),
		Date:                    time.Now().UTC(),
		TwitterFollowersGrowth:  raw.TwitterFollowersGrowth,
		RedditSubscribersGrowth: raw.RedditSubscribersGrowth,
		SentimentScore:          raw.SentimentScore,
	}, nil
}
