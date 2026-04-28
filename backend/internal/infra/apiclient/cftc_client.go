package apiclient

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
)

const cftcBaseURL = "https://api.cftc.gov/v1"

// CFTCClient fetches Commitment of Traders reports
type CFTCClient struct {
	httpClient *http.Client
	apiKey     string
}

// COTReport represents a single COT report record
type COTReport struct {
	ReportDateAsOf      string `json:"report_date_as_of"`
	CftcContractMarketCode string `json:"cftc_contract_market_code"`
	CftcMarketCode      string `json:"cftc_market_code"`
	CftcRegionCode      string `json:"cftc_region_code"`
	CftcCommodityCode   string `json:"cftc_commodity_code"`
	OpenInterestAll     int64  `json:"open_interest_all"`
	NoncommPositionsLongAll int64 `json:"noncomm_positions_long_all"`
	NoncommPositionsShortAll int64 `json:"noncomm_positions_short_all"`
	CommPositionsLongAll int64 `json:"comm_positions_long_all"`
	CommPositionsShortAll int64 `json:"comm_positions_short_all"`
	NonreptPositionsLongAll int64 `json:"nonrept_positions_long_all"`
	NonreptPositionsShortAll int64 `json:"nonrept_positions_short_all"`
}

// NewCFTCClient creates a new CFTC API client
func NewCFTCClient(apiKey string) *CFTCClient {
	return &CFTCClient{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		apiKey:     apiKey,
	}
}

// FetchLegacyReports fetches legacy COT reports (Disaggregated)
func (c *CFTCClient) FetchLegacyReports(ctx context.Context, marketCode string, from, to time.Time) ([]COTReport, error) {
	params := url.Values{}
	params.Set("cftc_contract_market_code", marketCode)
	params.Set("report_date_as_of.ge", from.Format("2006-01-02"))
	params.Set("report_date_as_of.le", to.Format("2006-01-02"))

	reqURL := fmt.Sprintf("%s/cot/futures-and-options/historical?%s", cftcBaseURL, params.Encode())
	if c.apiKey != "" {
		reqURL += "&key=" + c.apiKey
	}

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CFTC API returned status %d", resp.StatusCode)
	}

	var result struct {
		Records []COTReport `json:"records"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Records, nil
}

// FetchDisaggregatedReports fetches disaggregated COT reports with dealer/swap separation
func (c *CFTCClient) FetchDisaggregatedReports(ctx context.Context, marketCode string, from, to time.Time) ([]DisaggregatedReport, error) {
	params := url.Values{}
	params.Set("cftc_contract_market_code", marketCode)
	params.Set("report_date_as_of.ge", from.Format("2006-01-02"))
	params.Set("report_date_as_of.le", to.Format("2006-01-02"))

	reqURL := fmt.Sprintf("%s/cot-disaggregated/historical?%s", cftcBaseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CFTC API returned status %d", resp.StatusCode)
	}

	var result struct {
		Records []DisaggregatedReport `json:"records"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Records, nil
}

// DisaggregatedReport represents disaggregated COT data
type DisaggregatedReport struct {
	ReportDateAsOf           string `json:"report_date_as_of"`
	CftcContractMarketCode   string `json:"cftc_contract_market_code"`
	ProdMercPositionsLong    int64  `json:"prod_merc_positions_long"`
	ProdMercPositionsShort   int64  `json:"prod_merc_positions_short"`
	SwapPositionsLong        int64  `json:"swap_positions_long"`
	SwapPositionsShort       int64  `json:"swap_positions_short"`
	MmgrPositionsLong        int64  `json:"mmgr_positions_long"`
	MmgrPositionsShort       int64  `json:"mmgr_positions_short"`
	OtherReptPositionsLong   int64  `json:"other_rept_positions_long"`
	OtherReptPositionsShort  int64  `json:"other_rept_positions_short"`
	NonreptPositionsLong     int64  `json:"nonrept_positions_long"`
	NonreptPositionsShort    int64  `json:"nonrept_positions_short"`
}

// ParseReportDate parses the CFTC report date format
func ParseReportDate(dateStr string) (time.Time, error) {
	// CFTC format: "YYYY-MM-DD"
	return time.Parse("2006-01-02", dateStr)
}

// ContractToCurrency maps COT contract codes to currencies
// Note: These are legacy contract codes for COT Legacy Reports
var ContractToCurrency = map[string]string{
	"090741": "EUR", // EUR futures
	"096742": "GBP", // GBP futures
	"097741": "JPY", // JPY futures
	"092741": "CHF", // CHF futures
	"232741": "AUD", // AUD futures
	"090741F": "CAD", // CAD futures
	"112741": "NZD", // NZD futures
}

// FetchAllForCurrency fetches all COT reports for a currency
func (c *CFTCClient) FetchAllForCurrency(ctx context.Context, currency string, weeks int) ([]COTReport, error) {
	// Find contract code
	var contractCode string
	for code, curr := range ContractToCurrency {
		if curr == currency {
			contractCode = code
			break
		}
	}
	if contractCode == "" {
		return nil, fmt.Errorf("unknown currency: %s", currency)
	}

	to := time.Now()
	from := to.AddDate(0, 0, -weeks*7)

	return c.FetchLegacyReports(ctx, contractCode, from, to)
}

// ParseCSVReport parses a COT report from CSV format
func ParseCSVReport(r io.Reader) ([]COTReport, error) {
	reader := csv.NewReader(r)
	
	// Read header
	headers, err := reader.Read()
	if err != nil {
		return nil, err
	}

	// Create column index
	colMap := make(map[string]int)
	for i, h := range headers {
		colMap[strings.TrimSpace(strings.ToLower(h))] = i
	}

	var reports []COTReport
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		report := COTReport{
			ReportDateAsOf: getField(record, colMap, "report_date_as_of"),
			CftcContractMarketCode: getField(record, colMap, "cftc_contract_market_code"),
			CftcMarketCode: getField(record, colMap, "cftc_market_code"),
			CftcRegionCode: getField(record, colMap, "cftc_region_code"),
			CftcCommodityCode: getField(record, colMap, "cftc_commodity_code"),
		}

		if v, err := parseInt64(getField(record, colMap, "open_interest_all")); err == nil {
			report.OpenInterestAll = v
		}
		if v, err := parseInt64(getField(record, colMap, "noncomm_positions_long_all")); err == nil {
			report.NoncommPositionsLongAll = v
		}
		if v, err := parseInt64(getField(record, colMap, "noncomm_positions_short_all")); err == nil {
			report.NoncommPositionsShortAll = v
		}
		if v, err := parseInt64(getField(record, colMap, "comm_positions_long_all")); err == nil {
			report.CommPositionsLongAll = v
		}
		if v, err := parseInt64(getField(record, colMap, "comm_positions_short_all")); err == nil {
			report.CommPositionsShortAll = v
		}
		if v, err := parseInt64(getField(record, colMap, "nonrept_positions_long_all")); err == nil {
			report.NonreptPositionsLongAll = v
		}
		if v, err := parseInt64(getField(record, colMap, "nonrept_positions_short_all")); err == nil {
			report.NonreptPositionsShortAll = v
		}

		reports = append(reports, report)
	}

	return reports, nil
}

func getField(record []string, colMap map[string]int, name string) string {
	if idx, ok := colMap[name]; ok && idx < len(record) {
		return strings.TrimSpace(record[idx])
	}
	return ""
}

func parseInt64(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}
	return strconv.ParseInt(s, 10, 64)
}
