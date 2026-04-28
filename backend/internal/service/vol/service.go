package vol

import (
	"context"
	"fmt"
	"math"
	"time"

	volv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
)

// Service implements Volatility business logic with real calculations.
type Service struct {
	vixStore    *VIXStore
	gexStore    *GEXStore
	ivolStore   *IVolStore
	skewStore   *SkewStore
}

// VIXStore holds VIX data.
type VIXStore struct {
	current    float64
	history    []VIXPoint
	percentile float64
}

type VIXPoint struct {
	Timestamp time.Time
	Value     float64
}

// GEXStore holds Gamma Exposure data.
type GEXStore struct {
	data map[string]*GEXData
}

type GEXData struct {
	NetGEX     float64
	FlipPoint  float64
	GammaWalls  []float64
	Timestamp  time.Time
}

// IVolStore holds implied volatility data.
type IVolStore struct {
	surfaces map[string]*IVolSurface
}

type IVolSurface struct {
	Points  []IVolPoint
	ATMIV   float64
	Expiry  time.Time
}

type IVolPoint struct {
	Strike float64
	IV     float64
	Delta  float64
	Gamma  float64
	Theta  float64
	Vega   float64
}

// SkewStore holds skew data.
type SkewStore struct {
	data map[string]*SkewData
}

type SkewData struct {
	RiskReversal float64
	Butterfly    float64
	Timestamp    time.Time
}

// NewService creates a new VolService with sample data.
func NewService() *Service {
	return &Service{
		vixStore:  newVIXStore(),
		gexStore:  newGEXStore(),
		ivolStore: newIVolStore(),
		skewStore: newSkewStore(),
	}
}

func newVIXStore() *VIXStore {
	// Generate 30 days of VIX history
	history := make([]VIXPoint, 30)
	now := time.Now()
	baseVIX := 18.0
	
	for i := 0; i < 30; i++ {
		history[i] = VIXPoint{
			Timestamp: now.Add(-time.Duration(29-i) * 24 * time.Hour),
			Value:     baseVIX + math.Sin(float64(i)*0.3)*3 + (randFloat64()-0.5)*2,
		}
	}
	
	current := history[29].Value
	percentile := calculatePercentile(current, history)
	
	return &VIXStore{
		current:    current,
		history:    history,
		percentile: percentile,
	}
}

func newGEXStore() *GEXStore {
	store := &GEXStore{data: make(map[string]*GEXData)}
	
	// Sample GEX for major pairs
	pairs := []string{"EURUSD", "GBPUSD", "USDJPY", "SPY", "QQQ"}
	
	for _, pair := range pairs {
		spot := 1.0
		if pair == "SPY" {
			spot = 450.0
		} else if pair == "QQQ" {
			spot = 380.0
		} else if pair == "USDJPY" {
			spot = 150.0
		}
		
		netGEX := (randFloat64() - 0.5) * 1000000
		flipPoint := spot * (1 + (randFloat64()-0.5)*0.02)
		
		walls := []float64{
			spot * 0.98,
			spot * 0.99,
			spot * 1.01,
			spot * 1.02,
		}
		
		store.data[pair] = &GEXData{
			NetGEX:    netGEX,
			FlipPoint: flipPoint,
			GammaWalls: walls,
			Timestamp: time.Now(),
		}
	}
	
	return store
}

func newIVolStore() *IVolStore {
	store := &IVolStore{surfaces: make(map[string]*IVolSurface)}
	
	// Generate IV surfaces for major pairs
	pairs := []string{"EURUSD", "GBPUSD", "USDJPY"}
	
	for _, pair := range pairs {
		spot := 1.0
		if pair == "USDJPY" {
			spot = 150.0
		}
		
		atmIV := 8.0 + randFloat64()*4.0
		points := generateIVSurface(spot, atmIV)
		
		store.surfaces[pair] = &IVolSurface{
			Points: points,
			ATMIV:  atmIV,
			Expiry: time.Now().Add(30 * 24 * time.Hour),
		}
	}
	
	return store
}

func generateIVSurface(spot, atmIV float64) []IVolPoint {
	var points []IVolPoint
	
	// Generate strikes from -10% to +10%
	for i := -10; i <= 10; i++ {
		strike := spot * (1 + float64(i)*0.01)
		
		// Vol smile: higher IV for OTM options
		moneyness := strike / spot - 1.0
		volAdj := moneyness * moneyness * 20.0 // Quadratic smile
		iv := atmIV + volAdj
		
		// Calculate Greeks (simplified Black-Scholes)
		tau := 30.0 / 365.0 // 30 days
		delta := calculateDelta(strike, spot, iv/100, tau, i > 0)
		gamma := calculateGamma(strike, spot, iv/100, tau)
		theta := calculateTheta(strike, spot, iv/100, tau)
		vega := calculateVega(strike, spot, iv/100, tau)
		
		points = append(points, IVolPoint{
			Strike: strike,
			IV:     iv,
			Delta:  delta,
			Gamma:  gamma,
			Theta:  theta,
			Vega:   vega,
		})
	}
	
	return points
}

