# AntClaw 数据采集调度方案

> **版本**：v1.0
> **状态**：设计草案（待评审）
> **作者**：AntClaw 工程团队
> **关联文档**：`AntClaw-重构解决方案.md` §3.1、§4.2、§7

---

## 一、设计目标

构建一套**事件驱动 + 智能触发 + 多级降级**的数据采集调度系统，实现：

1. **零数据丢失**：所有外部数据增量持久化，永不重复抓取，永不遗漏
2. **最小 API 调用**：尊重外部 API 速率限制，缓存 + 去重 + 智能触发
3. **最低延迟**：财经事件发布后 ≤ 60 秒内捕获实际值
4. **最强鲁棒性**：单点故障不影响整体，多级降级保证可用性
5. **可观测性**：完整的指标、日志、分布式追踪
6. **可扩展性**：新增数据源只需实现一个接口

---

## 二、与简单定时任务的差异

| 维度 | 简单 cron 定时任务 | AntClaw 事件驱动调度 |
|------|-------------------|---------------------|
| **触发机制** | 固定间隔轮询 | 条件触发 + 事件窗口 + 用户偏好驱动 |
| **数据写入** | 全量覆盖 | 增量 UPSERT，diff 检测，only-on-change |
| **API 调用** | 无差别请求 | 滑动窗口限流 + Redis 计数器 + circuit breaker |
| **缓存策略** | 无缓存或简单 TTL | L1（内存）+ L2（Redis）+ L3（DB）三级缓存 |
| **失败处理** | 失败即丢弃或简单重试 | 指数退避 + 熔断 + 降级到上次成功数据 |
| **并发控制** | 无 | 信号量限制 + 任务幂等键 + 分布式锁 |
| **可观测性** | 仅日志 | OTel 全链路追踪 + Prometheus 指标 + Sentry 异常 |
| **数据质量** | 无校验 | Schema 校验 + 异常值检测 + 历史对照 |

---

## 三、整体架构

### 3.1 分层架构

```
┌─────────────────────────────────────────────────────────────┐
│              Trigger Layer（触发层）                          │
│  ConditionTicker | EventWindow | UserPrefs | Webhook         │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│              Coordinator Layer（协调层）                      │
│  Job Dispatcher | Idempotency Guard | Distributed Lock       │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│              Fetcher Layer（采集层）                          │
│  FRED | MQL5 | CFTC | ECB | OECD | ...（统一 Fetcher 接口）   │
│  + Rate Limiter + Circuit Breaker + Retry + Timeout          │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│              Validator Layer（校验层）                        │
│  Schema Check | Range Check | Outlier Detection | Diff       │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│              Persistence Layer（持久化层）                    │
│  PostgreSQL（TimescaleDB hypertable）+ Redis（热缓存）         │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│              Notification Layer（通知层）                     │
│  Surprise Scoring | Confluence Detection | SSE Fanout        │
└─────────────────────────────────────────────────────────────┘
```

### 3.2 进程结构

| 进程 | 职责 | 副本数 |
|------|------|--------|
| `antclaw-api` | 提供 Connect-RPC，读取持久化数据 | 1+（无状态可水平扩展）|
| `antclaw-worker` | **运行采集调度器**，写入 DB 和 Redis | **1**（首发单实例，分布式锁防双写）|
| `postgres` | TimescaleDB 持久化 | 1（主），1（备，硬化阶段）|
| `redis` | 缓存、限流、Pub/Sub、SSE 扇出 | 1 |

**关键决策**：采集调度器**只在 worker 进程**运行，api 进程纯读，避免双写竞争。

---

## 四、五大采集循环（继承 + 优化）

### 4.1 初始同步（Initial Sync）

**触发**：worker 启动时一次

**流程**：
```
1. 查询 DB 中今日数据是否存在
   ├── 存在 → 跳过初始同步，仅运行 missed-data 检查
   └── 不存在 → 全量抓取本周数据，批量写入 DB
2. 检查每个数据源的最近一次成功时间
   └── 超过 SLA 阈值的源 → 标记为 stale，进入快速恢复模式
```

**优化点（vs ark-intelligent）**：
- 增加 **stale 检测**：worker 重启时检测哪些数据源长时间未更新
- 使用 **批量 UPSERT**：单次事务写入多条，提升 5-10x 性能

### 4.2 周同步（Weekly Sync）

**触发**：每小时检查一次，仅在 `Sunday 23:00 UTC+8` 执行

