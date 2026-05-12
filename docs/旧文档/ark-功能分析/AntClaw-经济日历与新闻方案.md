# AntClaw 经济日历与新闻方案

> **版本**：v1.0  
> **对应 ark-intelligent 模块**：`internal/service/news/` (9 files) + `internal/service/marketdata/` (多数据源)  
> **对应 AntClaw Proto**：`CalendarService`, `AlertsService`

---

## 一、ark-intelligent 方法分析

### 1.1 核心职责
- 从 MQL5 抓取经济日历（事件 + forecast + previous + actual）
- 监控 Fed RSS（FOMC 声明、Fed 主席讲话）
- 计算 Surprise Score（actual vs forecast 的 sigma 归一化）
- 记录事件对价格的 impact（15m/30m/1h/4h 窗口）
- 自举（bootstrap）历史 impact 数据库

### 1.2 数据采集（fetcher.go）

**MQL5 隐藏 POST 端点**：
- 无需 API key，伪造浏览器请求头（UA, Referer, Origin, X-Requested-With）
- POST 参数：date_mode, from, to, importance, currencies
- 返回所有影响级别事件（filter 放在客户端）
- 使用 circuit breaker，失败立即降级到缓存

**Scrape vs Fetch**：
- `ScrapeCalendar(ctx, "this")`：本周事件
- `ScrapeCalendar(ctx, "next")`：下周事件
- `ScrapeActuals(ctx, dateStr)`：只抓某日已发布事件的实际值

### 1.3 调度器（scheduler.go，已在《数据采集调度方案》详细分析）
五大循环：初始同步、周同步、微采集、事前提醒、Fed RSS 监控。

### 1.4 Surprise Scoring（surprise.go）

```
surprise_sigma = (actual - forecast) / stddev(history of diffs)
```

**分类**：
- `|sigma| < 0.5` → `In Line`
- `0.5 ≤ |sigma| < 1.5` → `Minor Beat / Minor Miss`
- `|sigma| ≥ 1.5` → `Major Beat / Major Miss`

**ImpactDirection 修正**：
- 某些指标"数字越大越坏"（失业率、CPI 对股票）
- `ImpactDirection=2` → 翻转 sigma 符号

### 1.5 Impact Recorder（impact_recorder.go）

- 事件发布后非阻塞启动 goroutine
- 在 15m / 30m / 1h / 4h 四个时点查询价格
- 计算 `(price_after - price_before) / price_before`
- 保存到 `EventImpactRecord`，用于后续机器学习

### 1.6 Impact Bootstrap（impact_bootstrap.go）

启动时回填历史 impact：
- 遍历 `NewsRepository` 中过去 90 天所有事件
- 对每个已有 actual 的事件，查询对应价格历史，计算 impact
- 用于训练 confluence 模型

### 1.7 Fed RSS（fed_rss.go）

- 抓取 Fed 官网 RSS feed（speeches + press-release）
- 解析新条目（去重基于 GUID）
- 提取内容，关键词分类 HAWKISH/DOVISH/NEUTRAL
- 推送给用户 + 缓存供 AI 上下文使用

### 1.8 MarketData 扩展（marketdata/）

**多源采集**：
- `bybit/`：订单流、Orderbook snapshot（crypto 微观结构）
- `coingecko/`：TOTAL3 加密市值
- `cryptocompare/`：加密价格备选源
- `defillama/`：DeFi TVL
- `deribit/`：期权隐含波动率（DVOL）
- `finviz/`：股票情绪
- `massive/`：聚合提供商

每个子包实现独立的 fetcher + 缓存 + 降级。

---

## 二、AntClaw 设计方案

### 2.1 架构调整

```
CalendarService (Proto)
  ├── service/calendar/mql5    (MQL5 采集)
  ├── service/calendar/fed_rss (Fed RSS)
  ├── service/calendar/surprise
  └── service/calendar/impact  (Impact Recorder + Bootstrap)
      ↓
  infra/apiclient/mql5_client.go
  infra/apiclient/fed_rss_client.go
  infra/postgres/calendar_repo.go
  infra/redis/calendar_cache.go
```

### 2.2 核心接口

