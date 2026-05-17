-- Migration: Moderation cases + content status columns (A13-P0-04, A13-P0-05)
-- Run: cd backend && psql -d antclaw -f db/migrations/20260517_moderation.sql

CREATE TABLE IF NOT EXISTS moderation_cases (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source       VARCHAR(20) NOT NULL DEFAULT 'report',  -- report/auto/manual
    target_type  VARCHAR(20) NOT NULL,                    -- post/comment/user
    target_id    UUID NOT NULL,
    reporter_id  UUID,
    reason       TEXT NOT NULL,
    priority     VARCHAR(10) NOT NULL DEFAULT 'normal',   -- low/normal/high/critical
    status       VARCHAR(20) NOT NULL DEFAULT 'pending',  -- pending/in_review/actioned/rejected/appealed/closed
    assignee_id  UUID,
    reviewer_id  UUID,
    notes        TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_moderation_status ON moderation_cases(status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_moderation_assignee ON moderation_cases(assignee_id);

-- Content status tracking
ALTER TABLE alfq_posts ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'active';
ALTER TABLE alfq_comments ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'active';
CREATE INDEX IF NOT EXISTS idx_alfq_posts_status ON alfq_posts(status);

-- Update report table to link to moderation_cases
ALTER TABLE alfq_reports ADD COLUMN IF NOT EXISTS case_id UUID REFERENCES moderation_cases(id);
ALTER TABLE alfq_reports ADD COLUMN IF NOT EXISTS priority VARCHAR(10) NOT NULL DEFAULT 'normal';
ALTER TABLE alfq_reports ADD COLUMN IF NOT EXISTS assignee_id UUID;
ALTER TABLE alfq_reports ADD COLUMN IF NOT EXISTS resolved_by UUID;
ALTER TABLE alfq_reports ADD COLUMN IF NOT EXISTS resolution_note TEXT;
