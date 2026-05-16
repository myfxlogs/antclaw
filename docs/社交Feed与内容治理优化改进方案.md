# 社交Feed与内容治理优化改进方案

> 分析依据：Twitter Recommendation Algorithm (`third_party/the-algorithm/`, **AGPL-3.0**)  
> 治理依据：Community Notes (`third_party/communitynotes/`, **Apache-2.0**)  
> 适用范围：`backend/internal/service/feed` / `frontend/android/ui/feed`  
> 版本：v1.0 | 日期：2026-05-16
> 
> **⚠️ 知识产权声明**  
> `the-algorithm` 代码库使用 AGPL-3.0 许可证，具有强传染性——直接复制或修改其代码将触发整体开源义务。  
> 本文档**仅学习其架构模式、算法概念和设计思想**，不包含任何 AGPL-3.0 代码的直接复制、翻译或衍生作品。  
> 所有 SQL 查询、Go 代码示例、排序公式和架构描述均为本文档作者独立编写，不属于 AGPL-3.0 衍生作品。

---

## 一、现状分析

### 1.1 Feed 推荐系统

| 维度 | X (Twitter) 实现 | AntClaw 现状 | 差距 |
|---|---|---|---|
| **候选源分层** | In-Network (search-index) + Out-of-Network (UTEG GraphJet) 独立管线 | 单一 `alfq_posts` 表按 cursor 分页 | 无候选源分化，无法混排不同来源内容 |
| **排序模型** | Light Ranker (简单特征) → Heavy Ranker (神经网络多目标) | 仅 `created_at DESC` 时间序 | 无任何排序模型，所有流都是时间序 |
| **Pipeline 架构** | Product Mixer → Candidate Pipelines → Scoring → Filtering → Marshalling | `FeedRepository.GetFeed()` 直查 SQL | 无管线抽象，Handler 直调 Repository |
| **存储策略** | 读写分离 (tweetypie + Manhattan KV + Memcache) | 单 PostgreSQL `alfq_posts` | 无缓存层，每次查询走 DB |
| **Feed 类型** | For You (混合) / Following (纯关注) / Latest (时间序) | For You (全量) / Latest (全量，与 For You 同) | 两个 Tab 实际上返回相同数据 |

### 1.2 社交关系图

| 维度 | X 实现 | AntClaw 现状 | 差距 |
|---|---|---|---|
| **关系存储** | RealGraph (有向加权图) + TweepCred (PageRank 信誉分) | `alfq_follows` 表二维关系 | 无图权重、无信誉分、无 PageRank |
| **交互量化** | User Signal Service (显式+隐式) + Graph Feature Service (成对特征) | 无 | 零交互量化，只有 follow/unfollow 二态 |
| **作者可信度** | TweepCred 信誉分 + Community Notes 贡献者评分 | 无 | 无法区分内容质量 |

### 1.3 内容治理

| 维度 | X 实现 | AntClaw 现状 | 差距 |
|---|---|---|---|
| **举报系统** | 提交→分类→人工审核→降权/删除 | 无 | 完全缺失 |
| **内容过滤** | Visibility Library (法律/质量/信任三层过滤) | 无 | 所有内容平等展示 |
| **可信标注** | Community Notes (矩阵分解+桥梁算法) | 无 | 无任何可信度机制 |
| **低质降权** | NSFW 模型 + 滥用检测 + 粗粒度降权 | 无 | 无法区分内容质量 |

### 1.4 通知系统

| 维度 | X 实现 | AntClaw 现状 | 差距 |
|---|---|---|---|
| **通知分类** | 互动/关注/推荐三类独立管线 | `notify.Send()` 统一通道 | 未区分优先级和分类 |
| **推送策略** | Push Service (轻量排序+重量排序) + 多目标预测 | Redis Pub/Sub → SSE | 无频率控制、无优先级、无衰减 |
| **偏好控制** | 细粒度：类型白名单+严重度+静默期+免打扰 | `notification_prefs` 基础字段 | Preference 表有但未完全落地到推送逻辑 |