```go
type CalendarFetcher interface {
    FetchEvents(ctx, from, to time.Time) ([]CalendarEvent, error)
    FetchActuals(ctx, date time.Time) ([]CalendarEvent, error)
}

type FedRSSFetcher interface {
    FetchSpeeches(ctx) ([]FedSpeech, error)
    FetchPressReleases(ctx) ([]FedRelease, error)
}

type CalendarRepository interface {
    UpsertEvents(ctx, events []CalendarEvent) (int, error)
    UpdateActual(ctx, eventID, actual string, surprise float64) error
    GetByDate(ctx, date time.Time) ([]CalendarEvent, error)
    GetUpcoming(ctx, within time.Duration) ([]CalendarEvent, error)
    GetHistoricalSurprises(ctx, eventName, currency string, limit int) ([]float64, error)
    SaveImpactRecord(ctx, rec ImpactRecord) error
}
```

### 2.3 Schema

```sql
-- 事件表（见采集调度方案，已定义）
-- 复用 calendar_events

-- 历史 surprise 表（for stddev normalization）
CREATE TABLE calendar_surprise_history (
    id            BIGSERIAL PRIMARY KEY,
    event_name    VARCHAR(256),
    currency      VARCHAR(8),
    released_at   TIMESTAMPTZ,
    actual_val    DOUBLE PRECISION,
    forecast_val  DOUBLE PRECISION,
    diff          DOUBLE PRECISION,  -- actual - forecast
    sigma         DOUBLE PRECISION   -- normalized
);
CREATE INDEX idx_surprise_event_currency ON calendar_surprise_history (event_name, currency, released_at DESC);

-- 价格 impact 记录
CREATE TABLE event_impact_records (
    event_id      VARCHAR(64),
    window        VARCHAR(8),       -- '15m','30m','1h','4h'
    symbol        VARCHAR(32),
    price_before  DOUBLE PRECISION,
    price_after   DOUBLE PRECISION,
    pct_change    DOUBLE PRECISION,
    recorded_at   TIMESTAMPTZ,
    PRIMARY KEY (event_id, window, symbol)
);
SELECT create_hypertable('event_impact_records', 'recorded_at', chunk_time_interval => INTERVAL '90 days');

-- Fed 讲话
CREATE TABLE fed_speeches (
    guid          VARCHAR(256) PRIMARY KEY,
    title         VARCHAR(512),
    speaker       VARCHAR(128),
    published_at  TIMESTAMPTZ,
    url           TEXT,
    content       TEXT,
    tone          VARCHAR(16),       -- HAWKISH/DOVISH/NEUTRAL
    fetched_at    TIMESTAMPTZ DEFAULT NOW()
);
```

### 2.4 Redis

| Key | 类型 | TTL | 内容 |
|-----|------|-----|------|
| `cache:calendar:upcoming` | JSON | 1m | 未来 24h 事件 |
| `cache:calendar:week:{YYYY-WW}` | JSON | 10m | 周事件列表 |
| `cache:fed:rss:latest` | JSON | 1h | Fed 最新讲话 |
| `dedup:reminder:{event}:{user}:{mins}` | SETNX | 24h | 提醒去重 |
| `surprise:weekly:{year_week}:{currency}` | ZSET | 7d | 累积 sigma |
| `pubsub:calendar:actual_released` | Pub/Sub | - | 实际值发布事件 |

### 2.5 调度（详见《数据采集调度方案》）

此处仅补充 Fed RSS 与 Impact：
- **Fed RSS 轮询**：每 10 分钟
- **Impact 回填**：每日 02:00 补齐昨日所有事件的 4 个窗口 impact
- **Impact Bootstrap**：worker 启动时触发一次，回填过去 90 天

### 2.6 多语言采集策略（新增）

#### 2.6.1 设计原则
经济日历事件中**仅 `title`/`description` 需要翻译**，其余字段（`event_id`、`country`、`currency`、`impact` 枚举、`scheduled_at`、`previous`/`forecast`/`actual` 数字）**完全语言无关**，跨语言共享。

> **数据真实性**：金融术语（如"Non-Farm Payrolls"="非农就业人数"）的标准译法对告警匹配、用户搜索、AI 上下文均关键，纯 LLM 翻译易出现"创造性翻译"导致术语漂移。

#### 2.6.2 三种方案权衡

