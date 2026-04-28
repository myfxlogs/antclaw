package signals

import "time"

type Bar struct {
	Time   time.Time
	Symbol string
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
	Source string
}

type COTAnalysis struct {
	ReportDate     time.Time
	ContractCode   string
	NetPosition    int64
	COTIndex       float64
	Direction      string
	SentimentScore float64
	WoWChange      int64
	ZScore         float64
	Percentile     float64
}

type RegimeSnapshot struct {
	Time          time.Time
	Symbol        string
	Timeframe     string
	UnifiedScore  float64
	UnifiedLabel  string
	HMMState      string
	HMMConfidence float64
	GARCHRegime   string
	VolRatio      float64
	ADXStrength   string
	ADXValue      float64
}

type RegimeTransition struct {
	Time      time.Time
	Symbol    string
	Timeframe string
	FromLabel string
	ToLabel   string
	FromScore float64
	ToScore   float64
	Severity  string
}

type FactorBreakdown struct {
	Symbol    string
	AsOf      time.Time
	Momentum  float64
	LowVol    float64
	Trend     float64
	Carry     float64
	Crowding  float64
	Residual  float64
	Composite float64
}

type RankItem struct {
	Symbol    string
	Rank      int
	RawScore  float64
	NormScore float64
	Trend     string
}

type RankingResult struct {
	AsOf    time.Time
	Weights map[string]float64
	Items   []RankItem
}

type FlowDivergence struct {
	Time         time.Time
	PairA        string
	PairB        string
	Corr         float64
	BaselineMean float64
	BaselineStd  float64
	ZScore       float64
	LeadLag      int
}

type VolRegimeData struct {
	Symbol     string
	Regime     string
	Annualized float64
	Percentile float64
}

type TermPoint struct {
	TenorDays int
	IV        float64
}

type IVSurface struct {
	Underlying    string
	AtTheMoney    float64
	Skew25Delta   float64
	TermStructure []TermPoint
}

type UnifiedSignalRecord struct {
	ID             int64
	Symbol         string
	IssuedAt       time.Time
	Recommendation string
	UnifiedScore   float64
	Confidence     float64
	Components     map[string]float64
	MissingSubsys  []string
	WeightsUsed    map[string]float64
}

type SignalOutcome struct {
	SignalID       int64
	Horizon        string
	ReturnPct      float64
	DirectionMatch bool
	EvaluatedAt    time.Time
}

type OutcomeStats struct {
	SampleSize          int
	DirectionalAccuracy float64
	AvgReturn           float64
	HitRate             float64
	Sharpe              float64
	StdDev              float64
}
