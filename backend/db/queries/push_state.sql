-- name: GetPushState :one
SELECT id, user_id, event_key, push_type, last_sent_at, payload_hash, created_at, updated_at
  FROM notification_push_state
 WHERE user_id = $1 AND event_key = $2 AND push_type = $3;

-- name: UpsertPushState :one
INSERT INTO notification_push_state (user_id, event_key, push_type, last_sent_at, payload_hash)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (user_id, event_key, push_type)
DO UPDATE SET last_sent_at = EXCLUDED.last_sent_at,
              payload_hash = EXCLUDED.payload_hash,
              updated_at = NOW()
RETURNING id, user_id, event_key, push_type, last_sent_at, payload_hash, created_at, updated_at;

-- name: ListPushStatesByUser :many
SELECT id, user_id, event_key, push_type, last_sent_at, payload_hash, created_at, updated_at
  FROM notification_push_state
 WHERE user_id = $1
 ORDER BY last_sent_at DESC
 LIMIT $2;

-- name: DeleteOldPushStates :exec
DELETE FROM notification_push_state
 WHERE last_sent_at < $1;
