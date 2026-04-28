package oecd

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

// Client 封装 OECD Data Explorer API（SDMX-CSV，免 Key）。
// 接口示例：https://sdmx.oecd.org/public/rest/data/<flow>/<key>?format=csvfile
type Client struct {
	src  apiclient.Source
	base string
}

func NewClient(src apiclient.Source) *Client {
	return &Client{src: src, base: "https://sdmx.oecd.org/public/rest/data"}
}

type Point struct {
	Date  time.Time
	Value float64
}

// GetSeries 拉取 OECD 序列；seriesID 形如 "OECD.SDD.NAD,DSD_NAMAIN10@DF_TABLE1,1.0/A.USA.B1GQ.V"。
func (c *Client) GetSeries(ctx context.Context, seriesID string) ([]Point, error) {
	if seriesID == "" {
		return nil, fmt.Errorf("oecd: empty series_id")
	}
	url := fmt.Sprintf("%s/%s?format=csvfile", c.base, seriesID)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	req.Header.Set("Accept", "text/csv")
	resp, err := c.src.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("oecd http %d", resp.StatusCode)
	}
	r := csv.NewReader(resp.Body)
	r.FieldsPerRecord = -1
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) < 2 {
		return nil, nil
	}
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
		return nil, fmt.Errorf("oecd csv missing required columns")
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
		t, err := parseOECDDate(strings.TrimSpace(row[idxTime]))
		if err != nil {
			continue
		}
		out = append(out, Point{Date: t, Value: v})
	}
	return out, nil
}

func parseOECDDate(s string) (time.Time, error) {
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
	return time.Time{}, fmt.Errorf("unrecognized oecd date: %s", s)
}
