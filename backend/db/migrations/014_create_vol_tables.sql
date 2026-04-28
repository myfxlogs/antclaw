-- VIX 时间序列
CREATE TABLE vix_term_structure (
    time TIMESTAMPTZ PRIMARY KEY,
    spot DOUBLE PRECISION,
    m1 DOUBLE PRECISION,
    m2 DOUBLE PRECISION,
    m3 DOUBLE PRECISION,
    vvix DOUBLE PRECISION,
    vix9d DOUBLE PRECISION,
    vix3m DOUBLE PRECISION,
    skew DOUBLE PRECISION,
    ovx DOUBLE PRECISION,
    gvz DOUBLE PRECISION,
    rvx DOUBLE PRECISION,
    move DOUBLE PRECISION,
    contango BOOLEAN,
    cross_regime VARCHAR(32),
    raw JSONB
);

SELECT create_hypertable('vix_term_structure', 'time', chunk_time_interval => INTERVAL '90 days');

-- DVOL 快照
CREATE TABLE dvol_snapshots (
    time TIMESTAMPTZ NOT NULL,
    currency VARCHAR(8) NOT NULL,          -- 'BTC','ETH'
    current_iv DOUBLE PRECISION,
    change_24h_pct DOUBLE PRECISION,
    iv_hv_spread DOUBLE PRECISION,
    iv_hv_ratio DOUBLE PRECISION,
    spike BOOLEAN,
    PRIMARY KEY (time, currency)
);

SELECT create_hypertable('dvol_snapshots', 'time', chunk_time_interval => INTERVAL '30 days');

-- GEX 快照
CREATE TABLE gex_snapshots (
    time TIMESTAMPTZ NOT NULL,
    symbol VARCHAR(8) NOT NULL,
    spot_price DOUBLE PRECISION,
    total_gex DOUBLE PRECISION,
    flip_level DOUBLE PRECISION,
    max_call_wall DOUBLE PRECISION,
    max_put_wall DOUBLE PRECISION,
    levels JSONB,
    PRIMARY KEY (time, symbol)
);

SELECT create_hypertable('gex_snapshots', 'time', chunk_time_interval => INTERVAL '30 days');

-- IV 曲面与 Skew 历史
CREATE TABLE iv_skew_history (
    time TIMESTAMPTZ NOT NULL,
    symbol VARCHAR(8) NOT NULL,
    pc_iv_ratio DOUBLE PRECISION,
    skew_slope DOUBLE PRECISION,
    smile JSONB,                            -- 5-point moneyness IV
    term_slope DOUBLE PRECISION,
    flip_event VARCHAR(32),                 -- 'BEAR_TO_BULL','BULL_TO_BEAR' or NULL
    PRIMARY KEY (time, symbol)
);

SELECT create_hypertable('iv_skew_history', 'time', chunk_time_interval => INTERVAL '90 days');
