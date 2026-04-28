package signals

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/antclaw/antclaw/internal/domain/shared"
	"github.com/antclaw/antclaw/internal/infra/postgres"
)

// UnifiedSignalService fuses signals from multiple subsystems
type UnifiedSignalService struct {
	cotRepo       postgres.COTRepository
	calendarRepo  postgres.CalendarRepository
	priceRepo     postgres.PriceRepository
	logger        *slog.Logger
}

// ComponentScore represents a component's signal score
type ComponentScore struct {
	Component   string  `json:"component"`
	Direction   shared.Direction `json:"direction"`
	Confidence  float64 `json:"confidence"` // 0-1
	Weight      float64 `json:"weight"`
	WeightedScore float64 `json:"weighted_score"`
	Available   bool    `json:"available"`
	Error       string  `json:"error,omitempty"`
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
	logger *slog.Logger,
) *UnifiedSignalService {
	return &UnifiedSignalService{
		cotRepo:      cotRepo,
		calendarRepo: calendarRepo,
		priceRepo:    priceRepo,
		logger:       logger,
	}
}

// GenerateSignal generates unified signal for a currency
func (s *UnifiedSignalService) GenerateSignal(ctx context.Context, currency string) (*UnifiedSignal, error) {
	signal := &UnifiedSignal{
		Timestamp:  time.Now(),
		Currency:   currency,
		Components: make([]ComponentScore, 0),
	}

	// Gather component signals with graceful degradation
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
			s.logger.Warn("component failed", "component", comp.name, "error", err)
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

	// Calculate final signal
	if totalWeight > 0 {
		signal.Confidence = math.Abs(weightedSum) / totalWeight
		signal.Direction = s.determineDirection(weightedSum)
	} else {
		signal.Confidence = 0
		signal.Direction = shared.DirectionNeutral
	}

	// Determine grade
	signal.Grade = s.determineGrade(signal.Confidence, totalWeight)

	// Generate recommendation
	signal.Recommendation = s.generateRecommendation(signal)

	// Record degradation
	if len(degradation) > 0 {
		signal.Degradation = fmt.Sprintf("missing: %v", degradation)
	}

	// Persist signal
	s.persistSignal(ctx, signal)

	return signal, nil
}

// getCOTScore gets COT component score
func (s *UnifiedSignalService) getCOTScore(ctx context.Context, currency string) (*ComponentScore, error) {
	// Get COT analysis from repository
	analyses, err := s.cotRepo.GetLatestAll(ctx)
	if err != nil {
		return nil, err
	}

	// Find analysis for currency
	var analysis *postgres.COTAnalysis
	for _, a := range analyses {
		// Simple matching - in production use proper mapping
		if a.ContractCode == currency {
			analysis = a
			break
		}
	}

	if analysis == nil {
		return nil, fmt.Errorf("no COT data for %s", currency)
	}

	// Convert direction and calculate confidence
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

// getTAScore gets technical analysis score
func (s *UnifiedSignalService) getTAScore(ctx context.Context, currency string) (*ComponentScore, error) {
	// In production: fetch TA signals from repository
	// For now, return neutral
	return &ComponentScore{
		Component:     "TA",
		Direction:     shared.DirectionNeutral,
		Confidence:    0.5,
		Weight:        0.25,
		WeightedScore: 0,
		Available:     false,
		Error:         "TA not implemented",
	}, nil
}

// getSentimentScore gets sentiment score
func (s *UnifiedSignalService) getSentimentScore(ctx context.Context, currency string) (*ComponentScore, error) {
	// In production: fetch sentiment data
	return &ComponentScore{
		Component:     "Sentiment",
		Direction:     shared.DirectionNeutral,
		Confidence:    0.5,
		Weight:        0.20,
		WeightedScore: 0,
		Available:     false,
		Error:         "Sentiment not implemented",
	}, nil
}

// getMacroScore gets macro score
func (s *UnifiedSignalService) getMacroScore(ctx context.Context, currency string) (*ComponentScore, error) {
	// In production: fetch macro regime
	return &ComponentScore{
		Component:     "Macro",
		Direction:     shared.DirectionNeutral,
		Confidence:    0.5,
		Weight:        0.15,
		WeightedScore: 0,
		Available:     false,
		Error:         "Macro not implemented",
	}, nil
}

// getCalendarScore gets calendar impact score
func (s *UnifiedSignalService) getCalendarScore(ctx context.Context, currency string) (*ComponentScore, error) {
	// Get upcoming events for currency
	events, err := s.calendarRepo.GetByCurrencyAndImpact(ctx, currency, string(shared.ImpactHigh), 5)
	if err != nil {
		return nil, err
	}

	// Calculate impact based on upcoming events
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

	return &ComponentScore{
		Component:     "Calendar",
		Direction:     direction,
		Confidence:    math.Min(math.Abs(impact), 1.0),
		Weight:        0.15,
		WeightedScore: impact * 0.15,
		Available:     len(events) > 0,
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
	// Check if enough components are available
	if totalWeight < 0.5 {
		return shared.GradeC
	}

	// Grade based on confidence
	if confidence > 0.7 {
		return shared.GradeA
	} else if confidence > 0.4 {
		return shared.GradeB
	}
	return shared.GradeC
}

// generateRecommendation generates trading recommendation
func (s *UnifiedSignalService) generateRecommendation(signal *UnifiedSignal) shared.Recommendation {
	switch {
	case signal.Grade == shared.GradeA && signal.Direction == shared.DirectionBullish:
		return shared.RecStrongLong
	case signal.Grade == shared.GradeA && signal.Direction == shared.DirectionBearish:
		return shared.RecStrongShort
	case signal.Direction == shared.DirectionBullish:
		return shared.RecLong
	case signal.Direction == shared.DirectionBearish:
		return shared.RecShort
	default:
		return shared.RecNeutral
	}
}

// persistSignal persists signal to database
func (s *UnifiedSignalService) persistSignal(ctx context.Context, signal *UnifiedSignal) {
	// In production: persist to unified_signals table
	s.logger.Info("signal generated",
		"currency", signal.Currency,
		"direction", signal.Direction,
		"confidence", signal.Confidence,
		"grade", signal.Grade,
	)
}

// GetSignalHistory returns signal history for a currency
func (s *UnifiedSignalService) GetSignalHistory(ctx context.Context, currency string, days int) ([]UnifiedSignal, error) {
	// In production: fetch from database
	return nil, fmt.Errorf("not implemented")
}
