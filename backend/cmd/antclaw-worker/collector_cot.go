// COT采集器
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"


	apiclient "github.com/antclaw/antclaw/internal/infra/apiclient"
	cftcclient "github.com/antclaw/antclaw/internal/infra/apiclient/cftc"
)

const cftcTFFEndpoint = "https://publicreporting.cftc.gov/resource/gpe5-46if.json"

var cotContracts = []struct {
	Code     string
	Currency string
	Name     string
}{
	{"099741", "EUR", "EURO FX"},
	{"096742", "JPY", "JAPANESE YEN"},
	{"092741", "GBP", "BRITISH POUND"},
	{"092742", "AUD", "AUSTRALIAN DOLLAR"},
	{"090741", "CHF", "SWISS FRANC"},
	{"098662", "CAD", "CANADIAN DOLLAR"},
	{"112741", "NZD", "NZ DOLLAR"},
	{"088691", "USD", "USD INDEX"},
	{"067651", "GOLD", "GOLD"},
	{"084691", "SILVER", "SILVER"},
	{"067411", "OIL", "CRUDE OIL"},
}

type socrataCOT struct {
	ReportDate    string `json:"report_date_as_yyyy_mm_dd"`
	ContractCode  string `json:"cftc_contract_market_code"`
	OpenInterest  string `json:"open_interest_all"`
	DealerLong    string `json:"dealer_positions_long_all"`
	DealerShort   string `json:"dealer_positions_short_all"`
	AssetMgrLong  string `json:"asset_mgr_positions_long"`
	AssetMgrShort string `json:"asset_mgr_positions_short"`
	LevFundLong   string `json:"lev_money_positions_long"`
	LevFundShort  string `json:"lev_money_positions_short"`
	OtherRptLong  string `json:"other_rept_positions_long"`
	OtherRptShort string `json:"other_rept_positions_short"`
	NonRptLong    string `json:"non_reportable_positions_long_all"`
	NonRptShort   string `json:"non_reportable_positions_short_all"`
}

