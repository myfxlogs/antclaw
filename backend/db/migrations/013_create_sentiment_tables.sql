-- 情绪快照
CREATE TABLE sentiment_snapshots (
    time TIMESTAMPTZ PRIMARY KEY,
    score DOUBLE PRECISION,                -- -100..+100
    regime VARCHAR(16),                    -- COMPLACENCY/NORMAL/STRESS/PANIC
    pc_ratio DOUBLE PRECISION,
    pc_percentile DOUBLE PRECISION,
    fear_greed DOUBLE PRECISION,
    retail_long_pct DOUBLE PRECISION,
    insider_net DOUBLE PRECISION,
    raw JSONB
);

SELECT create_hypertable('sentiment_snapshots', 'time', chunk_time_interval => INTERVAL '30 days');

-- 链上指标
CREATE TABLE onchain_metrics (
    date DATE NOT NULL,
    asset VARCHAR(8) NOT NULL,             -- 'BTC','ETH'
    flow_in DOUBLE PRECISION,
    flow_out DOUBLE PRECISION,
    net_flow DOUBLE PRECISION,
    active_addr BIGINT,
    tx_count BIGINT,
    onchain_score DOUBLE PRECISION,
    PRIMARY KEY (date, asset)
);

SELECT create_hypertable('onchain_metrics', 'date', chunk_time_interval => INTERVAL '180 days');

-- DeFi 快照
CREATE TABLE defi_snapshots (
    time TIMESTAMPTZ PRIMARY KEY,
    total_tvl DOUBLE PRECISION,
    tvl_change_24h DOUBLE PRECISION,
    tvl_change_7d DOUBLE PRECISION,
    dex_vol_24h DOUBLE PRECISION,
    stablecoin_mc DOUBLE PRECISION,
    raw JSONB
);

SELECT create_hypertable('defi_snapshots', 'time', chunk_time_interval => INTERVAL '30 days');
