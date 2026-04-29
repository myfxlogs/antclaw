-- 财经事件表
CREATE TABLE IF NOT EXISTS calendar_events (
    event_id VARCHAR(64) PRIMARY KEY,        -- 'mql5-12345'
    title VARCHAR(256) NOT NULL,
    country VARCHAR(8),
    currency VARCHAR(8),
    impact VARCHAR(16),                      -- 'low', 'medium', 'high'
    scheduled_at TIMESTAMPTZ NOT NULL,
    previous_value TEXT,
    forecast_value TEXT,
    actual_value TEXT,
    impact_direction SMALLINT,              -- 0/1/2
    surprise_score DOUBLE PRECISION,        -- stddev-normalized
    surprise_label VARCHAR(32),              -- 'MAJOR_BEAT' etc.
    revision_label VARCHAR(32),
    fetched_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_events_scheduled ON calendar_events (scheduled_at);
CREATE INDEX IF NOT EXISTS idx_events_currency_impact ON calendar_events (currency, impact, scheduled_at DESC);
CREATE INDEX IF NOT EXISTS idx_events_updated ON calendar_events (updated_at DESC);

-- 历史 surprise 表（for stddev normalization）
CREATE TABLE IF NOT EXISTS calendar_surprise_history (
    id BIGSERIAL PRIMARY KEY,
    event_name VARCHAR(256),
    currency VARCHAR(8),
    released_at TIMESTAMPTZ,
    actual_val DOUBLE PRECISION,
    forecast_val DOUBLE PRECISION,
    diff DOUBLE PRECISION,                  -- actual - forecast
    sigma DOUBLE PRECISION                  -- normalized
);

CREATE INDEX IF NOT EXISTS idx_surprise_event_currency ON calendar_surprise_history (event_name, currency, released_at DESC);

-- 价格 impact 记录
CREATE TABLE IF NOT EXISTS event_impact_records (
    event_id VARCHAR(64),
    "window" VARCHAR(8),                      -- '15m','30m','1h','4h'
    symbol VARCHAR(32),
    price_before DOUBLE PRECISION,
    price_after DOUBLE PRECISION,
    pct_change DOUBLE PRECISION,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- TimescaleDB 要求分区列必须出现在 UNIQUE/PRIMARY KEY 中。
    PRIMARY KEY (event_id, "window", symbol, recorded_at)
);

SELECT create_hypertable('event_impact_records', 'recorded_at', chunk_time_interval => INTERVAL '90 days', if_not_exists => TRUE);

CREATE INDEX IF NOT EXISTS idx_impact_event ON event_impact_records (event_id, recorded_at DESC);
CREATE INDEX IF NOT EXISTS idx_impact_symbol ON event_impact_records (symbol, recorded_at DESC);

-- Fed 讲话
CREATE TABLE IF NOT EXISTS fed_speeches (
    guid VARCHAR(256) PRIMARY KEY,
    title VARCHAR(512),
    speaker VARCHAR(128),
    published_at TIMESTAMPTZ,
    url TEXT,
    content TEXT,
    tone VARCHAR(16),                       -- HAWKISH/DOVISH/NEUTRAL
    fetched_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_speeches_published ON fed_speeches (published_at DESC);
CREATE INDEX IF NOT EXISTS idx_speeches_tone ON fed_speeches (tone, published_at DESC);

-- 多语言标题表
CREATE TABLE IF NOT EXISTS calendar_event_titles (
    event_id VARCHAR(64) NOT NULL,
    lang VARCHAR(8) NOT NULL,               -- 'en','zh','ja','ru',...
    title VARCHAR(512) NOT NULL,
    description TEXT,
    source VARCHAR(16) NOT NULL,              -- 'mql5' | 'llm' | 'manual'
    confidence DOUBLE PRECISION,            -- LLM 翻译置信度（0..1）
    fetched_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (event_id, lang)
);

CREATE INDEX IF NOT EXISTS idx_titles_lang ON calendar_event_titles (lang, source);
CREATE INDEX IF NOT EXISTS idx_titles_event ON calendar_event_titles (event_id);
