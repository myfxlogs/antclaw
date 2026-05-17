package apiclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// --- MQL5 Fetcher Tests ---

func TestMQL5Fetcher_FetchWeek_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify method and headers
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Errorf("unexpected Content-Type: %s", r.Header.Get("Content-Type"))
		}

		// Parse form to verify params
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}

		// Return mock response
		events := []MQL5Event{
			{
				ID:            101,
				EventName:     "Non-Farm Payrolls",
				Importance:    "high",
				CurrencyCode:  "USD",
				Country:       1,
				ActualValue:   "250K",
				ForecastValue: "200K",
				PreviousValue: "180K",
				FullDate:      "2026-05-17T12:30:00",
				ReleaseDate:   1715950800000,
			},
			{
				ID:            102,
				EventName:     "CPI m/m",
				Importance:    "medium",
				CurrencyCode:  "USD",
				Country:       1,
				ActualValue:   "0.3%",
				ForecastValue: "0.4%",
				PreviousValue: "0.4%",
				FullDate:      "2026-05-18T12:30:00",
				ReleaseDate:   1716037200000,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(events)
	}))
	defer ts.Close()

	f := NewMQL5Fetcher()
	f.SetBaseURL(ts.URL)

	events, err := f.FetchWeek(context.Background(), "current")
	if err != nil {
		t.Fatalf("FetchWeek failed: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	// Verify first event
	e := events[0]
	if e.EventID != "mql5-101" {
		t.Errorf("expected event_id mql5-101, got %s", e.EventID)
	}
	if e.Title != "Non-Farm Payrolls" {
		t.Errorf("expected title Non-Farm Payrolls, got %s", e.Title)
	}
	if e.Currency != "USD" {
		t.Errorf("expected currency USD, got %s", e.Currency)
	}
	if e.Impact != "high" {
		t.Errorf("expected impact high, got %s", e.Impact)
	}
	if e.Country != "USA" {
		t.Errorf("expected country USA, got %s", e.Country)
	}
	if e.Actual != "250K" {
		t.Errorf("expected actual 250K, got %s", e.Actual)
	}
	if e.Forecast != "200K" {
		t.Errorf("expected forecast 200K, got %s", e.Forecast)
	}
	if e.Previous != "180K" {
		t.Errorf("expected previous 180K, got %s", e.Previous)
	}
	if e.ScheduledAt != "2026-05-17T12:30:00" {
		t.Errorf("expected scheduled_at 2026-05-17T12:30:00, got %s", e.ScheduledAt)
	}
}

func TestMQL5Fetcher_FetchWeek_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	f := NewMQL5Fetcher()
	f.SetBaseURL(ts.URL)

	_, err := f.FetchWeek(context.Background(), "current")
	if err == nil {
		t.Fatal("expected error for 500 status, got nil")
	}
}

func TestMQL5Fetcher_FetchWeek_InvalidJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not json"))
	}))
	defer ts.Close()

	f := NewMQL5Fetcher()
	f.SetBaseURL(ts.URL)

	_, err := f.FetchWeek(context.Background(), "current")
	if err == nil {
		t.Fatal("expected decode error, got nil")
	}
}

func TestMQL5Fetcher_FetchWeek_Empty(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
	}))
	defer ts.Close()

	f := NewMQL5Fetcher()
	f.SetBaseURL(ts.URL)

	events, err := f.FetchWeek(context.Background(), "current")
	if err != nil {
		t.Fatalf("FetchWeek failed: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected 0 events, got %d", len(events))
	}
}

func TestMQL5Fetcher_FetchWeek_ContextCanceled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
	}))
	defer ts.Close()

	f := NewMQL5Fetcher()
	f.SetBaseURL(ts.URL)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := f.FetchWeek(ctx, "current")
	if err == nil {
		t.Fatal("expected context canceled error, got nil")
	}
}

func TestMQL5Fetcher_FetchDateRange(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if from := r.FormValue("from"); from == "" {
			t.Error("expected from param, got empty")
		}
		if to := r.FormValue("to"); to == "" {
			t.Error("expected to param, got empty")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
	}))
	defer ts.Close()

	f := NewMQL5Fetcher()
	f.SetBaseURL(ts.URL)

	from := time.Date(2026, 5, 11, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 5, 17, 23, 59, 59, 0, time.UTC)
	events, err := f.FetchDateRange(context.Background(), from, to)
	if err != nil {
		t.Fatalf("FetchDateRange failed: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected 0 events, got %d", len(events))
	}
}

func TestMQL5Fetcher_SetBaseURL_EmptyIgnored(t *testing.T) {
	f := NewMQL5Fetcher()
	original := f.getBaseURL()
	f.SetBaseURL("")
	if f.getBaseURL() != original {
		t.Errorf("empty SetBaseURL should not change baseURL")
	}
}

