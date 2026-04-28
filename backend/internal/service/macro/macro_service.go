package macro

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/antclaw/antclaw/internal/domain/shared"
	"github.com/antclaw/antclaw/internal/infra/apiclient"
	"github.com/antclaw/antclaw/internal/infra/postgres"
)

// MacroService provides macroeconomic data operations
type MacroService struct {
	repo      postgres.MacroRepository
	fred      *apiclient.FredClient
	bis       *apiclient.BISClient
	fedWatch  *apiclient.FedWatchClient
	logger    *slog.Logger
}

// MacroRegime represents macro regime classification
type MacroRegime struct {
	Regime      shared.Regime `json:"regime"`
	Score       float64       `json:"score"`
	Timestamp   time.Time     `json:"timestamp"`
	Components  RegimeComponents `json:"components"`
}

// RegimeComponents breaks down regime factors
type RegimeComponents struct {
	Growth      float64 `json:"growth"`
	Inflation   float64 `json:"inflation"`
	Liquidity   float64 `json:"liquidity"`
	Rates       float64 `json:"rates"`
}

// CompositeIndex represents a composite macro index
type CompositeIndex struct {
	Name        string    `json:"name"`
	Value       float64   `json:"value"`
	Change1M    float64   `json:"change_1m"`
	Change3M    float64   `json:"change_3m"`
	Timestamp   time.Time `json:"timestamp"`
}

// NewMacroService creates a new macro service
func NewMacroService(repo postgres.MacroRepository, fredKey string, logger *slog.Logger) *MacroService {
	return &MacroService{
		repo:     repo,
		fred:     apiclient.NewFredClient(fredKey),
		bis:      apiclient.NewBISClient(),
		fedWatch: apiclient.NewFedWatchClient(),
		logger:   logger,
	}
}

// SetFredKey updates the FRED API key dynamically.
// Called by CredentialResolver when the key is hot-reloaded.
func (s *MacroService) SetFredKey(key string) {
	s.fred.SetAPIKey(key)
}

// SyncFREDIndicators syncs FRED economic indicators
func (s *MacroService) SyncFREDIndicators(ctx context.Context, series []string) (*SyncResult, error) {
	var totalInserted int

	for _, seriesID := range series {
		resp, err := s.fred.FetchObservations(ctx, seriesID, 100)
		if err != nil {
			s.logger.Warn("failed to fetch FRED series", "series", seriesID, "error", err)
			continue
		}

		var macroObs []postgres.MacroObservation
		for _, o := range resp.Observations {
			date, _ := time.Parse("2006-01-02", o.Date)
			val, _ := strconv.ParseFloat(o.Value, 64)
			macroObs = append(macroObs, postgres.MacroObservation{
				Time:         date,
				Source:       "fred",
				SeriesID:     seriesID,
				ValueNumeric: &val,
				FetchedAt:    time.Now(),
			})
		}

		count, err := s.repo.SaveObservations(ctx, macroObs)
		if err != nil {
			s.logger.Error("failed to save observations", "series", seriesID, "error", err)
			continue
		}
		totalInserted += count
	}

	s.logger.Info("FRED sync completed", "inserted", totalInserted)
	return &SyncResult{Inserted: totalInserted}, nil
}

// CalculateRegime calculates current macro regime
func (s *MacroService) CalculateRegime(ctx context.Context) (*MacroRegime, error) {
	// Fetch key indicators
	indicators := s.fetchKeyIndicators(ctx)
	
	// Classify regime based on indicators
	regime := s.classifyRegime(indicators)
	
	// Save to database
	if err := s.repo.SaveRegime(ctx, postgres.RegimeSnapshot{
		Time:    time.Now(),
		Regime:  string(regime.Regime),
		Score:   regime.Score,
		Details: s.serializeComponents(regime.Components),
	}); err != nil {
		s.logger.Error("failed to save regime", "error", err)
	}

	return regime, nil
}

// GetCurrentRegime returns current macro regime
func (s *MacroService) GetCurrentRegime(ctx context.Context) (*MacroRegime, error) {
	// Get recent regime history
	history, err := s.repo.GetRegimeHistory(ctx, time.Now().AddDate(0, 0, -7), time.Now())
	if err != nil {
		return nil, err
	}

	if len(history) == 0 {
		// Calculate fresh
		return s.CalculateRegime(ctx)
	}

	latest := history[0]
	var components RegimeComponents
	s.deserializeComponents(latest.Details, &components)

	return &MacroRegime{
		Regime:     shared.Regime(latest.Regime),
		Score:      latest.Score,
		Timestamp:  latest.Time,
		Components: components,
	}, nil
}

