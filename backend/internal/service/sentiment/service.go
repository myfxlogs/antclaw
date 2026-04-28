package sentiment

import (
	"context"
	"fmt"
	"time"

	sentv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
)

// Service implements Sentiment business logic.
type Service struct {
	sentimentCache map[string]*sentv1.SentimentData
	onchainCache   map[string][]*sentv1.OnchainMetric
	defiCache      map[string][]*sentv1.DefiMetric
	carryCache     map[string][]*sentv1.CarryData
}

// NewService creates a new SentimentService with sample data.
func NewService() *Service {
	svc := &Service{
		sentimentCache: make(map[string]*sentv1.SentimentData),
		onchainCache:   make(map[string][]*sentv1.OnchainMetric),
		defiCache:      make(map[string][]*sentv1.DefiMetric),
		carryCache:     make(map[string][]*sentv1.CarryData),
	}

	// Initialize sample sentiment data
	svc.generateSampleSentiment()
	svc.generateSampleOnchain()
	svc.generateSampleDefi()
	svc.generateSampleCarry()

	return svc
}

func (s *Service) generateSampleSentiment() {
	assets := []string{"BTC", "ETH", "EUR", "GBP", "GOLD"}
	sources := []string{"fear_greed", "social", "derivatives"}

	now := time.Now().Format(time.RFC3339)

	for _, asset := range assets {
		// Generate base sentiment score (-1 to 1)
		baseScore := (randFloat()-0.5)*2 + 0.2 // slight bullish bias
		label := sentimentLabel(baseScore)

		s.sentimentCache[asset] = &sentv1.SentimentData{
			Asset:     asset,
			Score:     baseScore,
			Label:     label,
			Source:    "composite",
			Timestamp: now,
		}

		// Components for this asset
		var components []*sentv1.SentimentData
		for _, src := range sources {
			var srcScore float64
			switch src {
			case "fear_greed":
				srcScore = baseScore + (randFloat()-0.5)*0.3
			case "social":
				srcScore = baseScore + (randFloat()-0.5)*0.5
			case "derivatives":
				srcScore = baseScore + (randFloat()-0.5)*0.2
			}
			components = append(components, &sentv1.SentimentData{
				Asset:     asset,
				Score:     srcScore,
				Label:     sentimentLabel(srcScore),
				Source:    src,
				Timestamp: now,
			})
		}
		s.onchainCache[asset+"_components"] = convertToOnchain(components)
	}
}

func (s *Service) generateSampleOnchain() {
	// Bitcoin onchain metrics
	s.onchainCache["BTC"] = []*sentv1.OnchainMetric{
		{Name: "active_addresses", Value: 890000, Trend: "rising"},
		{Name: "hash_rate", Value: 450.5, Trend: "stable"},
		{Name: "exchange_balance", Value: 2.3e6, Trend: "falling"},
		{Name: "sopr", Value: 1.02, Trend: "rising"},
		{Name: "nupl", Value: 0.45, Trend: "stable"},
		{Name: "mvrv", Value: 2.1, Trend: "rising"},
	}

	// Ethereum onchain metrics
	s.onchainCache["ETH"] = []*sentv1.OnchainMetric{
		{Name: "gas_usage", Value: 85.2, Trend: "stable"},
		{Name: "active_addresses", Value: 420000, Trend: "rising"},
		{Name: "defi_tvl", Value: 48.5e9, Trend: "rising"},
		{Name: "exchange_balance", Value: 18.5e6, Trend: "falling"},
		{Name: "validator_count", Value: 890000, Trend: "rising"},
		{Name: "staking_apr", Value: 3.8, Trend: "falling"},
	}
}

