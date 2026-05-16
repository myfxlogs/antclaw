# 安卓端 X 社交属性对标与用户体验改造开发指南

> 目标读者：Android 客户端 AI Agent  
> 适用范围：`frontend/android`  
> 对标对象：X 的社交网络体验，包括信息流、关注关系、转发/引用、对话串、发现、通知、私信和个人主页  
> 目标：找出 AntClaw Android 在社交属性和客户端用户体验上的不足，并给出可直接落地的代码开发方案  
> 不处理范围：Web 端、Admin 前端；服务端缺口仅记录依赖，不在 Android Agent 中直接修改服务端

---

## 一、结论

当前 Android 客户端已经有社交骨架：

- 首页 Feed：`ui/feed/FeedScreen.kt`、`FeedViewModel.kt`
- 帖子卡片：`ui/components/PostCard.kt`
- 发帖页：`ui/post/PostScreen.kt`
- 帖子详情：`ui/post/PostDetailScreen.kt`
- 评论区：`ui/social/CommentSection.kt`
- Profile：`ui/profile/ProfileScreen.kt`
- 发现页：`ui/discover/DiscoverScreen.kt`
- 通知中心：`ui/notification/NotificationCenterScreen.kt`
- 私信列表：`ui/chat/ChatScreen.kt`

但它目前更像“带帖子列表的行情/信号 App”，还不是一个用户愿意反复打开、互动和沉淀关系的社交软件。与 X 相比，主要短板是：

1. **关系网络弱**：关注、推荐、作者可信度和关系入口不够突出。
2. **信息流社交语义不足**：缺少回复链、引用转发、更多菜单、书签、分享意图、内容上下文。
3. **发布体验不顺**：成功后不回流 Feed，失败重试不是真重试，草稿和关联信号入口不完整。
4. **发现页不够社交**：搜索框没有真实输入触发，趋势、热门讨论、热门交易员、品种话题未形成 Explore 体验。
5. **通知不够像社交反馈中心**：没有按互动/关注/信号/系统分组，点击通知不能可靠跳转上下文。
6. **Profile 不像社交主页**：缺少头像、用户名、内容 Tab、关注关系入口、帖子列表、交易员可信标签布局。
7. **交互反馈多次不友好**：硬编码英文、重试按钮无实际重试、部分错误只显示原始 message、状态切换不够明确。

本指南要求先把 P0 社交闭环做顺，再做 P1 对标增强。

---

## 二、X 社交属性对标矩阵

| X 社交属性 | X 的体验特征 | 当前 Android 状态 | 缺口等级 | 改造方向 |
|---|---|---|---|---|
| Home 信息流 | For You / Following、无限滚动、互动即时反馈 | 有推荐/信号/最新 Tab 和分页 | P0 | 补关注流语义、顶部新内容提示、拉刷新体验、帖子上下文 |
| 关注关系 | 作者卡、关注按钮、关系推荐 | Profile 有关注按钮，Feed 卡片没有关注关系入口 | P0 | Feed 作者区展示关注/认证/等级，Profile 内容 Tab |
| 帖子互动 | 回复、转发、引用、点赞、分享、收藏、更多 | PostCard 有点赞/评论/分享计数 | P0/P1 | 把 share 拆成转发/引用/系统分享，补更多菜单和收藏 |
| 对话串 | 帖子详情展示主帖、上下文、回复链 | 详情页有主帖和评论 | P1 | 支持回复树、楼中楼、评论输入状态、定位到评论区 |
| 发布 | Compose 自动聚焦、草稿、失败保留、可引用内容 | PostScreen 基础可发布 | P0 | 草稿、字数、真实重试、发布成功回流 Feed |
| Explore | 搜索、趋势、推荐账号、热门话题 | Discover 只有搜索框和热门交易员列表 | P1 | 搜索触发、趋势品种、热门话题、推荐交易员模块化 |
| Notifications | 按社交互动分类，点击回到上下文 | 通知列表有未读和类别 chip | P1 | 分组 Tab、点击路由、互动通知样式 |
| DM | 会话列表、未读、进入会话聊天 | ChatScreen 只有会话列表 | P2 | 会话详情、发送消息、未读同步 |
| Profile | 头像、昵称、用户名、bio、关注数、内容 Tab | Profile 展示资料和指标，无帖子 Tab | P0 | 头图/头像/用户名、帖子/回复/信号 Tab、作者可信信息 |
| 信任与治理 | 认证、举报、屏蔽、静音 | 几乎缺失 | P2 | 交易员认证、举报/屏蔽/静音入口 |

