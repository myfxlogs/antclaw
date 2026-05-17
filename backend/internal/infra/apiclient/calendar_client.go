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

// MQL5Event represents a single economic calendar event from MQL5 API.
type MQL5Event struct {
	ID               int    `json:"Id"`
	EventName        string `json:"EventName"`
	Importance       string `json:"Importance"`       // "low", "medium", "high"
	CurrencyCode     string `json:"CurrencyCode"`     // e.g. "USD"
	Country          int    `json:"Country"`
	ActualValue      string `json:"ActualValue"`
	ForecastValue    string `json:"ForecastValue"`
	PreviousValue    string `json:"PreviousValue"`
	OldPreviousValue string `json:"OldPreviousValue"`
	FullDate         string `json:"FullDate"`         // "2026-03-17T14:00:00"
	ReleaseDate      int64  `json:"ReleaseDate"`      // Unix milliseconds
	ImpactDirection  int    `json:"ImpactDirection"`  // 0=neutral, 1=positive, 2=negative
	ImpactValue      string `json:"ImpactValue"`
	Processed        int    `json:"Processed"`        // 1=released, 0=upcoming
}

// CalendarEvent is the normalized calendar event structure.
type CalendarEvent struct {
	EventID     string `json:"event_id"`
	Title       string `json:"title"`
	Country     string `json:"country"`
	Currency    string `json:"currency"`
	Impact      string `json:"impact"`       // "low", "medium", "high"
	ScheduledAt string `json:"scheduled_at"` // ISO8601 (FullDate)
	ReleaseDate int64  `json:"release_date"`  // Unix milliseconds (ReleaseDate)
	Previous    string `json:"previous"`
	Forecast    string `json:"forecast"`
	Actual      string `json:"actual"`
}

// MQL5Fetcher fetches economic calendar data from MQL5's hidden POST endpoint.
// No API key required. Returns all impact levels (low, medium, high, holiday).
type MQL5Fetcher struct {
	httpClient *http.Client
	baseURL    string
	mu         sync.RWMutex
}

const mql5DefaultEndpoint = "https://www.mql5.com/en/economic-calendar/content"
const mql5ImportanceAll = "15"      // bitmask: low(1) + medium(2) + high(4) + holiday(8)
const mql5CurrenciesMask = "131071" // all major currencies

// NewMQL5Fetcher creates a new MQL5 fetcher.
func NewMQL5Fetcher() *MQL5Fetcher {
	return &MQL5Fetcher{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    mql5DefaultEndpoint,
	}
}

