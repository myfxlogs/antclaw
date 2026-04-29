-- +goose Up
CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action TEXT NOT NULL,
    resource TEXT NOT NULL DEFAULT '',
    details TEXT NOT NULL DEFAULT '',
    ip_address TEXT,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    hash_prev BYTEA NOT NULL DEFAULT '\x',
    hash_self BYTEA NOT NULL
);

-- 强制 append-only：禁止 UPDATE 和 DELETE
CREATE OR REPLACE FUNCTION audit_logs_no_update()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'audit_logs is append-only: UPDATE and DELETE are forbidden';
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS audit_logs_no_update_trigger ON audit_logs;
CREATE TRIGGER audit_logs_no_update_trigger
    BEFORE UPDATE OR DELETE ON audit_logs
    FOR EACH ROW
    EXECUTE FUNCTION audit_logs_no_update();

CREATE INDEX IF NOT EXISTS idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at DESC);

-- +goose Down
DROP TRIGGER IF EXISTS audit_logs_no_update_trigger ON audit_logs;
DROP TABLE IF EXISTS audit_logs;
