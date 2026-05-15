# X / Threads 对标社交客户端落地指导

## 一、文档目的

本文是 Android 客户端面向“交易员版 X / Threads”目标的当前执行依据，供客户端 AI Agent 在 `/frontend/android` 代码开发时使用。

本文只负责 Android 客户端落地，不负责服务端代码改造。服务端 AI Agent 必须阅读 `09-服务端配套改造优化文档.md`。

本文不替代以下文档：

- `00-文档索引.md`：文档入口。
- `01-API接口总览.md`：Service / RPC 查询。
- `02-API字段参考.md`：请求响应字段查询。
- `03-SSE实时通知接口.md`：实时通知与重连。
- `04-客户端开发指导.md`：Android 分层、状态、性能。
- `07-文档质量评审与落地边界.md`：接口范围与禁止实现边界。
- `09-服务端配套改造优化文档.md`：服务端 Agent 的接口、分层、鉴权、分页和测试整改依据。

AI Agent 开发社交功能时，必须先读本文，再按涉及模块读取 API 与字段文档。

## 二、产品目标

### 2.1 对标对象

| 产品 | 借鉴点 | AntClaw Android 取舍 |
|---|---|---|
| X | 信息流、关注关系、转发/引用、趋势话题、通知实时性、个人主页 | 保留高频社交交互，弱化泛娱乐内容 |
| Threads | 极简发布、轻量互动、对话串、低干扰视觉 | 保留低门槛发布和干净阅读体验 |

### 2.2 AntClaw 差异化定位

一句话：**交易员版 X / Threads，围绕交易信号、观点、账户可信度和实时讨论构建社交网络。**

核心差异：

- **信号即内容**：系统信号、价格变化、告警触发可以成为信息流内容源。
- **身份可信度**：用户主页不仅展示简介，还展示交易数据、绑定账户、胜率、风险指标等可信线索。
- **实时协作**：通知、评论、关注、私聊围绕交易机会快速发生。
- **弱图表、强讨论**：移动端优先阅读、互动、发布，不承载复杂桌面级分析终端。

## 三、当前代码缺陷与目标差距

### 3.1 首页 Feed 仍是固定看板，不是社交信息流

当前问题：

- `FeedViewModel` 使用硬编码品种列表。
- Feed Card 由信号接口拼装，缺少真实帖子流。
- Tab 显示为 `热门 / 推荐 / 信号 / 关注`，但点击无行为。
- 部分 RPC 失败被静默吞掉，用户不知道数据缺失。

目标状态：

- 首页应由 `FeedService` 或后端社交 Feed 返回统一内容流。
- 每条内容具有明确类型：用户帖子、系统信号、引用、回复、账户动态、告警事件。
- Tab 必须真实切换：推荐、关注、信号、最新。
- 支持下拉刷新、分页加载、失败重试、空态引导。

### 3.2 社交交互链路不完整

当前问题：

- 点赞、评论、转发、引用、收藏等能力不完整或不统一。
- 发帖页、详情页、社交流页面缺少统一交互规范。
- 失败反馈不稳定，部分操作失败静默。

目标状态：

- 每个 Feed Item 至少支持：点赞、评论入口、分享/引用入口、更多菜单。
- 与交易信号相关内容支持：查看信号详情、订阅告警、加入讨论。
- 所有用户操作必须有乐观更新、失败回滚或明确错误提示。

### 3.3 身份体系不足以支撑交易员社交

当前问题：

- 个人页与交易员页已有雏形，但关注、数据指标、认证标识、内容列表不完整。
- 多处加载失败静默，用户可能看到空白资料。

目标状态：

- 主页应类似 X / Threads 的 Profile：头像、昵称、用户名、简介、认证/等级、关注数、粉丝数、交易指标、内容 Tab。
- 内容 Tab 至少包括：帖子、回复、信号、账户。
- 关注/取关必须有明确状态、loading、防重复点击和失败回滚。

### 3.4 通知与实时体验不足

当前问题：

- SSE 存在重复连接、固定重连、后台断开、权限提示缺失等风险。
- Android 13+ 通知权限没有完整请求体验。
- 底部栏未充分使用通知计数。

