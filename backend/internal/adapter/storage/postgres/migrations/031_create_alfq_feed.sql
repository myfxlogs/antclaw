-- 031: AlfQ Feed 社交帖子表
-- 对应 proto: antclaw/v1/alfq_feed.proto
CREATE TABLE IF NOT EXISTS alfq_posts (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    author_id         UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    author_name       TEXT         NOT NULL DEFAULT '',
    content           TEXT         NOT NULL DEFAULT '',
    post_type         VARCHAR(32)  NOT NULL DEFAULT 'text',  -- text / signal_card / chart_share / share
    signal_pair       VARCHAR(32)  NOT NULL DEFAULT '',
    signal_direction  VARCHAR(16)  NOT NULL DEFAULT '',
    signal_confidence INTEGER      NOT NULL DEFAULT 0,
    visibility        VARCHAR(16)  NOT NULL DEFAULT 'public', -- public / followers / circle
    circle_id         UUID,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_alfq_posts_created ON alfq_posts(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_alfq_posts_author ON alfq_posts(author_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_alfq_posts_visibility ON alfq_posts(visibility, created_at DESC);

-- 点赞表
CREATE TABLE IF NOT EXISTS alfq_likes (
    post_id    UUID NOT NULL REFERENCES alfq_posts(id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (post_id, user_id)
);

-- 评论表
CREATE TABLE IF NOT EXISTS alfq_comments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id     UUID         NOT NULL REFERENCES alfq_posts(id) ON DELETE CASCADE,
    author_id   UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    author_name TEXT         NOT NULL DEFAULT '',
    content     TEXT         NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_alfq_comments_post ON alfq_comments(post_id, created_at ASC);