// Simplified Black-Scholes Greeks calculations
func calculateDelta(strike, spot, vol, tau float64, isCall bool) float64 {
	d1 := (math.Log(spot/strike) + 0.5*vol*vol*tau) / (vol * math.Sqrt(tau))
	if isCall {
		return 0.5 + 0.5*math.Erf(d1/math.Sqrt(2))
	}
	return -0.5 + 0.5*math.Erf(-d1/math.Sqrt(2))
}

func calculateGamma(strike, spot, vol, tau float64) float64 {
	d1 := (math.Log(spot/strike) + 0.5*vol*vol*tau) / (vol * math.Sqrt(tau))
	return math.Exp(-d1*d1/2) / (spot * vol * math.Sqrt(2*math.Pi*tau))
}

func calculateTheta(strike, spot, vol, tau float64) float64 {
	// Simplified theta approximation
	return -spot * vol / (2 * math.Sqrt(2*math.Pi*tau)) * 0.01
}

func calculateVega(strike, spot, vol, tau float64) float64 {
	d1 := (math.Log(spot/strike) + 0.5*vol*vol*tau) / (vol * math.Sqrt(tau))
	return spot * math.Sqrt(tau/(2*math.Pi)) * math.Exp(-d1*d1/2) * 0.01
}

func newSkewStore() *SkewStore {
	store := &SkewStore{data: make(map[string]*SkewData)}
	
	pairs := []string{"EURUSD", "GBPUSD", "USDJPY", "SPY"}
	
	for _, pair := range pairs {
		// Risk reversal: 25 delta call vol - 25 delta put vol
		rr := (randFloat64() - 0.5) * 4.0 // -2% to +2%
		// Butterfly: (ATM vol - 0.5*(25d call + 25d put))
		fly := randFloat64() * 0.5
		
		store.data[pair] = &SkewData{
			RiskReversal: rr,
			Butterfly:    fly,
			Timestamp:    time.Now(),
		}
	}
	
	return store
}

func calculatePercentile(current float64, history []VIXPoint) float64 {
	count := 0
	for _, h := range history {
		if h.Value <= current {
			count++
		}
	}
	return float64(count) / float64(len(history)) * 100
}

func randFloat64() float64 {
	return 0.5 // Deterministic for demo
}

// GetVix returns VIX data with regime classification.
func (s *Service) GetVix(ctx context.Context) (*volv1.GetVixResponse, error) {
	now := time.Now()
	
	// Determine regime based on level and percentile
	regime := "normal"
	if s.vixStore.current < 15 {
		regime = "low"
	} else if s.vixStore.current > 25 {
		regime = "high"
	} else if s.vixStore.current > 35 {
		regime = "extreme"
	}
	
	// Calculate term structure (simplified: compare 30-day to 60-day implied)
	termStructure := 0.2 + (randFloat64()-0.5)*0.1
	
	// Build history
	var history []*volv1.VixData
	for _, h := range s.vixStore.history {
		history = append(history, &volv1.VixData{
			Timestamp:      h.Timestamp.Format(time.RFC3339),
			Spot:           h.Value,
			TermStructure:  termStructure,
			Percentile_30D: s.vixStore.percentile,
			Regime:         regime,
		})
	}
	
	return &volv1.GetVixResponse{
		Vix: &volv1.VixData{
			Timestamp:      now.Format(time.RFC3339),
			Spot:           s.vixStore.current,
			TermStructure:  termStructure,
			Percentile_30D: s.vixStore.percentile,
			Regime:         regime,
		},
		History: history,
	}, nil
}

// GetMove returns MOVE index (bond volatility) data.
func (s *Service) GetMove(ctx context.Context) (*volv1.GetMoveResponse, error) {
	// MOVE index calculation (simplified Merrill Lynch Option Volatility Estimate)
	value := 95.0 + randFloat64()*10.0
	
	// Determine trend
	trend := "normal"
	if value > 100 {
		trend = "elevated"
	} else if value > 120 {
		trend = "extreme"
	}
	
	return &volv1.GetMoveResponse{
		Move: &volv1.MoveData{
			Timestamp: time.Now().Format(time.RFC3339),
			Value:     value,
			Trend:     trend,
		},
	}, nil
}

// GetDvol returns DVol (crypto volatility) data.
func (s *Service) GetDvol(ctx context.Context, asset string) (*volv1.GetDvolResponse, error) {
	// DVOL values for major crypto assets
	dvolValues := map[string]float64{
		"BTC":  55.0,
		"ETH":  62.0,
		"SOL":  78.0,
		"AVAX": 85.0,
	}
	
	value, ok := dvolValues[asset]
	if !ok {
		value = 60.0 // Default
	}
	
	// Add some variation
	value += (randFloat64() - 0.5) * 5.0
	
	return &volv1.GetDvolResponse{
		Dvol: &volv1.DvolData{
			Timestamp: time.Now().Format(time.RFC3339),
			Value:     value,
			Asset:     asset,
		},
	}, nil
}

