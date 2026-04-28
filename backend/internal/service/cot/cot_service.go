package cot

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/antclaw/antclaw/internal/domain/shared"
	"github.com/antclaw/antclaw/internal/infra/apiclient"
	"github.com/antclaw/antclaw/internal/infra/postgres"
)

// COTService provides COT analysis operations
type COTService struct {
	repo      postgres.COTRepository
	client    *apiclient.CFTCClient
	logger    *slog.Logger
}

// COTAnalysisResult represents COT analysis for a currency
type COTAnalysisResult struct {
	Currency        string                `json:"currency"`
	ReportDate      time.Time             `json:"report_date"`
	NetPosition     int64                 `json:"net_position"`
	COTIndex        float64               `json:"cot_index"`
	Direction       shared.Direction      `json:"direction"`
	SentimentScore  float64               `json:"sentiment_score"`
	WoWChange       int64                 `json:"wow_change"`
	ZScore          float64               `json:"zscore"`
	Percentile      float64               `json:"percentile"`
	SignalStrength  string                `json:"signal_strength"`
}

// NewCOTService creates a new COT service
func NewCOTService(repo postgres.COTRepository, apiKey string, logger *slog.Logger) *COTService {
	return &COTService{
		repo:   repo,
		client: apiclient.NewCFTCClient(apiKey),
		logger: logger,
	}
}

// SyncLatest fetches and syncs latest COT reports
func (s *COTService) SyncLatest(ctx context.Context) (*SyncResult, error) {
	// Fetch for major currencies
	currencies := []string{"EUR", "GBP", "JPY", "CHF", "AUD", "CAD", "NZD", "MXN", "BRL"}
	
	var totalInserted int
	var analyses []postgres.COTAnalysis

	for _, curr := range currencies {
		reports, err := s.client.FetchAllForCurrency(ctx, curr, 26) // 6 months
		if err != nil {
			s.logger.Warn("failed to fetch COT", "currency", curr, "error", err)
			continue
		}

		// Convert to internal records
		var records []postgres.COTRecord
		for _, r := range reports {
			date, _ := apiclient.ParseReportDate(r.ReportDateAsOf)
			record := postgres.COTRecord{
				ReportDate:       date,
				ContractCode:     r.CftcContractMarketCode,
				Currency:         curr,
				NoncommLong:      r.NoncommPositionsLongAll,
				NoncommShort:     r.NoncommPositionsShortAll,
				CommLong:         r.CommPositionsLongAll,
				CommShort:        r.CommPositionsShortAll,
				TotalOI:          r.OpenInterestAll,
			}
			records = append(records, record)
		}

		// Upsert records
		count, err := s.repo.UpsertRecords(ctx, records)
		if err != nil {
			s.logger.Error("failed to upsert COT records", "currency", curr, "error", err)
			continue
		}
		totalInserted += count

		// Generate analysis for latest report
		if len(records) > 0 {
			latest := records[len(records)-1]
			history := s.extractHistory(records, curr)
			analysis := s.analyze(latest, history)
			
			analyses = append(analyses, postgres.COTAnalysis{
				ReportDate:     latest.ReportDate,
				ContractCode:   latest.ContractCode,
				NetPosition:    analysis.NetPosition,
				COTIndex:       analysis.COTIndex,
				Direction:      string(analysis.Direction),
				SentimentScore: analysis.SentimentScore,
				WoWChange:      analysis.WoWChange,
				ZScore:         analysis.ZScore,
				Percentile:     analysis.Percentile,
			})
		}
	}

	// Save analyses
	if err := s.repo.SaveAnalysis(ctx, analyses); err != nil {
		s.logger.Error("failed to save analyses", "error", err)
	}

	s.logger.Info("COT sync completed", "inserted", totalInserted, "analyses", len(analyses))
	
	return &SyncResult{
		Inserted: totalInserted,
		Analyses: len(analyses),
	}, nil
}

// GetAnalysis returns COT analysis for a currency
func (s *COTService) GetAnalysis(ctx context.Context, currency string) (*COTAnalysisResult, error) {
	// Get latest from database
	allAnalyses, err := s.repo.GetLatestAll(ctx)
	if err != nil {
		return nil, err
	}

	analysis, ok := allAnalyses[currency]
	if !ok {
		return nil, fmt.Errorf("no COT data for currency: %s", currency)
	}

	// Get historical for additional context
	history, err := s.repo.GetHistory(ctx, analysis.ContractCode, 52)
	if err != nil {
		s.logger.Warn("failed to get history", "currency", currency, "error", err)
	}

	// Get 52-week high/low for normalization
	high, low := s.getHighLow(history)

	result := &COTAnalysisResult{
		Currency:       analysis.ContractCode,
		ReportDate:     analysis.ReportDate,
		NetPosition:    analysis.NetPosition,
		COTIndex:       analysis.COTIndex,
		Direction:      shared.Direction(analysis.Direction),
		SentimentScore: analysis.SentimentScore,
		WoWChange:      analysis.WoWChange,
		ZScore:         analysis.ZScore,
		Percentile:     analysis.Percentile,
	}

	// Calculate signal strength
	result.SignalStrength = s.calculateSignalStrength(result, high, low)

	return result, nil
}

