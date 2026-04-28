// 数据详情查看器 - 展示所有采集数据的详细内容
package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║              📊 AntClaw 数据详情查看器                        ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	dbpool, err := pgxpool.New(context.Background(), "postgres://antclaw:antclaw@localhost:5432/antclaw")
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer dbpool.Close()

	ctx := context.Background()

	// 1. COT分析详情
	showCOTAnalysis(ctx, dbpool)

	// 2. 宏观状态
	showMacroRegime(ctx, dbpool)

	// 3. VIX期限结构
	showVIXTerm(ctx, dbpool)

	// 4. DVOL数据
	showDVOL(ctx, dbpool)

	// 5. DeFi数据
	showDeFi(ctx, dbpool)

	// 6. 资金流向
	showFlowDivergence(ctx, dbpool)

	// 7. 成交量分布
	showVolumeProfile(ctx, dbpool)

	// 8. 链上数据
	showOnchain(ctx, dbpool)

	// 9. 情绪数据
	showSentiment(ctx, dbpool)
}

func showCOTAnalysis(ctx context.Context, dbpool *pgxpool.Pool) {
	fmt.Println("┌─────────────────────────────────────────────────────────────────┐")
	fmt.Println("│ 📊 COT分析结果 (cot_analyses) - COT Index / Z-score / 百分位   │")
	fmt.Println("└─────────────────────────────────────────────────────────────────┘")

	rows, _ := dbpool.Query(ctx, `
		SELECT contract_code, net_position, cot_index, direction, 
		       zscore, percentile, report_date
		FROM cot_analyses
		ORDER BY report_date DESC, cot_index DESC
		LIMIT 15
	`)
	fmt.Printf("%-10s %12s %8s %12s %8s %10s %12s\n", 
		"合约", "净头寸", "COTIdx", "方向", "Z-Score", "百分位", "报告日期")
	fmt.Println(strings.Repeat("─", 85))
	for rows.Next() {
		var code, dir string
		var net int64
		var idx, zscore, pct float64
		var date time.Time
		rows.Scan(&code, &net, &idx, &dir, &zscore, &pct, &date)
		fmt.Printf("%-10s %12d %7.1f%% %-12s %7.2f %9.1f%% %s\n",
			code, net, idx, dir, zscore, pct, date.Format("2006-01-02"))
	}
	rows.Close()
	fmt.Println()
}

func showMacroRegime(ctx context.Context, dbpool *pgxpool.Pool) {
	fmt.Println("┌─────────────────────────────────────────────────────────────────┐")
	fmt.Println("│ 📈 宏观状态历史 (macro_regime_history) - Risk-On/Off 分类     │")
	fmt.Println("└─────────────────────────────────────────────────────────────────┘")

	rows, _ := dbpool.Query(ctx, `
		SELECT time, regime, score, details
		FROM macro_regime_history
		ORDER BY time DESC
		LIMIT 12
	`)
	fmt.Printf("%-20s %-16s %8s %s\n", "时间", "状态", "评分", "详情")
	fmt.Println(strings.Repeat("─", 75))
	for rows.Next() {
		var t time.Time
		var regime string
		var score float64
		var details []byte
		rows.Scan(&t, &regime, &score, &details)
		emoji := "⚪"
		switch regime {
		case "risk_on", "mild_risk_on":
			emoji = "🟢"
		case "risk_off", "mild_risk_off":
			emoji = "🔴"
		}
		fmt.Printf("%-20s %s %-14s %7.1f\n", t.Format("2006-01-02 15:04"), emoji, regime, score)
	}
	rows.Close()
	fmt.Println()
}

func showVIXTerm(ctx context.Context, dbpool *pgxpool.Pool) {
	fmt.Println("┌─────────────────────────────────────────────────────────────────┐")
	fmt.Println("│ 📉 VIX期限结构 (vix_term_structure)                           │")
	fmt.Println("└─────────────────────────────────────────────────────────────────┘")

	rows, _ := dbpool.Query(ctx, `
		SELECT time, spot, vix9d, vix3m, contango, cross_regime
		FROM vix_term_structure
		ORDER BY time DESC
		LIMIT 10
	`)
	fmt.Printf("%-20s %8s %8s %8s %10s %s\n", "时间", "VIX", "VIX9D", "VIX3M", "期限结构", "状态")
	fmt.Println(strings.Repeat("─", 75))
	for rows.Next() {
		var t time.Time
		var spot, v9, v3m float64
		var cont bool
		var regime string
		rows.Scan(&t, &spot, &v9, &v3m, &cont, &regime)
		term := "Backwardation"
		if cont {
			term = "Contango"
		}
		fmt.Printf("%-20s %8.2f %8.2f %8.2f %10s %s\n",
			t.Format("2006-01-02"), spot, v9, v3m, term, regime)
	}
	rows.Close()
	fmt.Println()
}

