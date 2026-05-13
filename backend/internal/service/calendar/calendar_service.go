package calendar

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/antclaw/antclaw/internal/domain/shared"
	"github.com/antclaw/antclaw/internal/infra/apiclient"
	"github.com/antclaw/antclaw/internal/infra/postgres"
	"github.com/antclaw/antclaw/internal/infra/redis"
)

// CalendarService provides economic calendar operations
type CalendarService struct {
	repo      postgres.CalendarRepository
	client    *apiclient.MQL5Fetcher
	redis     *redis.Client
	logger    *slog.Logger
}

// NewCalendarService creates a new calendar service
func NewCalendarService(repo postgres.CalendarRepository, redis *redis.Client, logger *slog.Logger) *CalendarService {
	return &CalendarService{
		repo:   repo,
		client: apiclient.NewMQL5Fetcher(),
		redis:  redis,
		logger: logger,
	}
}

// SyncWeek fetches and syncs calendar events for a week
func (s *CalendarService) SyncWeek(ctx context.Context) (*SyncResult, error) {
	events, err := s.client.FetchWeek(ctx, "current")
	if err != nil {
		return nil, fmt.Errorf("fetch week failed: %w", err)
	}

	s.logger.Info("fetched calendar events", "count", len(events))

	// Transform to internal format
	var calendarEvents []postgres.CalendarEvent
	for _, e := range events {
		// Calculate surprise score if actual value available
		var surprise float64
		if e.Actual != "" && e.Forecast != "" {
			surprise = s.calculateSurprise(e.Actual, e.Forecast, e.Previous)
		}

		direction := s.calculateImpactDirection(e)
		surpriseLabel := s.classifySurprise(surprise)

		// Parse scheduled time - prefer ReleaseDate (Unix ms), fallback to FullDate
		scheduledAt, err := s.parseEventTime(e)
		if err != nil {
			s.logger.Warn("failed to parse event time, skipping", "event_id", e.EventID, "error", err)
			continue
		}

		event := postgres.CalendarEvent{
			EventID:         s.truncate(s.generateEventID(e), 64),
			Title:           s.truncate(e.Title, 256),
			Country:         s.truncate(e.Country, 8),
			Currency:        s.truncate(e.Currency, 8),
			Impact:          s.truncate(e.Impact, 16),
			ScheduledAt:     scheduledAt,
			PreviousValue:   s.truncate(e.Previous, 256),
			ForecastValue:   s.truncate(e.Forecast, 256),
			ActualValue:     s.truncate(e.Actual, 256),
			ImpactDirection: direction,
			SurpriseScore:   surprise,
			SurpriseLabel:   s.truncate(surpriseLabel, 32),
			FetchedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}
		calendarEvents = append(calendarEvents, event)
	}

	// Upsert to database
	count, err := s.repo.UpsertEvents(ctx, calendarEvents)
	if err != nil {
		return nil, fmt.Errorf("upsert events failed: %w", err)
	}

	// Cache in Redis
	s.cacheEvents(ctx, calendarEvents)

	return &SyncResult{
		Inserted: count,
		Total:    len(events),
	}, nil
}

// SyncActuals checks for updated actual values
func (s *CalendarService) SyncActuals(ctx context.Context) (*SyncResult, error) {
	// Get current week events again to refresh actual values
	events, err := s.client.FetchWeek(ctx, "current")
	if err != nil {
		return nil, fmt.Errorf("fetch current week failed: %w", err)
	}

	var updated int
	for _, e := range events {
		if e.Actual != "" {
			surprise := s.calculateSurprise(e.Actual, e.Forecast, e.Previous)
			eventID := s.generateEventID(e)
			if err := s.repo.UpdateActual(ctx, eventID, e.Actual, surprise); err != nil {
				s.logger.Warn("failed to update actual", "event", eventID, "error", err)
				continue
			}
			updated++
		}
	}

	return &SyncResult{Updated: updated}, nil
}

// GetUpcoming returns upcoming high-impact events
func (s *CalendarService) GetUpcoming(ctx context.Context, hours int) ([]postgres.CalendarEvent, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("calendar:upcoming:%d", hours)
	var cached []postgres.CalendarEvent
	if err := s.redis.GetJSON(ctx, cacheKey, &cached); err == nil && len(cached) > 0 {
		return cached, nil
	}

	// Fetch from database
	events, err := s.repo.GetUpcoming(ctx, time.Duration(hours)*time.Hour)
	if err != nil {
		return nil, err
	}

	// Filter high impact only
	var highImpact []postgres.CalendarEvent
	for _, e := range events {
		if e.Impact == string(shared.ImpactHigh) {
			highImpact = append(highImpact, e)
		}
	}

	// Cache for 5 minutes
	s.redis.SetJSON(ctx, cacheKey, highImpact, 5*time.Minute)

	return highImpact, nil
}

