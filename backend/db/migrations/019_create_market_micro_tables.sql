-- 跨市场相关性历史
CREATE TABLE intermarket_correlations (
    time TIMESTAMPTZ NOT NULL,
    pair_a VARCHAR(32),
    pair_b VARCHAR(32),
    window_days INT,
    correlation DOUBLE PRECISION,
    historical_mean DOUBLE PRECISION,
    historical_std DOUBLE PRECISION,
    z_score DOUBLE PRECISION,
    is_break BOOLEAN,
    PRIMARY KEY (time, pair_a, pair_b, window_days)
);

SELECT create_hypertable('intermarket_correlations', 'time', chunk_time_interval => INTERVAL '90 days');

-- 微观结构快照（高频，注意压缩）
CREATE TABLE micro_snapshots (
    time TIMESTAMPTZ NOT NULL,
    symbol VARCHAR(32) NOT NULL,
    obi_top10 DOUBLE PRECISION,
    spread_bps DOUBLE PRECISION,
    bid_depth DOUBLE PRECISION,
    ask_depth DOUBLE PRECISION,
    stress_score DOUBLE PRECISION,
    PRIMARY KEY (time, symbol)
);

SELECT create_hypertable('micro_snapshots', 'time', chunk_time_interval => INTERVAL '7 days');
SELECT add_compression_policy('micro_snapshots', INTERVAL '7 days');

-- Orderflow 事件
CREATE TABLE orderflow_absorptions (
    id BIGSERIAL PRIMARY KEY,
    time TIMESTAMPTZ NOT NULL,
    symbol VARCHAR(32),
    price DOUBLE PRECISION,
    direction VARCHAR(8),                   -- 'BUY','SELL'
    strength DOUBLE PRECISION,
    volume DOUBLE PRECISION
);

SELECT create_hypertable('orderflow_absorptions', 'time', chunk_time_interval => INTERVAL '30 days');

-- Volume Profile
CREATE TABLE volume_profiles (
    time TIMESTAMPTZ NOT NULL,
    symbol VARCHAR(32) NOT NULL,
    period VARCHAR(16),                    -- 'session','day','week'
    poc DOUBLE PRECISION,
    vah DOUBLE PRECISION,
    val DOUBLE PRECISION,
    profile JSONB,                          -- 价格-成交量分布
    PRIMARY KEY (time, symbol, period)
);

SELECT create_hypertable('volume_profiles', 'time', chunk_time_interval => INTERVAL '90 days');