func TestMQL5Fetcher_SetBaseURL_Updates(t *testing.T) {
	f := NewMQL5Fetcher()
	f.SetBaseURL("https://example.com")
	if f.getBaseURL() != "https://example.com" {
		t.Errorf("expected https://example.com, got %s", f.getBaseURL())
	}
}

func TestMQL5Fetcher_normalizeEvent_ImpactMapping(t *testing.T) {
	tests := []struct {
		name     string
		input    MQL5Event
		expected string
	}{
		{"high importance", MQL5Event{ID: 1, Importance: "high"}, "high"},
		{"medium importance", MQL5Event{ID: 2, Importance: "medium"}, "medium"},
		{"low importance", MQL5Event{ID: 3, Importance: "low"}, "low"},
		{"unknown maps to low", MQL5Event{ID: 4, Importance: "unknown"}, "low"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeEvent(tt.input)
			if result.Impact != tt.expected {
				t.Errorf("expected impact %s, got %s", tt.expected, result.Impact)
			}
		})
	}
}

// --- XM Fetcher Tests ---

func TestXMFetcher_FetchWeek_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Errorf("unexpected Content-Type: %s", r.Header.Get("Content-Type"))
		}

		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}

		events := []XMEvent{
			{
				ID:       201,
				Title:    "GDP q/q",
				Country:  "US",
				Currency: "USD",
				Impact:   "high",
				DateTime: "2026-05-17T12:30:00Z",
				Actual:   "3.2%",
				Forecast: "3.0%",
				Previous: "3.4%",
			},
			{
				ID:       202,
				Title:    "Unemployment Rate",
				Country:  "UK",
				Currency: "GBP",
				Impact:   "medium",
				DateTime: "2026-05-18T06:00:00Z",
				Actual:   "",
				Forecast: "4.0%",
				Previous: "4.2%",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(events)
	}))
	defer ts.Close()

	f := NewXMFetcher()
	f.SetBaseURL(ts.URL)

	events, err := f.FetchWeek(context.Background(), "current")
	if err != nil {
		t.Fatalf("FetchWeek failed: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	// Verify first event
	e := events[0]
	if e.EventID != "xm-201" {
		t.Errorf("expected event_id xm-201, got %s", e.EventID)
	}
	if e.Title != "GDP q/q" {
		t.Errorf("expected title GDP q/q, got %s", e.Title)
	}
	if e.Country != "US" {
		t.Errorf("expected country US, got %s", e.Country)
	}
	if e.Currency != "USD" {
		t.Errorf("expected currency USD, got %s", e.Currency)
	}
	if e.Impact != "high" {
		t.Errorf("expected impact high, got %s", e.Impact)
	}
	if e.Actual != "3.2%" {
		t.Errorf("expected actual 3.2%%, got %s", e.Actual)
	}
	if e.Forecast != "3.0%" {
		t.Errorf("expected forecast 3.0%%, got %s", e.Forecast)
	}
	if e.Previous != "3.4%" {
		t.Errorf("expected previous 3.4%%, got %s", e.Previous)
	}

	// Verify second event (no actual value yet - upcoming)
	e2 := events[1]
	if e2.Actual != "" {
		t.Errorf("expected empty actual for upcoming event, got %s", e2.Actual)
	}
	if e2.EventID != "xm-202" {
		t.Errorf("expected event_id xm-202, got %s", e2.EventID)
	}
}

func TestXMFetcher_FetchWeek_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer ts.Close()

	f := NewXMFetcher()
	f.SetBaseURL(ts.URL)

	_, err := f.FetchWeek(context.Background(), "current")
	if err == nil {
		t.Fatal("expected error for 503 status, got nil")
	}
}

func TestXMFetcher_FetchWeek_InvalidJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{invalid}"))
	}))
	defer ts.Close()

	f := NewXMFetcher()
	f.SetBaseURL(ts.URL)

	_, err := f.FetchWeek(context.Background(), "current")
	if err == nil {
		t.Fatal("expected decode error, got nil")
	}
}

func TestXMFetcher_FetchWeek_EmptyResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
	}))
	defer ts.Close()

	f := NewXMFetcher()
	f.SetBaseURL(ts.URL)

	events, err := f.FetchWeek(context.Background(), "current")
	if err != nil {
		t.Fatalf("FetchWeek failed: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected 0 events, got %d", len(events))
	}
}

func TestXMFetcher_SetHTTPClient(t *testing.T) {
	f := NewXMFetcher()
	custom := &http.Client{Timeout: 5 * time.Second}
	f.SetHTTPClient(custom)
	if f.getHTTPClient() != custom {
		t.Error("SetHTTPClient did not set the client")
	}
}