func (s *Service) generateSampleDefi() {
	chains := []string{"ethereum", "solana", "arbitrum"}

	protocols := map[string][]*sentv1.DefiMetric{
		"ethereum": {
			{Protocol: "Aave", Tvl: "12.5B", TvlChange_24H: "+2.3%", UtilizationRate: 0.72, HealthScore: "A"},
			{Protocol: "Uniswap", Tvl: "4.2B", TvlChange_24H: "+1.8%", UtilizationRate: 0.45, HealthScore: "A"},
			{Protocol: "Lido", Tvl: "25.8B", TvlChange_24H: "+0.5%", UtilizationRate: 0.91, HealthScore: "A+"},
			{Protocol: "MakerDAO", Tvl: "8.1B", TvlChange_24H: "-0.2%", UtilizationRate: 0.68, HealthScore: "A"},
		},
		"solana": {
			{Protocol: "Marinade", Tvl: "1.8B", TvlChange_24H: "+5.2%", UtilizationRate: 0.85, HealthScore: "B+"},
			{Protocol: "Jupiter", Tvl: "450M", TvlChange_24H: "+8.1%", UtilizationRate: 0.55, HealthScore: "A-"},
			{Protocol: "Raydium", Tvl: "380M", TvlChange_24H: "+3.5%", UtilizationRate: 0.62, HealthScore: "B+"},
		},
		"arbitrum": {
			{Protocol: "GMX", Tvl: "520M", TvlChange_24H: "+1.2%", UtilizationRate: 0.58, HealthScore: "A-"},
			{Protocol: "Camelot", Tvl: "180M", TvlChange_24H: "+4.5%", UtilizationRate: 0.48, HealthScore: "B"},
		},
	}

	for _, chain := range chains {
		s.defiCache[chain] = protocols[chain]
	}
}

func (s *Service) generateSampleCarry() {
	s.carryCache["fx"] = []*sentv1.CarryData{
		{Pair: "USDJPY", Spot: "150.25", Futures: "150.45", Basis: "+0.20", AnnualizedYield: "4.8%"},
		{Pair: "EURUSD", Spot: "1.0850", Futures: "1.0845", Basis: "-0.05", AnnualizedYield: "-1.2%"},
		{Pair: "GBPUSD", Spot: "1.2650", Futures: "1.2660", Basis: "+0.10", AnnualizedYield: "2.4%"},
		{Pair: "AUDUSD", Spot: "0.6520", Futures: "0.6510", Basis: "-0.10", AnnualizedYield: "-2.8%"},
	}

	s.carryCache["crypto"] = []*sentv1.CarryData{
		{Pair: "BTC", Spot: "67500", Futures: "68000", Basis: "+500", AnnualizedYield: "8.2%"},
		{Pair: "ETH", Spot: "3450", Futures: "3480", Basis: "+30", AnnualizedYield: "12.5%"},
		{Pair: "SOL", Spot: "145", Futures: "148", Basis: "+3", AnnualizedYield: "25.8%"},
	}
}

func randFloat() float64 {
	return 0.5
}

func sentimentLabel(score float64) string {
	switch {
	case score > 0.6:
		return "extreme_greed"
	case score > 0.2:
		return "greed"
	case score > -0.2:
		return "neutral"
	case score > -0.6:
		return "fear"
	default:
		return "extreme_fear"
	}
}

func convertToOnchain(sentiments []*sentv1.SentimentData) []*sentv1.OnchainMetric {
	metrics := make([]*sentv1.OnchainMetric, len(sentiments))
	for i, s := range sentiments {
		trend := "stable"
		if s.Score > 0.2 {
			trend = "rising"
		} else if s.Score < -0.2 {
			trend = "falling"
		}
		metrics[i] = &sentv1.OnchainMetric{
			Name:   s.Source,
			Value:  s.Score,
			Trend:  trend,
		}
	}
	return metrics
}

