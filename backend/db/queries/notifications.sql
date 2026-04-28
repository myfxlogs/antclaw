-- name: CreateNotification :one
INSERT INTO notifications (user_id, type, title, body, data, priority)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, user_id, type, title, body, data, priority, is_read, created_at, read_at;

-- name: GetNotificationByID :one
SELECT * FROM notifications WHERE id = $1;

-- name: GetUnreadNotifications :many
SELECT * FROM notifications 
WHERE user_id = $1 AND is_read = false 
ORDER BY created_at DESC;

-- name: GetNotificationHistory :many
SELECT * FROM notifications 
WHERE user_id = $1 
ORDER BY created_at DESC 
LIMIT $2;

-- name: MarkNotificationRead :exec
UPDATE notifications SET is_read = true, read_at = NOW() WHERE id = $1;

-- name: MarkAllRead :exec
UPDATE notifications SET is_read = true, read_at = NOW() 
WHERE user_id = $1 AND is_read = false;