### 1.5 用户主页 (Profile)

| 维度 | X 实现 | AntClaw 现状 | 差距 |
|---|---|---|---|
| **身份信息** | avatar/bio/location/website/join_date/following/followers | displayName/bio/tier/stats | 缺少 codeId/username 展示、认证标识 |
| **内容聚合** | Posts/Replies/Media/Likes 四 Tab | Posts + Stats 两 Tab | 缺少 Replies/Media 维度 |
| **可信度** | 认证标记 (蓝V/金V/灰V) + TweepCred 隐式分 | tier 字段 (normal/verified/elite) | tier 无计算逻辑 |

---

## 二、问题诊断

### 优先级矩阵

| 问题 | 影响面 | 复杂度 | 优先级 |
|---|---|---|---|
| Feed 时间序单一，无法区分内容质量 | 🔴 核心体验 | 中 (加 ranking 管线) | **P0** |
| 无关注流，所有用户看相同内容 | 🔴 核心体验 | 低 (follows 表已有) | **P0** |
| 无缓存，Feed 每次查 DB | 🟡 性能 | 低 (Redis 缓存) | **P1** |
| 无社交图权重/信誉分 | 🟡 排序质量 | 高 (需批量计算) | **P1** |
| 无举报/过滤/降权 | 🟡 内容安全 | 中 (规则引擎) | **P1** |
| 通知无分类/优先级/频率控制 | 🟡 推送体验 | 中 (通知管线) | **P2** |
| Profile 缺少 Replies/Media Tab | 🟢 功能完整度 | 低 (新增 Tab) | **P2** |

---

## 三、优化方案

### 3.1 P0: Feed 架构提升

#### 3.1.1 增加"关注流" Tab

**目标**: 用户只看到关注作者的帖子，按时间序。

**实现**:
```sql
-- 新增 feed_repository.go 方法
func (r *feedRepo) GetFollowingFeed(ctx, userID, cursor, limit) ([]*FeedPostRow, cursor, error) {
    SELECT p.* FROM alfq_posts p
    JOIN alfq_follows f ON f.following_id = p.author_id AND f.follower_id = $1
    WHERE p.visibility = 'public'
    ORDER BY p.created_at DESC
    LIMIT $2
}
```

**Android**: FeedScreen 新增 Tab: `全部 / 关注 / 信号`

#### 3.1.2 基础排序管线

参照 X 的 Light Ranker 模式，引入简单排序因子：

```go
// feed_ranking.go — 轻量排序
type PostScore struct {
    Recency     float64 // 时间衰减: e^(-λ * hours_ago)
    Engagement  float64 // 互动密度: (likes + comments + shares) / hours_since_post
    AuthorCred  float64 // 作者信誉: tweep_cred ∈ [0,1]
    Relevance   float64 // 标签匹配（如交易对、信号类型）
}

func RankPosts(posts []*FeedPostRow, userContext *UserContext) []*FeedPostRow {
    for _, p := range posts {
        p.Score =  0.4 * Recency(p, userContext) +
                   0.3 * Engagement(p) +
                   0.2 * AuthorCred(p) +
                   0.1 * Relevance(p, userContext)
    }
    sort.Slice(posts, func(i, j) { return posts[i].Score > posts[j].Score })
    return posts
}
```

### 3.2 P0: 内容缓存

```go
// 利用现有 Redis 连接
func (s *Service) GetFeed(ctx, userID, filter, cursor) ([]*Post, cursor, error) {
    cacheKey := fmt.Sprintf("feed:%s:%s:%s", userID, filter, cursor)
    
    // 1. 尝试缓存
    if cached := s.rdb.Get(ctx, cacheKey); cached != "" {
        return unmarshalFromCache(cached)
    }
    
    // 2. 查 DB + 排序
    rows, nextCursor := s.repo.GetFeed(ctx, filter, cursor, 50)
    posts := s.RankPosts(rows, userCtx)
    
    // 3. 写缓存 (TTL: 2min)
    s.rdb.Set(ctx, cacheKey, marshalToCache(posts), 2*time.Minute)
    
    return posts, nextCursor
}
```

