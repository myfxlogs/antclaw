-- Migration: Trend pre-aggregation tables (S12-P1-03)
-- Replaces real-time subquery-based trend computation with pre-aggregated windows.

CREATE TABLE IF NOT EXISTS trend_topics (
    id              BIGSERIAL PRIMARY KEY,
    window          VARCHAR(5) NOT NULL,   -- 1h, 24h, 7d
    topic           VARCHAR(200) NOT NULL,
    post_count      INT4 NOT NULL DEFAULT 0,
    engagement_count INT4 NOT NULL DEFAULT 0,
    calculated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(window, topic)
);
CREATE INDEX IF NOT EXISTS idx_trend_topics_window ON trend_topics(window, engagement_count DESC);

CREATE TABLE IF NOT EXISTS hot_symbols_agg (
    id              BIGSERIAL PRIMARY KEY,
    window          VARCHAR(5) NOT NULL,
    symbol          VARCHAR(50) NOT NULL,
    post_count      INT4 NOT NULL DEFAULT 0,
    signal_count    INT4 NOT NULL DEFAULT 0,
    engagement_count INT4 NOT NULL DEFAULT 0,
    calculated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(window, symbol)
);
CREATE INDEX IF NOT EXISTS idx_hot_symbols_window ON hot_symbols_agg(window, engagement_count DESC);

-- Backfill: compute initial 24h aggregates
INSERT INTO trend_topics (window, topic, post_count, engagement_count)
SELECT '24h', topic, post_count, engagement_count FROM (
    SELECT unnest(regexp_matches(content, '#([A-Za-z0-9_]+)', 'g')) AS topic,
           COUNT(*)::int4 AS post_count,
           SUM(
               (SELECT COUNT(*) FROM alfq_likes WHERE post_id = p.id) +
               (SELECT COUNT(*) FROM alfq_comments WHERE post_id = p.id) +
               (SELECT COUNT(*) FROM alfq_posts WHERE original_post_id = p.id AND post_type = 'share')
           )::int4 AS engagement_count
    FROM alfq_posts p
    WHERE p.visibility = 'public' AND p.created_at >= NOW() - INTERVAL '24 hours'
    GROUP BY topic
) sub
ON CONFLICT (window, topic) DO UPDATE SET
    post_count = EXCLUDED.post_count,
    engagement_count = EXCLUDED.engagement_count,
    calculated_at = NOW();

INSERT INTO hot_symbols_agg (window, symbol, post_count, signal_count, engagement_count)
SELECT '24h', symbol, post_count, signal_count, engagement_count FROM (
    SELECT signal_pair AS symbol,
           COUNT(*)::int4 AS post_count,
           COUNT(*) FILTER (WHERE post_type = 'signal_card')::int4 AS signal_count,
           SUM(
               (SELECT COUNT(*) FROM alfq_likes WHERE post_id = p.id) +
               (SELECT COUNT(*) FROM alfq_comments WHERE post_id = p.id) +
               (SELECT COUNT(*) FROM alfq_posts WHERE original_post_id = p.id AND post_type = 'share')
           )::int4 AS engagement_count
    FROM alfq_posts p
    WHERE p.visibility = 'public' AND p.created_at >= NOW() - INTERVAL '24 hours' AND p.signal_pair != ''
    GROUP BY signal_pair
) sub
ON CONFLICT (window, symbol) DO UPDATE SET
    post_count = EXCLUDED.post_count,
    signal_count = EXCLUDED.signal_count,
    engagement_count = EXCLUDED.engagement_count,
    calculated_at = NOW();
