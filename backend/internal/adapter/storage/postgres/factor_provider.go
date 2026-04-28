package postgres

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/antclaw/antclaw/internal/service/signals"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FactorPgProvider struct{ pool *pgxpool.Pool }

func NewFactorPgProvider(pool *pgxpool.Pool) *FactorPgProvider { return &FactorPgProvider{pool: pool} }

func (p *FactorPgProvider) GetRanking(ctx context.Context, category string, asOf time.Time) (*signals.RankingResult, error) {
	rows, err := p.pool.Query(ctx, `
WITH latest AS (
	SELECT MAX(time) AS t FROM factor_rankings WHERE time <= $1
)
SELECT e.symbol, e.rank, e.raw_score, e.norm_score, fr.weights
FROM factor_ranking_entries e
JOIN factor_rankings fr ON fr.snapshot_id = e.snapshot_id
WHERE fr.time = (SELECT t FROM latest)
ORDER BY e.rank ASC`, asOf)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := &signals.RankingResult{AsOf: asOf, Weights: map[string]float64{}, Items: []signals.RankItem{}}
	symbolSet := categorySymbolsSet(category)
	for rows.Next() {
		var item signals.RankItem
		var rawWeights []byte
		if err := rows.Scan(&item.Symbol, &item.Rank, &item.RawScore, &item.NormScore, &rawWeights); err != nil {
			return nil, err
		}
		if len(symbolSet) > 0 && !symbolSet[strings.ToUpper(item.Symbol)] {
			continue
		}
		if len(res.Weights) == 0 && len(rawWeights) > 0 {
			_ = json.Unmarshal(rawWeights, &res.Weights)
		}
		res.Items = append(res.Items, item)
	}
	return res, rows.Err()
}

func (p *FactorPgProvider) GetSymbolFactors(ctx context.Context, symbol string, asOf time.Time) (*signals.FactorBreakdown, error) {
	var b signals.FactorBreakdown
	var raw []byte
	err := p.pool.QueryRow(ctx, `
WITH latest AS (
	SELECT MAX(time) AS t FROM factor_rankings WHERE time <= $1
)
SELECT e.breakdown
FROM factor_ranking_entries e
JOIN factor_rankings fr ON fr.snapshot_id = e.snapshot_id
WHERE fr.time = (SELECT t FROM latest) AND e.symbol = $2
LIMIT 1`, asOf, strings.ToUpper(symbol)).Scan(&raw)
	if err != nil {
		return nil, err
	}
	var tmp map[string]float64
	if err := json.Unmarshal(raw, &tmp); err != nil {
		return nil, err
	}
	b.Symbol = strings.ToUpper(symbol)
	b.AsOf = asOf
	b.Momentum = tmp["momentum"]
	b.LowVol = tmp["low_vol"]
	b.Trend = tmp["trend"]
	b.Carry = tmp["carry"]
	b.Crowding = tmp["crowding"]
	b.Residual = tmp["residual"]
	b.Composite = tmp["composite"]
	return &b, nil
}

func categorySymbolsSet(category string) map[string]bool {
	if strings.EqualFold(category, "all") || strings.TrimSpace(category) == "" {
		return nil
	}
	set := map[string]bool{}
	for _, s := range signals.SymbolsByCategory(category) {
		set[strings.ToUpper(s)] = true
	}
	return set
}
