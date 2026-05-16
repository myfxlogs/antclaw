package calendar

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/antclaw/antclaw/internal/domain/apperror"

	"github.com/antclaw/antclaw/internal/infra/apiclient"
	calendarv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
)

// Service implements Calendar business logic.
type Service struct {
	fetcher *apiclient.MQL5Fetcher
}

// NewService creates a new CalendarService for production.
// Deprecated: Use NewServiceWithFetcher for explicit fetcher configuration.
func NewService() *Service {
	return NewServiceWithFetcher(nil)
}

// NewServiceWithFetcher creates a new CalendarService with MQL5 fetcher.
func NewServiceWithFetcher(fetcher *apiclient.MQL5Fetcher) *Service {
	if fetcher == nil {
		fetcher = apiclient.NewMQL5Fetcher()
	}
	return &Service{fetcher: fetcher}
}

// ListEvents returns economic calendar events from MQL5 API.
func (s *Service) ListEvents(ctx context.Context, date, currencyFilter string, minImpact calendarv1.ImpactLevel) (*calendarv1.ListEventsResponse, error) {
	events, err := s.fetchFromMQL5(ctx, date)
	if err != nil {
		log.Printf("MQL5 calendar fetch failed: %v", err)
		return nil, fmt.Errorf("%w: %v", apperror.ErrUpstreamUnavailable, err)
	}

	// Apply filters
	filtered := make([]*calendarv1.CalendarEvent, 0)
	for _, e := range events {
		if currencyFilter != "" && e.Currency != currencyFilter {
			continue
		}
		if minImpact > 0 && e.Impact < minImpact {
			continue
		}
		filtered = append(filtered, e)
	}

	return &calendarv1.ListEventsResponse{Events: filtered}, nil
}

// fetchFromMQL5 fetches real calendar data from MQL5 API.
func (s *Service) fetchFromMQL5(ctx context.Context, date string) ([]*calendarv1.CalendarEvent, error) {
	// Parse date if provided
	var from, to time.Time
	if date != "" {
		t, err := time.Parse("2006-01-02", date)
		if err == nil {
			from = t
			to = t.AddDate(0, 0, 1)
		}
	}

	// Default to this week if no date specified
	if from.IsZero() {
		return s.convertEvents(s.fetcher.FetchWeek(ctx, "this"))
	}

	return s.convertEvents(s.fetcher.FetchDateRange(ctx, from, to))
}

// convertEvents converts apiclient events to proto events.
func (s *Service) convertEvents(apiEvents []apiclient.CalendarEvent, err error) ([]*calendarv1.CalendarEvent, error) {
	if err != nil {
		return nil, err
	}

	events := make([]*calendarv1.CalendarEvent, len(apiEvents))
	for i, e := range apiEvents {
		impact := calendarv1.ImpactLevel_IMPACT_LEVEL_LOW
		switch e.Impact {
		case "high":
			impact = calendarv1.ImpactLevel_IMPACT_LEVEL_HIGH
		case "medium":
			impact = calendarv1.ImpactLevel_IMPACT_LEVEL_MEDIUM
		}
		events[i] = &calendarv1.CalendarEvent{
			EventId:     e.EventID,
			Title:       e.Title,
			Country:     e.Country,
			Currency:    e.Currency,
			Impact:      impact,
			ScheduledAt: e.ScheduledAt,
			Previous:    e.Previous,
			Forecast:    e.Forecast,
			Actual:      e.Actual,
		}
	}
	return events, nil
}

// GetEvent returns event details.
func (s *Service) GetEvent(ctx context.Context, eventID string) (*calendarv1.GetEventResponse, error) {
	if eventID == "" {
		return nil, fmt.Errorf("event_id required")
	}
	return nil, fmt.Errorf("%w: event %s", apperror.ErrNotFound, eventID)
}

// GetImpact returns impact analysis for an event.
func (s *Service) GetImpact(ctx context.Context, eventID string) (*calendarv1.GetImpactResponse, error) {
	if eventID == "" {
		return nil, fmt.Errorf("event_id required")
	}
	return nil, fmt.Errorf("%w: impact analysis for event %s", apperror.ErrNotFound, eventID)
}

// GetImpactHistory returns historical impact of events.
func (s *Service) GetImpactHistory(ctx context.Context, eventType, pair string, count int32) (*calendarv1.GetImpactHistoryResponse, error) {
	if eventType == "" {
		return nil, fmt.Errorf("event_type required")
	}
	return nil, fmt.Errorf("%w: impact history not available", apperror.ErrDataInsufficient)
}
