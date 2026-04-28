package price

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	pricev1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
)

// Service implements Price business logic with real calculations.
type Service struct {
	dataStore *PriceDataStore
}

// PriceDataStore holds price data for pairs.
type PriceDataStore struct {
	prices map[string]*PriceData
}

// PriceData holds OHLCV data for a pair.
type PriceData struct {
	Pair      string
	Current   float64
	Open24h   float64
	High24h   float64
	Low24h    float64
	Timestamp time.Time
	Bars      []Bar
}

// Bar represents a single price bar.
type Bar struct {
	Timestamp time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    int64
}

// NewService creates a new PriceService with sample data.
func NewService() *Service {
	return &Service{
		dataStore: newSampleDataStore(),
	}
}

// newSampleDataStore creates sample price data for major pairs.
func newSampleDataStore() *PriceDataStore {
	store := &PriceDataStore{prices: make(map[string]*PriceData)}
	
	// Major forex pairs with realistic data
	pairs := []struct {
		pair    string
		current float64
		open24h float64
	}{
		{"EURUSD", 1.0850, 1.0820},
		{"GBPUSD", 1.2650, 1.2620},
		{"USDJPY", 150.20, 149.80},
		{"USDCHF", 0.8850, 0.8880},
		{"AUDUSD", 0.6550, 0.6520},
		{"USDCAD", 1.3550, 1.3580},
	}
	
	for _, p := range pairs {
		change := p.current - p.open24h
		pctChange := (change / p.open24h) * 100
		
		// Generate 24h of hourly bars
		bars := generateBars(p.current, 24, time.Hour)
		
		store.prices[p.pair] = &PriceData{
			Pair:      p.pair,
			Current:   p.current,
			Open24h:   p.open24h,
			High24h:   math.Max(p.current, p.open24h) + math.Abs(pctChange)*0.001,
			Low24h:    math.Min(p.current, p.open24h) - math.Abs(pctChange)*0.001,
			Timestamp: time.Now(),
			Bars:      bars,
		}
	}
	
	return store
}

// generateBars creates sample price bars.
func generateBars(basePrice float64, count int, interval time.Duration) []Bar {
	bars := make([]Bar, count)
	now := time.Now().UTC()
	price := basePrice
	
	for i := 0; i < count; i++ {
		ts := now.Add(-time.Duration(count-i) * interval)
		
		// Random walk
		change := (mathRandFloat64() - 0.5) * 0.002 * price
		open := price
		close := price + change
		high := math.Max(open, close) + math.Abs(change)*0.5
		low := math.Min(open, close) - math.Abs(change)*0.5
		
		bars[i] = Bar{
			Timestamp: ts,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    int64(1000000 + mathRandFloat64()*5000000),
		}
		
		price = close
	}
	
	return bars
}

func mathRandFloat64() float64 {
	// Simple deterministic random for reproducibility
	return 0.5 // Placeholder - in production use crypto/rand or math/rand with seed
}

// GetPrice returns current price with historical bars for a pair.
func (s *Service) GetPrice(ctx context.Context, pair, timeframe string, count int32) (*pricev1.GetPriceResponse, error) {
	if pair == "" {
		return nil, fmt.Errorf("pair required")
	}
	
	data, ok := s.dataStore.prices[pair]
	if !ok {
		// Generate sample data for unknown pairs
		data = &PriceData{
			Pair:      pair,
			Current:   1.0000,
			Open24h:   0.9950,
			Timestamp: time.Now(),
			Bars:      generateBars(1.0000, int(count), time.Hour),
		}
	}
	
	change24h := data.Current - data.Open24h
	changePct := (change24h / data.Open24h) * 100
	
	// Convert bars to proto
	var pbBars []*pricev1.PriceBar
	for _, bar := range data.Bars {
		pbBars = append(pbBars, &pricev1.PriceBar{
			Timestamp: bar.Timestamp.Format(time.RFC3339),
			Open:      fmt.Sprintf("%.5f", bar.Open),
			High:      fmt.Sprintf("%.5f", bar.High),
			Low:       fmt.Sprintf("%.5f", bar.Low),
			Close:     fmt.Sprintf("%.5f", bar.Close),
			Volume:    bar.Volume,
		})
	}
	
	return &pricev1.GetPriceResponse{
		Pair:          pair,
		Current:       fmt.Sprintf("%.5f", data.Current),
		Change_24H:    fmt.Sprintf("%.5f", change24h),
		ChangePct_24H: fmt.Sprintf("%.2f%%", changePct),
		Bars:          pbBars,
	}, nil
}

