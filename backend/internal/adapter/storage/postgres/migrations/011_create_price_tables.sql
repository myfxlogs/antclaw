-- 日线 hypertable
CREATE TABLE IF NOT EXISTS price_daily (
    time TIMESTAMPTZ NOT NULL,
    symbol VARCHAR(32) NOT NULL,
    open DOUBLE PRECISION,
    high DOUBLE PRECISION,
    low DOUBLE PRECISION,
    close DOUBLE PRECISION,
    volume DOUBLE PRECISION,
    source VARCHAR(16),                     -- 'twelvedata' | 'yahoo' | ...
    PRIMARY KEY (time, symbol)
);

SELECT create_hypertable('price_daily', 'time', chunk_time_interval => INTERVAL '90 days', if_not_exists => TRUE);
ALTER TABLE price_daily SET (timescaledb.compress, timescaledb.compress_segmentby = 'symbol');
SELECT add_compression_policy('price_daily', INTERVAL '90 days', if_not_exists => TRUE);

CREATE INDEX IF NOT EXISTS idx_price_daily_symbol ON price_daily (symbol, time DESC);
CREATE INDEX IF NOT EXISTS idx_price_daily_source ON price_daily (source, time DESC);

-- 盘中 hypertable
CREATE TABLE IF NOT EXISTS price_intraday (
    time TIMESTAMPTZ NOT NULL,
    symbol VARCHAR(32) NOT NULL,
    interval VARCHAR(8) NOT NULL,           -- '1h', '4h', '15m'
    open DOUBLE PRECISION,
    high DOUBLE PRECISION,
    low DOUBLE PRECISION,
    close DOUBLE PRECISION,
    volume DOUBLE PRECISION,
    source VARCHAR(16),
    PRIMARY KEY (time, symbol, interval)
);

SELECT create_hypertable('price_intraday', 'time', chunk_time_interval => INTERVAL '7 days', if_not_exists => TRUE);
ALTER TABLE price_intraday SET (timescaledb.compress, timescaledb.compress_segmentby = 'symbol');
SELECT add_compression_policy('price_intraday', INTERVAL '30 days', if_not_exists => TRUE);

CREATE INDEX IF NOT EXISTS idx_price_intraday_symbol ON price_intraday (symbol, interval, time DESC);

-- 连续聚合：日线 → 周线 / 月线
CREATE MATERIALIZED VIEW IF NOT EXISTS price_weekly WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('1 week', time) AS week,
    symbol,
    first(open, time) AS open,
    max(high) AS high,
    min(low) AS low,
    last(close, time) AS close,
    sum(volume) AS volume
FROM price_daily
GROUP BY week, symbol;

SELECT add_continuous_aggregate_policy('price_weekly',
    start_offset => INTERVAL '1 year',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour', if_not_exists => TRUE);
