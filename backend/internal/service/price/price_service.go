package price

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/antclaw/antclaw/internal/infra/postgres"
)

// PriceService provides price data operations
type PriceService struct {
	repo      postgres.PriceRepository
	logger    *slog.Logger
}

// PriceContext represents price context with technical indicators
type PriceContext struct {
	Symbol      string                 `json:"symbol"`
	LastPrice   float64                `json:"last_price"`
	Change24h   float64                `json:"change_24h"`
	Volatility  float64                `json:"volatility"`
	ATR         float64                `json:"atr"`
	Trend       string                 `json:"trend"`
	Support     []float64              `json:"support"`
	Resistance  []float64              `json:"resistance"`
	Timestamp   time.Time              `json:"timestamp"`
}

// NewPriceService creates a new price service
func NewPriceService(repo postgres.PriceRepository, logger *slog.Logger) *PriceService {
	return &PriceService{
		repo:   repo,
		logger: logger,
	}
}

// GetPriceContext returns comprehensive price context
func (s *PriceService) GetPriceContext(ctx context.Context, symbol string) (*PriceContext, error) {
	// Get recent price data
	bars, err := s.GetRecentDailyBars(ctx, symbol, 100)
	if err != nil {
		return nil, err
	}

	if len(bars) < 20 {
		return nil, fmt.Errorf("insufficient price data for %s", symbol)
	}

	context := &PriceContext{
		Symbol:     symbol,
		LastPrice:  bars[len(bars)-1].Close,
		Timestamp:  time.Now(),
	}

	// Calculate 24h change
	if len(bars) >= 2 {
		context.Change24h = (bars[len(bars)-1].Close - bars[len(bars)-2].Close) / bars[len(bars)-2].Close * 100
	}

	// Calculate volatility
	context.Volatility = s.calculateVolatility(bars, 20)

	// Calculate ATR
	context.ATR = s.calculateATR(bars, 14)

	// Determine trend
	context.Trend = s.determineTrend(bars)

	// Find support/resistance
	context.Support, context.Resistance = s.findSupportResistance(bars, 20)

	return context, nil
}

// GetRecentDailyBars fetches recent daily bars
func (s *PriceService) GetRecentDailyBars(ctx context.Context, symbol string, days int) ([]postgres.DailyBar, error) {
	to := time.Now()
	from := to.AddDate(0, 0, -days)
	
	return s.repo.GetDailyBars(ctx, symbol, from, to)
}

// GetLatestPrice returns latest price for a symbol
func (s *PriceService) GetLatestPrice(ctx context.Context, symbol string) (float64, error) {
	bar, err := s.repo.GetLatest(ctx, symbol)
	if err != nil {
		return 0, err
	}
	if bar == nil {
		return 0, fmt.Errorf("no price data for %s", symbol)
	}
	return bar.Close, nil
}

// calculateVolatility calculates historical volatility
func (s *PriceService) calculateVolatility(bars []postgres.DailyBar, period int) float64 {
	if len(bars) < period+1 {
		return 0
	}

	// Calculate daily returns
	var returns []float64
	for i := 1; i <= period; i++ {
		idx := len(bars) - period + i - 1
		if bars[idx-1].Close == 0 {
			continue
		}
		ret := math.Log(bars[idx].Close / bars[idx-1].Close)
		returns = append(returns, ret)
	}

	if len(returns) == 0 {
		return 0
	}

	// Calculate standard deviation
	mean := s.mean(returns)
	variance := s.variance(returns, mean)
	stdDev := math.Sqrt(variance)

	// Annualize volatility
	return stdDev * math.Sqrt(252) * 100 // As percentage
}

// calculateATR calculates Average True Range
func (s *PriceService) calculateATR(bars []postgres.DailyBar, period int) float64 {
	if len(bars) < period+1 {
		return 0
	}

	var trValues []float64
	for i := len(bars) - period; i < len(bars); i++ {
		high := bars[i].High
		low := bars[i].Low
		prevClose := bars[i-1].Close

		tr1 := high - low
		tr2 := math.Abs(high - prevClose)
		tr3 := math.Abs(low - prevClose)

		tr := math.Max(tr1, math.Max(tr2, tr3))
		trValues = append(trValues, tr)
	}

	return s.mean(trValues)
}

// determineTrend determines price trend
func (s *PriceService) determineTrend(bars []postgres.DailyBar) string {
	if len(bars) < 50 {
		return "UNKNOWN"
	}

	// Calculate moving averages
	sma20 := s.calculateSMA(bars, 20)
	sma50 := s.calculateSMA(bars, 50)

	currentPrice := bars[len(bars)-1].Close

	// Trend determination
	if currentPrice > sma20 && sma20 > sma50 {
		return "BULLISH"
	} else if currentPrice < sma20 && sma20 < sma50 {
		return "BEARISH"
	}
	return "NEUTRAL"
}

// calculateSMA calculates Simple Moving Average
func (s *PriceService) calculateSMA(bars []postgres.DailyBar, period int) float64 {
	if len(bars) < period {
		return 0
	}

	var sum float64
	for i := len(bars) - period; i < len(bars); i++ {
		sum += bars[i].Close
	}
	return sum / float64(period)
}

// findSupportResistance finds support and resistance levels
func (s *PriceService) findSupportResistance(bars []postgres.DailyBar, lookback int) ([]float64, []float64) {
	if len(bars) < lookback {
		return nil, nil
	}

	recent := bars[len(bars)-lookback:]

	var highs, lows []float64
	for _, bar := range recent {
		highs = append(highs, bar.High)
		lows = append(lows, bar.Low)
	}

	// Find local maxima and minima
	resistance := s.findLocalExtrema(highs, true)
	support := s.findLocalExtrema(lows, false)

	return support, resistance
}

// findLocalExtrema finds local extrema in price data
func (s *PriceService) findLocalExtrema(prices []float64, findMax bool) []float64 {
	if len(prices) < 5 {
		return prices
	}

	var extrema []float64
	window := 2

	for i := window; i < len(prices)-window; i++ {
		isExtremum := true
		for j := 1; j <= window; j++ {
			if findMax {
				if prices[i] <= prices[i-j] || prices[i] <= prices[i+j] {
					isExtremum = false
					break
				}
			} else {
				if prices[i] >= prices[i-j] || prices[i] >= prices[i+j] {
					isExtremum = false
					break
				}
			}
		}
		if isExtremum {
			extrema = append(extrema, prices[i])
		}
	}

	// Limit to top 3 levels
	if len(extrema) > 3 {
		extrema = extrema[:3]
	}

	return extrema
}

// mean calculates arithmetic mean
func (s *PriceService) mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// variance calculates variance
func (s *PriceService) variance(values []float64, mean float64) float64 {
	if len(values) == 0 {
		return 0
	}

	var sumSquared float64
	for _, v := range values {
		diff := v - mean
		sumSquared += diff * diff
	}
	return sumSquared / float64(len(values))
}

// UpsertBars inserts or updates price bars
func (s *PriceService) UpsertBars(ctx context.Context, bars []postgres.DailyBar) error {
	return s.repo.UpsertDailyBars(ctx, bars)
}

// GetMultiContext fetches price context for multiple symbols
func (s *PriceService) GetMultiContext(ctx context.Context, symbols []string) (map[string]*PriceContext, error) {
	results := make(map[string]*PriceContext)
	
	for _, symbol := range symbols {
		ctx, err := s.GetPriceContext(ctx, symbol)
		if err != nil {
			s.logger.Warn("failed to get price context", "symbol", symbol, "error", err)
			continue
		}
		results[symbol] = ctx
	}
	
	return results, nil
}
