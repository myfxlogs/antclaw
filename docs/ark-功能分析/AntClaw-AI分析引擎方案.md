# AntClaw AI 分析引擎方案

> **版本**：v1.0  
> **对应 ark-intelligent 模块**：`internal/service/ai/` (15 files, 200KB+)  
> **对应 AntClaw Proto**：`AIService`

---

## 一、ark-intelligent 方法分析

### 1.1 核心职责
AI 模块是 ark-intelligent **最大单包**，提供两大类能力：
1. **结构化分析**（COT 解读、宏观解读、信号解读）— 后台调用，固定 prompt
2. **对话式聊天**（带工具：web search、memory、code execution）— 用户实时交互

支持 **多 model 并行**：Gemini（Google）+ Claude（Anthropic 通过自建代理）。

### 1.2 模块矩阵

| 文件 | 职责 |
|------|------|
| `gemini.go` | Google Gemini 客户端（generative-ai-go SDK）|
| `claude.go` | Claude 客户端（自建代理 marketriskmonitor.com/api/analyze）|
| `claude_analyzer.go` | 用 Claude 做结构化分析的适配层 |
| `interpreter.go` | Gemini Interpreter（结构化分析主入口）|
| `cached_interpreter.go` | 带缓存的 Interpreter（避免重复 LLM 调用）|
| `chat_service.go` | 对话服务（多轮聊天 + 工具调用）|
| `context_builder.go` | **上下文注入器**：检测用户问题相关性，从 FRED/sentiment 等服务取数据，注入 system prompt |
| `unified_outlook.go` | **大集成报告**：调用所有子服务（COT/FRED/BIS/Wyckoff/ICT/GEX/微观/IMF/WB/Treasury），生成统一展望 |
| `prompts.go` | **Prompt 模板库** (33KB)：SystemPromptTemplate + 各场景特定 prompt |
| `memory_store.go` | 用户级文件式记忆（适配 Anthropic memory_20250818 工具）|
| `tools.go` | **分级工具配置**（Free/Pro/Premium 不同工具）|
| `tool_executor.go` | 工具调用执行器（路由到 memory_store）|
| `ai_ratelimit.go` | RPM + 每日配额限流（滑动窗口 + 每日午夜重置）|

### 1.3 关键设计

#### A. 双模型 + 用户偏好
- 用户在 `/settings` 选 `PreferredModel`（gemini / claude / claude-opus）
- `WithModel()` 返回 per-request scoped copy，**线程安全**
- Prompt 共享：相同 prompt 模板适用两个 model，输出一致

#### B. Context Builder（智能注入）
- 检测用户问题关键词（"EUR"、"COT"、"VIX"、"recession"…）
- 按需调用：FRED MacroData、Sentiment、Fed Speeches、价格上下文
- 拼接到 system prompt（避免每次问答都全量注入，节省 token）

#### C. Unified Outlook
- 一次性调用 **15+ 子服务**：COT/FRED/Wyckoff/ICT/GEX/微观/IMF/WB/Treasury/BIS/Fed/Sentiment/DeFi/Macro/Price
- 每个子服务输出 → 标准化 markdown 段落
- 拼接成 4000+ token 的 prompt → LLM 综合输出展望

#### D. 工具系统
- **Server-side 工具**（Anthropic 自动执行）：web_search, web_fetch, code_execution
- **Client-side 工具**（需要回调）：memory_20250818
- **Tier 隔离**：Free 仅 memory；Pro+ 加 web tools；Premium 加 code execution

#### E. Memory Store
- 文件式：`view <path>`, `create <path>`, `delete <path>`, `rename`
- 持久化到磁盘（路径 namespace per user）
- 与 Anthropic memory tool 协议对齐

#### F. AI Rate Limiter
- RPM 滑动窗口 + 日级配额
- WIB（UTC+7）每日午夜重置
- 超限 → `ErrAIRateLimited`，返回缓存版本或模板降级

#### G. Caching
- `cached_interpreter`：相同输入指纹（COT 数据 hash）→ 返回缓存
- TTL：分析报告 6 小时；展望 1 小时

---

## 二、AntClaw 设计方案

### 2.1 架构

```
AIService (Proto, 已存在)
  ├── service/ai/clients
  │     ├── gemini.go
  │     ├── claude.go
  │     └── interface.go         (统一 ChatEngine 接口)
  ├── service/ai/analyzer        (结构化分析)
  │     ├── interpreter.go
  │     └── cached.go
  ├── service/ai/chat
  │     ├── service.go
  │     ├── tools.go
  │     └── tool_executor.go
  ├── service/ai/context_builder.go
  ├── service/ai/unified_outlook.go
  ├── service/ai/prompts/       (拆分子文件，<800 行)
  │     ├── system.go
  │     ├── cot.go
  │     ├── macro.go
  │     ├── outlook.go
  │     └── chat.go
  ├── service/ai/memory          (持久化到 Postgres + Redis)
  └── service/ai/ratelimit       (Redis 分布式限流)
      ↓
  infra/{anthropic,gemini}_client.go
  infra/postgres/ai_memory_repo.go
  infra/postgres/ai_usage_repo.go
```