**流程**：
```
1. 抓取下周所有财经事件
2. 与 DB 中已有数据 diff
3. UPSERT 新增/变更事件
4. 通过 Redis Pub/Sub 通知 api 进程刷新缓存
```

**优化点**：
- **预加载窗口扩大**：从下一周扩展到 **未来 14 天**，减少边界场景遗漏
- **Idempotency Key**：使用 `weekly-sync:YYYY-WW` 作为分布式锁键，防止并发执行

### 4.3 微采集（Micro-Scrape）— **核心**

**触发**：每 30 秒（**比 ark-intelligent 的 1 分钟更密集**）

**流程**：
```
对每个待发布事件：
  发布后窗口期 = [0, 30] 分钟
  触发抓取的时间点（秒级精度）：
    30秒 → 1分钟 → 2分钟 → 3分钟 → 5分钟 → 10分钟 → 15分钟 → 20分钟 → 30分钟
  
  对每次触发：
    1. 查询事件最新值
    2. 与 DB 旧值对比
    3. 如有变化 → UPSERT + 触发 onNewRelease
    4. 记录 fetch latency 到 Prometheus
```

**优化点**：
- **轮询频率提升**：30 秒 vs 60 秒，**首次捕获延迟从 1min 降至 30s**
- **指数退避后期窗口**：发布后 5 分钟内密集采集，5-30 分钟稀疏采集
- **批量抓取**：一次请求获取当日所有未结算事件，**减少 90% API 调用**
- **diff 写入**：仅当 actual / forecast / previous 有变化时才写 DB

### 4.4 事前提醒（Pre-Event Reminder）

**触发**：每分钟

**流程**：
```
对每个未发布事件：
  minsUntil = 事件时间 - 当前时间（分钟）
  
  对每个活跃用户：
    if minsUntil ∈ user.alert_minutes (例如 [60, 30, 15, 5])
       and event 匹配用户筛选条件
       and 未发送过此提醒 (Redis SETNX 防重)
       and 未触发 quiet hours / daily cap
    → 通过 SSE / 推送 / 站内信 发送提醒
```

**优化点**：
- **去重**改为 Redis SETNX（替代内存 map），支持多 worker 实例
- **批量查询用户偏好**：一次性加载所有活跃用户偏好到内存，避免 N+1
- **优先级队列**：高影响事件（NFP、CPI、利率决议）优先发送

### 4.5 慢速数据源（Slow-Poll）— FRED / OECD / BIS

**触发**：每 5 分钟检查一次（FRED 系列）

**流程**：
```
对每个 FRED 系列：
  1. 检查 Redis 缓存是否过期（TTL = 5 分钟）
  2. 已过期 → 并发抓取（最多 10 个并行）
  3. 写入 DB（增量）+ 更新 Redis 缓存
  4. 失败 → 返回缓存中的旧数据（graceful degradation）
```

**优化点（vs ark-intelligent）**：
- **持久化到 TimescaleDB**：原版只用内存缓存，重启即丢；新版写入 DB
- **continuous aggregates**：DB 自动计算每日/每周聚合，查询时无需重算
- **背压控制**：worker 维持的 fetch 队列长度超过阈值时，自动延长 TTL

---

## 五、PostgreSQL + TimescaleDB Schema

### 5.1 时间序列表（hypertable）

```sql
-- 通用时间序列快照表
CREATE TABLE data_snapshots (
    time          TIMESTAMPTZ NOT NULL,
    source        VARCHAR(32) NOT NULL,    -- 'fred', 'mql5', 'cot', 'ecb', etc.
    series_id     VARCHAR(64) NOT NULL,    -- 'GDP', 'UNRATE', 'EUR_NFP', etc.
    value_numeric DOUBLE PRECISION,
    value_text    TEXT,                    -- 非数值数据
    raw_json      JSONB,                   -- 原始 API 响应
    fetched_at    TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (time, source, series_id)
);

-- 转换为 hypertable，按 30 天分区
SELECT create_hypertable('data_snapshots', 'time',
    chunk_time_interval => INTERVAL '30 days');

-- 启用压缩（30 天前的数据自动压缩，节省 90% 空间）
ALTER TABLE data_snapshots SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'source, series_id'
);
SELECT add_compression_policy('data_snapshots', INTERVAL '30 days');

-- 索引
CREATE INDEX idx_snapshots_source_series ON data_snapshots (source, series_id, time DESC);
```

### 5.2 财经事件表

