package ta

import (
	"context"
	"fmt"
	"math"
	"time"

	tav1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
)

// Service implements Technical Analysis business logic.
type Service struct {
	priceCache map[string][]PriceBar
}

// PriceBar represents a price bar for TA calculations.
type PriceBar struct {
	Timestamp time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    int64
}

// NewService creates a new TAService.
func NewService() *Service {
	svc := &Service{
		priceCache: make(map[string][]PriceBar),
	}
	// Generate sample data for major pairs
	svc.generateSampleData("EURUSD", 1.0850, 200)
	svc.generateSampleData("GBPUSD", 1.2650, 200)
	svc.generateSampleData("USDJPY", 150.20, 200)
	return svc
}

func (s *Service) generateSampleData(pair string, basePrice float64, count int) {
	bars := make([]PriceBar, count)
	now := time.Now()
	price := basePrice

	for i := 0; i < count; i++ {
		ts := now.Add(-time.Duration(count-i) * time.Hour)
		// Random walk with slight upward bias
		change := (randFloat() - 0.48) * 0.002 * price
		open := price
		close := price + change
		high := math.Max(open, close) + math.Abs(change)*randFloat()
		low := math.Min(open, close) - math.Abs(change)*randFloat()

		bars[i] = PriceBar{
			Timestamp: ts,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    int64(1000000 + randFloat()*5000000),
		}
		price = close
	}
	s.priceCache[pair] = bars
}

func randFloat() float64 {
	return 0.5
}

// GetIndicators calculates technical indicators.
func (s *Service) GetIndicators(ctx context.Context, pair, timeframe string, indicators []string) (*tav1.GetIndicatorsResponse, error) {
	bars, ok := s.priceCache[pair]
	if !ok {
		return nil, fmt.Errorf("pair not found: %s", pair)
	}

	if len(indicators) == 0 {
		indicators = []string{"RSI", "MACD", "MA20", "MA50", "BB"}
	}

	var values []*tav1.IndicatorValue
	closes := extractCloses(bars)

	for _, ind := range indicators {
		switch ind {
		case "RSI":
			rsi := calculateRSI(closes, 14)
			signal := rsiSignal(rsi)
			values = append(values, &tav1.IndicatorValue{Name: "RSI", Value: rsi, Signal: signal})
		case "MACD":
			macd, signal := calculateMACD(closes)
			macdSignal := macdSignal(macd, signal)
			values = append(values, &tav1.IndicatorValue{Name: "MACD", Value: macd, Signal: macdSignal})
		case "MA20":
			ma := calculateSMA(closes, 20)
			signal := maSignal(closes[len(closes)-1], ma)
			values = append(values, &tav1.IndicatorValue{Name: "MA20", Value: ma, Signal: signal})
		case "MA50":
			ma := calculateSMA(closes, 50)
			signal := maSignal(closes[len(closes)-1], ma)
			values = append(values, &tav1.IndicatorValue{Name: "MA50", Value: ma, Signal: signal})
		case "BB":
			upper, lower := calculateBollingerBands(closes, 20)
			values = append(values, &tav1.IndicatorValue{Name: "BB_Upper", Value: upper, Signal: "neutral"})
			values = append(values, &tav1.IndicatorValue{Name: "BB_Lower", Value: lower, Signal: "neutral"})
		}
	}

	return &tav1.GetIndicatorsResponse{
		Pair:       pair,
		Timeframe:  timeframe,
		Values:     values,
	}, nil
}

// GetElliott performs Elliott Wave analysis.
func (s *Service) GetElliott(ctx context.Context, pair, timeframe string) (*tav1.GetElliottResponse, error) {
	bars, ok := s.priceCache[pair]
	if !ok {
		return nil, fmt.Errorf("pair not found: %s", pair)
	}

	// Simplified Elliott Wave detection
	waves := detectElliottWaves(bars)

	return &tav1.GetElliottResponse{
		Pair:          pair,
		CurrentCount:  fmt.Sprintf("Wave %d of %d", len(waves)%5+1, (len(waves)/5)+1),
		Waves:         waves,
		NextProjection: "Wave " + nextWaveProjection(len(waves)),
	}, nil
}

// GetWyckoff performs Wyckoff analysis.
func (s *Service) GetWyckoff(ctx context.Context, pair string) (*tav1.GetWyckoffResponse, error) {
	bars, ok := s.priceCache[pair]
	if !ok {
		return nil, fmt.Errorf("pair not found: %s", pair)
	}

	phase := detectWyckoffPhase(bars)

	return &tav1.GetWyckoffResponse{
		Pair: pair,
		Phase: &tav1.WyckoffPhase{
			Phase:        phase,
			StartPrice:   bars[0].Close,
			CurrentPrice: bars[len(bars)-1].Close,
			Evidence:     wyckoffEvidence(phase),
		},
		Structure: wyckoffStructure(bars),
	}, nil
}

