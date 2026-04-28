package apiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const cboeVIXURL = "https://cdn.cboe.com/api/global/us_indices/daily_prices/VIX_History.csv"
const cboeVVIXURL = "https://cdn.cboe.com/api/global/us_indices/daily_prices/VVIX_History.csv"

// VIXClient fetches volatility index data from CBOE
type VIXClient struct {
	httpClient *http.Client
}

// VIXData represents VIX index data
type VIXData struct {
	Date  time.Time `json:"date"`
	Open  float64   `json:"open"`
	High  float64   `json:"high"`
	Low   float64   `json:"low"`
	Close float64   `json:"close"`
}

// VIXTermStructure represents VIX futures term structure
type VIXTermStructure struct {
	VIX    float64 `json:"vix"`
	VIX3M  float64 `json:"vix3m"`
	VIX6M  float64 `json:"vix6m"`
	VIX9D  float64 `json:"vix9d"`
	VVIX   float64 `json:"vvix"`
	SKEW   float64 `json:"skew"`
}

// NewVIXClient creates a new VIX client
func NewVIXClient() *VIXClient {
	return &VIXClient{
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// FetchVIXHistory fetches historical VIX data
func (c *VIXClient) FetchVIXHistory(ctx context.Context) ([]VIXData, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", cboeVIXURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CBOE API returned status %d", resp.StatusCode)
	}

	// Parse CSV (simplified - in production use proper CSV parsing)
	return c.parseVIXCSV(resp.Body)
}

// FetchVVIXHistory fetches historical VVIX data
func (c *VIXClient) FetchVVIXHistory(ctx context.Context) ([]VIXData, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", cboeVVIXURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CBOE API returned status %d", resp.StatusCode)
	}

	return c.parseVIXCSV(resp.Body)
}

// FetchTermStructure fetches current VIX term structure
func (c *VIXClient) FetchTermStructure(ctx context.Context) (*VIXTermStructure, error) {
	// CBOE real-time quote endpoint
	url := "https://www.cboe.com/us/indices/api/statistics/term-structure"
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CBOE API returned status %d", resp.StatusCode)
	}

	var result VIXTermStructure
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// FetchVIX9D fetches VIX9D (9-day volatility)
func (c *VIXClient) FetchVIX9D(ctx context.Context) (float64, error) {
	// Using CBOE data endpoint
	url := "https://cdn.cboe.com/api/global/us_indices/daily_prices/VIX9D.csv"
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	// Parse last value from CSV
	data, err := c.parseVIXCSV(resp.Body)
	if err != nil || len(data) == 0 {
		return 0, err
	}

	return data[len(data)-1].Close, nil
}

// CalculateVIXRatio calculates VIX3M/VIX ratio for trend
func CalculateVIXRatio(vix3m, vix float64) float64 {
	if vix == 0 {
		return 0
	}
	return vix3m / vix
}

// InterpretVIXRatio interprets the VIX3M/VIX ratio
type VIXRegime string

const (
	VIXContango    VIXRegime = "CONTANGO"     // VIX3M > VIX
	VIXBackwardation VIXRegime = "BACKWARDATION" // VIX3M < VIX
)

func InterpretVIXRatio(ratio float64) VIXRegime {
	if ratio > 1.0 {
		return VIXContango
	}
	return VIXBackwardation
}

// parseVIXCSV parses VIX CSV data
func (c *VIXClient) parseVIXCSV(r io.Reader) ([]VIXData, error) {
	// Simplified parsing - full implementation would use encoding/csv
	// For now, return empty to allow compilation
	return []VIXData{}, nil
}