// SetBaseURL updates the base URL dynamically. Empty string is ignored.
func (f *MQL5Fetcher) SetBaseURL(url string) {
	if url == "" {
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.baseURL = strings.TrimRight(url, "/")
}

// SetHTTPClient replaces the http.Client (e.g. to inject a proxy transport).
// Useful when the data center IP is blocked by the upstream (as with MQL5).
func (f *MQL5Fetcher) SetHTTPClient(c *http.Client) {
	if c == nil {
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.httpClient = c
}

func (f *MQL5Fetcher) getBaseURL() string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if f.baseURL == "" {
		return mql5DefaultEndpoint
	}
	return f.baseURL
}

// getWeekRange returns the from/to ISO timestamps for a given week.
func getWeekRange(week string) (from, to string) {
	now := time.Now()
	// Walk back to Monday
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
func (f *MQL5Fetcher) FetchWeek(ctx context.Context, week string) ([]CalendarEvent, error) {
	from, to := getWeekRange(week)
	return f.fetch(ctx, from, to)
}

// FetchDateRange fetches calendar events for a date range.
func (f *MQL5Fetcher) FetchDateRange(ctx context.Context, from, to time.Time) ([]CalendarEvent, error) {
	return f.fetch(ctx,
		from.UTC().Format("2006-01-02T15:04:05"),
		to.UTC().Format("2006-01-02T15:04:05"))
}

func (f *MQL5Fetcher) fetch(ctx context.Context, from, to string) ([]CalendarEvent, error) {
	data := url.Values{}
	data.Set("date_mode", "2") // 2 = date range mode
	data.Set("from", from)
	data.Set("to", to)
	data.Set("importance", mql5ImportanceAll)
	data.Set("currencies", mql5CurrenciesMask)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, f.getBaseURL(),
		strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://www.mql5.com/en/economic-calendar")
	req.Header.Set("Origin", "https://www.mql5.com")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("mql5 request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mql5 returned status %d", resp.StatusCode)
	}

	var rawEvents []MQL5Event
	if err := json.NewDecoder(resp.Body).Decode(&rawEvents); err != nil {
		return nil, fmt.Errorf("mql5 decode failed: %w", err)
	}

	// Normalize to CalendarEvent
	events := make([]CalendarEvent, 0, len(rawEvents))
	for _, e := range rawEvents {
		events = append(events, normalizeEvent(e))
	}

	return events, nil
}

// normalizeEvent converts MQL5 event to normalized CalendarEvent.
func normalizeEvent(e MQL5Event) CalendarEvent {
	// Map importance
	impact := "low"
	switch e.Importance {
	case "high":
		impact = "high"
	case "medium":
		impact = "medium"
	}

	return CalendarEvent{
		EventID:     fmt.Sprintf("mql5-%d", e.ID),
		Title:       e.EventName,
		Country:     getCountryName(e.Country),
		Currency:    e.CurrencyCode,
		Impact:      impact,
		ScheduledAt: e.FullDate,
		ReleaseDate: e.ReleaseDate,
		Previous:    e.PreviousValue,
		Forecast:    e.ForecastValue,
		Actual:      e.ActualValue,
	}
}

// getCountryName maps MQL5 country codes to country names.
func getCountryName(code int) string {
	countryMap := map[int]string{
		1:   "USA",
		2:   "UK",
		3:   "Germany",
		4:   "France",
		5:   "Italy",
		6:   "Japan",
		7:   "Canada",
		8:   "Australia",
		9:   "Switzerland",
		10:  "China",
		11:  "New Zealand",
		12:  "Spain",
		13:  "Sweden",
		14:  "Norway",
		15:  "Denmark",
		16:  "Netherlands",
		17:  "Belgium",
		18:  "Austria",
		19:  "Ireland",
		20:  "Portugal",
		21:  "Greece",
		22:  "Finland",
		23:  "Russia",
		24:  "Brazil",
		25:  "India",
		26:  "Mexico",
		27:  "South Africa",
		28:  "Turkey",
		29:  "Poland",
		30:  "South Korea",
		31:  "Singapore",
		32:  "Indonesia",
		33:  "Hong Kong",
		34:  "Malaysia",
		35:  "Thailand",
		36:  "Philippines",
		37:  "Vietnam",
		38:  "Chile",
		39:  "Colombia",
		40:  "Peru",
		41:  "Argentina",
		42:  "Egypt",
		43:  "Saudi Arabia",
		44:  "UAE",
		45:  "Israel",
		46:  "Czech Republic",
		47:  "Hungary",
		48:  "Romania",
		49:  "Bulgaria",
		50:  "Croatia",
		51:  "EU",
		52:  "OPEC",
		53:  "IMF",
		54:  "World Bank",
		55:  "WTO",
		56:  "OECD",
		57:  "G7",
		58:  "G20",
		59:  "ECB",
		60:  "Fed",
		61:  "BOE",
		62:  "BOJ",
		63:  "SNB",
		64:  "RBA",
		65:  "RBNZ",
		66:  "BOC",
		67:  "PBOC",
		68:  "RBI",
		69:  "CBRT",
		70:  "BRSA",
		71:  "BCRA",
		72:  "CBR",
		73:  "BCB",
		74:  "CBM",
		75:  "CBRP",
	}
	if name, ok := countryMap[code]; ok {
		return name
	}
	return fmt.Sprintf("Country_%d", code)
}
