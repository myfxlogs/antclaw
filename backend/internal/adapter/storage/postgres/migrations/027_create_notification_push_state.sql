-- +goose Up
-- 客户端智能推送：跨重启去重状态表。
-- 每个 (user_id, event_key, push_type) 记录最近一次发送时间与 payload hash，
-- 保证 worker 重启后不会对同一事件重复推送。
CREATE TABLE IF NOT EXISTS notification_push_state (
    id           BIGSERIAL PRIMARY KEY,
    user_id      UUID        NOT NULL,
    event_key    TEXT        NOT NULL,
    push_type    TEXT        NOT NULL,  -- calendar_pre / calendar_actual / surprise / digest / cot / regime / macro / options / onchain
    last_sent_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    payload_hash TEXT        NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, event_key, push_type)
);

CREATE INDEX IF NOT EXISTS idx_notification_push_state_user_time
    ON notification_push_state(user_id, last_sent_at DESC);

-- 定期清理超过 90 天的旧 push state，防止表无限膨胀。
-- 由 cron / worker 周期调用，不在 migration 中创建定时任务。

-- +goose Down
DROP TABLE IF EXISTS notification_push_state;
