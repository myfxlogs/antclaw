# AlfQ · AI Agent 技术实现指南

> 文档版本：v1.0
> 创建日期：2026-05-14
> 适用范围：AI Agent 代码实现指导

---

## 一、技术架构设计

### 1.1 整体架构

```
┌─────────────────────────────────────────────────────────────────┐
│                      AlfQ 系统架构                              │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐    HTTP/2    ┌─────────────────────────────┐ │
│  │   Android   │◄────────────►│         Backend            │ │
│  │   Client    │   Connect-RPC │  (Go + Connect-RPC + SSE)  │ │
│  └─────────────┘              └─────────────────────────────┘ │
│         │                              │                       │
│         │ SSE                          │                       │
│         ▼                              ▼                       │
│  ┌─────────────┐              ┌─────────────────────────────┐ │
│  │  EventSource│              │ PostgreSQL | Redis | MinIO  │ │
│  │  (实时推送)  │              │        (数据层)             │ │
│  └─────────────┘              └─────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

### 1.2 客户端架构（Android）

| 层级 | 职责 | 技术选型 |
|-----|------|---------|
| **UI 层** | 界面渲染、用户交互 | Jetpack Compose |
| **ViewModel 层** | 状态管理、业务逻辑 | ViewModel + Flow |
| **Repository 层** | 数据访问、缓存 | Room + Connect-RPC |
| **API 层** | RPC 客户端、数据转换 | Connect-Web |
| **数据层** | 本地持久化 | Room Database |

### 1.3 服务端架构（Go）

| 层级 | 职责 | 包路径 |
|-----|------|-------|
| **Handler 层** | RPC 请求处理 | `adapter/rpc/` |
| **Service 层** | 业务逻辑 | `service/` |
| **Repository 层** | 数据访问 | `infra/postgres/` |
| **Domain 层** | 领域模型 | `domain/` |

---

## 二、核心功能模块

### 2.1 模块划分

| 模块 | 功能描述 | 状态 |
|-----|---------|------|
| **Auth** | 用户认证、登录注册 | ✅ 已实现 |
| **Feed** | 社交动态流、帖子互动 | ✅ 已实现 |
| **Signals** | 交易信号展示、详情 | ✅ 已实现 |
| **Discover** | 发现页、交易员推荐 | ✅ 已实现 |
| **Profile** | 用户主页、个人资料 | ✅ 已实现 |
| **MT Account** | MT4/MT5 账户绑定 | ✅ 已实现 |
| **Alerts** | 价格告警设置 | ✅ 已实现 |
| **Notifications** | 通知中心、推送管理 | ✅ 已实现 |
| **Chat** | 私聊功能 | ⚠️ 部分实现 |
| **Circle** | 圈子功能 | ⚠️ 部分实现 |

### 2.2 模块依赖关系

```
Auth (认证)
    ↓
Feed (动态) ←→ Signals (信号)
    ↓
Discover (发现) ←→ Profile (个人)
    ↓
MT Account ←→ Alerts
    ↓
Notifications ←→ Chat ←→ Circle
```

---

## 三、数据流程说明

### 3.1 认证流程

```
客户端                          服务端
  ↓                               ↓
LoginScreen ─→ LoginRequest ─→ AuthService.Login
                                   ↓
                              验证密码
                              查询用户
                              生成 Token
                                   ↓
LoginResponse ←─ access_token ──
   ↓
存储 Token 到 SharedPreferences
   ↓
设置 RPC 客户端 Authorization 头
```

### 3.2 Feed 加载流程

```
客户端                          服务端
  ↓                               ↓
FeedViewModel ─→ GetFeedRequest ─→ FeedService.GetFeed
                                        ↓
                                   查询 Feed 数据
                                   分页处理
                                   获取作者信息
                                        ↓
FeedResponse ←─  posts[] ───
   ↓
更新 UI 状态
   ↓
Compose 重组渲染
```

### 3.3 SSE 实时推送流程

```
客户端                          服务端
  ↓                               ↓
EventSource 连接 ──→ /stream/events
                              ↓
                         订阅用户事件
                         事件触发时推送
                              ↓
事件数据 ←──── SSE Message ───
   ↓
解析事件类型
   ↓
更新对应 UI
```

---

## 四、接口调用规则

### 4.1 RPC 调用规范

#### 4.1.1 认证头设置

```kotlin
// 所有 RPC 调用必须携带 Authorization 头
val headers = mapOf(
    "Authorization" to "Bearer ${tokenStore.getAccessToken()}"
)
```

#### 4.1.2 错误处理模板

```kotlin
suspend fun <T> safeRpcCall(block: suspend () -> T): Result<T> {
    return try {
        Result.success(block())
    } catch (e: ConnectException) {
        // 网络错误
        Result.failure(NetworkError)
    } catch (e: StatusException) {
        // RPC 状态错误
        when (e.code) {
            Code.UNAUTHENTICATED -> Result.failure(AuthError)
            Code.PERMISSION_DENIED -> Result.failure(ForbiddenError)
            else -> Result.failure(UnknownError)
        }
    }
}
```

### 4.2 SSE 订阅规范

```kotlin
// 订阅格式
// URL: /stream/events?user_id={userId}&topics={topics}
// topics: notifications, signals, feed, alerts