目标状态：

- 通知中心类似 X 的 Notifications：互动、关注、系统信号、告警分组展示。
- 前台通过 SSE 实时更新；App 回前台补拉未读和最近通知。
- 后台通知需要运行时权限说明；无权限时给出设置引导。
- 未读数在顶部或底部入口真实展示。

### 3.5 工程基础不足以支撑高频迭代

当前问题：

- ViewModel 直接拼 RPC path。
- `ConnectTransportProvider` 是全局单例且使用 `runBlocking`。
- Token 明文存储。
- 登录/注册密码输入按产品要求保持明文可见，但必须禁止本地保存、日志输出和崩溃上报采集。
- Room 使用破坏性迁移。
- 没有发现测试文件。

目标状态：

- 社交功能必须按 `Screen -> ViewModel -> Repository -> RpcClient -> Connect-RPC` 落地。
- UI 状态必须统一四态或 `AsyncState`。
- 用户操作事件使用 `SharedFlow<UiEvent>`，不要把一次性提示塞进持久状态。
- 关键 ViewModel 和 Repository 必须有单元测试。

## 四、信息架构

### 4.1 一级导航

一级导航建议固定为五个入口：

| Tab | 对标 | 职责 |
|---|---|---|
| 首页 | X Home / Threads For You | 推荐、关注、信号混合信息流 |
| 发现 | X Explore | 搜索、趋势品种、热门交易员、热门话题 |
| 发布 | X Compose | 发帖、引用信号、附加品种/观点 |
| 通知/消息 | X Notifications / Messages | 通知中心与私聊入口，可按阶段拆分 |
| 我的 | X Profile | 当前用户主页、设置、MT 账户、告警 |

如果当前底部栏没有中心发布按钮，不得命名为 `BottomNavBarWithFAB`。若要对标 X / Threads，应明确发布入口是底部中心按钮或首页浮动按钮。

### 4.2 首页 Feed Tab

| Tab | 内容 | 空态 |
|---|---|---|
| 推荐 | 服务端推荐的用户帖子、信号、热门讨论 | 引导关注交易员或查看热门品种 |
| 关注 | 已关注用户和品种动态 | 引导去发现页关注 |
| 信号 | 系统信号和用户引用信号的讨论 | 引导添加自选品种 |
| 最新 | 时间倒序全量可见内容 | 显示刷新入口 |

Tab 未实现时必须隐藏或禁用，禁止展示无效 Tab。

## 五、核心用户体验规范

### 5.1 信息流体验

每个 Feed Item 必须包含：

- 作者：头像、昵称、用户名、认证/等级。
- 时间：相对时间，例如 `3分钟前`，禁止固定显示 `刚刚`。
- 正文：支持纯文本和交易相关短标签。
- 上下文：如关联品种、信号方向、置信度、账户事件。
- 操作区：评论、点赞、引用/转发、分享或更多。
- 点击行为：正文进入详情；作者进入 Profile；品种进入信号详情或市场页。

列表行为：

- 首屏加载显示骨架屏或进度态。
- 下拉刷新必须保留用户当前位置或给出新内容提示。
- 分页失败只影响底部加载区，不清空已有内容。
- 部分 item 渲染失败不得导致整页崩溃。

### 5.2 发布体验

发帖页面应遵守：

- 输入框自动聚焦，显示字数限制。
- 支持关联品种或信号。
- 发送按钮在空内容或提交中禁用。
- 发布成功后返回上一页或插入信息流顶部。
- 发布失败保留草稿并显示可重试错误。

AI Agent 不得为了演示效果硬编码帖子或伪造发布成功；必须走后端 RPC。若 RPC 不存在，先补后端协议或向用户确认。

### 5.3 互动体验

点赞、关注、收藏等高频操作建议使用乐观更新：

```text
用户点击
  -> UI 立即更新
  -> 调用 RPC
  -> 成功保持
  -> 失败回滚并显示 Snackbar
```

必须处理：

- 重复点击防抖。
- 未登录或 token 过期。
- 403 无权限。
- 网络失败回滚。

