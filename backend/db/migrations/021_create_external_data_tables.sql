-- SEC 13F 持仓
CREATE TABLE sec_13f_holdings (
    institution_cik VARCHAR(16),
    quarter VARCHAR(8),                     -- '2026-Q1'
    issuer VARCHAR(256),
    cusip VARCHAR(16),
    value_usd BIGINT,
    shares BIGINT,
    fetched_at TIMESTAMPTZ,
    PRIMARY KEY (institution_cik, quarter, cusip)
);

-- 美国国债拍卖
CREATE TABLE treasury_auctions (
    cusip VARCHAR(16) PRIMARY KEY,
    security_type VARCHAR(16),
    security_term VARCHAR(16),
    auction_date DATE,
    high_yield DOUBLE PRECISION,
    bid_to_cover DOUBLE PRECISION,
    indirect_pct DOUBLE PRECISION,
    demand_label VARCHAR(16)               -- 'STRONG','NEUTRAL','WEAK'
);

CREATE INDEX idx_treasury_auctions_date ON treasury_auctions (auction_date DESC);

-- 世界银行宏观数据
CREATE TABLE worldbank_macro (
    country VARCHAR(8),
    currency VARCHAR(8),
    year INT,
    gdp_growth DOUBLE PRECISION,
    current_account DOUBLE PRECISION,
    cpi DOUBLE PRECISION,
    fx_reserves DOUBLE PRECISION,
    PRIMARY KEY (country, year)
);

-- IMF WEO 数据
CREATE TABLE imf_weo (
    country VARCHAR(8),
    indicator VARCHAR(32),                  -- 'GDP_GROWTH','CPI','CA'
    year INT,
    is_forecast BOOLEAN,
    value DOUBLE PRECISION,
    vintage VARCHAR(16),                    -- '2026-04','2026-10'
    PRIMARY KEY (country, indicator, year, vintage)
);

-- 采集任务审计表
CREATE TABLE fetch_jobs (
    id BIGSERIAL PRIMARY KEY,
    job_type VARCHAR(64) NOT NULL,          -- 'fred-fetch','mql5-micro'
    source VARCHAR(32) NOT NULL,
    started_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    status VARCHAR(16),                     -- 'running','success','failed'
    error_message TEXT,
    records_inserted INTEGER DEFAULT 0,
    records_updated INTEGER DEFAULT 0,
    duration_ms INTEGER
);

SELECT create_hypertable('fetch_jobs', 'started_at', chunk_time_interval => INTERVAL '7 days');

CREATE INDEX idx_fetch_jobs_source ON fetch_jobs (source, started_at DESC);
CREATE INDEX idx_fetch_jobs_status ON fetch_jobs (status, started_at DESC);
