-- Migration: Content governance tables
-- Run: cd backend && psql -d antclaw -f db/migrations/20260516_content_governance.sql

-- 举报表
CREATE TABLE IF NOT EXISTS alfq_reports (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id     UUID NOT NULL REFERENCES alfq_posts(id) ON DELETE CASCADE,
    reporter_id UUID NOT NULL REFERENCES users(id),
    reason      VARCHAR(100) NOT NULL,
    details     TEXT,
    status      VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);
CREATE INDEX idx_alfq_reports_post_status ON alfq_reports(post_id, status);

-- 用户信誉分
CREATE TABLE IF NOT EXISTS user_cred (
    user_id    UUID PRIMARY KEY REFERENCES users(id),
    score      REAL NOT NULL DEFAULT 0,
    page_rank  REAL NOT NULL DEFAULT 0,
    inter_rate REAL NOT NULL DEFAULT 0,
    sig_acc    REAL NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