### 5.4 错误与空态

禁止：

- `catch (_: Exception) {}` 静默失败。
- 请求失败后用硬编码假数据填充。
- 空态按钮 `onClick = {}`。

必须：

- 错误态显示用户可理解文案。
- 可恢复错误提供重试。
- 空态提供下一步动作。
- 调试日志记录原始异常。

### 5.5 视觉原则

对标 Threads 的低噪声阅读体验：

- 信息流背景简洁，卡片边界弱化。
- 文字层级清晰，不滥用高饱和颜色。
- 交易方向颜色统一：看多、看空、中性不得在不同页面随意变化。
- 深色模式必须跟随系统或可配置，禁止强制 `darkTheme = false`。

## 六、工程落地规范

### 6.1 推荐目录

社交模块建议按职责拆分：

```text
frontend/android/app/src/main/java/com/antclaw/alfq/
├── data/
│   ├── rpc/
│   │   ├── FeedRpc.kt
│   │   ├── SocialRpc.kt
│   │   ├── ProfileRpc.kt
│   │   └── NotificationRpc.kt
│   ├── repository/
│   │   ├── FeedRepository.kt
│   │   ├── SocialRepository.kt
│   │   └── ProfileRepository.kt
│   └── local/
│       ├── FeedCacheDao.kt
│       └── SocialDraftStore.kt
├── ui/
│   ├── feed/
│   │   ├── FeedScreen.kt
│   │   ├── FeedViewModel.kt
│   │   ├── FeedModels.kt
│   │   └── FeedItemView.kt
│   ├── post/
│   ├── social/
│   ├── profile/
│   └── components/
└── navigation/
    ├── Routes.kt
    ├── MainNavGraph.kt
    └── BottomNavBar.kt
```

文件职责以单一概念为准，不按机械行数拆分。

### 6.2 数据流

标准链路：

```text
Screen
  -> ViewModel
  -> Repository
  -> Rpc 封装
  -> Connect-RPC
  -> Repository 映射为 UI 模型
  -> ViewModel 更新 UiState
  -> Screen 渲染
```

禁止：

- ViewModel 直接拼 RPC path。
- Screen 直接访问 Repository 或 RPC。
- UI 层创建假数据 fallback。
- 手动编辑生成代码。

### 6.3 状态模型

复杂页面必须使用统一状态：

```text
idle
loading
success(data)
empty(reason)
error(message)
```

列表页建议拆分：

```text
initialLoadState
refreshing
appendState
items
lastUpdatedAt
```

一次性事件：

```text
Snackbar
Navigate
OpenSheet
RequireLogin
```

必须通过事件流处理，不要长期保存在 UiState 中。

### 6.4 Repository 规则

Repository 负责：

- 调用 Rpc。
- 合并本地缓存和远端结果。
- 将 proto 转为 UI/domain 模型。
- 处理分页游标。
- 抛出或返回明确错误。

Repository 不负责：

- Compose 状态。
- 导航。
- Snackbar 文案展示。

### 6.5 缓存策略

Feed 和通知是高频页面，应支持本地缓存：

| 数据 | 缓存 | 说明 |
|---|---|---|
| Feed 首屏 | Room | 弱网时可展示旧内容并提示刷新失败 |
| 通知 | Room | 已读状态本地同步，前台补拉校正 |
| 草稿 | DataStore 或 Room | 发布失败保留 |
| Token | 加密存储 | 禁止明文 refresh token |

缓存不得伪造服务端不存在的数据。

## 七、接口落地优先级

### 7.1 P0：社交基础闭环

- 登录态稳定：登录、刷新、登出、session expired。
- 当前用户：获取 `me`。
- Feed 列表：推荐/关注至少一个真实列表。
- 帖子详情：查看单条内容和回复列表。
- 个人主页内容 Tab：至少展示该用户真实帖子列表。
- 发布帖子：文本发布。
- 点赞：乐观更新和失败回滚。
- 关注/取关：Profile 和 Feed 作者卡可用。
- 通知未读：Badge 与通知中心。

### 7.2 P1：对标体验增强

