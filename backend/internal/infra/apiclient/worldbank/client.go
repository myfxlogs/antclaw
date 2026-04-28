package worldbank

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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

// GetSeries 拉取 country/indicator 时间序列；series_id 形如 "USA/NY.GDP.MKTP.CD"。
func (c *Client) GetSeries(ctx context.Context, seriesID string) ([]Point, error) {
	url := fmt.Sprintf("%s/country/%s?format=json&per_page=1000", c.base, seriesID)
	// World Bank uses /country/<iso>/indicator/<id>
	// seriesID expected like "USA/indicator/NY.GDP.MKTP.CD"
	if seriesID == "" {
		return nil, fmt.Errorf("worldbank: empty series_id")
	}
	url = fmt.Sprintf("%s/country/%s?format=json&per_page=1000", c.base, seriesID)
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