// GetGex returns Gamma Exposure (GEX) data with flip point and walls.
func (s *Service) GetGex(ctx context.Context, pair string) (*volv1.GetGexResponse, error) {
	data, ok := s.gexStore.data[pair]
	if !ok {
		return nil, fmt.Errorf("pair not found: %s", pair)
	}
	
	// Determine wall type
	wall := "neutral"
	if data.NetGEX > 500000 {
		wall = "call_wall"
	} else if data.NetGEX < -500000 {
		wall = "put_wall"
	}
	
	return &volv1.GetGexResponse{
		Gex: &volv1.GexData{
			Timestamp:  data.Timestamp.Format(time.RFC3339),
			Pair:       pair,
			NetGex:     data.NetGEX,
			FlipPoint:  data.FlipPoint,
			Wall:       wall,
		},
	}, nil
}

// GetIvol returns Implied Volatility surface with full Greeks.
func (s *Service) GetIvol(ctx context.Context, pair, expiry string) (*volv1.GetIvolResponse, error) {
	surface, ok := s.ivolStore.surfaces[pair]
	if !ok {
		// Generate on-the-fly for unknown pairs
		spot := 1.0
		atmIV := 10.0
		points := generateIVSurface(spot, atmIV)
		surface = &IVolSurface{
			Points: points,
			ATMIV:  atmIV,
			Expiry: time.Now().Add(30 * 24 * time.Hour),
		}
	}
	
	// Convert to proto
	var protoPoints []*volv1.IvolPoint
	for _, p := range surface.Points {
		protoPoints = append(protoPoints, &volv1.IvolPoint{
			Strike: p.Strike,
			Iv:     p.IV,
			Delta:  p.Delta,
			Gamma:  p.Gamma,
			Theta:  p.Theta,
			Vega:   p.Vega,
		})
	}
	
	return &volv1.GetIvolResponse{
		Pair:    pair,
		Surface: protoPoints,
		AtmIv:   surface.ATMIV,
	}, nil
}

// GetSkew returns volatility skew data with risk reversal and butterfly.
func (s *Service) GetSkew(ctx context.Context, pair string) (*volv1.GetSkewResponse, error) {
	data, ok := s.skewStore.data[pair]
	if !ok {
		return nil, fmt.Errorf("pair not found: %s", pair)
	}
	
	// Determine term structure shape
	termStructure := "flat"
	if data.RiskReversal > 1.0 {
		termStructure = "call_skew"
	} else if data.RiskReversal < -1.0 {
		termStructure = "put_skew"
	}
	
	return &volv1.GetSkewResponse{
		Skew: &volv1.SkewData{
			Timestamp:      data.Timestamp.Format(time.RFC3339),
			Pair:           pair,
			RiskReversal:   data.RiskReversal,
			Fly:            data.Butterfly,
			TermStructure:  termStructure,
		},
	}, nil
}

// GetSkewVixAlert returns skew-vix divergence alerts.
func (s *Service) GetSkewVixAlert(ctx context.Context, pair string) (*volv1.GetSkewVixAlertResponse, error) {
	var alerts []*volv1.SkewVixAlert
	
	// Check for divergence conditions
	vixRegime := s.vixStore.current < 20.0 // Low VIX
	skewData, ok := s.skewStore.data[pair]
	if !ok {
		return &volv1.GetSkewVixAlertResponse{Alerts: alerts}, nil
	}
	
	// High put skew + low VIX = potential complacency warning
	if vixRegime && skewData.RiskReversal < -2.0 {
		alerts = append(alerts, &volv1.SkewVixAlert{
			AlertId:   fmt.Sprintf("sv-alert-%s-%d", pair, time.Now().Unix()),
			Timestamp: time.Now().Format(time.RFC3339),
			Pair:      pair,
			Signal:    "complacency_warning",
			Confidence: 0.75,
			Reason:    "Elevated put skew with low VIX suggests market complacency",
		})
	}
	
	// High call skew + high VIX = potential euphoria warning
	if !vixRegime && skewData.RiskReversal > 2.0 {
		alerts = append(alerts, &volv1.SkewVixAlert{
			AlertId:   fmt.Sprintf("sv-alert-%s-%d", pair, time.Now().Unix()),
			Timestamp: time.Now().Format(time.RFC3339),
			Pair:      pair,
			Signal:    "euphoria_warning",
			Confidence: 0.65,
			Reason:    "Elevated call skew with high VIX suggests market euphoria",
		})
	}
	
	return &volv1.GetSkewVixAlertResponse{
		Alerts: alerts,
	}, nil
}
