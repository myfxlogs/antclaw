package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// COTRecord represents a COT report record
type COTRecord struct {
	ReportDate   time.Time `db:"report_date"`
	ContractCode string    `db:"contract_code"`
	Currency     string    `db:"currency"`
	NoncommLong  int64     `db:"noncomm_long"`
	NoncommShort int64     `db:"noncomm_short"`
	CommLong     int64     `db:"comm_long"`
	CommShort    int64     `db:"comm_short"`
	DealerLong   int64     `db:"dealer_long"`
	DealerShort  int64     `db:"dealer_short"`
	LevfundLong  int64     `db:"levfund_long"`
	LevfundShort int64     `db:"levfund_short"`
	MMLong       int64     `db:"mm_long"`
	MMShort      int64     `db:"mm_short"`
	SwapLong     int64     `db:"swap_long"`
	SwapShort    int64     `db:"swap_short"`
	TotalOI      int64     `db:"total_oi"`
	RawJSON      []byte    `db:"raw_json"`
}

// COTAnalysis represents COT analysis result
type COTAnalysis struct {
	ReportDate      time.Time `db:"report_date"`
	ContractCode    string    `db:"contract_code"`
	NetPosition     int64     `db:"net_position"`
	COTIndex        float64   `db:"cot_index"`
	Direction       string    `db:"direction"`
	SentimentScore  float64   `db:"sentiment_score"`
	WoWChange       int64     `db:"wow_change"`
	ZScore          float64   `db:"zscore"`
	Percentile      float64   `db:"percentile"`
}

// COTSignalOutcome tracks signal performance
type COTSignalOutcome struct {
	SignalID       int64     `db:"signal_id"`
	SignalType     string    `db:"signal_type"`
	ContractCode   string    `db:"contract_code"`
	IssuedAt       time.Time `db:"issued_at"`
	RawConfidence  float64   `db:"raw_confidence"`
	Return1W       *float64  `db:"return_1w"`
	Return2W       *float64  `db:"return_2w"`
	Return4W       *float64  `db:"return_4w"`
	Win            *bool     `db:"win"`
	EvaluatedAt    *time.Time `db:"evaluated_at"`
}

// COTCalibration stores calibration parameters
type COTCalibration struct {
	SignalType  string    `db:"signal_type"`
	PlattA      float64   `db:"platt_a"`
	PlattB      float64   `db:"platt_b"`
	WinRate     float64   `db:"win_rate"`
	SampleSize  int       `db:"sample_size"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// COTRepository defines COT data operations
type COTRepository interface {
	UpsertRecords(ctx context.Context, records []COTRecord) (int, error)
	GetHistory(ctx context.Context, contractCode string, weeks int) ([]COTRecord, error)
	GetLatestAll(ctx context.Context) (map[string]*COTAnalysis, error)
	SaveAnalysis(ctx context.Context, analyses []COTAnalysis) error
	SaveSignalOutcome(ctx context.Context, outcome COTSignalOutcome) error
	GetSignalStats(ctx context.Context, signalType string) (*COTCalibration, error)
	UpdateCalibration(ctx context.Context, cal COTCalibration) error
}

type cotRepo struct {
	pool *pgxpool.Pool
}

// NewCOTRepository creates a new COT repository
func NewCOTRepository(pool *pgxpool.Pool) COTRepository {
	return &cotRepo{pool: pool}
}

// UpsertRecords inserts or updates COT records
func (r *cotRepo) UpsertRecords(ctx context.Context, records []COTRecord) (int, error) {
	if len(records) == 0 {
		return 0, nil
	}

	batch := &pgx.Batch{}
	for _, rec := range records {
		batch.Queue(`
			INSERT INTO cot_records 
			(report_date, contract_code, currency, noncomm_long, noncomm_short,
			 comm_long, comm_short, dealer_long, dealer_short, levfund_long, levfund_short,
			 mm_long, mm_short, swap_long, swap_short, total_oi, raw_json)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
			ON CONFLICT (report_date, contract_code) DO UPDATE SET
				currency = EXCLUDED.currency,
				noncomm_long = EXCLUDED.noncomm_long,
				noncomm_short = EXCLUDED.noncomm_short,
				comm_long = EXCLUDED.comm_long,
				comm_short = EXCLUDED.comm_short,
				dealer_long = EXCLUDED.dealer_long,
				dealer_short = EXCLUDED.dealer_short,
				levfund_long = EXCLUDED.levfund_long,
				levfund_short = EXCLUDED.levfund_short,
				mm_long = EXCLUDED.mm_long,
				mm_short = EXCLUDED.mm_short,
				swap_long = EXCLUDED.swap_long,
				swap_short = EXCLUDED.swap_short,
				total_oi = EXCLUDED.total_oi,
				raw_json = EXCLUDED.raw_json
		`, rec.ReportDate, rec.ContractCode, rec.Currency, rec.NoncommLong, rec.NoncommShort,
			rec.CommLong, rec.CommShort, rec.DealerLong, rec.DealerShort, rec.LevfundLong, rec.LevfundShort,
			rec.MMLong, rec.MMShort, rec.SwapLong, rec.SwapShort, rec.TotalOI, rec.RawJSON)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	var count int
	for i := 0; i < len(records); i++ {
		if _, err := br.Exec(); err != nil {
			return count, fmt.Errorf("upsert record %d failed: %w", i, err)
		}
		count++
	}
	return count, br.Close()
}

// GetHistory retrieves historical COT records
func (r *cotRepo) GetHistory(ctx context.Context, contractCode string, weeks int) ([]COTRecord, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT report_date, contract_code, currency, noncomm_long, noncomm_short,
			   comm_long, comm_short, dealer_long, dealer_short, levfund_long, levfund_short,
			   mm_long, mm_short, swap_long, swap_short, total_oi, raw_json
		FROM cot_records
		WHERE contract_code = $1 
		  AND report_date >= NOW() - INTERVAL '1 week' * $2
		ORDER BY report_date DESC
	`, contractCode, weeks)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[COTRecord])
}