func TestXMFetcher_SetHTTPClient_NilIgnored(t *testing.T) {
	f := NewXMFetcher()
	original := f.getHTTPClient()
	f.SetHTTPClient(nil)
	if f.getHTTPClient() != original {
		t.Error("nil SetHTTPClient should not change client")
	}
}

func TestXMFetcher_SetBaseURL_EmptyIgnored(t *testing.T) {
	f := NewXMFetcher()
	original := f.getBaseURL()
	f.SetBaseURL("")
	if f.getBaseURL() != original {
		t.Errorf("empty SetBaseURL should not change baseURL")
	}
}

func TestXMFetcher_normalizeXMEvent_ImpactMapping(t *testing.T) {
	tests := []struct {
		name     string
		input    XMEvent
		expected string
	}{
		{"high impact", XMEvent{ID: 1, Impact: "high"}, "high"},
		{"medium impact", XMEvent{ID: 2, Impact: "medium"}, "medium"},
		{"low impact", XMEvent{ID: 3, Impact: "low"}, "low"},
		{"unknown defaults to low", XMEvent{ID: 4, Impact: "unknown"}, "low"},
		{"empty defaults to low", XMEvent{ID: 5, Impact: ""}, "low"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeXMEvent(tt.input)
			if result.Impact != tt.expected {
				t.Errorf("expected impact %s, got %s", tt.expected, result.Impact)
			}
		})
	}
}

func TestXMFetcher_FetchDateRange(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if from := r.FormValue("from"); from == "" {
			t.Error("expected from param, got empty")
		}
		if to := r.FormValue("to"); to == "" {
			t.Error("expected to param, got empty")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
	}))
	defer ts.Close()

	f := NewXMFetcher()
	f.SetBaseURL(ts.URL)

	from := time.Date(2026, 5, 11, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 5, 17, 23, 59, 59, 0, time.UTC)
	events, err := f.FetchDateRange(context.Background(), from, to)
	if err != nil {
		t.Fatalf("FetchDateRange failed: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected 0 events, got %d", len(events))
	}
}

func TestXMFetcher_ContextCanceled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
	}))
	defer ts.Close()

	f := NewXMFetcher()
	f.SetBaseURL(ts.URL)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := f.FetchWeek(ctx, "current")
	if err == nil {
		t.Fatal("expected context canceled error, got nil")
	}
}

// --- getWeekRange / getXMWeekRange tests ---

func TestGetWeekRange_Current(t *testing.T) {
	from, to := getWeekRange("current")
	if from == "" || to == "" {
		t.Fatal("expected non-empty from/to for current week")
	}

	// from should be a Monday 00:00:00
	fromTime, err := time.Parse("2006-01-02T15:04:05", from)
	if err != nil {
		t.Fatalf("failed to parse from: %v", err)
	}
	if fromTime.Weekday() != time.Monday {
		t.Errorf("expected Monday, got %s", fromTime.Weekday())
	}
	if fromTime.Hour() != 0 || fromTime.Minute() != 0 || fromTime.Second() != 0 {
		t.Errorf("expected 00:00:00, got %02d:%02d:%02d", fromTime.Hour(), fromTime.Minute(), fromTime.Second())
	}

	// to should be Sunday 23:59:59
	toTime, err := time.Parse("2006-01-02T15:04:05", to)
	if err != nil {
		t.Fatalf("failed to parse to: %v", err)
	}
	if toTime.Weekday() != time.Sunday {
		t.Errorf("expected Sunday, got %s", toTime.Weekday())
	}
}

func TestGetWeekRange_Next(t *testing.T) {
	curFrom, _ := getWeekRange("current")
	nextFrom, _ := getWeekRange("next")

	curTime, _ := time.Parse("2006-01-02T15:04:05", curFrom)
	nextTime, _ := time.Parse("2006-01-02T15:04:05", nextFrom)

	diff := nextTime.Sub(curTime)
	if diff != 7*24*time.Hour {
		t.Errorf("expected 7 days between weeks, got %s", diff)
	}
}

func TestGetWeekRange_Prev(t *testing.T) {
	curFrom, _ := getWeekRange("current")
	prevFrom, _ := getWeekRange("prev")

	curTime, _ := time.Parse("2006-01-02T15:04:05", curFrom)
	prevTime, _ := time.Parse("2006-01-02T15:04:05", prevFrom)

	diff := curTime.Sub(prevTime)
	if diff != 7*24*time.Hour {
		t.Errorf("expected 7 days between weeks, got %s", diff)
	}
}

func TestGetWeekRange_Concurrent(t *testing.T) {
	// Verify thread-safety of the MQL5 and XM fetchers under concurrent reads
	m5 := NewMQL5Fetcher()
	xm := NewXMFetcher()

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			_ = m5.getBaseURL()
			_ = xm.getBaseURL()
			done <- true
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
}