- 引用/转发。
- 评论回复树或对话串。
- 搜索用户、帖子、品种。
- 趋势话题和热门品种。
- 关联信号发帖。
- 下拉刷新和分页。

### 7.3 P2：交易员差异化

- 交易表现认证展示。
- MT 账户可信标识。
- 信号讨论区。
- 告警触发自动生成讨论入口。
- 高质量交易员推荐。

## 八、当前代码整改清单

### 8.1 必须先整改

- [x] 登录/注册密码输入按产品要求保持明文可见，禁止改为默认隐藏；安全性依赖 HTTPS/TLS 传输、服务端哈希存储、禁止客户端日志采集密码。
- [x] Release 签名信息移出仓库配置。
- [x] Release 禁用 cleartext traffic。
- [x] Token 加密存储，至少 refresh token 加密。
- [x] 移除或限制 `fallbackToDestructiveMigration()`。
- [x] `AlfQTheme` 支持系统深色模式。
- [x] Feed Tab 无实现前隐藏或接入真实状态。
- [x] 空态按钮必须可点击并执行有效动作。

### 8.2 社交基础设施

- [x] 拆分 `AuthManager` 或 `SessionViewModel` 管理登录态。 → `ui/session/SessionViewModel.kt`
- [x] `ConnectTransportProvider` 改为可注入、可测试的 RPC 客户端工厂。 → tokenProvider 改为 suspend lambda
- [x] OkHttpClient 单例复用。
- [x] 401 refresh 串行化，失败后通知 UI 跳登录。 → ConnectTransportProvider.refreshTokenBlocking() + SessionViewModel.onSessionExpired()
- [ ] SSE 连接由会话统一管理，支持幂等 connect 和指数退避。
- [x] ViewModel 不再直接拼 RPC path。 → FeedRpc/ProfileRpc/NotificationRpc/SearchRpc/TrendRpc 封装

### 8.3 Feed / Social 改造

- [x] 建立 `FeedRepository` 和 `FeedRpc`。 → SocialRpc.kt 重构为 FeedRpc + ProfileRpc + NotificationRpc + SearchRpc + TrendRpc
- [x] 建立统一 `FeedItemUiModel`。 → PostUi 新增 originalPostId，SocialModel 新增 TraderProfileUi/UserInfoUi
- [x] 首页支持真实 Tab 状态。
- [x] 支持刷新、分页、错误重试。
- [x] 点赞/关注操作有乐观更新和失败回滚。
- [x] 帖子详情页展示回复与关联信号。 → PostDetailViewModel 新增 loadComments/loadMoreComments

### 8.4 测试与验收

- [ ] `AuthManager` / token refresh 单元测试。
- [x] `FeedViewModel` 首屏加载、刷新、分页、错误测试。 → `FeedViewModelTest.kt` (5 tests)
- [x] 点赞/关注乐观更新测试。 → `FeedViewModelTest` includes like rollback test
- [ ] `SseManager` 解析、重连、401 场景测试。
- [ ] Compose 登录页、Feed 空态、错误态、成功态测试。

## 九、AI Agent 开发流程

每次实现社交功能前，AI Agent 必须按以下顺序执行：

1. 阅读本文，确认功能属于 P0 / P1 / P2。
2. 阅读 `07-文档质量评审与落地边界.md`，确认不是管理端能力。
3. 在 `01-API接口总览.md` 查询 Service / RPC。
4. 在 `02-API字段参考.md` 查询字段。
5. 检查当前 Android 代码是否已有同类 Repository / Rpc / ViewModel 模式。
6. 先补 Rpc 封装和 Repository，再写 ViewModel，最后写 Screen。
7. 为错误态、空态、加载态、成功态分别实现 UI。
8. 增加或更新测试。
9. 运行构建和测试。
10. 如果涉及文档验收项，更新对应 `[ ]` 为 `[x]`，并写明依据。

如果接口字段、分页参数、排序策略、缓存 TTL、推荐算法含义不明确，不能猜测，必须先提问或查后端实现。

客户端 Agent 不得直接修改服务端代码；发现服务端能力缺口时，记录到联调问题或交给服务端 Agent 按 `09-服务端配套改造优化文档.md` 处理。

