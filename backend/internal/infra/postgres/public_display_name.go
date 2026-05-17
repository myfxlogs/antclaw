// Package postgres provides shared public display name helpers.
package postgres

// PublicDisplayNameExpr is the SQL expression for a safe public display name.
// Falls back from display_name → username → code_id → anonymous.
// Must be used in all public-facing queries (Feed, Search, Trader, UserList).
// Never leaks email.
const PublicDisplayNameExpr = `COALESCE(NULLIF(display_name, ''), username, code_id, 'User-' || LEFT(id::text, 8))`
