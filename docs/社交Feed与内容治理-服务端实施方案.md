# 社交Feed与内容治理 — 服务端实施方案

> 基于 Twitter Recommendation Algorithm 与 Community Notes 架构研究  
> 适用范围：`backend/internal/service/feed` / `backend/internal/infra/postgres`  
> 版本：v1.0 | 日期：2026-05-16

---

## 一、Feed 推荐系统

### 1.1 关注流拆分

**目标**：用户只看关注作者的帖子，按时间序。

**实现文件**：`backend/internal/infra/postgres/feed_repository.go`

```go
// GetFollowingFeed returns posts from users the viewer follows.
func (r *feedRepo) GetFollowingFeed(ctx context.Context, userID string, cursor *SocialCursor, limit int32) ([]*FeedPostRow, *SocialCursor, error) {
    args := []interface{}{limit + 1, userID}
    query := `SELECT ` + feedSelectColumns + ` FROM alfq_posts p
        WHERE p.author_id IN (SELECT following_id FROM alfq_follows WHERE follower_id = $2)
          AND p.visibility = 'public'
        ORDER BY p.created_at DESC`
    query, args = AppendCursor(query, args, cursor, "p.created_at, p.id", CursorDesc, 2)
    rows, _ := r.pool.Query(ctx, query, args...)
    defer rows.Close()
    return scanFeedPosts(rows)
}
```

**Service 层**（`backend/internal/service/feed/service.go`）：
```go
func (s *Service) GetFollowingFeed(ctx context.Context, userID string, cursor string, limit int32) ([]*alfqv1.Post, string, error) {
    if userID == "" {
        return nil, "", connect.NewError(connect.CodeUnauthenticated, errors.New("login required"))
    }
    c := parseCursor(cursor)
    rows, next, err := s.repo.GetFollowingFeed(ctx, userID, c, limit)
    if err != nil {
        return nil, "", err
    }
    posts := toProtoPosts(rows)
    return posts, next.Encode(), nil
}
```

### 1.2 基础排序管线

**新文件**：`backend/internal/service/feed/ranking.go`

```go
package feed

import (
    "math"
    "sort"
    "time"
)

const (
    RecencyDecay  = 0.1  // λ: 时间衰减系数
    EngagementNorm = 3600 // 1小时归一化
)

type PostScore struct {
    Recency    float64
    Engagement float64
    AuthorCred float64
}

func RankPosts(posts []*FeedPostRow) []*FeedPostRow {
    now := time.Now()
    for _, p := range posts {
        hoursAgo := now.Sub(p.CreatedAt).Hours()
        p.Score = 0.4*math.Exp(-RecencyDecay*hoursAgo) +
                  0.3*(float64(p.LikeCount+p.CommentCount+p.ShareCount))/(hoursAgo+1)*EngagementNorm +
                  0.3*p.AuthorCredScore
    }
    sort.Slice(posts, func(i, j int) bool {
        return posts[i].Score > posts[j].Score
    })
    return posts
}
```

### 1.3 Redis 缓存层

**修改文件**：`backend/internal/service/feed/service.go`

```go
func (s *Service) GetFeedCached(ctx context.Context, userID, filter, cursor string, limit int32) ([]*alfqv1.Post, string, error) {
    cacheKey := fmt.Sprintf("feed:%s:%s:%s:%d", userID, filter, cursor, limit)
    if cached := s.rdb.Get(ctx, cacheKey); cached != "" {
        var wrapper struct { Posts []*alfqv1.Post; Next string }
        json.Unmarshal([]byte(cached), &wrapper)
        return wrapper.Posts, wrapper.Next, nil
    }
    rows, next, err := s.repo.GetFeed(ctx, filter, parseCursor(cursor), limit)
    if err != nil {
        return nil, "", err
    }
    posts := RankPosts(rows)
    protoPosts := toProtoPosts(posts)
    wrapper, _ := json.Marshal(struct{ Posts []*alfqv1.Post; Next string }{protoPosts, next.Encode()})
    s.rdb.Set(ctx, cacheKey, string(wrapper), 2*time.Minute)
    return protoPosts, next.Encode(), nil
}
```

---

## 二、社交图算法

### 2.1 TueepCred 信誉分

**新文件**：`backend/internal/service/cred/tweepcred.go`

```go
package cred

// TweepCredScore holds a user's reputation score.
type TweepCredScore struct {
    UserID          string
    PageRank        float64 // 关注图 PageRank 迭代结果
    InteractionRate float64 // 帖子平均互动率
    SignalAccuracy  float64 // 交易信号预测准确率
    FinalScore      float64
}

// CalculateBulk computes scores for all active users (hourly cron job).
func CalculateBulk(pgPool *pgxpool.Pool) ([]TweepCredScore, error) {
    ctx := context.Background()
    
    // 1. PageRank on follows graph
    ranks := computePageRank(ctx, pgPool)
    
    // 2. Interaction rate per user
    rows, _ := pgPool.Query(ctx, `
        SELECT author_id, 
               AVG((COALESCE(l.likes,0) + COALESCE(c.comments,0) + COALESCE(s.shares,0))::float 
                    / GREATEST(EXTRACT(EPOCH FROM NOW() - p.created_at)/3600, 1)) AS rate
        FROM alfq_posts p
        LEFT JOIN LATERAL (SELECT COUNT(*) AS likes FROM alfq_likes WHERE post_id = p.id) l ON true
        LEFT JOIN LATERAL (SELECT COUNT(*) AS comments FROM alfq_comments WHERE post_id = p.id) c ON true
        LEFT JOIN LATERAL (SELECT COUNT(*) AS shares FROM alfq_posts WHERE original_post_id = p.id AND post_type = 'share') s ON true
        WHERE p.created_at > NOW() - INTERVAL '30 days'
        GROUP BY author_id`)
    defer rows.Close()
    
    // ... aggregate + compute FinalScore
    return scores, nil
}
```

