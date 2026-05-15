-- +goose Up
-- Create mt_accounts table for MetaTrader trading account bindings.

CREATE TABLE IF NOT EXISTS mt_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    login TEXT NOT NULL,
    password TEXT NOT NULL,
    mt_type TEXT NOT NULL CHECK (mt_type IN ('MT4', 'MT5')),
    broker_company TEXT NOT NULL DEFAULT '',
    broker_server TEXT NOT NULL DEFAULT '',
    broker_host TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'connecting'
        CHECK (status IN ('connecting', 'connected', 'disconnected', 'error', 'disabled')),
    token TEXT NOT NULL DEFAULT '',
    currency TEXT NOT NULL DEFAULT '',
    account_type TEXT NOT NULL DEFAULT '',
    alias TEXT NOT NULL DEFAULT '',
    last_error TEXT NOT NULL DEFAULT '',
    is_disabled BOOLEAN NOT NULL DEFAULT false,
    is_investor BOOLEAN NOT NULL DEFAULT false,
    balance DOUBLE PRECISION NOT NULL DEFAULT 0,
    credit DOUBLE PRECISION NOT NULL DEFAULT 0,
    equity DOUBLE PRECISION NOT NULL DEFAULT 0,
    margin DOUBLE PRECISION NOT NULL DEFAULT 0,
    free_margin DOUBLE PRECISION NOT NULL DEFAULT 0,
    margin_level DOUBLE PRECISION NOT NULL DEFAULT 0,
    profit DOUBLE PRECISION NOT NULL DEFAULT 0,
    profit_percent DOUBLE PRECISION NOT NULL DEFAULT 0,
    leverage INTEGER NOT NULL DEFAULT 0,
    connected_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, login, mt_type)
);

CREATE INDEX IF NOT EXISTS idx_mt_accounts_user ON mt_accounts(user_id);
CREATE INDEX IF NOT EXISTS idx_mt_accounts_status ON mt_accounts(status);
CREATE INDEX IF NOT EXISTS idx_mt_accounts_mt_type ON mt_accounts(mt_type);

-- Update timestamp trigger
DROP TRIGGER IF EXISTS trigger_mt_accounts_updated_at ON mt_accounts;
CREATE TRIGGER trigger_mt_accounts_updated_at
    BEFORE UPDATE ON mt_accounts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- +goose Down
DROP TRIGGER IF EXISTS trigger_mt_accounts_updated_at ON mt_accounts;
DROP TABLE IF EXISTS mt_accounts;