// RecordImpact records price impact of an event
func (s *CalendarService) RecordImpact(ctx context.Context, eventID, symbol string) error {
	// Get event
	events, err := s.repo.GetByDate(ctx, time.Now().Add(-7*24*time.Hour))
	if err != nil {
		return err
	}

	var targetEvent *postgres.CalendarEvent
	for _, e := range events {
		if e.EventID == eventID {
			targetEvent = &e
			break
		}
	}

	if targetEvent == nil {
		return fmt.Errorf("event %s not found", eventID)
	}

	// Record impacts at different time windows
	windows := []string{"5m", "15m", "30m", "1h", "4h", "1d"}
	for _, window := range windows {
		impact := s.calculateImpact(targetEvent, symbol, window)
		if err := s.repo.SaveImpactRecord(ctx, impact); err != nil {
			s.logger.Warn("failed to save impact record",
				"event", eventID, "window", window, "error", err)
		}
	}

	return nil
}

// calculateSurprise calculates surprise score (sigma)
func (s *CalendarService) calculateSurprise(actual, forecast, previous string) float64 {
	actualVal := s.parseValue(actual)
	forecastVal := s.parseValue(forecast)

	if forecastVal == 0 {
		return 0
	}

	// Percentage surprise
	diff := (actualVal - forecastVal) / forecastVal * 100
	return diff
}

// calculateImpactDirection determines if impact is bullish (-1), bearish (1), or neutral (0)
func (s *CalendarService) calculateImpactDirection(e apiclient.CalendarEvent) int16 {
	// This is a simplified logic - in production, use event-specific logic
	if e.Actual == "" || e.Forecast == "" {
		return 0
	}

	actual := s.parseValue(e.Actual)
	forecast := s.parseValue(e.Forecast)

	// Higher than expected
	if actual > forecast {
		// For most economic indicators, better than expected is bullish for currency
		return 1
	}
	
	return -1
}

// classifySurprise classifies surprise magnitude
func (s *CalendarService) classifySurprise(surprise float64) string {
	absSurprise := math.Abs(surprise)
	switch {
	case absSurprise >= 3:
		return "EXTREME"
	case absSurprise >= 2:
		return "HIGH"
	case absSurprise >= 1:
		return "MODERATE"
	default:
		return "NORMAL"
	}
}

// parseEventTime parses event time from CalendarEvent.
// Prefers ReleaseDate (Unix milliseconds), falls back to FullDate.
func (s *CalendarService) parseEventTime(e apiclient.CalendarEvent) (time.Time, error) {
	// Prefer ReleaseDate (Unix milliseconds)
	if e.ReleaseDate > 0 {
		return time.UnixMilli(e.ReleaseDate).UTC(), nil
	}

	// Fallback to FullDate: format is "2026-05-11T14:00:00" (no timezone)
	if e.ScheduledAt != "" {
		// Parse in UTC since MQL5 FullDate is UTC
		t, err := time.ParseInLocation("2006-01-02T15:04:05", e.ScheduledAt, time.UTC)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid FullDate format: %w", err)
		}
		return t, nil
	}

	return time.Time{}, fmt.Errorf("both ReleaseDate and ScheduledAt are empty")
}

// generateEventID generates unique event ID
func (s *CalendarService) generateEventID(e apiclient.CalendarEvent) string {
	// Parse event time for ID
	t, err := s.parseEventTime(e)
	if err != nil {
		// Fallback to current time if parsing fails (should not happen after parseEventTime check)
		t = time.Now()
	}
	return fmt.Sprintf("%s_%s_%d",
		e.Currency,
		strings.ReplaceAll(strings.ToLower(e.Title), " ", "_"),
		t.Unix())
}

// truncate truncates string to max length
func (s *CalendarService) truncate(val string, maxLen int) string {
	if len(val) > maxLen {
		return val[:maxLen]
	}
	return val
}

// parseValue parses string value to float
func (s *CalendarService) parseValue(val string) float64 {
	val = strings.TrimSpace(val)
	val = strings.ReplaceAll(val, "K", "")
	val = strings.ReplaceAll(val, "M", "")
	val = strings.ReplaceAll(val, "B", "")
	val = strings.ReplaceAll(val, "%", "")
	val = strings.ReplaceAll(val, ",", "")

	if val == "" {
		return 0
	}

	v, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0
	}
	return v
}

// calculateImpact calculates price impact for a time window
func (s *CalendarService) calculateImpact(event *postgres.CalendarEvent, symbol, window string) postgres.EventImpactRecord {
	// In production, fetch actual price data from price service
	// For now, return placeholder
	return postgres.EventImpactRecord{
		EventID:    event.EventID,
		Window:     window,
		Symbol:     symbol,
		PriceBefore: 0,
		PriceAfter:  0,
		PctChange:   0,
		RecordedAt:  time.Now(),
	}
}

// cacheEvents caches events in Redis
func (s *CalendarService) cacheEvents(ctx context.Context, events []postgres.CalendarEvent) {
	// Cache by date
	cacheKey := fmt.Sprintf("calendar:date:%s", time.Now().Format("2006-01-02"))
	s.redis.SetJSON(ctx, cacheKey, events, 1*time.Hour)
}

// SyncResult represents sync operation result
type SyncResult struct {
	Inserted int
	Updated  int
	Total    int
}
