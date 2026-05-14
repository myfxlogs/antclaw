# 客户端 Feed 发布帖子"点击无响应"问题

> 目标读者：客户端 AI Agent（Android）
> 诊断日期：2026-05-14
> 确诊：**纯客户端问题**，服务端 `CreatePost` RPC 逻辑完整可用

---

## 一、问题现象

用户在帖子发布界面输入内容后点击「发布」按钮，没有任何响应——不报错、不提示成功、界面不变化。

## 二、根因分析

### 2.1 错误被静默吞掉（主因）

**文件**: `frontend/android/app/src/main/java/com/antclaw/alfq/ui/post/PostViewModel.kt`

```kotlin
// 当前代码（有问题）
fun post(...) {
    viewModelScope.launch {
        try {
            // ... RPC 调用 ...
            _posted.value = true
        } catch (_: Exception) { _posted.value = false }  // ← 吞掉所有异常
    }
}
```

问题：无论什么原因失败（Token 过期、网络不通、服务端报错、DB 写入失败），`catch` 块只是把 `_posted` 设回 `false`。由于初始值也是 `false`，UI 层感知不到任何状态变化 → 用户看到"什么都没发生"。

### 2.2 绕过了已有的 RPC 封装层

**文件**: `frontend/android/app/src/main/java/com/antclaw/alfq/data/rpc/FeedRpc.kt`

已存在 `FeedRpcClient` 类，封装了 `createPost()` 方法，且由 Hilt 单例注入：

```kotlin
class FeedRpcClient @Inject constructor(
    private val client: ProtocolClientInterface,
) {
    suspend fun createPost(req: CreatePostReq) { ... }
}
```

但 `PostViewModel` 没有注入 `FeedRpcClient`，而是自己裸调 `ConnectTransportProvider.createProtocolClient().unary(...)`：

```kotlin
// PostViewModel 当前做法（应改为注入 FeedRpcClient）
val spec = MethodSpec("antclaw.v1.FeedService/CreatePost", ...)
ConnectTransportProvider.createProtocolClient().unary(req, emptyMap(), spec).getOrThrow()
```

这导致：
- 每次发布都新建 `ProtocolClient` 实例（浪费）
- 无法复用统一错误处理 / 重试 / 日志

### 2.3 UI 缺少加载态和错误态

**文件**: `frontend/android/app/src/main/java/com/antclaw/alfq/ui/post/PostScreen.kt`

当前只有两个状态：
- 初始（`posted = false`）
- 成功（`posted = true` → 显示成功文字）

缺失：
- 加载中（按钮禁用 + spinner）
- 失败（SnackBar / 错误文字）

---

## 三、修复方案

### 3.1 PostViewModel 改造

```kotlin
@HiltViewModel
class PostViewModel @Inject constructor(
    private val feedRpc: FeedRpcClient,  // ← 注入已有 RPC 客户端
) : ViewModel() {

    // 用 sealed class 替换单一的 Boolean
    sealed class PostState {
        object Idle : PostState()
        object Loading : PostState()
        object Success : PostState()
        data class Error(val message: String) : PostState()
    }

    private val _postState = MutableStateFlow<PostState>(PostState.Idle)
    val postState: StateFlow<PostState> = _postState.asStateFlow()

    fun post(content: String, signalPair: String, visibility: String) {
        if (_postState.value is PostState.Loading) return  // 防重复提交
        viewModelScope.launch {
            _postState.value = PostState.Loading
            try {
                feedRpc.createPost(
                    CreatePostReq(
                        content = content,
                        post_type = if (signalPair.isBlank()) "post" else "signal",
                        signal_pair = signalPair,
                        signal_direction = "",
                        signal_confidence = 0,
                        visibility = visibility,
                    )
                )
                _postState.value = PostState.Success
            } catch (e: Exception) {
                _postState.value = PostState.Error(
                    e.message ?: "发布失败，请重试"
                )
            }
        }
    }

    fun reset() { _postState.value = PostState.Idle }
}
```

### 3.2 PostScreen 改造

