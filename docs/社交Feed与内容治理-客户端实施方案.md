# 社交Feed与内容治理 — 客户端实施方案

> 基于 X (Twitter) 客户端布局与交互对标  
> 适用范围：`frontend/android/app/src/main/java/com/antclaw/alfq/`  
> 版本：v2.0 | 日期：2026-05-16

---

## 零、导航体系（对标 X 完整设计）

### 0.1 页面对照表

| X 页面 | AntClaw 路由 | 文件 | 状态 |
|---|---|---|---|
| 首页 (Home) | `feed` | `ui/feed/FeedScreen.kt` | ⚠️ 缺"关注" Tab |
| 搜索/发现 (Explore) | `discover` | `ui/discover/DiscoverScreen.kt` | ✅ |
| 发贴 (+) | `post` | `ui/post/PostScreen.kt` | ✅ |
| 通知 (Notifications) | `notifications` | `ui/notification/NotificationCenterScreen.kt` | ⚠️ 缺分类过滤 |
| 私信 (Messages) | `chat` | `ui/chat/ChatScreen.kt` | ✅ |
| 帖子详情 | `postDetail/{postId}` | `ui/post/PostDetailScreen.kt` | ✅ |
| 个人主页 | `profile/{userId}` | `ui/profile/ProfileScreen.kt` | ⚠️ 缺 Media Tab |
| 设置 | `settings/language` | `ui/settings/LanguagePickerScreen.kt` | ✅ |
| 信号详情 | `signal/{pair}` | `ui/feed/SignalDetailScreen.kt` | ✅ |

### 0.2 导航规则（对标 X 行为）

```kotlin
// AlfQNavigation.kt — 底部栏导航常量
object NavRules {
    // 每个 Tab 维护独立的返回栈
    const val ROOT = "feed"

    // 从二级页点底部 Tab → 回到该 Tab 的根页面
    fun navigateToTab(nav: NavController, route: String) {
        nav.navigate(route) {
            popUpTo(ROOT) { inclusive = false; saveState = true }
            launchSingleTop = true
            restoreState = true
        }
    }

    // 从通知/个人页点帖子 → 目标 Tab 下新建栈
    fun navigateToPost(nav: NavController, postId: String) {
        nav.navigate("postDetail/$postId")
        // 不做 popUpTo，保留返回栈
    }
}
```

**核心行为**：
- 底部 5 Tab 之间切换：`popUpTo("feed") { inclusive=false }`，每个 Tab 保留自己的滚动位置
- Tab 内的二级页（帖子详情、个人主页）：正常 push，返回键回到 Tab 根
- 从通知点击跳帖子：push 到当前栈，返回回到通知页
- 底部栏始终在 5 个主 Tab 显示，进入二级页隐藏

---

## 一、首页 (Feed) — 对标 X Home

### 1.1 Tab 布局

```
┌──────────────────────────┐
│  [关注]  [发现探索]  [关注中]  │  ← ScrollableTabRow，X 风格
├──────────────────────────┤
│  帖子流 (LazyColumn)     │
│  ┌────────────────────┐  │
│  │ 头像  用户名 @codeId  │  │  ← PostCard，X 风格
│  │       2小时前          │  │
│  │  帖子内容...          │  │
│  │  💬12  🔄5  ❤️34  📊 │  │
│  └────────────────────┘  │
│  ┌────────────────────┐  │
│  │ ...                 │  │
│  └────────────────────┘  │
└──────────────────────────┘
```

**文件**：`ui/feed/FeedScreen.kt`

