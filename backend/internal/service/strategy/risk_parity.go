package strategy

import (
	"context"
	"fmt"
	"math"
)

// RiskParitySizer manages portfolio-level risk allocation
type RiskParitySizer struct{}

// NewRiskParitySizer creates a new risk parity sizer
func NewRiskParitySizer() *RiskParitySizer {
	return &RiskParitySizer{}
}

// AccountInfo represents account information
type AccountInfo struct {
	Balance      float64
	MaxHeat      float64 // Max total risk (e.g., 0.06 = 6%)
	WinRate      float64 // Historical win rate
	AvgWinLoss   float64 // Avg win / avg loss ratio
}

// PlannedPosition represents a position to be allocated
type PlannedPosition struct {
	Symbol      string
	Direction   string
	EntryPrice  float64
	StopPrice   float64
	Conviction  ConvictionLevel
}

// Allocation represents position sizing result
type Allocation struct {
	Positions   []PositionAllocation
	TotalRisk   float64
	KellyFraction float64
}

// PositionAllocation represents sizing for one position
type PositionAllocation struct {
	Symbol       string
	Shares       float64
	RiskAmount   float64
	RiskPercent  float64
	PositionSize float64
}

// Allocate calculates position sizes based on risk parity
func (s *RiskParitySizer) Allocate(ctx context.Context, positions []PlannedPosition, account AccountInfo) (*Allocation, error) {
	if len(positions) == 0 {
		return &Allocation{}, nil
	}

	// Calculate Kelly fraction
	kelly := s.calculateKelly(account.WinRate, account.AvgWinLoss)

	// Calculate initial risk per position
	equalRisk := account.MaxHeat / float64(len(positions))

	allocation := &Allocation{
		Positions:     make([]PositionAllocation, 0, len(positions)),
		KellyFraction: kelly,
	}

	for _, pos := range positions {
		// Calculate risk distance
		riskDistance := math.Abs(pos.EntryPrice - pos.StopPrice) / pos.EntryPrice
		if riskDistance == 0 {
			continue // Skip invalid stops
		}

		// Adjust risk based on conviction
		adjustedRisk := equalRisk * s.convictionMultiplier(pos.Conviction)

		// Apply Kelly fraction
		finalRisk := adjustedRisk * kelly

		// Calculate shares
		riskAmount := account.Balance * finalRisk
		shares := riskAmount / (pos.EntryPrice * riskDistance)

		allocation.Positions = append(allocation.Positions, PositionAllocation{
			Symbol:       pos.Symbol,
			Shares:       shares,
			RiskAmount:   riskAmount,
			RiskPercent:  finalRisk,
			PositionSize: shares * pos.EntryPrice,
		})

		allocation.TotalRisk += finalRisk
	}

	// Scale down if total risk exceeds max
	if allocation.TotalRisk > account.MaxHeat {
		scale := account.MaxHeat / allocation.TotalRisk
		allocation = s.scaleAllocation(allocation, scale)
	}

	return allocation, nil
}

// calculateKelly calculates Kelly criterion fraction
func (s *RiskParitySizer) calculateKelly(winRate, avgWinLoss float64) float64 {
	// Kelly = (p*b - (1-p)) / b
	// where p = win rate, b = avg win/loss
	if avgWinLoss == 0 {
		return 0.5 // Default to half Kelly
	}

	kelly := (winRate*avgWinLoss - (1-winRate)) / avgWinLoss

	// Cap between 0.1 and 0.5 (half Kelly to full Kelly)
	if kelly > 0.5 {
		kelly = 0.5
	} else if kelly < 0.1 {
		kelly = 0.1
	}

	return kelly
}

// convictionMultiplier returns risk multiplier based on conviction
func (s *RiskParitySizer) convictionMultiplier(conviction ConvictionLevel) float64 {
	multipliers := map[ConvictionLevel]float64{
		ConvHigh:   1.3,
		ConvMedium: 1.0,
		ConvLow:    0.7,
		ConvAvoid:  0,
	}

	m, ok := multipliers[conviction]
	if !ok {
		return 0.5
	}
	return m
}

// scaleAllocation scales all positions by factor
func (s *RiskParitySizer) scaleAllocation(alloc *Allocation, scale float64) *Allocation {
	scaled := &Allocation{
		Positions:     make([]PositionAllocation, len(alloc.Positions)),
		KellyFraction: alloc.KellyFraction,
	}

	for i, pos := range alloc.Positions {
		scaled.Positions[i] = PositionAllocation{
			Symbol:       pos.Symbol,
			Shares:       pos.Shares * scale,
			RiskAmount:   pos.RiskAmount * scale,
			RiskPercent:  pos.RiskPercent * scale,
			PositionSize: pos.PositionSize * scale,
		}
		scaled.TotalRisk += scaled.Positions[i].RiskPercent
	}

	return scaled
}

// AdjustForVolatility adjusts allocation based on volatility regime
func (s *RiskParitySizer) AdjustForVolatility(alloc *Allocation, regime VolRegime) *Allocation {
	multiplier := 1.0

	switch regime {
	case VolExpanding:
		multiplier = 0.7
	case VolContracting:
		multiplier = 1.2
	}

	return s.scaleAllocation(alloc, multiplier)
}

// Validate checks if allocation meets risk constraints
func (s *RiskParitySizer) Validate(alloc *Allocation, account AccountInfo) error {
	if alloc.TotalRisk > account.MaxHeat {
		return fmt.Errorf("total risk %.2f%% exceeds max heat %.2f%%",
			alloc.TotalRisk*100, account.MaxHeat*100)
	}

	for _, pos := range alloc.Positions {
		if pos.RiskPercent > 0.03 { // No single position > 3%
			return fmt.Errorf("position %s risk %.2f%% exceeds 3%% limit",
				pos.Symbol, pos.RiskPercent*100)
		}
	}

	return nil
}