// GetIct performs ICT/SMC analysis.
func (s *Service) GetIct(ctx context.Context, pair, timeframe string) (*tav1.GetIctResponse, error) {
	bars, ok := s.priceCache[pair]
	if !ok {
		return nil, fmt.Errorf("pair not found: %s", pair)
	}

	levels := detectIctLevels(bars)

	return &tav1.GetIctResponse{
		Pair:   pair,
		Levels: levels,
		Bias:   determineIctBias(levels, bars[len(bars)-1].Close),
	}, nil
}

// GetAmt performs Auction Market Theory analysis.
func (s *Service) GetAmt(ctx context.Context, pair string, lookbackDays int32) (*tav1.GetAmtResponse, error) {
	bars, ok := s.priceCache[pair]
	if !ok {
		return nil, fmt.Errorf("pair not found: %s", pair)
	}

	if lookbackDays == 0 {
		lookbackDays = 30
	}

	zones := calculateVolumeProfile(bars, int(lookbackDays))

	return &tav1.GetAmtResponse{
		Pair:  pair,
		Zones: zones,
		AuctionContext: determineAuctionContext(zones, bars[len(bars)-1].Close),
	}, nil
}

// GetOrderflow performs order flow analysis.
func (s *Service) GetOrderflow(ctx context.Context, pair, timeframe string) (*tav1.GetOrderflowResponse, error) {
	bars, ok := s.priceCache[pair]
	if !ok {
		return nil, fmt.Errorf("pair not found: %s", pair)
	}

	imbalances := detectOrderflowImbalances(bars)

	return &tav1.GetOrderflowResponse{
		Pair:       pair,
		Imbalances: imbalances,
	}, nil
}

// GetVolumeProfile calculates volume profile.
func (s *Service) GetVolumeProfile(ctx context.Context, pair string) (*tav1.GetVolumeProfileResponse, error) {
	bars, ok := s.priceCache[pair]
	if !ok {
		return nil, fmt.Errorf("pair not found: %s", pair)
	}

	zones := calculateVolumeProfile(bars, len(bars))

	var poc, vah, val float64
	profile := make([]*tav1.VpLevel, 0, len(zones))
	for _, z := range zones {
		switch z.Type {
		case "point_of_control":
			poc = z.Price
		case "value_area_high":
			vah = z.Price
		case "value_area_low":
			val = z.Price
		}
		profile = append(profile, &tav1.VpLevel{Price: z.Price, Volume: z.Volume})
	}

	return &tav1.GetVolumeProfileResponse{
		Pair:          pair,
		Poc:           poc,
		ValueAreaHigh: vah,
		ValueAreaLow:  val,
		Profile:       profile,
	}, nil
}

// GetIntermarket performs intermarket analysis.
func (s *Service) GetIntermarket(ctx context.Context, pair string) (*tav1.GetIntermarketResponse, error) {
	// Simplified intermarket correlations
	correlations := []*tav1.Correlation{
		{AssetA: pair, AssetB: "DXY", Correlation: -0.75, Timeframe: "1d"},
		{AssetA: pair, AssetB: "SPY", Correlation: 0.45, Timeframe: "1d"},
		{AssetA: pair, AssetB: "GOLD", Correlation: -0.35, Timeframe: "1d"},
		{AssetA: pair, AssetB: "OIL", Correlation: 0.25, Timeframe: "1d"},
	}

	return &tav1.GetIntermarketResponse{
		Pair:           pair,
		Correlations:   correlations,
		DominantDriver: determineDominantDriver(correlations),
	}, nil
}

func determineDominantDriver(correlations []*tav1.Correlation) string {
	for _, corr := range correlations {
		if math.Abs(corr.Correlation) > 0.5 {
			return corr.AssetB
		}
	}
	return "none"
}

// Technical indicator calculations
func extractCloses(bars []PriceBar) []float64 {
	closes := make([]float64, len(bars))
	for i, bar := range bars {
		closes[i] = bar.Close
	}
	return closes
}

func calculateRSI(closes []float64, period int) float64 {
	if len(closes) < period+1 {
		return 50.0
	}

	var gains, losses float64
	for i := len(closes) - period; i < len(closes); i++ {
		change := closes[i] - closes[i-1]
		if change > 0 {
			gains += change
		} else {
			losses += -change
		}
	}

	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)

	if avgLoss == 0 {
		return 100.0
	}

	rs := avgGain / avgLoss
	rsi := 100.0 - (100.0 / (1.0 + rs))
	return rsi
}

