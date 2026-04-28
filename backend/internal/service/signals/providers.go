package signals

import (
	"context"
	"time"
)

type PriceProvider interface {
	GetDailyBars(ctx context.Context, symbol string, from, to time.Time) ([]Bar, error)
	GetIntradayBars(ctx context.Context, symbol, interval string, from, to time.Time) ([]Bar, error)
	GetLatestPrice(ctx context.Context, symbol string) (*Bar, error)
}

type COTProvider interface {
	GetLatestAnalysis(ctx context.Context, contractCode string) (*COTAnalysis, error)
	GetAnalysisHistory(ctx context.Context, contractCode string, weeks int) ([]COTAnalysis, error)
	GetCurrencyExtremes(ctx context.Context, threshold float64) ([]COTAnalysis, error)
}

type MacroRegimeProvider interface {
	GetCurrent(ctx context.Context, symbol, timeframe string) (*RegimeSnapshot, error)
	GetHistory(ctx context.Context, symbol, timeframe string, lookback int) ([]RegimeSnapshot, error)
	GetTransitions(ctx context.Context, symbol string, since time.Time) ([]RegimeTransition, error)
}

type FactorProvider interface {
	GetRanking(ctx context.Context, category string, asOf time.Time) (*RankingResult, error)
	GetSymbolFactors(ctx context.Context, symbol string, asOf time.Time) (*FactorBreakdown, error)
}

type VolProvider interface {
	GetVIX(ctx context.Context) (float64, error)
	GetGARCH(ctx context.Context, symbol string) (*VolRegimeData, error)
	GetIVSurface(ctx context.Context, underlying string) (*IVSurface, error)
}

type FlowProvider interface {
	GetDivergence(ctx context.Context, pairA, pairB string, since time.Time) ([]FlowDivergence, error)
	GetTopDivergent(ctx context.Context, since time.Time, limit int) ([]FlowDivergence, error)
}

type SignalRepo interface {
	SaveUnified(ctx context.Context, sig UnifiedSignalRecord) (int64, error)
	GetByID(ctx context.Context, id int64) (*UnifiedSignalRecord, error)
	GetRecentBySymbol(ctx context.Context, symbol string, limit int) ([]UnifiedSignalRecord, error)
	SaveOutcome(ctx context.Context, outcome SignalOutcome) error
	GetOutcomeStats(ctx context.Context, signalType, symbol, horizon string, since time.Time) (*OutcomeStats, error)
	GetActiveWeights(ctx context.Context) (map[string]float64, error)
}