---

## 三、当前代码中的体验问题

### 3.1 PostCard 视觉和交互不像社交内容单元

当前文件：`frontend/android/app/src/main/java/com/antclaw/alfq/ui/components/PostCard.kt`

问题：

- `Card` 使用 `surfaceVariant` 和完整卡片边界，阅读感偏“看板卡片”，不像 X/Threads 的流式内容。
- 作者区只有首字母头像和名称，缺少用户名、认证/等级、关注关系、更多菜单。
- 操作区只有 like/comment/share，且 share 语义不清，是系统分享、转发还是引用不明确。
- `ActionButton` 中 `IconButton` 与 count 分离，点击目标和计数布局不够紧凑。
- `contentDescription = null`，无障碍和可测试性不足。
- `visibilityLabel` 使用 emoji，视觉风格不稳定且不可本地化。
- `ChartShareBody` 里有 `📈` 和 `[Chart]` 硬编码，体验粗糙。

目标：

把 `PostCard` 改成社交流 item，而不是独立业务卡片。

### 3.2 PostScreen 发布链路不友好

当前文件：`frontend/android/app/src/main/java/com/antclaw/alfq/ui/post/PostScreen.kt`

问题：

- 成功后只清空输入并延迟 reset，不返回 Feed，不插入新帖。
- 错误态“重试”只调用 `viewModel.reset()`，不是重试发布。
- `signalPair` 输入框只有 `signalPair.isNotBlank()` 时才显示，用户没有入口添加交易对。
- 没有草稿保存、字数统计、自动聚焦、图片/信号引用扩展位。
- 失败后没有明确保留草稿的反馈。

目标：

发布体验要像 X Compose：轻量、明确、失败不丢内容、成功立即回流。

### 3.3 PostDetail 详情页不够像对话串

当前文件：

- `ui/post/PostDetailScreen.kt`
- `ui/post/PostDetailViewModel.kt`

问题：

- 顶部标题硬编码 `"Post"`，返回按钮硬编码 `"Back"`，错误重试硬编码 `"Retry"`。
- 评论加载错误只存在 state 中，UI 没有明显展示 `commentError`。
- `sendComment` 没有输入 loading、防重复发送、失败保留输入策略。
- 详情页没有主帖上下文、引用原帖、回复树、评论区定位。
- `PostCard` 在详情页仍是卡片样式，没有详情页主帖的沉浸布局。

目标：

帖子详情页要成为社交对话中心，而不是列表卡片加评论列表。

### 3.4 Profile 不是社交主页

当前文件：`frontend/android/app/src/main/java/com/antclaw/alfq/ui/profile/ProfileScreen.kt`

问题：

- 页面只有 displayName、bio、关注数和交易指标，没有头像、用户名、codeId、认证说明。
- 没有帖子/回复/信号/账户 Tab。
- 关注按钮没有显式 loading、防重复点击、失败回滚视觉。
- 没有从 Profile 进入该用户帖子详情的路径。
- 交易指标是卡片展示，和社交身份结合不够。

目标：

Profile 要像 X 的个人主页，但强化交易员可信度。

### 3.5 Discover 没有真实 Explore 行为

当前文件：`frontend/android/app/src/main/java/com/antclaw/alfq/ui/discover/DiscoverScreen.kt`

问题：