// GetAllAnalysis returns COT analysis for all currencies
func (s *COTService) GetAllAnalysis(ctx context.Context) (map[string]*COTAnalysisResult, error) {
	analyses, err := s.repo.GetLatestAll(ctx)
	if err != nil {
		return nil, err
	}

	results := make(map[string]*COTAnalysisResult)
	for code, analysis := range analyses {
		results[code] = &COTAnalysisResult{
			Currency:       code,
			ReportDate:     analysis.ReportDate,
			NetPosition:    analysis.NetPosition,
			COTIndex:       analysis.COTIndex,
			Direction:      shared.Direction(analysis.Direction),
			SentimentScore: analysis.SentimentScore,
			WoWChange:      analysis.WoWChange,
			ZScore:         analysis.ZScore,
			Percentile:     analysis.Percentile,
		}
	}

	return results, nil
}

// analyze performs COT analysis on a record
func (s *COTService) analyze(record postgres.COTRecord, history []int64) postgres.COTAnalysis {
	// Calculate net position
	netPos := record.NoncommLong - record.NoncommShort

	// Calculate COT Index (0-100)
	var cotIndex float64
	if len(history) >= 3 {
		high, low := s.getHighLowValues(history)
		if high != low {
			cotIndex = float64(netPos-low) / float64(high-low) * 100
		}
	}

	// Calculate WoW change
	var wowChange int64
	if len(history) >= 2 {
		wowChange = netPos - history[len(history)-2]
	}

	// Calculate Z-Score
	zscore := s.calculateZScore(float64(netPos), history)

	// Calculate percentile
	percentile := s.calculatePercentile(float64(netPos), history)

	// Determine direction
	direction := shared.DirectionNeutral
	if cotIndex > 70 {
		direction = shared.DirectionBullish
	} else if cotIndex < 30 {
		direction = shared.DirectionBearish
	}

	// Calculate sentiment score (-1 to 1)
	sentiment := (cotIndex - 50) / 50

	return postgres.COTAnalysis{
		ReportDate:     record.ReportDate,
		ContractCode:   record.ContractCode,
		NetPosition:    netPos,
		COTIndex:       cotIndex,
		Direction:      string(direction),
		SentimentScore: sentiment,
		WoWChange:      wowChange,
		ZScore:         zscore,
		Percentile:     percentile,
	}
}

// calculateZScore calculates Z-score
func (s *COTService) calculateZScore(value float64, history []int64) float64 {
	if len(history) < 2 {
		return 0
	}

	mean := s.calculateMean(history)
	std := s.calculateStdDev(history, mean)
	
	if std == 0 {
		return 0
	}
	
	return (value - mean) / std
}

// calculatePercentile calculates percentile rank
func (s *COTService) calculatePercentile(value float64, history []int64) float64 {
	if len(history) == 0 {
		return 50
	}

	lessThan := 0
	for _, h := range history {
		if float64(h) < value {
			lessThan++
		}
	}

	return float64(lessThan) / float64(len(history)) * 100
}

// calculateMean calculates mean
func (s *COTService) calculateMean(values []int64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	var sum int64
	for _, v := range values {
		sum += v
	}
	return float64(sum) / float64(len(values))
}

// calculateStdDev calculates standard deviation
func (s *COTService) calculateStdDev(values []int64, mean float64) float64 {
	if len(values) < 2 {
		return 0
	}

	var sumSquared float64
	for _, v := range values {
		diff := float64(v) - mean
		sumSquared += diff * diff
	}
	
	variance := sumSquared / float64(len(values)-1)
	return math.Sqrt(variance)
}

// getHighLow gets 52-week high and low from history
func (s *COTService) getHighLow(history []postgres.COTRecord) (int64, int64) {
	if len(history) == 0 {
		return 0, 0
	}

	var values []int64
	for _, h := range history {
		netPos := h.NoncommLong - h.NoncommShort
		values = append(values, netPos)
	}

	return s.getHighLowValues(values)
}

// getHighLowValues gets high/low from values
func (s *COTService) getHighLowValues(values []int64) (int64, int64) {
	if len(values) == 0 {
		return 0, 0
	}

	high := values[0]
	low := values[0]

	for _, v := range values[1:] {
		if v > high {
			high = v
		}
		if v < low {
			low = v
		}
	}

	return high, low
}

// extractHistory extracts net position history from records
func (s *COTService) extractHistory(records []postgres.COTRecord, currency string) []int64 {
	var history []int64
	for _, r := range records {
		if r.Currency == currency {
			netPos := r.NoncommLong - r.NoncommShort
			history = append(history, netPos)
		}
	}
	return history
}

// calculateSignalStrength calculates signal strength based on multiple factors
func (s *COTService) calculateSignalStrength(analysis *COTAnalysisResult, high, low int64) string {
	score := 0

	// COT Index score
	if analysis.COTIndex > 80 || analysis.COTIndex < 20 {
		score += 3
	} else if analysis.COTIndex > 70 || analysis.COTIndex < 30 {
		score += 2
	} else if analysis.COTIndex > 60 || analysis.COTIndex < 40 {
		score += 1
	}

	// Z-Score
	if math.Abs(analysis.ZScore) > 2 {
		score += 2
	} else if math.Abs(analysis.ZScore) > 1 {
		score += 1
	}

	// WoW change
	if math.Abs(float64(analysis.WoWChange)) > float64(high-low)*0.1 {
		score += 1
	}

	switch {
	case score >= 5:
		return "STRONG"
	case score >= 3:
		return "MODERATE"
	case score >= 1:
		return "WEAK"
	default:
		return "NONE"
	}
}

// SyncResult represents sync operation result
type SyncResult struct {
	Inserted int
	Analyses int
}
