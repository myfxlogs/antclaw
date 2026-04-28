// 链上数据采集器 - 使用CoinGecko免费API
package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/antclaw/antclaw/internal/infra/apiclient"
	cg "github.com/antclaw/antclaw/internal/infra/apiclient/coingecko"
	"github.com/jackc/pgx/v5/pgxpool"
)

// onchainAssets 要采集的加密资产列表
var onchainAssets = []struct {
	ID     string // CoinGecko ID
	Symbol string // 内部符号
}{
	{"bitcoin", "BTC"},
	{"ethereum", "ETH"},
	{"solana", "SOL"},
	{"binancecoin", "BNB"},
	{"ripple", "XRP"},
}

// cg.MarketResp CoinGecko市场数据响应

// collectOnchain 采集链上数据并持久化
func collectOnchain(ctx context.Context, dbpool *pgxpool.Pool, logger *slog.Logger) error {
	logger.Info("Starting onchain collection from CoinGecko")

	src := apiclient.NewSource("coingecko", apiclient.Options{Timeout: 30 * time.Second})
	client := cg.NewClient(src)
	totalInserted := 0

	for _, asset := range onchainAssets {
		data, err := client.GetMarketChart(ctx, asset.ID, "usd", 30, "daily")
		if err != nil {
			logger.Warn("CoinGecko fetch failed", "asset", asset.Symbol, "error", err)
			continue
		}

		inserted := saveOnchainMetrics(ctx, dbpool, asset.Symbol, data)
		totalInserted += inserted
		logger.Info("Onchain synced", "asset", asset.Symbol, "records", inserted)

		// API限速: CoinGecko免费版约30 calls/min
		time.Sleep(2 * time.Second)
	}

	logger.Info("Onchain collection completed", "total_inserted", totalInserted)
	return nil
}

// saveOnchainMetrics 保存链上指标到数据库
func saveOnchainMetrics(ctx context.Context, dbpool *pgxpool.Pool, symbol string, data *cg.MarketResp) int {
	count := 0
	// CoinGecko返回的是 [timestamp_ms, value] 数组
	for i := 0; i < len(data.Prices); i++ {
		if i >= len(data.TotalVolumes) || i >= len(data.MarketCaps) {
			break
		}
		tsMs := int64(data.Prices[i][0])
		date := time.Unix(tsMs/1000, 0).UTC().Truncate(24 * time.Hour)
		volume := data.TotalVolumes[i][1]
		marketCap := data.MarketCaps[i][1]

		// 使用交易量作为flow代理,市值作为onchain_score基础
		_, err := dbpool.Exec(ctx, `
			INSERT INTO onchain_metrics 
			(date, asset, flow_in, flow_out, net_flow, onchain_score)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (date, asset) DO UPDATE SET
			  flow_in = EXCLUDED.flow_in,
			  flow_out = EXCLUDED.flow_out,
			  net_flow = EXCLUDED.net_flow,
			  onchain_score = EXCLUDED.onchain_score`,
			date, symbol, volume*0.5, volume*0.5, 0.0, marketCap)
		if err == nil {
			count++
		}
	}
	return count
}
