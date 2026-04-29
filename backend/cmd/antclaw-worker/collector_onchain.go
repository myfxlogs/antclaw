// 链上数据采集器：通过 CoinGecko 公共 API 抓取价格/市值/成交量。
// 表 onchain_metrics 为长表 (time, asset, metric, value, source)，单点拆为多行。
// MVRV/SOPR/active_addresses 等深度链上指标不在 CoinGecko 范围，留待 Coinmetrics 薄客户端接入。
package main

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/antclaw/antclaw/internal/infra/apiclient"
	cg "github.com/antclaw/antclaw/internal/infra/apiclient/coingecko"
	"github.com/jackc/pgx/v5/pgxpool"
)

// onchainAssets 要采集的加密资产（CoinGecko ID + 内部 Symbol）。
var onchainAssets = []struct {
	ID     string
	Symbol string
}{
	{"bitcoin", "BTC"},
	{"ethereum", "ETH"},
	{"solana", "SOL"},
	{"binancecoin", "BNB"},
	{"ripple", "XRP"},
}

const onchainSource = "coingecko"

// collectOnchain 拉取最近 30 日数据并写入 onchain_metrics。
func collectOnchain(ctx context.Context, dbpool *pgxpool.Pool, logger *slog.Logger) error {
	logger.Info("Starting onchain collection from CoinGecko")

	src := apiclient.NewSource(onchainSource, apiclient.Options{Timeout: 30 * time.Second})
	client := cg.NewClient(src)
	totalInserted := 0
	var firstErr error

	for _, asset := range onchainAssets {
		data, err := client.GetMarketChart(ctx, asset.ID, "usd", 30, "daily")
		if err != nil {
			logger.Warn("CoinGecko fetch failed", "asset", asset.Symbol, "error", err)
			if firstErr == nil {
				firstErr = err
			}
			continue
		}

		inserted, err := saveOnchainMetrics(ctx, dbpool, asset.Symbol, data, logger)
		if err != nil {
			logger.Warn("onchain insert failed", "asset", asset.Symbol, "error", err)
			if firstErr == nil {
				firstErr = err
			}
		}
		totalInserted += inserted
		logger.Info("Onchain synced", "asset", asset.Symbol, "records", inserted)

		// CoinGecko free tier ≈ 30 req/min
		time.Sleep(2 * time.Second)
	}

	logger.Info("Onchain collection completed", "total_inserted", totalInserted)
	if totalInserted == 0 && firstErr != nil {
		return firstErr
	}
	return nil
}

// saveOnchainMetrics 把市场快照拆成长表行：每个 (time,asset) 写 price / market_cap / total_volume 三条。
func saveOnchainMetrics(ctx context.Context, dbpool *pgxpool.Pool, symbol string, data *cg.MarketResp, logger *slog.Logger) (int, error) {
	if data == nil || len(data.Prices) == 0 {
		return 0, errors.New("empty market chart response")
	}
	const sql = `
		INSERT INTO onchain_metrics (time, asset, metric, value, source)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (time, asset, metric) DO UPDATE SET
		  value = EXCLUDED.value,
		  source = EXCLUDED.source`

	count := 0
	n := len(data.Prices)
	if len(data.MarketCaps) < n {
		n = len(data.MarketCaps)
	}
	if len(data.TotalVolumes) < n {
		n = len(data.TotalVolumes)
	}
	for i := 0; i < n; i++ {
		tsMs := int64(data.Prices[i][0])
		ts := time.UnixMilli(tsMs).UTC()
		points := []struct {
			metric string
			value  float64
		}{
			{"price_usd", data.Prices[i][1]},
			{"market_cap_usd", data.MarketCaps[i][1]},
			{"total_volume_usd", data.TotalVolumes[i][1]},
		}
		for _, p := range points {
			if _, err := dbpool.Exec(ctx, sql, ts, symbol, p.metric, p.value, onchainSource); err != nil {
				logger.Warn("onchain row insert failed",
					"asset", symbol, "metric", p.metric, "time", ts, "error", err)
				continue
			}
			count++
		}
	}
	return count, nil
}
