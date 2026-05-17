-- Migration: Add pg_trgm GIN indexes for ILIKE search acceleration
-- Run: psql -d antclaw -f db/migrations/20260517_search_index.sql
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX IF NOT EXISTS idx_users_display_name_trgm ON users USING GIN (display_name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_alfq_posts_content_trgm ON alfq_posts USING GIN (content gin_trgm_ops);
