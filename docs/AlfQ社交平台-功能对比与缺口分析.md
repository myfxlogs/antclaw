# AlfQ 社交平台 vs X / Threads 核心功能对比

> 审计日期：2026-05-14
> 审计范围：服务端 proto + handler + DB 表（含通知/告警/SSE/在线追踪）
> 目的：识别功能缺口，制定开发规划

---

## 一、内容发布

| 功能 | AlfQ 现状 | X (Twitter) | Threads | 缺口 |
|---|---|---|---|---|
| 文字帖子 | ✅ `CreatePost` (text/signal_card/chart_share) | ✅ | ✅ | — |
| 图片附件 | ❌ | ✅ (最多 4 张) | ✅ (最多 10 张) | **需新增媒体上传 + post 关联** |
| 视频附件 | ❌ | ✅ (最长 2:20) | ✅ (最长 5 分钟) | **需视频上传/转码/存储** |
| 帖子串 (Thread) | ❌ | ✅ (连续回复自己) | ✅ | **需 `parent_post_id` 字段** |
| 引用转发 (Quote) | ✅ `SharePost` (含 comment) | ✅ Quote Tweet | ✅ Quote | — |
| 纯转发 (Repost) | ❌ | ✅ Retweet | ✅ Repost | **需 `repost` 类型 + 源帖引用** |
| 编辑帖子 | ❌ | ✅ (付费用户) | ✅ (15 分钟内) | **需 `UpdatePost` RPC** |
| 删除帖子 | ❌ | ✅ | ✅ | **需 `DeletePost` RPC** |
| 草稿 | ❌ | ✅ | ✅ | **客户端本地实现** |
| 定时发布 | ❌ | ❌ (第三方) | ❌ | 可选 |
| 帖子可见性 | ✅ public/followers/circle | ✅ public/followers | ✅ public | — |
| 投票帖 | ❌ | ✅ | ✅ | **需 `poll` 类型** |
| @提及用户 | ❌ | ✅ | ✅ | **需解析 + 通知被提及者** |
| #话题标签 | ❌ | ✅ | ✅ | **需 `hashtags` 字段 + 搜索** |
| 信号卡片帖 | ✅ `signal_card` | ❌ (不适用) | ❌ | AlfQ 独有优势 |

### 缺口汇总（内容发布）

| 缺口 | 优先级 | 涉及范围 |
|---|---|---|
| 图片/视频附件 | 高 | MinIO 存储 + proto 扩展 + 新 migration |
| 帖子串 (threading) | 高 | `parent_post_id` 字段 + `GetThread` RPC |
| 删除帖子 | 高 | `DeletePost` RPC + 级联删评论/赞 |
| 编辑帖子 | 中 | `UpdatePost` RPC |
| @提及 + #话题 | 中 | 内容解析 + 通知 + 搜索索引 |
| 投票帖 | 低 | `poll` post_type 扩展 |

---

## 二、互动机制

| 功能 | AlfQ 现状 | X | Threads | 缺口 |
|---|---|---|---|---|
| 点赞 | ✅ `LikePost` / `UnlikePost` | ✅ | ✅ | — |
| 评论 | ✅ `CommentOnPost` | ✅ Reply | ✅ Reply | — |
| 评论嵌套 | ❌ (平铺) | ✅ (多层线程) | ✅ (多层线程) | **需 `parent_comment_id` 字段** |
| 转发 | ✅ `SharePost` | ✅ | ✅ | — |
| 收藏/书签 | ❌ | ✅ Bookmark | ✅ Save | **需 `alfq_bookmarks` 表** |
| 帖子浏览数 | ❌ | ✅ | ✅ | **需 `view_count` 字段 + 记数** |
| 转发数 | ❌ (只有 share_count 字段) | ✅ | ✅ | **需区分 repost vs quote** |
| 帖子置顶 | ❌ | ✅ Pin to profile | ✅ | **需 `pinned_post_id` 字段** |
| 举报/屏蔽 | ❌ | ✅ | ✅ | **需 `alfq_reports` 表 + 审核流** |
| 拉黑用户 | ❌ | ✅ Block | ✅ Block | **需 `alfq_blocks` 表** |
| 静音用户 | ❌ | ✅ Mute | ✅ | **需 `alfq_mutes` 表** |

### 缺口汇总（互动）

| 缺口 | 优先级 | 涉及范围 |
|---|---|---|
| 评论嵌套 (threaded replies) | 高 | `alfq_comments` 加 `parent_id` 字段 |
| 收藏/书签 | 高 | 新表 + `BookmarkPost` / `GetBookmarks` RPC |
| 帖子浏览计数 | 中 | `view_count` 字段 + 去重逻辑 |
| 拉黑/静音 | 中 | 新表 + Feed 过滤逻辑 |
| 举报系统 | 中 | 新表 + 审核后台 |