```kotlin
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun FeedScreen(
    viewModel: FeedViewModel = hiltViewModel(),
    notificationCount: Int = 0,
    onPostClick: (String) -> Unit = {},
    onAuthorClick: (String) -> Unit = {},
    onNotificationClick: () -> Unit = {},
    onSearchClick: () -> Unit = {},
) {
    val state by viewModel.uiState.collectAsStateWithLifecycle()
    val tabs = listOf(
        FeedTab.FOLLOWING to stringResource(R.string.feed_tab_following),
        FeedTab.RECOMMENDED to stringResource(R.string.feed_tab_recommended),
        FeedTab.SIGNALS to stringResource(R.string.feed_tab_signals),
    )

    Scaffold(topBar = {
        // X 风格顶部：用户头像(点击进"我") + Logo + 搜索/通知图标
        TopAppBar(
            title = { Text("AntClaw", fontWeight = FontWeight.Bold) },
            navigationIcon = {
                // 点击头像 → "我" 页面（ProfileScreen with userId=session）
                IconButton(onClick = { onAuthorClick("me") }) {
                    // Avatar placeholder
                    Surface(Modifier.size(32.dp), CircleShape) { Box(contentAlignment = Alignment.Center) {
                        Text("A") // 首字母
                    }}
                }
            },
            actions = {
                IconButton(onClick = onSearchClick) { Icon(Icons.Default.Search, "搜索") }
                // 通知铃铛 + 红点
                BadgedBox(badge = { if (notificationCount > 0) Badge { Text("$notificationCount") } }) {
                    IconButton(onClick = onNotificationClick) { Icon(Icons.Default.Notifications, "通知") }
                }
            },
        )
    }) { padding ->
        Column(Modifier.padding(padding)) {
            // Tab 栏
            ScrollableTabRow(
                selectedTabIndex = state.currentTab.ordinal,
                edgePadding = 0.dp,
                indicator = { tabPositions ->
                    TabRowDefaults.SecondaryIndicator(
                        Modifier.tabIndicatorOffset(tabPositions[state.currentTab.ordinal]),
                        color = MaterialTheme.colorScheme.primary,
                    )
                },
            ) {
                tabs.forEachIndexed { idx, (tab, label) ->
                    Tab(
                        selected = state.currentTab == tab,
                        onClick = { viewModel.selectTab(tab) },
                        text = {
                            Text(label,
                                fontWeight = if (state.currentTab == tab) FontWeight.Bold else FontWeight.Normal,
                                color = if (state.currentTab == tab) MaterialTheme.colorScheme.onSurface
                                        else MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f),
                            )
                        },
                    )
                }
            }

            // 内容区
            when (state.phase) {
                AsyncPhase.Loading -> Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) { CircularProgressIndicator() }
                else -> {
                    val listState = rememberLazyListState()
                    LazyColumn(state = listState) {
                        items(state.posts, key = { it.postId }) { post ->
                            PostCard(post = post,
                                onPostClick = onPostClick,
                                onAuthorClick = onAuthorClick,
                                onLikeClick = { viewModel.toggleLike(post.postId) },
                                onShareClick = { viewModel.sharePost(post.postId) },
                            )
                        }
                        if (state.hasMore) {
                            item {
                                Box(Modifier.fillMaxWidth().padding(16.dp), contentAlignment = Alignment.Center) {
                                    CircularProgressIndicator(Modifier.size(24.dp))
                                }
                            }
                            LaunchedEffect(listState) {
                                snapshotFlow { listState.layoutInfo.visibleItemsInfo.lastOrNull()?.index }
                                    .collect { if (it != null && it >= state.posts.size - 3) viewModel.loadMore() }
                            }
                        }
                    }
                }
            }
        }
    }
}
```

### 1.2 字符串资源

```xml
<!-- res/values/strings.xml -->
<string name="feed_tab_following">关注</string>
<string name="feed_tab_recommended">发现探索</string>
<string name="feed_tab_signals">关注中</string>
```

---

## 二、帖子卡片 (PostCard) — 对标 X Tweet

### 2.1 完整卡片布局

