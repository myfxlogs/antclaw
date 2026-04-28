// Package postgres provides PostgreSQL access using pgx and sqlc.
// Supports: primary writes + read replica pools, hypertables for time-series.
// See: AntClaw-重构解决方案.md §4.2
package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool creates a new PostgreSQL connection pool
func NewPool(connString string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}
	return pool, nil
}
