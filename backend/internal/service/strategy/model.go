// Package strategy provides strategy management service.
package strategy

import (
	"time"

	"github.com/google/uuid"
)

// Strategy represents a trading strategy configuration.
type Strategy struct {
	ID            uuid.UUID      `json:"id"`
	Name          string         `json:"name"`
	Kind          string         `json:"kind"`
	Symbol        string         `json:"symbol"`
	Timeframe     string         `json:"timeframe"`
	Params        map[string]any `json:"params"`
	ScheduleCron  string         `json:"schedule_cron"`
	Enabled       bool           `json:"enabled"`
	Status        string         `json:"status"`
	Description   string         `json:"description"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	CreatedBy     string         `json:"created_by"`
	UpdatedBy     string         `json:"updated_by"`
	LastRunAt     *time.Time     `json:"last_run_at,omitempty"`
	LastRunStatus string         `json:"last_run_status,omitempty"`
	DeletedAt     *time.Time     `json:"deleted_at,omitempty"`
}

// RunResult represents a single backtest run result.
type RunResult struct {
	RunID        uuid.UUID      `json:"run_id"`
	StrategyID   uuid.UUID      `json:"strategy_id"`
	StartedAt    time.Time      `json:"started_at"`
	FinishedAt   time.Time      `json:"finished_at"`
	Status       string         `json:"status"`
	Metrics      map[string]any `json:"metrics"`
	Mock         bool           `json:"mock"`
	ErrorMessage string         `json:"error_message,omitempty"`
}

// RunMetrics holds common backtest metrics.
type RunMetrics struct {
	TotalReturn float64 `json:"total_return"`
	Sharpe      float64 `json:"sharpe"`
	MaxDrawdown float64 `json:"max_drawdown"`
	WinRate     float64 `json:"win_rate"`
	Trades      int     `json:"trades"`
}

// validKinds defines allowed strategy types.
var validKinds = map[string]bool{
	"ma_cross":               true,
	"rsi_reversal":           true,
	"cot_extreme":            true,
	"breakout":               true,
	"quant_momentum":         true,
	"quant_trend":            true,
	"quant_mean_reversion":   true,
	"quant_carry":            true,
	"quant_multi_factor":     true,
	"vp_rotation":            true,
	"vp_breakout":            true,
	"vp_absorption_reversal": true,
	"cta_dual_ema":           true,
	"cta_donchian_breakout":  true,
	"cta_atr_trail":          true,
	"cta_multi_tf":           true,
}

// validStatuses defines allowed statuses.
var validStatuses = map[string]bool{
	"draft":    true,
	"active":   true,
	"archived": true,
}

// IsValidKind checks if the strategy kind is valid.
func IsValidKind(k string) bool {
	return validKinds[k]
}

// IsValidStatus checks if the status is valid.
func IsValidStatus(s string) bool {
	return validStatuses[s]
}

// IsValidCron validates cron expression (simplified: @hourly, @daily, or basic cron).
func IsValidCron(c string) bool {
	if c == "" {
		return false
	}
	// Preset macros
	if c == "@hourly" || c == "@daily" {
		return true
	}
	// Must have at least 5 fields
	parts := 0
	for _, r := range c {
		if r == ' ' {
			parts++
		}
	}
	return parts >= 4 // 5 fields = 4 spaces
}