```sql
CREATE TABLE calendar_events (
    event_id          VARCHAR(64) PRIMARY KEY,        -- 'mql5-12345'
    title             VARCHAR(256) NOT NULL,
    country           VARCHAR(8),
    currency          VARCHAR(8),
    impact            VARCHAR(16),                    -- 'low', 'medium', 'high'
    scheduled_at      TIMESTAMPTZ NOT NULL,
    previous_value    TEXT,
    forecast_value    TEXT,
    actual_value      TEXT,
    impact_direction  SMALLINT,                       -- 0/1/2
    surprise_score    DOUBLE PRECISION,               -- stddev-normalized
    surprise_label    VARCHAR(32),                    -- 'MAJOR_BEAT' etc.
    revision_label    VARCHAR(32),
    fetched_at        TIMESTAMPTZ DEFAULT NOW(),
    updated_at        TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_events_scheduled ON calendar_events (scheduled_at);
CREATE INDEX idx_events_currency_impact ON calendar_events (currency, impact, scheduled_at DESC);
```

### 5.3 采集任务审计表

```sql
CREATE TABLE fetch_jobs (
    id               BIGSERIAL PRIMARY KEY,
    job_type         VARCHAR(64) NOT NULL,            -- 'fred-fetch', 'mql5-micro'
    source           VARCHAR(32) NOT NULL,
    started_at       TIMESTAMPTZ DEFAULT NOW(),
    completed_at     TIMESTAMPTZ,
    status           VARCHAR(16),                     -- 'running', 'success', 'failed'
    error_message    TEXT,
    records_inserted INTEGER DEFAULT 0,
    records_updated  INTEGER DEFAULT 0,
    duration_ms      INTEGER
);

SELECT create_hypertable('fetch_jobs', 'started_at',
    chunk_time_interval => INTERVAL '7 days');
```

### 5.4 Continuous Aggregates（连续聚合）

```sql
-- 自动计算每日 FRED 指标聚合，查询时无需扫全表
CREATE MATERIALIZED VIEW fred_daily_agg
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 day', time) AS day,
    source,
    series_id,
    last(value_numeric, time) AS daily_value,
    count(*) AS observations
FROM data_snapshots
WHERE source = 'fred'
GROUP BY day, source, series_id;

-- 每 1 小时刷新
SELECT add_continuous_aggregate_policy('fred_daily_agg',
    start_offset => INTERVAL '7 days',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour');
```

---

## 六、Redis 数据结构

| Key 模式 | 类型 | TTL | 用途 |
|---------|------|-----|------|
| `cache:fred:{series_id}` | String (JSON) | 5min | FRED 数据热缓存 |
| `cache:mql5:events:{date}` | String (JSON) | 1min | 当日事件列表缓存 |
| `lock:job:{job_type}:{key}` | String | 5min | 分布式幂等锁（SETNX）|
| `ratelimit:fred:{minute}` | Counter | 60s | FRED API 滑动窗口限流 |
| `dedup:reminder:{event_id}:{user_id}:{mins}` | String | 86400s | 提醒去重 |
| `surprise:weekly:{year_week}:{currency}` | Sorted Set | 7d | 周度 surprise 累积 |
| `pubsub:data_updated` | Pub/Sub | - | 数据更新通知 → SSE 扇出 |
| `circuit:fred` | Hash | - | 熔断器状态 |

---

## 七、Go 接口设计

### 7.1 通用 Fetcher 接口

```go
package fetcher

// Fetcher is the unified interface for all data source clients.
type Fetcher interface {
    // Name returns a unique source identifier (e.g., "fred", "mql5").
    Name() string
    
    // Fetch retrieves data for the given key (series_id, event_id, etc.).
    // Returns FetchResult with normalized data points.
    Fetch(ctx context.Context, key string) (*FetchResult, error)
    
    // Healthcheck verifies the source is reachable.
    Healthcheck(ctx context.Context) error
}

type FetchResult struct {
    Source    string
    Key       string
    DataPoints []DataPoint
    RawJSON   []byte
    FetchedAt time.Time
}

type DataPoint struct {
    Time         time.Time
    ValueNumeric *float64
    ValueText    *string
    Metadata     map[string]string
}
```

### 7.2 调度器接口

