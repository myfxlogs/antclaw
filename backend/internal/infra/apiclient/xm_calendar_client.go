// Package apiclient provides HTTP clients for external data sources.
package apiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// XMEvent represents a single economic calendar event from XM API.
// NOTE: The actual XM API response schema is unverified (XM blocks data-center IPs).
// Adjust these fields once the API is reachable via a residential proxy.
type XMEvent struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Country     string `json:"country"`   // ISO country code, e.g. "US"
	Currency    string `json:"currency"`  // e.g. "USD"
	Impact      string `json:"impact"`    // "low", "medium", "high"
	DateTime    string `json:"datetime"`  // ISO8601
	Actual      string `json:"actual"`
	Forecast    string `json:"forecast"`
	Previous    string `json:"previous"`
}

// XMFetcher fetches economic calendar data from XM's API.
// Requires a residential proxy or non-data-center IP because XM (like MQL5)
// blocks VPS / hosting-provider IP ranges at the TCP level.
type XMFetcher struct {
	httpClient *http.Client
	baseURL    string
	mu         sync.RWMutex
}

const xmDefaultEndpoint = "https://www.xm.com/economic-calendar"
const xmAPIPath = "/api/data" // best-guess; adjust when endpoint is verified

// NewXMFetcher creates a new XM fetcher.
func NewXMFetcher() *XMFetcher {
	return &XMFetcher{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    xmDefaultEndpoint,
	}
}

// SetBaseURL updates the base URL dynamically. Empty string is ignored.
func (f *XMFetcher) SetBaseURL(rawURL string) {
	if rawURL == "" {
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.baseURL = strings.TrimRight(rawURL, "/")
}

// SetHTTPClient replaces the http.Client (e.g. to inject a proxy transport).
func (f *XMFetcher) SetHTTPClient(c *http.Client) {
	if c == nil {
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.httpClient = c
}

func (f *XMFetcher) getBaseURL() string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if f.baseURL == "" {
		return xmDefaultEndpoint
	}
	return f.baseURL
}

func (f *XMFetcher) getHTTPClient() *http.Client {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.httpClient
}

// getWeekRange returns the from/to ISO timestamps for a given week.
func getXMWeekRange(week string) (from, to string) {
	now := time.Now()
	for now.Weekday() != time.Monday {
		now = now.AddDate(0, 0, -1)
	}
	monday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	switch week {
	case "next":
		monday = monday.AddDate(0, 0, 7)
	case "prev":
		monday = monday.AddDate(0, 0, -7)
	}

	sunday := monday.AddDate(0, 0, 6)
	end := time.Date(sunday.Year(), sunday.Month(), sunday.Day(), 23, 59, 59, 0, time.UTC)

	return monday.UTC().Format("2006-01-02T15:04:05"),
		end.UTC().Format("2006-01-02T15:04:05")
}

// FetchWeek fetches calendar events for a specific week.
func (f *XMFetcher) FetchWeek(ctx context.Context, week string) ([]CalendarEvent, error) {
	from, to := getXMWeekRange(week)
	return f.fetch(ctx, from, to)
}

// FetchDateRange fetches calendar events for a date range.
func (f *XMFetcher) FetchDateRange(ctx context.Context, from, to time.Time) ([]CalendarEvent, error) {
	return f.fetch(ctx,
		from.UTC().Format("2006-01-02T15:04:05"),
		to.UTC().Format("2006-01-02T15:04:05"))
}

func (f *XMFetcher) fetch(ctx context.Context, from, to string) ([]CalendarEvent, error) {
	endpoint := f.getBaseURL() + xmAPIPath

	data := url.Values{}
	data.Set("from", from)
	data.Set("to", to)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint,
		strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("xm: failed to build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")

	resp, err := f.getHTTPClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("xm request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("xm returned status %d", resp.StatusCode)
	}

	var rawEvents []XMEvent
	if err := json.NewDecoder(resp.Body).Decode(&rawEvents); err != nil {
		return nil, fmt.Errorf("xm decode failed: %w", err)
	}

	events := make([]CalendarEvent, 0, len(rawEvents))
	for _, e := range rawEvents {
		events = append(events, normalizeXMEvent(e))
	}

	return events, nil
}

func normalizeXMEvent(e XMEvent) CalendarEvent {
	impact := "low"
	switch e.Impact {
	case "high":
		impact = "high"
	case "medium":
		impact = "medium"
	}

	return CalendarEvent{
		EventID:     fmt.Sprintf("xm-%d", e.ID),
		Title:       e.Title,
		Country:     e.Country,
		Currency:    e.Currency,
		Impact:      impact,
		ScheduledAt: e.DateTime,
		Previous:    e.Previous,
		Forecast:    e.Forecast,
		Actual:      e.Actual,
	}
}
