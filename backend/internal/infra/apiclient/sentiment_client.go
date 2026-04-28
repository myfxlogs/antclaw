package apiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SentimentClient aggregates sentiment data from multiple sources
type SentimentClient struct {
	httpClient *http.Client
	apiKeys    map[string]string
}

// PutCallData represents CBOE put/call ratio data
type PutCallData struct {
	Date      time.Time `json:"date"`
	IndexPCR  float64   `json:"index_pcr"`
	EquityPCR float64   `json:"equity_pcr"`
	TotalPCR  float64   `json:"total_pcr"`
}

// FearGreedData represents CNN Fear & Greed Index
type FearGreedData struct {
	Date      time.Time `json:"date"`
	Value     float64   `json:"value"`
	Sentiment string    `json:"sentiment"` // Extreme Fear, Fear, Neutral, Greed, Extreme Greed
}

// DVOLData represents Deribit Volatility Index
type DVOLData struct {
	Timestamp time.Time `json:"timestamp"`
	DVOL      float64   `json:"dvol"`
	Currency  string    `json:"currency"`
}

// NewSentimentClient creates a new sentiment client
func NewSentimentClient(apiKeys map[string]string) *SentimentClient {
	return &SentimentClient{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		apiKeys:    apiKeys,
	}
}

// FetchPutCallRatio fetches CBOE put/call ratio
func (c *SentimentClient) FetchPutCallRatio(ctx context.Context) (*PutCallData, error) {
	// CBOE data endpoint
	url := "https://cdn.cboe.com/api/global/us_indices/daily_prices/PCRATIO.json"
	
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

	var result struct {
		Data []struct {
			Date      string  `json:"date"`
			IndexPCR  float64 `json:"index_pcr"`
			EquityPCR float64 `json:"equity_pcr"`
			TotalPCR  float64 `json:"total_pcr"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("no data available")
	}

	latest := result.Data[len(result.Data)-1]
	date, _ := time.Parse("2006-01-02", latest.Date)

	return &PutCallData{
		Date:      date,
		IndexPCR:  latest.IndexPCR,
		EquityPCR: latest.EquityPCR,
		TotalPCR:  latest.TotalPCR,
	}, nil
}

// FetchFearGreedIndex fetches CNN Fear & Greed Index
func (c *SentimentClient) FetchFearGreedIndex(ctx context.Context) (*FearGreedData, error) {
	url := "https://production.dataviz.cnn.io/fear-and-greed/index.json"
	
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
		return nil, fmt.Errorf("CNN API returned status %d", resp.StatusCode)
	}

	var result struct {
		FearAndGreed struct {
			Score float64 `json:"score"`
			Text  string  `json:"rating"`
		} `json:"fear_and_greed"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &FearGreedData{
		Date:      time.Now(),
		Value:     result.FearAndGreed.Score,
		Sentiment: result.FearAndGreed.Text,
	}, nil
}

// FetchDVOL fetches Deribit Volatility Index for BTC/ETH
func (c *SentimentClient) FetchDVOL(ctx context.Context, currency string) (*DVOLData, error) {
	// Deribit index endpoint
	url := fmt.Sprintf("https://www.deribit.com/api/v2/public/get_index?currency=%s", currency)
	
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
		return nil, fmt.Errorf("Deribit API returned status %d", resp.StatusCode)
	}

	var result struct {
		Result struct {
			DVOL float64 `json:"volatility_30"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &DVOLData{
		Timestamp: time.Now(),
		DVOL:      result.Result.DVOL,
		Currency:  currency,
	}, nil
}

// InterpretPutCallRatio interprets the put/call ratio
type PCRRegime string

const (
	PCRNeutral  PCRRegime = "NEUTRAL"   // ~1.0
	PCRBearish  PCRRegime = "BEARISH"   // >1.1 (high puts = fear)
	PCRBullish  PCRRegime = "BULLISH"   // <0.9 (high calls = greed)
)

func InterpretPutCallRatio(pcr float64) PCRRegime {
	if pcr > 1.1 {
		return PCRBearish
	} else if pcr < 0.9 {
		return PCRBullish
	}
	return PCRNeutral
}

// InterpretFearGreed interprets Fear & Greed value
func InterpretFearGreed(value float64) string {
	switch {
	case value <= 20:
		return "Extreme Fear"
	case value <= 40:
		return "Fear"
	case value <= 60:
		return "Neutral"
	case value <= 80:
		return "Greed"
	default:
		return "Extreme Greed"
	}
}

// SentimentScore aggregates multiple sentiment indicators
func CalculateSentimentScore(pcr *PutCallData, fearGreed *FearGreedData) float64 {
	score := 50.0 // Neutral base

	if pcr != nil {
		pcrRegime := InterpretPutCallRatio(pcr.TotalPCR)
		switch pcrRegime {
		case PCRBullish:
			score += 10
		case PCRBearish:
			score -= 10
		}
	}

	if fearGreed != nil {
		// Incorporate Fear & Greed directly
		score = (score + fearGreed.Value) / 2
	}

	// Clamp to 0-100
	if score > 100 {
		score = 100
	} else if score < 0 {
		score = 0
	}

	return score
}

// SentimentComposite aggregates all sentiment data
type SentimentComposite struct {
	Timestamp     time.Time `json:"timestamp"`
	PutCallRatio  *PutCallData   `json:"put_call_ratio,omitempty"`
	FearGreed     *FearGreedData `json:"fear_greed,omitempty"`
	DVOL          *DVOLData      `json:"dvol,omitempty"`
	CompositeScore float64       `json:"composite_score"`
}

// FetchComposite fetches all sentiment indicators
func (c *SentimentClient) FetchComposite(ctx context.Context) (*SentimentComposite, error) {
	composite := &SentimentComposite{
		Timestamp: time.Now(),
	}

	// Fetch put/call ratio
	if pcr, err := c.FetchPutCallRatio(ctx); err == nil {
		composite.PutCallRatio = pcr
	}

	// Fetch Fear & Greed
	if fg, err := c.FetchFearGreedIndex(ctx); err == nil {
		composite.FearGreed = fg
	}

	// Fetch DVOL for BTC
	if dvol, err := c.FetchDVOL(ctx, "BTC"); err == nil {
		composite.DVOL = dvol
	}

	composite.CompositeScore = CalculateSentimentScore(composite.PutCallRatio, composite.FearGreed)

	return composite, nil
}