// collectCOT fetches COT data from CFTC Socrata and persists to database.
func collectCOT(ctx context.Context, dbpool *pgxpool.Pool, logger *slog.Logger) error {
	logger.Info("COT: collecting from CFTC Socrata")

	client := &http.Client{Timeout: 30 * time.Second}
	// 优先使用 CFTC Socrata legacy（薄客户端），失败回退手写 Socrata 调用
	cftcSrc := apiclient.NewSource("cftc", apiclient.Options{Timeout: 30 * time.Second})
	cftc := cftcclient.NewClient(cftcSrc, "")
	totalInserted := 0

	for _, contract := range cotContracts {
				// 先尝试 CFTC v1 legacy（52 周）
		from := time.Now().AddDate(0, 0, -52*7)
		to := time.Now()
		if records, err := cftc.FetchLegacyReports(ctx, contract.Code, from, to); err == nil && len(records) > 0 {
			ins := saveLegacyReports(ctx, dbpool, records, contract.Currency, logger)
			totalInserted += ins
			logger.Info("COT legacy synced", "currency", contract.Currency, "records", ins)
			continue
		}

params := url.Values{}
		params.Set("cftc_contract_market_code", contract.Code)
		params.Set("$order", "report_date_as_yyyy_mm_dd DESC")
		params.Set("$limit", "52")

		reqURL := fmt.Sprintf("%s?%s", cftcTFFEndpoint, params.Encode())

		req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
		if err != nil {
			logger.Warn("COT request build failed", "currency", contract.Currency, "error", err)
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			logger.Warn("COT fetch failed", "currency", contract.Currency, "error", err)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil || resp.StatusCode != 200 {
			logger.Warn("COT response error", "currency", contract.Currency, "status", resp.StatusCode)
			continue
		}

		var records []socrataCOT
		if err := json.Unmarshal(body, &records); err != nil {
			logger.Warn("COT parse failed", "currency", contract.Currency, "error", err)
			continue
		}

		inserted := saveCOTRecords(ctx, dbpool, records, contract.Currency, logger)
		totalInserted += inserted
	}

	logger.Info("COT collection completed", "total_inserted", totalInserted)
	return nil
}

// TFF field mapping:
// noncomm (non-reportable) = calculated as open_interest - total_reportable
// comm (commercial) = dealer_positions
// dealer = dealer_positions (same as comm in TFF)
// levfund = lev_money_positions
// mm (money managers) = other_rept_positions
// swap = asset_mgr_positions
func saveCOTRecords(ctx context.Context, dbpool *pgxpool.Pool, records []socrataCOT, currency string, logger *slog.Logger) int {
	count := 0
	for _, r := range records {
		reportDate, err := time.Parse("2006-01-02T15:04:05.000", r.ReportDate)
		if err != nil {
			reportDate, err = time.Parse("2006-01-02", r.ReportDate[:10])
			if err != nil {
				continue
			}
		}

		oi := parseInt64(r.OpenInterest)
		nonCommLong := oi - parseInt64(r.NonRptLong)
		nonCommShort := oi - parseInt64(r.NonRptShort)

		rawJSON, _ := json.Marshal(r)
		_, err = dbpool.Exec(ctx, `
			INSERT INTO cot_records 
			(report_date, contract_code, currency, noncomm_long, noncomm_short, comm_long, comm_short,
			 dealer_long, dealer_short, levfund_long, levfund_short, mm_long, mm_short, 
			 swap_long, swap_short, total_oi, raw_json)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
			ON CONFLICT (report_date, contract_code) DO UPDATE SET
			  noncomm_long = EXCLUDED.noncomm_long,
			  noncomm_short = EXCLUDED.noncomm_short,
			  comm_long = EXCLUDED.comm_long,
			  comm_short = EXCLUDED.comm_short,
			  raw_json = EXCLUDED.raw_json`,
			reportDate, currency, currency,
			nonCommLong, nonCommShort,
			parseInt64(r.DealerLong), parseInt64(r.DealerShort),
			parseInt64(r.DealerLong), parseInt64(r.DealerShort),
			parseInt64(r.LevFundLong), parseInt64(r.LevFundShort),
			parseInt64(r.OtherRptLong), parseInt64(r.OtherRptShort),
			parseInt64(r.AssetMgrLong), parseInt64(r.AssetMgrShort),
			oi,
			rawJSON,
		)
		if err == nil {
			count++
		}
	}
	return count
}


// saveLegacyReports: 落地 CFTC v1 legacy 报表至 cot_records（细分项缺失置 0）
func saveLegacyReports(ctx context.Context, dbpool *pgxpool.Pool, records []cftcclient.COTReport, currency string, logger *slog.Logger) int {
	count := 0
	for _, r := range records {
		reportDate, err := time.Parse("2006-01-02", r.ReportDateAsOf)
		if err != nil {
			continue
		}
		oi := r.OpenInterestAll
		nonCommLong := r.NoncommPositionsLongAll
		nonCommShort := r.NoncommPositionsShortAll
		commLong := r.CommPositionsLongAll
		commShort := r.CommPositionsShortAll
		rawJSON, _ := json.Marshal(r)
		_, err = dbpool.Exec(ctx, `
			INSERT INTO cot_records 
			(report_date, contract_code, currency, noncomm_long, noncomm_short, comm_long, comm_short,
			 dealer_long, dealer_short, levfund_long, levfund_short, mm_long, mm_short, 
			 swap_long, swap_short, total_oi, raw_json)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
			ON CONFLICT (report_date, contract_code) DO UPDATE SET
			  noncomm_long = EXCLUDED.noncomm_long,
			  noncomm_short = EXCLUDED.noncomm_short,
			  comm_long = EXCLUDED.comm_long,
			  comm_short = EXCLUDED.comm_short,
			  raw_json = EXCLUDED.raw_json`,
			reportDate, currency, currency,
			nonCommLong, nonCommShort,
			commLong, commShort,
			commLong, commShort,
			int64(0), int64(0),
			int64(0), int64(0),
			int64(0), int64(0),
			oi,
			rawJSON,
		)
		if err == nil {
			count++
		}
	}
	return count
}
