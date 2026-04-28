# AntClaw FRED 宏观数据方案

> **版本**：v1.0  
> **对应 ark-intelligent 模块**：`internal/service/fred/` (23 files)  
> **对应 AntClaw Proto**：`MacroService`

---

## 一、ark-intelligent 方法分析

### 1.1 核心职责
FRED 模块是**宏观数据的中枢**，从美联储 FRED API 拉取 50+ 指标，计算复合指数、宏观状态分类、收益率利差、资产回报矩阵，并对外提供 Fed 言论语调分析。

### 1.2 数据采集与缓存（fetcher.go + cache.go）

**采集**：
- 并发抓取（最多 10 并行）50+ 个 FRED 系列（T10Y2Y, UNRATE, CPIAUCSL, NFCI, FEDFUNDS...）
- 每个 series 独立错误处理，单点失败不影响整体
- Panic recovery，保证 goroutine 不崩溃

**缓存**：
- 全局单例 `cachedMacroData`，TTL 5 min
- `GetCachedOrFetch`：命中返回，过期重新抓
- `InvalidateCache`：外部显式刷新（如 Fed 讲话后触发）
- `postFetchHook`：抓取成功后触发持久化、下游监听

### 1.3 宏观状态分类（regime.go + regime_asset.go）

**`ClassifyMacroRegime(MacroData) → Regime`**：
- 规则引擎：多维度打分（收益率曲线、通胀、劳动、金融条件）
- 输出六类状态：
  - `INFLATIONARY`（通胀高 + 劳动紧）
  - `GOLDILOCKS`（低通胀 + 稳健增长）
  - `STAGFLATION`（高通胀 + 弱增长）
  - `DEFLATION`（通缩 + 衰退）
  - `STRESS`（金融条件紧 + 信贷利差宽）
  - `NEUTRAL`（混合）

**历史回测（regime_history.go）**：
- 拉取过去 5 年的 FRED 数据，按周重构 `RegimeSnapshot`
- `regime_performance.go`：在每个 regime 下统计各货币的周度 return
- `regime_asset.go`：输出资产-状态矩阵（如 "通胀期 USD +1.2%, GOLD +2.5%"）

### 1.4 复合指数（composites.go）

**Macro Component Score**（-100 ~ +100）：
- 收益率曲线（25 分权重）：正值 = 健康，深度倒挂 = 衰退信号
- 通胀目标偏离（CPI 离 2%）
- NFCI 金融条件（连续值而非阈值，避免边界抖动）
- 劳动市场（失业率 + 非农）
- GDP 增长

**单一分数**避免了原先 "stress + FRED" 分别计算导致的 NFCI 三重计算 bug。

### 1.5 利率差与套息（rate_differential.go + carry_monitor.go）

**`FetchCarryRanking`**：
- 抓取主要货币政策利率（USD Fed Funds, EUR EONIA, GBP BoE, JPY BoJ, AUD, NZD, CAD, CHF）
- 计算相对 USD 的差值，排序输出 `CarryRanking`
- **carry unwind detection**：监控利差区间（max - min）环比变化；区间收窄 > 20% 视为套息平仓信号

### 1.6 Fed 讲话监控（speeches.go）

- 通过 Firecrawl 服务抓取 https://federalreserve.gov/newsevents/speeches.htm
- 关键词分类 HAWKISH/DOVISH/NEUTRAL
- 缓存 6 小时；若 `FIRECRAWL_API_KEY` 缺失则 `Available=false`（优雅降级）

### 1.7 Alerts（alerts.go）
- 监测 NFCI、失业率、通胀、收益率曲线阈值触发
- Regime 切换触发广播事件
- 防抖：同一指标 1 小时内只报警一次

### 1.8 持久化（persistence.go）
- `SaveSnapshots`：每次 FetchMacroData 成功后持久化所有 series 观测值到 BadgerDB
- 记录 `SeriesID + Date + Value`

---

## 二、AntClaw 设计方案

### 2.1 架构调整

```
MacroService (Proto)
  ├── service/macro/fred      (FRED 领域逻辑)
  ├── service/macro/regime    (宏观状态分类)
  ├── service/macro/composite (复合指数)
  ├── service/macro/carry     (利率差+套息)
  └── service/macro/speeches  (Fed 讲话)
      ↓
  infra/apiclient/fred_client.go
  infra/apiclient/firecrawl_client.go
      ↓
  infra/postgres/macro_repo.go   (TimescaleDB)
  infra/redis/macro_cache.go     (热缓存)
```

