-- 情绪快照
CREATE TABLE IF NOT EXISTS sentiment_snapshots (
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

SELECT create_hypertable('sentiment_snapshots', 'time', chunk_time_interval => INTERVAL '30 days', if_not_exists => TRUE);

-- 链上指标（长表：单点拆为多行，便于扩展指标维度）
-- 与 worker collector_onchain 及 OnchainHandler 的实际读写对齐。
CREATE TABLE IF NOT EXISTS onchain_metrics (
    time   TIMESTAMPTZ        NOT NULL,
    asset  VARCHAR(16)        NOT NULL,    -- 'BTC','ETH'
    metric VARCHAR(32)        NOT NULL,    -- 'active_addresses','tx_count','net_flow','mvrv','sopr','onchain_score',...
    value  DOUBLE PRECISION   NOT NULL,
    source VARCHAR(32),                    -- 'coingecko','coinmetrics',...
    PRIMARY KEY (time, asset, metric)
);

SELECT create_hypertable('onchain_metrics', 'time', chunk_time_interval => INTERVAL '30 days', if_not_exists => TRUE);

CREATE INDEX IF NOT EXISTS idx_onchain_asset_metric ON onchain_metrics (asset, metric, time DESC);

-- DeFi 快照
CREATE TABLE IF NOT EXISTS defi_snapshots (
    time TIMESTAMPTZ PRIMARY KEY,
    total_tvl DOUBLE PRECISION,
    tvl_change_24h DOUBLE PRECISION,
    tvl_change_7d DOUBLE PRECISION,
    dex_vol_24h DOUBLE PRECISION,
    stablecoin_mc DOUBLE PRECISION,
    raw JSONB
);

SELECT create_hypertable('defi_snapshots', 'time', chunk_time_interval => INTERVAL '30 days', if_not_exists => TRUE);
