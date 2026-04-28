-- name: CreateOrUpdateUserAIKey :exec
INSERT INTO user_ai_keys (user_id, provider, key_enc, key_fingerprint)
VALUES ($1, $2, $3, $4)
ON CONFLICT (user_id, provider) DO UPDATE
SET key_enc = $3, key_fingerprint = $4, updated_at = NOW();

-- name: GetUserAIKey :one
SELECT * FROM user_ai_keys WHERE user_id = $1 AND provider = $2;

-- name: DeleteUserAIKey :exec
DELETE FROM user_ai_keys WHERE user_id = $1 AND provider = $2;

-- name: UpdateAIKeyVerified :exec
UPDATE user_ai_keys SET last_verified_at = NOW(), last_error = NULL WHERE user_id = $1 AND provider = $2;

-- name: UpdateAIKeyError :exec
UPDATE user_ai_keys SET last_error = $3 WHERE user_id = $1 AND provider = $2;

-- name: ListActiveAIKeys :many
SELECT * FROM user_ai_keys WHERE last_error IS NULL OR last_error = '';

-- name: UpdateAIKeyHealth :exec
UPDATE user_ai_keys 
SET last_verified_at = COALESCE($3, last_verified_at), 
    last_error = COALESCE($4, last_error),
    updated_at = NOW()
WHERE user_id = $1 AND provider = $2;
