-- 统一信号历史
CREATE TABLE unified_signals (
    id BIGSERIAL PRIMARY KEY,
    symbol VARCHAR(32) NOT NULL,
    issued_at TIMESTAMPTZ NOT NULL,
    recommendation VARCHAR(16),              -- STRONG_LONG/LONG/NEUTRAL/SHORT/STRONG_SHORT
    unified_score DOUBLE PRECISION,          -- -1..+1
    confidence DOUBLE PRECISION,
    components JSONB,
    missing_subsys TEXT[],
    weights_used JSONB
);

SELECT create_hypertable('unified_signals', 'issued_at', chunk_time_interval => INTERVAL '90 days');

CREATE INDEX idx_unified_signals_symbol ON unified_signals (symbol, issued_at DESC);

-- 信号准确率统计
CREATE TABLE signal_outcomes (
    signal_id BIGINT REFERENCES unified_signals(id),
    horizon VARCHAR(8),                      -- '1D','1W','2W','1M'
    return_pct DOUBLE PRECISION,
    direction_match BOOLEAN,
    evaluated_at TIMESTAMPTZ,
    PRIMARY KEY (signal_id, horizon)
);

-- 权重配置
CREATE TABLE signal_weight_config (
    id SERIAL PRIMARY KEY,
    name VARCHAR(64) UNIQUE,                 -- 'default','aggressive','conservative'
    weights JSONB,
    is_active BOOLEAN,
    updated_at TIMESTAMPTZ
);

-- Playbooks
CREATE TABLE playbooks (
    id BIGSERIAL PRIMARY KEY,
    generated_at TIMESTAMPTZ NOT NULL,
    user_id VARCHAR(64),
    regime VARCHAR(32),
    entries JSONB,
    global_risk JSONB,
    weights JSONB
);

SELECT create_hypertable('playbooks', 'generated_at', chunk_time_interval => INTERVAL '30 days');

CREATE INDEX idx_playbooks_user ON playbooks (user_id, generated_at DESC);

-- Playbook 决策
CREATE TABLE playbook_decisions (
    id BIGSERIAL PRIMARY KEY,
    playbook_id BIGINT REFERENCES playbooks(id),
    user_id VARCHAR(64),
    symbol VARCHAR(32),
    decision VARCHAR(16),                  -- 'TAKEN','SKIPPED','MODIFIED'
    actual_entry DOUBLE PRECISION,
    actual_stop DOUBLE PRECISION,
    notes TEXT,
    decided_at TIMESTAMPTZ
);

-- 因子排名快照
CREATE TABLE factor_rankings (
    time TIMESTAMPTZ NOT NULL,
    snapshot_id BIGSERIAL,
    weights JSONB,
    PRIMARY KEY (time, snapshot_id)
);

SELECT create_hypertable('factor_rankings', 'time', chunk_time_interval => INTERVAL '30 days');

CREATE TABLE factor_ranking_entries (
    snapshot_id BIGINT,
    symbol VARCHAR(32),
    rank INT,
    raw_score DOUBLE PRECISION,
    norm_score DOUBLE PRECISION,
    breakdown JSONB,
    PRIMARY KEY (snapshot_id, symbol)
);

-- Flow divergence 历史
CREATE TABLE flow_divergence_history (
    time TIMESTAMPTZ NOT NULL,
    pair_a VARCHAR(32),
    pair_b VARCHAR(32),
    corr DOUBLE PRECISION,
    baseline_mean DOUBLE PRECISION,
    baseline_std DOUBLE PRECISION,
    z_score DOUBLE PRECISION,
    lead_lag INT,
    PRIMARY KEY (time, pair_a, pair_b)
);

SELECT create_hypertable('flow_divergence_history', 'time', chunk_time_interval => INTERVAL '30 days');