val eventSource = EventSource(
    url = "$BASE_URL/stream/events?topics=notifications,signals",
    headers = mapOf("Authorization" to "Bearer $token")
)

eventSource.addEventListener("notification") { event ->
    // 处理通知事件
}
```

### 4.3 数据格式约定

| 字段类型 | 格式要求 | 示例 |
|---------|---------|------|
| 日期时间 | Unix 毫秒时间戳 | `1704067200000` |
| 金额 | 字符串（避免浮点误差） | `"1234.56"` |
| 货币对 | 大写 6 字符 | `"EURUSD"` |
| UUID | 小写标准格式 | `"550e8400-e29b-41d4-a716-446655440000"` |

---

## 五、错误处理机制

### 5.1 客户端错误分类

| 错误类型 | 触发条件 | 处理策略 |
|---------|---------|---------|
| **AuthError** | Token 无效/过期 | 清除本地 Token，跳转到登录页 |
| **NetworkError** | 网络不可用 | 显示错误提示，提供重试按钮 |
| **ApiError** | 服务端业务错误 | 显示具体错误信息 |
| **ValidationError** | 请求参数校验失败 | 显示表单错误提示 |

### 5.2 服务端错误码

| 错误码 | 含义 | HTTP 状态码 |
|-------|------|------------|
| `UNAUTHENTICATED` | 未认证 | 401 |
| `PERMISSION_DENIED` | 无权限 | 403 |
| `INVALID_ARGUMENT` | 参数无效 | 400 |
| `NOT_FOUND` | 资源不存在 | 404 |
| `INTERNAL` | 服务器内部错误 | 500 |
| `UNAVAILABLE` | 服务暂时不可用 | 503 |

### 5.3 重试策略

```kotlin
// 指数退避重试
val retryDelays = listOf(1000L, 2000L, 4000L, 8000L)

suspend fun <T> withRetry(maxRetries: Int = 3, block: suspend () -> T): T {
    var lastException: Exception? = null
    
    for (i in 0 until maxRetries) {
        try {
            return block()
        } catch (e: Exception) {
            lastException = e
            if (i < maxRetries - 1) {
                delay(retryDelays[minOf(i, retryDelays.size - 1)])
            }
        }
    }
    
    throw lastException ?: RuntimeException("Unknown error")
}
```

---

## 六、性能优化策略

### 6.1 客户端优化

| 优化项 | 策略 | 实现位置 |
|-------|------|---------|
| **列表滚动** | 使用 `LazyColumn` + 分页加载 | FeedScreen.kt |
| **图片加载** | Coil 缓存 + 占位图 | SharedComposables.kt |
| **状态管理** | 合理使用 `remember` + `mutableState` | ViewModel |
| **网络请求** | 防抖 + 缓存 | Repository |
| **内存管理** | 使用 `DisposableEffect` 清理资源 | Composables |

### 6.2 服务端优化

| 优化项 | 策略 | 实现位置 |
|-------|------|---------|
| **数据库查询** | 索引优化 + 分页 | Repository |
| **缓存策略** | Redis 缓存热点数据 | infra/redis/ |
| **限流** | 基于用户的请求频率限制 | auth/rate_limit.go |
| **异步处理** | 非关键路径异步执行 | Service |

### 6.3 缓存策略

| 数据类型 | 缓存位置 | TTL | 失效策略 |
|---------|---------|-----|---------|
| 用户信息 | SharedPreferences | 1 小时 | 登录/登出时刷新 |
| Feed 数据 | Room | 5 分钟 | 下拉刷新时清除 |
| 价格数据 | Memory + Room | 30 秒 | SSE 更新时清除 |
| 信号详情 | Room | 10 分钟 | 信号更新时清除 |

---

## 七、技术选型依据

### 7.1 客户端技术栈

| 技术 | 版本 | 选型理由 |
|-----|------|---------|
| **Kotlin** | 1.9 | 现代 JVM 语言，空安全，协程支持 |
| **Jetpack Compose** | 1.6 | 声明式 UI，性能优秀，官方支持 |
| **ViewModel** | 2.6 | 生命周期感知，状态管理 |
| **Flow** | 1.0 | 响应式编程，异步数据流 |
| **Room** | 2.6 | 本地数据库，ORM 支持 |
| **Connect-RPC** | 1.1 | 类型安全，HTTP/2，代码生成 |
| **Hilt** | 2.4 | 依赖注入，减少样板代码 |

### 7.2 服务端技术栈

| 技术 | 版本 | 选型理由 |
|-----|------|---------|
| **Go** | 1.21 | 高性能，并发模型优秀 |
| **Connect-RPC** | 1.1 | 类型安全，与客户端无缝对接 |
| **PostgreSQL** | 16 | 关系型数据库，JSONB 支持 |
| **Redis** | 7.2 | 缓存、消息队列 |
| **Wire** | 0.5 | 编译时依赖注入 |

---

## 八、实现标准

### 8.1 代码风格规范

#### 8.1.1 Kotlin 规范

```kotlin
// 命名规则
// ✅ 推荐
val userName: String
fun fetchUserData(): Flow<User>
class UserRepository

