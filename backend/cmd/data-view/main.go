// 数据查看器 - 展示所有采集的数据
package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║              📊 AntClaw 数据采集总览                          ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	dbpool, err := pgxpool.New(context.Background(), "postgres://antclaw:antclaw@localhost:5432/antclaw")
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer dbpool.Close()

	ctx := context.Background()

	// 1. 财经日历
	showCalendar(ctx, dbpool)

	// 2. COT数据
	showCOT(ctx, dbpool)

	// 3. 宏观数据
	showMacro(ctx, dbpool)

	// 4. 价格数据
	showPrice(ctx, dbpool)

	// 5. 所有表统计
	showAllTables(ctx, dbpool)
}

func showCalendar(ctx context.Context, dbpool *pgxpool.Pool) {
	fmt.Println("┌─────────────────────────────────────────────────────────────────┐")
	fmt.Println("│ 📅 财经日历 (calendar_events)                                   │")
	fmt.Println("└─────────────────────────────────────────────────────────────────┘")

	var total int
	dbpool.QueryRow(ctx, "SELECT COUNT(*) FROM calendar_events").Scan(&total)
	fmt.Printf("总记录数: %d\n\n", total)

	// 按影响等级统计
	rows, _ := dbpool.Query(ctx, `
		SELECT impact, COUNT(*) as cnt 
		FROM calendar_events 
		GROUP BY impact 
		ORDER BY cnt DESC
	`)
	fmt.Println("影响等级分布:")
	for rows.Next() {
		var impact string
		var cnt int
		rows.Scan(&impact, &cnt)
		fmt.Printf("  %-10s %s %d\n", impact, bar(cnt, total), cnt)
	}
	rows.Close()

	// 按货币统计
	rows, _ = dbpool.Query(ctx, `
		SELECT currency, COUNT(*) as cnt 
		FROM calendar_events 
		WHERE currency != ''
		GROUP BY currency 
		ORDER BY cnt DESC 
		LIMIT 8
	`)
	fmt.Println("\n货币分布 (前8):")
	for rows.Next() {
		var curr string
		var cnt int
		rows.Scan(&curr, &cnt)
		fmt.Printf("  %-6s %s %d\n", curr, bar(cnt, total), cnt)
	}
	rows.Close()

	// 最新高影响事件
	fmt.Println("\n最新高影响事件:")
	rows, _ = dbpool.Query(ctx, `
		SELECT title, currency, forecast_value, actual_value, previous_value
		FROM calendar_events 
		WHERE impact = 'high'
		ORDER BY updated_at DESC 
		LIMIT 5
	`)
	for rows.Next() {
		var title, curr, forecast, actual, previous string
		rows.Scan(&title, &curr, &forecast, &actual, &previous)
		display := title
		if len(display) > 40 {
			display = display[:37] + "..."
		}
		fmt.Printf("  [%s] %-40s\n", curr, display)
		if actual != "" || forecast != "" {
			fmt.Printf("      公布:%s | 预期:%s | 前值:%s\n", actual, forecast, previous)
		}
	}
	rows.Close()
	fmt.Println()
}

func showCOT(ctx context.Context, dbpool *pgxpool.Pool) {
	fmt.Println("┌─────────────────────────────────────────────────────────────────┐")
	fmt.Println("│ 📊 COT持仓数据 (cot_records)                                    │")
	fmt.Println("└─────────────────────────────────────────────────────────────────┘")

	var total int
	err := dbpool.QueryRow(ctx, "SELECT COUNT(*) FROM cot_records").Scan(&total)
	if err != nil || total == 0 {
		fmt.Println("暂无数据 (表已创建，但未采集)")
		fmt.Println()
		return
	}
	fmt.Printf("总记录数: %d\n", total)

	// 按合约统计
	rows, _ := dbpool.Query(ctx, `
		SELECT contract_code, COUNT(*) as cnt 
		FROM cot_records 
		GROUP BY contract_code 
		ORDER BY cnt DESC 
		LIMIT 5
	`)
	fmt.Println("合约分布:")
	for rows.Next() {
		var code string
		var cnt int
		rows.Scan(&code, &cnt)
		fmt.Printf("  %-15s %s %d\n", code, bar(cnt, total), cnt)
	}
	rows.Close()
	fmt.Println()
}

