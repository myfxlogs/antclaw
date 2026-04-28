// 情绪采集器 - 使用Alternative.me Fear&Greed (免费API)
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// alternativeMeResp Alternative.me Fear&Greed API响应
type alternativeMeResp struct {
	Data []struct {
		Value     string `json:"value"`
		Class     string `json:"value_classification"`
		Timestamp string `json:"timestamp"`
	} `json:"data"`
}

// collectSentiment 采集情绪指标并持久化
func collectSentiment(ctx context.Context, dbpool *pgxpool.Pool, logger *slog.Logger) error {
	logger.Info("Starting sentiment collection: Fear & Greed index")

	// 1. 从Alternative.me获取Crypto Fear & Greed (免费,无需API key)
	cryptoFG, err := fetchAlternativeMeFG(ctx, 30)
	if err != nil {
		logger.Warn("Fear&Greed fetch failed", "error", err)
	}

	inserted := 0
	for _, fg := range cryptoFG {
		regime := classifyFearGreed(fg.Value)
		raw, _ := json.Marshal(fg)

		_, err := dbpool.Exec(ctx, `
			INSERT INTO sentiment_snapshots 
			(time, score, regime, fear_greed, raw)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (time) DO UPDATE SET
			  score = EXCLUDED.score,
			  regime = EXCLUDED.regime,
			  fear_greed = EXCLUDED.fear_greed,
			  raw = EXCLUDED.raw`,
			fg.Time, fg.Value, regime, fg.Value, raw)
		if err == nil {
			inserted++
		}
	}

	logger.Info("Sentiment collection completed", "fear_greed_records", inserted)
	return nil
}

type fgRecord struct {
	Time  time.Time
	Value float64
	Class string
}

// fetchAlternativeMeFG 从Alternative.me获取Fear & Greed历史
func fetchAlternativeMeFG(ctx context.Context, limit int) ([]fgRecord, error) {
	url := fmt.Sprintf("https://api.alternative.me/fng/?limit=%d", limit)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("alternative.me status %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var data alternativeMeResp
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	var records []fgRecord
	for _, d := range data.Data {
		val, _ := strconv.ParseFloat(d.Value, 64)
		ts, _ := strconv.ParseInt(d.Timestamp, 10, 64)
		records = append(records, fgRecord{
			Time:  time.Unix(ts, 0),
			Value: val,
			Class: d.Class,
		})
	}
	return records, nil
}

// classifyFearGreed 将Fear&Greed数值分类为市场状态
func classifyFearGreed(value float64) string {
	switch {
	case value <= 25:
		return "extreme_fear"
	case value <= 45:
		return "fear"
	case value <= 55:
		return "neutral"
	case value <= 75:
		return "greed"
	default:
		return "extreme_greed"
	}
}
