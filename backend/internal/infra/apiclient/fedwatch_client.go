package apiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// FedWatchClient fetches CME FedWatch probability data
type FedWatchClient struct {
	httpClient *http.Client
}

// FedWatchData represents Fed funds futures implied probabilities
type FedWatchData struct {
	MeetingDate   time.Time            `json:"meeting_date"`
	Probabilities map[string]float64   `json:"probabilities"` // Rate -> Probability%
	TargetRate    string               `json:"target_rate"`   // Most likely rate
	Confidence    float64              `json:"confidence"`    // Probability of target
	FetchedAt     time.Time            `json:"fetched_at"`
}

// MeetingProbability represents probability for a specific meeting
type MeetingProbability struct {
	MeetingDate string             `json:"meeting_date"`
	Scenarios   []RateScenario     `json:"scenarios"`
}

// RateScenario represents a rate scenario and its probability
type RateScenario struct {
	Rate        string  `json:"rate"`
	Probability float64 `json:"probability"`
}

// NewFedWatchClient creates a new FedWatch client
func NewFedWatchClient() *FedWatchClient {
	return &FedWatchClient{
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// FetchProbabilities fetches Fed funds futures implied probabilities
// Uses CME data feed or scraping approach
func (c *FedWatchClient) FetchProbabilities(ctx context.Context) ([]FedWatchData, error) {
	// CME FedWatch data endpoint
	url := "https://www.cmegroup.com/CmeWS/mvc/Quotes/Future/446/G?pageSize=500"
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CME API returned status %d", resp.StatusCode)
	}

	var result struct {
		Quotes []struct {
			ExpDate       string  `json:"expirationDate"`
			Last          string  `json:"last"`
			Change        float64 `json:"change"`
			Probabilities []struct {
				Scenario    string  `json:"scenario"`
				Probability float64 `json:"probability"`
			} `json:"probabilities"`
		} `json:"quotes"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var data []FedWatchData
	now := time.Now()
	
	for _, q := range result.Quotes {
		meetingDate, err := time.Parse("20060102", q.ExpDate)
		if err != nil {
			continue
		}
		
		fw := FedWatchData{
			MeetingDate:   meetingDate,
			Probabilities: make(map[string]float64),
			FetchedAt:     now,
		}
		
		maxProb := 0.0
		for _, p := range q.Probabilities {
			fw.Probabilities[p.Scenario] = p.Probability
			if p.Probability > maxProb {
				maxProb = p.Probability
				fw.TargetRate = p.Scenario
				fw.Confidence = p.Probability
			}
		}
		
		data = append(data, fw)
	}

	return data, nil
}

// ParseFedRate parses Fed funds rate from string like "5.25-5.50%"
func ParseFedRate(s string) (float64, error) {
	s = strings.TrimSuffix(s, "%")
	s = strings.TrimSpace(s)
	
	// Handle range like "5.25-5.50"
	if strings.Contains(s, "-") {
		parts := strings.Split(s, "-")
		if len(parts) == 2 {
			mid, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
			if err != nil {
				return 0, err
			}
			high, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
			if err != nil {
				return 0, err
			}
			return (mid + high) / 2, nil
		}
	}
	
	return strconv.ParseFloat(s, 64)
}

// GetHikeProbability gets probability of any rate hike at a meeting
func (c *FedWatchClient) GetHikeProbability(ctx context.Context, meetingDate time.Time) (float64, error) {
	data, err := c.FetchProbabilities(ctx)
	if err != nil {
		return 0, err
	}
	
	for _, d := range data {
		if d.MeetingDate.Equal(meetingDate) || 
		   (d.MeetingDate.Year() == meetingDate.Year() && 
		    d.MeetingDate.Month() == meetingDate.Month()) {
			hikeProb := 0.0
			for rate, prob := range d.Probabilities {
				if r, err := ParseFedRate(rate); err == nil && r > 5.0 { // Current approximate rate
					hikeProb += prob
				}
			}
			return hikeProb, nil
		}
	}
	
	return 0, fmt.Errorf("no data for meeting date %v", meetingDate)
}

// GetNextMeetingProbability gets probability for next FOMC meeting
func (c *FedWatchClient) GetNextMeetingProbability(ctx context.Context) (*FedWatchData, error) {
	data, err := c.FetchProbabilities(ctx)
	if err != nil {
		return nil, err
	}
	
	now := time.Now()
	for _, d := range data {
		if d.MeetingDate.After(now) {
			return &d, nil
		}
	}
	
	return nil, fmt.Errorf("no future meeting data available")
}

// CalculateRatePath calculates implied rate path from probabilities
func CalculateRatePath(data []FedWatchData) []struct {
	Date       time.Time `json:"date"`
	ImpliedRate float64  `json:"implied_rate"`
} {
	var path []struct {
		Date       time.Time `json:"date"`
		ImpliedRate float64  `json:"implied_rate"`
	}
	
	for _, d := range data {
		var weightedRate float64
		var totalProb float64
		
		for rateStr, prob := range d.Probabilities {
			if rate, err := ParseFedRate(rateStr); err == nil {
				weightedRate += rate * prob
				totalProb += prob
			}
		}
		
		if totalProb > 0 {
			path = append(path, struct {
				Date       time.Time `json:"date"`
				ImpliedRate float64  `json:"implied_rate"`
			}{
				Date:        d.MeetingDate,
				ImpliedRate: weightedRate / totalProb,
			})
		}
	}
	
	return path
}

// FOMCMeetings returns scheduled FOMC meeting dates
func FOMCMeetings(year int) []time.Time {
	// FOMC meets 8 times per year, typically on:
	// Jan/Feb, Mar, Apr/May, Jun, Jul, Sep, Nov, Dec
	meetings := []time.Time{
		time.Date(year, 1, 29, 0, 0, 0, 0, time.UTC),
		time.Date(year, 3, 19, 0, 0, 0, 0, time.UTC),
		time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC),
		time.Date(year, 6, 12, 0, 0, 0, 0, time.UTC),
		time.Date(year, 7, 31, 0, 0, 0, 0, time.UTC),
		time.Date(year, 9, 18, 0, 0, 0, 0, time.UTC),
		time.Date(year, 11, 7, 0, 0, 0, 0, time.UTC),
		time.Date(year, 12, 18, 0, 0, 0, 0, time.UTC),
	}
	return meetings
}