// GetLatestAll retrieves latest analysis for all contracts
func (r *cotRepo) GetLatestAll(ctx context.Context) (map[string]*COTAnalysis, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT ON (contract_code)
			report_date, contract_code, net_position, cot_index, direction,
			sentiment_score, wow_change, zscore, percentile
		FROM cot_analyses
		ORDER BY contract_code, report_date DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make(map[string]*COTAnalysis)
	for rows.Next() {
		var a COTAnalysis
		if err := rows.Scan(&a.ReportDate, &a.ContractCode, &a.NetPosition, &a.COTIndex,
			&a.Direction, &a.SentimentScore, &a.WoWChange, &a.ZScore, &a.Percentile); err != nil {
			return nil, err
		}
		results[a.ContractCode] = &a
	}
	return results, rows.Err()
}

// SaveAnalysis saves COT analysis results
func (r *cotRepo) SaveAnalysis(ctx context.Context, analyses []COTAnalysis) error {
	batch := &pgx.Batch{}
	for _, a := range analyses {
		batch.Queue(`
			INSERT INTO cot_analyses 
			(report_date, contract_code, net_position, cot_index, direction,
			 sentiment_score, wow_change, zscore, percentile)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (report_date, contract_code) DO UPDATE SET
				net_position = EXCLUDED.net_position,
				cot_index = EXCLUDED.cot_index,
				direction = EXCLUDED.direction,
				sentiment_score = EXCLUDED.sentiment_score,
				wow_change = EXCLUDED.wow_change,
				zscore = EXCLUDED.zscore,
				percentile = EXCLUDED.percentile
		`, a.ReportDate, a.ContractCode, a.NetPosition, a.COTIndex, a.Direction,
			a.SentimentScore, a.WoWChange, a.ZScore, a.Percentile)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < len(analyses); i++ {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("save analysis %d failed: %w", i, err)
		}
	}
	return br.Close()
}

// SaveSignalOutcome saves a signal outcome
func (r *cotRepo) SaveSignalOutcome(ctx context.Context, outcome COTSignalOutcome) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO cot_signal_outcomes 
		(signal_type, contract_code, issued_at, raw_confidence, return_1w, return_2w, return_4w, win, evaluated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (signal_id) DO UPDATE SET
			return_1w = EXCLUDED.return_1w,
			return_2w = EXCLUDED.return_2w,
			return_4w = EXCLUDED.return_4w,
			win = EXCLUDED.win,
			evaluated_at = EXCLUDED.evaluated_at
	`, outcome.SignalType, outcome.ContractCode, outcome.IssuedAt, outcome.RawConfidence,
		outcome.Return1W, outcome.Return2W, outcome.Return4W, outcome.Win, outcome.EvaluatedAt)
	return err
}

// GetSignalStats retrieves calibration stats for a signal type
func (r *cotRepo) GetSignalStats(ctx context.Context, signalType string) (*COTCalibration, error) {
	var cal COTCalibration
	err := r.pool.QueryRow(ctx, `
		SELECT signal_type, platt_a, platt_b, win_rate, sample_size, updated_at
		FROM cot_calibration
		WHERE signal_type = $1
	`, signalType).Scan(&cal.SignalType, &cal.PlattA, &cal.PlattB, &cal.WinRate, &cal.SampleSize, &cal.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &cal, nil
}

// UpdateCalibration updates calibration parameters
func (r *cotRepo) UpdateCalibration(ctx context.Context, cal COTCalibration) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO cot_calibration (signal_type, platt_a, platt_b, win_rate, sample_size, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (signal_type) DO UPDATE SET
			platt_a = EXCLUDED.platt_a,
			platt_b = EXCLUDED.platt_b,
			win_rate = EXCLUDED.win_rate,
			sample_size = EXCLUDED.sample_size,
			updated_at = EXCLUDED.updated_at
	`, cal.SignalType, cal.PlattA, cal.PlattB, cal.WinRate, cal.SampleSize, cal.UpdatedAt)
	return err
}
