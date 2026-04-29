-- +goose Up
-- 通知推送系统：扩展字段 + 用户偏好表
-- 设计：参考 Emulator/ark-intelligent 的 Telegram 推送模型，落地为 antclaw 的
--       "持久化 + 实时 SSE + 用户偏好 + 去重" 模式。

ALTER TABLE notifications
    ADD COLUMN IF NOT EXISTS category   TEXT NOT NULL DEFAULT 'system',
    ADD COLUMN IF NOT EXISTS dedup_key  TEXT,
    ADD COLUMN IF NOT EXISTS severity   VARCHAR(16) NOT NULL DEFAULT 'normal';

-- 历史数据：旧 priority 视为 severity 的初值（仅做一次性回填）。
UPDATE notifications
   SET severity = COALESCE(priority, 'normal')
 WHERE severity = 'normal' AND priority IS NOT NULL;

-- dedup_key 唯一性窗口由调用方在 Redis 上控制 TTL；此处仅做查询索引。
CREATE INDEX IF NOT EXISTS idx_notifications_user_category
    ON notifications (user_id, category, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_dedup
    ON notifications (dedup_key) WHERE dedup_key IS NOT NULL;

-- 用户通知偏好（每用户单行）。所有字段都给安全默认值，未设置 → 全开。
CREATE TABLE IF NOT EXISTS user_notification_prefs (
    user_id        UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    enabled_types  TEXT[]      NOT NULL DEFAULT ARRAY['alert','signal','system','digest']::TEXT[],
    min_severity   VARCHAR(16) NOT NULL DEFAULT 'low',  -- low|normal|high|critical
    quiet_start    TIME        NOT NULL DEFAULT '00:00',
    quiet_end      TIME        NOT NULL DEFAULT '00:00', -- start==end → 不静默
    timezone       TEXT        NOT NULL DEFAULT 'UTC',
    push_enabled   BOOLEAN     NOT NULL DEFAULT true,
    email_enabled  BOOLEAN     NOT NULL DEFAULT false,
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS user_notification_prefs;
DROP INDEX IF EXISTS idx_notifications_dedup;
DROP INDEX IF EXISTS idx_notifications_user_category;
ALTER TABLE notifications
    DROP COLUMN IF EXISTS severity,
    DROP COLUMN IF EXISTS dedup_key,
    DROP COLUMN IF EXISTS category;
