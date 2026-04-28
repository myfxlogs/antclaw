package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CalendarEvent represents an economic calendar event
type CalendarEvent struct {
	EventID       string    `db:"event_id"`
	Title         string    `db:"title"`
	Country       string    `db:"country"`
	Currency      string    `db:"currency"`
	Impact        string    `db:"impact"`
	ScheduledAt   time.Time `db:"scheduled_at"`
	PreviousValue string    `db:"previous_value"`
	ForecastValue string    `db:"forecast_value"`
	ActualValue   string    `db:"actual_value"`
	ImpactDirection int16   `db:"impact_direction"`
	SurpriseScore float64   `db:"surprise_score"`
	SurpriseLabel string    `db:"surprise_label"`
	RevisionLabel string    `db:"revision_label"`
	FetchedAt     time.Time `db:"fetched_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

// EventImpactRecord represents price impact of an event
type EventImpactRecord struct {
	EventID      string    `db:"event_id"`
	Window       string    `db:"window"`
	Symbol       string    `db:"symbol"`
	PriceBefore  float64   `db:"price_before"`
	PriceAfter   float64   `db:"price_after"`
	PctChange    float64   `db:"pct_change"`
	RecordedAt   time.Time `db:"recorded_at"`
}

// CalendarRepository defines calendar data operations
type CalendarRepository interface {
	UpsertEvents(ctx context.Context, events []CalendarEvent) (int, error)
	UpdateActual(ctx context.Context, eventID, actual string, surprise float64) error
	GetByDate(ctx context.Context, date time.Time) ([]CalendarEvent, error)
	GetUpcoming(ctx context.Context, within time.Duration) ([]CalendarEvent, error)
	GetByCurrencyAndImpact(ctx context.Context, currency, impact string, limit int) ([]CalendarEvent, error)
	SaveImpactRecord(ctx context.Context, rec EventImpactRecord) error
	GetImpactRecords(ctx context.Context, eventID string) ([]EventImpactRecord, error)
	GetHistoricalSurprises(ctx context.Context, eventName, currency string, limit int) ([]float64, error)
}

// calendarRepo implements CalendarRepository
type calendarRepo struct {
	pool *pgxpool.Pool
}

// NewCalendarRepository creates a new calendar repository
func NewCalendarRepository(pool *pgxpool.Pool) CalendarRepository {
	return &calendarRepo{pool: pool}
}

// UpsertEvents inserts or updates calendar events
func (r *calendarRepo) UpsertEvents(ctx context.Context, events []CalendarEvent) (int, error) {
	if len(events) == 0 {
		return 0, nil
	}

	batch := &pgx.Batch{}
	for _, e := range events {
		batch.Queue(`
			INSERT INTO calendar_events 
			(event_id, title, country, currency, impact, scheduled_at, 
			 previous_value, forecast_value, actual_value, impact_direction,
			 surprise_score, surprise_label, revision_label, fetched_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW(), NOW())
			ON CONFLICT (event_id) DO UPDATE SET
				title = EXCLUDED.title,
				country = EXCLUDED.country,
				currency = EXCLUDED.currency,
				impact = EXCLUDED.impact,
				scheduled_at = EXCLUDED.scheduled_at,
				previous_value = EXCLUDED.previous_value,
				forecast_value = EXCLUDED.forecast_value,
				actual_value = EXCLUDED.actual_value,
				impact_direction = EXCLUDED.impact_direction,
				surprise_score = EXCLUDED.surprise_score,
				surprise_label = EXCLUDED.surprise_label,
				revision_label = EXCLUDED.revision_label,
				updated_at = NOW()
			WHERE calendar_events.actual_value IS DISTINCT FROM EXCLUDED.actual_value
		`, e.EventID, e.Title, e.Country, e.Currency, e.Impact, e.ScheduledAt,
			e.PreviousValue, e.ForecastValue, e.ActualValue, e.ImpactDirection,
			e.SurpriseScore, e.SurpriseLabel, e.RevisionLabel)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	var count int
	for i := 0; i < len(events); i++ {
		if _, err := br.Exec(); err != nil {
			return count, fmt.Errorf("upsert event %d failed: %w", i, err)
		}
		count++
	}

	if err := br.Close(); err != nil {
		return count, err
	}
	return count, nil
}

// UpdateActual updates the actual value and surprise score
func (r *calendarRepo) UpdateActual(ctx context.Context, eventID, actual string, surprise float64) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE calendar_events 
		SET actual_value = $1, surprise_score = $2, updated_at = NOW()
		WHERE event_id = $3
	`, actual, surprise, eventID)
	return err
}

