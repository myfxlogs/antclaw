package ecb

import (
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/antclaw/antclaw/internal/infra/apiclient"
)

// Client 封装 ECB Statistical Data Warehouse（CSV 端点，免 Key）。
// 接口示例：https://sdw-wsrest.ecb.europa.eu/service/data/<flow>/<key>?format=csvdata
type Client struct {
	src  apiclient.Source
	base string
}

func NewClient(src apiclient.Source) *Client {
	// 新版 Data Portal 端点（旧 sdw-wsrest 已下线 2024-12）
	return &Client{src: src, base: "https://data-api.ecb.europa.eu/service/data"}
}

type Point struct {
	Date  time.Time
	Value float64
}

// GetSeries 拉取 ECB 序列；seriesID 形如 "EXR/D.USD.EUR.SP00.A"（每日欧元/美元）。
func (c *Client) GetSeries(ctx context.Context, seriesID string) ([]Point, error) {
	if seriesID == "" {
		return nil, fmt.Errorf("ecb: empty series_id")
	}
	url := fmt.Sprintf("%s/%s?format=csvdata", c.base, seriesID)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	req.Header.Set("Accept", "text/csv")
	resp, err := c.src.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ecb http %d", resp.StatusCode)
	}
	r := csv.NewReader(resp.Body)
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) < 2 {
		return nil, nil
	}
	// 定位 TIME_PERIOD / OBS_VALUE 列
	header := rows[0]
	idxTime, idxValue := -1, -1
	for i, h := range header {
		switch strings.TrimSpace(h) {
		case "TIME_PERIOD":
			idxTime = i
		case "OBS_VALUE":
			idxValue = i
		}
	}
	if idxTime < 0 || idxValue < 0 {
		return nil, fmt.Errorf("ecb csv missing required columns")
	}
	out := make([]Point, 0, len(rows)-1)
	for _, row := range rows[1:] {
		if len(row) <= idxTime || len(row) <= idxValue {
			continue
		}
		v, err := strconv.ParseFloat(strings.TrimSpace(row[idxValue]), 64)
		if err != nil {
			continue
		}
		t, err := parseECBDate(strings.TrimSpace(row[idxTime]))
		if err != nil {
			continue
		}
		out = append(out, Point{Date: t, Value: v})
	}
	return out, nil
}

// parseECBDate 兼容 ECB 多种粒度："2024", "2024-Q1", "2024-01", "2024-01-15"。
func parseECBDate(s string) (time.Time, error) {
	formats := []string{"2006-01-02", "2006-01", "2006"}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t.UTC(), nil
		}
	}
	// Quarter: 2024-Q1
	if strings.Contains(s, "-Q") {
		parts := strings.Split(s, "-Q")
		if len(parts) == 2 {
			y, err1 := strconv.Atoi(parts[0])
			q, err2 := strconv.Atoi(parts[1])
			if err1 == nil && err2 == nil && q >= 1 && q <= 4 {
				return time.Date(y, time.Month((q-1)*3+1), 1, 0, 0, 0, 0, time.UTC), nil
			}
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized ECB date: %s", s)
}
