package signals

import (
	"context"
	"fmt"
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
	ctx := context.Background()

	// Get price data
	priceResp, err := p.priceSvc.GetPrice(ctx, pair, "1h", 50)
	if err != nil {
		return nil, fmt.Errorf("provider: price unavailable for %s: %w", pair, err)
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
	vixValue := 18.0 // default fallback
	if vixResp, vixErr := p.volSvc.GetVix(ctx); vixErr == nil && vixResp.Vix != nil {
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

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
