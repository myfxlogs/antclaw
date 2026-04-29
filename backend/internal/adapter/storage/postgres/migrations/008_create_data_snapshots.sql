-- TimescaleDB 通用时间序列快照表
-- 用于存储所有外部数据源的原始数据

CREATE TABLE IF NOT EXISTS data_snapshots (
    time TIMESTAMPTZ NOT NULL,
    source VARCHAR(32) NOT NULL,      -- 'fred', 'mql5', 'cot', 'ecb', etc.
    series_id VARCHAR(64) NOT NULL,   -- 'GDP', 'UNRATE', 'EUR_NFP', etc.
    value_numeric DOUBLE PRECISION,
    value_text TEXT,                  -- 非数值数据
    raw_json JSONB,                   -- 原始 API 响应
    fetched_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (time, source, series_id)
);

-- 转换为 hypertable，按 30 天分区
SELECT create_hypertable('data_snapshots', 'time',
    chunk_time_interval => INTERVAL '30 days', if_not_exists => TRUE);

-- 启用压缩（30 天前的数据自动压缩）
ALTER TABLE data_snapshots SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'source, series_id'
);

SELECT add_compression_policy('data_snapshots', INTERVAL '30 days', if_not_exists => TRUE);

-- 索引
CREATE INDEX IF NOT EXISTS idx_snapshots_source_series ON data_snapshots (source, series_id, time DESC);
CREATE INDEX IF NOT EXISTS idx_snapshots_fetched ON data_snapshots (fetched_at DESC);
