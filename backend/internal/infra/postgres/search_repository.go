// Package postgres provides the Search Repository for cross-entity search.
package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrQueryTooShort is returned when a search query is shorter than the minimum length.
var ErrQueryTooShort = errors.New("search: query must be at least 2 characters")

// SearchUserRow is a flat DB row for user search results.
type SearchUserRow struct {
	UserID        string
	DisplayName   string
	Tier          string
	FollowerCount int32
}

// SearchSymbolRow is a flat DB row for symbol search results.
type SearchSymbolRow struct {
	Symbol      string
	DisplayName string
	Market      string
}

// SearchRepository defines data operations for cross-entity search (P1).
type SearchRepository interface {
	SearchUsers(ctx context.Context, query string, limit int32) ([]*SearchUserRow, error)
	SearchPosts(ctx context.Context, query string, limit int32) ([]*FeedPostRow, [][]string, error)
	SearchSymbols(ctx context.Context, query string, limit int32) ([]*SearchSymbolRow, error)
}

type searchRepo struct{ pool *pgxpool.Pool }

func NewSearchRepository(pool *pgxpool.Pool) SearchRepository { return &searchRepo{pool: pool} }

func (r *searchRepo) SearchUsers(ctx context.Context, query string, limit int32) ([]*SearchUserRow, error) {
	if len(query) < 2 {
		return nil, ErrQueryTooShort
	}
	rows, err := r.pool.Query(ctx,
		`SELECT id::text, `+PublicDisplayNameExpr+`, 'normal',
			(SELECT COUNT(*)::int4 FROM alfq_follows WHERE following_id = users.id)
		 FROM users
		 WHERE display_name ILIKE '%' || $1 || '%'
		 ORDER BY (SELECT COUNT(*) FROM alfq_follows WHERE following_id = users.id) DESC
		 LIMIT $2`,
		query, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*SearchUserRow
	for rows.Next() {
		u := &SearchUserRow{}
		if err := rows.Scan(&u.UserID, &u.DisplayName, &u.Tier, &u.FollowerCount); err != nil {
			return nil, err
		}
		results = append(results, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *searchRepo) SearchPosts(ctx context.Context, query string, limit int32) ([]*FeedPostRow, [][]string, error) {
	if len(query) < 2 {
		return nil, nil, ErrQueryTooShort
	}
	rows, err := r.pool.Query(ctx,
		`SELECT `+feedSelectColumns+`
		 FROM alfq_posts p LEFT JOIN alfq_post_stats s ON s.post_id = p.id
		 WHERE p.visibility = 'public'
		   AND p.content ILIKE '%' || $1 || '%'
		 ORDER BY p.created_at DESC, p.id DESC
		 LIMIT $2`,
		query, limit,
	)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var posts []*FeedPostRow
	for rows.Next() {
		row, err := scanFeedPost(rows)
		if err != nil {
			return nil, nil, err
		}
		posts = append(posts, row)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	// Search results: liked_by not needed (no currentUserID in search context), return empty arrays
	likedBy := make([][]string, len(posts))
	return posts, likedBy, nil
}

func (r *searchRepo) SearchSymbols(ctx context.Context, query string, limit int32) ([]*SearchSymbolRow, error) {
	// Search from price_daily table for real symbols only.
	// Skip synthetic/derived symbols; only return what exists in the database.
	rows, err := r.pool.Query(ctx,
		`SELECT DISTINCT symbol, symbol AS display_name,
			CASE
				WHEN symbol LIKE '%:%' THEN 'crypto'
				WHEN symbol LIKE '%=X' THEN 'forex'
				ELSE 'stock'
			END AS market
		 FROM price_daily
		 WHERE symbol ILIKE '%' || $1 || '%'
		 ORDER BY symbol
		 LIMIT $2`,
		query, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*SearchSymbolRow
	for rows.Next() {
		s := &SearchSymbolRow{}
		if err := rows.Scan(&s.Symbol, &s.DisplayName, &s.Market); err != nil {
			return nil, err
		}
		results = append(results, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}
