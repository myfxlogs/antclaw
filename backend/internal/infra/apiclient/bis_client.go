package apiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const bisBaseURL = "https://stats.bis.org/api/v2"

// BISClient fetches Bank for International Settlements data
type BISClient struct {
	httpClient *http.Client
}

// PolicyRate represents central bank policy rates
type PolicyRate struct {
	CountryCode string    `json:"country_code"`
	CountryName string    `json:"country_name"`
	Date        time.Time `json:"date"`
	Rate        float64   `json:"rate"`
	RateType    string    `json:"rate_type"`
}

// CreditGap represents credit-to-GDP gap data
type CreditGap struct {
	CountryCode string    `json:"country_code"`
	CountryName string    `json:"country_name"`
	Date        time.Time `json:"date"`
	CreditGap   float64   `json:"credit_gap"`
	CreditGDP   float64   `json:"credit_gdp"`
}

// REER represents Real Effective Exchange Rate
type REER struct {
	CountryCode string    `json:"country_code"`
	CountryName string    `json:"country_name"`
	Date        time.Time `json:"date"`
	REERValue   float64   `json:"reer"`
	Broad       float64   `json:"broad"`  // Broad REER
	Narrow      float64   `json:"narrow"` // Narrow REER
}

// NewBISClient creates a new BIS client
func NewBISClient() *BISClient {
	return &BISClient{
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// FetchPolicyRates fetches central bank policy rates
func (c *BISClient) FetchPolicyRates(ctx context.Context, countryCode string) ([]PolicyRate, error) {
	url := fmt.Sprintf("%s/data/BISWEB_IRF/c/%s/latest", bisBaseURL, countryCode)
	
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
		return nil, fmt.Errorf("BIS API returned status %d", resp.StatusCode)
	}

	var result struct {
		Observations []struct {
			TimePeriod string  `json:"TIME_PERIOD"`
			Value      float64 `json:"OBS_VALUE"`
		} `json:"observations"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var rates []PolicyRate
	for _, obs := range result.Observations {
		date, err := time.Parse("2006-01", obs.TimePeriod)
		if err != nil {
			continue
		}
		rates = append(rates, PolicyRate{
			CountryCode: countryCode,
			Date:        date,
			Rate:        obs.Value,
			RateType:    "Policy",
		})
	}

	return rates, nil
}

// FetchCreditGap fetches credit-to-GDP gap data
func (c *BISClient) FetchCreditGap(ctx context.Context, countryCode string) ([]CreditGap, error) {
	url := fmt.Sprintf("%s/data/BISWEB_CPGAP/c/%s/latest", bisBaseURL, countryCode)
	
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
		return nil, fmt.Errorf("BIS API returned status %d", resp.StatusCode)
	}

	var result struct {
		Observations []struct {
			TimePeriod string  `json:"TIME_PERIOD"`
			GapValue   float64 `json:"OBS_VALUE"`
		} `json:"observations"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var gaps []CreditGap
	for _, obs := range result.Observations {
		date, err := time.Parse("2006-Q1", obs.TimePeriod)
		if err != nil {
			continue
		}
		gaps = append(gaps, CreditGap{
			CountryCode: countryCode,
			Date:        date,
			CreditGap:   obs.GapValue,
		})
	}

	return gaps, nil
}

// FetchREER fetches Real Effective Exchange Rates
func (c *BISClient) FetchREER(ctx context.Context, countryCode string, broad bool) ([]REER, error) {
	series := "REER-N"
	if broad {
		series = "REER-B"
	}
	
	url := fmt.Sprintf("%s/data/BISWEB_EER/M.%s.US", bisBaseURL, series)
	
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
		return nil, fmt.Errorf("BIS API returned status %d", resp.StatusCode)
	}

	var result struct {
		Observations []struct {
			TimePeriod string  `json:"TIME_PERIOD"`
			Value      float64 `json:"OBS_VALUE"`
		} `json:"observations"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var rates []REER
	for _, obs := range result.Observations {
		date, err := time.Parse("2006-01", obs.TimePeriod)
		if err != nil {
			continue
		}
		rates = append(rates, REER{
			CountryCode: countryCode,
			Date:        date,
			REERValue:   obs.Value,
		})
	}

	return rates, nil
}

// CalculateCarryRate calculates carry rate between two currencies
func CalculateCarryRate(highRate, lowRate float64) float64 {
	return highRate - lowRate
}

// CarryTrade represents a carry trade opportunity
type CarryTrade struct {
	HighCurrency string  `json:"high_currency"`
	HighRate     float64 `json:"high_rate"`
	LowCurrency  string  `json:"low_currency"`
	LowRate      float64 `json:"low_rate"`
	CarryRate    float64 `json:"carry_rate"`
}

// FindBestCarryTrades finds the best carry trade opportunities
func FindBestCarryTrades(rates map[string]float64) []CarryTrade {
	// Find highest and lowest rates
	var maxRate, minRate float64
	var maxCurrency, minCurrency string
	
	first := true
	for curr, rate := range rates {
		if first {
			maxRate = rate
			minRate = rate
			maxCurrency = curr
			minCurrency = curr
			first = false
			continue
		}
		
		if rate > maxRate {
			maxRate = rate
			maxCurrency = curr
		}
		if rate < minRate {
			minRate = rate
			minCurrency = curr
		}
	}
	
	return []CarryTrade{{
		HighCurrency: maxCurrency,
		HighRate:     maxRate,
		LowCurrency:  minCurrency,
		LowRate:      minRate,
		CarryRate:    maxRate - minRate,
	}}
}

// MajorEconomies defines major economies to track
var MajorEconomies = map[string]string{
	"US": "United States",
	"EA": "Euro Area",
	"JP": "Japan",
	"GB": "United Kingdom",
	"CH": "Switzerland",
	"CA": "Canada",
	"AU": "Australia",
	"NZ": "New Zealand",
}
