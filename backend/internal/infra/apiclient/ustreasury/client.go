package ustreasury

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/antclaw/antclaw/internal/infra/apiclient"
)

// Client 封装 home.treasury.gov 每日收益率曲线 CSV feed。
// 接口形如：/resource-center/data-chart-center/interest-rates/daily-treasury-rates.csv/{year}/all?type=daily_treasury_yield_curve
// （XML/Atom 版已下线，仅保留 CSV）。
type Client struct {
	src  apiclient.Source
	base string
}

func NewClient(src apiclient.Source) *Client {
	return &Client{src: src, base: "https://home.treasury.gov/resource-center/data-chart-center/interest-rates/daily-treasury-rates.csv"}
}

// CurveRow 单日收益率曲线（年限对应字段）。
type CurveRow struct {
	Date time.Time
	Y1M  float64
	Y2M  float64
	Y3M  float64
	Y6M  float64
	Y1Y  float64
	Y2Y  float64
	Y3Y  float64
	Y5Y  float64
	Y7Y  float64
	Y10Y float64
	Y20Y float64
	Y30Y float64
}

// FetchYearXML 历史方法名（保留兼容），实际拉取 CSV。
// 返回按日期升序排列的曲线序列；最近一日在末尾。
func (c *Client) FetchYearXML(ctx context.Context, year int) ([]CurveRow, error) {
	return c.FetchYear(ctx, year)
}

// FetchYear 拉取指定年份的全部交易日收益率曲线。
func (c *Client) FetchYear(ctx context.Context, year int) ([]CurveRow, error) {
	url := fmt.Sprintf("%s/%d/all?type=daily_treasury_yield_curve", c.base, year)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := c.src.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ustreasury http %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return parseCurveCSV(body)
}

// parseCurveCSV 解析 Treasury CSV，按日期升序返回。
// 表头形如：Date,"1 Mo","1.5 Month","2 Mo","3 Mo","4 Mo","6 Mo","1 Yr","2 Yr","3 Yr","5 Yr","7 Yr","10 Yr","20 Yr","30 Yr"
func parseCurveCSV(b []byte) ([]CurveRow, error) {
	r := csv.NewReader(strings.NewReader(string(b)))
	r.FieldsPerRecord = -1
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) < 2 {
		return nil, fmt.Errorf("ustreasury csv: empty")
	}
	header := records[0]
	idx := map[string]int{}
	for i, h := range header {
		idx[strings.TrimSpace(h)] = i
	}
	col := func(name string) int {
		if v, ok := idx[name]; ok {
			return v
		}
		return -1
	}
	dateCol := col("Date")
	if dateCol < 0 {
		return nil, fmt.Errorf("ustreasury csv: missing Date column")
	}
	type binding struct {
		idx int
		set func(*CurveRow, float64)
	}
	bindings := []binding{
		{col("1 Mo"), func(r *CurveRow, v float64) { r.Y1M = v }},
		{col("2 Mo"), func(r *CurveRow, v float64) { r.Y2M = v }},
		{col("3 Mo"), func(r *CurveRow, v float64) { r.Y3M = v }},
		{col("6 Mo"), func(r *CurveRow, v float64) { r.Y6M = v }},
		{col("1 Yr"), func(r *CurveRow, v float64) { r.Y1Y = v }},
		{col("2 Yr"), func(r *CurveRow, v float64) { r.Y2Y = v }},
		{col("3 Yr"), func(r *CurveRow, v float64) { r.Y3Y = v }},
		{col("5 Yr"), func(r *CurveRow, v float64) { r.Y5Y = v }},
		{col("7 Yr"), func(r *CurveRow, v float64) { r.Y7Y = v }},
		{col("10 Yr"), func(r *CurveRow, v float64) { r.Y10Y = v }},
		{col("20 Yr"), func(r *CurveRow, v float64) { r.Y20Y = v }},
		{col("30 Yr"), func(r *CurveRow, v float64) { r.Y30Y = v }},
	}
	rows := make([]CurveRow, 0, len(records)-1)
	for _, rec := range records[1:] {
		if dateCol >= len(rec) {
			continue
		}
		dt, err := time.Parse("01/02/2006", strings.TrimSpace(rec[dateCol]))
		if err != nil {
			continue
		}
		row := CurveRow{Date: dt}
		for _, b := range bindings {
			if b.idx < 0 || b.idx >= len(rec) {
				continue
			}
			s := strings.TrimSpace(rec[b.idx])
			if s == "" {
				continue
			}
			v, err := strconv.ParseFloat(s, 64)
			if err != nil {
				continue
			}
			b.set(&row, v)
		}
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Date.Before(rows[j].Date) })
	return rows, nil
}