- 搜索框只是本地 state，没有触发搜索。
- 页面只展示热门交易员，缺少趋势话题、热门品种、热门帖子。
- 错误态不明显，`state.error` 没有展示。
- `onCircleClick` 参数存在但页面没有圈子入口。

目标：

Discover 要承担 X Explore 的职责：找人、找话题、找热门讨论、找品种。

### 3.6 Navigation 命名和发布入口不一致

当前文件：`frontend/android/app/src/main/java/com/antclaw/alfq/navigation/AlfQNavigation.kt`

问题：

- 组件名是 `BottomNavBarWithFAB`，但没有 FAB。
- 导航里没有一级发布按钮，发布入口不明显。
- mainTabs 包含 `"social"`，但 `AlfQApp` 中 `"social"` 仍渲染 `FeedScreen`，和 `"feed"` 重复。
- 消息入口有，但通知不是一级入口，Feed 顶部和底部入口分散。

目标：

一级导航要体现社交产品：Home、Explore、Compose、Notifications/Messages、Me。

---

## 四、目标信息架构

### 4.1 一级导航

建议短期采用 5 个一级入口：

| 入口 | 路由 | 职责 |
|---|---|---|
| 首页 | `feed` | 推荐、关注、信号、最新信息流 |
| 发现 | `discover` | 搜索、趋势、热门交易员、热门品种 |
| 发布 | `composePost` | 发帖、引用信号、关联品种 |
| 通知 | `notifications` | 社交互动、关注、信号、系统通知 |
| 我的 | `me` | 当前用户主页、设置、MT 账户、告警 |

消息 `chat` 可以先作为通知页或我的页入口。若要保留底部消息入口，则通知入口必须在首页顶部继续稳定存在。

### 4.2 Feed Tab

| Tab | 数据来源 | 当前策略 |
|---|---|---|
| 推荐 | `FeedService/GetFeed filter=all` 或后端推荐流 | P0 保留 |
| 关注 | 关注流，后端未支持前隐藏 | P1 开启 |
| 信号 | `signals_only` | P0 保留 |
| 最新 | 时间倒序 | P0 保留 |

禁止展示没有真实行为的 Tab。

---

## 五、代码开发方案

### 5.1 重构 PostCard 为 TimelinePostItem

新增或重命名：

```text
ui/feed/components/TimelinePostItem.kt
ui/feed/components/PostAuthorRow.kt
ui/feed/components/PostActionBar.kt
ui/feed/components/PostContextBlock.kt
```

目标结构：

```kotlin
@Composable
fun TimelinePostItem(
    post: PostUi,
    actionState: PostActionState = PostActionState(),
    onPostClick: (String) -> Unit,
    onAuthorClick: (String) -> Unit,
    onLikeClick: (String) -> Unit,
    onReplyClick: (String) -> Unit,
    onRepostClick: (String) -> Unit,
    onQuoteClick: (String) -> Unit,
    onBookmarkClick: (String) -> Unit,
    onMoreClick: (String) -> Unit,
)
```

`PostActionState`：

```kotlin
data class PostActionState(
    val likingPostIds: Set<String> = emptySet(),
    val repostingPostIds: Set<String> = emptySet(),
    val bookmarkingPostIds: Set<String> = emptySet(),
)
```

视觉要求：

- 列表 item 使用 `Surface` 或无边界布局，弱化卡片感。
- 左侧头像固定 40dp，右侧内容占满。
- 作者行显示：昵称、用户名/codeId、认证/等级、时间、更多按钮。
- 正文最多 8 行，点击进入详情。
- 信号/品种上下文使用轻量 chip，不使用嵌套大卡片。
- 操作栏图标顺序：回复、转发/引用、点赞、浏览/收藏、分享/更多。
- 所有图标必须有 `contentDescription`。

禁止：

- emoji 作为核心 UI 标签。
- `contentDescription = null`。
- 操作失败无反馈。

### 5.2 扩展 PostUi 社交字段

当前 `PostUi` 已有基本字段，但对标 X 还不够。建议扩展：