## 十、验收标准

### 10.1 产品验收

- [ ] 用户打开 App 后能看到真实社交信息流，而不是固定信号看板。
- [ ] 用户可以发布一条真实帖子。
- [ ] 用户可以点赞并看到计数变化。
- [ ] 用户可以进入帖子详情并看到回复或详情内容。
- [ ] 用户可以关注/取关另一个用户。
- [ ] 用户可以进入个人主页查看帖子、指标和关注关系。
- [ ] Feed 有空态、错误态、加载态、分页态。
- [ ] 通知未读数准确显示，前台通知可实时更新。

### 10.2 工程验收

- [ ] ViewModel 不直接拼 RPC path。
- [ ] 无硬编码假数据 fallback。
- [ ] 用户操作失败有反馈。
- [ ] 关键社交 ViewModel 有单元测试。
- [ ] Release 不包含硬编码签名密码。
- [ ] Release 不允许 cleartext traffic。
- [ ] Token 不以明文形式保存 refresh token。
- [ ] Room migration 不使用生产破坏性迁移。

### 10.3 对标验收

- [ ] 信息流阅读体验接近 Threads：简洁、低噪声、内容优先。
- [ ] 互动效率接近 X：点赞、评论、转发/引用入口清晰。
- [ ] 发现页能帮助用户找到交易员、话题或品种。
- [ ] 个人主页能建立交易员可信度，而不只是普通资料页。
- [ ] 通知中心能支撑高频社交反馈。

## 十一、禁止事项

- 禁止用硬编码帖子、随机用户、随机点赞数模拟社交繁荣。
- 禁止把管理端接口做进 Android。
- 禁止新增 REST 客户端、Retrofit、WebSocket 作为主业务通道。
- 禁止无效按钮、无效 Tab、点击无反馈。
- 禁止吞错。
- 禁止 Screen 直接访问 RPC。
- 禁止为了视觉效果牺牲真实错误态。

## 十二、推荐里程碑

### M-Social-0：基础设施止血

- 修复安全配置、明文密码输入边界、深色模式、Token 存储。
- 建立 Session / RPC / 错误处理基础。
- 建立统一 `AsyncState` 和 `UiEvent`。

### M-Social-1：真实 Feed

- 接入真实 Feed RPC。
- 替换硬编码信号看板为信息流。
- 实现刷新、分页、Tab。

### M-Social-2：发布与互动

- 实现发帖。
- 实现点赞、评论入口、帖子详情。
- 实现乐观更新和失败回滚。

### M-Social-3：关系与主页

- 实现关注/取关。
- 完善 Profile 内容 Tab。
- 展示交易员可信指标。

### M-Social-4：通知与发现

- 完善通知中心和未读 Badge。
- 完善发现页搜索、趋势品种、热门交易员。
- 优化 SSE 与前后台补拉。

### M-Social-5：体验打磨与测试

- 补全 Compose UI 测试。
- 补全 ViewModel / Repository 单测。
- 性能优化、无障碍、深色模式、弱网体验。

## 十三、后端依赖与接口矩阵

Android Agent 开工前必须先确认后端分支已同步 `09-服务端配套改造优化文档.md` 的 P0 基线。客户端不得在服务端能力缺失时伪造数据。

