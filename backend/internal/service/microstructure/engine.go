package microstructure

import (
	"math"
	"time"
)

// Engine analyzes order book microstructure
type Engine struct{}

// NewEngine creates microstructure engine
func NewEngine() *Engine {
	return &Engine{}
}

// MicrostructureSignal represents microstructure analysis
type MicrostructureSignal struct {
	Timestamp    time.Time
	Symbol       string
	OBI          float64 // Order Book Imbalance
	Spread       float64 // Bid-ask spread
	SpreadPct    float64 // Spread as percentage
	Depth        DepthMetrics
	StressScore  float64
	Recommendation string
}

// DepthMetrics represents order book depth
type DepthMetrics struct {
	BidDepth       float64
	AskDepth       float64
	TotalDepth     float64
	DepthImbalance float64
}

// Analyze performs microstructure analysis
func (e *Engine) Analyze(symbol string, orderbook OrderBook) (*MicrostructureSignal, error) {
	signal := &MicrostructureSignal{
		Timestamp: time.Now(),
		Symbol:    symbol,
		OBI:       e.calculateOBI(orderbook),
		Spread:    orderbook.Asks[0].Price - orderbook.Bids[0].Price,
	}
	signal.SpreadPct = signal.Spread / ((orderbook.Asks[0].Price + orderbook.Bids[0].Price) / 2)
	signal.Depth = e.calculateDepth(orderbook)
	signal.StressScore = e.calculateStress(signal.SpreadPct, signal.Depth)
	signal.Recommendation = e.generateRecommendation(signal)
	return signal, nil
}

// OrderBook represents order book data
type OrderBook struct {
	Bids []Level
	Asks []Level
}

// Level represents a price level
type Level struct {
	Price  float64
	Volume float64
}

func (e *Engine) calculateOBI(ob OrderBook) float64 {
	var bidVol, askVol float64
	for _, bid := range ob.Bids {
		bidVol += bid.Volume
	}
	for _, ask := range ob.Asks {
		askVol += ask.Volume
	}
	total := bidVol + askVol
	if total == 0 {
		return 0
	}
	return (bidVol - askVol) / total
}

func (e *Engine) calculateDepth(ob OrderBook) DepthMetrics {
	var bidDepth, askDepth float64
	for i := 0; i < min(5, len(ob.Bids)); i++ {
		bidDepth += ob.Bids[i].Volume
	}
	for i := 0; i < min(5, len(ob.Asks)); i++ {
		askDepth += ob.Asks[i].Volume
	}
	total := bidDepth + askDepth
	return DepthMetrics{
		BidDepth:       bidDepth,
		AskDepth:       askDepth,
		TotalDepth:     total,
		DepthImbalance: (bidDepth - askDepth) / total,
	}
}

func (e *Engine) calculateStress(spreadPct float64, depth DepthMetrics) float64 {
	spreadScore := math.Min(spreadPct*10000, 50)
	depthScore := 0.0
	if depth.TotalDepth < 1000000 {
		depthScore = 50 * (1 - depth.TotalDepth/1000000)
	}
	return spreadScore + depthScore
}

func (e *Engine) generateRecommendation(signal *MicrostructureSignal) string {
	if signal.StressScore > 70 {
		return "AVOID - High stress"
	}
	if signal.SpreadPct > 0.005 {
		return "CAUTION - Wide spread"
	}
	if signal.OBI > 0.3 {
		return "BULLISH - Buying pressure"
	} else if signal.OBI < -0.3 {
		return "BEARISH - Selling pressure"
	}
	return "NEUTRAL"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