func rsiSignal(rsi float64) string {
	if rsi > 70 {
		return "bearish"
	} else if rsi < 30 {
		return "bullish"
	}
	return "neutral"
}

func calculateMACD(closes []float64) (macd, signal float64) {
	ema12 := calculateEMA(closes, 12)
	ema26 := calculateEMA(closes, 26)
	macd = ema12 - ema26
	signal = calculateEMA(closes[:len(closes)-9], 9)
	return macd, signal
}

func calculateEMA(closes []float64, period int) float64 {
	if len(closes) < period {
		return closes[len(closes)-1]
	}

	multiplier := 2.0 / (float64(period) + 1.0)
	ema := calculateSMA(closes[:period], period)

	for i := period; i < len(closes); i++ {
		ema = (closes[i]-ema)*multiplier + ema
	}
	return ema
}

func macdSignal(macd, signal float64) string {
	if macd > signal {
		return "bullish"
	} else if macd < signal {
		return "bearish"
	}
	return "neutral"
}

func calculateSMA(closes []float64, period int) float64 {
	if len(closes) < period {
		period = len(closes)
	}

	var sum float64
	for i := len(closes) - period; i < len(closes); i++ {
		sum += closes[i]
	}
	return sum / float64(period)
}

func maSignal(price, ma float64) string {
	if price > ma*1.01 {
		return "bullish"
	} else if price < ma*0.99 {
		return "bearish"
	}
	return "neutral"
}

func calculateBollingerBands(closes []float64, period int) (upper, lower float64) {
	sma := calculateSMA(closes, period)
	var sumSq float64
	for i := len(closes) - period; i < len(closes); i++ {
		diff := closes[i] - sma
		sumSq += diff * diff
	}
	stdDev := math.Sqrt(sumSq / float64(period))
	return sma + 2*stdDev, sma - 2*stdDev
}

func detectElliottWaves(bars []PriceBar) []*tav1.ElliottWave {
	var waves []*tav1.ElliottWave
	closes := extractCloses(bars)

	// Simplified wave detection
	if len(closes) < 100 {
		return waves
	}

	waveCount := 1
	startIdx := len(closes) - 100
	startPrice := closes[startIdx]

	for i := startIdx; i < len(closes)-20 && waveCount <= 5; i += 20 {
		endPrice := closes[i+19]
		direction := "up"
		if endPrice < startPrice {
			direction = "down"
		}

		waves = append(waves, &tav1.ElliottWave{
			WaveNumber: int32(waveCount),
			Direction:  direction,
			PriceStart: startPrice,
			PriceEnd:   endPrice,
		})

		waveCount++
		startPrice = endPrice
	}

	return waves
}

func nextWaveProjection(count int) string {
	waves := []string{"1", "2", "3", "4", "5", "A", "B", "C"}
	return waves[count%len(waves)]
}

func detectWyckoffPhase(bars []PriceBar) string {
	if len(bars) < 50 {
		return "unknown"
	}

	recent := bars[len(bars)-50:]
	high := recent[0].High
	low := recent[0].Low

	for _, bar := range recent {
		if bar.High > high {
			high = bar.High
		}
		if bar.Low < low {
			low = bar.Low
		}
	}

	current := recent[len(recent)-1].Close
	range_ := high - low

	if current < low+range_*0.25 {
		return "accumulation"
	} else if current > high-range_*0.25 {
		return "distribution"
	} else if current > low+range_*0.5 {
		return "markup"
	}
	return "markdown"
}

func wyckoffEvidence(phase string) string {
	evidenceMap := map[string]string{
		"accumulation":  "Spring pattern, decreasing volume on declines",
		"markup":        "Higher highs, higher lows, increasing volume",
		"distribution":  "Upthrust, decreasing volume on rallies",
		"markdown":      "Lower highs, lower lows, increasing volume on declines",
	}
	return evidenceMap[phase]
}

func wyckoffStructure(bars []PriceBar) string {
	if len(bars) < 30 {
		return "insufficient_data"
	}

	recent := bars[len(bars)-30:]
	upCount, downCount := 0, 0

	for i := 1; i < len(recent); i++ {
		if recent[i].Close > recent[i-1].Close {
			upCount++
		} else {
			downCount++
		}
	}

	if upCount > downCount*2 {
		return "strong_uptrend"
	} else if downCount > upCount*2 {
		return "strong_downtrend"
	} else if upCount > downCount {
		return "weak_uptrend"
	}
	return "weak_downtrend"
}

