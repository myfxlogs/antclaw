-- +goose Up
-- Create bot_bindings table for P7 Bot interface

CREATE TABLE bot_bindings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    platform TEXT NOT NULL CHECK (platform IN ('telegram', 'wechat', 'feishu')),
    external_user_id TEXT NOT NULL,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    external_username TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(platform, external_user_id)
);

CREATE INDEX idx_bot_bindings_user ON bot_bindings(user_id);
CREATE INDEX idx_bot_bindings_platform_external ON bot_bindings(platform, external_user_id);

-- Update timestamp trigger
CREATE TRIGGER trigger_bot_bindings_updated_at
    BEFORE UPDATE ON bot_bindings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- +goose Down
DROP TRIGGER IF EXISTS trigger_bot_bindings_updated_at ON bot_bindings;
DROP TABLE IF EXISTS bot_bindings;