| 客户端能力 | 后端依赖 | Android 接入层 | P0/P1 | 失败处理 |
|---|---|---|---|---|
| Feed 首页列表 | `FeedService/GetFeed` | `FeedRpc.getFeed` -> `FeedRepository.getFeed` | P0 | 显示错误态和重试，不填充假帖子 |
| Feed Tab 过滤 | `GetFeedRequest.filter` | `FeedTab.toFilter()` | P0 | 后端不支持时隐藏对应 Tab 或提示暂不可用 |
| 帖子详情 | `FeedService/GetPost` | `SocialRpc.getPost` -> `PostDetailViewModel` | P0 | `NotFound` 显示内容不存在 |
| 评论列表 | `FeedService/ListComments` | `CommentRpc.listComments` 或 `SocialRpc.listComments` | P0 | 失败仅影响评论区，不清空帖子详情 |
| 发布评论 | `FeedService/CommentOnPost` | `SocialRepository.commentOnPost` | P0 | 保留输入内容并 Snackbar 提示 |
| 用户帖子 Tab | `FeedService/ListUserPosts` | `ProfileRepository.listUserPosts` | P0 | Profile 资料仍显示，帖子 Tab 显示错误态 |
| 发布帖子 | `FeedService/CreatePost` | `ComposePostViewModel` | P0 | 保留草稿，不伪造发布成功 |
| 点赞 / 取消点赞 | `LikePost` / `UnlikePost` | `FeedViewModel.toggleLike` | P0 | 乐观更新失败回滚 |
| 关注 / 取关 | `TraderService/Follow` / `Unfollow` | `ProfileViewModel.toggleFollow` | P0 | 乐观更新失败回滚 |
| 关注状态 | `TraderProfile.is_following` | `ProfileRepository.getProfile` | P0 | 未返回时显示未知状态并禁用按钮 |
| 未读通知 | `NotificationService/UnreadCount` | `NotificationViewModel` | P0 | Badge 隐藏或显示重试状态 |
| 实时通知 | `/sse/notifications` | `SseManager` / session layer | P1 | 断线指数退避，前台补拉 |
| 搜索 | `SearchService/Search` | `DiscoverRepository.search` | P1 | 搜索页错误态 |
| 趋势 | `TrendService/ListTrendingTopics`、`ListHotSymbols` | `DiscoverRepository` | P1 | 模块级错误态 |
| 推荐交易员 | `TraderService/ListRecommendedTraders` | `DiscoverRepository` | P2 | 未实现时隐藏模块 |

## 十四、详细实施步骤

### 14.1 开工前检查

每次实现社交功能前执行：

1. 确认当前分支已拉取最新代码。
2. 检查 `proto/antclaw/v1/alfq_feed.proto` 是否包含：
   - `ListComments`
   - `ListUserPosts`
   - `Comment.parent_comment_id`
   - `Post.original_post_id`
3. 检查 `proto/antclaw/v1/alfq_trader.proto` 是否包含：
   - `TraderProfile.is_following`
4. 执行或确认 Android proto 生成链路可用。
5. 检查 `BuildConfig.BASE_URL` 是否指向：

```text
https://api.alfq.org/
```

6. 确认没有新增 `fetch`、`axios`、Retrofit 或 WebSocket 主业务通道。

### 14.2 Feed 首页实施步骤

目标：Feed 首页使用真实社交流，不使用硬编码数据。

实施顺序：

1. 在 `data/rpc/` 建立或补齐 `FeedRpc` / `SocialRpc` 封装。
2. 在 `data/repository/` 建立或补齐 `FeedRepository`。
3. 建立 `FeedItemUiModel`，统一承载：
   - 帖子 ID
   - 作者 ID
   - 作者名
   - 内容
   - 类型
   - 关联信号
   - 点赞数
   - 评论数
   - 转发数
   - 当前用户是否点赞
   - 创建时间
   - 原帖 ID
4. `FeedViewModel` 只调用 Repository，不直接拼 RPC path。
5. `FeedScreen` 渲染：
   - 首屏 loading
   - 空态
   - 错误态
   - 成功列表
   - 底部分页 loading / error
6. Tab 切换时映射 filter：

| Tab | filter |
|---|---|
| 推荐 | `all` |
| 信号 | `signals_only` |
| 最新 | `all`，由服务端时间倒序返回 |

关注 Tab 只有在后端提供关注流或明确 filter 时才展示。未实现前不得显示无效 Tab。

### 14.3 帖子详情与评论实施步骤

目标：帖子详情页展示真实帖子和真实评论列表。

实施顺序：

1. `PostDetailViewModel.load(postId)` 并发或串行调用：
   - `GetPost`
   - `ListComments`
2. 评论区使用独立分页状态：
   - `initialLoadState`
   - `comments`
   - `nextCursor`
   - `appendState`
3. 发布评论调用 `CommentOnPost`。
4. 发布成功后：
   - 将返回评论插入本地列表，或重新刷新评论第一页。
   - 清空输入框。
