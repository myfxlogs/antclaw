// Package cftc 封装 CFTC Commitment of Traders 报表 API（薄客户端）。
//
// 入口：使用 publicreporting.cftc.gov Socrata 数据集（免 Key 也可读，配 Key 提速率）。
// 业务派生（持仓 z-score、置信度等）属 service/cot 域，不在此包。
package cftc

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/antclaw/antclaw/internal/infra/apiclient"
)

// Socrata legacy futures-and-options 数据集 URL（免 Key 可读）。
const legacyDatasetURL = "https://publicreporting.cftc.gov/resource/6dca-aqww.json"

// Client 通过 apiclient.Source 中间件调用 CFTC Socrata 接口。
type Client struct {
	src      apiclient.Source
	endpoint string
	apiKey   string
}

// NewClient 使用 Source 构造客户端；apiKey 可空（公共数据，配 Key 提速率）。
func NewClient(src apiclient.Source, apiKey string) *Client {
	return &Client{src: src, endpoint: legacyDatasetURL, apiKey: apiKey}
}

// COTReport 单条 legacy 报表记录（仅暴露上层关心字段）。
type COTReport struct {
	ReportDateAsOf           string `json:"report_date_as_of"`
	CftcContractMarketCode   string `json:"cftc_contract_market_code"`
	CftcMarketCode           string `json:"cftc_market_code"`
	CftcRegionCode           string `json:"cftc_region_code"`
	CftcCommodityCode        string `json:"cftc_commodity_code"`
	OpenInterestAll          int64  `json:"open_interest_all"`
	NoncommPositionsLongAll  int64  `json:"noncomm_positions_long_all"`
	NoncommPositionsShortAll int64  `json:"noncomm_positions_short_all"`
	CommPositionsLongAll     int64  `json:"comm_positions_long_all"`
	CommPositionsShortAll    int64  `json:"comm_positions_short_all"`
	NonreptPositionsLongAll  int64  `json:"nonrept_positions_long_all"`
	NonreptPositionsShortAll int64  `json:"nonrept_positions_short_all"`
}

// rawRecord Socrata 返回的字符串字段，需要二次解析数值。
type rawRecord map[string]string

// FetchLegacyReports 拉取指定合约代码、日期区间内的 legacy COT 报表。
func (c *Client) FetchLegacyReports(ctx context.Context, marketCode string, from, to time.Time) ([]COTReport, error) {
	q := url.Values{}
	q.Set("$where", fmt.Sprintf(
		"cftc_contract_market_code='%s' AND report_date_as_of between '%s' AND '%s'",
		marketCode,
		from.UTC().Format("2006-01-02T15:04:05"),
		to.UTC().Format("2006-01-02T15:04:05"),
	))
	q.Set("$order", "report_date_as_of DESC")
	q.Set("$limit", "200")
	reqURL := fmt.Sprintf("%s?%s", c.endpoint, q.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	if c.apiKey != "" {
		req.Header.Set("X-App-Token", c.apiKey)
	}
	resp, err := c.src.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cftc socrata http %d", resp.StatusCode)
	}
	var raws []rawRecord
	if err := json.NewDecoder(resp.Body).Decode(&raws); err != nil {
		return nil, fmt.Errorf("cftc decode: %w", err)
	}
	out := make([]COTReport, 0, len(raws))
	for _, r := range raws {
		out = append(out, mapRawToReport(r))
	}
	return out, nil
}

// ContractToCurrency 主要外汇期货合约代码到币种的映射（legacy 数据集）。
var ContractToCurrency = map[string]string{
	"099741": "EUR", // Euro FX
	"096742": "GBP", // British Pound
	"097741": "JPY", // Japanese Yen
	"092741": "CHF", // Swiss Franc
	"232741": "AUD", // Australian Dollar
	"090741": "CAD", // Canadian Dollar
	"112741": "NZD", // New Zealand Dollar
}

// FetchAllForCurrency 拉取指定币种最近 weeks 周的 legacy 报表。
func (c *Client) FetchAllForCurrency(ctx context.Context, currency string, weeks int) ([]COTReport, error) {
	var contractCode string
	for code, curr := range ContractToCurrency {
		if curr == currency {
			contractCode = code
			break
		}
	}
	if contractCode == "" {
		return nil, fmt.Errorf("cftc: unknown currency %q", currency)
	}
	to := time.Now().UTC()
	from := to.AddDate(0, 0, -weeks*7)
	return c.FetchLegacyReports(ctx, contractCode, from, to)
}

// ParseReportDate 把 CFTC 字段里的 "YYYY-MM-DDTHH:MM:SS.000" 截取为 time.Time。
func ParseReportDate(dateStr string) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)
	if len(dateStr) >= 10 {
		if t, err := time.Parse("2006-01-02", dateStr[:10]); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cftc: bad date %q", dateStr)
}

// mapRawToReport 把字符串字段转换为强类型字段。
func mapRawToReport(r rawRecord) COTReport {
	return COTReport{
		ReportDateAsOf:           r["report_date_as_of"],
		CftcContractMarketCode:   r["cftc_contract_market_code"],
		CftcMarketCode:           r["cftc_market_code"],
		CftcRegionCode:           r["cftc_region_code"],
		CftcCommodityCode:        r["cftc_commodity_code"],
		OpenInterestAll:          atoi64(r["open_interest_all"]),
		NoncommPositionsLongAll:  atoi64(r["noncomm_positions_long_all"]),
		NoncommPositionsShortAll: atoi64(r["noncomm_positions_short_all"]),
		CommPositionsLongAll:     atoi64(r["comm_positions_long_all"]),
		CommPositionsShortAll:    atoi64(r["comm_positions_short_all"]),
		NonreptPositionsLongAll:  atoi64(r["nonrept_positions_long_all"]),
		NonreptPositionsShortAll: atoi64(r["nonrept_positions_short_all"]),
	}
}

func atoi64(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}

// ParseCSVReport 解析 CFTC 周报 CSV 文件（兼容历史脚本的离线导入路径）。
func ParseCSVReport(r io.Reader) ([]COTReport, error) {
	reader := csv.NewReader(r)
	headers, err := reader.Read()
	if err != nil {
		return nil, err
	}
	col := map[string]int{}
	for i, h := range headers {
		col[strings.TrimSpace(strings.ToLower(h))] = i
	}
	pick := func(rec []string, name string) string {
		if i, ok := col[name]; ok && i < len(rec) {
			return strings.TrimSpace(rec[i])
		}
		return ""
	}
	var out []COTReport
	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		out = append(out, COTReport{
			ReportDateAsOf:           pick(rec, "report_date_as_of"),
			CftcContractMarketCode:   pick(rec, "cftc_contract_market_code"),
			CftcMarketCode:           pick(rec, "cftc_market_code"),
			CftcRegionCode:           pick(rec, "cftc_region_code"),
			CftcCommodityCode:        pick(rec, "cftc_commodity_code"),
			OpenInterestAll:          atoi64(pick(rec, "open_interest_all")),
			NoncommPositionsLongAll:  atoi64(pick(rec, "noncomm_positions_long_all")),
			NoncommPositionsShortAll: atoi64(pick(rec, "noncomm_positions_short_all")),
			CommPositionsLongAll:     atoi64(pick(rec, "comm_positions_long_all")),
			CommPositionsShortAll:    atoi64(pick(rec, "comm_positions_short_all")),
			NonreptPositionsLongAll:  atoi64(pick(rec, "nonrept_positions_long_all")),
			NonreptPositionsShortAll: atoi64(pick(rec, "nonrept_positions_short_all")),
		})
	}
	return out, nil
}