### 2.2 批量写入

```sql
CREATE TABLE IF NOT EXISTS user_cred (
    user_id    UUID PRIMARY KEY REFERENCES users(id),
    score      REAL NOT NULL DEFAULT 0,
    page_rank  REAL NOT NULL DEFAULT 0,
    inter_rate REAL NOT NULL DEFAULT 0,
    sig_acc    REAL NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

---

## 三、内容治理

### 3.1 可见性过滤链

**新文件**：`backend/internal/service/feed/visibility.go`

```go
package feed

type FilterDecision struct {
    Allow    bool
    Drop     bool
    DownRank float64
    Reason   string
}

type FilterRule struct {
    Name     string
    Priority int
    Apply    func(ctx context.Context, pg *pgxpool.Pool, post *FeedPostRow) FilterDecision
}

var DefaultRules = []FilterRule{
    {Name: "low_content", Priority: 0, Apply: checkLowContent},
    {Name: "report_penalty", Priority: 1, Apply: checkReportPenalty},
}

func checkLowContent(ctx context.Context, pg *pgxpool.Pool, post *FeedPostRow) FilterDecision {
    if len(post.Content) < 5 && post.PostType == "text" {
        return FilterDecision{Allow: true, DownRank: 0.5, Reason: "low_content"}
    }
    return FilterDecision{Allow: true}
}

func checkReportPenalty(ctx context.Context, pg *pgxpool.Pool, post *FeedPostRow) FilterDecision {
    var count int
    pg.QueryRow(ctx, `SELECT COUNT(*) FROM alfq_reports WHERE post_id = $1 AND status = 'pending'`, post.ID).Scan(&count)
    if count >= 3 {
        return FilterDecision{Allow: true, DownRank: 0.3, Reason: "reported"}
    }
    return FilterDecision{Allow: true}
}

func ApplyFilters(ctx context.Context, pg *pgxpool.Pool, posts []*FeedPostRow) []*FeedPostRow {
    filtered := make([]*FeedPostRow, 0, len(posts))
    for _, p := range posts {
        allow := true
        for _, rule := range DefaultRules {
            d := rule.Apply(ctx, pg, p)
            if d.Drop { allow = false; break }
            p.VisibilityScore *= d.DownRank
        }
        if allow { filtered = append(filtered, p) }
    }
    return filtered
}
```

### 3.2 举报建表

```sql
CREATE TABLE IF NOT EXISTS alfq_reports (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id     UUID NOT NULL REFERENCES alfq_posts(id),
    reporter_id UUID NOT NULL REFERENCES users(id),
    reason      VARCHAR(100) NOT NULL,
    details     TEXT,
    status      VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending / resolved / dismissed
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);
```

---

## 四、通知系统

### 4.1 优先级常量

**修改文件**：`backend/internal/notify/notify.go`

```go
const (
    NotifPriorityLow      = 0 // 关注、系统通知
    NotifPriorityNormal   = 1 // 点赞、分享
    NotifPriorityHigh     = 2 // 评论、@提及
    NotifPriorityCritical = 3 // 信号告警
)

// 频率控制：同一 category 对同一用户 30 分钟内最多 3 条
func (s *Service) shouldSend(ctx context.Context, n *Notification) bool {
    key := fmt.Sprintf("notify:rate:%s:%s", n.UserID, n.Category)
    count, _ := s.rdb.Incr(ctx, key).Result()
    if count == 1 { s.rdb.Expire(ctx, key, 30*time.Minute) }
    return count <= 3
}
```

### 4.2 通知分类发送

```go
func (s *Service) SendWithFreqControl(ctx context.Context, n *Notification) error {
    if !s.shouldSend(ctx, n) {
        return nil
    }
    return s.Send(ctx, n)
}
```

---

## 五、API 扩展 (Proto)

```proto
// feed.proto
service FeedService {
  rpc GetFollowingFeed(GetFollowingFeedRequest) returns (GetFollowingFeedResponse);
  rpc ReportPost(ReportPostRequest) returns (ReportPostResponse);
}

message GetFollowingFeedRequest {
  int32 page_size = 1;
  string cursor = 2;
}

message ReportPostRequest {
  string post_id = 1;
  string reason = 2;
  string details = 3;
}
```

---

## 六、实施顺序

| 序号 | 任务 | 文件 | 工期 |
|---|---|---|---|
| 1 | `GetFollowingFeed` Repository + Service | `feed_repository.go` + `feed/service.go` | 1 天 |
| 2 | `RankPosts` 排序管线 | `feed/ranking.go` (新) | 1 天 |
| 3 | Redis Feed 缓存 | `feed/service.go` | 0.5 天 |
| 4 | TweepCred 计算 + `user_cred` 表 | `cred/tweepcred.go` (新) | 2 天 |
| 5 | 可见性过滤链 | `feed/visibility.go` (新) | 1 天 |
| 6 | 举报表 + `ReportPost` RPC | `feed_repository.go` + proto | 1 天 |
| 7 | 通知优先级/频率控制 | `notify/notify.go` | 0.5 天 |

---

## 七、验收

```bash
cd backend && go vet ./... && go test ./...
# 新增测试: feed/ranking_test.go, feed/visibility_test.go, notify/rate_test.go
```