### 2.2 核心接口

```go
type FREDFetcher interface {
    FetchSeries(ctx, seriesID string, limit int) ([]Observation, error)
    FetchBatch(ctx, seriesIDs []string, limit int) (map[string][]Observation, error)
}

type MacroRepository interface {
    SaveObservations(ctx, obs []Observation) (int, error)            // 增量 UPSERT
    GetLatest(ctx, seriesID string) (*Observation, error)
    GetHistory(ctx, seriesID string, from, to time.Time) ([]Observation, error)
    GetRegimeHistory(ctx, from, to time.Time) ([]RegimeSnapshot, error)
}

type RegimeClassifier interface {
    Classify(data *MacroData) *RegimeResult
    BacktestHistory(ctx, lookbackYears int) ([]RegimeSnapshot, error)
}
```

### 2.3 持久化 Schema

```sql
-- 复用 data_snapshots (source='fred')，另建分析视图
CREATE TABLE macro_regime_history (
    time     TIMESTAMPTZ NOT NULL,
    regime   VARCHAR(32) NOT NULL,
    score    DOUBLE PRECISION,
    details  JSONB,
    PRIMARY KEY (time)
);
SELECT create_hypertable('macro_regime_history', 'time', chunk_time_interval => INTERVAL '1 year');

-- 资产-状态矩阵预聚合
CREATE MATERIALIZED VIEW macro_regime_asset_perf
WITH (timescaledb.continuous) AS
SELECT time_bucket('1 week', p.time) AS week,
       r.regime, p.symbol,
       avg((p.close - lag(p.close) OVER w) / lag(p.close) OVER w * 100) AS weekly_ret
FROM price_daily p
JOIN macro_regime_history r ON time_bucket('1 day', p.time) = time_bucket('1 day', r.time)
WINDOW w AS (PARTITION BY p.symbol ORDER BY p.time)
GROUP BY week, r.regime, p.symbol;
```

### 2.4 Redis 缓存

| Key | 类型 | TTL | 内容 |
|-----|------|-----|------|
| `cache:macro:data` | JSON | 5m | 最新 MacroData 快照 |
| `cache:macro:regime` | JSON | 5m | 当前 regime |
| `cache:macro:carry` | JSON | 15m | CarryRanking |
| `cache:macro:composite` | JSON | 5m | 复合指数 |
| `cache:fed:speeches` | JSON | 6h | Fed 讲话列表 |
| `pubsub:macro:regime_change` | Pub/Sub | - | regime 切换事件 |

### 2.5 调度集成

| 任务 | 频率 | 说明 |
|------|------|------|
| `fred-full-sync` | 每 5 分钟 | 拉取 50+ 核心 series，UPSERT |
| `fred-regime-classify` | 每 5 分钟 | 分类并写 `macro_regime_history` |
| `fred-carry-monitor` | 每小时 | 利差监控 + unwind 检测 |
| `fed-speeches-scrape` | 每 6 小时 | Firecrawl 抓取 |
| `fred-regime-backtest` | 每日 02:00 | 回溯 5 年 regime，填补历史 |

### 2.6 vs ark-intelligent 提升

| 维度 | ark-intelligent | AntClaw |
|------|----------------|---------|
| 持久化 | BadgerDB | TimescaleDB（复杂查询 + 压缩）|
| 缓存 | 进程内 map | Redis，多 worker 共享 |
| Regime 历史 | 每次按需计算 | 连续聚合，查询 O(log n) |
| 资产矩阵 | 实时扫全量 | 物化视图，预聚合 |
| Fed 讲话 | Firecrawl 单源 | Firecrawl + RSS 双路 |
| Alerts | 内嵌打印 | 走统一 `AlertsService` + SSE |
| 并发 | 10 并行抓取 | 可配置 + 限流 |

---

## 三、参考文件

- ark-intelligent：`internal/service/fred/{fetcher,cache,composites,regime,regime_asset,regime_history,rate_differential,carry_monitor,speeches,worldbank,alerts,persistence}.go`
- AntClaw proto：`proto/antclaw/v1/macro.proto`
- AntClaw service：`backend/internal/service/macro/service.go`