```kotlin
data class PostUi(
    val postId: String,
    val authorId: String,
    val authorName: String,
    val authorHandle: String = "",
    val authorAvatar: String? = null,
    val authorTier: String = "normal",
    val authorVerified: Boolean = false,
    val content: String,
    val postType: PostType,
    val signalCard: SignalCardUi? = null,
    val visibility: PostVisibility = PostVisibility.PUBLIC,
    val likeCount: Int = 0,
    val replyCount: Int = 0,
    val repostCount: Int = 0,
    val quoteCount: Int = 0,
    val viewCount: Int = 0,
    val isLiked: Boolean = false,
    val isReposted: Boolean = false,
    val isBookmarked: Boolean = false,
    val createdAt: Instant = Instant.EPOCH,
    val originalPostId: String? = null,
    val quotedPost: PostUi? = null,
)
```

如果后端 proto 暂无字段：

- 不伪造计数。
- UI 隐藏对应状态或显示不可用。
- 在文档或 issue 中记录后端缺口。

### 5.3 FeedViewModel 拆出 TimelineActionController

当前 `FeedViewModel` 和 `SocialFeedViewModel` 重复实现加载与互动。建议新增：

```text
ui/feed/timeline/TimelineState.kt
ui/feed/timeline/TimelineActionController.kt
ui/feed/timeline/TimelineReducer.kt
```

职责：

- `TimelineState`：统一首屏、刷新、分页、互动状态。
- `TimelineActionController`：处理 load/refresh/loadMore/like/repost/bookmark。
- `TimelineReducer`：纯函数处理乐观更新和回滚。

状态建议：

```kotlin
data class TimelineState(
    val posts: List<PostUi> = emptyList(),
    val currentTab: HomeFeedTab = HomeFeedTab.RECOMMENDED,
    val initial: AsyncPhase = AsyncPhase.Idle,
    val refreshing: Boolean = false,
    val append: AsyncPhase = AsyncPhase.Idle,
    val nextCursor: String? = null,
    val hasMore: Boolean = true,
    val actionState: PostActionState = PostActionState(),
)

sealed class AsyncPhase {
    data object Idle : AsyncPhase()
    data object Loading : AsyncPhase()
    data class Error(val message: String) : AsyncPhase()
}
```

验收：

- 首屏错误不影响已有缓存列表。
- 分页错误只在列表底部显示。
- 点赞/转发/收藏失败回滚。
- 重复点击同一 action 时按钮显示 loading 或禁用。

### 5.4 发布页改为 ComposePost 体验

新增：

```text
ui/post/compose/ComposePostScreen.kt
ui/post/compose/ComposePostViewModel.kt
ui/post/compose/ComposePostState.kt
data/local/SocialDraftStore.kt
```

`ComposePostState`：

```kotlin
data class ComposePostState(
    val draft: PostDraft = PostDraft(),
    val submitState: AsyncPhase = AsyncPhase.Idle,
    val charLimit: Int = 280,
    val lastSubmittedPost: PostUi? = null,
)

data class PostDraft(
    val content: String = "",
    val visibility: String = "public",
    val signalPair: String = "",
    val signalDirection: String = "",
    val signalConfidence: Int = 0,
    val quotePostId: String? = null,
)
```

必须实现：

- 自动聚焦输入框。
- 字数统计。
- 空内容禁用。
- 提交中禁用。
- 失败保留草稿。
- 重试重新提交同一 draft。
- 成功后返回 Feed 并插入顶部，或导航到新帖子详情。
- 显式“关联品种/信号”按钮，不靠 `signalPair.isNotBlank()` 显示输入框。

### 5.5 帖子详情改造为 Conversation Screen

建议重命名或新增：

```text
ui/post/detail/PostConversationScreen.kt
ui/post/detail/PostConversationViewModel.kt
ui/post/detail/ReplyComposer.kt
ui/post/detail/ReplyThread.kt
```

目标 UI：