5. 发布失败后：
   - 保留输入内容。
   - Snackbar 显示错误。

P0 评论可以先扁平展示。P1 再按 `parent_comment_id` 组装回复树。

### 14.4 个人主页实施步骤

目标：Profile 页面展示真实资料、关注状态和该用户内容 Tab。

实施顺序：

1. `ProfileRepository.getProfile(userId)` 调用 `TraderService/GetProfile`。
2. `ProfileRepository.listUserPosts(userId, filter, cursor)` 调用 `FeedService/ListUserPosts`。
3. `ProfileViewModel` 页面状态拆分：
   - `profileState`
   - `postsState`
   - `selectedTab`
   - `followActionState`
4. 关注按钮使用乐观更新：

```text
保存 previous state
立即切换 isFollowing 与 followerCount
调用 Follow / Unfollow
失败则恢复 previous state 并 Snackbar
```

5. Profile 内容 Tab P0 至少包含：
   - 帖子
   - 信号

回复、账户、交易指标可作为 P1/P2。

### 14.5 发布入口实施步骤

目标：用户可以发布真实文本帖子。

实施顺序：

1. 建立 `ComposePostScreen` 和 `ComposePostViewModel`。
2. 输入框支持：
   - 自动聚焦
   - 字数统计
   - 空内容禁用发布按钮
   - 提交中禁用重复点击
3. 调用 `FeedService/CreatePost`。
4. 成功后：
   - 返回 Feed 并触发刷新，或把返回的 Post 插入 Feed 顶部。
5. 失败后：
   - 保留草稿。
   - 显示错误。

P0 只做文本帖。关联信号、引用、可见性高级配置放到 P1。

### 14.6 通知与 SSE 实施步骤

目标：未读数准确，前台通知实时更新。

实施顺序：

1. `NotificationViewModel` 在 App 前台调用 `UnreadCount`。
2. `SseManager` 由 Session 层统一管理，不由页面重复创建。
3. 登录后连接 SSE，登出后断开。
4. 前台收到事件后：
   - 更新未读数。
   - 必要时刷新通知列表。
5. App 回前台时补拉未读数，修正 SSE 断线期间的状态。
6. 401 时停止重连并触发重新登录事件。

禁止使用 `setInterval` 轮询替代 SSE。

## 十五、关键集成点

### 15.1 Session 与 Token

必须保证：

- access token 注入所有 Connect-RPC 请求。
- refresh token 加密存储。
- 当前用户 ID 在登录 / 注册成功后保存，用于 `liked_by` 映射和“我的主页”。
- 401 refresh 必须串行化，避免多个请求同时刷新 token。
- refresh 失败后清理会话并跳转登录。

### 15.2 Connect-RPC 客户端

必须保证：

- RPC 客户端可注入、可替换、可测试。
- OkHttpClient 单例复用。
- Repository 只依赖 RPC 封装，不依赖 Compose。
- ViewModel 不直接构造 RPC path。

### 15.3 Navigation

至少定义以下路由：

```text
feed
postDetail/{postId}
composePost
profile/{userId}
notifications
discover
settings
```

点击行为：

- Feed item 正文：进入 `postDetail/{postId}`。
- 作者头像 / 名称：进入 `profile/{userId}`。
- 评论按钮：进入详情并定位评论区。
- 发布按钮：进入 `composePost`。
- 通知 item：按通知类型进入帖子、Profile、告警或系统页面。

### 15.4 UI 状态和错误文案

推荐错误映射：

| 场景 | 用户文案 |
|---|---|
| `Unauthenticated` | 登录已过期，请重新登录 |
| `PermissionDenied` | 你没有权限执行此操作 |
| `NotFound` | 内容不存在或已被删除 |
| `InvalidArgument` | 请求参数无效，请重试 |
| 网络失败 | 网络连接失败，请检查后重试 |
| 服务端错误 | 服务暂时不可用，请稍后再试 |

调试日志可以记录原始异常，但不得记录 token、密码、refresh token。

## 十六、测试流程

