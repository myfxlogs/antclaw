package fetcher

import (
	"context"
	"fmt"
	"time"
)

// Fetcher is the unified interface for all data source clients
type Fetcher interface {
	// Name returns a unique source identifier (e.g., "fred", "mql5")
	Name() string

	// Fetch retrieves data for the given key (series_id, event_id, etc.)
	// Returns FetchResult with normalized data points
	Fetch(ctx context.Context, key string) (*FetchResult, error)

	// Healthcheck verifies the source is reachable
	Healthcheck(ctx context.Context) error

	// Available returns true if this fetcher is configured and available
	Available() bool
}

// FetchResult contains fetched data
type FetchResult struct {
	Source     string      `json:"source"`
	Key        string      `json:"key"`
	DataPoints []DataPoint `json:"data_points"`
	RawJSON    []byte      `json:"raw_json"`
	FetchedAt  time.Time   `json:"fetched_at"`
}

// DataPoint represents a single time-series data point
type DataPoint struct {
	Time         time.Time       `json:"time"`
	ValueNumeric *float64        `json:"value_numeric,omitempty"`
	ValueText    *string         `json:"value_text,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// NewDataPoint creates a numeric data point
func NewDataPoint(t time.Time, value float64) DataPoint {
	return DataPoint{
		Time:         t,
		ValueNumeric: &value,
	}
}

// NewTextDataPoint creates a text data point
func NewTextDataPoint(t time.Time, value string) DataPoint {
	return DataPoint{
		Time:      t,
		ValueText: &value,
	}
}

// MultiFetcher combines multiple fetchers with fallback
type MultiFetcher struct {
	fetchers []Fetcher
}

// NewMultiFetcher creates a multi-fetcher with fallback chain
func NewMultiFetcher(fetchers ...Fetcher) *MultiFetcher {
	return &MultiFetcher{fetchers: fetchers}
}

// Fetch tries each fetcher in order until one succeeds
func (m *MultiFetcher) Fetch(ctx context.Context, key string) (*FetchResult, error) {
	var lastErr error
	for _, f := range m.fetchers {
		if !f.Available() {
			continue
		}
		result, err := f.Fetch(ctx, key)
		if err == nil {
			return result, nil
		}
		lastErr = err
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, ErrNoFetcherAvailable
}

// Healthcheck checks all fetchers
func (m *MultiFetcher) Healthcheck(ctx context.Context) map[string]error {
	results := make(map[string]error)
	for _, f := range m.fetchers {
		results[f.Name()] = f.Healthcheck(ctx)
	}
	return results
}

// Available returns true if any fetcher is available
func (m *MultiFetcher) Available() bool {
	for _, f := range m.fetchers {
		if f.Available() {
			return true
		}
	}
	return false
}

// Name returns the multi-fetcher name
func (m *MultiFetcher) Name() string {
	return "multi"
}

// ErrNoFetcherAvailable is returned when no fetcher is available
var ErrNoFetcherAvailable = &FetcherError{Message: "no fetcher available"}

// FetcherError represents a fetcher error
type FetcherError struct {
	Source  string
	Key     string
	Message string
	Cause   error
}

func (e *FetcherError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (key=%s): %v", e.Source, e.Message, e.Key, e.Cause)
	}
	return fmt.Sprintf("%s: %s (key=%s)", e.Source, e.Message, e.Key)
}

func (e *FetcherError) Unwrap() error {
	return e.Cause
}