```go
package scheduler

// Scheduler manages background data collection jobs.
type Scheduler interface {
    // Start begins all collection loops; blocks until ctx cancelled.
    Start(ctx context.Context) error
    
    // RegisterJob adds a new collection job to the scheduler.
    RegisterJob(job Job) error
    
    // TriggerNow forces immediate execution of a named job (for admin/testing).
    TriggerNow(ctx context.Context, jobName string) error
    
    // Stats returns current scheduler statistics.
    Stats() SchedulerStats
}

// Job describes a single collection job.
type Job interface {
    Name() string                                   // Unique job name
    Schedule() Schedule                             // When to run
    Execute(ctx context.Context) (*JobResult, error) // Job logic
}

// Schedule defines when a job should execute.
type Schedule interface {
    NextRun(now time.Time) time.Time
    ShouldRun(now time.Time, lastRun time.Time) bool
}
```

### 7.3 仓储接口（Hexagonal Ports）

```go
package ports

// SnapshotRepository persists time-series data points.
type SnapshotRepository interface {
    // BatchUpsert inserts/updates multiple snapshots in a single transaction.
    BatchUpsert(ctx context.Context, snapshots []Snapshot) (inserted, updated int, err error)
    
    // GetLatest returns the most recent snapshot for source+series.
    GetLatest(ctx context.Context, source, seriesID string) (*Snapshot, error)
    
    // GetHistory returns N recent snapshots ordered by time desc.
    GetHistory(ctx context.Context, source, seriesID string, limit int) ([]Snapshot, error)
}

// CalendarRepository persists economic calendar events.
type CalendarRepository interface {
    UpsertEvents(ctx context.Context, events []CalendarEvent) error
    UpdateActual(ctx context.Context, eventID string, actual string, surprise float64) error
    GetByDate(ctx context.Context, date time.Time) ([]CalendarEvent, error)
    GetUpcoming(ctx context.Context, within time.Duration) ([]CalendarEvent, error)
}
```

---

## 八、关键算法

### 8.1 增量 UPSERT（PostgreSQL）

```sql
INSERT INTO data_snapshots (time, source, series_id, value_numeric, raw_json)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (time, source, series_id) DO UPDATE SET
    value_numeric = EXCLUDED.value_numeric,
    raw_json = EXCLUDED.raw_json,
    fetched_at = NOW()
WHERE data_snapshots.value_numeric IS DISTINCT FROM EXCLUDED.value_numeric;
-- 关键：WHERE 子句确保只在值实际变化时更新，避免无效写
```

### 8.2 滑动窗口限流

```go
// Redis 实现：每分钟最多 N 次调用
func (r *RateLimiter) Allow(ctx context.Context, source string, limit int) (bool, error) {
    key := fmt.Sprintf("ratelimit:%s:%d", source, time.Now().Unix()/60)
    count, err := r.redis.Incr(ctx, key).Result()
    if err != nil {
        return false, err
    }
    if count == 1 {
        r.redis.Expire(ctx, key, 90*time.Second)
    }
    return count <= int64(limit), nil
}
```

### 8.3 分布式幂等锁

```go
// 防止多 worker 重复执行同一任务
func (l *Lock) Acquire(ctx context.Context, jobKey string, ttl time.Duration) (bool, error) {
    key := "lock:job:" + jobKey
    ok, err := l.redis.SetNX(ctx, key, l.workerID, ttl).Result()
    return ok, err
}
```

### 8.4 熔断器

```go
// 连续失败 N 次 → 熔断 5 分钟
type CircuitBreaker struct {
    failures    int
    lastFailure time.Time
    state       string // "closed", "open", "half-open"
}

func (cb *CircuitBreaker) Allow() bool {
    if cb.state == "open" {
        if time.Since(cb.lastFailure) > 5*time.Minute {
            cb.state = "half-open" // 允许试探性请求
            return true
        }
        return false
    }
    return true
}
```

### 8.5 Surprise Scoring（标准差归一化）

```go
// stddev 归一化的 surprise 评分
func ComputeSurprise(actual, forecast float64, history []float64) float64 {
    diff := actual - forecast
    if len(history) < 3 {
        return diff // 历史数据不足，使用原始差值
    }
    
    // 计算历史 (actual - forecast) 的标准差
    sigma := stddev(history)
    if sigma == 0 {
        return diff
    }
    return diff / sigma // 返回 sigma 单位的归一化值
}
```

---

## 九、可观测性

### 9.1 Prometheus 指标

| 指标名称 | 类型 | 标签 | 说明 |
|---------|------|------|------|
| `antclaw_fetch_total` | Counter | `source, status` | 采集请求总数 |
| `antclaw_fetch_duration_seconds` | Histogram | `source` | 采集延迟分布 |
| `antclaw_fetch_records_inserted` | Counter | `source` | 新增记录数 |
| `antclaw_fetch_records_updated` | Counter | `source` | 更新记录数 |
| `antclaw_circuit_state` | Gauge | `source` | 熔断器状态（0=closed, 1=open）|
| `antclaw_event_capture_latency_seconds` | Histogram | `currency` | 事件发布到捕获的延迟 |
| `antclaw_ratelimit_rejected` | Counter | `source` | 限流拒绝次数 |

