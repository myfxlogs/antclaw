-- TA 信号历史
CREATE TABLE IF NOT EXISTS ta_signals (
    id BIGSERIAL,
    symbol VARCHAR(32),
    timeframe VARCHAR(8),                  -- 'D','4H','1H'
    issued_at TIMESTAMPTZ,
    signal_type VARCHAR(64),               -- 'FVG_LONG','OB_SHORT','WYCKOFF_SPRING'...
    grade VARCHAR(2),                      -- 'A','B','C'
    direction VARCHAR(8),                  -- 'LONG','SHORT'
    entry_price DOUBLE PRECISION,
    stop_loss DOUBLE PRECISION,
    take_profit DOUBLE PRECISION,
    metadata JSONB,
    PRIMARY KEY (id, issued_at)
);

SELECT create_hypertable('ta_signals', 'issued_at', chunk_time_interval => INTERVAL '90 days', if_not_exists => TRUE);

CREATE INDEX IF NOT EXISTS idx_ta_signals_symbol ON ta_signals (symbol, timeframe, issued_at DESC);
CREATE INDEX IF NOT EXISTS idx_ta_signals_type ON ta_signals (signal_type, issued_at DESC);

-- ICT 结构持久化
CREATE TABLE IF NOT EXISTS ict_structures (
    id BIGSERIAL,
    symbol VARCHAR(32),
    timeframe VARCHAR(8),
    type VARCHAR(16),                      -- 'FVG','OB','BREAKER','LIQUIDITY','BOS','CHOCH','SWEEP'
    formed_at TIMESTAMPTZ,
    high DOUBLE PRECISION,
    low DOUBLE PRECISION,
    direction VARCHAR(8),
    mitigated BOOLEAN DEFAULT FALSE,
    mitigated_at TIMESTAMPTZ,
    reversed BOOLEAN DEFAULT FALSE,
    raw JSONB,
    PRIMARY KEY (id, formed_at)
);

SELECT create_hypertable('ict_structures', 'formed_at', chunk_time_interval => INTERVAL '90 days', if_not_exists => TRUE);

CREATE INDEX IF NOT EXISTS idx_ict_structures_active ON ict_structures (symbol, timeframe, type, mitigated);

-- Elliott 浪型历史
CREATE TABLE IF NOT EXISTS elliott_counts (
    id BIGSERIAL PRIMARY KEY,
    symbol VARCHAR(32),
    timeframe VARCHAR(8),
    detected_at TIMESTAMPTZ,
    wave_type VARCHAR(16),                 -- IMPULSE/CORRECTIVE/DIAGONAL...
    waves JSONB,
    valid BOOLEAN,
    targets JSONB,
    confidence DOUBLE PRECISION
);

-- Wyckoff 事件
CREATE TABLE IF NOT EXISTS wyckoff_events (
    id BIGSERIAL,
    symbol VARCHAR(32),
    timeframe VARCHAR(8),
    event_name VARCHAR(16),                -- PS,SC,AR,ST,SPRING,...
    bar_time TIMESTAMPTZ,
    price DOUBLE PRECISION,
    volume DOUBLE PRECISION,
    confidence DOUBLE PRECISION,
    raw JSONB,
    PRIMARY KEY (id, bar_time)
);

SELECT create_hypertable('wyckoff_events', 'bar_time', chunk_time_interval => INTERVAL '90 days', if_not_exists => TRUE);

-- Wyckoff 阶段快照
CREATE TABLE IF NOT EXISTS wyckoff_phases (
    time TIMESTAMPTZ NOT NULL,
    symbol VARCHAR(32) NOT NULL,
    timeframe VARCHAR(8) NOT NULL,
    phase VARCHAR(16),                     -- ACCUMULATION/DISTRIBUTION/TRANSITION
    confidence DOUBLE PRECISION,
    cause_score DOUBLE PRECISION,
    PRIMARY KEY (time, symbol, timeframe)
);