func showMacro(ctx context.Context, dbpool *pgxpool.Pool) {
	fmt.Println("┌─────────────────────────────────────────────────────────────────┐")
	fmt.Println("│ 📈 宏观数据 (data_snapshots)                                    │")
	fmt.Println("└─────────────────────────────────────────────────────────────────┘")

	var total int
	err := dbpool.QueryRow(ctx, "SELECT COUNT(*) FROM data_snapshots").Scan(&total)
	if err != nil || total == 0 {
		fmt.Println("暂无数据 (表已创建，但未采集)")
		fmt.Println()
		return
	}
	fmt.Printf("总记录数: %d\n", total)

	// 按系列统计
	rows, _ := dbpool.Query(ctx, `
		SELECT series_id, COUNT(*) as cnt 
		FROM data_snapshots 
		GROUP BY series_id 
		ORDER BY cnt DESC 
		LIMIT 5
	`)
	fmt.Println("数据系列:")
	for rows.Next() {
		var series string
		var cnt int
		rows.Scan(&series, &cnt)
		fmt.Printf("  %-20s %s %d\n", series, bar(cnt, total), cnt)
	}
	rows.Close()
	fmt.Println()
}

func showPrice(ctx context.Context, dbpool *pgxpool.Pool) {
	fmt.Println("┌─────────────────────────────────────────────────────────────────┐")
	fmt.Println("│ 💹 价格数据 (price_daily)                                       │")
	fmt.Println("└─────────────────────────────────────────────────────────────────┘")

	var total int
	err := dbpool.QueryRow(ctx, "SELECT COUNT(*) FROM price_daily").Scan(&total)
	if err != nil || total == 0 {
		fmt.Println("暂无数据 (表已创建，但未采集)")
		fmt.Println()
		return
	}
	fmt.Printf("总记录数: %d\n", total)

	// 按品种统计
	rows, _ := dbpool.Query(ctx, `
		SELECT symbol, COUNT(*) as cnt 
		FROM price_daily 
		GROUP BY symbol 
		ORDER BY cnt DESC 
		LIMIT 5
	`)
	fmt.Println("品种分布:")
	for rows.Next() {
		var symbol string
		var cnt int
		rows.Scan(&symbol, &cnt)
		fmt.Printf("  %-10s %s %d\n", symbol, bar(cnt, total), cnt)
	}
	rows.Close()
	fmt.Println()
}

func showAllTables(ctx context.Context, dbpool *pgxpool.Pool) {
	fmt.Println("┌─────────────────────────────────────────────────────────────────┐")
	fmt.Println("│ 📁 所有数据表状态                                               │")
	fmt.Println("└─────────────────────────────────────────────────────────────────┘")

	tables := []struct {
		name  string
		desc  string
	}{
		{"calendar_events", "财经日历事件"},
		{"cot_records", "COT持仓数据"},
		{"cot_analyses", "COT分析结果"},
		{"price_daily", "日K价格数据"},
		{"price_intraday", "分时价格数据"},
		{"data_snapshots", "FRED宏观数据"},
		{"macro_regime_history", "宏观状态历史"},
		{"sentiment_snapshots", "情绪指标"},
		{"onchain_metrics", "链上数据"},
		{"defi_snapshots", "DeFi数据"},
		{"vix_term_structure", "VIX期限结构"},
		{"dvol_snapshots", "DVOL数据"},
		{"backtest_jobs", "回测任务"},
		{"backtest_results", "回测结果"},
		{"flow_divergence_history", "资金流向"},
		{"volume_profiles", "成交量分布"},
	}

	fmt.Println("表名                          描述                        记录数    状态")
	fmt.Println(strings.Repeat("─", 80))

	for _, t := range tables {
		var count int
		err := dbpool.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", t.name)).Scan(&count)
		status := "✓"
		if err != nil {
			status = "✗"
			count = 0
		}
		desc := t.desc
		if len(desc) > 25 {
			desc = desc[:22] + "..."
		}
		name := t.name
		if len(name) > 28 {
			name = name[:25] + "..."
		}

		if count > 0 {
			fmt.Printf("%-28s %-28s %6d    %s\n", name, desc, count, status)
		} else {
			fmt.Printf("%-28s %-28s %6s    %s\n", name, desc, "-", status)
		}
	}

	fmt.Println()
	fmt.Println("图例: ✓ 表存在  |  - 无数据")
}

func bar(value, total int) string {
	if total == 0 {
		return ""
	}
	width := 20
	filled := int(float64(value) * float64(width) / float64(total))
	if filled == 0 && value > 0 {
		filled = 1
	}
	return strings.Repeat("█", filled)
}