// ❌ 不推荐
val UserName: String  // PascalCase 仅用于类名
fun GetUser()         // 函数名使用 camelCase
```

#### 8.1.2 Go 规范

```go
// 命名规则
// ✅ 推荐
type User struct {}
func (u *User) GetName() string

// 错误处理
err := doSomething()
if err != nil {
    return fmt.Errorf("context: %w", err)
}
```

### 8.2 测试标准

| 测试类型 | 覆盖要求 | 实现位置 |
|---------|---------|---------|
| **单元测试** | ViewModel / Service 层 | `*_test.kt` / `*_test.go` |
| **集成测试** | Repository / Handler 层 | `*_test.go` |
| **E2E 测试** | 核心业务流程 | `scripts/e2e/` |

### 8.3 文档标准

| 文档类型 | 要求 | 位置 |
|---------|------|-----|
| **API 文档** | 每个 RPC 方法必须有注释 | `.proto` 文件 |
| **设计文档** | 每个里程碑一份 | `docs/alfq/` |
| **实现指南** | AI Agent 执行参考 | 本文档 |

---

## 九、代码实现模板

### 9.1 ViewModel 模板

```kotlin
class ExampleViewModel(
    private val repository: ExampleRepository,
    private val dispatcher: CoroutineDispatcher = Dispatchers.IO
) : ViewModel() {
    
    private val _state = MutableStateFlow(AsyncState.Idle)
    val state: StateFlow<AsyncState<ExampleData>> = _state
    
    fun loadData() {
        viewModelScope.launch(dispatcher) {
            _state.value = AsyncState.Loading
            val result = repository.fetchData()
            _state.value = when (result) {
                is Result.Success -> AsyncState.Success(result.data)
                is Result.Failure -> AsyncState.Error(result.error)
            }
        }
    }
}

sealed class AsyncState<out T> {
    object Idle : AsyncState<Nothing>()
    object Loading : AsyncState<Nothing>()
    data class Success<out T>(val data: T) : AsyncState<T>()
    data class Error(val error: Throwable) : AsyncState<Nothing>()
}
```

### 9.2 Repository 模板

```kotlin
class ExampleRepository(
    private val api: ExampleApi,
    private val db: AppDatabase,
    private val tokenStore: TokenStore
) {
    
    suspend fun fetchData(): Result<ExampleData> {
        return safeRpcCall {
            // 1. 优先从缓存读取
            val cached = db.exampleDao().getLatest()
            if (cached != null && !isExpired(cached.timestamp)) {
                return Result.success(cached.toDomain())
            }
            
            // 2. 从网络获取
            val response = api.getExample(
                GetExampleRequest.newBuilder().build()
            )
            
            // 3. 缓存到本地
            db.exampleDao().insert(response.toEntity())
            
            Result.success(response.toDomain())
        }
    }
}
```

### 9.3 Handler 模板

```go
type ExampleHandler struct {
    service ExampleService
}

func (h *ExampleHandler) GetExample(
    ctx context.Context,
    req *pb.GetExampleRequest,
) (*pb.GetExampleResponse, error) {
    // 1. 参数校验
    if err := validateRequest(req); err != nil {
        return nil, status.Error(codes.InvalidArgument, err.Error())
    }
    
    // 2. 调用 Service
    data, err := h.service.GetExample(ctx, req)
    if err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }
    
    // 3. 转换响应
    return &pb.GetExampleResponse{
        Data: convertToProto(data),
    }, nil
}
```

---

## 十、部署与集成

### 10.1 环境配置

| 环境 | 配置文件 | API 地址 |
|-----|---------|---------|
| **开发** | `local.properties` | `http://localhost:8082` |
| **测试** | `staging.properties` | `https://api.staging.antclaw.io` |
| **生产** | `production.properties` | `https://api.antclaw.io` |

### 10.2 构建命令

```bash
# Android 构建
cd frontend/android
./gradlew assembleDebug    # Debug 构建
./gradlew assembleRelease  # Release 构建
./gradlew installDebug     # 安装到设备

# 后端构建
cd backend
go build ./cmd/antclaw-api
go run ./cmd/antclaw-api

# Docker 构建
docker compose -f deploy/docker-compose.yaml build
```

---

## 附录：常用工具函数

### A.1 时间格式化

```kotlin
fun formatTimestamp(timestamp: Long): String {
    val date = Date(timestamp)
    val formatter = SimpleDateFormat("yyyy-MM-dd HH:mm", Locale.getDefault())
    return formatter.format(date)
}
```

### A.2 价格格式化

```kotlin
fun formatPrice(price: String, pair: String): String {
    val decimals = if (pair.endsWith("JPY")) 2 else 4
    return "%.${decimals}f".format(price.toDouble())
}
```

### A.3 错误日志

```kotlin
fun logError(tag: String, message: String, e: Throwable? = null) {
    if (BuildConfig.DEBUG) {
        Log.e(tag, message, e)
    }
    // 生产环境上报到监控平台
}
```

---

**文档结束**
