-- 宏观状态历史
CREATE TABLE IF NOT EXISTS macro_regime_history (
    time TIMESTAMPTZ NOT NULL,
    regime VARCHAR(32) NOT NULL,            -- INFLATIONARY/GOLDILOCKS/STAGFLATION/DEFLATION/STRESS/NEUTRAL
    score DOUBLE PRECISION,
    details JSONB,
    PRIMARY KEY (time)
);

SELECT create_hypertable('macro_regime_history', 'time', chunk_time_interval => INTERVAL '1 year', if_not_exists => TRUE);

CREATE INDEX IF NOT EXISTS idx_macro_regime ON macro_regime_history (regime, time DESC);

-- FRED 连续聚合（每日聚合）
CREATE MATERIALIZED VIEW IF NOT EXISTS fred_daily_agg
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 day', time) AS day,
    source,
    series_id,
    last(value_numeric, time) AS daily_value,
    count(*) AS observations
FROM data_snapshots
WHERE source = 'fred'
GROUP BY day, source, series_id;

SELECT add_continuous_aggregate_policy('fred_daily_agg',
    start_offset => INTERVAL '7 days',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour', if_not_exists => TRUE);

-- BIS 政策利率
CREATE TABLE IF NOT EXISTS bis_policy_rates (
    date DATE NOT NULL,
    currency VARCHAR(8) NOT NULL,
    rate DOUBLE PRECISION,
    PRIMARY KEY (date, currency)
);

-- BIS 信贷缺口
CREATE TABLE IF NOT EXISTS bis_credit_gap (
    date DATE NOT NULL,
    country VARCHAR(8) NOT NULL,
    gap_pct DOUBLE PRECISION,
    risk_label VARCHAR(16),                -- 'NORMAL','WARN','CRITICAL'
    PRIMARY KEY (date, country)
);

-- BIS REER
CREATE TABLE IF NOT EXISTS bis_reer (
    date DATE NOT NULL,
    currency VARCHAR(8) NOT NULL,
    reer DOUBLE PRECISION,
    z_score DOUBLE PRECISION,
    PRIMARY KEY (date, currency)
);

-- FedWatch 概率
CREATE TABLE IF NOT EXISTS fed_watch_probabilities (
    snapshot_at TIMESTAMPTZ NOT NULL,
    meeting_date DATE NOT NULL,
    rate_change_bps INT NOT NULL,         -- -50, -25, 0, +25, +50
    probability DOUBLE PRECISION,
    PRIMARY KEY (snapshot_at, meeting_date, rate_change_bps)
);

-- 宏观指标通用表
CREATE TABLE IF NOT EXISTS macro_indicators (
    source VARCHAR(32) NOT NULL,           -- 'ECB','OECD','EUROSTAT','SNB','DTCC','TE'
    indicator VARCHAR(64) NOT NULL,
    region VARCHAR(16),
    period DATE,
    value DOUBLE PRECISION,
    metadata JSONB,
    fetched_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (source, indicator, region, period)
);

-- Carry 利率差
CREATE TABLE IF NOT EXISTS carry_rates (
    date DATE NOT NULL,
    currency VARCHAR(8) NOT NULL,
    rate DOUBLE PRECISION,
    vs_usd_spread DOUBLE PRECISION,
    PRIMARY KEY (date, currency)
);