---

## 三、用户关注与关系

| 功能 | AlfQ 现状 | X | Threads | 缺口 |
|---|---|---|---|---|
| 关注/取消关注 | ✅ Follow / Unfollow | ✅ | ✅ | — |
| 粉丝列表 | ✅ GetFollowers | ✅ | ✅ | — |
| 关注列表 | ✅ GetFollowing | ✅ | ✅ | — |
| 关注数/粉丝数 | ✅ follower_count / following_count | ✅ | ✅ | — |
| 个人资料页 | ✅ GetProfile (含交易统计) | ✅ | ✅ | AlfQ 独有交易数据 |
| 编辑资料 | ✅ UpdateProfile | ✅ | ✅ | — |
| 头像 | ❌ | ✅ | ✅ | **需头像上传/存储** |
| 个人简介 | ✅ bio 字段 | ✅ | ✅ | — |
| 顶图/封面 | ❌ | ✅ | ✅ | **可选** |
| 验证标记 | ✅ tier (normal/verified/elite) | ✅ 蓝 V | ✅ | — |
| 拉黑 | ❌ | ✅ | ✅ | 同上 |
| 静音 | ❌ | ✅ Mute | ✅ Restrict | 同上 |
| 关注请求 (私密账号) | ❌ | ✅ Protected | ✅ Private | **需 `follow_request` 状态** |
| 列表 (Lists) | ❌ | ✅ | ❌ | 可选 |
| 圈子 | ✅ Circles (Create/Join/Leave) | ❌ Communities | ❌ | AlfQ 独有 |

### 缺口汇总（关注）

| 缺口 | 优先级 | 涉及范围 |
|---|---|---|
| 头像上传 | 高 | 存储 + `avatar_url` 字段 |
| 拉黑/静音 | 中 | 关系过滤 + Feed 排除 |
| 私密账号/关注请求 | 低 | 用户状态扩展 |

---

## 四、信息流与发现

| 功能 | AlfQ 现状 | X | Threads | 缺口 |
|---|---|---|---|---|
| 公共时间线 | ✅ `GetFeed` (created_at DESC) | ✅ | ✅ | — |
| 关注者时间线 | ❌ | ✅ Following | ✅ Following | **需 `GetFollowingFeed` RPC (JOIN alfq_follows)** |
| 算法推荐流 | ❌ | ✅ For You | ✅ For You | **需推荐引擎** |
| 圈子 Feed | ✅ `GetCircleFeed` (部分实现) | ❌ | ❌ | AlfQ 独有 |
| 搜索帖子 | ❌ | ✅ | ✅ | **需全文搜索 (pg_trgm/Elasticsearch)** |
| 搜索用户 | ❌ | ✅ | ✅ | **需 `SearchUsers` RPC** |
| 热门话题 | ❌ | ✅ Trending | ✅ Trending | **需话题聚合 + 趋势计算** |
| 个性化推荐 | ❌ | ✅ | ✅ | 后期 |
| 无限滚动 (游标) | ✅ cursor 分页 | ✅ | ✅ | — |
| 信号流 (Signal Bar) | ✅ (客户端 FeedScreen) | ❌ | ❌ | AlfQ 独有 |

### 缺口汇总（信息流）

| 缺口 | 优先级 | 涉及范围 |
|---|---|---|
| 关注者时间线 | 高 | `GetFollowingFeed` RPC + `alfq_follows` JOIN |
| 用户搜索 | 高 | 全文搜索 (pg_trgm GIN 索引) |
| 帖子搜索 | 中 | 全文搜索 |
| 算法推荐 | 低 (后期) | 独立推荐服务 |
| 热门话题 | 低 (后期) | 聚合 + 趋势算法 |

---

## 五、通知系统

| 功能 | AlfQ 现状 | X | Threads | 缺口 |
|---|---|---|---|---|
| 应用内通知 | ✅ NotificationService CRUD | ✅ | ✅ | — |
| SSE 实时推送 | ✅ /sse/notifications (Redis Pub/Sub) | — | — | 架构更优 |
| 通知偏好 | ✅ 类型过滤 + 严重度 + 静默时段 | ✅ | ✅ | — |
| 未读计数 | ✅ UnreadCount | ✅ | ✅ | — |
| 推送通知 (FCM/APNs) | ❌ (仅有 push 类型字段) | ✅ | ✅ | **需 FCM 集成** |
| 邮件通知 | ❌ (仅有 email 类型字段) | ✅ | ✅ | **需 SMTP 集成** |
| Webhook 通知 | ✅ RegisterWebhook | ❌ | ❌ | AlfQ 独有 |
| 通知分组 | ❌ | ✅ (按类型折叠) | ✅ | 可选 |
| 营销通知 | ❌ | ✅ | ✅ | 不适用 |

