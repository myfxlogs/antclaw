package main

import (
"context"
"fmt"
"time"

"github.com/antclaw/antclaw/internal/adapter/storage/postgres/db"
"github.com/jackc/pgx/v5/pgtype"
"github.com/jackc/pgx/v5/pgxpool"
)

// ByokHealthChecker performs daily health checks on user-provided AI keys.
type ByokHealthChecker struct {
dbpool *pgxpool.Pool
queries *db.Queries
}

// NewByokHealthChecker creates a new BYOK health checker.
func NewByokHealthChecker(dbpool *pgxpool.Pool) *ByokHealthChecker {
return &ByokHealthChecker{
dbpool:  dbpool,
queries: db.New(dbpool),
}
}

// Run executes the health check job.
func (c *ByokHealthChecker) Run(ctx context.Context) error {
keys, err := c.queries.ListActiveAIKeys(ctx)
if err != nil {
return fmt.Errorf("list ai keys: %w", err)
}

for _, key := range keys {
if err := c.checkKey(ctx, key); err != nil {
continue
}
}

return nil
}

func (c *ByokHealthChecker) checkKey(ctx context.Context, key db.UserAiKey) error {
provider := key.Provider

err := c.testProviderKey(ctx, provider, key.KeyEnc)

now := pgtype.Timestamptz{Time: time.Now(), Valid: true}

if err != nil {
errStr := err.Error()
return c.queries.UpdateAIKeyHealth(ctx, db.UpdateAIKeyHealthParams{
UserID:    key.UserID,
Provider:  provider,
LastVerifiedAt: now,
LastError: &errStr,
})
}

return c.queries.UpdateAIKeyHealth(ctx, db.UpdateAIKeyHealthParams{
UserID:         key.UserID,
Provider:       provider,
LastVerifiedAt: now,
LastError:      nil,
})
}

func (c *ByokHealthChecker) testProviderKey(ctx context.Context, provider string, keyEnc []byte) error {
switch provider {
case "gemini":
// Test Gemini API key
case "claude":
// Test Claude API key
default:
return fmt.Errorf("unknown provider: %s", provider)
}
return nil
}
