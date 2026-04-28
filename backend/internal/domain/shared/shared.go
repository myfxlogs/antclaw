package shared

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// Direction represents signal direction
type Direction string

const (
	DirectionLong   Direction = "LONG"
	DirectionShort  Direction = "SHORT"
	DirectionFlat   Direction = "FLAT"
	DirectionBullish Direction = "BULLISH"
	DirectionBearish Direction = "BEARISH"
	DirectionNeutral Direction = "NEUTRAL"
)

// ImpactLevel represents event impact level
type ImpactLevel string

const (
	ImpactLow      ImpactLevel = "low"
	ImpactMedium   ImpactLevel = "medium"
	ImpactHigh     ImpactLevel = "high"
	ImpactHoliday  ImpactLevel = "holiday"
)

// Regime represents macro regime
type Regime string

const (
	RegimeInflationary Regime = "INFLATIONARY"
	RegimeGoldilocks   Regime = "GOLDILOCKS"
	RegimeStagflation  Regime = "STAGFLATION"
	RegimeDeflation    Regime = "DEFLATION"
	RegimeStress       Regime = "STRESS"
	RegimeNeutral      Regime = "NEUTRAL"
)

// InstrumentKind represents financial instrument type
type InstrumentKind string

const (
	KindFX       InstrumentKind = "FX"
	KindCrypto   InstrumentKind = "CRYPTO"
	KindEquity   InstrumentKind = "EQUITY"
	KindCommodity InstrumentKind = "COMMODITY"
	KindIndex    InstrumentKind = "INDEX"
	KindMacro    InstrumentKind = "MACRO"
)

// Instrument represents a tradeable instrument
type Instrument struct {
	Symbol string         `json:"symbol"`
	Venue  string         `json:"venue"`
	Kind   InstrumentKind `json:"kind"`
}

func (i Instrument) String() string {
	return fmt.Sprintf("%s:%s", i.Venue, i.Symbol)
}

// Interval represents time interval
type Interval string

const (
	IntervalM1  Interval = "M1"
	IntervalM5  Interval = "M5"
	IntervalM15 Interval = "M15"
	IntervalH1  Interval = "H1"
	IntervalH4  Interval = "H4"
	IntervalD1  Interval = "D1"
	IntervalW1  Interval = "W1"
)

// TimeRange represents a time range with validation
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

func (tr TimeRange) Validate() error {
	if tr.Start.After(tr.End) {
		return fmt.Errorf("start time %v after end time %v", tr.Start, tr.End)
	}
	return nil
}

// Percent represents a percentage value (stored as decimal, 0.1234 = 12.34%)
type Percent float64

func (p Percent) Validate() error {
	v := float64(p)
	if math.IsNaN(v) || math.IsInf(v, 0) || v < -10000 || v > 10000 {
		return fmt.Errorf("invalid percent: %v", v)
	}
	return nil
}

func (p Percent) String() string {
	return fmt.Sprintf("%.2f%%", float64(p)*100)
}

// Locale represents BCP-47 locale
type Locale string

// ValidLocales defines supported locales
var ValidLocales = map[Locale]bool{
	"zh-CN": true,
	"en-US": true,
}

func (l Locale) Validate() error {
	if !ValidLocales[l] {
		return fmt.Errorf("unsupported locale: %s", l)
	}
	return nil
}

// Timezone represents IANA timezone
type Timezone string

func (t Timezone) Validate() error {
	if _, err := time.LoadLocation(string(t)); err != nil {
		return fmt.Errorf("invalid timezone %s: %w", t, err)
	}
	return nil
}

// SignalType represents a signal classification
type SignalType string

const (
	SignalTypeCOT           SignalType = "COT"
	SignalTypeCTA           SignalType = "CTA"
	SignalTypeQuant         SignalType = "QUANT"
	SignalTypeSentiment     SignalType = "SENTIMENT"
	SignalTypeSeasonal      SignalType = "SEASONAL"
	SignalTypeUnified       SignalType = "UNIFIED"
	SignalTypeTA            SignalType = "TA"
)

// Grade represents confluence grade
type Grade string

const (
	GradeA Grade = "A"  // High confluence
	GradeB Grade = "B"  // Partial confluence
	GradeC Grade = "C"  // Noise
)

// Recommendation represents a trading recommendation
type Recommendation string

const (
	RecStrongLong  Recommendation = "STRONG_LONG"
	RecLong        Recommendation = "LONG"
	RecNeutral     Recommendation = "NEUTRAL"
	RecShort       Recommendation = "SHORT"
	RecStrongShort Recommendation = "STRONG_SHORT"
)

// ConvictionLevel represents signal conviction
type ConvictionLevel string

const (
	ConvHigh    ConvictionLevel = "HIGH"
	ConvMedium  ConvictionLevel = "MEDIUM"
	ConvLow     ConvictionLevel = "LOW"
	ConvAvoid   ConvictionLevel = "AVOID"
)

// UserRole represents user role
type UserRole string

const (
	RoleFree     UserRole = "free"
	RolePremium  UserRole = "premium"
	RoleAdmin    UserRole = "admin"
)

// Helper functions
func NormalizeSymbol(sym string) string {
	return strings.ToUpper(strings.TrimSpace(sym))
}

func IsValidImpact(s string) bool {
	switch ImpactLevel(s) {
	case ImpactLow, ImpactMedium, ImpactHigh, ImpactHoliday:
		return true
	default:
		return false
	}
}

func IsValidDirection(s string) bool {
	switch Direction(s) {
	case DirectionLong, DirectionShort, DirectionFlat,
		 DirectionBullish, DirectionBearish, DirectionNeutral:
		return true
	default:
		return false
	}
}

// VolRegime represents volatility regime
type VolRegime string

const (
	VolExpanding   VolRegime = "EXPANDING"
	VolContracting VolRegime = "CONTRACTING"
	VolNormal      VolRegime = "NORMAL"
)

// CrossVolRegime represents cross-asset volatility regime
type CrossVolRegime string

const (
	CrossVolNormal       CrossVolRegime = "NORMAL"
	CrossVolEnergyRisk   CrossVolRegime = "ENERGY_RISK"
	CrossVolBroadRiskOff CrossVolRegime = "BROAD_RISK_OFF"
	CrossVolSmallCap     CrossVolRegime = "SMALL_CAP_STRESS"
	CrossVolSystemic     CrossVolRegime = "SYSTEMIC"
)
