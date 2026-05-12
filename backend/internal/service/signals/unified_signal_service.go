package signals

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/antclaw/antclaw/internal/domain/shared"
	"github.com/antclaw/antclaw/internal/infra/postgres"
)

// UnifiedSignalService fuses signals from multiple subsystems
type UnifiedSignalService struct {
	cotRepo      postgres.COTRepository
	calendarRepo postgres.CalendarRepository
	priceRepo    postgres.PriceRepository
	macroRepo    postgres.MacroRepository
	pool         *pgxpool.Pool
}

// ComponentScore represents a component's signal score
type ComponentScore struct {
	Component     string            `json:"component"`
	Direction     shared.Direction  `json:"direction"`
	Confidence    float64           `json:"confidence"` // 0-1
	Weight        float64           `json:"weight"`
	WeightedScore float64           `json:"weighted_score"`
	Available     bool              `json:"available"`
	Error         string            `json:"error,omitempty"`
}

// UnifiedSignal represents the final fused signal
type UnifiedSignal struct {
	Timestamp       time.Time        `json:"timestamp"`
	Currency        string           `json:"currency"`
	Direction       shared.Direction `json:"direction"`
	Confidence      float64          `json:"confidence"` // 0-1
	Grade           shared.Grade     `json:"grade"`
	Components      []ComponentScore `json:"components"`
	Recommendation  shared.Recommendation `json:"recommendation"`
	Degradation     string           `json:"degradation,omitempty"`
}

// NewUnifiedSignalService creates a new unified signal service
func NewUnifiedSignalService(
	cotRepo postgres.COTRepository,
	calendarRepo postgres.CalendarRepository,
	priceRepo postgres.PriceRepository,
	macroRepo postgres.MacroRepository,
	pool *pgxpool.Pool,
) *UnifiedSignalService {
	return &UnifiedSignalService{
		cotRepo:      cotRepo,
		calendarRepo: calendarRepo,
		priceRepo:    priceRepo,
		macroRepo:    macroRepo,
		pool:         pool,
	}
}

// GenerateSignal generates unified signal for a currency
func (s *UnifiedSignalService) GenerateSignal(ctx context.Context, currency string) (*UnifiedSignal, error) {
	signal := &UnifiedSignal{
		Timestamp:  time.Now(),
		Currency:   currency,
		Components: make([]ComponentScore, 0),
	}

	components := []struct {
		name string
		fn   func(context.Context, string) (*ComponentScore, error)
	}{
		{"COT", s.getCOTScore},
		{"TA", s.getTAScore},
		{"Sentiment", s.getSentimentScore},
		{"Macro", s.getMacroScore},
		{"Calendar", s.getCalendarScore},
	}

	var totalWeight, weightedSum float64
	var degradation []string

	for _, comp := range components {
		score, err := comp.fn(ctx, currency)
		if err != nil {
			score = &ComponentScore{
				Component: comp.name,
				Available: false,
				Error:     err.Error(),
			}
			degradation = append(degradation, comp.name)
		}

		if score.Available {
			totalWeight += score.Weight
			weightedSum += score.WeightedScore
		}

		signal.Components = append(signal.Components, *score)
	}

	if totalWeight > 0 {
		signal.Confidence = math.Abs(weightedSum) / totalWeight
		signal.Direction = s.determineDirection(weightedSum)
	} else {
		signal.Confidence = 0
		signal.Direction = shared.DirectionNeutral
	}

	signal.Grade = s.determineGrade(signal.Confidence, totalWeight)
	signal.Recommendation = s.generateRecommendation(signal)

	if len(degradation) > 0 {
		signal.Degradation = fmt.Sprintf("missing: %v", degradation)
	}

	s.persistSignal(ctx, signal)

	return signal, nil
}

// getCOTScore gets COT component score
func (s *UnifiedSignalService) getCOTScore(ctx context.Context, currency string) (*ComponentScore, error) {
	analyses, err := s.cotRepo.GetLatestAll(ctx)
	if err != nil {
		return nil, err
	}

	var analysis *postgres.COTAnalysis
	for _, a := range analyses {
		if a.ContractCode == currency {
			analysis = a
			break
		}
	}

	if analysis == nil {
		return nil, fmt.Errorf("no COT data for %s", currency)
	}

	direction := shared.Direction(analysis.Direction)
	confidence := analysis.Percentile / 100
	if confidence < 0.5 {
		confidence = 0.5
	}

	score := confidence
	if direction == shared.DirectionBearish {
		score = -score
	}

	return &ComponentScore{
		Component:     "COT",
		Direction:     direction,
		Confidence:    confidence,
		Weight:        0.25,
		WeightedScore: score * 0.25,
		Available:     true,
	}, nil
}