### 缺口汇总（通知）

| 缺口 | 优先级 | 涉及范围 |
|---|---|---|
| FCM 推送 | 高 | Android 集成 + 服务端 FCM SDK |
| APNs 推送 | 高 | iOS 集成 + 服务端 APNs |
| 邮件通知 | 中 | SMTP 配置 + 模板 |

---

## 六、实时互动能力

| 功能 | AlfQ 现状 | X | Threads | 缺口 |
|---|---|---|---|---|
| 通知实时推送 | ✅ SSE /sse/notifications | ✅ WebSocket | ✅ | — |
| 告警实时推送 | ✅ 6 个 SSE 端点 | ❌ | ❌ | AlfQ 独有 |
| 在线用户追踪 | ✅ presence.Tracker | ✅ | ✅ | — |
| 消息推送 | ✅ SSE + Redis Pub/Sub | ✅ | ✅ | — |
| 打字指示器 | ❌ | ✅ DM typing | ✅ typing | **需 WebSocket (Chat)** |
| 已读回执 | ❌ | ✅ | ✅ | **需 read_at 字段** |
| 直播间/Spaces | ❌ | ✅ Spaces | ❌ | 后期 |
| 实时点赞数 | ❌ | ✅ | ✅ | **需 SSE/WS 推送计数更新** |
| 实时评论 | ❌ | ✅ | ✅ | **需 SSE/WS 推送新评论** |

### 缺口汇总（实时）

| 缺口 | 优先级 | 涉及范围 |
|---|---|---|
| WebSocket 聊天升级 | 中 | Chat 从轮询升级到 WS |
| 实时社交事件推送 | 中 | 新 SSE channel "social:{userID}" |
| 打字指示器 / 已读 | 低 (后期) | WebSocket 协议 |

---

## 七、AlfQ 独有优势（差异化能力）

| 功能 | 说明 |
|---|---|
| 信号卡片帖 | Post 可附带交易对/方向/置信度，交易者直接分享分析 |
| 交易者档案 | Profile 含胜率、盈亏比、夏普比率等量化指标 |
| 交易圈子 | 围绕交易品种建圈子，圈子内独立 Feed |
| 告警系统 | 规则驱动的价格/波动率/宏观告警，含冷却/配额/偏好 |
| 策略市场 | 策略/指标/EA 发布、试用到购买闭环 |
| 实时行情 SSE | 行情、宏观、COT、期权等多维度 SSE 实时流 |

---

## 八、优先级路线图建议

### P0 — 本周（核心社交闭环）

```
□ DeletePost RPC + 级联删除
□ GetFollowingFeed RPC (关注者时间线)
□ SocialFeedScreen (客户端 — 见 docs/客户端-Feed发布无响应问题.md)
□ 评论嵌套 (alfq_comments.parent_id)
```

### P1 — 本月（基础社交体验）

```
□ 图片上传 (MinIO + proto 扩展)
□ 帖子串 (alfq_posts.parent_post_id)
□ 收藏/书签
□ 拉黑/静音用户
□ 用户搜索 (pg_trgm)
□ FCM/APNs 推送
```

### P2 — 下月（增强与差异化）

```
□ 帖子浏览计数
□ @提及 + #话题
□ 编辑帖子
□ 头像上传
□ 实时社交事件 SSE
□ 举报系统
```

### P3 — 远期（规模化）

```
□ 算法推荐引擎
□ 热门话题聚合
□ 视频附件
□ 投票帖
□ Spaces/直播
□ Elasticsearch 全文搜索
```

---

## 九、数据模型缺口清单

以下 DB 表/字段需要新建：

| 表/字段 | 用途 | 优先级 |
|---|---|---|
| `alfq_posts.parent_post_id` | 帖子串 | P1 |
| `alfq_posts.view_count` | 浏览计数 | P2 |
| `alfq_posts.pinned` | 帖子置顶 | P2 |
| `alfq_posts.updated_at` | 编辑时间 | P2 |
| `alfq_comments.parent_id` | 评论嵌套 | P0 |
| `alfq_bookmarks` | 收藏 | P1 |
| `alfq_blocks` | 拉黑 | P1 |
| `alfq_mutes` | 静音 | P2 |
| `alfq_reports` | 举报 | P3 |
| `alfq_post_media` | 媒体附件 | P1 |
| `users.avatar_url` | 头像 | P2 |
| `users.account_type` | 公开/私密 | P3 |
| `alfq_hashtags` + `alfq_post_hashtags` | 话题标签 | P2 |
| `alfq_post_reposts` | 纯转发记录 | P2 |