// GetSentiment returns market sentiment for an asset.
func (s *Service) GetSentiment(ctx context.Context, asset string) (*sentv1.GetSentimentResponse, error) {
	sentiment, ok := s.sentimentCache[asset]
	if !ok {
		// Generate default sentiment
		sentiment = &sentv1.SentimentData{
			Asset:     asset,
			Score:     0.1,
			Label:     "neutral",
			Source:    "composite",
			Timestamp: time.Now().Format(time.RFC3339),
		}
	}

	components := s.onchainCache[asset+"_components"]
	if components == nil {
		components = []*sentv1.OnchainMetric{
			{Name: "fear_greed", Value: sentiment.Score + 0.1, Trend: "stable"},
			{Name: "social", Value: sentiment.Score - 0.1, Trend: "stable"},
		}
	}

	return &sentv1.GetSentimentResponse{
		Sentiment:  sentiment,
		Components: convertToSentiments(components),
	}, nil
}

func convertToSentiments(metrics []*sentv1.OnchainMetric) []*sentv1.SentimentData {
	now := time.Now().Format(time.RFC3339)
	result := make([]*sentv1.SentimentData, len(metrics))
	for i, m := range metrics {
		result[i] = &sentv1.SentimentData{
			Asset:     m.Name,
			Score:     m.Value,
			Label:     sentimentLabel(m.Value),
			Source:    m.Name,
			Timestamp: now,
		}
	}
	return result
}

// GetOnchain returns onchain metrics for an asset.
func (s *Service) GetOnchain(ctx context.Context, asset string) (*sentv1.GetOnchainResponse, error) {
	metrics, ok := s.onchainCache[asset]
	if !ok {
		return &sentv1.GetOnchainResponse{
			Asset:   asset,
			Metrics: []*sentv1.OnchainMetric{},
			Signal:  "neutral",
		}, nil
	}

	signal := deriveOnchainSignal(metrics)

	return &sentv1.GetOnchainResponse{
		Asset:   asset,
		Metrics: metrics,
		Signal:  signal,
	}, nil
}

func deriveOnchainSignal(metrics []*sentv1.OnchainMetric) string {
	rising, falling := 0, 0
	for _, m := range metrics {
		switch m.Trend {
		case "rising":
			rising++
		case "falling":
			falling++
		}
	}

	if rising > falling*2 {
		return "bullish"
	} else if falling > rising*2 {
		return "bearish"
	}
	return "neutral"
}

// GetDefiHealth returns DeFi health metrics for a chain.
func (s *Service) GetDefiHealth(ctx context.Context, chain string) (*sentv1.GetDefiHealthResponse, error) {
	protocols, ok := s.defiCache[chain]
	if !ok {
		protocols = []*sentv1.DefiMetric{
			{Protocol: "Generic Protocol", Tvl: "100M", TvlChange_24H: "+0.5%", UtilizationRate: 0.60, HealthScore: "B"},
		}
	}

	return &sentv1.GetDefiHealthResponse{
		Chain:          chain,
		Protocols:      protocols,
		OverallHealth:  calculateOverallHealth(protocols),
	}, nil
}

func calculateOverallHealth(protocols []*sentv1.DefiMetric) string {
	if len(protocols) == 0 {
		return "unknown"
	}

	totalScore := 0.0
	for _, p := range protocols {
		switch p.HealthScore {
		case "A+":
			totalScore += 5
		case "A":
			totalScore += 4
		case "A-":
			totalScore += 3.5
		case "B+":
			totalScore += 3
		case "B":
			totalScore += 2
		default:
			totalScore += 2.5
		}
	}

	avg := totalScore / float64(len(protocols))
	switch {
	case avg >= 4.5:
		return "excellent"
	case avg >= 3.5:
		return "healthy"
	case avg >= 2.5:
		return "moderate"
	default:
		return "at_risk"
	}
}

// GetCarryMonitor returns carry trade opportunities.
func (s *Service) GetCarryMonitor(ctx context.Context, category string) (*sentv1.GetCarryMonitorResponse, error) {
	carries, ok := s.carryCache[category]
	if !ok {
		if category == "" {
			carries = s.carryCache["fx"]
		} else {
			return nil, fmt.Errorf("unknown category: %s", category)
		}
	}

	return &sentv1.GetCarryMonitorResponse{
		Carries: carries,
	}, nil
}
