package cot

import (
	"context"
	"fmt"
	"time"

	cotv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
)

// Service implements COT business logic.
type Service struct {
	// repo would be injected here for real data access
}

func NewService() *Service {
	return &Service{}
}

// GetSummary returns latest COT data and history for a pair.
func (s *Service) GetSummary(ctx context.Context, pair string) (*cotv1.GetSummaryResponse, error) {
	if pair == "" {
		return nil, fmt.Errorf("pair required")
	}

	// MVP: return skeleton data
	latest := &cotv1.COTEntry{
		Pair:         pair,
		Date:         time.Now().Format("2006-01-02"),
		NonCommLong:  1000,
		NonCommShort: 500,
		CommLong:     2000,
		CommShort:    1500,
		NonRepLong:   300,
		NonRepShort:  200,
	}

	return &cotv1.GetSummaryResponse{
		Latest: latest,
		History: []*cotv1.COTEntry{
			latest,
		},
	}, nil
}

// Compare returns COT comparison between two dates.
func (s *Service) Compare(ctx context.Context, pair, dateA, dateB string) (*cotv1.CompareResponse, error) {
	if pair == "" || dateA == "" || dateB == "" {
		return nil, fmt.Errorf("pair, date_a and date_b required")
	}

	entryA := &cotv1.COTEntry{
		Pair:         pair,
		Date:         dateA,
		NonCommLong:  1000,
		NonCommShort: 500,
		CommLong:     2000,
		CommShort:    1500,
		NonRepLong:   300,
		NonRepShort:  200,
	}
	entryB := &cotv1.COTEntry{
		Pair:         pair,
		Date:         dateB,
		NonCommLong:  1100,
		NonCommShort: 450,
		CommLong:     2100,
		CommShort:    1400,
		NonRepLong:   350,
		NonRepShort:  150,
	}

	return &cotv1.CompareResponse{
		EntryA:        entryA,
		EntryB:        entryB,
		ChangeSummary: fmt.Sprintf("Net non-commercial changed from %d to %d", 
			entryA.NonCommLong-entryA.NonCommShort, 
			entryB.NonCommLong-entryB.NonCommShort),
	}, nil
}

// GetSignals returns COT-based trading signals.
func (s *Service) GetSignals(ctx context.Context, pairFilter string) (*cotv1.GetSignalsResponse, error) {
	signals := []*cotv1.COTSignal{
		{
			Pair:        "EURUSD",
			SignalType:  "extreme_positioning",
			Direction:   "bearish",
			Strength:    0.75,
			CreatedAt:   time.Now().Format(time.RFC3339),
		},
	}

	if pairFilter != "" {
		filtered := make([]*cotv1.COTSignal, 0)
		for _, s := range signals {
			if s.Pair == pairFilter {
				filtered = append(filtered, s)
			}
		}
		signals = filtered
	}

	return &cotv1.GetSignalsResponse{Signals: signals}, nil
}

// GetHistory returns COT historical data.
func (s *Service) GetHistory(ctx context.Context, pair string, limit int32) (*cotv1.COTServiceGetHistoryResponse, error) {
	if pair == "" {
		return nil, fmt.Errorf("pair required")
	}

	entries := make([]*cotv1.COTEntry, 0, limit)
	for i := int32(0); i < limit && i < 10; i++ {
		entries = append(entries, &cotv1.COTEntry{
			Pair:         pair,
			Date:         time.Now().AddDate(0, 0, -int(i)*7).Format("2006-01-02"),
			NonCommLong:  1000 + int64(i)*50,
			NonCommShort: 500 - int64(i)*20,
			CommLong:     2000 + int64(i)*30,
			CommShort:    1500 - int64(i)*15,
			NonRepLong:   300 + int64(i)*10,
			NonRepShort:  200 - int64(i)*5,
		})
	}

	return &cotv1.COTServiceGetHistoryResponse{Entries: entries}, nil
}

// SubscribePairAlert creates a COT alert subscription.
func (s *Service) SubscribePairAlert(ctx context.Context, pair string, threshold float64) (*cotv1.SubscribePairAlertResponse, error) {
	if pair == "" {
		return nil, fmt.Errorf("pair required")
	}
	if threshold <= 0 {
		return nil, fmt.Errorf("threshold must be positive")
	}

	return &cotv1.SubscribePairAlertResponse{
		SubscriptionId: fmt.Sprintf("cot-%s-%d", pair, time.Now().Unix()),
	}, nil
}
