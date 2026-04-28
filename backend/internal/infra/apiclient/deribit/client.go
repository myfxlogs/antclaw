package deribit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/antclaw/antclaw/internal/infra/apiclient"
)

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
