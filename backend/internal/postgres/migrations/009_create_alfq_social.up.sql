-- AlfQ 社交功能表（M3-M5）
-- 迁移编号：009

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
CREATE INDEX IF NOT EXISTS idx_alfq_posts_author ON alfq_posts(author_id);
CREATE INDEX IF NOT EXISTS idx_alfq_posts_created ON alfq_posts(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_alfq_posts_visibility ON alfq_posts(visibility);

CREATE TABLE IF NOT EXISTS alfq_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID NOT NULL REFERENCES alfq_posts(id) ON DELETE CASCADE,
    author_id UUID NOT NULL REFERENCES users(id),
    author_name TEXT NOT NULL DEFAULT '',
    content TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_alfq_comments_post ON alfq_comments(post_id);

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

-- ===== Chat =====
CREATE TABLE IF NOT EXISTS alfq_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL,
    sender_id UUID NOT NULL REFERENCES users(id),
    sender_name TEXT NOT NULL DEFAULT '',
    content TEXT NOT NULL DEFAULT '',
    message_type TEXT NOT NULL DEFAULT 'text',
    signal_data TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_alfq_messages_conv ON alfq_messages(conversation_id, created_at);

-- ===== Circles =====
CREATE TABLE IF NOT EXISTS alfq_circles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    symbol TEXT NOT NULL DEFAULT '',
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS alfq_circle_members (
    circle_id UUID NOT NULL REFERENCES alfq_circles(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (circle_id, user_id)
);

-- ===== Marketplace =====
CREATE TABLE IF NOT EXISTS alfq_products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    author_id UUID NOT NULL REFERENCES users(id),
    author_name TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT 'strategy',
    description TEXT NOT NULL DEFAULT '',
    symbol TEXT NOT NULL DEFAULT '',
    purchase_type TEXT NOT NULL DEFAULT 'subscription',
    price NUMERIC(10,2) NOT NULL DEFAULT 0,
    trial_days INT NOT NULL DEFAULT 0,
    rating NUMERIC(3,2) NOT NULL DEFAULT 0,
    purchase_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS alfq_purchases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    product_id UUID NOT NULL REFERENCES alfq_products(id),
    is_trial BOOLEAN NOT NULL DEFAULT FALSE,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_alfq_purchases_user ON alfq_purchases(user_id);
