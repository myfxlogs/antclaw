package postgres

import (
	"context"

	"github.com/antclaw/antclaw/internal/service/signals"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type COTPgProvider struct{ pool *pgxpool.Pool }

func NewCOTPgProvider(pool *pgxpool.Pool) *COTPgProvider { return &COTPgProvider{pool: pool} }

func (p *COTPgProvider) GetLatestAnalysis(ctx context.Context, contractCode string) (*signals.COTAnalysis, error) {
	var a signals.COTAnalysis
	err := p.pool.QueryRow(ctx, `SELECT report_date, contract_code, net_position, cot_index, direction, sentiment_score, wow_change, zscore, percentile FROM cot_analyses WHERE contract_code=$1 ORDER BY report_date DESC LIMIT 1`, contractCode).
		Scan(&a.ReportDate, &a.ContractCode, &a.NetPosition, &a.COTIndex, &a.Direction, &a.SentimentScore, &a.WoWChange, &a.ZScore, &a.Percentile)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (p *COTPgProvider) GetAnalysisHistory(ctx context.Context, contractCode string, weeks int) ([]signals.COTAnalysis, error) {
	rows, err := p.pool.Query(ctx, `SELECT report_date, contract_code, net_position, cot_index, direction, sentiment_score, wow_change, zscore, percentile FROM cot_analyses WHERE contract_code=$1 AND report_date >= NOW() - INTERVAL '1 week' * $2 ORDER BY report_date DESC`, contractCode, weeks)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (signals.COTAnalysis, error) {
		var a signals.COTAnalysis
		err := row.Scan(&a.ReportDate, &a.ContractCode, &a.NetPosition, &a.COTIndex, &a.Direction, &a.SentimentScore, &a.WoWChange, &a.ZScore, &a.Percentile)
		return a, err
	})
}

func (p *COTPgProvider) GetCurrencyExtremes(ctx context.Context, threshold float64) ([]signals.COTAnalysis, error) {
	rows, err := p.pool.Query(ctx, `SELECT report_date, contract_code, net_position, cot_index, direction, sentiment_score, wow_change, zscore, percentile FROM cot_analyses WHERE ABS(zscore) >= $1 ORDER BY ABS(zscore) DESC LIMIT 100`, threshold)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (signals.COTAnalysis, error) {
		var a signals.COTAnalysis
		err := row.Scan(&a.ReportDate, &a.ContractCode, &a.NetPosition, &a.COTIndex, &a.Direction, &a.SentimentScore, &a.WoWChange, &a.ZScore, &a.Percentile)
		return a, err
	})
}
