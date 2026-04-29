// Package apiclient provides HTTP clients for external data sources.
package apiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const fredDefaultBaseURL = "https://api.stlouisfed.org/fred"

// FredClient handles FRED API requests.
type FredClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	mu         sync.RWMutex
}

// FredObservation represents a single FRED data point.
type FredObservation struct {
	RealtimeStart string `json:"realtime_start"`
	RealtimeEnd   string `json:"realtime_end"`
	Date          string `json:"date"`
	Value         string `json:"value"`
}

// FredResponse is the API response structure.
type FredResponse struct {
	RealtimeStart    string            `json:"realtime_start"`
	RealtimeEnd      string            `json:"realtime_end"`
	ObservationStart string            `json:"observation_start"`
	ObservationEnd   string            `json:"observation_end"`
	Units            string            `json:"units"`
	OutputType       int               `json:"output_type"`
	FileType         string            `json:"file_type"`
	OrderBy          string            `json:"order_by"`
	SortOrder        string            `json:"sort_order"`
	Count            int               `json:"count"`
	Offset           int               `json:"offset"`
	Limit            int               `json:"limit"`
	Observations     []FredObservation `json:"observations"`
}

// NewFredClient creates a new FRED API client.
func NewFredClient(apiKey string) *FredClient {
	return &FredClient{
		apiKey:     apiKey,
		baseURL:    fredDefaultBaseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// SetBaseURL updates the base URL dynamically. Empty string is ignored.
//
// 兜底归一化：用户在 /datasources 表单里很容易只填 host（例如
// "https://api.stlouisfed.org"），漏掉 FRED API 的 "/fred" 子路径。
// 为了避免在线上配置错把 BaseURL 覆盖成不可用的值，这里统一做：
//   - 去掉尾部斜杠
//   - 若没有以 "/fred" 结尾（也不是已包含 v2 路径），补 "/fred"
func (c *FredClient) SetBaseURL(url string) {
	if url == "" {
		return
	}
	url = strings.TrimRight(url, "/")
	if !strings.HasSuffix(url, "/fred") && !strings.Contains(url, "/fred/") {
		url = url + "/fred"
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.baseURL = url
}

func (c *FredClient) getBaseURL() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.baseURL == "" {
		return fredDefaultBaseURL
	}
	return c.baseURL
}

// SetAPIKey updates the API key dynamically. Empty string is ignored.
func (c *FredClient) SetAPIKey(key string) {
	if key == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.apiKey = key
}

func (c *FredClient) getAPIKey() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.apiKey
}

// IsConfigured returns true if API key is set.
func (c *FredClient) IsConfigured() bool {
	return c.apiKey != ""
}

// FetchObservations retrieves FRED series observations.
func (c *FredClient) FetchObservations(ctx context.Context, seriesID string, limit int) (*FredResponse, error) {
	if c.getAPIKey() == "" {
		return nil, fmt.Errorf("FRED API key not configured")
	}

	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	u, _ := url.Parse(c.getBaseURL() + "/series/observations")
	q := u.Query()
	q.Set("series_id", seriesID)
	q.Set("api_key", c.getAPIKey())
	q.Set("file_type", "json")
	q.Set("sort_order", "desc")
	q.Set("limit", fmt.Sprintf("%d", limit))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fred api request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fred api returned status %d", resp.StatusCode)
	}

	var result FredResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("fred api decode failed: %w", err)
	}

	return &result, nil
}

// FetchSeriesInfo retrieves metadata for a FRED series.
func (c *FredClient) FetchSeriesInfo(ctx context.Context, seriesID string) (map[string]interface{}, error) {
	if c.getAPIKey() == "" {
		return nil, fmt.Errorf("FRED API key not configured")
	}

	u, _ := url.Parse(c.getBaseURL() + "/series")
	q := u.Query()
	q.Set("series_id", seriesID)
	q.Set("api_key", c.getAPIKey())
	q.Set("file_type", "json")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fred api returned status %d", resp.StatusCode)
	}

	var result struct {
		Seriess []map[string]interface{} `json:"seriess"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Seriess) > 0 {
		return result.Seriess[0], nil
	}
	return nil, fmt.Errorf("series not found: %s", seriesID)
}

// ParseValues extracts float64 values from FRED observations.
// Returns values in descending chronological order (values[0] = most recent).
// Skips missing values (represented as "." in FRED).
func ParseValues(resp *FredResponse) []float64 {
	var values []float64
	for _, obs := range resp.Observations {
		if obs.Value == "." || obs.Value == "" {
			continue
		}
		v, err := strconv.ParseFloat(obs.Value, 64)
		if err != nil {
			continue
		}
		values = append(values, v)
	}
	return values
}

// ParseLatest returns the most recent non-missing value from FRED response.
func ParseLatest(resp *FredResponse) (float64, string, bool) {
	for _, obs := range resp.Observations {
		if obs.Value == "." || obs.Value == "" {
			continue
		}
		v, err := strconv.ParseFloat(obs.Value, 64)
		if err != nil {
			continue
		}
		return v, obs.Date, true
	}
	return 0, "", false
}