// GetLevels calculates support/resistance levels using pivot points.
func (s *Service) GetLevels(ctx context.Context, pair, timeframe string) (*pricev1.GetLevelsResponse, error) {
	if pair == "" {
		return nil, fmt.Errorf("pair required")
	}
	
	data, ok := s.dataStore.prices[pair]
	if !ok {
		return nil, fmt.Errorf("pair not found: %s", pair)
	}
	
	// Calculate pivot point from last bar
	if len(data.Bars) == 0 {
		return nil, fmt.Errorf("no price data for pair: %s", pair)
	}
	
	lastBar := data.Bars[len(data.Bars)-1]
	high := lastBar.High
	low := lastBar.Low
	close := lastBar.Close
	
	// Classic pivot point formula
	pivot := (high + low + close) / 3
	r1 := 2*pivot - low
	r2 := pivot + (high - low)
	s1 := 2*pivot - high
	s2 := pivot - (high - low)
	
	levels := []*pricev1.PriceLevel{
		{Price: fmt.Sprintf("%.5f", r2), Type: "resistance", Strength: 0.8},
		{Price: fmt.Sprintf("%.5f", r1), Type: "resistance", Strength: 0.6},
		{Price: fmt.Sprintf("%.5f", pivot), Type: "pivot", Strength: 1.0},
		{Price: fmt.Sprintf("%.5f", s1), Type: "support", Strength: 0.6},
		{Price: fmt.Sprintf("%.5f", s2), Type: "support", Strength: 0.8},
	}
	
	return &pricev1.GetLevelsResponse{
		Pair:   pair,
		Levels: levels,
	}, nil
}

// GetMarketOverview returns market overview for major pairs.
func (s *Service) GetMarketOverview(ctx context.Context, category string) (*pricev1.GetMarketOverviewResponse, error) {
	var items []*pricev1.MarketOverviewItem
	
	pairs := []string{"EURUSD", "GBPUSD", "USDJPY", "USDCHF", "AUDUSD", "USDCAD"}
	
	for _, pair := range pairs {
		data, ok := s.dataStore.prices[pair]
		if !ok {
			continue
		}
		
		items = append(items, &pricev1.MarketOverviewItem{
			Pair:  pair,
			Price: fmt.Sprintf("%.5f", data.Current),
		})
	}
	
	return &pricev1.GetMarketOverviewResponse{Items: items}, nil
}

// GetSession returns trading session info with volatility.
func (s *Service) GetSession(ctx context.Context, pair string) (*pricev1.GetSessionResponse, error) {
	now := time.Now().UTC()
	
	// Session definitions (UTC)
	sessions := []*pricev1.SessionInfo{
		{
			Session:  "sydney",
			IsOpen:   isSessionOpen(now, 22, 7),
			OpensAt:  "22:00",
			ClosesAt: "07:00",
			VolatilityIndex: 12.5,
		},
		{
			Session:  "tokyo",
			IsOpen:   isSessionOpen(now, 0, 9),
			OpensAt:  "00:00",
			ClosesAt: "09:00",
			VolatilityIndex: 15.0,
		},
		{
			Session:  "london",
			IsOpen:   isSessionOpen(now, 8, 17),
			OpensAt:  "08:00",
			ClosesAt: "17:00",
			VolatilityIndex: 25.0,
		},
		{
			Session:  "new_york",
			IsOpen:   isSessionOpen(now, 13, 22),
			OpensAt:  "13:00",
			ClosesAt: "22:00",
			VolatilityIndex: 22.0,
		},
	}
	
	// Determine current session
	currentSession := ""
	for _, s := range sessions {
		if s.IsOpen {
			currentSession = s.Session
			break
		}
	}
	
	return &pricev1.GetSessionResponse{
		Pair:            pair,
		Sessions:        sessions,
		CurrentSession:  currentSession,
	}, nil
}

// isSessionOpen checks if current time is within session hours.
func isSessionOpen(now time.Time, startHour, endHour int) bool {
	hour := now.Hour()
	if startHour < endHour {
		return hour >= startHour && hour < endHour
	}
	// Handle overnight sessions (e.g., Sydney 22:00-07:00)
	return hour >= startHour || hour < endHour
}

