-- name: CreateNotification :one
-- 注意：调用方必须负责 dedup_key 的去重（Redis SETEX）；这里只做存档。
INSERT INTO notifications (user_id, type, category, title, body, data, priority, severity, dedup_key)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, user_id, type, category, title, body, data, priority, severity, dedup_key, is_read, created_at, read_at;

-- name: GetNotificationByID :one
SELECT * FROM notifications WHERE id = $1;

-- name: GetNotificationByIDForUser :one
SELECT * FROM notifications WHERE id = $1 AND user_id = $2;

-- name: GetUnreadNotifications :many
SELECT * FROM notifications
 WHERE user_id = $1 AND is_read = false
 ORDER BY created_at DESC
 LIMIT $2;

-- name: CountUnreadNotifications :one
SELECT COUNT(*)::bigint AS n FROM notifications
 WHERE user_id = $1 AND is_read = false;

-- name: GetNotificationHistory :many
SELECT * FROM notifications
 WHERE user_id = $1
 ORDER BY created_at DESC
 LIMIT $2;

-- name: MarkNotificationRead :exec
UPDATE notifications
   SET is_read = true, read_at = NOW()
 WHERE id = $1 AND user_id = $2;

-- name: MarkAllRead :exec
UPDATE notifications
   SET is_read = true, read_at = NOW()
 WHERE user_id = $1 AND is_read = false;