```kotlin
@Composable
fun PostCard(
    post: PostUi,
    onPostClick: (String) -> Unit = {},
    onAuthorClick: (String) -> Unit = {},
    onLikeClick: () -> Unit = {},
    onShareClick: () -> Unit = {},
    onReportClick: () -> Unit = {},
) {
    var showMenu by remember { mutableStateOf(false) }

    Card(
        modifier = Modifier
            .fillMaxWidth()
            .clickable { onPostClick(post.postId) },
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
        elevation = CardDefaults.cardElevation(defaultElevation = 0.dp),
    ) {
        Row(Modifier.padding(12.dp)) {
            // 左列：头像
            Surface(
                modifier = Modifier.size(40.dp).clickable { onAuthorClick(post.authorId) },
                shape = CircleShape,
                color = MaterialTheme.colorScheme.primaryContainer,
            ) {
                Box(contentAlignment = Alignment.Center) {
                    Text(post.authorName.take(1).uppercase(), fontWeight = FontWeight.Bold)
                }
            }

            Spacer(Modifier.width(12.dp))

            // 右列：内容
            Column(Modifier.weight(1f)) {
                // 头部：用户名 + handle + 时间 + 菜单
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Text(post.authorName, fontWeight = FontWeight.Bold, style = MaterialTheme.typography.bodyMedium)
                    Spacer(Modifier.width(4.dp))
                    Text("@${post.codeId.ifEmpty { post.authorId.take(8) }}",
                        color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f),
                        style = MaterialTheme.typography.bodySmall)
                    Text(" · ${timeAgo(post.createdAt)}",
                        color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.4f),
                        style = MaterialTheme.typography.bodySmall)
                    Spacer(Modifier.weight(1f))
                    Box {
                        IconButton(onClick = { showMenu = true }, modifier = Modifier.size(24.dp)) {
                            Icon(Icons.Default.MoreVert, "更多", modifier = Modifier.size(16.dp))
                        }
                        DropdownMenu(expanded = showMenu, onDismissRequest = { showMenu = false }) {
                            DropdownMenuItem(text = { Text("举报") }, onClick = { onReportClick(); showMenu = false })
                            DropdownMenuItem(text = { Text("不感兴趣") }, onClick = { showMenu = false })
                        }
                    }
                }

                Spacer(Modifier.height(4.dp))

                // 正文
                Text(post.content, style = MaterialTheme.typography.bodyLarge)

                // 关联品种标签 (如果有 signalPair)
                if (post.signalPair.isNotEmpty()) {
                    Spacer(Modifier.height(4.dp))
                    SuggestionChip(
                        onClick = { },
                        label = { Text(post.signalPair, style = MaterialTheme.typography.labelSmall) },
                        modifier = Modifier.height(24.dp),
                    )
                }

                Spacer(Modifier.height(8.dp))

                // 互动按钮行
                Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
                    // 评论
                    ActionButton(icon = Icons.Default.ChatBubbleOutline, text = "${post.commentCount}", onClick = { onPostClick(post.postId) })
                    // 分享
                    ActionButton(icon = Icons.Default.Repeat, text = "${post.shareCount}", onClick = onShareClick)
                    // 点赞
                    ActionButton(
                        icon = if (post.isLiked) Icons.Default.Favorite else Icons.Default.FavoriteBorder,
                        text = "${post.likeCount}",
                        onClick = onLikeClick,
                        tint = if (post.isLiked) Color.Red else MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f),
                    )
                    // 浏览量（可选）
                    ActionButton(icon = Icons.Default.BarChart, text = "${post.viewCount}", onClick = {})
                }
            }
        }
    }
}

@Composable
private fun ActionButton(icon: ImageVector, text: String, onClick: () -> Unit, tint: Color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f)) {
    Row(Modifier.clickable(onClick = onClick).padding(4.dp), verticalAlignment = Alignment.CenterVertically) {
        Icon(icon, contentDescription = null, modifier = Modifier.size(16.dp), tint = tint)
        if (text != "0") {
            Spacer(Modifier.width(2.dp))
            Text(text, style = MaterialTheme.typography.labelSmall, color = tint)
        }
    }
}

private fun timeAgo(instant: Instant): String {
    val now = Instant.now()
    val seconds = Duration.between(instant, now).seconds
    return when {
        seconds < 60 -> "刚刚"
        seconds < 3600 -> "${seconds / 60}分钟"
        seconds < 86400 -> "${seconds / 3600}小时"
        else -> "${seconds / 86400}天"
    }
}
```

### 2.2 PostUi 扩展字段

```kotlin
// ui/social/SocialModel.kt
data class PostUi(
    // ... 现有字段 ...
    val codeId: String = "",       // 新增：作者 code_id
    val viewCount: Int = 0,        // 新增：浏览量
)
```

---

## 三、通知中心 — 对标 X Notifications

### 3.1 分类 Chip 过滤

```kotlin
@Composable
fun NotificationCenterScreen(
    onBack: () -> Unit,
    onNotificationClick: (ClientNotification) -> Unit,
    viewModel: NotificationViewModel = hiltViewModel(),
) {
    val state by viewModel.uiState.collectAsStateWithLifecycle()

    Scaffold(topBar = {
        TopAppBar(
            title = { Text("通知", fontWeight = FontWeight.Bold) },
            navigationIcon = { IconButton(onClick = onBack) { Icon(Icons.Default.ArrowBack, "返回") } },
            actions = {
                IconButton(onClick = { navController.navigate("notification_prefs") }) {
                    Icon(Icons.Default.Settings, "设置")
                }
            },
        )
    }) { padding ->
        Column(Modifier.padding(padding)) {
            // X 风格分类过滤
            ScrollableTabRow(selectedTabIndex = state.currentFilter, edgePadding = 0.dp) {
                listOf("全部", "互动", "关注", "信号").forEachIndexed { idx, label ->
                    Tab(
                        selected = state.currentFilter == idx,
                        onClick = { viewModel.setFilter(idx) },
                        text = { Text(label) },
                    )
                }
            }
            when (state.phase) {
                AsyncPhase.Loading -> Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) { CircularProgressIndicator() }
                else -> LazyColumn {
                    items(state.filteredItems, key = { it.id }) { notif ->
                        NotificationCard(notif = notif, onClick = { onNotificationClick(notif) })
                    }
                }
            }
        }
    }
}
```

