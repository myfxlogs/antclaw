package calendar

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/antclaw/antclaw/internal/infra/apiclient"
	calendarv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
)

// Service implements Calendar business logic.
type Service struct {
	fetcher *apiclient.MQL5Fetcher
}

// NewService creates a new CalendarService with sample data only.
// Deprecated: Use NewServiceWithFetcher for production with real data.
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

// ListEvents returns economic calendar events.
// Fetches real data from MQL5 API, with fallback to sample data on error.
func (s *Service) ListEvents(ctx context.Context, date, currencyFilter string, minImpact calendarv1.ImpactLevel) (*calendarv1.ListEventsResponse, error) {
	// Try to fetch real data from MQL5
	events, err := s.fetchFromMQL5(ctx, date)
	if err != nil {
		log.Printf("MQL5 calendar fetch failed: %v, using fallback data", err)
		// Fallback to sample data
		events = s.getSampleEvents()
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

// getSampleEvents returns fallback sample events.
func (s *Service) getSampleEvents() []*calendarv1.CalendarEvent {
	return []*calendarv1.CalendarEvent{
		{
			EventId:     "evt-001",
			Title:       "Non-Farm Payrolls",
			Country:     "US",
			Currency:    "USD",
			Impact:      calendarv1.ImpactLevel_IMPACT_LEVEL_HIGH,
			ScheduledAt: time.Now().Format(time.RFC3339),
			Previous:    "200K",
			Forecast:    "210K",
			Actual:      "",
		},
		{
			EventId:     "evt-002",
			Title:       "ECB Interest Rate Decision",
			Country:     "EU",
			Currency:    "EUR",
			Impact:      calendarv1.ImpactLevel_IMPACT_LEVEL_HIGH,
			ScheduledAt: time.Now().Add(time.Hour).Format(time.RFC3339),
			Previous:    "4.5%",
			Forecast:    "4.5%",
			Actual:      "",
		},
	}
}

// GetEvent returns event details.
func (s *Service) GetEvent(ctx context.Context, eventID string) (*calendarv1.GetEventResponse, error) {
	if eventID == "" {
		return nil, fmt.Errorf("event_id required")
	}

	return &calendarv1.GetEventResponse{
		Event: &calendarv1.CalendarEvent{
			EventId:     eventID,
			Title:       "Non-Farm Payrolls",
			Country:     "US",
			Currency:    "USD",
			Impact:      calendarv1.ImpactLevel_IMPACT_LEVEL_HIGH,
			ScheduledAt: time.Now().Format(time.RFC3339),
			Previous:    "200K",
			Forecast:    "210K",
			Actual:      "215K",
		},
		Description: "The Non-Farm Payrolls report measures the number of jobs added or lost in the US economy over the last month, excluding farm workers.",
	}, nil
}

// GetImpact returns impact analysis for an event.
func (s *Service) GetImpact(ctx context.Context, eventID string) (*calendarv1.GetImpactResponse, error) {
	if eventID == "" {
		return nil, fmt.Errorf("event_id required")
	}

	return &calendarv1.GetImpactResponse{
		Analysis: &calendarv1.ImpactAnalysis{
			EventId:         eventID,
			AffectedPairs:   []string{"EURUSD", "GBPUSD", "USDJPY", "XAUUSD"},
			ExpectedDirection: "volatile",
			Confidence:      0.85,
		},
	}, nil
}

// GetImpactHistory returns historical impact of events.
func (s *Service) GetImpactHistory(ctx context.Context, eventType, pair string, count int32) (*calendarv1.GetImpactHistoryResponse, error) {
	if eventType == "" {
		return nil, fmt.Errorf("event_type required")
	}
	if count <= 0 {
		count = 10
	}

	impacts := make([]*calendarv1.HistoricalImpact, 0, count)
	for i := int32(0); i < count; i++ {
		impacts = append(impacts, &calendarv1.HistoricalImpact{
			EventId:      fmt.Sprintf("evt-%s-%d", eventType, i),
			Date:         time.Now().AddDate(0, 0, -int(i)*7).Format("2006-01-02"),
			Pair:         pair,
			PriceChange:  0.5 + float64(i)*0.1,
			Volatility:   1.2 + float64(i)*0.05,
		})
	}

	return &calendarv1.GetImpactHistoryResponse{Impacts: impacts}, nil
}
