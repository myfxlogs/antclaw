package trend

import (
	"context"
	"testing"

	alfqv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/internal/infra/postgres"
)

// fakeTrendRepo implements TrendRepository with canned test data.
type fakeTrendRepo struct {
	topics  []*postgres.TrendingTopicRow
	symbols []*postgres.HotSymbolRow
}

func newFakeTrendRepo() *fakeTrendRepo {
	return &fakeTrendRepo{}
}

func (f *fakeTrendRepo) ListTrendingTopics(_ context.Context, window string, limit int32) ([]*postgres.TrendingTopicRow, error) {
	var out []*postgres.TrendingTopicRow
	for _, t := range f.topics {
		if int32(len(out)) >= limit {
			break
		}
		out = append(out, t)
	}
	return out, nil
}

func (f *fakeTrendRepo) ListHotSymbols(_ context.Context, window string, limit int32) ([]*postgres.HotSymbolRow, error) {
	var out []*postgres.HotSymbolRow
	for _, s := range f.symbols {
		if int32(len(out)) >= limit {
			break
		}
		out = append(out, s)
	}
	return out, nil
}

// ----- Tests -----

func TestTrendingTopics_InvalidWindow(t *testing.T) {
	svc := NewService(newFakeTrendRepo())
	_, err := svc.ListTrendingTopics(context.Background(), &alfqv1.ListTrendingTopicsRequest{Window: "3d"})
	if err == nil {
		t.Fatal("expected InvalidArgument for invalid window")
	}
}

func TestTrendingTopics_DefaultWindow(t *testing.T) {
	repo := newFakeTrendRepo()
	repo.topics = []*postgres.TrendingTopicRow{
		{Topic: "#BTC", PostCount: 10, EngagementCount: 50},
	}
	svc := NewService(repo)
	resp, err := svc.ListTrendingTopics(context.Background(), &alfqv1.ListTrendingTopicsRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Topics) != 1 {
		t.Fatalf("expected 1 topic, got %d", len(resp.Topics))
	}
}

func TestTrendingTopics_ValidWindows(t *testing.T) {
	for _, w := range []string{"1h", "24h", "7d"} {
		repo := newFakeTrendRepo()
		repo.topics = []*postgres.TrendingTopicRow{
			{Topic: "#" + w, PostCount: 1, EngagementCount: 1},
		}
		svc := NewService(repo)
		resp, err := svc.ListTrendingTopics(context.Background(), &alfqv1.ListTrendingTopicsRequest{Window: w})
		if err != nil {
			t.Fatalf("unexpected error for window %s: %v", w, err)
		}
		if len(resp.Topics) != 1 {
			t.Fatalf("expected 1 topic for %s, got %d", w, len(resp.Topics))
		}
	}
}

func TestTrendingTopics_FieldsMapped(t *testing.T) {
	repo := newFakeTrendRepo()
	repo.topics = []*postgres.TrendingTopicRow{
		{Topic: "#ETH", PostCount: 25, EngagementCount: 120},
	}
	svc := NewService(repo)
	resp, err := svc.ListTrendingTopics(context.Background(), &alfqv1.ListTrendingTopicsRequest{Window: "24h"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Topics[0].Topic != "#ETH" {
		t.Fatalf("expected #ETH, got %s", resp.Topics[0].Topic)
	}
	if resp.Topics[0].PostCount != 25 {
		t.Fatalf("expected post_count 25, got %d", resp.Topics[0].PostCount)
	}
	if resp.Topics[0].EngagementCount != 120 {
		t.Fatalf("expected engagement 120, got %d", resp.Topics[0].EngagementCount)
	}
}

func TestHotSymbols_InvalidWindow(t *testing.T) {
	svc := NewService(newFakeTrendRepo())
	_, err := svc.ListHotSymbols(context.Background(), &alfqv1.ListHotSymbolsRequest{Window: "2d"})
	if err == nil {
		t.Fatal("expected InvalidArgument for invalid window")
	}
}

func TestHotSymbols_DefaultWindow(t *testing.T) {
	repo := newFakeTrendRepo()
	repo.symbols = []*postgres.HotSymbolRow{
		{Symbol: "EURUSD", PostCount: 8, SignalCount: 3, EngagementCount: 40},
	}
	svc := NewService(repo)
	resp, err := svc.ListHotSymbols(context.Background(), &alfqv1.ListHotSymbolsRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Symbols) != 1 {
		t.Fatalf("expected 1 symbol, got %d", len(resp.Symbols))
	}
}

func TestHotSymbols_FieldsMapped(t *testing.T) {
	repo := newFakeTrendRepo()
	repo.symbols = []*postgres.HotSymbolRow{
		{Symbol: "XAUUSD", PostCount: 15, SignalCount: 7, EngagementCount: 85},
	}
	svc := NewService(repo)
	resp, err := svc.ListHotSymbols(context.Background(), &alfqv1.ListHotSymbolsRequest{Window: "7d"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := resp.Symbols[0]
	if s.Symbol != "XAUUSD" {
		t.Fatalf("expected XAUUSD, got %s", s.Symbol)
	}
	if s.PostCount != 15 {
		t.Fatalf("expected post_count 15, got %d", s.PostCount)
	}
	if s.SignalCount != 7 {
		t.Fatalf("expected signal_count 7, got %d", s.SignalCount)
	}
	if s.EngagementCount != 85 {
		t.Fatalf("expected engagement 85, got %d", s.EngagementCount)
	}
}

func TestHotSymbols_LimitClamping(t *testing.T) {
	repo := newFakeTrendRepo()
	for i := 0; i < 60; i++ {
		repo.symbols = append(repo.symbols, &postgres.HotSymbolRow{
			Symbol: "SYM", PostCount: 1, SignalCount: 0, EngagementCount: 0,
		})
	}
	svc := NewService(repo)
	resp, err := svc.ListHotSymbols(context.Background(), &alfqv1.ListHotSymbolsRequest{Window: "24h", Limit: 100})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Symbols) > 50 {
		t.Fatalf("expected max 50 symbols, got %d", len(resp.Symbols))
	}
}