func showDVOL(ctx context.Context, dbpool *pgxpool.Pool) {
	fmt.Println("┌─────────────────────────────────────────────────────────────────┐")
	fmt.Println("│ 📊 DVOL隐含波动率 (dvol_snapshots) - Deribit                  │")
	fmt.Println("└─────────────────────────────────────────────────────────────────┘")

	// 最新数据
	fmt.Println("最新DVOL:")
	rows, _ := dbpool.Query(ctx, `
		SELECT time, currency, current_iv, change_24h_pct, spike
		FROM dvol_snapshots
		WHERE time >= NOW() - INTERVAL '1 day'
		ORDER BY currency, time DESC
	`)
	fmt.Printf("%-20s %-8s %10s %12s %s\n", "时间", "币种", "当前IV", "24h变化%", "Spike")
	fmt.Println(strings.Repeat("─", 65))
	for rows.Next() {
		var t time.Time
		var ccy string
		var iv, chg float64
		var spike bool
		rows.Scan(&t, &ccy, &iv, &chg, &spike)
		spikeStr := ""
		if spike {
			spikeStr = "⚡ SPIKE"
		}
		fmt.Printf("%-20s %-8s %9.2f%% %10.2f%% %s\n",
			t.Format("2006-01-02"), ccy, iv, chg, spikeStr)
	}
	rows.Close()

	// 历史统计
	fmt.Println("\nDVOL历史统计 (最近3个月):")
	rows, _ = dbpool.Query(ctx, `
		SELECT currency, 
		       AVG(current_iv) as avg_iv,
		       MAX(current_iv) as max_iv,
		       MIN(current_iv) as min_iv,
		       COUNT(*) as cnt
		FROM dvol_snapshots
		WHERE time >= NOW() - INTERVAL '90 days'
		GROUP BY currency
	`)
	fmt.Printf("%-8s %10s %10s %10s %8s\n", "币种", "平均IV", "最高IV", "最低IV", "样本数")
	fmt.Println(strings.Repeat("─", 55))
	for rows.Next() {
		var ccy string
		var avg, max, min float64
		var cnt int
		rows.Scan(&ccy, &avg, &max, &min, &cnt)
		fmt.Printf("%-8s %9.2f%% %9.2f%% %9.2f%% %8d\n", ccy, avg, max, min, cnt)
	}
	rows.Close()
	fmt.Println()
}

func showDeFi(ctx context.Context, dbpool *pgxpool.Pool) {
	fmt.Println("┌─────────────────────────────────────────────────────────────────┐")
	fmt.Println("│ 🏦 DeFi数据 (defi_snapshots) - DefiLlama                     │")
	fmt.Println("└─────────────────────────────────────────────────────────────────┘")

	// 最新TVL
	var latestTVL, chg24, chg7d, stableMC float64
	err := dbpool.QueryRow(ctx, `
		SELECT total_tvl, tvl_change_24h, tvl_change_7d, stablecoin_mc
		FROM defi_snapshots
		ORDER BY time DESC
		LIMIT 1
	`).Scan(&latestTVL, &chg24, &chg7d, &stableMC)
	if err == nil {
		fmt.Printf("当前总锁仓价值 (TVL): $%.2fB\n", latestTVL/1e9)
		fmt.Printf("24h变化: %+.2f%%  |  7d变化: %+.2f%%\n", chg24, chg7d)
		fmt.Printf("稳定币市值: $%.2fB\n\n", stableMC/1e9)
	}

	// 历史趋势
	fmt.Println("TVL历史趋势 (最近14天):")
	rows, _ := dbpool.Query(ctx, `
		SELECT time, total_tvl, tvl_change_24h
		FROM defi_snapshots
		ORDER BY time DESC
		LIMIT 14
	`)
	fmt.Printf("%-20s %15s %12s\n", "日期", "TVL", "24h变化%")
	fmt.Println(strings.Repeat("─", 50))
	for rows.Next() {
		var t time.Time
		var tvl, chg float64
		rows.Scan(&t, &tvl, &chg)
		fmt.Printf("%-20s $%12.2fB %10.2f%%\n",
			t.Format("2006-01-02"), tvl/1e9, chg)
	}
	rows.Close()
	fmt.Println()
}