```kotlin
@Composable
fun PostScreen(
    viewModel: PostViewModel = hiltViewModel(),
    onClose: () -> Unit = {},
) {
    var content by remember { mutableStateOf("") }
    var signalPair by remember { mutableStateOf("") }
    var visibility by remember { mutableStateOf("public") }

    val postState by viewModel.postState.collectAsState()

    // 成功后清空内容
    LaunchedEffect(postState) {
        if (postState is PostViewModel.PostState.Success) {
            content = ""
            signalPair = ""
        }
    }

    // 发布按钮
    val isSending = postState is PostViewModel.PostState.Loading
    Button(
        onClick = { viewModel.post(content, signalPair, visibility) },
        enabled = content.isNotBlank() && !isSending,
    ) {
        if (isSending) {
            CircularProgressIndicator(modifier = Modifier.size(16.dp), strokeWidth = 2.dp)
        } else {
            Text(stringResource(R.string.post_publish))
        }
    }

    // 成功提示
    if (postState is PostViewModel.PostState.Success) {
        Text(stringResource(R.string.post_success), color = MaterialTheme.colorScheme.primary)
    }

    // 错误提示
    if (postState is PostViewModel.PostState.Error) {
        Text(
            (postState as PostViewModel.PostState.Error).message,
            color = MaterialTheme.colorScheme.error,
        )
    }
}
```

---

## 四、验证方式

修复后，按以下场景逐项验证：

| 场景 | 预期行为 |
|---|---|
| 已登录，正常发布文字帖 | 按钮变 spinner → 成功后清空输入框，显示"发布成功" |
| 已登录，发布信号帖（填了交易对） | 同上，post_type 自动设为 "signal" |
| Token 过期 | 按钮变 spinner 后恢复，显示错误提示（如"401 Unauthorized"） |
| 服务端不可达（断网） | 同上，显示网络错误提示 |
| 快速双击发布 | 第二次点击被忽略（Loading 态拦截） |
| 内容为空时 | 按钮灰色不可点 |

---

## 五、发布成功后看不到帖子——缺失社交 Feed 展示页面

**当前状态：PostScreen 只清空输入框，且客户端没有社交 Feed 展示页面。**

```
客户端现有:
  PostScreen → CreatePost ✅  (发帖)
  FeedScreen → SignalsService ✅ (信号卡，非社交帖子)
  FeedRpc.kt → createPost() ✅ (RPC 封装)

客户端缺失:
  SocialFeedScreen → FeedService.GetFeed ❌  (查看帖子)
  PostViewModel 发布后通知刷新 ❌
```

帖子发到了服务端 `alfq_posts` 表，但客户端**没有任何 UI 调用 `GetFeed`**。首页 FeedScreen 显示的是交易信号，不是用户社交帖子。

### 需要的改动

**1. 新建 SocialFeedViewModel**

```kotlin
@HiltViewModel
class SocialFeedViewModel @Inject constructor(
    private val feedRpc: FeedRpcClient
) : ViewModel() {
    private val _posts = MutableStateFlow<List<Post>>(emptyList())
    val posts: StateFlow<List<Post>> = _posts.asStateFlow()

    fun loadFeed() {
        viewModelScope.launch {
            val resp = feedRpc.getFeed(pageSize = 20)
            _posts.value = resp
        }
    }
}
```

**2. 新建 SocialFeedScreen**

```kotlin
@Composable
fun SocialFeedScreen(viewModel: SocialFeedViewModel = hiltViewModel()) {
    val posts by viewModel.posts.collectAsState()
    LazyColumn {
        items(posts) { post ->
            PostCard(post)
        }
    }
}
```

**3. 路由注册** — 在 `AlfQApp.kt` NavHost 添加:
```kotlin
composable("social_feed") { SocialFeedScreen() }
```

**4. PostViewModel 发布后通知刷新**

```kotlin
private val _postSuccess = MutableSharedFlow<Unit>()
val postSuccess: SharedFlow<Unit> = _postSuccess.asSharedFlow()
// CreatePost 成功后: _postSuccess.emit(Unit)
```

---

## 六、涉及文件清单

| 文件 | 改动 |
|---|---|
| `ui/post/PostViewModel.kt` | 注入 `FeedRpcClient`，引入 `PostState` sealed class，`post()` 用 `feedRpc.createPost()` |
| `ui/post/PostScreen.kt` | 加 loading spinner、error 文字、防重复提交 |
| （无需改动）`data/rpc/FeedRpc.kt` | 已有 `FeedRpcClient`，直接复用 |
| （无需改动）`data/rpc/ConnectTransportProvider.kt` | 已有 Token 注入机制 |