| 方案 | 数据真实性 | 成本 | 一致性 | 备注 |
|------|----------|------|--------|------|
| 仅英文 + LLM 翻译 | ⚠️ 中 | LLM token | ❌ 多次翻译可能不一致 | 不推荐 |
| 每种语言各采一次 | ✅ 高（MQL5 人工本地化）| N× HTTP | ✅ MQL5 全局统一 | MQL5 不支持的语言无法覆盖 |
| **混合策略（采纳）** | ✅ 高 | 适中 | ✅ 高 | MQL5 直采 + LLM 兜底 |

#### 2.6.3 采纳方案：混合策略

1. **数值主源（canonical）**：固定从 MQL5 **英文版**（`/en/economic-calendar/content`）采集，数值字段写入 `calendar_events`，作为**唯一真值**
2. **多语言并发直采**：对 MQL5 官方支持的语言（en/zh/ja/ru/de/es/fr/pt/it/tr/ar 等）并发拉取，**只提取 `title`**，按 `event_id` 写入 `calendar_event_titles`
   - 同一会话复用 cookie，并发 errgroup，10 种语言总耗时 ≤ 单语言的 1.5×
3. **LLM 兜底**：MQL5 不支持的语言（如 id、vi）→ LLM 翻译 → 永久缓存到 `calendar_event_titles`，相同 `event_id+lang` 仅翻译一次
4. **管理员校正**：admin 可通过 `AdminService.UpdateCalendarTitle` 覆盖，`source='manual'` 优先级最高

#### 2.6.4 Schema 补充

```sql
-- 多语言标题表（独立于 calendar_events）
CREATE TABLE calendar_event_titles (
    event_id     VARCHAR(64) NOT NULL,
    lang         VARCHAR(8)  NOT NULL,        -- 'en','zh','ja','ru','de','es','fr','id',...
    title        VARCHAR(512) NOT NULL,
    description  TEXT,
    source       VARCHAR(16) NOT NULL,        -- 'mql5' | 'llm' | 'manual'
    confidence   DOUBLE PRECISION,            -- LLM 翻译置信度（0..1）；mql5/manual = 1.0
    fetched_at   TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (event_id, lang)
);
CREATE INDEX idx_titles_lang ON calendar_event_titles (lang, source);
```

#### 2.6.5 查询回退链
```
title = repo.GetTitle(event_id, user_lang)        -- 命中
     ?? repo.GetTitle(event_id, 'en')              -- 回退英文
     ?? llm.Translate(en_title, user_lang)         -- 异步翻译并写库
     ?? en_title                                   -- 最终兜底
```

#### 2.6.6 调度
- **多语言并发采集**：与英文主采集**同步触发**（同一调度循环内 errgroup），失败的语言不影响主流程
- **LLM 翻译队列**：检测到新 `event_id` 且某 `lang` 缺失 → 入队 Redis Stream → 翻译 worker 消费
- **失败重试**：MQL5 某语言失败 → 加入兜底队列 → 下次采集重试或转 LLM

#### 2.6.7 配置驱动
```yaml
calendar:
  canonical_lang: en
  mql5_languages: [en, zh, ja, ru, de, es, fr, pt, it, tr, ar]
  llm_fallback_languages: [id, vi, th]
  llm_translator: claude-sonnet      # or gemini-flash
  concurrency: 8
```

### 2.7 优化与提升

| 维度 | ark-intelligent | AntClaw |
|------|----------------|---------|
| 微采集频率 | 1 min | 30 sec |
| 去重 | 内存 map | Redis SETNX（多 worker）|
| Impact 记录 | 串行 goroutine | 异步队列（Redis Streams）|
| 历史 surprise 查询 | BadgerDB 遍历 | PostgreSQL 索引查询 O(log n) |
| Fed RSS | 单次 HTTP | 带 ETag 缓存，避免重复下载 |
| SSE 推送 | 无（仅 Telegram） | 原生 SSE + Telegram bot 双通道 |
| 多语言 | 仅英文 | MQL5 多语言直采 + LLM 边缘语言兜底 |

---

## 三、参考文件

- ark-intelligent：`internal/service/news/{fetcher,scheduler,surprise,impact_recorder,impact_bootstrap,fed_rss,analyzer}.go`, `marketdata/*/`
- AntClaw proto：`proto/antclaw/v1/calendar.proto`, `alerts.proto`
- AntClaw service：`backend/internal/service/calendar/service.go`
- 关联方案：《AntClaw-数据采集调度方案.md》
