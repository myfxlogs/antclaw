package snb

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/antclaw/antclaw/internal/infra/apiclient"
)

// Client 封装 SNB（瑞士国家银行）公开数据 API（CSV，免 Key）。
// 接口示例：https://data.snb.ch/api/cube/devkua/data/csv/en（央行政策利率）
type Client struct {
	src  apiclient.Source
	base string
}

func NewClient(src apiclient.Source) *Client {
	return &Client{src: src, base: "https://data.snb.ch/api/cube"}
}

type Point struct {
	Date  time.Time
	Value float64
}

// GetSeries 拉取 SNB 立方体；seriesID 形如 "devkua"（cube name）。
// SNB CSV 头部含元数据，"Date,Value" 数据从空行后开始。
func (c *Client) GetSeries(ctx context.Context, seriesID string) ([]Point, error) {
	if seriesID == "" {
		return nil, fmt.Errorf("snb: empty series_id")
	}
	url := fmt.Sprintf("%s/%s/data/csv/en", c.base, seriesID)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	req.Header.Set("Accept", "text/csv")
	resp, err := c.src.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("snb http %d", resp.StatusCode)
	}
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	out := make([]Point, 0, 256)
	inData := false
	headerSeen := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !inData {
			if line == "" {
				inData = true
			}
			continue
		}
		if !headerSeen {
			headerSeen = true // skip header row "Date,..."
			continue
		}
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) < 2 {
			continue
		}
		dateStr := strings.TrimSpace(parts[0])
		valStr := strings.TrimSpace(parts[len(parts)-1])
		t, err := parseSNBDate(dateStr)
		if err != nil {
			continue
		}
		v, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			continue
		}
		out = append(out, Point{Date: t, Value: v})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func parseSNBDate(s string) (time.Time, error) {
	for _, f := range []string{"2006-01-02", "2006-01", "2006"} {
		if t, err := time.Parse(f, s); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized snb date: %s", s)
}
