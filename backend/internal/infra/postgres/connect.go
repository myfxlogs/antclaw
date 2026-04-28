// Package postgres provides database connection utilities.
package postgres

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPoolFromEnv creates a *pgxpool.Pool from environment variables.
// Priority: DATABASE_URL > ANTCLAW_DB_URL > individual ANTCLAW_DB_* vars.
func NewPoolFromEnv() (*pgxpool.Pool, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = os.Getenv("ANTCLAW_DB_URL")
	}
	if dbURL == "" {
		dbURL = buildDBURL()
	}
	return pgxpool.New(context.Background(), dbURL)
}

func buildDBURL() string {
	host := getenv("ANTCLAW_DB_HOST", "localhost")
	port := getenv("ANTCLAW_DB_PORT", "5432")
	user := getenv("ANTCLAW_DB_USER", "antclaw")
	pass := getenv("ANTCLAW_DB_PASSWORD", "antclaw")
	name := getenv("ANTCLAW_DB_NAME", "antclaw")
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, name)
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
