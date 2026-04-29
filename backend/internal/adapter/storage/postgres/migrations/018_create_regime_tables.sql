-- 状态融合历史
CREATE TABLE IF NOT EXISTS regime_overlay_history (
    time TIMESTAMPTZ NOT NULL,
    symbol VARCHAR(32) NOT NULL,
    timeframe VARCHAR(8) NOT NULL,
    unified_score DOUBLE PRECISION,
    unified_label VARCHAR(16),             -- STRONG_BULL/BULL/NEUTRAL/BEAR/STRONG_BEAR
    hmm_state VARCHAR(16),
    hmm_confidence DOUBLE PRECISION,
    hmm_score DOUBLE PRECISION,
    garch_regime VARCHAR(16),
    vol_ratio DOUBLE PRECISION,
    garch_score DOUBLE PRECISION,
    adx_strength VARCHAR(16),
    adx_value DOUBLE PRECISION,
    adx_score DOUBLE PRECISION,
    cot_score DOUBLE PRECISION,
    available_models JSONB,
    PRIMARY KEY (time, symbol, timeframe)
);

SELECT create_hypertable('regime_overlay_history', 'time', chunk_time_interval => INTERVAL '90 days', if_not_exists => TRUE);

CREATE INDEX IF NOT EXISTS idx_regime_overlay_symbol ON regime_overlay_history (symbol, timeframe, time DESC);

-- 状态转换事件
CREATE TABLE IF NOT EXISTS regime_transitions (
    id BIGSERIAL,
    time TIMESTAMPTZ NOT NULL,
    symbol VARCHAR(32),
    timeframe VARCHAR(8),
    from_label VARCHAR(16),
    to_label VARCHAR(16),
    from_score DOUBLE PRECISION,
    to_score DOUBLE PRECISION,
    severity VARCHAR(8),                    -- 'INFO','WARN','CRITICAL'
    PRIMARY KEY (id, time)
);

SELECT create_hypertable('regime_transitions', 'time', chunk_time_interval => INTERVAL '90 days', if_not_exists => TRUE);
