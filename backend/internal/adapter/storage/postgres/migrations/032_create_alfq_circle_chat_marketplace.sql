-- 032: AlfQ 圈子、关注、聊天、市场表
-- handler 引用但迁移遗漏的表集合

-- 圈子
CREATE TABLE IF NOT EXISTS alfq_circles (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name         TEXT        NOT NULL DEFAULT '',
    description  TEXT        NOT NULL DEFAULT '',
    symbol       VARCHAR(32) NOT NULL DEFAULT '',
    created_by   UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    member_count INTEGER     NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_alfq_circles_members ON alfq_circles(member_count DESC);

-- 圈子成员
CREATE TABLE IF NOT EXISTS alfq_circle_members (
    circle_id UUID NOT NULL REFERENCES alfq_circles(id) ON DELETE CASCADE,
    user_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (circle_id, user_id)
);

-- 关注
CREATE TABLE IF NOT EXISTS alfq_follows (
    follower_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    following_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (follower_id, following_id)
);
CREATE INDEX IF NOT EXISTS idx_alfq_follows_following ON alfq_follows(following_id);
CREATE INDEX IF NOT EXISTS idx_alfq_follows_follower ON alfq_follows(follower_id);

-- 聊天消息
CREATE TABLE IF NOT EXISTS alfq_messages (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id VARCHAR(64) NOT NULL DEFAULT '',
    sender_id       UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    sender_name     TEXT        NOT NULL DEFAULT '',
    content         TEXT        NOT NULL DEFAULT '',
    message_type    VARCHAR(32) NOT NULL DEFAULT 'text',
    signal_data     TEXT        NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_alfq_messages_conv ON alfq_messages(conversation_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_alfq_messages_sender ON alfq_messages(sender_id, created_at DESC);

-- 市场产品
CREATE TABLE IF NOT EXISTS alfq_products (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    author_id      UUID           NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    author_name    TEXT           NOT NULL DEFAULT '',
    name           TEXT           NOT NULL DEFAULT '',
    category       VARCHAR(32)    NOT NULL DEFAULT 'signal',
    description    TEXT           NOT NULL DEFAULT '',
    symbol         VARCHAR(32)    NOT NULL DEFAULT '',
    purchase_type  VARCHAR(32)    NOT NULL DEFAULT 'free',
    price          NUMERIC(10,2)  NOT NULL DEFAULT 0,
    trial_days     INTEGER        NOT NULL DEFAULT 0,
    rating         NUMERIC(3,2)   NOT NULL DEFAULT 0,
    purchase_count INTEGER        NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_alfq_products_category ON alfq_products(category, rating DESC);
CREATE INDEX IF NOT EXISTS idx_alfq_products_author ON alfq_products(author_id, created_at DESC);

-- 购买记录
CREATE TABLE IF NOT EXISTS alfq_purchases (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_id UUID        NOT NULL REFERENCES alfq_products(id) ON DELETE CASCADE,
    is_trial   BOOLEAN     NOT NULL DEFAULT false,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_alfq_purchases_user ON alfq_purchases(user_id, created_at DESC);
