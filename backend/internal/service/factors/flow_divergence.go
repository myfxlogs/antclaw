package factors

import (
	"math"
	"time"

	"github.com/antclaw/antclaw/internal/infra/postgres"
)

// FlowDivergenceDetector detects when correlated asset pairs decouple
type FlowDivergenceDetector struct {
	priceRepo postgres.PriceRepository
}

// NewFlowDivergenceDetector creates a new detector
func NewFlowDivergenceDetector(priceRepo postgres.PriceRepository) *FlowDivergenceDetector {
	return &FlowDivergenceDetector{priceRepo: priceRepo}
}

// DivergenceSignal represents a detected divergence
type DivergenceSignal struct {
	Asset1        string
	Asset2        string
	CurrentCorr   float64
	BaselineMean  float64
	BaselineStd   float64
	ZScore        float64
	IsDiverged    bool
	Direction1    string
	Direction2    string
	DetectedAt    time.Time
}

// CorrelationPair represents an asset pair to monitor
type CorrelationPair struct {
	Asset1 string
	Asset2 string
	Window int // Rolling window for correlation (default 20)
}

// DefaultPairs defines classic intermarket pairs to monitor
var DefaultPairs = []CorrelationPair{
	{Asset1: "DXY", Asset2: "EURUSD", Window: 20},
	{Asset1: "DXY", Asset2: "XAUUSD", Window: 20},
	{Asset1: "SPX", Asset2: "VIX", Window: 20},
	{Asset1: "USOIL", Asset2: "USDCAD", Window: 20},
	{Asset1: "AUDUSD", Asset2: "XAUUSD", Window: 20},
	{Asset1: "AUDUSD", Asset2: "XCUUSD", Window: 20},
}

// Detect analyzes pairs for divergence
func (d *FlowDivergenceDetector) Detect(pairs []CorrelationPair, historyDays int) ([]DivergenceSignal, error) {
	var signals []DivergenceSignal
	baselineWindow := 60 // 60-day baseline

	for _, pair := range pairs {
		// Fetch prices for both assets
		prices1, err := d.fetchPrices(pair.Asset1, historyDays+baselineWindow)
		if err != nil {
			continue
		}
		prices2, err := d.fetchPrices(pair.Asset2, historyDays+baselineWindow)
		if err != nil {
			continue
		}

		if len(prices1) < baselineWindow+pair.Window || len(prices2) < baselineWindow+pair.Window {
			continue
		}

		// Calculate baseline (60-day) correlation statistics
		baselineCorrs := d.calculateRollingCorrelations(prices1, prices2, pair.Window, baselineWindow)
		baselineMean := mean(baselineCorrs)
		baselineStd := stdDev(baselineCorrs, baselineMean)

		// Calculate current correlation
		currentCorr := correlation(
			prices1[len(prices1)-pair.Window:],
			prices2[len(prices2)-pair.Window:],
		)

		// Calculate z-score
		zScore := (currentCorr - baselineMean) / baselineStd
		if baselineStd == 0 {
			zScore = 0
		}

		// Determine divergence (> 2 sigma deviation)
		isDiverged := math.Abs(zScore) > 2.0

		signal := DivergenceSignal{
			Asset1:       pair.Asset1,
			Asset2:       pair.Asset2,
			CurrentCorr:  currentCorr,
			BaselineMean: baselineMean,
			BaselineStd:  baselineStd,
			ZScore:       zScore,
			IsDiverged:   isDiverged,
			DetectedAt:   time.Now(),
		}

		// Determine direction
		if len(prices1) > 0 && len(prices2) > 0 {
			ret1 := (prices1[len(prices1)-1] - prices1[len(prices1)-2]) / prices1[len(prices1)-2]
			ret2 := (prices2[len(prices2)-1] - prices2[len(prices2)-2]) / prices2[len(prices2)-2]

			if ret1 > 0 {
				signal.Direction1 = "UP"
			} else {
				signal.Direction1 = "DOWN"
			}

			if ret2 > 0 {
				signal.Direction2 = "UP"
			} else {
				signal.Direction2 = "DOWN"
			}
		}

		signals = append(signals, signal)
	}

	return signals, nil
}

// fetchPrices fetches historical prices for an asset
func (d *FlowDivergenceDetector) fetchPrices(symbol string, days int) ([]float64, error) {
	// In production, fetch from price repository
	// For now, return empty (would need actual implementation)
	return nil, nil
}

// calculateRollingCorrelations calculates rolling correlations
func (d *FlowDivergenceDetector) calculateRollingCorrelations(prices1, prices2 []float64, window, periods int) []float64 {
	var correlations []float64

	for i := 0; i < periods; i++ {
		end := len(prices1) - i
		start := end - window

		if start < 0 {
			break
		}

		corr := correlation(prices1[start:end], prices2[start:end])
		correlations = append(correlations, corr)
	}

	return correlations
}

// LeadLagAnalysis performs cross-correlation at different lags to find leading/lagging relationship
func (d *FlowDivergenceDetector) LeadLagAnalysis(prices1, prices2 []float64, maxLag int) (leadLag map[int]float64, leader string) {
	leadLag = make(map[int]float64)
	maxCorr := -1.0
	bestLag := 0

	for lag := -maxLag; lag <= maxLag; lag++ {
		var corr float64

		if lag >= 0 {
			// Asset1 leads
			if len(prices1) > lag+20 && len(prices2) > 20 {
				corr = correlation(prices1[lag:lag+20], prices2[:20])
			}
		} else {
			// Asset2 leads
			lag = -lag
			if len(prices2) > lag+20 && len(prices1) > 20 {
				corr = correlation(prices2[lag:lag+20], prices1[:20])
			}
			corr = -corr // Negative indicates asset2 leads
		}

		leadLag[lag] = corr

		if math.Abs(corr) > maxCorr {
			maxCorr = math.Abs(corr)
			bestLag = lag
		}
	}

	if bestLag > 0 {
		return leadLag, "asset1"
	} else if bestLag < 0 {
		return leadLag, "asset2"
	}
	return leadLag, "neutral"
}

// correlation calculates Pearson correlation
func correlation(x, y []float64) float64 {
	if len(x) != len(y) || len(x) < 2 {
		return 0
	}

	meanX := mean(x)
	meanY := mean(y)

	var sumXY, sumX2, sumY2 float64
	for i := 0; i < len(x); i++ {
		dx := x[i] - meanX
		dy := y[i] - meanY
		sumXY += dx * dy
		sumX2 += dx * dx
		sumY2 += dy * dy
	}

	if sumX2 == 0 || sumY2 == 0 {
		return 0
	}

	return sumXY / math.Sqrt(sumX2*sumY2)
}

// mean calculates arithmetic mean
func mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// stdDev calculates standard deviation
func stdDev(values []float64, mean float64) float64 {
	if len(values) < 2 {
		return 0
	}

	var sumSquared float64
	for _, v := range values {
		diff := v - mean
		sumSquared += diff * diff
	}

	return math.Sqrt(sumSquared / float64(len(values)))
}