```text
顶部栏：返回 + 标题“帖子”
主帖区：作者、正文、信号上下文、统计、操作栏
回复输入区：固定底部或评论列表前
回复列表：按时间或树结构
分页区：加载更多 / 错误重试
```

必须补：

- `commentError` 展示在评论区。
- `sendComment` 有 loading、防重复、失败保留输入。
- 评论发送成功插入列表并滚动到新评论。
- 主帖不存在时显示“内容不存在或已被删除”。
- 文案全部资源化。

P1 增强：

- 支持 `parentCommentId` 回复。
- 支持从通知跳转并定位某条评论。
- 支持引用原帖块。

### 5.6 Profile 改造成社交主页

建议拆分：

```text
ui/profile/ProfileScreen.kt          # 薄入口
ui/profile/ProfileHeader.kt
ui/profile/ProfileTabs.kt
ui/profile/ProfileTimeline.kt
ui/profile/ProfileViewModel.kt
ui/profile/ProfileState.kt
```

目标结构：

```text
Header:
  avatar / displayName / handle / verified tier / bio
  follower / following / joined date
  Follow button or Edit profile
TraderTrust:
  win rate / profit factor / sharpe / total trades / MT verified
Tabs:
  Posts / Replies / Signals / Accounts
Timeline:
  ListUserPosts / replies / signal posts
```

P0 必做：

- 显示用户名或 codeId。
- 关注按钮 loading、防重复点击、失败回滚。
- 帖子 Tab 调用 `FeedService/ListUserPosts`。
- 点击 Profile 中帖子进入详情。

P1：

- 回复 Tab。
- 信号 Tab。
- 账户可信标识。

### 5.7 Discover 改造成 Explore

建议拆分：

```text
ui/discover/DiscoverScreen.kt
ui/discover/DiscoverSearchBar.kt
ui/discover/TrendingSymbolsSection.kt
ui/discover/TrendingTopicsSection.kt
ui/discover/RecommendedTradersSection.kt
ui/discover/SearchResultsSection.kt
```

状态：

```kotlin
data class DiscoverState(
    val query: String = "",
    val searchState: AsyncPhase = AsyncPhase.Idle,
    val results: List<SearchResultUi> = emptyList(),
    val trendingSymbols: List<TrendSymbolUi> = emptyList(),
    val trendingTopics: List<TrendTopicUi> = emptyList(),
    val recommendedTraders: List<TraderProfileUi> = emptyList(),
    val error: String? = null,
)
```

必须实现：

- 搜索框 debounce 后触发搜索。
- 空 query 展示趋势和推荐。
- 有 query 展示搜索结果。
- 错误态按模块展示，不整页空白。
- 点击品种进入信号/市场页。
- 点击交易员进入 Profile。

### 5.8 通知中心改造成社交反馈中心

当前 `NotificationCenterScreen` 有基本列表，但应对标 X Notifications。

建议新增 Tab：

| Tab | 内容 |
|---|---|
| 全部 | 所有通知 |
| 互动 | 点赞、评论、转发、引用 |
| 关注 | 新关注、交易员动态 |
| 信号 | 信号、告警、市场事件 |
| 系统 | 系统通知 |

必须实现：

- 通知 item 点击后根据 `data` 跳转：post/profile/signal/alert。
- `markRead` 失败要回滚本地 unread 状态。
- `markAllRead` 提交中禁用按钮。
- 前台 SSE 收到通知插入顶部并显示轻提示。
- App 回前台补拉，修正漏收。

### 5.9 私信体验分阶段

当前 `ChatScreen` 只有会话列表。P2 再做完整 DM，但 P1 至少要避免“点了没反应”。

短期要求：

- 会话 item 可点击。
- 如果会话详情未实现，显示明确“私信详情暂不可用”，不要无响应。
- 未读数和时间显示稳定。

---

## 六、用户体验统一规则

### 6.1 所有社交操作必须有即时反馈