**vs ark-intelligent 重大改动**：
1. **BYOK（Bring Your Own Key）**：用户在 `/settings` 设置自己的 API key（已在 Auth 设计）
2. **Memory 持久化**：从文件改为 Postgres + Redis，支持多 worker
3. **限流分布式**：Redis 滑动窗口替代单进程
4. **Prompt 拆分**：33KB 单文件拆分为多个子文件
5. **Outlook 异步化**：unified_outlook 改为后台任务，前端 SSE 进度

### 2.2 核心接口

```go
type ChatEngine interface {
    Name() string
    Chat(ctx, req ChatRequest) (*ChatResponse, error)
    ChatStream(ctx, req ChatRequest) (<-chan ChatChunk, error)
    Models() []ModelInfo
}

type Interpreter interface {
    AnalyzeCOT(ctx, analysis *COTAnalysis) (string, error)
    AnalyzeMacro(ctx, macro *MacroData) (string, error)
    AnalyzeSignal(ctx, signal *UnifiedSignal) (string, error)
    GenerateOutlook(ctx, snapshot *FullMarketSnapshot) (string, error)
}

type ContextBuilder interface {
    BuildSystemPrompt(ctx, query string, userID string) (string, error)
}

type MemoryStore interface {
    View(ctx, userID, path string) (string, error)
    Create(ctx, userID, path, content string) error
    Delete(ctx, userID, path string) error
    Rename(ctx, userID, oldPath, newPath string) error
}

type AIRateLimiter interface {
    Check(ctx, userID string, model string) error
    Record(ctx, userID, model string, tokens int) error
}
```

### 2.3 Schema

```sql
-- AI 调用日志（审计 + 用量统计）
CREATE TABLE ai_usage (
    id          BIGSERIAL PRIMARY KEY,
    user_id     VARCHAR(64),
    model       VARCHAR(64),
    operation   VARCHAR(32),    -- 'chat','interpret','outlook'
    prompt_tokens INT,
    completion_tokens INT,
    cached      BOOLEAN,
    duration_ms INT,
    error       TEXT,
    created_at  TIMESTAMPTZ
);
SELECT create_hypertable('ai_usage', 'created_at', chunk_time_interval => INTERVAL '30 days');

-- 用户记忆（替代文件存储）
CREATE TABLE ai_memory (
    user_id     VARCHAR(64),
    path        VARCHAR(512),
    content     TEXT,
    updated_at  TIMESTAMPTZ,
    PRIMARY KEY (user_id, path)
);

-- AI 缓存（结构化分析结果）
CREATE TABLE ai_cache (
    fingerprint VARCHAR(64) PRIMARY KEY,    -- SHA-256 of input
    operation   VARCHAR(32),
    model       VARCHAR(64),
    result      TEXT,
    created_at  TIMESTAMPTZ,
    expires_at  TIMESTAMPTZ
);
CREATE INDEX idx_ai_cache_expires ON ai_cache (expires_at);
```

### 2.4 Redis

| Key | TTL |
|-----|-----|
| `ai:ratelimit:{user_id}:{model}:rpm` | 60s（滑动窗口）|
| `ai:ratelimit:{user_id}:{model}:daily` | 24h |
| `cache:ai:result:{fingerprint}` | 1-6h |
| `cache:ai:context:{user_id}:{query_hash}` | 5m |
| `pubsub:ai:outlook:{job_id}` | - |

### 2.5 调度
- **Outlook 后台任务**：用户请求 → 入队 → worker 调用 15+ 子服务 → LLM → 写入 cache → SSE 推送
- **Cache 清理**：每日 03:00 删除 `expires_at < now`
- **Usage 报表**：每日 04:00 聚合 `ai_usage`，写入 `ai_usage_daily` 物化视图

### 2.6 优化与提升

| 维度 | ark-intelligent | AntClaw |
|------|----------------|---------|
| API key | 系统级 env | 用户级 BYOK + 系统级 fallback |
| Memory | 磁盘文件 | Postgres + Redis 缓存 |
| 限流 | 进程内 mutex | Redis 分布式 |
| 缓存 | 内存 LRU | DB + Redis 双层 |
| Outlook | 同步阻塞 | 异步 + SSE 进度 |
| Prompt | 33KB 单文件 | 拆分多文件，<800 行 |
| 多模型 | 硬编码 if/else | Strategy 模式，可插件 |
| 审计 | 仅日志 | `ai_usage` 表，可计费 |

---

## 三、参考文件

- ark-intelligent：`internal/service/ai/*.go`（15 文件）
- AntClaw proto：`proto/antclaw/v1/ai.proto`
- AntClaw service：`backend/internal/service/ai/`
