package orderflow

import (
	"math"
	"time"
)

// Engine analyzes order flow data
type Engine struct{}

// NewEngine creates order flow engine
func NewEngine() *Engine {
	return &Engine{}
}

// OrderFlowSignal represents order flow analysis
type OrderFlowSignal struct {
	Timestamp      time.Time
	Symbol         string
	CumDelta       float64
	DeltaTrend     string
	Absorption     []AbsorptionEvent
	POC            float64
	ValueAreaHigh  float64
	ValueAreaLow   float64
	Recommendation string
}

// AbsorptionEvent represents absorption detection
type AbsorptionEvent struct {
	Price    float64
	Strength float64
	Type     string // "ACCUMULATION" or "DISTRIBUTION"
}

// Analyze performs order flow analysis
func (e *Engine) Analyze(symbol string, ticks []Tick) (*OrderFlowSignal, error) {
	signal := &OrderFlowSignal{
		Timestamp: time.Now(),
		Symbol:    symbol,
	}

	// Calculate delta
	signal.CumDelta = e.calculateDelta(ticks)
	signal.DeltaTrend = e.determineDeltaTrend(ticks)

	// Detect absorption
	signal.Absorption = e.detectAbsorption(ticks)

	// Calculate volume profile
	profile := e.calculateVolumeProfile(ticks)
	signal.POC = profile.POC
	signal.ValueAreaHigh = profile.VAH
	signal.ValueAreaLow = profile.VAL

	signal.Recommendation = e.generateRecommendation(signal)
	return signal, nil
}

// Tick represents a single trade
type Tick struct {
	Price     float64
	Volume    float64
	IsBuy     bool
	Timestamp time.Time
}

func (e *Engine) calculateDelta(ticks []Tick) float64 {
	var delta float64
	for _, tick := range ticks {
		if tick.IsBuy {
			delta += tick.Volume
		} else {
			delta -= tick.Volume
		}
	}
	return delta
}

func (e *Engine) determineDeltaTrend(ticks []Tick) string {
	if len(ticks) < 20 {
		return "NEUTRAL"
	}

	half := len(ticks) / 2
	delta1 := e.calculateDelta(ticks[:half])
	delta2 := e.calculateDelta(ticks[half:])

	if delta2 > delta1*1.2 {
		return "INCREASING_BUYING"
	} else if delta2 < delta1*0.8 {
		return "INCREASING_SELLING"
	}
	return "NEUTRAL"
}

func (e *Engine) detectAbsorption(ticks []Tick) []AbsorptionEvent {
	var events []AbsorptionEvent

	// Group ticks by price level
	volumeByPrice := make(map[float64]float64)
	for _, tick := range ticks {
		// Round price to nearest whole number for grouping
		price := math.Round(tick.Price)
		volumeByPrice[price] += tick.Volume
	}

	// Find price levels with high volume but no movement
	for price, vol := range volumeByPrice {
		avgVol := e.averageVolume(volumeByPrice)
		if vol > avgVol*3 { // 3x average volume
			events = append(events, AbsorptionEvent{
				Price:    price,
				Strength: vol / avgVol,
				Type:     "ACCUMULATION",
			})
		}
	}

	return events
}

// VolumeProfile represents volume profile data
type VolumeProfile struct {
	POC float64 // Point of Control
	VAH float64 // Value Area High
	VAL float64 // Value Area Low
}

func (e *Engine) calculateVolumeProfile(ticks []Tick) VolumeProfile {
	// Build histogram
	volByPrice := make(map[float64]float64)
	for _, tick := range ticks {
		price := math.Round(tick.Price)
		volByPrice[price] += tick.Volume
	}

	// Find POC (most volume)
	var poc float64
	var maxVol float64
	for price, vol := range volByPrice {
		if vol > maxVol {
			maxVol = vol
			poc = price
		}
	}

	// Calculate value area (70% of volume)
	totalVol := 0.0
	for _, vol := range volByPrice {
		totalVol += vol
	}
	targetVol := totalVol * 0.7

	// Walk from POC outwards to find VAH and VAL
	var vah, val float64
	cumVol := maxVol

	for offset := 1.0; ; offset++ {
		upperVol := volByPrice[poc+offset]
		lowerVol := volByPrice[poc-offset]

		if cumVol+upperVol <= targetVol {
			cumVol += upperVol
			vah = poc + offset
		} else {
			break
		}

		if cumVol+lowerVol <= targetVol {
			cumVol += lowerVol
			val = poc - offset
		} else {
			break
		}
	}

	return VolumeProfile{
		POC: poc,
		VAH: vah,
		VAL: val,
	}
}

func (e *Engine) generateRecommendation(signal *OrderFlowSignal) string {
	if signal.CumDelta > 0 && signal.DeltaTrend == "INCREASING_BUYING" {
		return "BULLISH - Strong buying"
	}
	if signal.CumDelta < 0 && signal.DeltaTrend == "INCREASING_SELLING" {
		return "BEARISH - Strong selling"
	}
	return "NEUTRAL"
}

func (e *Engine) averageVolume(volumes map[float64]float64) float64 {
	if len(volumes) == 0 {
		return 0
	}
	var sum float64
	for _, v := range volumes {
		sum += v
	}
	return sum / float64(len(volumes))
}