func showFlowDivergence(ctx context.Context, dbpool *pgxpool.Pool) {
	fmt.Println("┌─────────────────────────────────────────────────────────────────┐")
	fmt.Println("│ 💰 资金流向背离 (flow_divergence_history) - Pearson相关性     │")
	fmt.Println("└─────────────────────────────────────────────────────────────────┘")

	rows, _ := dbpool.Query(ctx, `
		SELECT time, pair_a, pair_b, corr, z_score, lead_lag
		FROM flow_divergence_history
		ORDER BY time DESC
		LIMIT 16
	`)
	fmt.Printf("%-20s %-10s %-10s %8s %8s %8s\n", "时间", "货币对A", "货币对B", "相关系数", "Z-Score", "领先lag")
	fmt.Println(strings.Repeat("─", 80))
	for rows.Next() {
		var t time.Time
		var a, b string
		var corr, zscore float64
		var lag int
		rows.Scan(&t, &a, &b, &corr, &zscore, &lag)
		alert := ""
		if zscore > 2 || zscore < -2 {
			alert = "⚠️ 背离"
		}
		fmt.Printf("%-20s %-10s %-10s %8.3f %8.2f %8d %s\n",
			t.Format("2006-01-02"), a, b, corr, zscore, lag, alert)
	}
	rows.Close()
	fmt.Println()
}

func showVolumeProfile(ctx context.Context, dbpool *pgxpool.Pool) {
	fmt.Println("┌─────────────────────────────────────────────────────────────────┐")
	fmt.Println("│ 📊 成交量分布 (volume_profiles) - POC/VAH/VAL                 │")
	fmt.Println("└─────────────────────────────────────────────────────────────────┘")

	rows, _ := dbpool.Query(ctx, `
		SELECT symbol, poc, vah, val, period
		FROM volume_profiles
		ORDER BY symbol
	`)
	fmt.Printf("%-10s %15s %15s %15s %8s\n", "品种", "POC", "VAH", "VAL", "周期")
	fmt.Println(strings.Repeat("─", 70))
	for rows.Next() {
		var sym, period string
		var poc, vah, val float64
		rows.Scan(&sym, &poc, &vah, &val, &period)
		fmt.Printf("%-10s %15.5f %15.5f %15.5f %8s\n", sym, poc, vah, val, period)
	}
	rows.Close()
	fmt.Println()
	fmt.Println("POC: Point of Control (控制点/最大成交量价格)")
	fmt.Println("VAH: Value Area High (价值区高点 - 70%成交量上沿)")
	fmt.Println("VAL: Value Area Low (价值区低点 - 70%成交量下沿)")
	fmt.Println()
}

func showOnchain(ctx context.Context, dbpool *pgxpool.Pool) {
	fmt.Println("┌─────────────────────────────────────────────────────────────────┐")
	fmt.Println("│ ⛓️  链上数据 (onchain_metrics) - CoinGecko                    │")
	fmt.Println("└─────────────────────────────────────────────────────────────────┘")

	rows, _ := dbpool.Query(ctx, `
		SELECT date, asset, net_flow, onchain_score, active_addr, tx_count
		FROM onchain_metrics
		ORDER BY date DESC, asset
		LIMIT 15
	`)
	fmt.Printf("%-12s %-8s %12s %12s %12s %10s\n", "日期", "资产", "净流入", "链上评分", "活跃地址", "交易数")
	fmt.Println(strings.Repeat("─", 75))
	for rows.Next() {
		var date time.Time
		var asset string
		var flow, score float64
		var addr, tx int64
		rows.Scan(&date, &asset, &flow, &score, &addr, &tx)
		fmt.Printf("%-12s %-8s %11.2f %12.2f %12d %10d\n",
			date.Format("2006-01-02"), asset, flow, score, addr, tx)
	}
	rows.Close()
	fmt.Println()
}

func showSentiment(ctx context.Context, dbpool *pgxpool.Pool) {
	fmt.Println("┌─────────────────────────────────────────────────────────────────┐")
	fmt.Println("│ 🎭 情绪指标 (sentiment_snapshots) - Fear & Greed              │")
	fmt.Println("└─────────────────────────────────────────────────────────────────┘")

	rows, _ := dbpool.Query(ctx, `
		SELECT time, fear_greed, regime, score
		FROM sentiment_snapshots
		ORDER BY time DESC
		LIMIT 10
	`)
	fmt.Printf("%-20s %6s %16s %8s\n", "时间", "F&G值", "市场情绪", "评分")
	fmt.Println(strings.Repeat("─", 60))
	for rows.Next() {
		var t time.Time
		var fg, score float64
		var regime string
		rows.Scan(&t, &fg, &regime, &score)
		emoji := "😐"
		switch regime {
		case "extreme_fear":
			emoji = "😱"
		case "fear":
			emoji = "😰"
		case "greed":
			emoji = "🤑"
		case "extreme_greed":
			emoji = "🚀"
		}
		fmt.Printf("%-20s %5.0f%% %s %-14s %7.2f\n",
			t.Format("2006-01-02"), fg, emoji, regime, score)
	}
	rows.Close()
	fmt.Println()
}
