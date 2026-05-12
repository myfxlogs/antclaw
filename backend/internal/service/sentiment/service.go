package sentiment

import (
	"context"
	"fmt"
	"strings"
	"time"

	sentv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Service 情绪服务。优先读 DB：sentiment_snapshots / onchain_metrics / defi_snapshots。
// pool=nil 时返回可读的默认值（所有指标 0），不使用随机数。
type Service struct {
	pool *pgxpool.Pool
}

// NewService 兼容旧调用，pool=nil。
func NewService() *Service { return &Service{} }

// NewServiceWithPool 推荐的构造路径。
func NewServiceWithPool(pool *pgxpool.Pool) *Service { return &Service{pool: pool} }

// scoreToFG 把 -100..100 score 映射为 0..100 fear_greed（兼容 alternative.me）。
// 当前 GetSentiment 直接读 fear_greed 列，本函数保留供未来反向归一化时使用。
func scoreToFG(score float64) float64 { return (score + 100) / 2 }

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

// GetSentiment 从 sentiment_snapshots 读取最近一条真实 fear_greed，赋予全资产复用（F&G 是跨资产加密情绪指标）。
// score: -100..100 已存 score 列；fear_greed: 0..100。
func (s *Service) GetSentiment(ctx context.Context, asset string) (*sentv1.GetSentimentResponse, error) {
	asset = strings.ToUpper(strings.TrimSpace(asset))
	if asset == "" {
		asset = "BTC"
	}
	if s.pool == nil {
		return nil, fmt.Errorf("sentiment: postgres pool not configured")
	}
	var (
		t              time.Time
		score, fg, pcp float64
		regime         string
	)
	err := s.pool.QueryRow(ctx, `
		SELECT time, COALESCE(score,0), COALESCE(fear_greed,0), COALESCE(pc_percentile,0), COALESCE(regime,'')
		  FROM sentiment_snapshots
		 ORDER BY time DESC LIMIT 1`).Scan(&t, &score, &fg, &pcp, &regime)
	if err != nil {
		return nil, fmt.Errorf("sentiment: %w", err)
	}
	norm := score / 100 // 返回 -1..1
	now := t.UTC().Format(time.RFC3339)
	primary := &sentv1.SentimentData{
		Asset:     asset,
		Score:     norm,
		Label:     sentimentLabel(norm),
		Source:    "composite",
		Timestamp: now,
	}
	// 各子项：fear_greed 是真实拉取的；social / derivatives 在未接入前留空。
	components := []*sentv1.SentimentData{
		{Asset: asset, Score: (fg/100)*2 - 1, Label: sentimentLabel((fg/100)*2 - 1), Source: "fear_greed", Timestamp: now},
		{Asset: asset, Score: 0, Label: "neutral", Source: "social", Timestamp: now},
		{Asset: asset, Score: pcp/100*2 - 1, Label: sentimentLabel(pcp/100*2 - 1), Source: "derivatives", Timestamp: now},
	}
	_ = scoreToFG // 引用保留
	return &sentv1.GetSentimentResponse{Sentiment: primary, Components: components}, nil
}

// GetOnchain 从 onchain_metrics 读取最近一天 asset 的各项指标（宝塔表）。
func (s *Service) GetOnchain(ctx context.Context, asset string) (*sentv1.GetOnchainResponse, error) {
	asset = strings.ToUpper(strings.TrimSpace(asset))
	if s.pool == nil {
		return nil, fmt.Errorf("onchain: postgres pool not configured")
	}
	// onchain_metrics 现行 schema 为长表 (time, asset, metric, value)，每个时间点一组记录。
	rows, err := s.pool.Query(ctx, `
		SELECT metric, value FROM onchain_metrics
		 WHERE asset = $1
		   AND time = (SELECT MAX(time) FROM onchain_metrics WHERE asset = $1)`, asset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	metrics := []*sentv1.OnchainMetric{}
	for rows.Next() {
		var name string
		var val float64
		if err := rows.Scan(&name, &val); err == nil {
			metrics = append(metrics, &sentv1.OnchainMetric{Name: name, Value: val, Trend: "stable"})
		}
	}
	return &sentv1.GetOnchainResponse{
		Asset:   asset,
		Metrics: metrics,
		Signal:  deriveOnchainSignal(metrics),
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

// GetDefiHealth 默认读 defi_snapshots 最近一条总体 TVL 与变化，输出以“全生态”为单一 protocol 的调用接口形状。
// 如需各链明细，调 DeFiService.GetTopProtocols。
func (s *Service) GetDefiHealth(ctx context.Context, chain string) (*sentv1.GetDefiHealthResponse, error) {
	if s.pool == nil {
		return nil, fmt.Errorf("defi: postgres pool not configured")
	}
	var tvl, c24, c7 float64
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(total_tvl,0), COALESCE(tvl_change_24h,0), COALESCE(tvl_change_7d,0)
		  FROM defi_snapshots ORDER BY time DESC LIMIT 1`).Scan(&tvl, &c24, &c7)
	if err != nil {
		return nil, fmt.Errorf("defi: %w", err)
	}
	prot := []*sentv1.DefiMetric{{
		Protocol:        "DeFi (全生态)",
		Tvl:             fmt.Sprintf("%.2fB", tvl/1e9),
		TvlChange_24H:   fmt.Sprintf("%+.2f%%", c24),
		UtilizationRate: 0,
		HealthScore:     healthScoreFromChange(c24, c7),
	}}
	return &sentv1.GetDefiHealthResponse{
		Chain:         chain,
		Protocols:     prot,
		OverallHealth: calculateOverallHealth(prot),
	}, nil
}

func healthScoreFromChange(c24, c7 float64) string {
	switch {
	case c7 > 5:
		return "A+"
	case c7 > 0:
		return "A"
	case c24 > -2:
		return "B+"
	case c24 > -5:
		return "B"
	default:
		return "C"
	}
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

// GetCarryMonitor 该能力需要实时现货与期货价格。在未接入 FX/Crypto 期货牌价前，返回说明性错误，
// 以避免上下游误以为合成 carry 价为实盘。
func (s *Service) GetCarryMonitor(ctx context.Context, category string) (*sentv1.GetCarryMonitorResponse, error) {
	return nil, fmt.Errorf("sentiment carry: live FX/crypto futures price feed not yet wired (category=%q)", category)
}