// getTAScore computes technical analysis score from real price data.
func (s *UnifiedSignalService) getTAScore(ctx context.Context, currency string) (*ComponentScore, error) {
	from := time.Now().AddDate(0, 0, -60) // 60 days of daily data
	to := time.Now()

	bars, err := s.priceRepo.GetDailyBars(ctx, currency, from, to)
	if err != nil {
		return nil, fmt.Errorf("TA: price data unavailable: %w", err)
	}
	if len(bars) < 20 {
		return nil, fmt.Errorf("TA: insufficient bars (%d < 20) for %s", len(bars), currency)
	}

	// Compute SMA(5) and SMA(20) on last 20 bars
	closes := make([]float64, len(bars))
	for i, b := range bars {
		closes[i] = b.Close
	}

	// Use last 20 bars for computation
	window := closes[max(0, len(closes)-20):]
	sma5 := sma(window, min(5, len(window)))
	sma20 := sma(window, min(20, len(window)))

	if sma20 == 0 {
		return nil, fmt.Errorf("TA: cannot compute SMA for %s", currency)
	}

	// Direction: SMA(5) > SMA(20) = bullish
	diff := (sma5 - sma20) / sma20
	var direction shared.Direction
	if diff > 0.005 {
		direction = shared.DirectionBullish
	} else if diff < -0.005 {
		direction = shared.DirectionBearish
	} else {
		direction = shared.DirectionNeutral
	}

	// Confidence: based on separation magnitude (capped at 1.0)
	confidence := math.Min(math.Abs(diff)*100, 1.0)
	if confidence < 0.3 {
		confidence = 0.3
	}

	score := math.Abs(diff) * 100
	if direction == shared.DirectionBearish {
		score = -score
	}

	return &ComponentScore{
		Component:     "TA",
		Direction:     direction,
		Confidence:    confidence,
		Weight:        0.25,
		WeightedScore: score * 0.25,
		Available:     true,
	}, nil
}

