-- Migration: Post stats table for S12-P1-01
-- Replaces per-post subquery counts with a dedicated aggregated table.
-- Run: cd backend && psql -d antclaw -f db/migrations/20260517_post_stats.sql

CREATE TABLE IF NOT EXISTS alfq_post_stats (
    post_id       UUID PRIMARY KEY REFERENCES alfq_posts(id) ON DELETE CASCADE,
    like_count    INT4 NOT NULL DEFAULT 0,
    comment_count INT4 NOT NULL DEFAULT 0,
    share_count   INT4 NOT NULL DEFAULT 0,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Backfill existing posts
INSERT INTO alfq_post_stats (post_id, like_count, comment_count, share_count)
SELECT id,
    (SELECT COUNT(*) FROM alfq_likes WHERE post_id = alfq_posts.id),
    (SELECT COUNT(*) FROM alfq_comments WHERE post_id = alfq_posts.id),
    (SELECT COUNT(*) FROM alfq_posts sp WHERE sp.original_post_id = alfq_posts.id AND sp.post_type = 'share')
FROM alfq_posts
ON CONFLICT (post_id) DO NOTHING;
