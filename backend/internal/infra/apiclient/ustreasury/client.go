package ustreasury

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/antclaw/antclaw/internal/infra/apiclient"
)

// Client 封装 home.treasury.gov daily yield curve XML feed。
type Client struct {
	src  apiclient.Source
	base string
}

func NewClient(src apiclient.Source) *Client {
	return &Client{src: src, base: "https://home.treasury.gov/resource-center/data-chart-center/interest-rates/daily-treasury-rates.csv/all/"}
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

// FetchYearXML 拉取指定年份的 XML 数据并解析为按日期排序的曲线序列。
func (c *Client) FetchYearXML(ctx context.Context, year int) ([]CurveRow, error) {
	url := fmt.Sprintf("%s%d?type=daily_treasury_yield_curve&_format=xml", c.base, year)
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
	return parseAtomXML(body)
}

type atomFeed struct {
	XMLName xml.Name    `xml:"feed"`
	Entries []atomEntry `xml:"entry"`
}
type atomEntry struct {
	Content struct {
		Properties struct {
			NewDate struct {
				Value string `xml:",chardata"`
			} `xml:"NEW_DATE"`
			Fields []rawField `xml:",any"`
		} `xml:"properties"`
	} `xml:"content"`
}
type rawField struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}

func parseAtomXML(b []byte) ([]CurveRow, error) {
	var f atomFeed
	if err := xml.Unmarshal(b, &f); err != nil {
		return nil, err
	}
	rows := make([]CurveRow, 0, len(f.Entries))
	for _, e := range f.Entries {
		dt, err := time.Parse("2006-01-02T15:04:05", e.Content.Properties.NewDate.Value)
		if err != nil {
			continue
		}
		row := CurveRow{Date: dt}
		for _, fld := range e.Content.Properties.Fields {
			v, err := strconv.ParseFloat(fld.Value, 64)
			if err != nil {
				continue
			}
			switch fld.XMLName.Local {
			case "BC_1MONTH":
				row.Y1M = v
			case "BC_2MONTH":
				row.Y2M = v
			case "BC_3MONTH":
				row.Y3M = v
			case "BC_6MONTH":
				row.Y6M = v
			case "BC_1YEAR":
				row.Y1Y = v
			case "BC_2YEAR":
				row.Y2Y = v
			case "BC_3YEAR":
				row.Y3Y = v
			case "BC_5YEAR":
				row.Y5Y = v
			case "BC_7YEAR":
				row.Y7Y = v
			case "BC_10YEAR":
				row.Y10Y = v
			case "BC_20YEAR":
				row.Y20Y = v
			case "BC_30YEAR":
				row.Y30Y = v
			}
		}
		rows = append(rows, row)
	}
	return rows, nil
}