| 操作 | 立即反馈 | 失败处理 |
|---|---|---|
| 点赞 | 图标变色、计数变化 | 回滚并 Snackbar |
| 转发 | 计数变化或弹出菜单 | 回滚并 Snackbar |
| 收藏 | 图标状态变化 | 回滚并 Snackbar |
| 关注 | 按钮变 loading/已关注 | 回滚 followerCount |
| 发帖 | 按钮 loading | 保留草稿，重试 |
| 评论 | 输入框 loading | 保留输入，重试 |
| 已读通知 | 本地标记已读 | 回滚 unread |

### 6.2 错误文案不能直接暴露异常

建立错误映射：

```kotlin
fun Throwable.toUserMessage(): Int
```

常用文案：

| 场景 | 文案 |
|---|---|
| 网络失败 | 网络连接失败，请稍后重试 |
| 未登录 | 登录已过期，请重新登录 |
| 无权限 | 你没有权限执行此操作 |
| 内容不存在 | 内容不存在或已被删除 |
| 服务不可用 | 服务暂时不可用，请稍后再试 |

原始异常只写日志。

### 6.3 空态必须给下一步动作

| 页面 | 空态动作 |
|---|---|
| 推荐 Feed 空 | 去发现页找交易员 / 刷新 |
| 关注 Feed 空 | 去发现页关注交易员 |
| 信号 Feed 空 | 添加自选品种 / 刷新 |
| Profile 帖子空 | 如果是本人，去发布；如果是他人，提示暂无内容 |
| 通知空 | 提示“暂无通知”，不需要按钮 |
| Discover 空 | 换关键词 / 查看热门交易员 |

禁止空态按钮 `onClick = {}`。

### 6.4 文案和图标

要求：

- 所有用户可见文案进入 `strings.xml`。
- 不使用 emoji 表达核心状态。
- 图标按钮必须有 contentDescription。
- 不同页面的 bullish/bearish/neutral 颜色统一从 theme 获取。

---

## 七、开发实施顺序

### M1：修复发布和帖子详情体验

文件：

- `ui/post/PostScreen.kt`
- `ui/post/PostViewModel.kt`
- `ui/post/PostDetailScreen.kt`
- `ui/post/PostDetailViewModel.kt`
- `ui/social/CommentSection.kt`

任务：

1. 发帖失败保留草稿并真重试。
2. 发布成功回流 Feed 或进入详情。
3. 评论发送 loading、防重复、失败保留。
4. 详情页评论错误态展示。
5. 文案资源化。

### M2：重构 PostCard / Timeline UI

文件：

- `ui/components/PostCard.kt`
- 新增 `ui/feed/components/*`

任务：

1. 拆作者区、正文区、上下文区、操作区。
2. 操作栏支持回复、转发/引用、点赞、收藏、更多。
3. 消除大卡片感，改为社交流布局。
4. 所有图标 contentDescription。

### M3：Profile 社交主页

文件：

- `ui/profile/ProfileScreen.kt`
- `ui/profile/ProfileViewModel.kt`
- 新增 `ProfileHeader/ProfileTabs/ProfileTimeline`

任务：

1. Header 展示头像、昵称、handle/codeId、bio、认证/等级。
2. 关注按钮 loading 和失败回滚。
3. Posts Tab 接 `ListUserPosts`。
4. 点击帖子进详情。

### M4：Discover Explore 化

文件：

- `ui/discover/DiscoverScreen.kt`
- `ui/discover/DiscoverViewModel.kt`

任务：

1. 搜索框 debounce 触发搜索。
2. 空 query 展示趋势品种、热门话题、推荐交易员。
3. 错误按模块展示。
4. 点击结果进入 Profile / Signal / Post。

### M5：通知中心社交化

文件：

- `ui/notification/NotificationCenterScreen.kt`
- `ui/notification/NotificationViewModel.kt`

任务：

1. 通知分类 Tab。
2. 通知点击路由。
3. markRead / markAllRead loading 和失败回滚。
4. SSE 新通知插入顶部并补拉校正。

### M6：导航和发布入口

文件：

