-- +goose Up
-- 客户端智能推送：用户业务级告警偏好。
-- 与 user_notification_prefs（通用通知偏好）互补：
--   - user_notification_prefs 控制通知通道（enabled_types / min_severity / push_enabled / 静默）
--   - user_alert_preferences 控制业务内容（货币 / 品种 / 影响级别 / 提醒提前量 / 各 channel 开关）
CREATE TABLE IF NOT EXISTS user_alert_preferences (
    user_id                UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    currencies             TEXT[]   NOT NULL DEFAULT ARRAY['USD','EUR','GBP','JPY','CHF','CAD','AUD','NZD']::TEXT[],
    symbols                TEXT[]   NOT NULL DEFAULT ARRAY[]::TEXT[],
    impacts                TEXT[]   NOT NULL DEFAULT ARRAY['high','medium']::TEXT[],
    reminder_minutes       INT[]    NOT NULL DEFAULT ARRAY[60,15]::INT[],
    high_impact_only       BOOLEAN  NOT NULL DEFAULT FALSE,
    daily_digest_enabled   BOOLEAN  NOT NULL DEFAULT TRUE,
    weekly_digest_enabled  BOOLEAN  NOT NULL DEFAULT TRUE,
    cot_alerts_enabled     BOOLEAN  NOT NULL DEFAULT TRUE,
    macro_alerts_enabled   BOOLEAN  NOT NULL DEFAULT TRUE,
    options_alerts_enabled BOOLEAN  NOT NULL DEFAULT TRUE,
    onchain_alerts_enabled BOOLEAN  NOT NULL DEFAULT TRUE,
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS user_alert_preferences;
