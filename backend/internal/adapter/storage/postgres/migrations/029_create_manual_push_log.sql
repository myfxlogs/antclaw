-- +goose Up
-- 管理端手动推送日志表。
-- 记录每次管理员手动推送的标题、正文、目标用户数、实际发送数、操作人等信息。
CREATE TABLE IF NOT EXISTS manual_push_log (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title         TEXT        NOT NULL,
    body          TEXT        NOT NULL DEFAULT '',
    severity      TEXT        NOT NULL DEFAULT 'normal',
    category      TEXT        NOT NULL DEFAULT 'system',
    target_count  INT         NOT NULL DEFAULT 0,
    sent_count    INT         NOT NULL DEFAULT 0,
    admin_user_id UUID        NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    target_user_ids TEXT[]    NOT NULL DEFAULT '{}',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_manual_push_log_created_at
    ON manual_push_log(created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS manual_push_log;
