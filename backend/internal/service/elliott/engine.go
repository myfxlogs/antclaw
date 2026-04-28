package elliott

import (
	"math"
	"time"
)

// Engine analyzes Elliott Wave patterns
type Engine struct {
	MinRetracement float64 // Default 5%
}

// NewEngine creates Elliott Wave engine
func NewEngine() *Engine {
	return &Engine{MinRetracement: 0.05}
}

// WaveCount represents detected wave pattern
type WaveCount struct {
	Detected    bool
	Waves       []Wave
	Confidence  float64
	Targets     WaveTargets
	Trend       string // "BULLISH" or "BEARISH"
}

// Wave represents a single Elliott Wave
type Wave struct {
	Number     int     // 1-5 for impulse, A-C for corrective
	Type       string  // "IMPULSE" or "CORRECTIVE"
	StartPrice float64
	EndPrice   float64
	Length     float64
	Duration   int     // Bars
}

// WaveTargets represents projected targets
type WaveTargets struct {
	W5TargetConservative float64
	W5TargetAggressive float64
}

// Analyze performs Elliott Wave analysis
func (e *Engine) Analyze(bars []OHLCV) (*WaveCount, error) {
	// Step 1: Find swing points using ZigZag
	swings := e.findSwingPoints(bars)
	if len(swings) < 5 {
		return &WaveCount{Detected: false}, nil
	}

	// Step 2: Try to identify wave pattern
	waves := e.identifyWaves(swings)

	// Step 3: Validate against Elliott rules
	valid, confidence := e.validateWaves(waves)

	if !valid {
		return &WaveCount{Detected: false}, nil
	}

	// Step 4: Project targets
	targets := e.projectTargets(waves)

	return &WaveCount{
		Detected:   true,
		Waves:      waves,
		Confidence: confidence,
		Targets:    targets,
		Trend:      e.determineTrend(waves),
	}, nil
}

// OHLCV represents price bar
type OHLCV struct {
	Open, High, Low, Close float64
	Volume                 float64
	Time                   time.Time
}

// SwingPoint represents a swing high or low
type SwingPoint struct {
	Index  int
	Price  float64
	IsHigh bool
}

func (e *Engine) findSwingPoints(bars []OHLCV) []SwingPoint {
	var swings []SwingPoint

	for i := 2; i < len(bars)-2; i++ {
		// Check for swing high
		if bars[i].High > bars[i-1].High && bars[i].High > bars[i-2].High &&
			bars[i].High > bars[i+1].High && bars[i].High > bars[i+2].High {
			swings = append(swings, SwingPoint{Index: i, Price: bars[i].High, IsHigh: true})
		}

		// Check for swing low
		if bars[i].Low < bars[i-1].Low && bars[i].Low < bars[i-2].Low &&
			bars[i].Low < bars[i+1].Low && bars[i].Low < bars[i+2].Low {
			swings = append(swings, SwingPoint{Index: i, Price: bars[i].Low, IsHigh: false})
		}
	}

	return swings
}

func (e *Engine) identifyWaves(swings []SwingPoint) []Wave {
	if len(swings) < 5 {
		return nil
	}

	var waves []Wave
	for i := 0; i < 5 && i < len(swings)-1; i++ {
		waves = append(waves, Wave{
			Number:     i + 1,
			Type:       "IMPULSE",
			StartPrice: swings[i].Price,
			EndPrice:   swings[i+1].Price,
			Length:     math.Abs(swings[i+1].Price - swings[i].Price),
			Duration:   swings[i+1].Index - swings[i].Index,
		})
	}

	return waves
}

func (e *Engine) validateWaves(waves []Wave) (bool, float64) {
	if len(waves) < 5 {
		return false, 0
	}

	confidence := 1.0

	// Rule 1: Wave 2 cannot retrace more than 100% of Wave 1
	w2Retrace := waves[1].Length / waves[0].Length
	if w2Retrace > 1.0 {
		return false, 0
	}
	confidence *= (1.0 - w2Retrace*0.3)

	// Rule 2: Wave 3 cannot be the shortest
	if waves[2].Length < waves[0].Length && waves[2].Length < waves[4].Length {
		return false, 0
	}

	// Rule 3: Wave 4 cannot enter Wave 1 price territory
	w4Low := math.Min(waves[3].StartPrice, waves[3].EndPrice)
	w1High := math.Max(waves[0].StartPrice, waves[0].EndPrice)
	if w4Low < w1High {
		return false, 0
	}

	return true, confidence
}

func (e *Engine) projectTargets(waves []Wave) WaveTargets {
	if len(waves) < 3 {
		return WaveTargets{}
	}

	w1Length := waves[0].Length
	w3End := waves[2].EndPrice

	// Conservative: W5 = W1 length
	// Aggressive: W5 = 1.618 * W1 length
	return WaveTargets{
		W5TargetConservative: w3End + w1Length,
		W5TargetAggressive:   w3End + w1Length*1.618,
	}
}

func (e *Engine) determineTrend(waves []Wave) string {
	if len(waves) == 0 {
		return "NEUTRAL"
	}
	if waves[0].EndPrice > waves[0].StartPrice {
		return "BULLISH"
	}
	return "BEARISH"
}
