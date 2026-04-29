package worldbank

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/antclaw/antclaw/internal/infra/apiclient"
)

// Client 封装 World Bank 公开 API（无需 Key）。
// 形如：https://api.worldbank.org/v2/country/USA/indicator/NY.GDP.MKTP.CD?format=json&per_page=1000
type Client struct {
	src  apiclient.Source
	base string
}

func NewClient(src apiclient.Source) *Client {
	return &Client{src: src, base: "https://api.worldbank.org/v2"}
}

type Point struct {
	Date  time.Time
	Value float64
}

// GetSeries 拉取 country/indicator 时间序列。series_id 接受三种写法：
//   1) "NY.GDP.MKTP.CD"            —— 只填 indicator，默认 country=WLD（世界总量）
//   2) "USA/NY.GDP.MKTP.CD"        —— ISO + indicator
//   3) "USA/indicator/NY.GDP.MKTP.CD" —— 完整 v2 子路径
func (c *Client) GetSeries(ctx context.Context, seriesID string) ([]Point, error) {
	seriesID = strings.TrimSpace(seriesID)
	if seriesID == "" {
		return nil, fmt.Errorf("worldbank: empty series_id")
	}
	parts := strings.SplitN(seriesID, "/", 3)
	var path string
	switch len(parts) {
	case 1:
		// 只给了 indicator，自动落到 World 聚合，避免要求用户记忆 ISO 代码。
		path = fmt.Sprintf("country/WLD/indicator/%s", parts[0])
	case 2:
		path = fmt.Sprintf("country/%s/indicator/%s", parts[0], parts[1])
	case 3:
		path = fmt.Sprintf("country/%s", seriesID)
	default:
		return nil, fmt.Errorf("worldbank: invalid series_id %q", seriesID)
	}
	url := fmt.Sprintf("%s/%s?format=json&per_page=1000", c.base, path)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := c.src.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("worldbank http %d", resp.StatusCode)
	}
	// Response: [meta, [{date,value,...}, ...]]
	var raw []json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	if len(raw) < 2 {
		return nil, nil
	}
	type row struct {
		Date  string   `json:"date"`
		Value *float64 `json:"value"`
	}
	var items []row
	if err := json.Unmarshal(raw[1], &items); err != nil {
		return nil, err
	}
	out := make([]Point, 0, len(items))
	for _, it := range items {
		if it.Value == nil {
			continue
		}
		t, err := time.Parse("2006", it.Date)
		if err != nil {
			t, err = time.Parse("2006-01", it.Date)
			if err != nil {
				continue
			}
		}
		out = append(out, Point{Date: t, Value: *it.Value})
	}
	return out, nil
}
