-- M-B 回测产物：交易明细、状态分层指标
CREATE TABLE IF NOT EXISTS backtest_trades (
    job_id TEXT NOT NULL,
    seq INT NOT NULL,
    opened_at TIMESTAMPTZ,
    closed_at TIMESTAMPTZ,
    side TEXT NOT NULL,
    entry DOUBLE PRECISION,
    exit DOUBLE PRECISION,
    pnl DOUBLE PRECISION,
    pnl_pct DOUBLE PRECISION,
    mfe DOUBLE PRECISION,
    mae DOUBLE PRECISION,
    cost DOUBLE PRECISION,
    regime TEXT,
    PRIMARY KEY (job_id, seq)
);

CREATE TABLE IF NOT EXISTS backtest_metrics_by_regime (
    job_id TEXT NOT NULL,
    regime TEXT NOT NULL,
    n_trades INT,
    sharpe DOUBLE PRECISION,
    sortino DOUBLE PRECISION,
    max_drawdown DOUBLE PRECISION,
    win_rate DOUBLE PRECISION,
    PRIMARY KEY (job_id, regime)
);

-- M-C 通用校准（除 signal_calibration 外的细分模型）
CREATE TABLE IF NOT EXISTS signal_calibrations (
    model_id TEXT PRIMARY KEY,
    type TEXT NOT NULL,          -- 'platt' / 'isotonic'
    params JSONB NOT NULL,
    n_samples INT,
    brier DOUBLE PRECISION,
    fitted_at TIMESTAMPTZ DEFAULT NOW()
);

-- M-E 用户偏好与配额
CREATE TABLE IF NOT EXISTS user_preferences (
    user_id TEXT PRIMARY KEY,
    pairs TEXT[] DEFAULT '{}',
    high_impact_only BOOLEAN DEFAULT FALSE,
    quiet_hours_start INT,
    quiet_hours_end INT,
    timezone TEXT DEFAULT 'UTC'
);

CREATE TABLE IF NOT EXISTS user_quotas (
    user_id TEXT PRIMARY KEY,
    tier TEXT NOT NULL DEFAULT 'free',
    ai_calls_today INT DEFAULT 0,
    ai_max_per_day INT DEFAULT 20,
    reset_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS alert_log (
    id BIGSERIAL PRIMARY KEY,
    user_id TEXT,
    alert_type TEXT,
    severity TEXT,
    payload JSONB,
    sent BOOLEAN,
    reason TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_alert_log_user_time ON alert_log(user_id, created_at DESC);

-- M-F AI 记忆与多轮会话
CREATE TABLE IF NOT EXISTS ai_memories (
    id TEXT PRIMARY KEY,
    user_id TEXT,
    scope TEXT,                          -- 'global' / 'thread'
    key TEXT,
    value TEXT,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_ai_memories_user_scope_key
    ON ai_memories(user_id, scope, key);

CREATE TABLE IF NOT EXISTS ai_conversations (
    thread_id TEXT PRIMARY KEY,
    user_id TEXT,
    started_at TIMESTAMPTZ DEFAULT NOW(),
    last_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ai_messages (
    thread_id TEXT REFERENCES ai_conversations(thread_id) ON DELETE CASCADE,
    seq INT,
    role TEXT,                           -- 'user' / 'assistant' / 'tool' / 'system'
    content TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (thread_id, seq)
);
