-- name: CreateUser :one
INSERT INTO users (
    email, username, display_name, password_hash, locale, timezone, code_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1 AND deleted_at IS NULL;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL;

-- name: GetUserByCodeID :one
SELECT * FROM users WHERE code_id = $1 AND deleted_at IS NULL;

-- name: UpdateUserCodeID :exec
UPDATE users SET code_id = $2, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdatePasswordVersion :exec
UPDATE users SET password_version = password_version + 1, updated_at = NOW() WHERE id = $1;

-- name: UpdateUserPassword :exec
UPDATE users SET password_hash = $2, password_version = password_version + 1, updated_at = NOW() WHERE id = $1;

-- name: BanUser :exec
UPDATE users SET status = 'banned', updated_at = NOW() WHERE id = $1;

-- name: UnbanUser :exec
UPDATE users SET status = 'active', updated_at = NOW() WHERE id = $1;

-- name: UpdateUserRole :exec
UPDATE users SET role = $2, updated_at = NOW() WHERE id = $1;

-- name: UpdateEmailVerifiedAt :exec
UPDATE users SET email_verified_at = $2, updated_at = NOW() WHERE id = $1;

-- name: CreateSession :one
INSERT INTO sessions (user_id, user_agent, ip)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetSessionByID :one
SELECT * FROM sessions WHERE id = $1;

-- name: UpdateSessionLastSeen :exec
UPDATE sessions SET last_seen_at = NOW(), ip = $2 WHERE id = $1;

-- name: RevokeSession :exec
UPDATE sessions SET revoked_at = NOW() WHERE id = $1;

-- name: RevokeAllUserSessions :exec
UPDATE sessions SET revoked_at = NOW() WHERE user_id = $1;

-- name: ListUserSessions :many
SELECT * FROM sessions WHERE user_id = $1 ORDER BY last_seen_at DESC;

-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (session_id, user_id, expires_at)
VALUES ($1, $2, $3)
RETURNING jti;

-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens WHERE jti = $1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens SET revoked_at = NOW() WHERE jti = $1;

-- name: MarkRefreshTokenRotated :exec
UPDATE refresh_tokens SET rotated_to = $2 WHERE jti = $1;

-- name: RevokeAllUserRefreshTokens :exec
UPDATE refresh_tokens SET revoked_at = NOW() WHERE user_id = $1;

-- name: CreatePasswordResetToken :exec
INSERT INTO password_resets (token_hash, user_id, purpose)
VALUES ($1, $2, $3);

-- name: GetPasswordResetToken :one
SELECT * FROM password_resets WHERE token_hash = $1;

-- name: MarkPasswordResetTokenConsumed :exec
UPDATE password_resets SET consumed_at = NOW() WHERE token_hash = $1;

-- name: IsSessionRevoked :one
SELECT EXISTS(SELECT 1 FROM sessions WHERE id = $1 AND revoked_at IS NOT NULL);

-- name: IsRefreshTokenRevoked :one
SELECT EXISTS(SELECT 1 FROM refresh_tokens WHERE jti = $1 AND revoked_at IS NOT NULL);

-- name: ListUsers :many
SELECT * FROM users
WHERE deleted_at IS NULL
  AND ($1::text = '' OR email ILIKE '%' || $1 || '%')
  AND ($2::text = '' OR role = $2)
  AND ($3::boolean = false OR status = 'banned')
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: CountUsers :one
SELECT COUNT(*) FROM users WHERE deleted_at IS NULL;
