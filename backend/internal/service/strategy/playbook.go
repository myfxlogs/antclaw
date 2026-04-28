package strategy

import (
	"context"
	"fmt"
	"time"

	"github.com/antclaw/antclaw/internal/domain/shared"
)

// PlaybookEngine generates executable trading plans
type PlaybookEngine struct {
	factorEngine    FactorProvider
	macroProvider   MacroProvider
	cotProvider     COTProvider
	volProvider     VolProvider
	carryProvider   CarryProvider
}

// FactorProvider provides factor rankings
type FactorProvider interface {
	GetRanking(ctx context.Context) (*RankingResult, error)
}

// MacroProvider provides macro regime info
type MacroProvider interface {
	GetCurrentRegime(ctx context.Context) (*MacroRegime, error)
}

// COTProvider provides COT bias data
type COTProvider interface {
	GetBias(ctx context.Context, symbol string) (*COTBias, error)
}

// VolProvider provides volatility regime
type VolProvider interface {
	GetVolRegime(ctx context.Context, symbol string) (VolRegime, error)
}

// CarryProvider provides carry rates
type CarryProvider interface {
	GetCarryRate(ctx context.Context, symbol string) (float64, error)
}

// NewPlaybookEngine creates a new playbook engine
func NewPlaybookEngine(
	factors FactorProvider,
	macro MacroProvider,
	cot COTProvider,
	vol VolProvider,
	carry CarryProvider,
) *PlaybookEngine {
	return &PlaybookEngine{
		factorEngine:  factors,
		macroProvider: macro,
		cotProvider:   cot,
		volProvider:   vol,
		carryProvider: carry,
	}
}

// Playbook represents a trading plan
type Playbook struct {
	GeneratedAt      time.Time
	Regime           string
	IsTransition     bool
	TransitionProb   float64
	Entries          []PlaybookEntry
	GlobalRisk       GlobalRiskAssessment
}

// PlaybookEntry represents a single trade idea
type PlaybookEntry struct {
	Symbol          string
	Direction       shared.Direction
	Conviction      ConvictionLevel
	Score           float64
	Evidence        []Evidence
	SuggestedSize   float64 // % of account
	MaxRisk         float64 // % per trade
	HoldPeriod      time.Duration
}

// Evidence supports the trade decision
type Evidence struct {
	Source    string
	Signal    string
	Weight    float64
	Alignment bool // true if supports direction
}

// GlobalRiskAssessment represents portfolio-level risk
type GlobalRiskAssessment struct {
	TotalHeat       float64   // sum of all position risks
	MaxHeatAllowed  float64   // typically 6%
	VolRegime       string
	Recommendation  string
}

// ConvictionLevel represents confidence level
type ConvictionLevel string

const (
	ConvHigh   ConvictionLevel = "HIGH"
	ConvMedium ConvictionLevel = "MEDIUM"
	ConvLow    ConvictionLevel = "LOW"
	ConvAvoid  ConvictionLevel = "AVOID"
)

// VolRegime represents volatility regime
type VolRegime string

const (
	VolExpanding   VolRegime = "EXPANDING"
	VolContracting VolRegime = "CONTRACTING"
	VolNormal      VolRegime = "NORMAL"
)

// RankingResult from factor engine
type RankingResult struct {
	Ranked []AssetRank
	Top    []AssetRank
	Bottom []AssetRank
}

// AssetRank represents a single asset ranking
type AssetRank struct {
	Symbol       string
	Currency     string
	RawScore     float64
	NormScore    float64
	Rank         int
}

// MacroRegime represents macro state
type MacroRegime struct {
	Regime           string
	Score            float64
	IsTransition     bool
	TransitionProb   float64
	TransitionFrom   string
	TransitionTo     string
}

// COTBias represents COT positioning bias
type COTBias struct {
	Symbol      string
	Direction   shared.Direction
	Confidence  float64
	Index       float64 // 0-100
}

// Generate creates a playbook from all inputs
func (e *PlaybookEngine) Generate(ctx context.Context, opts PlaybookOpts) (*Playbook, error) {
	// Gather inputs
	ranking, err := e.factorEngine.GetRanking(ctx)
	if err != nil {
		return nil, fmt.Errorf("factor ranking failed: %w", err)
	}

	macro, err := e.macroProvider.GetCurrentRegime(ctx)
	if err != nil {
		return nil, fmt.Errorf("macro regime failed: %w", err)
	}

	// Build playbook
	pb := &Playbook{
		GeneratedAt:    time.Now(),
		Regime:         macro.Regime,
		IsTransition:   macro.IsTransition,
		TransitionProb: macro.TransitionProb,
	}

	// Process top candidates (long)
	for _, asset := range ranking.Top {
		entry, err := e.evaluateAsset(ctx, asset, shared.DirectionLong, macro)
		if err != nil {
			continue
		}
		if entry.Conviction != ConvAvoid {
			pb.Entries = append(pb.Entries, *entry)
		}
	}

	// Process bottom candidates (short)
	for _, asset := range ranking.Bottom {
		entry, err := e.evaluateAsset(ctx, asset, shared.DirectionShort, macro)
		if err != nil {
			continue
		}
		if entry.Conviction != ConvAvoid {
			pb.Entries = append(pb.Entries, *entry)
		}
	}

	// Calculate global risk
	pb.GlobalRisk = e.assessGlobalRisk(pb)

	return pb, nil
}

