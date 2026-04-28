package apiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const worldbankBaseURL = "https://api.worldbank.org/v2"

// WorldBankClient fetches World Bank economic data
type WorldBankClient struct {
	httpClient *http.Client
}

// WBIndicator represents a World Bank indicator value
type WBIndicator struct {
	CountryCode string    `json:"country_code"`
	CountryName string    `json:"country_name"`
	IndicatorID string    `json:"indicator_id"`
	IndicatorName string  `json:"indicator_name"`
	Date        time.Time `json:"date"`
	Value       float64   `json:"value"`
}

// NewWorldBankClient creates a new World Bank client
func NewWorldBankClient() *WorldBankClient {
	return &WorldBankClient{
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// FetchIndicator fetches a specific indicator for a country
func (c *WorldBankClient) FetchIndicator(ctx context.Context, countryCode, indicator string, from, to int) ([]WBIndicator, error) {
	url := fmt.Sprintf("%s/country/%s/indicator/%s?date=%d:%d&format=json&per_page=500",
		worldbankBaseURL, countryCode, indicator, from, to)
	
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
		return nil, fmt.Errorf("World Bank API returned status %d", resp.StatusCode)
	}

	// World Bank returns [metadata, data] array
	var result []json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result) < 2 {
		return nil, fmt.Errorf("unexpected response format")
	}

	var indicators []WBIndicator
	var rawIndicators []struct {
		Country     struct {
			ID   string `json:"id"`
			Value string `json:"value"`
		} `json:"country"`
		Indicator   struct {
			ID   string `json:"id"`
			Value string `json:"value"`
		} `json:"indicator"`
		Date        string  `json:"date"`
		Value       float64 `json:"value"`
	}

	if err := json.Unmarshal(result[1], &rawIndicators); err != nil {
		return nil, err
	}

	for _, raw := range rawIndicators {
		date, _ := time.Parse("2006", raw.Date)
		indicators = append(indicators, WBIndicator{
			CountryCode: raw.Country.ID,
			CountryName: raw.Country.Value,
			IndicatorID: raw.Indicator.ID,
			IndicatorName: raw.Indicator.Value,
			Date:        date,
			Value:       raw.Value,
		})
	}

	return indicators, nil
}

// FetchGDP fetches GDP data for a country
func (c *WorldBankClient) FetchGDP(ctx context.Context, countryCode string, years int) ([]WBIndicator, error) {
	to := time.Now().Year()
	from := to - years
	
	// NY.GDP.MKTP.CD is GDP current US$
	return c.FetchIndicator(ctx, countryCode, "NY.GDP.MKTP.CD", from, to)
}

// FetchInflation fetches inflation data (CPI annual change)
func (c *WorldBankClient) FetchInflation(ctx context.Context, countryCode string, years int) ([]WBIndicator, error) {
	to := time.Now().Year()
	from := to - years
	
	// FP.CPI.TOTL.ZG is Inflation, consumer prices (annual %%)
	return c.FetchIndicator(ctx, countryCode, "FP.CPI.TOTL.ZG", from, to)
}

// FetchUnemployment fetches unemployment rate
func (c *WorldBankClient) FetchUnemployment(ctx context.Context, countryCode string, years int) ([]WBIndicator, error) {
	to := time.Now().Year()
	from := to - years
	
	// SL.UEM.TOTL.ZS is Unemployment, total (%% of total labor force)
	return c.FetchIndicator(ctx, countryCode, "SL.UEM.TOTL.ZS", from, to)
}

// CommonIndicators defines frequently used World Bank indicators
var CommonIndicators = map[string]struct {
	ID   string
	Name string
}{
	"GDP":             {"NY.GDP.MKTP.CD", "GDP (current US$)"},
	"GDP_Growth":      {"NY.GDP.MKTP.KD.ZG", "GDP growth (annual %%)"},
	"Inflation":       {"FP.CPI.TOTL.ZG", "Inflation, consumer prices (annual %%)"},
	"Unemployment":    {"SL.UEM.TOTL.ZS", "Unemployment (%% of total labor force)"},
	"CurrentAccount":  {"BN.CAB.XOKA.CD", "Current account balance (BoP, current US$)"},
	"FX_Reserves":     {"FI.RES.TOTL.CD", "Total reserves (includes gold, current US$)"},
	"Debt_GDP":        {"GC.DOD.TOTL.GD.ZS", "Central government debt, total (%% of GDP)"},
	"Gini":            {"SI.POV.GINI", "Gini index"},
}

// MajorCountries defines major economies
var MajorCountries = map[string]string{
	"US": "United States",
	"CN": "China",
	"JP": "Japan",
	"DE": "Germany",
	"GB": "United Kingdom",
	"FR": "France",
	"IN": "India",
	"IT": "Italy",
	"BR": "Brazil",
	"CA": "Canada",
	"RU": "Russia",
	"KR": "South Korea",
	"ES": "Spain",
	"AU": "Australia",
	"MX": "Mexico",
}