### 16.1 本地构建

Android 修改后必须执行：

```bash
cd frontend/android
./gradlew :app:compileDebugKotlin
```

涉及资源、Manifest、依赖或打包时执行：

```bash
cd frontend/android
./gradlew :app:assembleDebug
```

### 16.2 单元测试

新增或修改 ViewModel / Repository 后执行：

```bash
cd frontend/android
./gradlew :app:testDebugUnitTest --no-daemon --max-workers=2
```

最低测试覆盖：

- Feed 首屏加载成功。
- Feed 首屏加载失败。
- Feed 刷新失败不清空旧数据。
- Feed 分页失败只影响 append 状态。
- 点赞乐观更新成功。
- 点赞失败回滚。
- 关注乐观更新成功。
- 关注失败回滚。
- 评论发布失败保留草稿。
- 401 触发登录过期事件。

### 16.3 手工联调测试

P0 手工验收步骤：

1. 使用真实账号登录。
2. 打开 Feed，确认不是硬编码数据。
3. 下拉刷新，确认数据稳定。
4. 滑到底部触发分页，确认不重复、不跳项。
5. 点赞一条帖子，退出重进后状态仍正确。
6. 打开帖子详情，确认评论列表来自后端。
7. 发布评论，刷新后评论仍存在。
8. 点击作者进入 Profile。
9. 关注 / 取关作者，确认粉丝数和状态变化。
10. 打开 Profile 内容 Tab，确认显示该用户真实帖子。
11. 打开通知中心，确认未读数与 Badge 一致。
12. 断网后重试，确认有明确错误态而不是假数据。

### 16.4 回归检查

每次提交前检查：

- 没有新增硬编码社交数据。
- 没有 `catch (_: Exception) {}`。
- 没有 ViewModel 直接拼 RPC path。
- 没有新增无效 Tab 或无效按钮。
- 没有把服务端缺口在客户端用假数据绕过。
- 文档 checklist 与实际代码一致。

## 十七、故障排查指南

### 17.1 Feed 为空

排查顺序：

1. 确认账号已登录且 token 有效。
2. 确认 `BASE_URL` 是 `https://api.alfq.org/`。
3. 查看 `GetFeed` 是否返回错误。
4. 检查 filter 是否为后端支持值。
5. 若后端返回空列表，显示空态，不造数据。

### 17.2 点赞状态不准确

排查顺序：

1. 确认登录 / 注册后已保存当前 `userId`。
2. 确认 `Post.liked_by` 是否包含当前用户 ID。
3. 确认 Repository 映射 `isLiked` 时使用当前用户 ID。
4. 确认乐观更新失败时有回滚。

### 17.3 Profile 关注状态不准确

排查顺序：

1. 确认 `TraderProfile.is_following` 已生成到 Android proto。
2. 确认 `GetProfile` 请求携带 Authorization。
3. 确认 Follow / Unfollow 返回真实 `follower_count`。
4. 确认失败时恢复 previous state。

### 17.4 评论无法显示

排查顺序：

1. 确认 Android proto 包含 `ListComments`。
2. 确认后端已注册 `FeedService/ListComments`。
3. 确认 `post_id` 正确。
4. 确认评论区错误不会覆盖帖子详情。

### 17.5 SSE 重复连接或断线

排查顺序：

1. 确认 SSE 由 Session 层统一管理。
2. 确认页面没有在每次重组时创建连接。
3. 确认后台断开、前台重连。
4. 确认 401 不无限重连。
5. 确认日志不输出 token。

### 17.6 编译找不到 proto 方法

排查顺序：

1. 确认后端 proto 已更新并推送。
2. 确认 Android 已拉取最新 proto。
3. 清理并重新生成 proto。
4. 重新执行：

```bash
cd frontend/android
./gradlew :app:compileDebugKotlin
```

### 17.7 页面出现假数据或演示数据

处理要求：

1. 立即移除硬编码 fallback。
2. 改为 loading / empty / error 状态。
3. 如果是后端缺口，记录到 `09-服务端配套改造优化文档.md` 对应项。
4. 不允许为了通过 UI 验收保留假数据。
