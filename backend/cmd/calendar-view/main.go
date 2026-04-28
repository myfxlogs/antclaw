// 财经日历中文版查看器
package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// 影响等级翻译
var impactCN = map[string]string{
	"high":   "★★★ 高",
	"medium": "★★☆ 中",
	"low":    "★☆☆ 低",
}

// 货币翻译
var currencyCN = map[string]string{
	"USD": "美元",
	"EUR": "欧元",
	"GBP": "英镑",
	"JPY": "日元",
	"AUD": "澳元",
	"CAD": "加元",
	"CHF": "瑞郎",
	"NZD": "纽元",
	"CNY": "人民币",
	"MXN": "墨西哥比索",
	"ZAR": "南非兰特",
	"KRW": "韩元",
	"BRL": "巴西雷亚尔",
	"INR": "印度卢比",
	"RUB": "俄罗斯卢布",
}

// 常见事件名称翻译
var eventCN = map[string]string{
	"Interest Rate Decision":                    "利率决议",
	"Federal Funds Rate":                          "联邦基金利率",
	"Non-Farm Payrolls":                           "非农就业人数",
	"Unemployment Rate":                           "失业率",
	"CPI":                                         "消费者物价指数",
	"GDP":                                         "国内生产总值",
	"Retail Sales":                                "零售销售",
	"Industrial Production":                       "工业生产",
	"Manufacturing PMI":                           "制造业PMI",
	"Services PMI":                                "服务业PMI",
	"Trade Balance":                               "贸易帐",
	"Current Account":                             "经常帐",
	"Consumer Confidence":                         "消费者信心指数",
	"Business Confidence":                         "商业信心指数",
	"ZEW Economic Sentiment":                    "ZEW经济景气指数",
	"IFO Business Climate":                        "IFO商业景气指数",
	"ECB Press Conference":                        "欧央行新闻发布会",
	"Fed Press Conference":                        "美联储新闻发布会",
	"Initial Jobless Claims":                      "初请失业金人数",
	"Continuing Claims":                           "续请失业金人数",
	"Building Permits":                            "建筑许可",
	"Housing Starts":                             "新屋开工",
	"Existing Home Sales":                         "成屋销售",
	"New Home Sales":                              "新屋销售",
	"Durable Goods Orders":                        "耐用品订单",
	"Factory Orders":                              "工厂订单",
	"Wholesale Inventories":                       "批发库存",
	"Retail Inventories":                          "零售库存",
	"Philadelphia Fed Manufacturing Index":        "费城联储制造业指数",
	"Empire State Manufacturing Index":            "纽约联储制造业指数",
	"Richmond Fed Manufacturing Index":          "里士满联储制造业指数",
	"Chicago PMI":                                 "芝加哥PMI",
	"Dallas Fed Manufacturing Index":              "达拉斯联储制造业指数",
	"Kansas City Fed Manufacturing Index":         "堪萨斯联储制造业指数",
	"Pending Home Sales":                          "成屋待完成销售",
	"Personal Income":                             "个人收入",
	"Personal Spending":                           "个人支出",
	"Core PCE Price Index":                        "核心PCE物价指数",
	"PCE Price Index":                             "PCE物价指数",
	"Average Hourly Earnings":                     "平均时薪",
	"Average Weekly Hours":                        "平均每周工时",
	"Labor Force Participation Rate":              "劳动参与率",
	"Capacity Utilization Rate":                   "产能利用率",
	"Producer Price Index":                        "生产者物价指数",
	"Core PPI":                                    "核心PPI",
	"Crude Oil Inventories":                       "原油库存",
	"Gasoline Inventories":                        "汽油库存",
	"Distillate Inventories":                      "精炼油库存",
	"Natural Gas Storage":                         "天然气库存",
	"Baker Hughes Oil Rig Count":                  "贝克休斯石油钻井数",
	"CFTC Net Positions":                          "CFTC持仓",
}

