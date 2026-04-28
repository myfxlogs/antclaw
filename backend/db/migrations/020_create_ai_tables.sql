-- AI 调用日志
CREATE TABLE ai_usage (
    id BIGSERIAL PRIMARY KEY,
    user_id VARCHAR(64),
    model VARCHAR(64),
    operation VARCHAR(32),                  -- 'chat','interpret','outlook'
    prompt_tokens INT,
    completion_tokens INT,
    cached BOOLEAN,
    duration_ms INT,
    error TEXT,
    created_at TIMESTAMPTZ
);

SELECT create_hypertable('ai_usage', 'created_at', chunk_time_interval => INTERVAL '30 days');

CREATE INDEX idx_ai_usage_user ON ai_usage (user_id, created_at DESC);
CREATE INDEX idx_ai_usage_model ON ai_usage (model, created_at DESC);

-- 用户记忆
CREATE TABLE ai_memory (
    user_id VARCHAR(64),
    path VARCHAR(512),
    content TEXT,
    updated_at TIMESTAMPTZ,
    PRIMARY KEY (user_id, path)
);

-- AI 缓存
CREATE TABLE ai_cache (
    fingerprint VARCHAR(64) PRIMARY KEY,    -- SHA-256 of input
    operation VARCHAR(32),
    model VARCHAR(64),
    result TEXT,
    created_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ
);

CREATE INDEX idx_ai_cache_expires ON ai_cache (expires_at);
CREATE INDEX idx_ai_cache_operation ON ai_cache (operation, created_at DESC);
