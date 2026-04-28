-- COT 原始记录 hypertable
CREATE TABLE cot_records (
    report_date DATE NOT NULL,
    contract_code VARCHAR(16) NOT NULL,
    currency VARCHAR(8),
    noncomm_long BIGINT,
    noncomm_short BIGINT,
    comm_long BIGINT,
    comm_short BIGINT,
    dealer_long BIGINT,
    dealer_short BIGINT,
    levfund_long BIGINT,
    levfund_short BIGINT,
    mm_long BIGINT,
    mm_short BIGINT,
    swap_long BIGINT,
    swap_short BIGINT,
    total_oi BIGINT,
    raw_json JSONB,
    PRIMARY KEY (report_date, contract_code)
);

SELECT create_hypertable('cot_records', 'report_date', chunk_time_interval => INTERVAL '1 year');

CREATE INDEX idx_cot_currency ON cot_records (currency, report_date DESC);
CREATE INDEX idx_cot_contract ON cot_records (contract_code, report_date DESC);

-- COT 分析结果缓存（每周覆盖）
CREATE TABLE cot_analyses (
    report_date DATE NOT NULL,
    contract_code VARCHAR(16) NOT NULL,
    net_position BIGINT,
    cot_index DOUBLE PRECISION,
    direction VARCHAR(16),                  -- BULLISH/BEARISH/NEUTRAL
    sentiment_score DOUBLE PRECISION,
    wow_change BIGINT,
    zscore DOUBLE PRECISION,
    percentile DOUBLE PRECISION,
    PRIMARY KEY (report_date, contract_code)
);

CREATE INDEX idx_cot_analysis_contract ON cot_analyses (contract_code, report_date DESC);
CREATE INDEX idx_cot_analysis_direction ON cot_analyses (direction, report_date DESC);

-- 信号回测记录（用于 Platt 校准）
CREATE TABLE cot_signal_outcomes (
    signal_id BIGSERIAL PRIMARY KEY,
    signal_type VARCHAR(32) NOT NULL,
    contract_code VARCHAR(16) NOT NULL,
    issued_at TIMESTAMPTZ NOT NULL,
    raw_confidence DOUBLE PRECISION,
    return_1w DOUBLE PRECISION,
    return_2w DOUBLE PRECISION,
    return_4w DOUBLE PRECISION,
    win BOOLEAN,
    evaluated_at TIMESTAMPTZ
);

CREATE INDEX idx_signal_type_outcome ON cot_signal_outcomes (signal_type, win);
CREATE INDEX idx_signal_contract ON cot_signal_outcomes (contract_code, issued_at DESC);

-- 信号校准参数
CREATE TABLE cot_calibration (
    signal_type VARCHAR(32) PRIMARY KEY,
    platt_a DOUBLE PRECISION,
    platt_b DOUBLE PRECISION,
    win_rate DOUBLE PRECISION,
    sample_size INTEGER,
    updated_at TIMESTAMPTZ
);
