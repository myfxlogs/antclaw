-- 回测任务
CREATE TABLE backtest_jobs (
    job_id UUID PRIMARY KEY,
    user_id VARCHAR(64),
    type VARCHAR(32),                       -- 'evaluator','walkforward','montecarlo','composer'
    request JSONB,
    status VARCHAR(16),                     -- 'queued','running','done','failed','canceled'
    progress DOUBLE PRECISION,
    created_at TIMESTAMPTZ,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    error TEXT
);

SELECT create_hypertable('backtest_jobs', 'created_at', chunk_time_interval => INTERVAL '30 days');

CREATE INDEX idx_backtest_jobs_user ON backtest_jobs (user_id, status, created_at DESC);
CREATE INDEX idx_backtest_jobs_status ON backtest_jobs (status, created_at DESC);

-- 回测结果
CREATE TABLE backtest_results (
    job_id UUID PRIMARY KEY REFERENCES backtest_jobs(job_id),
    summary JSONB,                          -- 关键指标
    equity_curve JSONB,
    trades JSONB,
    detailed JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 信号校准参数
CREATE TABLE signal_calibration (
    signal_type VARCHAR(64) PRIMARY KEY,
    logistic_a DOUBLE PRECISION,
    logistic_b DOUBLE PRECISION,
    sample_size INT,
    win_rate DOUBLE PRECISION,
    avg_return DOUBLE PRECISION,
    updated_at TIMESTAMPTZ
);

-- Walk-forward 优化历史
CREATE TABLE walkforward_history (
    id BIGSERIAL PRIMARY KEY,
    fold_idx INT,
    train_from DATE,
    train_to DATE,
    test_from DATE,
    test_to DATE,
    optimal_weights JSONB,
    in_sample_sharpe DOUBLE PRECISION,
    oos_sharpe DOUBLE PRECISION,
    created_at TIMESTAMPTZ
);
