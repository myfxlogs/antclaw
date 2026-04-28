package imf

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/antclaw/antclaw/internal/infra/apiclient"
)

// Client 封装 IMF Data API（无需 Key）。
// 接口示例：https://www.imf.org/external/datamapper/api/v1/<indicator>/<country>?periods=2010-2024
type Client struct {
	src  apiclient.Source
	base string
}

func NewClient(src apiclient.Source) *Client {
	return &Client{src: src, base: "https://www.imf.org/external/datamapper/api/v1"}
}

type Point struct {
	Date  time.Time
	Value float64
}

// GetSeries 拉取 indicator+country 时间序列；
// seriesID 形如 "NGDP_RPCH/USA"（实际 GDP 增速）。
func (c *Client) GetSeries(ctx context.Context, seriesID string) ([]Point, error) {
	if seriesID == "" {
		return nil, fmt.Errorf("imf: empty series_id")
	}
	url := fmt.Sprintf("%s/%s", c.base, seriesID)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := c.src.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("imf http %d", resp.StatusCode)
	}
	// Response 形如 {"values": {"<indicator>": {"<country>": {"2010":..., "2011":...}}}}
	var raw struct {
		Values map[string]map[string]map[string]float64 `json:"values"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	var out []Point
	for _, indMap := range raw.Values {
		for _, yearMap := range indMap {
			for y, v := range yearMap {
				yi, err := strconv.Atoi(y)
				if err != nil {
					continue
				}
				out = append(out, Point{
					Date:  time.Date(yi, 1, 1, 0, 0, 0, 0, time.UTC),
					Value: v,
				})
			}
		}
	}
	return out, nil
}
