package deribit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/antclaw/antclaw/internal/infra/apiclient"
)

var timeNow = time.Now

// Client 封装 Deribit 公开 API（无需鉴权）。
type Client struct {
	src  apiclient.Source
	base string
}

func NewClient(src apiclient.Source) *Client {
	return &Client{src: src, base: "https://www.deribit.com/api/v2"}
}

type Instrument struct {
	InstrumentName string  `json:"instrument_name"`
	Strike         float64 `json:"strike"`
	OptionType     string  `json:"option_type"`
	ExpirationTs   int64   `json:"expiration_timestamp"`
	IsActive       bool    `json:"is_active"`
}

type instrumentsResp struct {
	Result []Instrument `json:"result"`
}

// GetInstruments 列出指定 currency 的期权合约（kind=option）。
func (c *Client) GetInstruments(ctx context.Context, currency string) ([]Instrument, error) {
	url := fmt.Sprintf("%s/public/get_instruments?currency=%s&kind=option&expired=false", c.base, currency)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := c.src.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("deribit http %d", resp.StatusCode)
	}
	var out instrumentsResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out.Result, nil
}

type BookSummary struct {
	InstrumentName string  `json:"instrument_name"`
	MarkIV         float64 `json:"mark_iv"`
	OpenInterest   float64 `json:"open_interest"`
	UnderlyingPrice float64 `json:"underlying_price"`
}

type bookSummariesResp struct {
	Result []BookSummary `json:"result"`
}

// GetBookSummaries 返回所有期权合约的市场汇总（含 mark_iv 与 open_interest）。
func (c *Client) GetBookSummaries(ctx context.Context, currency string) ([]BookSummary, error) {
	url := fmt.Sprintf("%s/public/get_book_summary_by_currency?currency=%s&kind=option", c.base, currency)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := c.src.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("deribit http %d", resp.StatusCode)
	}
	var out bookSummariesResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out.Result, nil
}

// DVOLPoint 单条 DVOL 时间序列点（OHLC + ts，单位百分点）。
type DVOLPoint struct {
	Timestamp int64
	Open      float64
	High      float64
	Low       float64
	Close     float64
}

// GetVolatilityIndexData 拉取 Deribit DVOL 指数（BTC/ETH 等）历史数据，返回升序点。
// resolution 单位秒；常用 60(1m)/3600(1h)/43200(12h)/86400(1d)。
// startMs/endMs 为 Unix 毫秒；endMs<=0 时取当前时间。
func (c *Client) GetVolatilityIndexData(ctx context.Context, currency string, resolutionSec int, startMs, endMs int64) ([]DVOLPoint, error) {
	if resolutionSec <= 0 {
		resolutionSec = 86400
	}
	if endMs <= 0 {
		endMs = nowMs()
	}
	if startMs <= 0 {
		startMs = endMs - int64(30)*86400000 // 默认 30 天
	}
	url := fmt.Sprintf("%s/public/get_volatility_index_data?currency=%s&start_timestamp=%d&end_timestamp=%d&resolution=%d",
		c.base, currency, startMs, endMs, resolutionSec)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := c.src.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("deribit http %d", resp.StatusCode)
	}
	var raw struct {
		Result struct {
			Data           [][]float64 `json:"data"`
			Continuation   string      `json:"continuation"`
			Volatility     float64     `json:"volatility"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	out := make([]DVOLPoint, 0, len(raw.Result.Data))
	for _, row := range raw.Result.Data {
		if len(row) < 5 {
			continue
		}
		out = append(out, DVOLPoint{
			Timestamp: int64(row[0]),
			Open:      row[1],
			High:      row[2],
			Low:       row[3],
			Close:     row[4],
		})
	}
	return out, nil
}

func nowMs() int64 { return timeNow().UnixMilli() }