func detectIctLevels(bars []PriceBar) []*tav1.IctLevel {
	if len(bars) < 20 {
		return nil
	}

	var levels []*tav1.IctLevel
	recent := bars[len(bars)-20:]

	// Detect order blocks (3+ consecutive bullish/bearish candles)
	for i := 2; i < len(recent)-1; i++ {
		if recent[i-2].Close < recent[i-2].Open &&
			recent[i-1].Close < recent[i-1].Open &&
			recent[i].Close > recent[i].Open {
			// Bullish order block
			levels = append(levels, &tav1.IctLevel{
				Type:      "order_block",
				Price:     recent[i-1].Low,
				Timeframe: "1h",
			})
		}
	}

	// Detect fair value gaps
	for i := 1; i < len(recent)-1; i++ {
		if recent[i+1].Low > recent[i-1].High {
			levels = append(levels, &tav1.IctLevel{
				Type:      "fair_value_gap",
				Price:     (recent[i-1].High + recent[i+1].Low) / 2,
				Timeframe: "1h",
			})
		}
	}

	return levels
}

func determineIctBias(levels []*tav1.IctLevel, currentPrice float64) string {
	if len(levels) == 0 {
		return "neutral"
	}

	bullishLevels, bearishLevels := 0, 0
	for _, level := range levels {
		if level.Type == "order_block" && level.Price < currentPrice {
			bullishLevels++
		} else if level.Price > currentPrice {
			bearishLevels++
		}
	}

	if bullishLevels > bearishLevels {
		return "bullish"
	} else if bearishLevels > bullishLevels {
		return "bearish"
	}
	return "neutral"
}

func calculateVolumeProfile(bars []PriceBar, lookback int) []*tav1.AmtZone {
	if len(bars) < lookback {
		lookback = len(bars)
	}

	recent := bars[len(bars)-lookback:]

	// Find POC (price with highest volume)
	volumeByPrice := make(map[float64]int64)
	for _, bar := range recent {
		price := roundToTick(bar.Close, 0.0001)
		volumeByPrice[price] += bar.Volume
	}

	var pocPrice float64
	var maxVolume int64
	for price, vol := range volumeByPrice {
		if vol > maxVolume {
			maxVolume = vol
			pocPrice = price
		}
	}

	// Calculate value area (70% of volume)
	totalVolume := int64(0)
	for _, vol := range volumeByPrice {
		totalVolume += vol
	}
	targetVolume := int64(float64(totalVolume) * 0.7)

	var vah, val float64
	cumulativeVolume := int64(0)

	// Expand from POC outward
	for step := 1.0; ; step += 0.0001 {
		highPrice := pocPrice + step
		lowPrice := pocPrice - step

		if vol, ok := volumeByPrice[highPrice]; ok {
			cumulativeVolume += vol
			vah = highPrice
		}
		if vol, ok := volumeByPrice[lowPrice]; ok {
			cumulativeVolume += vol
			val = lowPrice
		}

		if cumulativeVolume >= targetVolume {
			break
		}
	}

	return []*tav1.AmtZone{
		{Type: "point_of_control", Price: pocPrice, Volume: float64(maxVolume)},
		{Type: "value_area_high", Price: vah, Volume: float64(volumeByPrice[vah])},
		{Type: "value_area_low", Price: val, Volume: float64(volumeByPrice[val])},
	}
}

func roundToTick(price, tick float64) float64 {
	return math.Round(price/tick) * tick
}

func determineAuctionContext(zones []*tav1.AmtZone, currentPrice float64) string {
	if len(zones) < 3 {
		return "unknown"
	}

	var poc, vah, val float64
	for _, z := range zones {
		switch z.Type {
		case "point_of_control":
			poc = z.Price
		case "value_area_high":
			vah = z.Price
		case "value_area_low":
			val = z.Price
		}
	}

	if currentPrice > vah {
		return "high_volume_breakout"
	} else if currentPrice < val {
		return "low_volume_breakdown"
	} else if math.Abs(currentPrice-poc) < (vah-val)*0.1 {
		return "at_poc"
	}
	return "within_value_area"
}

func detectOrderflowImbalances(bars []PriceBar) []*tav1.Imbalance {
	var imbalances []*tav1.Imbalance

	for i := 1; i < len(bars); i++ {
		// Simplified imbalance detection based on delta
		delta := bars[i].Volume - bars[i-1].Volume
		if math.Abs(float64(delta)) > float64(bars[i-1].Volume)*0.5 {
			imbalanceType := "buy_imbalance"
			if bars[i].Close < bars[i].Open {
				imbalanceType = "sell_imbalance"
			}
			imbalances = append(imbalances, &tav1.Imbalance{
				Type:  imbalanceType,
				Price: bars[i].Close,
				Delta: float64(delta),
			})
		}
	}

	return imbalances
}