- `navigation/AlfQNavigation.kt`
- `AlfQApp.kt`

任务：

1. `BottomNavBarWithFAB` 要么实现真正发布 FAB，要么改名。
2. 去掉重复的 `social` Tab 或赋予真实语义。
3. 发布入口固定可见。
4. 通知/消息入口清晰。

---

## 八、后端依赖清单

Android Agent 不直接改服务端。遇到以下缺口，记录给服务端 Agent：

| 客户端能力 | 后端依赖 |
|---|---|
| 关注流 | `FeedService/GetFeed` 支持 following filter |
| 引用转发 | `SharePost` 返回 `original_post_id` 和 quoted post 信息 |
| 收藏 | 新增 Bookmark RPC |
| Profile 帖子 Tab | `FeedService/ListUserPosts` |
| 回复树 | `Comment.parent_comment_id` + `ListComments` |
| 通知跳转 | Notification `data` 包含 `post_id/user_id/signal_id` |
| 搜索 | `SearchService` 支持 users/posts/symbols |
| 趋势话题 | `TrendService` 返回 topics/symbols |
| 推荐交易员 | `TraderService/ListRecommendedTraders` |

如果后端未支持，不允许客户端假造数据或假装功能可用。

---

## 九、测试要求

### 9.1 必补 ViewModel 测试

- `ComposePostViewModelTest`
  - 空内容不能提交。
  - 提交中防重复。
  - 失败保留草稿。
  - retry 使用同一 draft。
  - 成功返回 PostUi。
- `PostConversationViewModelTest`
  - 主帖加载失败显示错误。
  - 评论加载失败不影响主帖。
  - 发送评论失败保留输入。
  - 发送成功插入评论并增加计数。
- `TimelineReducerTest`
  - like/repost/bookmark 乐观更新和回滚。
  - append failure 不清空旧列表。
- `ProfileViewModelTest`
  - follow/unfollow 成功和失败回滚。
  - posts tab 加载失败只影响 tab。
- `DiscoverViewModelTest`
  - query debounce 后搜索。
  - 空 query 加载趋势和推荐。
- `NotificationViewModelTest`
  - markRead 失败回滚。
  - SSE 通知插入顶部。

### 9.2 Compose UI 测试

至少覆盖：

- Feed 空态按钮能跳 Discover。
- Post item 操作按钮有 contentDescription。
- ComposePost 失败后文本仍在。
- Profile 有 Posts Tab。
- Notification Tab 可切换。

### 9.3 手工验收

1. 登录后进入 Feed，能阅读真实内容。
2. 点赞后立即变化，断网失败能回滚。
3. 发布失败不丢草稿。
4. 发布成功后能看到新帖。
5. 进入帖子详情能评论。
6. 评论失败不丢输入。
7. 点击作者进入 Profile。
8. Profile 能关注/取关并看到帖子。
9. Discover 搜索能返回结果或明确空态。
10. 通知点击能跳到对应上下文。

---

## 十、禁止事项

- 禁止硬编码帖子、用户、点赞数、趋势话题作为 fallback。
- 禁止点击无反馈。
- 禁止错误后清空用户输入。
- 禁止显示无效 Tab。
- 禁止用 emoji 替代正式 UI 状态。
- 禁止 ViewModel 直接拼 RPC path。
- 禁止 Screen 直接访问 RPC。
- 禁止把服务端缺口用客户端假数据绕过。
- 禁止用户可见文案硬编码英文。

---

## 十一、最终验收标准

完成本指南后，Android 客户端应具备以下体验：

- 像 X 一样可以围绕关注关系消费信息流。
- 可以顺畅发布、回复、点赞、转发/引用。
- Profile 能承载交易员身份和内容沉淀。
- Discover 能帮助用户找到人、品种和讨论。
- Notifications 能反馈社交互动并跳转上下文。
- 所有失败都有明确反馈，用户输入不丢失。
- 所有社交状态来自真实后端或本地缓存，不伪造繁荣。

