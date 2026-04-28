package ai

import (
	"context"
	"fmt"
	"strings"
	"time"

	aiv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
)

// defaultScopes 默认聚合的领域。
var defaultScopes = []string{"macro", "onchain", "options", "treasury"}

// BuildContext 按 asset/scope 聚合数据库中已有的最新指标，
// 输出可直接注入 Prompt 的紧凑文本与结构化 sections。
func (s *Service) BuildContext(ctx context.Context, asset string, scope []string, locale string) (*aiv1.BuildContextResponse, error) {
	if asset == "" {
		return nil, fmt.Errorf("ai build_context: asset is required")
	}
	if len(scope) == 0 {
		scope = defaultScopes
	}
	resp := &aiv1.BuildContextResponse{
		Asset:       asset,
		Sections:    make(map[string]*aiv1.ContextSection, len(scope)),
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
	}
	if s.pool == nil {
		return resp, nil
	}

	for _, sc := range scope {
		key := strings.ToLower(strings.TrimSpace(sc))
		var section *aiv1.ContextSection
		switch key {
		case "macro":
			section = s.buildMacroSection(ctx)
		case "onchain":
			section = s.buildOnchainSection(ctx, asset)
		case "options":
			section = s.buildOptionsSection(ctx, asset)
		case "treasury":
			section = s.buildTreasurySection(ctx)
		default:
			continue
		}
		if section != nil {
			resp.Sections[key] = section
		}
	}
	resp.PromptReady = renderPromptReady(asset, resp.Sections, locale)
	return resp, nil
}

// buildMacroSection 取最近一次 macro_regime_history。
func (s *Service) buildMacroSection(ctx context.Context) *aiv1.ContextSection {
	sec := &aiv1.ContextSection{Scope: "macro", Indicators: map[string]string{}}
	var regime string
	var score float64
	var t time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT regime, score, time FROM macro_regime_history
		 ORDER BY time DESC LIMIT 1`).Scan(&regime, &score, &t)
	if err != nil {
		sec.Notes = []string{"macro_regime_history 暂无数据"}
		return sec
	}
	sec.Indicators["regime"] = regime
	sec.Indicators["score"] = fmt.Sprintf("%.3f", score)
	sec.Indicators["as_of"] = t.UTC().Format(time.RFC3339)
	return sec
}

// buildOnchainSection 取最近 onchain_metrics 一条记录的关键字段。
func (s *Service) buildOnchainSection(ctx context.Context, asset string) *aiv1.ContextSection {
	sec := &aiv1.ContextSection{Scope: "onchain", Indicators: map[string]string{}}
	rows, err := s.pool.Query(ctx, `
		SELECT metric_name, metric_value, time FROM onchain_metrics
		 WHERE asset = $1 AND time >= NOW() - INTERVAL '14 days'
		 ORDER BY time DESC LIMIT 30`, strings.ToUpper(asset))
	if err != nil {
		sec.Notes = []string{"onchain_metrics 查询失败"}
		return sec
	}
	defer rows.Close()
	seen := map[string]bool{}
	for rows.Next() {
		var name string
		var val float64
		var t time.Time
		if err := rows.Scan(&name, &val, &t); err != nil {
			continue
		}
		if seen[name] {
			continue
		}
		sec.Indicators[name] = fmt.Sprintf("%.4f", val)
		seen[name] = true
	}
	if len(sec.Indicators) == 0 {
		sec.Notes = []string{fmt.Sprintf("onchain_metrics 无 %s 近期数据", asset)}
	}
	return sec
}

// buildOptionsSection 取最近 GEX 与 DVOL。
func (s *Service) buildOptionsSection(ctx context.Context, asset string) *aiv1.ContextSection {
	sec := &aiv1.ContextSection{Scope: "options", Indicators: map[string]string{}}
	upper := strings.ToUpper(asset)

	var spot, totalGex, zeroGamma float64
	var t time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT spot_price, total_gex, zero_gamma_strike, time FROM gex_snapshots
		 WHERE symbol = $1 ORDER BY time DESC LIMIT 1`, upper).Scan(&spot, &totalGex, &zeroGamma, &t)
	if err == nil {
		sec.Indicators["spot"] = fmt.Sprintf("%.2f", spot)
		sec.Indicators["total_gex"] = fmt.Sprintf("%.2f", totalGex)
		sec.Indicators["zero_gamma"] = fmt.Sprintf("%.2f", zeroGamma)
		sec.Indicators["gex_as_of"] = t.UTC().Format(time.RFC3339)
	}

	var iv float64
	err = s.pool.QueryRow(ctx, `
		SELECT current_iv FROM dvol_snapshots
		 WHERE currency = $1 ORDER BY time DESC LIMIT 1`, upper).Scan(&iv)
	if err == nil {
		sec.Indicators["dvol_iv"] = fmt.Sprintf("%.2f", iv)
	}

	if len(sec.Indicators) == 0 {
		sec.Notes = []string{fmt.Sprintf("options 表无 %s 近期数据", asset)}
	}
	return sec
}

// buildTreasurySection 取最近 BIS 政策利率（USD）作为基准。
func (s *Service) buildTreasurySection(ctx context.Context) *aiv1.ContextSection {
	sec := &aiv1.ContextSection{Scope: "treasury", Indicators: map[string]string{}}
	var rate float64
	var d time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT rate, date FROM bis_policy_rates
		 WHERE currency = 'USD' ORDER BY date DESC LIMIT 1`).Scan(&rate, &d)
	if err == nil {
		sec.Indicators["usd_policy_rate"] = fmt.Sprintf("%.3f", rate)
		sec.Indicators["as_of"] = d.UTC().Format("2006-01-02")
		return sec
	}
	sec.Notes = []string{"bis_policy_rates USD 暂无数据"}
	return sec
}

// renderPromptReady 把 sections 序列化为紧凑可读文本，便于下游 Prompt 直接注入。
func renderPromptReady(asset string, sections map[string]*aiv1.ContextSection, locale string) string {
	var b strings.Builder
	if locale == "zh" || locale == "zh-CN" {
		fmt.Fprintf(&b, "标的：%s\n", asset)
	} else {
		fmt.Fprintf(&b, "Asset: %s\n", asset)
	}
	for _, scope := range []string{"macro", "onchain", "options", "treasury"} {
		sec, ok := sections[scope]
		if !ok {
			continue
		}
		fmt.Fprintf(&b, "\n[%s]\n", strings.ToUpper(scope))
		for k, v := range sec.Indicators {
			fmt.Fprintf(&b, "- %s: %s\n", k, v)
		}
		for _, n := range sec.Notes {
			fmt.Fprintf(&b, "- note: %s\n", n)
		}
	}
	return b.String()
}
