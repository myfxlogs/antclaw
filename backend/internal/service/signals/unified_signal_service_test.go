package signals

import (
	"context"
	"testing"
	"time"

	"github.com/antclaw/antclaw/internal/domain/shared"
	"github.com/antclaw/antclaw/internal/infra/postgres"
)

// stubPriceRepo implements postgres.PriceRepository for testing.
type stubPriceRepo struct {
	bars []postgres.DailyBar
	err  error
}

func (s *stubPriceRepo) GetDailyBars(ctx context.Context, symbol string, from, to time.Time) ([]postgres.DailyBar, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.bars, nil
}
func (s *stubPriceRepo) GetLatest(ctx context.Context, symbol string) (*postgres.DailyBar, error) { return nil, nil }
func (s *stubPriceRepo) GetIntradayBars(ctx context.Context, symbol, interval string, from, to time.Time) ([]postgres.IntradayBar, error) {
	return nil, nil
}
func (s *stubPriceRepo) GetLatestIntraday(ctx context.Context, symbol, interval string) (*postgres.IntradayBar, error) {
	return nil, nil
}
func (s *stubPriceRepo) UpsertDailyBars(ctx context.Context, bars []postgres.DailyBar) error  { return nil }
func (s *stubPriceRepo) UpsertIntradayBars(ctx context.Context, bars []postgres.IntradayBar) error { return nil }

// stubMacroRepo implements postgres.MacroRepository for testing.
type stubMacroRepo struct {
	snapshots []postgres.RegimeSnapshot
	err       error
}

func (s *stubMacroRepo) GetRegimeHistory(ctx context.Context, from, to time.Time) ([]postgres.RegimeSnapshot, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.snapshots, nil
}
func (s *stubMacroRepo) SaveObservations(ctx context.Context, obs []postgres.MacroObservation) (int, error) { return 0, nil }
func (s *stubMacroRepo) GetLatest(ctx context.Context, seriesID string) (*postgres.MacroObservation, error)   { return nil, nil }
func (s *stubMacroRepo) GetHistory(ctx context.Context, seriesID string, from, to time.Time) ([]postgres.MacroObservation, error) {
	return nil, nil
}
func (s *stubMacroRepo) SaveRegime(ctx context.Context, snapshot postgres.RegimeSnapshot) error { return nil }

func TestGetTAScoreInsufficientBars(t *testing.T) {
	svc := &UnifiedSignalService{
		priceRepo: &stubPriceRepo{bars: nil},
	}
	_, err := svc.getTAScore(context.Background(), "EURUSD")
	if err == nil {
		t.Fatal("expected error for insufficient bars, got nil")
	}
}

func TestGetTAScoreBullish(t *testing.T) {
	// Build 20 daily bars with a clear uptrend (last 5 SMA >> last 20 SMA)
	now := time.Now().UTC()
	bars := make([]postgres.DailyBar, 25)
	for i := 0; i < 25; i++ {
		bars[i] = postgres.DailyBar{
			Time:   now.Add(-time.Duration(25-i) * 24 * time.Hour),
			Close:  1.0 + float64(i)*0.01, // steady uptrend
		}
	}
	svc := &UnifiedSignalService{
		priceRepo: &stubPriceRepo{bars: bars},
	}
	score, err := svc.getTAScore(context.Background(), "EURUSD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !score.Available {
		t.Fatal("expected TA score to be available")
	}
	if score.Direction != shared.DirectionBullish {
		t.Fatalf("expected bullish direction, got %s", score.Direction)
	}
	if score.Confidence <= 0 {
		t.Fatal("expected positive confidence")
	}
}

func TestGetTAScoreBearish(t *testing.T) {
	now := time.Now().UTC()
	bars := make([]postgres.DailyBar, 25)
	for i := 0; i < 25; i++ {
		bars[i] = postgres.DailyBar{
			Time:  now.Add(-time.Duration(25-i) * 24 * time.Hour),
			Close: 1.5 - float64(i)*0.01, // steady downtrend
		}
	}
	svc := &UnifiedSignalService{
		priceRepo: &stubPriceRepo{bars: bars},
	}
	score, err := svc.getTAScore(context.Background(), "EURUSD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if score.Direction != shared.DirectionBearish {
		t.Fatalf("expected bearish direction, got %s", score.Direction)
	}
}

func TestGetMacroScoreRiskOn(t *testing.T) {
	svc := &UnifiedSignalService{
		macroRepo: &stubMacroRepo{
			snapshots: []postgres.RegimeSnapshot{
				{Time: time.Now(), Regime: "risk_on", Score: 0.85},
			},
		},
	}
	score, err := svc.getMacroScore(context.Background(), "EURUSD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !score.Available {
		t.Fatal("expected macro score to be available")
	}
	if score.Direction != shared.DirectionBullish {
		t.Fatalf("expected bullish for risk_on, got %s", score.Direction)
	}
}

func TestGetMacroScoreRiskOff(t *testing.T) {
	svc := &UnifiedSignalService{
		macroRepo: &stubMacroRepo{
			snapshots: []postgres.RegimeSnapshot{
				{Time: time.Now(), Regime: "risk_off", Score: 0.9},
			},
		},
	}
	score, err := svc.getMacroScore(context.Background(), "EURUSD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if score.Direction != shared.DirectionBearish {
		t.Fatalf("expected bearish for risk_off, got %s", score.Direction)
	}
}

func TestGetMacroScoreEmpty(t *testing.T) {
	svc := &UnifiedSignalService{
		macroRepo: &stubMacroRepo{snapshots: nil},
	}
	_, err := svc.getMacroScore(context.Background(), "EURUSD")
	if err == nil {
		t.Fatal("expected error for empty regime history")
	}
}

func TestSMABasic(t *testing.T) {
	vals := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	got := sma(vals, 5)
	expected := (6.0 + 7 + 8 + 9 + 10) / 5.0
	if got != expected {
		t.Fatalf("SMA(5)=%v, want %v", got, expected)
	}
}

func TestSMAInsufficient(t *testing.T) {
	if v := sma([]float64{1}, 5); v != 1 {
		t.Fatalf("SMA(%d)=%v, want 1", 1, v)
	}
}

func TestDetermineDirection(t *testing.T) {
	tests := []struct {
		score float64
		want  shared.Direction
	}{
		{0.5, shared.DirectionBullish},
		{0.15, shared.DirectionBullish},
		{0.05, shared.DirectionNeutral},
		{-0.05, shared.DirectionNeutral},
		{-0.15, shared.DirectionBearish},
		{-0.5, shared.DirectionBearish},
	}
	svc := &UnifiedSignalService{}
	for _, tt := range tests {
		got := svc.determineDirection(tt.score)
		if got != tt.want {
			t.Errorf("determineDirection(%v)=%s, want %s", tt.score, got, tt.want)
		}
	}
}

func TestGetSignalHistoryNoPool(t *testing.T) {
	svc := &UnifiedSignalService{}
	_, err := svc.GetSignalHistory(context.Background(), "EURUSD", 30)
	if err == nil {
		t.Fatal("expected error when pool is nil")
	}
}
