-- name: GetUserNotificationPrefs :one
SELECT user_id, enabled_types, min_severity, quiet_start, quiet_end, timezone,
       push_enabled, email_enabled, updated_at
  FROM user_notification_prefs
 WHERE user_id = $1;

-- name: UpsertUserNotificationPrefs :one
INSERT INTO user_notification_prefs (
    user_id, enabled_types, min_severity, quiet_start, quiet_end, timezone, push_enabled, email_enabled
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (user_id) DO UPDATE SET
    enabled_types = EXCLUDED.enabled_types,
    min_severity  = EXCLUDED.min_severity,
    quiet_start   = EXCLUDED.quiet_start,
    quiet_end     = EXCLUDED.quiet_end,
    timezone      = EXCLUDED.timezone,
    push_enabled  = EXCLUDED.push_enabled,
    email_enabled = EXCLUDED.email_enabled,
    updated_at    = NOW()
RETURNING user_id, enabled_types, min_severity, quiet_start, quiet_end, timezone,
          push_enabled, email_enabled, updated_at;