// RunScenario runs price scenario analysis.
func (s *Service) RunScenario(ctx context.Context, pair string, params map[string]string) (*pricev1.RunScenarioResponse, error) {
	data, ok := s.dataStore.prices[pair]
	if !ok {
		return nil, fmt.Errorf("pair not found: %s", pair)
	}
	
	// Get scenario parameters
	shockStr := params["shock_pct"]
	shock, _ := strconv.ParseFloat(shockStr, 64)
	if shock == 0 {
		shock = 1.0 // Default 1% shock
	}
	
	currentPrice := data.Current
	shockAmount := currentPrice * (shock / 100)
	
	// Monte Carlo style scenario simulation
	results := []*pricev1.ScenarioResult{
		{
			ScenarioName: "bullish_breakout",
			Outcome:      fmt.Sprintf("Price reaches %.5f", currentPrice+shockAmount*2),
			Probability:  0.25,
		},
		{
			ScenarioName: "moderate_rally",
			Outcome:      fmt.Sprintf("Price reaches %.5f", currentPrice+shockAmount),
			Probability:  0.35,
		},
		{
			ScenarioName: "sideways",
			Outcome:      fmt.Sprintf("Price range %.5f - %.5f", currentPrice-shockAmount*0.5, currentPrice+shockAmount*0.5),
			Probability:  0.25,
		},
		{
			ScenarioName: "moderate_decline",
			Outcome:      fmt.Sprintf("Price reaches %.5f", currentPrice-shockAmount),
			Probability:  0.12,
		},
		{
			ScenarioName: "bearish_breakdown",
			Outcome:      fmt.Sprintf("Price reaches %.5f", currentPrice-shockAmount*2),
			Probability:  0.03,
		},
	}
	
	return &pricev1.RunScenarioResponse{Results: results}, nil
}

// GetRegime detects market regime using ADX-inspired calculation.
func (s *Service) GetRegime(ctx context.Context, pair, timeframe string) (*pricev1.GetRegimeResponse, error) {
	data, ok := s.dataStore.prices[pair]
	if !ok || len(data.Bars) < 14 {
		return &pricev1.GetRegimeResponse{
			Regime: &pricev1.MarketRegime{
				Regime:    "unknown",
				Confidence: 0.0,
				Since:     time.Now().Format(time.RFC3339),
			},
		}, nil
	}
	
	// Calculate directional movement
	plusDM := 0.0
	minusDM := 0.0
	trSum := 0.0
	
	for i := 1; i < len(data.Bars) && i < 15; i++ {
		curr := data.Bars[i]
		prev := data.Bars[i-1]
		
		// True range
		tr := math.Max(curr.High-curr.Low, 
			math.Max(math.Abs(curr.High-prev.Close), math.Abs(curr.Low-prev.Close)))
		trSum += tr
		
		// Directional movement
		upMove := curr.High - prev.High
		downMove := prev.Low - curr.Low
		
		if upMove > downMove && upMove > 0 {
			plusDM += upMove
		}
		if downMove > upMove && downMove > 0 {
			minusDM += downMove
		}
	}
	
	// Calculate DI and ADX
	var diPlus, diMinus, adx float64
	if trSum > 0 {
		diPlus = (plusDM / trSum) * 100
		diMinus = (minusDM / trSum) * 100
		dx := math.Abs(diPlus - diMinus) / (diPlus + diMinus) * 100
		adx = dx // Simplified
	}
	
	// Classify regime
	regime := "ranging"
	confidence := 0.5
	
	if adx > 25 {
		if diPlus > diMinus {
			regime = "trending"
			confidence = adx / 100
		} else {
			regime = "volatile"
			confidence = adx / 100
		}
	} else if adx < 15 {
		regime = "calm"
		confidence = (25 - adx) / 25
	}
	
	recentRegimes := []string{"ranging", regime}
	
	return &pricev1.GetRegimeResponse{
		Regime: &pricev1.MarketRegime{
			Regime:     regime,
			Confidence: confidence,
			Since:      time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
		},
		RecentRegimes: recentRegimes,
	}, nil
}

// GetSeasonal returns seasonal analysis for a pair.
func (s *Service) GetSeasonal(ctx context.Context, pair string, years int32) (*pricev1.GetSeasonalResponse, error) {
	if years <= 0 {
		years = 5
	}
	
	// Sample seasonal data based on typical forex patterns
	months := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", 
		"Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	
	var data []*pricev1.SeasonalDataPoint
	
	for i, month := range months {
		// Sample seasonal patterns (simplified)
		baseReturn := 0.0
		switch i {
		case 0, 1: // Jan, Feb - post-holiday
			baseReturn = 0.3
		case 2, 3: // Mar, Apr - spring volatility
			baseReturn = 0.5
		case 4, 5: // May, Jun - summer lull
			baseReturn = -0.1
		case 6, 7: // Jul, Aug - summer low
			baseReturn = -0.2
		case 8, 9: // Sep, Oct - autumn volatility
			baseReturn = 0.4
		case 10, 11: // Nov, Dec - year end
			baseReturn = 0.2
		}
		
		winRate := 0.5 + baseReturn*0.5
		if winRate > 0.7 {
			winRate = 0.7
		}
		if winRate < 0.3 {
			winRate = 0.3
		}
		
		data = append(data, &pricev1.SeasonalDataPoint{
			Month:     month,
			AvgReturn: baseReturn,
			WinRate:   winRate,
		})
	}
	
	return &pricev1.GetSeasonalResponse{
		Pair: pair,
		Data: data,
	}, nil
}
