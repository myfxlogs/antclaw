package backtest

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/antclaw/antclaw/internal/domain/shared"
	"github.com/antclaw/antclaw/internal/infra/postgres"
)

// BacktestService provides backtesting operations
type BacktestService struct {
	priceRepo postgres.PriceRepository
	repo      postgres.BacktestRepository
	logger    *slog.Logger
}

// BacktestRequest represents a backtest request
type BacktestRequest struct {
	StrategyName string                 `json:"strategy_name"`
	Symbols      []string               `json:"symbols"`
	From         time.Time              `json:"from"`
	To           time.Time              `json:"to"`
	Parameters   map[string]interface{} `json:"parameters"`
}

// BacktestResult represents backtest results
type BacktestResult struct {
	StrategyName     string                 `json:"strategy_name"`
	Symbols          []string               `json:"symbols"`
	Period           string                 `json:"period"`
	TotalReturn      float64                `json:"total_return"`
	AnnualizedReturn float64                `json:"annualized_return"`
	SharpeRatio      float64                `json:"sharpe_ratio"`
	MaxDrawdown      float64                `json:"max_drawdown"`
	WinRate          float64                `json:"win_rate"`
	TotalTrades      int                    `json:"total_trades"`
	AvgTradeReturn   float64                `json:"avg_trade_return"`
	ProfitFactor     float64                `json:"profit_factor"`
	Trades           []Trade                `json:"trades"`
	EquityCurve      []EquityPoint          `json:"equity_curve"`
}

// Trade represents a single trade
type Trade struct {
	Symbol      string    `json:"symbol"`
	EntryTime   time.Time `json:"entry_time"`
	ExitTime    time.Time `json:"exit_time"`
	Direction   shared.Direction `json:"direction"`
	EntryPrice  float64   `json:"entry_price"`
	ExitPrice   float64   `json:"exit_price"`
	Quantity    float64   `json:"quantity"`
	Pnl         float64   `json:"pnl"`
	PnlPct      float64   `json:"pnl_pct"`
	ExitReason  string    `json:"exit_reason"`
}

// EquityPoint represents a point on the equity curve
type EquityPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Equity    float64   `json:"equity"`
	Drawdown  float64   `json:"drawdown"`
}

// NewBacktestService creates a new backtest service
func NewBacktestService(priceRepo postgres.PriceRepository, repo postgres.BacktestRepository, logger *slog.Logger) *BacktestService {
	return &BacktestService{
		priceRepo: priceRepo,
		repo:      repo,
		logger:    logger,
	}
}

// RunBacktest runs a backtest for a strategy
func (s *BacktestService) RunBacktest(ctx context.Context, req BacktestRequest) (*BacktestResult, error) {
	s.logger.Info("starting backtest", "strategy", req.StrategyName, "symbols", req.Symbols)

	// Fetch historical data
	var allBars []postgres.DailyBar
	for _, symbol := range req.Symbols {
		bars, err := s.priceRepo.GetDailyBars(ctx, symbol, req.From, req.To)
		if err != nil {
			return nil, fmt.Errorf("fetch bars for %s failed: %w", symbol, err)
		}
		allBars = append(allBars, bars...)
	}

	if len(allBars) == 0 {
		return nil, fmt.Errorf("no price data available")
	}

	// Run simulation
	trades := s.simulateTrades(allBars, req.Parameters)

	// Calculate metrics
	result := s.calculateMetrics(trades, allBars)
	result.StrategyName = req.StrategyName
	result.Symbols = req.Symbols
	result.Period = fmt.Sprintf("%s to %s", req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))

	s.logger.Info("backtest completed",
		"strategy", req.StrategyName,
		"trades", result.TotalTrades,
		"return", result.TotalReturn,
		"sharpe", result.SharpeRatio,
	)

	return result, nil
}

// simulateTrades simulates trades based on strategy logic
func (s *BacktestService) simulateTrades(bars []postgres.DailyBar, params map[string]interface{}) []Trade {
	var trades []Trade
	var position *Trade

	// Simple moving average crossover strategy
	shortPeriod := 10
	longPeriod := 30

	for i := longPeriod; i < len(bars); i++ {
		shortMA := s.calculateSMA(bars, i, shortPeriod)
		longMA := s.calculateSMA(bars, i, longPeriod)

		currentBar := bars[i]

		// Entry signal: short MA crosses above long MA
		if position == nil && shortMA > longMA && s.previousShortMA(bars, i, shortPeriod) <= s.previousLongMA(bars, i, longPeriod) {
			position = &Trade{
				Symbol:     currentBar.Symbol,
				EntryTime:  currentBar.Time,
				Direction:  shared.DirectionLong,
				EntryPrice: currentBar.Close,
				Quantity:   1.0,
			}
		}

		// Exit signal: short MA crosses below long MA
		if position != nil && shortMA < longMA && s.previousShortMA(bars, i, shortPeriod) >= s.previousLongMA(bars, i, longPeriod) {
			position.ExitTime = currentBar.Time
			position.ExitPrice = currentBar.Close
			position.Pnl = (position.ExitPrice - position.EntryPrice) * position.Quantity
			position.PnlPct = position.Pnl / position.EntryPrice * 100
			position.ExitReason = "Signal"
			
			if position.Direction == shared.DirectionShort {
				position.Pnl = -position.Pnl
				position.PnlPct = -position.PnlPct
			}
			
			trades = append(trades, *position)
			position = nil
		}
	}

	return trades
}

