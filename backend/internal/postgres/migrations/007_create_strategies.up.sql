-- 创建策略管理表
CREATE TABLE IF NOT EXISTS strategies (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            TEXT NOT NULL,
    kind            TEXT NOT NULL,          -- 策略类型标识：ma_cross / rsi_reversal / cot_extreme / ...
    symbol          TEXT NOT NULL,          -- 标的：EURUSD / BTCUSDT / SPX / ...
    timeframe       TEXT NOT NULL,          -- 1h / 4h / 1d
    params          JSONB NOT NULL DEFAULT '{}'::jsonb,
    schedule_cron   TEXT NOT NULL DEFAULT '@hourly',  -- @hourly / @daily / 0 */4 * * *
    enabled         BOOLEAN NOT NULL DEFAULT FALSE,
    status          TEXT NOT NULL DEFAULT 'draft',    -- draft / active / archived
    description     TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by      TEXT NOT NULL DEFAULT '',
    updated_by      TEXT NOT NULL DEFAULT '',
    last_run_at     TIMESTAMPTZ,
    last_run_status TEXT,                   -- success / failed / running
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_strategies_enabled ON strategies(enabled) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_strategies_kind ON strategies(kind) WHERE deleted_at IS NULL;

-- 扩展 backtest_results 表（如果不存在）
ALTER TABLE backtest_results ADD COLUMN IF NOT EXISTS strategy_id UUID;
ALTER TABLE backtest_results ADD COLUMN IF NOT EXISTS params JSONB DEFAULT '{}'::jsonb;
ALTER TABLE backtest_results ADD COLUMN IF NOT EXISTS mock BOOLEAN DEFAULT FALSE;
