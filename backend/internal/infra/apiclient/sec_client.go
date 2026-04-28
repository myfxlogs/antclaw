package apiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const secEDGARURL = "https://www.sec.gov/Archives/edgar/daily-index"
const secAPIBase = "https://api.sec-api.io"

// SECClient fetches SEC 13F filings
type SECClient struct {
	httpClient *http.Client
	apiKey     string
	userAgent  string
}

// Holding13F represents a single 13F holding
type Holding13F struct {
	CIK           string  `json:"cik"`
	IssuerName    string  `json:"issuer_name"`
	ClassTitle    string  `json:"class_title"`
	CUSIP         string  `json:"cusip"`
	Shares        int64   `json:"shares"`
	ShareType     string  `json:"share_type"` // SH, PRN, etc
	Value         float64 `json:"value"`
	FilingQuarter string  `json:"filing_quarter"`
}

// Institution13F represents an institution's 13F filing
type Institution13F struct {
	CIK           string       `json:"cik"`
	Name          string       `json:"name"`
	FilingDate    time.Time    `json:"filing_date"`
	PeriodEndDate time.Time    `json:"period_end_date"`
	Holdings      []Holding13F `json:"holdings"`
	TotalValue    float64      `json:"total_value"`
}

// NewSECClient creates a new SEC client
func NewSECClient(apiKey, userAgent string) *SECClient {
	return &SECClient{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		apiKey:     apiKey,
		userAgent:  userAgent,
	}
}

// Fetch13FFilings fetches 13F filings for a CIK
func (c *SECClient) Fetch13FFilings(ctx context.Context, cik string, quarter string) (*Institution13F, error) {
	params := url.Values{}
	params.Set("cik", cik)
	params.Set("quarter", quarter)
	
	reqURL := fmt.Sprintf("%s/v1/form-13f?%s", secAPIBase, params.Encode())
	
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Authorization", c.apiKey)
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("SEC API returned status %d", resp.StatusCode)
	}

	var result Institution13F
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// FetchLatest13F fetches the most recent 13F filing
func (c *SECClient) FetchLatest13F(ctx context.Context, cik string) (*Institution13F, error) {
	// Get current quarter
	now := time.Now()
	quarter := fmt.Sprintf("%d-Q%d", now.Year(), (now.Month()-1)/3+1)
	
	return c.Fetch13FFilings(ctx, cik, quarter)
}

// SearchByCUSIP searches for holdings by CUSIP
func (c *SECClient) SearchByCUSIP(ctx context.Context, cusip string) ([]Institution13F, error) {
	params := url.Values{}
	params.Set("cusip", cusip)
	
	reqURL := fmt.Sprintf("%s/v1/form-13f/search?%s", secAPIBase, params.Encode())
	
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Authorization", c.apiKey)
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("SEC API returned status %d", resp.StatusCode)
	}

	var result struct {
		Filings []Institution13F `json:"filings"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Filings, nil
}

// CalculateChanges calculates changes between two filings
func CalculateChanges(old, new []Holding13F) []HoldingChange {
	oldMap := make(map[string]Holding13F)
	for _, h := range old {
		oldMap[h.CUSIP] = h
	}
	
	newMap := make(map[string]Holding13F)
	for _, h := range new {
		newMap[h.CUSIP] = h
	}

	var changes []HoldingChange
	
	// Check for new positions and changes
	for cusip, newHolding := range newMap {
		change := HoldingChange{
			CUSIP:      cusip,
			IssuerName: newHolding.IssuerName,
		}
		
		if oldHolding, exists := oldMap[cusip]; exists {
			change.ChangeType = "MODIFIED"
			change.SharesChange = newHolding.Shares - oldHolding.Shares
			change.ValueChange = newHolding.Value - oldHolding.Value
			change.PctChange = (newHolding.Value - oldHolding.Value) / oldHolding.Value * 100
		} else {
			change.ChangeType = "NEW"
			change.SharesChange = newHolding.Shares
			change.ValueChange = newHolding.Value
			change.PctChange = 100
		}
		
		changes = append(changes, change)
	}
	
	// Check for closed positions
	for cusip, oldHolding := range oldMap {
		if _, exists := newMap[cusip]; !exists {
			changes = append(changes, HoldingChange{
				CUSIP:        cusip,
				IssuerName:   oldHolding.IssuerName,
				ChangeType:   "CLOSED",
				SharesChange: -oldHolding.Shares,
				ValueChange:  -oldHolding.Value,
				PctChange:    -100,
			})
		}
	}
	
	return changes
}

// HoldingChange represents a change in holding
type HoldingChange struct {
	CUSIP        string  `json:"cusip"`
	IssuerName   string  `json:"issuer_name"`
	ChangeType   string  `json:"change_type"` // NEW, MODIFIED, CLOSED
	SharesChange int64   `json:"shares_change"`
	ValueChange  float64 `json:"value_change"`
	PctChange    float64 `json:"pct_change"`
}

// GetInstitutionalFlows calculates aggregate institutional flows
func GetInstitutionalFlows(changes []HoldingChange) (inflow, outflow float64) {
	for _, c := range changes {
		switch c.ChangeType {
		case "NEW":
			inflow += c.ValueChange
		case "MODIFIED":
			if c.ValueChange > 0 {
				inflow += c.ValueChange
			} else {
				outflow -= c.ValueChange
			}
		case "CLOSED":
			outflow -= c.ValueChange
		}
	}
	return inflow, outflow
}

// NotableInstitutions contains major institutional investors to track
var NotableInstitutions = map[string]string{
	"0001067983": "Berkshire Hathaway",
	"0001336528": "Bridgewater Associates",
	"0001350694": "Renaissance Technologies",
	"0001422224": "Two Sigma",
	"0001610876": "Citadel",
	"0001103804": "Appaloosa",
}
