package ict

import (
	"time"

	"github.com/antclaw/antclaw/internal/domain/shared"
)

// Engine analyzes ICT (Inner Circle Trader) concepts
type Engine struct{}

// NewEngine creates ICT engine
func NewEngine() *Engine {
	return &Engine{}
}

// ICTReport represents analysis result
type ICTReport struct {
	Timestamp     time.Time
	BOS           []StructureEvent // Break of Structure
	ChoCH         []StructureEvent // Change of Character
	Sweeps        []LiquiditySweep
	FairValueGaps []FVG
	OrderBlocks   []OrderBlock
}

// StructureEvent represents BOS or ChoCH
type StructureEvent struct {
	Time      time.Time
	Price     float64
	Type      string // "BOS" or "ChoCH"
	Direction string // "BULLISH" or "BEARISH"
}

// LiquiditySweep represents liquidity sweep detection
type LiquiditySweep struct {
	Time      time.Time
	Price     float64
	Type      string // "SWEEP_HIGH" or "SWEEP_LOW"
	Reversed  bool
}

// FVG represents Fair Value Gap
type FVG struct {
	Time     time.Time
	Top      float64
	Bottom   float64
	Type     string // "BULLISH" or "BEARISH"
	IsFilled bool
}

// OrderBlock represents Order Block
type OrderBlock struct {
	Time      time.Time
	High      float64
	Low       float64
	Type      string // "BULLISH" or "BEARISH"
	Mitigated bool
}

// OHLCV represents price bar
type OHLCV struct {
	Open, High, Low, Close float64
	Volume                 float64
	Time                   time.Time
}

// Analyze performs ICT analysis
func (e *Engine) Analyze(bars []OHLCV, symbol string) (*ICTReport, error) {
	report := &ICTReport{
		Timestamp:     time.Now(),
		BOS:           make([]StructureEvent, 0),
		ChoCH:         make([]StructureEvent, 0),
		Sweeps:        make([]LiquiditySweep, 0),
		FairValueGaps: make([]FVG, 0),
		OrderBlocks:   make([]OrderBlock, 0),
	}

	// Find swing points
	swings := e.findSwingPoints(bars)

	// Detect BOS and ChoCH
	bosEvents, chochEvents := e.detectStructure(bars, swings)
	report.BOS = bosEvents
	report.ChoCH = chochEvents

	// Detect liquidity sweeps
	report.Sweeps = e.detectSweeps(bars, swings)

	// Detect Fair Value Gaps
	report.FairValueGaps = e.detectFVGs(bars)

	// Detect Order Blocks
	report.OrderBlocks = e.detectOrderBlocks(bars)

	return report, nil
}

// SwingPoint represents a swing high/low
type SwingPoint struct {
	Index  int
	Price  float64
	IsHigh bool
}

func (e *Engine) findSwingPoints(bars []OHLCV) []SwingPoint {
	var swings []SwingPoint

	for i := 2; i < len(bars)-2; i++ {
		// Swing high
		if bars[i].High > bars[i-1].High && bars[i].High > bars[i-2].High &&
			bars[i].High > bars[i+1].High && bars[i].High > bars[i+2].High {
			swings = append(swings, SwingPoint{Index: i, Price: bars[i].High, IsHigh: true})
		}

		// Swing low
		if bars[i].Low < bars[i-1].Low && bars[i].Low < bars[i-2].Low &&
			bars[i].Low < bars[i+1].Low && bars[i].Low < bars[i+2].Low {
			swings = append(swings, SwingPoint{Index: i, Price: bars[i].Low, IsHigh: false})
		}
	}

	return swings
}

func (e *Engine) detectStructure(bars []OHLCV, swings []SwingPoint) ([]StructureEvent, []StructureEvent) {
	var bosEvents []StructureEvent
	var chochEvents []StructureEvent

	if len(swings) < 4 {
		return bosEvents, chochEvents
	}

	// Need at least 2 highs and 2 lows for structure analysis
	var highs, lows []SwingPoint
	for _, s := range swings {
		if s.IsHigh {
			highs = append(highs, s)
		} else {
			lows = append(lows, s)
		}
	}

	if len(highs) < 2 || len(lows) < 2 {
		return bosEvents, chochEvents
	}

	// BOS: Break of Structure (trend continuation)
	// In uptrend: price closes above previous swing high
	lastHigh := highs[len(highs)-1]
	prevHigh := highs[len(highs)-2]

	if len(bars) > lastHigh.Index+1 {
		lastBar := bars[len(bars)-1]
		if lastBar.Close > prevHigh.Price && lastBar.Close > lastHigh.Price {
			bosEvents = append(bosEvents, StructureEvent{
				Time:      lastBar.Time,
				Price:     lastBar.Close,
				Type:      "BOS",
				Direction: "BULLISH",
			})
		}
	}

	// ChoCH: Change of Character (trend reversal)
	// Uptrend broken: price closes below recent swing low
	lastLow := lows[len(lows)-1]
	if len(bars) > lastLow.Index+1 {
		lastBar := bars[len(bars)-1]
		if lastBar.Close < lastLow.Price {
			chochEvents = append(chochEvents, StructureEvent{
				Time:      lastBar.Time,
				Price:     lastBar.Close,
				Type:      "ChoCH",
				Direction: "BEARISH",
			})
		}
	}

	return bosEvents, chochEvents
}

