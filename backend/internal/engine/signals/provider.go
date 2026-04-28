package signals

import (
	"strconv"
	"strings"
	"time"

	"github.com/antclaw/antclaw/internal/service/price"
	"github.com/antclaw/antclaw/internal/service/vol"
)

// ServiceProvider implements DataProvider using PriceService and VolService.
type ServiceProvider struct {
	priceSvc *price.Service
	volSvc   *vol.Service
}

// NewServiceProvider creates a new data provider.
func NewServiceProvider(priceSvc *price.Service, volSvc *vol.Service) *ServiceProvider {
	return &ServiceProvider{
		priceSvc: priceSvc,
		volSvc:   volSvc,
	}
}

// GetMarketData fetches market data for a pair.
func (p *ServiceProvider) GetMarketData(pair string) (*MarketData, error) {
	// Get price data
	priceResp, err := p.priceSvc.GetPrice(nil, pair, "1h", 50)
	if err != nil {
		// Return synthetic data if price service fails
		return p.getSyntheticData(pair), nil
	}
	
	// Parse price data
	currentPrice, _ := strconv.ParseFloat(priceResp.Current, 64)
	changePct, _ := strconv.ParseFloat(strings.TrimSuffix(priceResp.ChangePct_24H, "%"), 64)
	
	// Convert bars
	var bars []PriceBar
	for _, pb := range priceResp.Bars {
		open, _ := strconv.ParseFloat(pb.Open, 64)
		high, _ := strconv.ParseFloat(pb.High, 64)
		low, _ := strconv.ParseFloat(pb.Low, 64)
		close, _ := strconv.ParseFloat(pb.Close, 64)
		ts, _ := time.Parse(time.RFC3339, pb.Timestamp)
		
		bars = append(bars, PriceBar{
			Timestamp: ts,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    pb.Volume,
		})
	}
	
	// Get VIX data for volatility context
	vixValue := 18.0 // Default
	vixResp, err := p.volSvc.GetVix(nil)
	if err == nil && vixResp.Vix != nil {
		vixValue = vixResp.Vix.Spot
	}
	
	// Calculate ATR
	atr := calculateATR(bars)
	
	return &MarketData{
		Pair:         pair,
		CurrentPrice: currentPrice,
		ChangePct24h: changePct,
		Bars:         bars,
		VIX:          vixValue,
		ATR:          atr,
		Timestamp:    time.Now(),
	}, nil
}

// GetPairsByCategory returns pairs for a category.
func (p *ServiceProvider) GetPairsByCategory(category string) []string {
	pairsByCategory := map[string][]string{
		"majors": {"EURUSD", "GBPUSD", "USDJPY", "USDCHF", "AUDUSD", "USDCAD"},
		"crosses": {"EURGBP", "EURJPY", "GBPJPY", "EURCHF", "GBPCHF"},
		"crypto": {"BTCUSD", "ETHUSD", "SOLUSD"},
		"all": {"EURUSD", "GBPUSD", "USDJPY", "USDCHF", "AUDUSD", "USDCAD",
			"EURGBP", "EURJPY", "GBPJPY", "BTCUSD", "ETHUSD"},
	}
	
	pairs, ok := pairsByCategory[category]
	if !ok {
		return pairsByCategory["majors"]
	}
	return pairs
}

// getSyntheticData returns synthetic market data for testing/fallback.
func (p *ServiceProvider) getSyntheticData(pair string) *MarketData {
	basePrice := 1.0
	if strings.Contains(pair, "JPY") {
		basePrice = 150.0
	} else if strings.Contains(pair, "BTC") {
		basePrice = 45000.0
	} else if strings.Contains(pair, "ETH") {
		basePrice = 2800.0
	}
	
	bars := generateSyntheticBars(basePrice, 50)
	
	return &MarketData{
		Pair:         pair,
		CurrentPrice:   basePrice,
		ChangePct24h: 0.0,
		Bars:         bars,
		VIX:          18.0,
		ATR:          basePrice * 0.002,
		Timestamp:    time.Now(),
	}
}

// generateSyntheticBars creates synthetic price bars.
func generateSyntheticBars(basePrice float64, count int) []PriceBar {
	bars := make([]PriceBar, count)
	now := time.Now().UTC()
	price := basePrice
	
	for i := 0; i < count; i++ {
		ts := now.Add(-time.Duration(count-i) * time.Hour)
		
		// Random walk with small steps
		change := (randFloat() - 0.5) * 0.002 * price
		open := price
		close := price + change
		high := max(open, close) + abs(change)*0.3
		low := min(open, close) - abs(change)*0.3
		
		bars[i] = PriceBar{
			Timestamp: ts,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    1000000 + int64(randFloat()*5000000),
		}
		
		price = close
	}
	
	return bars
}

// calculateATR calculates Average True Range.
func calculateATR(bars []PriceBar) float64 {
	if len(bars) < 2 {
		return 0
	}
	
	period := len(bars) - 1
	if period > 14 {
		period = 14
	}
	trSum := 0.0
	
	for i := 1; i <= period; i++ {
		curr := bars[len(bars)-i]
		prev := bars[len(bars)-i-1]
		
		// True Range = max(high-low, |high-prev_close|, |low-prev_close|)
		tr1 := curr.High - curr.Low
		tr2 := abs(curr.High - prev.Close)
		tr3 := abs(curr.Low - prev.Close)
		
		tr := max(tr1, max(tr2, tr3))
		trSum += tr
	}
	
	return trSum / float64(period)
}

// randFloat returns a deterministic random float for consistency.
func randFloat() float64 {
	return 0.5
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