// evaluateAsset evaluates a single asset for playbook inclusion
func (e *PlaybookEngine) evaluateAsset(ctx context.Context, asset AssetRank, direction shared.Direction, macro *MacroRegime) (*PlaybookEntry, error) {
	entry := &PlaybookEntry{
		Symbol:    asset.Symbol,
		Direction: direction,
		Score:     asset.NormScore,
	}

	var evidence []Evidence
	var score float64 = asset.NormScore

	// Check COT alignment
	cot, err := e.cotProvider.GetBias(ctx, asset.Symbol)
	if err == nil && cot != nil {
		aligned := cot.Direction == direction
		weight := 0.25
		if !aligned {
			weight = -0.15
			score *= 0.8 // Reduce score if COT opposes
		} else {
			score *= 1.2 // Boost if COT aligns
		}
		
		evidence = append(evidence, Evidence{
			Source:    "COT",
			Signal:    fmt.Sprintf("%s (index: %.1f)", cot.Direction, cot.Index),
			Weight:    weight,
			Alignment: aligned,
		})
	}

	// Check regime fit
	regimeFit := e.checkRegimeFit(asset.Symbol, direction, macro.Regime)
	if regimeFit {
		score *= 1.1
		evidence = append(evidence, Evidence{
			Source:    "Macro",
			Signal:    fmt.Sprintf("Regime %s supports %s", macro.Regime, direction),
			Weight:    0.15,
			Alignment: true,
		})
	}

	// Check volatility regime
	volRegime, _ := e.volProvider.GetVolRegime(ctx, asset.Symbol)
	if volRegime == VolExpanding {
		score *= 0.7 // Reduce in expanding vol
		evidence = append(evidence, Evidence{
			Source:    "Volatility",
			Signal:    "Expanding volatility",
			Weight:    -0.10,
			Alignment: false,
		})
	}

	// Check carry
	carry, _ := e.carryProvider.GetCarryRate(ctx, asset.Symbol)
	if carry > 0 && direction == shared.DirectionLong {
		score *= 1.05
		evidence = append(evidence, Evidence{
			Source:    "Carry",
			Signal:    fmt.Sprintf("Positive carry: %.2f%%", carry*100),
			Weight:    0.05,
			Alignment: true,
		})
	}

	// Handle transition period
	if macro.IsTransition && macro.TransitionProb > 0.5 {
		score *= 0.6 // Significant reduction in transitions
	}

	entry.Score = score
	entry.Evidence = evidence
	entry.Conviction = e.determineConviction(score, asset.Rank, evidence)
	entry.MaxRisk = e.calculateRisk(entry.Conviction, volRegime)

	return entry, nil
}

// checkRegimeFit checks if direction fits macro regime
func (e *PlaybookEngine) checkRegimeFit(symbol string, direction shared.Direction, regime string) bool {
	// Simplified logic - in production use proper mapping
	switch regime {
	case "EXPANSION", "GOLDILOCKS":
		return direction == shared.DirectionLong // Risk-on regimes favor longs
	case "RECESSION", "DEFLATION":
		return direction == shared.DirectionShort // Risk-off favors shorts
	default:
		return true
	}
}

// determineConviction determines conviction level
func (e *PlaybookEngine) determineConviction(score float64, rank int, evidence []Evidence) ConvictionLevel {
	// High: top 3 rank + aligned COT + good regime fit
	if rank <= 3 && score > 1.5 {
		for _, e := range evidence {
			if e.Source == "COT" && e.Alignment {
				return ConvHigh
			}
		}
	}

	// Medium: top 5 rank + partial support
	if rank <= 5 && score > 1.0 {
		return ConvMedium
	}

	// Low: lower rank or weak evidence
	if score > 0.5 {
		return ConvLow
	}

	return ConvAvoid
}

// calculateRisk calculates max risk per trade
func (e *PlaybookEngine) calculateRisk(conviction ConvictionLevel, volRegime VolRegime) float64 {
	baseRisk := map[ConvictionLevel]float64{
		ConvHigh:   0.02,  // 2%
		ConvMedium: 0.015, // 1.5%
		ConvLow:    0.01,  // 1%
		ConvAvoid:  0,
	}

	risk := baseRisk[conviction]

	// Volatility adjustment
	if volRegime == VolExpanding {
		risk *= 0.7
	}

	return risk
}

// assessGlobalRisk assesses portfolio-level risk
func (e *PlaybookEngine) assessGlobalRisk(pb *Playbook) GlobalRiskAssessment {
	var totalHeat float64

	for _, entry := range pb.Entries {
		totalHeat += entry.MaxRisk
	}

	assessment := GlobalRiskAssessment{
		TotalHeat:      totalHeat,
		MaxHeatAllowed: 0.06, // 6% standard
		VolRegime:      "NORMAL",
	}

	if totalHeat > assessment.MaxHeatAllowed {
		assessment.Recommendation = "REDUCE_POSITIONS"
	} else if totalHeat > assessment.MaxHeatAllowed*0.8 {
		assessment.Recommendation = "CAUTION"
	} else {
		assessment.Recommendation = "NORMAL"
	}

	return assessment
}

// PlaybookOpts for playbook generation
type PlaybookOpts struct {
	MaxLongs    int
	MaxShorts   int
	MinScore    float64
}
