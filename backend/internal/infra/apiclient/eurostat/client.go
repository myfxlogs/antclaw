package eurostat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/antclaw/antclaw/internal/infra/apiclient"
)

// Client 封装 Eurostat Statistics API（SDMX-JSON v1，免 Key）。
// 接口示例：https://ec.europa.eu/eurostat/api/dissemination/statistics/1.0/data/<dataset>?format=JSON
type Client struct {
	src  apiclient.Source
	base string
}

func NewClient(src apiclient.Source) *Client {
	return &Client{src: src, base: "https://ec.europa.eu/eurostat/api/dissemination/statistics/1.0/data"}
}

type Point struct {
	Date  time.Time
	Value float64
}

// GetSeries 拉取 Eurostat 数据集；
// seriesID 形如 "tps00001?geo=EU27_2020"（带 query 参数选择维度）。
func (c *Client) GetSeries(ctx context.Context, seriesID string) ([]Point, error) {
	if seriesID == "" {
		return nil, fmt.Errorf("eurostat: empty series_id")
	}
	sep := "?"
	if containsQuery(seriesID) {
		sep = "&"
	}
	url := fmt.Sprintf("%s/%s%sformat=JSON&lang=EN", c.base, seriesID, sep)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := c.src.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("eurostat http %d", resp.StatusCode)
	}
	// SDMX-JSON: dimension 含 time category index → label，value 为 {flat_index: number}
	var raw struct {
		Value     map[string]float64 `json:"value"`
		Dimension struct {
			Time struct {
				Category struct {
					Index map[string]int `json:"index"`
				} `json:"category"`
			} `json:"time"`
		} `json:"dimension"`
		Size []int    `json:"size"`
		ID   []string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	timeIdx := raw.Dimension.Time.Category.Index
	if len(timeIdx) == 0 {
		return nil, nil
	}
	// 反转 index：flat_position → time_label
	posToTime := make(map[int]string, len(timeIdx))
	for label, pos := range timeIdx {
		posToTime[pos] = label
	}
	// 找出 time 维度在 ID 中的位置
	timeAxis := -1
	for i, name := range raw.ID {
		if name == "time" {
			timeAxis = i
			break
		}
	}
	if timeAxis < 0 || len(raw.Size) == 0 {
		return nil, fmt.Errorf("eurostat: time dimension not found")
	}
	// 计算每个 time index 在 flat 索引中的 stride
	stride := 1
	for i := timeAxis + 1; i < len(raw.Size); i++ {
		stride *= raw.Size[i]
	}
	timeSize := raw.Size[timeAxis]
	out := make([]Point, 0, timeSize)
	// 假设其他维度均已固定到唯一值（API 调用方应通过 query 过滤）
	for tPos := 0; tPos < timeSize; tPos++ {
		flat := tPos * stride
		v, ok := raw.Value[strconv.Itoa(flat)]
		if !ok {
			continue
		}
		label, ok := posToTime[tPos]
		if !ok {
			continue
		}
		t, err := parseEurostatDate(label)
		if err != nil {
			continue
		}
		out = append(out, Point{Date: t, Value: v})
	}
	return out, nil
}

func containsQuery(s string) bool {
	for _, ch := range s {
		if ch == '?' {
			return true
		}
	}
	return false
}

func parseEurostatDate(s string) (time.Time, error) {
	for _, f := range []string{"2006-01-02", "2006-01", "2006"} {
		if t, err := time.Parse(f, s); err == nil {
			return t.UTC(), nil
		}
	}
	// Quarter "2024Q1"
	if len(s) == 6 && s[4] == 'Q' {
		y, e1 := strconv.Atoi(s[:4])
		q, e2 := strconv.Atoi(s[5:])
		if e1 == nil && e2 == nil && q >= 1 && q <= 4 {
			return time.Date(y, time.Month((q-1)*3+1), 1, 0, 0, 0, 0, time.UTC), nil
		}
	}
	// Month "2024M03"
	if len(s) == 7 && s[4] == 'M' {
		y, e1 := strconv.Atoi(s[:4])
		m, e2 := strconv.Atoi(s[5:])
		if e1 == nil && e2 == nil && m >= 1 && m <= 12 {
			return time.Date(y, time.Month(m), 1, 0, 0, 0, 0, time.UTC), nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized eurostat date: %s", s)
}