// getSentimentScore reads sentiment from sentiment_snapshots table.
func (s *UnifiedSignalService) getSentimentScore(ctx context.Context, currency string) (*ComponentScore, error) {
	if s.pool == nil {
		return nil, fmt.Errorf("sentiment: postgres pool not configured")
	}

	var (
		score float64
		fg    float64
	)
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(score,0), COALESCE(fear_greed,0)
		  FROM sentiment_snapshots
		 ORDER BY time DESC LIMIT 1`).Scan(&score, &fg)
	if err != nil {
		return nil, fmt.Errorf("sentiment: %w", err)
	}

	// Normalize to -1..1
	norm := math.Max(-1, math.Min(1, score/100))
	_ = fg // reserved for future use

	var direction shared.Direction
	if norm > 0.15 {
		direction = shared.DirectionBullish
	} else if norm < -0.15 {
		direction = shared.DirectionBearish
	} else {
		direction = shared.DirectionNeutral
	}

	confidence := math.Abs(norm)
	if confidence < 0.3 {
		confidence = 0.3
	}

	return &ComponentScore{
		Component:     "Sentiment",
		Direction:     direction,
		Confidence:    confidence,
		Weight:        0.20,
		WeightedScore: norm * 0.20,
		Available:     true,
	}, nil
}

// getMacroScore reads macro regime from macro_regime_history.
func (s *UnifiedSignalService) getMacroScore(ctx context.Context, currency string) (*ComponentScore, error) {
	from := time.Now().AddDate(0, 0, -30)
	to := time.Now()

	snapshots, err := s.macroRepo.GetRegimeHistory(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("macro: regime unavailable: %w", err)
	}
	if len(snapshots) == 0 {
		return nil, fmt.Errorf("macro: no regime data in last 30 days")
	}

	// Use the most recent regime
	latest := snapshots[0]

	var direction shared.Direction
	switch latest.Regime {
	case "risk_on", "expansion":
		direction = shared.DirectionBullish
	case "risk_off", "contraction":
		direction = shared.DirectionBearish
	default:
		direction = shared.DirectionNeutral
	}

	confidence := math.Abs(latest.Score)
	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.3 {
		confidence = 0.3
	}

	score := confidence
	if direction == shared.DirectionBearish {
		score = -score
	}

	return &ComponentScore{
		Component:     "Macro",
		Direction:     direction,
		Confidence:    confidence,
		Weight:        0.15,
		WeightedScore: score * 0.15,
		Available:     true,
	}, nil
}

// getCalendarScore gets calendar impact score
func (s *UnifiedSignalService) getCalendarScore(ctx context.Context, currency string) (*ComponentScore, error) {
	events, err := s.calendarRepo.GetByCurrencyAndImpact(ctx, currency, string(shared.ImpactHigh), 5)
	if err != nil {
		return nil, err
	}

	var impact float64
	for _, e := range events {
		if e.ScheduledAt.After(time.Now()) && e.ScheduledAt.Before(time.Now().Add(24*time.Hour)) {
			impact += e.SurpriseScore * 0.1
		}
	}

	direction := shared.DirectionNeutral
	if impact > 0.3 {
		direction = shared.DirectionBullish
	} else if impact < -0.3 {
		direction = shared.DirectionBearish
	}

	avail := len(events) > 0

	return &ComponentScore{
		Component:     "Calendar",
		Direction:     direction,
		Confidence:    math.Min(math.Abs(impact), 1.0),
		Weight:        0.15,
		WeightedScore: impact * 0.15,
		Available:     avail,
	}, nil
}

// determineDirection converts weighted score to direction
func (s *UnifiedSignalService) determineDirection(weightedSum float64) shared.Direction {
	threshold := 0.1
	if weightedSum > threshold {
		return shared.DirectionBullish
	} else if weightedSum < -threshold {
		return shared.DirectionBearish
	}
	return shared.DirectionNeutral
}

// determineGrade determines signal grade based on confidence and weight coverage
func (s *UnifiedSignalService) determineGrade(confidence float64, totalWeight float64) shared.Grade {
	if totalWeight < 0.5 {
		return shared.GradeC
	}

	if confidence > 0.7 {
		return shared.GradeB // downgraded from A since components are simple
	} else if confidence > 0.4 {
		return shared.GradeB
	}
	return shared.GradeC
}

// generateRecommendation generates trading recommendation
func (s *UnifiedSignalService) generateRecommendation(signal *UnifiedSignal) shared.Recommendation {
	switch {
	case signal.Direction == shared.DirectionBullish:
		return shared.RecLong
	case signal.Direction == shared.DirectionBearish:
		return shared.RecShort
	default:
		return shared.RecNeutral
	}
}

// persistSignal persists signal to database (best-effort).
func (s *UnifiedSignalService) persistSignal(ctx context.Context, signal *UnifiedSignal) {
	if s.pool == nil {
		return
	}
	_, _ = s.pool.Exec(ctx, `
		INSERT INTO signals_history (time, currency, direction, confidence, grade, recommendation)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT DO NOTHING`,
		signal.Timestamp, signal.Currency,
		string(signal.Direction), signal.Confidence,
		string(signal.Grade), string(signal.Recommendation),
	)
}

// GetSignalHistory returns signal history for a currency from DB.
func (s *UnifiedSignalService) GetSignalHistory(ctx context.Context, currency string, days int) ([]UnifiedSignal, error) {
	if s.pool == nil {
		return nil, fmt.Errorf("signal history: postgres pool not configured")
	}
	from := time.Now().AddDate(0, 0, -days)

	rows, err := s.pool.Query(ctx, `
		SELECT time, currency, direction, confidence, grade, recommendation
		  FROM signals_history
		 WHERE currency = $1 AND time >= $2
		 ORDER BY time DESC`, currency, from)
	if err != nil {
		return nil, fmt.Errorf("signal history: %w", err)
	}
	defer rows.Close()

	var signals []UnifiedSignal
	for rows.Next() {
		var sig UnifiedSignal
		var dir, grd, rec string
		if err := rows.Scan(&sig.Timestamp, &sig.Currency, &dir, &sig.Confidence, &grd, &rec); err != nil {
			continue
		}
		sig.Direction = shared.Direction(dir)
		sig.Grade = shared.Grade(grd)
		sig.Recommendation = shared.Recommendation(rec)
		signals = append(signals, sig)
	}
	return signals, nil
}

// --- helpers ---

func sma(values []float64, period int) float64 {
	if period <= 0 || len(values) == 0 {
		return 0
	}
	n := min(period, len(values))
	window := values[len(values)-n:]
	sum := 0.0
	for _, v := range window {
		sum += v
	}
	return sum / float64(n)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