// GetByDate retrieves events for a specific date
func (r *calendarRepo) GetByDate(ctx context.Context, date time.Time) ([]CalendarEvent, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT event_id, title, country, currency, impact, scheduled_at,
			   previous_value, forecast_value, actual_value, impact_direction,
			   surprise_score, surprise_label, revision_label, fetched_at, updated_at
		FROM calendar_events
		WHERE DATE(scheduled_at) = DATE($1)
		ORDER BY scheduled_at ASC
	`, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[CalendarEvent])
}

// GetUpcoming retrieves upcoming events within duration
func (r *calendarRepo) GetUpcoming(ctx context.Context, within time.Duration) ([]CalendarEvent, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT event_id, title, country, currency, impact, scheduled_at,
			   previous_value, forecast_value, actual_value, impact_direction,
			   surprise_score, surprise_label, revision_label, fetched_at, updated_at
		FROM calendar_events
		WHERE scheduled_at > NOW() 
		  AND scheduled_at <= NOW() + $1::interval
		  AND (actual_value IS NULL OR actual_value = '')
		ORDER BY scheduled_at ASC
	`, within.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[CalendarEvent])
}

// GetByCurrencyAndImpact retrieves events by currency and impact level
func (r *calendarRepo) GetByCurrencyAndImpact(ctx context.Context, currency, impact string, limit int) ([]CalendarEvent, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT event_id, title, country, currency, impact, scheduled_at,
			   previous_value, forecast_value, actual_value, impact_direction,
			   surprise_score, surprise_label, revision_label, fetched_at, updated_at
		FROM calendar_events
		WHERE currency = $1 AND impact = $2
		ORDER BY scheduled_at DESC
		LIMIT $3
	`, currency, impact, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[CalendarEvent])
}

// SaveImpactRecord saves an impact record
func (r *calendarRepo) SaveImpactRecord(ctx context.Context, rec EventImpactRecord) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO event_impact_records 
		(event_id, window, symbol, price_before, price_after, pct_change, recorded_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (event_id, window, symbol) DO UPDATE SET
			price_before = EXCLUDED.price_before,
			price_after = EXCLUDED.price_after,
			pct_change = EXCLUDED.pct_change,
			recorded_at = EXCLUDED.recorded_at
	`, rec.EventID, rec.Window, rec.Symbol, rec.PriceBefore, rec.PriceAfter, rec.PctChange, rec.RecordedAt)
	return err
}

// GetImpactRecords retrieves impact records for an event
func (r *calendarRepo) GetImpactRecords(ctx context.Context, eventID string) ([]EventImpactRecord, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT event_id, window, symbol, price_before, price_after, pct_change, recorded_at
		FROM event_impact_records
		WHERE event_id = $1
		ORDER BY window ASC
	`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[EventImpactRecord])
}

// GetHistoricalSurprises retrieves historical surprise scores
func (r *calendarRepo) GetHistoricalSurprises(ctx context.Context, eventName, currency string, limit int) ([]float64, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT sigma FROM calendar_surprise_history
		WHERE event_name = $1 AND currency = $2 AND sigma IS NOT NULL
		ORDER BY released_at DESC
		LIMIT $3
	`, eventName, currency, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []float64
	for rows.Next() {
		var sigma float64
		if err := rows.Scan(&sigma); err != nil {
			return nil, err
		}
		results = append(results, sigma)
	}
	return results, rows.Err()
}