// CalculateCompositeIndex calculates a composite economic index
func (s *MacroService) CalculateCompositeIndex(ctx context.Context, name string, weights map[string]float64) (*CompositeIndex, error) {
	var weightedSum, totalWeight float64
	
	for seriesID, weight := range weights {
		latest, err := s.repo.GetLatest(ctx, seriesID)
		if err != nil || latest == nil || latest.ValueNumeric == nil {
			s.logger.Warn("missing data for series", "series", seriesID)
			continue
		}
		
		// Normalize value (simplified - in production use proper normalization)
		value := *latest.ValueNumeric
		weightedSum += value * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return nil, fmt.Errorf("no data available for composite index")
	}

	index := &CompositeIndex{
		Name:      name,
		Value:     weightedSum / totalWeight,
		Timestamp: time.Now(),
	}

	// Calculate 1M and 3M changes (would need historical data in production)
	index.Change1M = 0
	index.Change3M = 0

	return index, nil
}

// FetchFedWatch fetches Fed probability data
func (s *MacroService) FetchFedWatch(ctx context.Context) (*apiclient.FedWatchData, error) {
	return s.fedWatch.GetNextMeetingProbability(ctx)
}

// fetchKeyIndicators fetches key macro indicators
func (s *MacroService) fetchKeyIndicators(ctx context.Context) map[string]float64 {
	indicators := make(map[string]float64)
	
	// GDP Growth
	if gdp, err := s.repo.GetLatest(ctx, "GDP"); err == nil && gdp != nil && gdp.ValueNumeric != nil {
		indicators["growth"] = *gdp.ValueNumeric
	}
	
	// Inflation (CPI)
	if cpi, err := s.repo.GetLatest(ctx, "CPIAUCSL"); err == nil && cpi != nil && cpi.ValueNumeric != nil {
		indicators["inflation"] = *cpi.ValueNumeric
	}
	
	// Unemployment
	if unemp, err := s.repo.GetLatest(ctx, "UNRATE"); err == nil && unemp != nil && unemp.ValueNumeric != nil {
		indicators["unemployment"] = *unemp.ValueNumeric
	}
	
	// Fed Funds Rate
	if fed, err := s.repo.GetLatest(ctx, "FEDFUNDS"); err == nil && fed != nil && fed.ValueNumeric != nil {
		indicators["rates"] = *fed.ValueNumeric
	}
	
	return indicators
}

// classifyRegime classifies macro regime
func (s *MacroService) classifyRegime(indicators map[string]float64) *MacroRegime {
	regime := &MacroRegime{
		Timestamp: time.Now(),
		Components: RegimeComponents{
			Growth:    s.getOrDefault(indicators, "growth", 2.0),
			Inflation: s.getOrDefault(indicators, "inflation", 2.0),
			Liquidity: s.getOrDefault(indicators, "liquidity", 5.0),
			Rates:     s.getOrDefault(indicators, "rates", 5.0),
		},
	}

	growth := regime.Components.Growth
	inflation := regime.Components.Inflation

	// Classify based on growth/inflation matrix
	switch {
	case growth > 2.5 && inflation < 2.5:
		regime.Regime = shared.RegimeGoldilocks
		regime.Score = 1.0
	case growth > 2.5 && inflation >= 2.5:
		regime.Regime = shared.RegimeInflationary
		regime.Score = -0.3
	case growth <= 2.5 && inflation < 2.5:
		regime.Regime = shared.RegimeDeflation
		regime.Score = -0.5
	case growth <= 2.5 && inflation >= 2.5:
		regime.Regime = shared.RegimeStagflation
		regime.Score = -0.7
	default:
		regime.Regime = shared.RegimeNeutral
		regime.Score = 0.0
	}

	return regime
}

// getOrDefault gets value from map or returns default
func (s *MacroService) getOrDefault(m map[string]float64, key string, defaultVal float64) float64 {
	if v, ok := m[key]; ok {
		return v
	}
	return defaultVal
}

// serializeComponents serializes components to bytes
func (s *MacroService) serializeComponents(c RegimeComponents) []byte {
	// Simplified - in production use proper JSON serialization
	return []byte(fmt.Sprintf("{\"growth\":%f,\"inflation\":%f,\"liquidity\":%f,\"rates\":%f}",
		c.Growth, c.Inflation, c.Liquidity, c.Rates))
}

// deserializeComponents deserializes components from bytes
func (s *MacroService) deserializeComponents(data []byte, c *RegimeComponents) {
	// Simplified - in production use proper JSON deserialization
}

// SyncResult represents sync operation result
type SyncResult struct {
	Inserted int
}
