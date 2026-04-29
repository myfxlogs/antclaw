// FRED数据采集演示
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/antclaw/antclaw/internal/infra/apiclient/fred"
	"github.com/antclaw/antclaw/internal/infra/postgres"
	"github.com/antclaw/antclaw/internal/service/macro"
)

func main() {
	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║              📈 FRED宏观数据采集演示                            ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// 连接数据库
	dbpool, err := pgxpool.New(context.Background(), "postgres://antclaw:antclaw@localhost:5432/antclaw")
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer dbpool.Close()
	fmt.Println("✓ 数据库已连接")

	// Create FRED client
	fredKey := os.Getenv("ANTCLAW_FRED_API_KEY")
	fredClient := fred.NewClient(fredKey)

	// 创建Repository和Service
	macroRepo := postgres.NewMacroRepository(dbpool)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	macroSvc := macro.NewMacroService(macroRepo, fredKey, logger)

	ctx := context.Background()

	// 测试获取单个系列
	fmt.Println("\n测试获取FRED数据系列...")
	seriesList := []string{"GDP", "CPIAUCSL", "UNRATE", "FEDFUNDS"}

	for _, seriesID := range seriesList {
		fmt.Printf("\n📊 获取 %s ...\n", seriesID)
		resp, err := fredClient.FetchObservations(ctx, seriesID, 5)
		if err != nil {
			fmt.Printf("  ✗ 获取失败: %v\n", err)
			continue
		}

		fmt.Printf("  ✓ 获取成功: %d 条记录\n", len(resp.Observations))
		for i, obs := range resp.Observations {
			if i < 3 {
				fmt.Printf("    %s: %s\n", obs.Date, obs.Value)
			}
		}
	}

	// 同步到数据库
	fmt.Println("\n同步数据到数据库...")
	result, err := macroSvc.SyncFREDIndicators(ctx, seriesList)
	if err != nil {
		fmt.Printf("✗ 同步失败: %v\n", err)
	} else {
		fmt.Printf("✓ 同步完成: %d 条记录已插入\n", result.Inserted)
	}

	// 查看数据库中的数据
	fmt.Println("\n查看已存储的宏观数据:")
	showMacroData(dbpool)
}

func showMacroData(dbpool *pgxpool.Pool) {
	ctx := context.Background()

	rows, err := dbpool.Query(ctx, `
		SELECT series_id, COUNT(*) as cnt, MAX(time) as latest
		FROM data_snapshots
		GROUP BY series_id
		ORDER BY series_id
	`)
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}
	defer rows.Close()

	fmt.Println("\n数据系列统计:")
	fmt.Println("────────────────────────────────────────")
	for rows.Next() {
		var seriesID string
		var count int
		var latest string
		rows.Scan(&seriesID, &count, &latest)
		latestStr := latest
		if len(latestStr) > 10 {
			latestStr = latestStr[:10]
		}
		fmt.Printf("  %-12s %4d 条记录  最新: %s\n", seriesID, count, latestStr)
	}
}
