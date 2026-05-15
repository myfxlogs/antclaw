-- Migration 031: alfq 社交表（幂等创建 + 扩展）
-- 处理两种场景：
--   1) 已有 alfq_* 表则仅 ALTER 加新字段/索引
--   2) 空数据库则先 CREATE TABLE IF NOT EXISTS 再加字段/索引
-- 对应文档：docs/安卓客户端技术文档包/09-服务端配套改造优化文档.md §4.2

-- ===== Feed =====
CREATE TABLE IF NOT EXISTS alfq_posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    author_id UUID NOT NULL REFERENCES users(id),
    author_name TEXT NOT NULL DEFAULT '',
    content TEXT NOT NULL DEFAULT '',
    post_type TEXT NOT NULL DEFAULT 'text',
    signal_pair TEXT NOT NULL DEFAULT '',
    signal_direction TEXT NOT NULL DEFAULT '',
    signal_confidence INT NOT NULL DEFAULT 0,
    visibility TEXT NOT NULL DEFAULT 'public',
    circle_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- alfq_posts 扩展字段（幂等）
ALTER TABLE alfq_posts ADD COLUMN IF NOT EXISTS original_post_id uuid NULL REFERENCES alfq_posts(id);

-- alfq_posts 索引（幂等）
CREATE INDEX IF NOT EXISTS idx_alfq_posts_feed_cursor   ON alfq_posts (created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_alfq_posts_author_cursor ON alfq_posts (author_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_alfq_posts_type_cursor   ON alfq_posts (post_type, created_at DESC, id DESC);

-- ===== Comments =====
CREATE TABLE IF NOT EXISTS alfq_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID NOT NULL REFERENCES alfq_posts(id) ON DELETE CASCADE,
    author_id UUID NOT NULL REFERENCES users(id),
    author_name TEXT NOT NULL DEFAULT '',
    content TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- alfq_comments 扩展字段（幂等）
ALTER TABLE alfq_comments ADD COLUMN IF NOT EXISTS parent_comment_id uuid NULL REFERENCES alfq_comments(id);

-- alfq_comments 索引（幂等）
CREATE INDEX IF NOT EXISTS idx_alfq_comments_post_cursor ON alfq_comments (post_id, created_at ASC, id ASC);
CREATE INDEX IF NOT EXISTS idx_alfq_comments_parent ON alfq_comments (parent_comment_id);

-- ===== Likes =====
CREATE TABLE IF NOT EXISTS alfq_likes (
    post_id UUID NOT NULL REFERENCES alfq_posts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (post_id, user_id)
);

-- ===== Follows =====
CREATE TABLE IF NOT EXISTS alfq_follows (
    follower_id UUID NOT NULL REFERENCES users(id),
    following_id UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (follower_id, following_id)
);
