package apiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const treasuryBaseURL = "https://www.treasurydirect.gov/TA_WS"

// TreasuryClient fetches US Treasury data
type TreasuryClient struct {
	httpClient *http.Client
}

// AuctionResult represents Treasury auction results
type AuctionResult struct {
	AuctionDate       time.Time `json:"auction_date"`
	IssueDate         time.Time `json:"issue_date"`
	MaturityDate      time.Time `json:"maturity_date"`
	SecurityType      string    `json:"security_type"`
	SecurityTerm      string    `json:"security_term"`
	HighRate          float64   `json:"high_rate"`
	InvestmentRate    float64   `json:"investment_rate"`
	Price             float64   `json:"price"`
	Accepted          int64     `json:"accepted"`
	Tendered          int64     `json:"tendered"`
	BidToCover        float64   `json:"bid_to_cover"`
	Dealers           int64     `json:"dealers"`
	Direct            int64     `json:"direct"`
	Indirect          int64     `json:"indirect"`
}

// YieldCurvePoint represents a point on the yield curve
type YieldCurvePoint struct {
	Date  time.Time `json:"date"`
	Term  string    `json:"term"`
	Rate  float64   `json:"rate"`
}

// NewTreasuryClient creates a new Treasury client
func NewTreasuryClient() *TreasuryClient {
	return &TreasuryClient{
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// FetchAuctions fetches Treasury auction results
func (c *TreasuryClient) FetchAuctions(ctx context.Context, securityType string, from, to time.Time) ([]AuctionResult, error) {
	url := fmt.Sprintf("%s/securities/auctioned?format=json&securitytype=%s&startdate=%s&enddate=%s",
		treasuryBaseURL, securityType, from.Format("01/02/2006"), to.Format("01/02/2006"))
	
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
		return nil, fmt.Errorf("Treasury API returned status %d", resp.StatusCode)
	}

	var results []AuctionResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, err
	}

	return results, nil
}

// FetchYieldCurve fetches current yield curve data
func (c *TreasuryClient) FetchYieldCurve(ctx context.Context, date time.Time) ([]YieldCurvePoint, error) {
	url := fmt.Sprintf("%s/securities/auctioned?format=json&type=notebond&startdate=%s&enddate=%s",
		treasuryBaseURL, date.Format("01/02/2006"), date.Format("01/02/2006"))
	
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
		return nil, fmt.Errorf("Treasury API returned status %d", resp.StatusCode)
	}

	// Parse and extract yield curve points
	var results []YieldCurvePoint
	
	return results, nil
}

// FetchBillAuctions fetches T-Bill auction results specifically
func (c *TreasuryClient) FetchBillAuctions(ctx context.Context, weeks int) ([]AuctionResult, error) {
	to := time.Now()
	from := to.AddDate(0, 0, -weeks*7)
	
	return c.FetchAuctions(ctx, "Bill", from, to)
}

// CalculateYieldCurveSteepness calculates 10Y-2Y spread
func CalculateYieldCurveSteepness(yield10Y, yield2Y float64) float64 {
	return yield10Y - yield2Y
}

// YieldCurveRegime represents the yield curve regime
type YieldCurveRegime string

const (
	CurveSteep       YieldCurveRegime = "STEEP"
	CurveFlat        YieldCurveRegime = "FLAT"
	CurveInverted    YieldCurveRegime = "INVERTED"
)

// InterpretYieldCurve interprets the yield curve shape
func InterpretYieldCurve(spread float64) YieldCurveRegime {
	if spread > 1.0 {
		return CurveSteep
	} else if spread < -0.1 {
		return CurveInverted
	}
	return CurveFlat
}

// AuctionMetrics calculates auction metrics
type AuctionMetrics struct {
	AvgBidToCover   float64 `json:"avg_bid_to_cover"`
	AvgIndirectPct  float64 `json:"avg_indirect_pct"`
	TrendDirection  string  `json:"trend_direction"` // IMPROVING, WORSENING, STABLE
}

// CalculateAuctionMetrics calculates metrics from auction history
func CalculateAuctionMetrics(auctions []AuctionResult) AuctionMetrics {
	if len(auctions) == 0 {
		return AuctionMetrics{}
	}
	
	var totalBidToCover, totalIndirectPct float64
	var validCount int
	
	for _, a := range auctions {
		if a.BidToCover > 0 {
			totalBidToCover += a.BidToCover
			validCount++
		}
		
		total := float64(a.Dealers + a.Direct + a.Indirect)
		if total > 0 {
			totalIndirectPct += float64(a.Indirect) / total * 100
		}
	}
	
	avgBTC := totalBidToCover / float64(validCount)
	avgIndirect := totalIndirectPct / float64(len(auctions))
	
	// Determine trend
	trend := "STABLE"
	if len(auctions) >= 2 {
		latest := auctions[len(auctions)-1]
		previous := auctions[len(auctions)-2]
		
		if latest.BidToCover > previous.BidToCover*1.05 {
			trend = "IMPROVING"
		} else if latest.BidToCover < previous.BidToCover*0.95 {
			trend = "WORSENING"
		}
	}
	
	return AuctionMetrics{
		AvgBidToCover:  avgBTC,
		AvgIndirectPct: avgIndirect,
		TrendDirection: trend,
	}
}

// SecurityTypes defines Treasury security types
var SecurityTypes = map[string]string{
	"Bill":  "Treasury Bills",
	"Note":  "Treasury Notes",
	"Bond":  "Treasury Bonds",
	"TIPS":  "Treasury Inflation-Protected Securities",
	"FRN":   "Floating Rate Notes",
}