func main() {
	fmt.Println("╔════════════════════════════════════════╗")
	fmt.Println("║       📊 AntClaw 财经日历 (中文版)      ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Println()

	// 连接数据库
	dbpool, err := pgxpool.New(context.Background(), "postgres://antclaw:antclaw@localhost:5432/antclaw")
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer dbpool.Close()

	ctx := context.Background()

	// 获取高影响事件
	fmt.Println("【★★★ 高影响事件】")
	fmt.Println(strings.Repeat("─", 80))
	queryAndDisplay(ctx, dbpool, "high")

	// 获取中影响事件
	fmt.Println()
	fmt.Println("【★★☆ 中影响事件】")
	fmt.Println(strings.Repeat("─", 80))
	queryAndDisplay(ctx, dbpool, "medium")

	// 统计信息
	fmt.Println()
	fmt.Println("【📈 统计数据】")
	fmt.Println(strings.Repeat("─", 80))
	displayStats(ctx, dbpool)
}

func queryAndDisplay(ctx context.Context, dbpool *pgxpool.Pool, impact string) {
	rows, err := dbpool.Query(ctx, `
		SELECT title, currency, impact, scheduled_at, forecast_value, actual_value, previous_value
		FROM calendar_events
		WHERE impact = $1
		ORDER BY scheduled_at DESC
		LIMIT 10
	`, impact)
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var title, currency, impactLevel string
		var scheduledAt time.Time
		var forecast, actual, previous string

		err := rows.Scan(&title, &currency, &impactLevel, &scheduledAt, &forecast, &actual, &previous)
		if err != nil {
			continue
		}

		count++
		cnTitle := translateEvent(title)
		cnCurrency := translateCurrency(currency)
		cnImpact := translateImpact(impactLevel)

		// 格式化时间
		timeStr := scheduledAt.Format("01-02 15:04")
		if scheduledAt.Year() < 1000 {
			timeStr = "待定"
		}

		// 显示事件
		fmt.Printf("⏰ %s  [%s]  %s\n", timeStr, cnCurrency, cnTitle)
		fmt.Printf("   %s | ", cnImpact)

		// 显示数值
		if actual != "" {
			fmt.Printf("公布: %s | ", actual)
		}
		if forecast != "" {
			fmt.Printf("预期: %s | ", forecast)
		}
		if previous != "" && previous != actual {
			fmt.Printf("前值: %s", previous)
		}
		fmt.Println()
		fmt.Println()
	}

	if count == 0 {
		fmt.Println("(暂无数据)")
	}
}

func displayStats(ctx context.Context, dbpool *pgxpool.Pool) {
	// 总事件数
	var total int
	dbpool.QueryRow(ctx, "SELECT COUNT(*) FROM calendar_events").Scan(&total)

	// 按货币统计
	rows, _ := dbpool.Query(ctx, `
		SELECT currency, COUNT(*) as cnt
		FROM calendar_events
		GROUP BY currency
		ORDER BY cnt DESC
		LIMIT 8
	`)
	defer rows.Close()

	fmt.Printf("总事件数: %d\n\n", total)
	fmt.Println("按货币分布:")

	for rows.Next() {
		var currency string
		var cnt int
		rows.Scan(&currency, &cnt)
		cnCurr := translateCurrency(currency)
		bar := strings.Repeat("█", cnt/5)
		fmt.Printf("  %-8s %s %d\n", cnCurr, bar, cnt)
	}
}

func translateEvent(en string) string {
	// 尝试完全匹配
	if cn, ok := eventCN[en]; ok {
		return cn
	}

	// 尝试部分匹配
	for enPattern, cn := range eventCN {
		if strings.Contains(en, enPattern) {
			return cn
		}
	}

	return en // 返回英文原名
}

func translateCurrency(code string) string {
	if cn, ok := currencyCN[code]; ok {
		return cn
	}
	return code
}

func translateImpact(impact string) string {
	if cn, ok := impactCN[impact]; ok {
		return cn
	}
	return impact
}