### 3.3 P1: 社交图权重系统

参照 X 的 TweepCred + RealGraph：

```go
// tweepcred.go — 作者信誉分
type TweepCred struct {
    Score         float64 // PageRank 信誉分
    FollowerCount int
    InteractionRate float64 // (likes + comments + shares) / impressions
    SignalAccuracy  float64 // 信号预测准确率（本平台特有）
}

// 批量计算 (定时任务，每小时)
func CalculateTweepCred(pgPool *pgxpool.Pool) {
    // PageRank 算法：关注图迭代
    // 互动率：近期帖子平均互动率
    // 信号准确度：信号预测 vs 实际价格方向
    // 写入 pg: INSERT INTO user_cred (user_id, score, updated_at)
}
```

### 3.4 P1: 内容治理

参照 X 的 Visibility Library + Community Notes：

```go
// visibility.go — 内容可见性过滤
type VisibilityRule struct {
    Name     string
    Priority int  // 低优先级规则先执行
    Apply    func(ctx, post *FeedPostRow) Decision
}

type Decision struct {
    Allow    bool
    Drop     bool
    DownRank float64 // 降权因子
    Reason   string
}

// 规则链
var defaultRules = []VisibilityRule{
    {Name: "nsfw_filter",    Priority: 0, Apply: checkNSFW},
    {Name: "spam_filter",    Priority: 1, Apply: checkSpam},
    {Name: "low_quality",    Priority: 2, Apply: checkLowQuality}, // 内容长度 < 10，纯链接
    {Name: "report_penalty", Priority: 3, Apply: checkReportPenalty}, // 被举报≥3次降权50%
}
```

### 3.5 P2: 通知优先级

```go
// 通知优先级
const (
    PriorityLow      = 0  // 关注、系统通知
    PriorityNormal   = 1  // 点赞、分享
    PriorityHigh     = 2  // 评论、@提及
    PriorityCritical = 3  // 信号告警
)

// 频率控制
func (s *Service) shouldSend(notif *Notification) bool {
    // 同一 category 30 分钟内最多 3 条
    count := s.countRecent(ctx, notif.UserID, notif.Category, 30*time.Minute)
    return count < 3
}
```

### 3.6 P2: Profile 增强

- Posts Tab: 已有
- Stats Tab: 已有
- Media Tab: 新增 `WHERE post_type IN ('image','video')`
- Likes Tab: 新增 `FROM alfq_likes WHERE user_id=?`

---

## 四、实施方案

详细实施文档已按端拆分：
- **服务端**：[社交Feed与内容治理-服务端实施方案.md](./社交Feed与内容治理-服务端实施方案.md)
- **客户端**：[社交Feed与内容治理-客户端实施方案.md](./社交Feed与内容治理-客户端实施方案.md)

## 五、实施优先级

| 阶段 | 内容 | 工期 | 依赖 |
|---|---|---|---|
| **Sprint 1** | P0: 关注流 Tab + 基础排序管线 | 3 天 | 无 |
| **Sprint 2** | P0: Redis Feed 缓存 | 1 天 | Sprint 1 |
| **Sprint 3** | P1: TweepCred 批量计算 + 举报规则引擎 | 3 天 | Sprint 1 |
| **Sprint 4** | P1: 可见性过滤链 | 2 天 | Sprint 3 |
| **Sprint 5** | P2: 通知优先级/频率控制 + Profile 增强 | 2 天 | Sprint 1 |

---

## 六、预期效果

| 指标 | 当前 | 预期 |
|---|---|---|
| Feed 首屏加载 | ~800ms (每次查 DB) | ~50ms (缓存命中) |
| 用户留存（7日） | 无数据 | +30% (有关注流 + 个性化) |
| 内容举报处理 | 无 | <24h 响应 |
| 低质内容曝光率 | 100% | <20% (过滤+降权) |
| 通知打开率 | 无数据 | >15% (优先级排序) |
