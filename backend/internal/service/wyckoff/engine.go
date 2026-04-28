package wyckoff

import (
	"time"

	"github.com/antclaw/antclaw/internal/domain/shared"
)

// Engine analyzes Wyckoff market phases
type Engine struct{}

// NewEngine creates Wyckoff engine
func NewEngine() *Engine {
	return &Engine{}
}

// WyckoffReport represents analysis result
type WyckoffReport struct {
	DetectedAt    time.Time
	Phase         string // ACCUMULATION, DISTRIBUTION, TRANSITION
	Confidence    float64
	Events        []WyckoffEvent
	Schematic     string
	CauseScore    float64 // Measure of potential move
	DirectionHint shared.Direction
}

// WyckoffEvent represents detected event
type WyckoffEvent struct {
	Name      string
	Type      string // PS, SC, AR, ST, SOS, LPS, etc.
	Time      time.Time
	Price     float64
	Volume    float64
	Strength  float64
}

// Analyze performs Wyckoff analysis
func (e *Engine) Analyze(bars []OHLCV) (*WyckoffReport, error) {
	report := &WyckoffReport{
		DetectedAt: time.Now(),
		Events:     make([]WyckoffEvent, 0),
	}

	// Detect events in recent 60 bars
	recentBars := bars
	if len(bars) > 60 {
		recentBars = bars[len(bars)-60:]
	}

	// Phase A events
	if ps := e.detectPS(recentBars); ps != nil {
		report.Events = append(report.Events, *ps)
	}
	if sc := e.detectSC(recentBars); sc != nil {
		report.Events = append(report.Events, *sc)
	}
	if ar := e.detectAR(recentBars); ar != nil {
		report.Events = append(report.Events, *ar)
	}
	if st := e.detectST(recentBars); st != nil {
		report.Events = append(report.Events, *st)
	}

	// Phase C events
	if spring := e.detectSpring(recentBars); spring != nil {
		report.Events = append(report.Events, *spring)
	}

	// Determine phase
	report.Phase = e.classifyPhase(report.Events)
	report.Confidence = e.calculateConfidence(report.Events)
	report.Schematic = e.determineSchematic(report.Events)
	report.CauseScore = e.calculateCause(report.Events)
	report.DirectionHint = e.inferDirection(report.Phase)

	return report, nil
}

// OHLCV represents price bar
type OHLCV struct {
	Open, High, Low, Close float64
	Volume                 float64
	Time                   time.Time
}

func (e *Engine) detectPS(bars []OHLCV) *WyckoffEvent {
	// Preliminary Support: High volume support after decline
	if len(bars) < 10 {
		return nil
	}

	recent := bars[len(bars)-10:]
	avgVol := e.avgVolume(recent)

	for _, bar := range recent {
		if bar.Volume > avgVol*2 && bar.Low == e.findLow(recent) {
			return &WyckoffEvent{
				Name:     "Preliminary Support",
				Type:     "PS",
				Time:     bar.Time,
				Price:    bar.Low,
				Volume:   bar.Volume,
				Strength: bar.Volume / avgVol,
			}
		}
	}
	return nil
}

func (e *Engine) detectSC(bars []OHLCV) *WyckoffEvent {
	// Selling Climax: Very high volume, wide range, often spring
	if len(bars) < 5 {
		return nil
	}

	recent := bars[len(bars)-5:]
	avgVol := e.avgVolume(recent)
	avgRange := e.avgRange(recent)

	for _, bar := range recent {
		range_ := bar.High - bar.Low
		if bar.Volume > avgVol*3 && range_ > avgRange*2 {
			return &WyckoffEvent{
				Name:     "Selling Climax",
				Type:     "SC",
				Time:     bar.Time,
				Price:    bar.Low,
				Volume:   bar.Volume,
				Strength: bar.Volume / avgVol,
			}
		}
	}
	return nil
}

func (e *Engine) detectAR(bars []OHLCV) *WyckoffEvent {
	// Automatic Rally: After SC, price rallies on lower volume
	// Simplified detection
	return nil
}

func (e *Engine) detectST(bars []OHLCV) *WyckoffEvent {
	// Secondary Test: Lower volume test of SC area
	return nil
}

func (e *Engine) detectSpring(bars []OHLCV) *WyckoffEvent {
	// Spring: Brief break below support with recovery
	if len(bars) < 5 {
		return nil
	}

	recent := bars[len(bars)-5:]
	support := e.findSupport(bars[:len(bars)-5])

	for i, bar := range recent {
		if bar.Low < support && bar.Close > support && i < len(recent)-1 {
			return &WyckoffEvent{
				Name:     "Spring",
				Type:     "SPRING",
				Time:     bar.Time,
				Price:    bar.Low,
				Volume:   bar.Volume,
				Strength: (support - bar.Low) / support * 100,
			}
		}
	}
	return nil
}

func (e *Engine) classifyPhase(events []WyckoffEvent) string {
	accumEvents := 0
	distEvents := 0

	for _, ev := range events {
		switch ev.Type {
		case "PS", "SC", "AR", "ST", "SPRING", "SOS":
			accumEvents++
		case "BC", "UPTHRUST":
			distEvents++
		}
	}

	if accumEvents > distEvents {
		return "ACCUMULATION"
	} else if distEvents > accumEvents {
		return "DISTRIBUTION"
	}
	return "TRANSITION"
}

func (e *Engine) calculateConfidence(events []WyckoffEvent) float64 {
	if len(events) == 0 {
		return 0
	}

	var totalStrength float64
	for _, ev := range events {
		totalStrength += ev.Strength
	}

	confidence := totalStrength / float64(len(events)) / 10
	if confidence > 1.0 {
		confidence = 1.0
	}
	return confidence
}

func (e *Engine) determineSchematic(events []WyckoffEvent) string {
	var types []string
	for _, ev := range events {
		types = append(types, ev.Type)
	}
	if len(types) == 0 {
		return "NONE"
	}
	return types[0]
}

func (e *Engine) calculateCause(events []WyckoffEvent) float64 {
	// Measure of potential move based on accumulation/distribution
	var cause float64
	for _, ev := range events {
		if ev.Type == "PS" || ev.Type == "SC" {
			cause += ev.Strength
		}
	}
	return cause
}

func (e *Engine) inferDirection(phase string) shared.Direction {
	switch phase {
	case "ACCUMULATION":
		return shared.DirectionLong
	case "DISTRIBUTION":
		return shared.DirectionShort
	default:
		return shared.DirectionNeutral
	}
}

func (e *Engine) avgVolume(bars []OHLCV) float64 {
	var sum float64
	for _, bar := range bars {
		sum += bar.Volume
	}
	return sum / float64(len(bars))
}

func (e *Engine) avgRange(bars []OHLCV) float64 {
	var sum float64
	for _, bar := range bars {
		sum += bar.High - bar.Low
	}
	return sum / float64(len(bars))
}

func (e *Engine) findLow(bars []OHLCV) float64 {
	low := bars[0].Low
	for _, bar := range bars {
		if bar.Low < low {
			low = bar.Low
		}
	}
	return low
}

func (e *Engine) findSupport(bars []OHLCV) float64 {
	// Simplified: find most frequent low
	return e.findLow(bars)
}
