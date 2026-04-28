package apiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// OnChainClient fetches blockchain data from multiple sources
type OnChainClient struct {
	httpClient *http.Client
	apiKeys    map[string]string
}

// CoinMetricsData represents CoinMetrics data
type CoinMetricsData struct {
	Timestamp    time.Time `json:"timestamp"`
	Asset        string    `json:"asset"`
	Price        float64   `json:"price"`
	Mktcap       float64   `json:"mktcap"`
	SplyCur      float64   `json:"sply_cur"`
	AdrActCnt    float64   `json:"adr_act_cnt"`
	TxCnt        float64   `json:"tx_cnt"`
	TxTfrValMean float64   `json:"tx_tfr_val_mean"`
	TxTfrValMed  float64   `json:"tx_tfr_val_med"`
	HashRate     float64   `json:"hash_rate"`
	DiffMean     float64   `json:"diff_mean"`
	FeeMean      float64   `json:"fee_mean"`
}

// ExchangeFlow represents exchange inflow/outflow
type ExchangeFlow struct {
	Timestamp time.Time `json:"timestamp"`
	Asset     string    `json:"asset"`
	Exchange  string    `json:"exchange"`
	Inflow    float64   `json:"inflow"`
	Outflow   float64   `json:"outflow"`
	Netflow   float64   `json:"netflow"`
}

// NewOnChainClient creates a new on-chain client
func NewOnChainClient(apiKeys map[string]string) *OnChainClient {
	return &OnChainClient{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		apiKeys:    apiKeys,
	}
}

// FetchCoinMetrics fetches CoinMetrics market data
func (c *OnChainClient) FetchCoinMetrics(ctx context.Context, asset string, from, to time.Time) ([]CoinMetricsData, error) {
	apiKey := c.apiKeys["coinmetrics"]
	if apiKey == "" {
		return nil, fmt.Errorf("coinmetrics API key not configured")
	}

	url := fmt.Sprintf("https://api.coinmetrics.io/v4/timeseries/market-candles?assets=%s&start_time=%s&end_time=%s",
		asset, from.Format(time.RFC3339), to.Format(time.RFC3339))

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CoinMetrics API returned status %d", resp.StatusCode)
	}

	var result struct {
		Data []CoinMetricsData `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

// FetchBlockchainInfo fetches blockchain.info stats
func (c *OnChainClient) FetchBlockchainInfo(ctx context.Context) (*BlockchainInfo, error) {
	url := "https://api.blockchain.info/stats"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("blockchain.info API returned status %d", resp.StatusCode)
	}

	var result BlockchainInfo
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// BlockchainInfo represents blockchain.info stats
type BlockchainInfo struct {
	MarketPriceUSD               float64 `json:"market_price_usd"`
	HashRate                     float64 `json:"hash_rate"`
	TotalFeesBTC                 float64 `json:"total_fees_btc"`
	TradeVolumeUSD               float64 `json:"trade_volume_usd"`
	EstimatedTransactionVolume   float64 `json:"estimated_transaction_volume_usd"`
	MinersRevenueUSD             float64 `json:"miners_revenue_usd"`
}

// FetchExchangeFlows fetches exchange inflow/outflow data
func (c *OnChainClient) FetchExchangeFlows(ctx context.Context, asset string, days int) ([]ExchangeFlow, error) {
	// Using Glassnode or similar API
	apiKey := c.apiKeys["glassnode"]
	if apiKey == "" {
		return nil, fmt.Errorf("glassnode API key not configured")
	}

	url := fmt.Sprintf("https://api.glassnode.com/v1/metrics/transactions/transfers_volume_to_exchanges_sum?a=%s&i=1d&c=native", asset)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Api-Key", apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Glassnode API returned status %d", resp.StatusCode)
	}

	var result []struct {
		Timestamp int64   `json:"t"`
		Value     float64 `json:"v"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var flows []ExchangeFlow
	for _, r := range result {
		flows = append(flows, ExchangeFlow{
			Timestamp: time.Unix(r.Timestamp, 0),
			Asset:     asset,
			Exchange:  "aggregate",
			Inflow:    r.Value,
		})
	}

	return flows, nil
}

// FetchActiveAddresses fetches active address count
func (c *OnChainClient) FetchActiveAddresses(ctx context.Context, asset string) (float64, error) {
	// Use CoinMetrics for this
	data, err := c.FetchCoinMetrics(ctx, asset, time.Now().AddDate(0, 0, -1), time.Now())
	if err != nil {
		return 0, err
	}

	if len(data) == 0 {
		return 0, fmt.Errorf("no data available")
	}

	return data[len(data)-1].AdrActCnt, nil
}

// FetchMempoolSize fetches mempool size
func (c *OnChainClient) FetchMempoolSize(ctx context.Context) (int64, error) {
	url := "https://mempool.space/api/mempool"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("mempool.space API returned status %d", resp.StatusCode)
	}

	var result struct {
		Count int64 `json:"count"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	return result.Count, nil
}

// OnChainMetrics aggregates on-chain metrics
type OnChainMetrics struct {
	Timestamp        time.Time `json:"timestamp"`
	Asset            string    `json:"asset"`
	Price            float64   `json:"price"`
	ActiveAddresses  float64   `json:"active_addresses"`
	TxCount24h       float64   `json:"tx_count_24h"`
	HashRate         float64   `json:"hash_rate"`
	MempoolSize      int64     `json:"mempool_size"`
	ExchangeInflow   float64   `json:"exchange_inflow"`
	ExchangeOutflow  float64   `json:"exchange_outflow"`
}

// FetchAllMetrics fetches all on-chain metrics for an asset
func (c *OnChainClient) FetchAllMetrics(ctx context.Context, asset string) (*OnChainMetrics, error) {
	metrics := &OnChainMetrics{
		Timestamp: time.Now(),
		Asset:     asset,
	}

	// Fetch CoinMetrics data
	if data, err := c.FetchCoinMetrics(ctx, asset, time.Now().AddDate(0, 0, -1), time.Now()); err == nil && len(data) > 0 {
		latest := data[len(data)-1]
		metrics.Price = latest.Price
		metrics.ActiveAddresses = latest.AdrActCnt
		metrics.TxCount24h = latest.TxCnt
		metrics.HashRate = latest.HashRate
	}

	// Fetch mempool size for BTC
	if asset == "BTC" {
		if size, err := c.FetchMempoolSize(ctx); err == nil {
			metrics.MempoolSize = size
		}
	}

	// Fetch exchange flows
	if flows, err := c.FetchExchangeFlows(ctx, asset, 1); err == nil && len(flows) > 0 {
		latest := flows[len(flows)-1]
		metrics.ExchangeInflow = latest.Inflow
		metrics.ExchangeOutflow = latest.Outflow
	}

	return metrics, nil
}

// CalculateNVT calculates Network Value to Transactions ratio
func CalculateNVT(marketCap, txVolume float64) float64 {
	if txVolume == 0 {
		return 0
	}
	return marketCap / txVolume
}

// InterpretNVT interprets the NVT ratio
type NVTRegime string

const (
	NVTOvervalued NVTRegime = "OVERVALUED" // NVT > 100
	NVTNeutral    NVTRegime = "NEUTRAL"    // NVT 30-100
	NVTUndervalued NVTRegime = "UNDERVALUED" // NVT < 30
)

func InterpretNVT(nvt float64) NVTRegime {
	if nvt > 100 {
		return NVTOvervalued
	} else if nvt < 30 {
		return NVTUndervalued
	}
	return NVTNeutral
}
