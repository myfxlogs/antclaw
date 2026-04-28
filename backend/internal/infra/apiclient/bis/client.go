package bis

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

// Client 封装 BIS Statistics API（SDMX 2.1，CSV 输出，免 Key）。
// 接口示例：https://stats.bis.org/api/v2/data/<flow>/<key>?format=csv
type Client struct {
	src  apiclient.Source
	base string
}

func NewClient(src apiclient.Source) *Client {
	return &Client{src: src, base: "https://stats.bis.org/api/v2/data"}
}

type Point struct {
	Date  time.Time
	Value float64
}

// GetSeries 拉取 BIS 序列；seriesID 形如 "BIS,WS_EER_D,1.0/D.N.B.US"（每日美元名义有效汇率）。
func (c *Client) GetSeries(ctx context.Context, seriesID string) ([]Point, error) {
	if seriesID == "" {
		return nil, fmt.Errorf("bis: empty series_id")
	}
	url := fmt.Sprintf("%s/%s?format=csv", c.base, seriesID)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	req.Header.Set("Accept", "text/csv")
	resp, err := c.src.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bis http %d", resp.StatusCode)
	}
	r := csv.NewReader(resp.Body)
	r.FieldsPerRecord = -1 // BIS CSV 列数有时不一致
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) < 2 {
		return nil, nil
	}
	// 与 ECB 同为 SDMX-CSV：TIME_PERIOD / OBS_VALUE 列
	idxTime, idxValue := -1, -1
	for i, h := range rows[0] {
		switch strings.TrimSpace(h) {
		case "TIME_PERIOD":
			idxTime = i
		case "OBS_VALUE":
			idxValue = i
		}
	}
	if idxTime < 0 || idxValue < 0 {
		return nil, fmt.Errorf("bis csv missing required columns")
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
		t, err := parseBISDate(strings.TrimSpace(row[idxTime]))
		if err != nil {
			continue
		}
		out = append(out, Point{Date: t, Value: v})
	}
	return out, nil
}

// parseBISDate 兼容 BIS 多种粒度（年/季/月/日）。
func parseBISDate(s string) (time.Time, error) {
	for _, f := range []string{"2006-01-02", "2006-01", "2006"} {
		if t, err := time.Parse(f, s); err == nil {
			return t.UTC(), nil
		}
	}
	if strings.Contains(s, "-Q") {
		parts := strings.Split(s, "-Q")
		if len(parts) == 2 {
			y, e1 := strconv.Atoi(parts[0])
			q, e2 := strconv.Atoi(parts[1])
			if e1 == nil && e2 == nil && q >= 1 && q <= 4 {
				return time.Date(y, time.Month((q-1)*3+1), 1, 0, 0, 0, 0, time.UTC), nil
			}
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized BIS date: %s", s)
}