func (e *Engine) detectSweeps(bars []OHLCV, swings []SwingPoint) []LiquiditySweep {
	var sweeps []LiquiditySweep

	if len(swings) < 2 || len(bars) < 2 {
		return sweeps
	}

	// Get last swing high and low
	var lastHigh, lastLow float64

	for i := len(swings) - 1; i >= 0; i-- {
		if swings[i].IsHigh && lastHigh == 0 {
			lastHigh = swings[i].Price
		}
		if !swings[i].IsHigh && lastLow == 0 {
			lastLow = swings[i].Price
		}
		if lastHigh != 0 && lastLow != 0 {
			break
		}
	}

	// Check last bar for sweep
	lastBar := bars[len(bars)-1]
	prevBar := bars[len(bars)-2]

	// Sweep high: wick above previous high, close below
	if lastBar.High > lastHigh && lastBar.Close < lastHigh && lastBar.Close < prevBar.Close {
		sweeps = append(sweeps, LiquiditySweep{
			Time:     lastBar.Time,
			Price:    lastBar.High,
			Type:     "SWEEP_HIGH",
			Reversed: true,
		})
	}

	// Sweep low: wick below previous low, close above
	if lastBar.Low < lastLow && lastBar.Close > lastLow && lastBar.Close > prevBar.Close {
		sweeps = append(sweeps, LiquiditySweep{
			Time:     lastBar.Time,
			Price:    lastBar.Low,
			Type:     "SWEEP_LOW",
			Reversed: true,
		})
	}

	return sweeps
}

func (e *Engine) detectFVGs(bars []OHLCV) []FVG {
	var fvgs []FVG

	if len(bars) < 3 {
		return fvgs
	}

	// Check last 3 bars for FVG
	for i := 0; i < len(bars)-2; i++ {
		bar1 := bars[i]
		bar2 := bars[i+1]
		bar3 := bars[i+2]

		// Bullish FVG: bar1 high < bar3 low (gap up)
		if bar1.High < bar3.Low {
			fvgs = append(fvgs, FVG{
				Time:     bar2.Time,
				Top:      bar3.Low,
				Bottom:   bar1.High,
				Type:     "BULLISH",
				IsFilled: false,
			})
		}

		// Bearish FVG: bar1 low > bar3 high (gap down)
		if bar1.Low > bar3.High {
			fvgs = append(fvgs, FVG{
				Time:     bar2.Time,
				Top:      bar1.Low,
				Bottom:   bar3.High,
				Type:     "BEARISH",
				IsFilled: false,
			})
		}
	}

	return fvgs
}

func (e *Engine) detectOrderBlocks(bars []OHLCV) []OrderBlock {
	var obs []OrderBlock

	if len(bars) < 5 {
		return obs
	}

	// Look for strong momentum candles followed by reversal
	for i := 4; i < len(bars); i++ {
		candles := bars[i-4 : i]

		// Bullish OB: Strong bearish candle before bullish move
		if e.isStrongBearish(candles[0]) && bars[i].Close > candles[0].Open {
			obs = append(obs, OrderBlock{
				Time:      candles[0].Time,
				High:      candles[0].Open,
				Low:       candles[0].Low,
				Type:      "BULLISH",
				Mitigated: false,
			})
		}

		// Bearish OB: Strong bullish candle before bearish move
		if e.isStrongBullish(candles[0]) && bars[i].Close < candles[0].Open {
			obs = append(obs, OrderBlock{
				Time:      candles[0].Time,
				High:      candles[0].High,
				Low:       candles[0].Open,
				Type:      "BEARISH",
				Mitigated: false,
			})
		}
	}

	return obs
}

func (e *Engine) isStrongBearish(bar OHLCV) bool {
	body := bar.Open - bar.Close
	range_ := bar.High - bar.Low
	return body > 0 && body > range_*0.6
}

func (e *Engine) isStrongBullish(bar OHLCV) bool {
	body := bar.Close - bar.Open
	range_ := bar.High - bar.Low
	return body > 0 && body > range_*0.6
}

// GetSignal provides simplified signal from ICT analysis
func (e *Engine) GetSignal(report *ICTReport) shared.Direction {
	// Check for ChoCH (strongest signal)
	for _, choch := range report.ChoCH {
		if choch.Direction == "BULLISH" {
			return shared.DirectionLong
		} else if choch.Direction == "BEARISH" {
			return shared.DirectionShort
		}
	}

	// Check for BOS
	for _, bos := range report.BOS {
		if bos.Direction == "BULLISH" {
			return shared.DirectionLong
		} else if bos.Direction == "BEARISH" {
			return shared.DirectionShort
		}
	}

	return shared.DirectionNeutral
}