---

## 四、个人主页 — 对标 X Profile

### 4.1 增强页面布局

```
┌──────────────────────────┐
│  ← 返回          ⚙ 设置  │  ← TopAppBar
├──────────────────────────┤
│  [大图封面背景]           │  ← headerImage (可选)
│       ┌────┐             │
│       │头像│  displayName │
│       └────┘  @codeId    │
│  bio 描述文字...          │
│  认证标记: ✓ 认证交易员    │
│                           │
│  100 关注  200 粉丝  3 信号│  ← TraderStatRow
│  [关注/取消关注] [分享]    │
├──────────────────────────┤
│  [帖子] [回复] [媒体] [赞] │  ← TabRow
├──────────────────────────┤
│  内容 (LazyColumn)        │
└──────────────────────────┘
```

```kotlin
@Composable
fun ProfileScreen(
    userId: String,
    viewModel: ProfileViewModel = hiltViewModel(),
    onBack: () -> Unit = {},
    onPostClick: (String) -> Unit = {},
    onSettingsClick: () -> Unit = {},
) {
    val state by viewModel.uiState.collectAsStateWithLifecycle()
    val tabs = listOf("帖子", "媒体", "赞过")

    Scaffold(topBar = { /* 同前 */ }) { padding ->
        LazyColumn(Modifier.padding(padding)) {
            // 头部 (同上版实现)
            item { ProfileHeader(state, viewModel) }
            // Tab 栏
            item { ProfileTabs(state.currentTab, tabs) { viewModel.selectTab(it) } }
            // 内容
            when (state.currentTab) {
                0 -> postsSection(state, onPostClick)
                1 -> mediaGrid(state)
                2 -> likesSection(state, onPostClick)
            }
        }
    }
}
```

---

## 五、搜索/发现 — 对标 X Explore

```kotlin
@Composable
fun DiscoverContent() {
    val vm = hiltViewModel<DiscoverViewModel>()
    LazyColumn(Modifier.fillMaxSize()) {
        item {
            // X 风格搜索栏
            OutlinedTextField(value = vm.query,
                onValueChange = { vm.search(it) },
                placeholder = { Text("搜索交易员、品种...") },
                leadingIcon = { Icon(Icons.Default.Search, null) },
                modifier = Modifier.fillMaxWidth().padding(12.dp),
                shape = RoundedCornerShape(24.dp),
            )
        }
        // 热门话题 / 推荐交易员
        items(vm.trendingTopics) { TraderCard(it, onClick = { nav("profile/${it.userId}") }) }
    }
}
```

---

## 六、实施清单

| # | 功能 | 文件 | 对标 X | 工时 |
|---|---|---|---|---|
| 1 | FeedScreen 关注 Tab + X 风格 TopBar | `FeedScreen.kt` | 首页 Tab 栏 | 1天 |
| 2 | PostCard 重构 (codeId + 时间 + 菜单 + 互动按钮) | `PostCard.kt` | Tweet 卡片 | 1天 |
| 3 | 通知分类过滤 Chip | `NotificationCenterScreen.kt` | Notifications 过滤 | 0.5天 |
| 4 | Profile Media + Likes Tab | `ProfileScreen.kt` | Profile Tabs | 0.5天 |
| 5 | Discover 搜索栏 | `DiscoverScreen.kt` | Explore 搜索 | 0.5天 |
| 6 | 导航返回栈修正 | `AlfQNavigation.kt` | Tab 独立栈 | 已 ✅ |
| 7 | 字符串资源中英补齐 | `strings.xml` | i18n | 0.5天 |
| 8 | 编译 + 测试 | — | 冒烟 | 0.5天 |

**总计：4.5 天**

---

## 七、验收

```bash
cd frontend/android && ./gradlew :app:compileDebugKotlin :app:testDebugUnitTest
adb install app/build/outputs/apk/debug/app-debug.apk
```

验收清单：
- [ ] 5 个 Tab 独立返回栈，切换不丢失滚动位置
- [ ] "关注" Tab 显示关注作者的帖子
- [ ] 帖子卡片显示 @codeId + 时间 + 互动按钮
- [ ] 通知分类过滤正常工作
- [ ] 个人页 Media/Likes Tab 数据正确
- [ ] 长按帖子显示举报菜单
- [ ] 通知点击正确跳转帖子/个人页/信号
