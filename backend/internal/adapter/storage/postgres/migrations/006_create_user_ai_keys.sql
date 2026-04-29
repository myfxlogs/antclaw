-- +goose Up
CREATE TABLE IF NOT EXISTS user_ai_keys (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider TEXT NOT NULL CHECK (provider IN ('gemini', 'claude')),
    key_enc BYTEA NOT NULL,
    key_fingerprint TEXT NOT NULL,
    last_verified_at TIMESTAMPTZ,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, provider)
);

DROP TRIGGER IF EXISTS trigger_user_ai_keys_updated_at ON user_ai_keys;
CREATE TRIGGER trigger_user_ai_keys_updated_at
    BEFORE UPDATE ON user_ai_keys
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- +goose Down
DROP TABLE IF EXISTS user_ai_keys;
