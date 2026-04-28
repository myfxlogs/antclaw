-- name: CreateAuditLog :one
INSERT INTO audit_logs (user_id, action, resource, details, ip_address, user_agent, hash_prev, hash_self)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id;

-- name: ListAuditLogs :many
SELECT * FROM audit_logs
WHERE ($1::uuid IS NULL OR user_id = $1)
  AND ($2::text = '' OR action = $2)
  AND ($3::timestamptz IS NULL OR created_at >= $3)
  AND ($4::timestamptz IS NULL OR created_at <= $4)
ORDER BY id DESC
LIMIT $5 OFFSET $6;

-- name: GetLastAuditLog :one
SELECT * FROM audit_logs ORDER BY id DESC LIMIT 1;

-- name: GetAuditLogByID :one
SELECT * FROM audit_logs WHERE id = $1;