### 9.2 OTel Tracing

每个采集任务作为一个 span，包含子 span：
```
fetch.fred-gdp (root span)
├── ratelimit.check
├── http.request (FRED API)
├── validation
├── db.upsert
└── cache.invalidate
```

### 9.3 SLA 监控

| SLA | 阈值 | 告警渠道 |
|-----|------|---------|
| FRED 数据新鲜度 | < 10 min | Sentry + Grafana 告警 |
| MQL5 事件捕获延迟 | < 60 sec | Sentry |
| 采集成功率 | > 99% / 24h | Grafana 告警 |
| Worker 心跳 | < 30 sec 间隔 | 直接重启 |

---

## 十、与 ark-intelligent 的核心提升

| 能力 | ark-intelligent | AntClaw 新方案 | 提升 |
|------|----------------|----------------|------|
| 持久化 | BadgerDB（嵌入式 KV） | PostgreSQL + TimescaleDB | **支持复杂查询、聚合分析** |
| 缓存 | 单进程内存 | Redis 分布式缓存 | **多 worker 共享** |
| 多 worker | 不支持（单进程） | 分布式锁 + 幂等键 | **可水平扩展** |
| 微采集频率 | 1 min | 30 sec | **延迟降低 50%** |
| 数据写入 | 全量覆盖 | 增量 UPSERT + diff 检测 | **写入量减少 80%** |
| 限流 | 无 | Redis 滑动窗口 | **避免 API 封禁** |
| 熔断 | 简单标志位 | 三态熔断器 | **故障恢复更平滑** |
| 可观测 | zerolog | OTel + Prometheus + Sentry | **全链路追踪** |
| 聚合分析 | 实时计算 | continuous aggregates | **查询性能 10-100x** |
| 数据归档 | 无 | 自动压缩 + 永久保留 | **存储成本降 90%** |

---

## 十一、实施路线

### 阶段 1：基础设施（1-2 天）
- [ ] PostgreSQL + TimescaleDB Docker 部署
- [ ] Redis Docker 部署
- [ ] Schema 创建 + migration 工具
- [ ] 基础 repository 层（pgx）

### 阶段 2：核心调度器（2-3 天）
- [ ] Fetcher 接口 + FRED/MQL5 实现
- [ ] Scheduler 主循环
- [ ] 分布式锁 + 限流 + 熔断
- [ ] 五大采集循环骨架

### 阶段 3：智能分析（2 天）
- [ ] Surprise Scoring
- [ ] Confluence Detection（COT + FRED + Surprise）
- [ ] Revision Tracking
- [ ] Continuous Aggregates 配置

### 阶段 4：可观测性（1 天）
- [ ] Prometheus metrics 接入
- [ ] OTel tracing 接入
- [ ] Sentry 集成
- [ ] Grafana dashboard

### 阶段 5：测试与上线（2 天）
- [ ] 单元测试（覆盖率 ≥ 80%）
- [ ] 集成测试（mock 外部 API）
- [ ] 压测（1000 RPS 持续 1 小时）
- [ ] 上线监控

**总工期**：约 8-10 个工作日

---

## 十二、未决项

1. **FRED API Key 池化**：是否支持多个 API key 轮询，提升速率限制？
2. **MQL5 反爬升级**：如果 MQL5 加强反爬，是否需要 Playwright 渲染？
3. **数据回填**：历史数据回填策略（首次启动时是否补全过去 5 年数据？）
4. **多区域部署**：未来是否需要多 region worker（亚洲/欧美各一）？
5. **告警精度**：surprise threshold 的初始值如何确定（需要历史数据校准）？

---

## 十三、参考资料

- TimescaleDB Best Practices: https://docs.timescale.com/timescaledb/latest/how-to-guides/
- PostgreSQL UPSERT: https://www.postgresql.org/docs/current/sql-insert.html#SQL-ON-CONFLICT
- Redis Sliding Window Rate Limit: https://redis.io/docs/manual/patterns/distributed-locks/
- Circuit Breaker Pattern: https://martinfowler.com/bliki/CircuitBreaker.html
- OpenTelemetry Go SDK: https://opentelemetry.io/docs/languages/go/
