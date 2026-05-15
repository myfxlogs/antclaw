-- Migration 031: 社交 Feed 表扩展（parent_comment_id / original_post_id / 游标索引）
-- 对应文档：docs/安卓客户端技术文档包/09-服务端配套改造优化文档.md §4.2

-- alfq_comments: 新增 parent_comment_id 用于评论回复树
ALTER TABLE alfq_comments
  ADD COLUMN IF NOT EXISTS parent_comment_id uuid NULL REFERENCES alfq_comments(id);

-- alfq_posts: 新增 original_post_id 用于 Share / Quote 引用关系
ALTER TABLE alfq_posts
  ADD COLUMN IF NOT EXISTS original_post_id uuid NULL REFERENCES alfq_posts(id);

-- 稳定游标索引（Feed 按 created_at DESC, id DESC）
CREATE INDEX IF NOT EXISTS idx_alfq_posts_feed_cursor
  ON alfq_posts (created_at DESC, id DESC);

-- 作者维度游标索引（个人主页）
CREATE INDEX IF NOT EXISTS idx_alfq_posts_author_cursor
  ON alfq_posts (author_id, created_at DESC, id DESC);

-- 帖子类型维度游标索引（filter）
CREATE INDEX IF NOT EXISTS idx_alfq_posts_type_cursor
  ON alfq_posts (post_type, created_at DESC, id DESC);

-- 评论游标索引（按时间升序）
CREATE INDEX IF NOT EXISTS idx_alfq_comments_post_cursor
  ON alfq_comments (post_id, created_at ASC, id ASC);

-- 父评论索引（回复树查询）
CREATE INDEX IF NOT EXISTS idx_alfq_comments_parent
  ON alfq_comments (parent_comment_id);
