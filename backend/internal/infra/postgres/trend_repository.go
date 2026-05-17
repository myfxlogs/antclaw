// Package postgres provides the Trend Repository for trending topics and hot symbols (P1).
package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TrendingTopicRow is a flat DB row for trending topic aggregation.
type TrendingTopicRow struct {
	Topic           string
	PostCount       int32
	EngagementCount int32
}

// HotSymbolRow is a flat DB row for hot symbol aggregation.
type HotSymbolRow struct {
	Symbol          string
	PostCount       int32
	SignalCount     int32
	EngagementCount int32
}

// TrendRepository defines data operations for trends (P1).
type TrendRepository interface {
	ListTrendingTopics(ctx context.Context, window string, limit int32) ([]*TrendingTopicRow, error)
	ListHotSymbols(ctx context.Context, window string, limit int32) ([]*HotSymbolRow, error)
}

type trendRepo struct{ pool *pgxpool.Pool }

func NewTrendRepository(pool *pgxpool.Pool) TrendRepository { return &trendRepo{pool: pool} }

// windowInterval returns the PostgreSQL INTERVAL for the given window.
// Only 1h, 24h, 7d are valid; caller must validate.
func windowInterval(window string) string {
	switch window {
	case "1h":
		return "1 hour"
	case "24h":
		return "24 hours"
	case "7d":
		return "7 days"
	default:
		return "24 hours"
	}
}

func (r *trendRepo) ListTrendingTopics(ctx context.Context, window string, limit int32) ([]*TrendingTopicRow, error) {
	// S12-P1-03: prefer pre-aggregated table, fall back to real-time
	rows, err := r.pool.Query(ctx,
		`SELECT topic, post_count, engagement_count
		 FROM trend_topics
		 WHERE window = $1
		 ORDER BY engagement_count DESC, post_count DESC
		 LIMIT $2`,
		window, limit,
	)
	if err != nil || rows == nil {
		// Fall back to real-time computation if pre-agg table not available
		return r.listTrendingTopicsRealtime(ctx, window, limit)
	}
	defer rows.Close()

	var results []*TrendingTopicRow
	for rows.Next() {
		t := &TrendingTopicRow{}
		if err := rows.Scan(&t.Topic, &t.PostCount, &t.EngagementCount); err != nil {
			return r.listTrendingTopicsRealtime(ctx, window, limit)
		}
		results = append(results, t)
	}
	if len(results) == 0 {
		return r.listTrendingTopicsRealtime(ctx, window, limit)
	}
	return results, rows.Err()
}

// listTrendingTopicsRealtime computes trending topics via live scan (fallback).
func (r *trendRepo) listTrendingTopicsRealtime(ctx context.Context, window string, limit int32) ([]*TrendingTopicRow, error) {
	iv := windowInterval(window)
	rows, err := r.pool.Query(ctx,
		`WITH recent_posts AS (
			SELECT content,
				(SELECT COUNT(*) FROM alfq_likes WHERE post_id = p.id) +
				(SELECT COUNT(*) FROM alfq_comments WHERE post_id = p.id) +
				(SELECT COUNT(*) FROM alfq_posts WHERE original_post_id = p.id AND post_type = 'share')
				AS engagement
			FROM alfq_posts p
			WHERE p.visibility = 'public'
			  AND p.created_at >= NOW() - $1::INTERVAL
		)
		SELECT topic, COUNT(*)::int4 AS post_count, SUM(engagement)::int4 AS engagement_count
		FROM (
			SELECT unnest(regexp_matches(content, '#([A-Za-z0-9_]+)', 'g')) AS topic, engagement
			FROM recent_posts
		) t
		GROUP BY topic
		ORDER BY engagement_count DESC, post_count DESC
		LIMIT $2`,
		iv, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*TrendingTopicRow
	for rows.Next() {
		t := &TrendingTopicRow{}
		if err := rows.Scan(&t.Topic, &t.PostCount, &t.EngagementCount); err != nil {
			return nil, err
		}
		results = append(results, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *trendRepo) ListHotSymbols(ctx context.Context, window string, limit int32) ([]*HotSymbolRow, error) {
	// S12-P1-03: prefer pre-aggregated table, fall back to real-time
	rows, err := r.pool.Query(ctx,
		`SELECT symbol, post_count, signal_count, engagement_count
		 FROM hot_symbols_agg
		 WHERE window = $1
		 ORDER BY engagement_count DESC, post_count DESC
		 LIMIT $2`,
		window, limit,
	)
	if err != nil || rows == nil {
		return r.listHotSymbolsRealtime(ctx, window, limit)
	}
	defer rows.Close()

	var results []*HotSymbolRow
	for rows.Next() {
		h := &HotSymbolRow{}
		if err := rows.Scan(&h.Symbol, &h.PostCount, &h.SignalCount, &h.EngagementCount); err != nil {
			return r.listHotSymbolsRealtime(ctx, window, limit)
		}
		results = append(results, h)
	}
	if len(results) == 0 {
		return r.listHotSymbolsRealtime(ctx, window, limit)
	}
	return results, rows.Err()
}

// listHotSymbolsRealtime computes hot symbols via live scan (fallback).
func (r *trendRepo) listHotSymbolsRealtime(ctx context.Context, window string, limit int32) ([]*HotSymbolRow, error) {
	iv := windowInterval(window)
	rows, err := r.pool.Query(ctx,
		`WITH recent_posts AS (
			SELECT signal_pair,
				post_type,
				(SELECT COUNT(*) FROM alfq_likes WHERE post_id = p.id) +
				(SELECT COUNT(*) FROM alfq_comments WHERE post_id = p.id) +
				(SELECT COUNT(*) FROM alfq_posts WHERE original_post_id = p.id AND post_type = 'share')
				AS engagement
			FROM alfq_posts p
			WHERE p.visibility = 'public'
			  AND p.created_at >= NOW() - $1::INTERVAL
			  AND p.signal_pair != ''
		)
		SELECT signal_pair AS symbol,
			COUNT(*)::int4 AS post_count,
			COUNT(*) FILTER (WHERE post_type = 'signal_card')::int4 AS signal_count,
			SUM(engagement)::int4 AS engagement_count
		FROM recent_posts
		GROUP BY signal_pair
		ORDER BY engagement_count DESC, post_count DESC
		LIMIT $2`,
		iv, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*HotSymbolRow
	for rows.Next() {
		h := &HotSymbolRow{}
		if err := rows.Scan(&h.Symbol, &h.PostCount, &h.SignalCount, &h.EngagementCount); err != nil {
			return nil, err
		}
		results = append(results, h)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}