// calculateMetrics calculates backtest metrics
func (s *BacktestService) calculateMetrics(trades []Trade, bars []postgres.DailyBar) *BacktestResult {
	result := &BacktestResult{
		Trades: trades,
	}

	if len(trades) == 0 {
		return result
	}

	// Calculate returns
	var totalPnl, winningPnl, losingPnl float64
	var winners int

	for _, t := range trades {
		result.TotalTrades++
		totalPnl += t.Pnl
		
		if t.Pnl > 0 {
			winners++
			winningPnl += t.Pnl
		} else {
			losingPnl -= t.Pnl
		}
	}

	result.TotalReturn = totalPnl
	result.WinRate = float64(winners) / float64(len(trades)) * 100
	result.AvgTradeReturn = totalPnl / float64(len(trades))
	
	if losingPnl > 0 {
		result.ProfitFactor = winningPnl / losingPnl
	}

	// Calculate Sharpe ratio
	returns := s.calculateReturns(trades)
	result.SharpeRatio = s.calculateSharpe(returns)

	// Calculate max drawdown
	result.EquityCurve = s.calculateEquityCurve(trades)
	result.MaxDrawdown = s.calculateMaxDrawdown(result.EquityCurve)

	// Annualized return
	days := bars[len(bars)-1].Time.Sub(bars[0].Time).Hours() / 24
	if days > 0 {
		result.AnnualizedReturn = math.Pow(1+result.TotalReturn/100, 365/days) - 1
	}

	return result
}

// calculateSMA calculates Simple Moving Average
func (s *BacktestService) calculateSMA(bars []postgres.DailyBar, index, period int) float64 {
	if index < period {
		return 0
	}

	var sum float64
	for i := index - period; i < index; i++ {
		sum += bars[i].Close
	}
	return sum / float64(period)
}

// previousShortMA calculates previous short MA
func (s *BacktestService) previousShortMA(bars []postgres.DailyBar, index, period int) float64 {
	return s.calculateSMA(bars, index-1, period)
}

// previousLongMA calculates previous long MA
func (s *BacktestService) previousLongMA(bars []postgres.DailyBar, index, period int) float64 {
	return s.calculateSMA(bars, index-1, period)
}

// calculateReturns calculates trade returns
func (s *BacktestService) calculateReturns(trades []Trade) []float64 {
	var returns []float64
	for _, t := range trades {
		returns = append(returns, t.PnlPct)
	}
	return returns
}

// calculateSharpe calculates Sharpe ratio
func (s *BacktestService) calculateSharpe(returns []float64) float64 {
	if len(returns) == 0 {
		return 0
	}

	mean := s.mean(returns)
	stdDev := s.stdDev(returns, mean)
	
	if stdDev == 0 {
		return 0
	}

	return mean / stdDev * math.Sqrt(252) // Annualized
}

// calculateEquityCurve calculates equity curve
func (s *BacktestService) calculateEquityCurve(trades []Trade) []EquityPoint {
	var equity float64 = 10000 // Start with $10,000
	var maxEquity float64 = equity
	var points []EquityPoint

	for _, t := range trades {
		equity += t.Pnl
		if equity > maxEquity {
			maxEquity = equity
		}
		
		drawdown := (maxEquity - equity) / maxEquity * 100
		
		points = append(points, EquityPoint{
			Timestamp: t.ExitTime,
			Equity:    equity,
			Drawdown:  drawdown,
		})
	}

	return points
}

// calculateMaxDrawdown calculates maximum drawdown
func (s *BacktestService) calculateMaxDrawdown(equityCurve []EquityPoint) float64 {
	var maxDrawdown float64
	for _, p := range equityCurve {
		if p.Drawdown > maxDrawdown {
			maxDrawdown = p.Drawdown
		}
	}
	return maxDrawdown
}

// mean calculates arithmetic mean
func (s *BacktestService) mean(values []float64) float64 {
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
func (s *BacktestService) stdDev(values []float64, mean float64) float64 {
	if len(values) < 2 {
		return 0
	}

	var sumSquared float64
	for _, v := range values {
		diff := v - mean
		sumSquared += diff * diff
	}
	
	variance := sumSquared / float64(len(values)-1)
	return math.Sqrt(variance)
}

// RunMonteCarlo runs Monte Carlo simulation
func (s *BacktestService) RunMonteCarlo(ctx context.Context, trades []Trade, iterations int) (*MonteCarloResult, error) {
	if len(trades) < 10 {
		return nil, fmt.Errorf("insufficient trades for Monte Carlo")
	}

	// Extract returns
	var returns []float64
	for _, t := range trades {
		returns = append(returns, t.PnlPct)
	}

	// Run simulations
	var results []float64
	for i := 0; i < iterations; i++ {
		shuffled := s.shuffle(returns)
		finalReturn := s.simulateReturns(shuffled)
		results = append(results, finalReturn)
	}

	// Calculate percentiles
	sort.Float64s(results)
	
	mcResult := &MonteCarloResult{
		Iterations: iterations,
		P95:        results[int(float64(iterations)*0.95)],
		P50:        results[int(float64(iterations)*0.50)],
		P5:         results[int(float64(iterations)*0.05)],
	}

	return mcResult, nil
}

// MonteCarloResult represents Monte Carlo simulation results
type MonteCarloResult struct {
	Iterations int     `json:"iterations"`
	P95        float64 `json:"p95"`
	P50        float64 `json:"p50"`
	P5         float64 `json:"p5"`
}

// shuffle randomly shuffles returns
func (s *BacktestService) shuffle(returns []float64) []float64 {
	shuffled := make([]float64, len(returns))
	copy(shuffled, returns)
	
	// Fisher-Yates shuffle
	for i := len(shuffled) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}
	
	return shuffled
}

// simulateReturns simulates equity curve from returns
func (s *BacktestService) simulateReturns(returns []float64) float64 {
	equity := 100.0
	for _, r := range returns {
		equity *= (1 + r/100)
	}
	return equity - 100
}
