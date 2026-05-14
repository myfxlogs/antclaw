-- name: GetUserAlertPreferences :one
SELECT user_id, currencies, symbols, impacts, reminder_minutes,
       high_impact_only,
       daily_digest_enabled, weekly_digest_enabled,
       cot_alerts_enabled, macro_alerts_enabled,
       options_alerts_enabled, onchain_alerts_enabled,
       updated_at
  FROM user_alert_preferences
 WHERE user_id = $1;

-- name: UpsertUserAlertPreferences :one
INSERT INTO user_alert_preferences (
    user_id, currencies, symbols, impacts, reminder_minutes,
    high_impact_only,
    daily_digest_enabled, weekly_digest_enabled,
    cot_alerts_enabled, macro_alerts_enabled,
    options_alerts_enabled, onchain_alerts_enabled
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
ON CONFLICT (user_id) DO UPDATE SET
    currencies             = EXCLUDED.currencies,
    symbols                = EXCLUDED.symbols,
    impacts                = EXCLUDED.impacts,
    reminder_minutes       = EXCLUDED.reminder_minutes,
    high_impact_only       = EXCLUDED.high_impact_only,
    daily_digest_enabled   = EXCLUDED.daily_digest_enabled,
    weekly_digest_enabled  = EXCLUDED.weekly_digest_enabled,
    cot_alerts_enabled     = EXCLUDED.cot_alerts_enabled,
    macro_alerts_enabled   = EXCLUDED.macro_alerts_enabled,
    options_alerts_enabled = EXCLUDED.options_alerts_enabled,
    onchain_alerts_enabled = EXCLUDED.onchain_alerts_enabled,
    updated_at             = NOW()
RETURNING user_id, currencies, symbols, impacts, reminder_minutes,
          high_impact_only,
          daily_digest_enabled, weekly_digest_enabled,
          cot_alerts_enabled, macro_alerts_enabled,
          options_alerts_enabled, onchain_alerts_enabled,
          updated_at;

-- name: ListUsersWithAlertPrefs :many
-- 分批扫描用户（用于 digest 批量推送），返回有偏好的用户及偏好。
SELECT u.id AS user_id,
       COALESCE(ap.currencies,           ARRAY['USD','EUR','GBP','JPY','CHF','CAD','AUD','NZD']::TEXT[])  AS currencies,
       COALESCE(ap.symbols,              ARRAY[]::TEXT[])                                                 AS symbols,
       COALESCE(ap.impacts,              ARRAY['high','medium']::TEXT[])                                  AS impacts,
       COALESCE(ap.reminder_minutes,     ARRAY[60,15]::INT[])                                            AS reminder_minutes,
       COALESCE(ap.high_impact_only,     FALSE)                                                           AS high_impact_only,
       COALESCE(ap.daily_digest_enabled, TRUE)                                                            AS daily_digest_enabled,
       COALESCE(ap.weekly_digest_enabled, TRUE)                                                           AS weekly_digest_enabled,
       COALESCE(ap.cot_alerts_enabled,   TRUE)                                                            AS cot_alerts_enabled,
       COALESCE(ap.macro_alerts_enabled, TRUE)                                                            AS macro_alerts_enabled,
       COALESCE(ap.options_alerts_enabled, TRUE)                                                          AS options_alerts_enabled,
       COALESCE(ap.onchain_alerts_enabled, TRUE)                                                          AS onchain_alerts_enabled,
       COALESCE(np.push_enabled,   TRUE)                                                                  AS push_enabled,
       COALESCE(np.min_severity,   'low')                                                                 AS min_severity,
       COALESCE(np.quiet_start,    '00:00')                                                               AS quiet_start,
       COALESCE(np.quiet_end,      '00:00')                                                               AS quiet_end,
       COALESCE(np.timezone,       'UTC')                                                                 AS timezone
  FROM users u
  LEFT JOIN user_alert_preferences ap ON ap.user_id = u.id
  LEFT JOIN user_notification_prefs np ON np.user_id = u.id
 WHERE u.id > $1   -- 游标分页（按 UUID 升序），避免 OFFSET 扫描膨胀
   AND u.role != 'admin'
   AND u.deleted_at IS NULL
 ORDER BY u.id
 LIMIT $2;
