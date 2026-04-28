# AntClaw COT 持仓分析方案

> **版本**：v1.0  
> **对应 ark-intelligent 模块**：`internal/service/cot/` (23 files)  
> **对应 AntClaw Proto**：`COTService`（注：现有 proto 可能需补充，当前功能清单归类到 `SentimentService`）

---

## 一、ark-intelligent 方法分析

### 1.1 核心职责
COT（Commitment of Traders）分析是**机构持仓智能**的核心，从 CFTC 每周公布的 Legacy/Disaggregated 报告提取非商业、商业、管理基金、交易商、杠杆基金等分类持仓，计算多维度指数和信号。

### 1.2 数据采集（fetcher.go）

- **源**：CFTC Socrata API（Legacy + Disaggregated + TFF 三个端点）
- **频率**：周度更新（每周五 15:30 ET 发布）
- **合约**：所有主要外汇、指数期货、商品期货、加密期货（约 25+ 合约）
- **历史深度**：默认拉 5 年（260 周），用于计算百分位和 z-score
- 支持 **resiliency**：CFTC API 失败时读取本地缓存 + 告警
- 校验：date 递增、字段完整、数值范围

### 1.3 COT Index（index.go）

**3 年百分位**：
```
COTIndex = (CurrentNet - MinNet_3y) / (MaxNet_3y - MinNet_3y) * 100
```
- `[0, 100]` 区间，越高越"多头拥挤"
- 每周更新，滚动窗口

**阈值（thresholds.go）**：
- `COTBullishThreshold = 60`（偏多）
- `COTBearishThreshold = 40`（偏空）
- `COTExtremeLong = 75`（极端多头，反转风险）
- `COTExtremeShort = 25`（极端空头）

### 1.4 分析器（analyzer.go）

**核心算法**：
1. **Net Position**：非商业多仓 - 非商业空仓
2. **WoW Change**：周环比变化 + Z-score
3. **Percentile Rank**：当前 net 在历史分布的百分位
4. **Direction**：BULLISH / BEARISH / NEUTRAL
5. **Confidence**：基于 z-score + 百分位的综合置信度

输出 `COTAnalysis { Contract, NetPosition, COTIndex, Direction, SentimentScore, ... }`。

### 1.5 类别 Z-Score（category_zscore.go）

单独计算 **Dealer / LevFund / ManagedMoney / SwapDealer** 四类的 WoW 净变化 Z-score：
- 历史窗口 ≥ 4 周
- `|Z| ≥ 2.0` 触发 alert
- 跨类别 divergence 检测（如 Dealer 买 vs LevFund 卖 → 智能资金分歧）

### 1.6 重新校准检测（recalibrated_detector.go）

**核心创新 — Platt 缩放置信度**：
- 传统信号用硬编码置信度（如 "EXTREME 70%"），实际历史命中率可能只有 45%
- 通过历史回测统计每种信号类型的真实 `WinRate` 和 `AvgReturn`
- `SampleSize ≥ 5` → 用历史命中率替换硬编码
- `SampleSize ≥ 20` → 用 **Platt scaling** 拟合 `(a, b)`：`calibrated = sigmoid(a·raw + b)`
- `SampleSize ≥ 10 且 WinRate < 50%` → **抑制信号**（`Suppressed = true`）

### 1.7 Confluence Score（confluence.go + confluence_score.go）

**多因子共振评分**：综合 COT 信号 + FRED 宏观 + 价格 + 季节性 → 单一 `ConvictionScore (0-100)`

**V3 算法**：
1. COT Bias（40% 权重）：方向 + 强度
2. FRED Regime（25%）：当前 regime 是否利好该方向
3. Surprise Accumulator（20%）：本周数据超预期方向
4. Seasonal（10%）：该周历史表现
5. Price Confirmation（5%）：价格与 COT 是否同向

输出：Score + Label (`HIGH CONVICTION`, `MODERATE`, `LOW`, `CONFLICTING`)。

### 1.8 Signals（signals.go）

生成具体交易信号：
- **EXTREME_LONG/SHORT**：COT Index > 75 或 < 25
- **DIVERGENCE**：价格走势与 COT 相反
- **NET_CHANGE_SPIKE**：WoW 变化 > 2 sigma
- **CATEGORY_DIVERGE**：Dealer vs LevFund 分歧
- **SEASONAL_ALIGNED**：当前 COT 方向与该周历史均值一致

每个信号附带 `Confidence`, `ExpectedReturn`, `TimeHorizon`。

### 1.9 USD Aggregate（usd_aggregate.go）

**跨币对合成 USD 方向信号**：
- 对 EUR、GBP、JPY、AUD、CAD、CHF、NZD 七大对手反向映射为 USD 方向（EUR 多 → USD 空）
- 加权求和 → Score -100 ~ +100
- **DX Direct**：如有 DX 期货 COT，直接读取作为主信号
- **Divergence**：合成方向与 DX 直读方向不一致 → 标记分歧
- **Conviction**：同方向币对占比

### 1.10 Regime（regime.go）

**基于 COT 的 Risk-On/Risk-Off**：
- Safe haven basket：JPY + CHF + XAU
- Risk basket：AUD + NZD + CAD + GBP
- 比较两个篮子的聚合方向 → RISK-ON / RISK-OFF / UNCERTAINTY

### 1.11 Seasonal（seasonal.go）

ISO 周度 COT 净持仓均值 + stddev，与当前周比较：
- 季节性偏离 Z-score
- 52 周正常范围
- Seasonal-aligned 信号补强

---

## 二、AntClaw 设计方案

### 2.1 架构调整

```
COTService (新 proto，或并入 SentimentService)
  ├── service/cot/fetcher    (CFTC API)
  ├── service/cot/analyzer   (Index + ZScore + Percentile)
  ├── service/cot/signals    (信号生成 + Platt 校准)
  ├── service/cot/confluence (多因子共振)
  ├── service/cot/usd_aggregate (USD 合成信号)
  └── service/cot/seasonal   (季节性)
      ↓
  infra/apiclient/cftc_client.go
  infra/postgres/cot_repo.go
  infra/redis/cot_cache.go
```

### 2.2 核心接口

```go
type CFTCFetcher interface {
    FetchLegacy(ctx, contractCode string, weeks int) ([]COTRecord, error)
    FetchDisaggregated(ctx, contractCode string, weeks int) ([]COTRecord, error)
    FetchAll(ctx, weeks int) (map[string][]COTRecord, error)
}

type COTRepository interface {
    UpsertRecords(ctx, records []COTRecord) (int, error)
    GetHistory(ctx, contractCode string, weeks int) ([]COTRecord, error)
    GetLatestAll(ctx) (map[string]*COTAnalysis, error)
    SaveAnalysis(ctx, analyses []COTAnalysis) error
    SaveSignalOutcome(ctx, signalID string, outcome float64) error  // 用于 Platt 校准
}

type CalibrationRepository interface {
    GetSignalStats(ctx, signalType string) (*SignalTypeStats, error)
    UpdatePlattParams(ctx, signalType string, a, b float64) error
}
```

### 2.3 Schema

```sql
-- COT 原始记录 hypertable
CREATE TABLE cot_records (
    report_date      DATE NOT NULL,
    contract_code    VARCHAR(16) NOT NULL,
    currency         VARCHAR(8),
    noncomm_long     BIGINT, noncomm_short BIGINT,
    comm_long        BIGINT, comm_short    BIGINT,
    dealer_long      BIGINT, dealer_short  BIGINT,
    levfund_long     BIGINT, levfund_short BIGINT,
    mm_long          BIGINT, mm_short      BIGINT,
    swap_long        BIGINT, swap_short    BIGINT,
    total_oi         BIGINT,
    raw_json         JSONB,
    PRIMARY KEY (report_date, contract_code)
);
SELECT create_hypertable('cot_records', 'report_date', chunk_time_interval => INTERVAL '1 year');

-- 分析结果缓存（每周覆盖）
CREATE TABLE cot_analyses (
    report_date      DATE NOT NULL,
    contract_code    VARCHAR(16) NOT NULL,
    net_position     BIGINT,
    cot_index        DOUBLE PRECISION,
    direction        VARCHAR(16),
    sentiment_score  DOUBLE PRECISION,
    wow_change       BIGINT,
    zscore           DOUBLE PRECISION,
    percentile       DOUBLE PRECISION,
    PRIMARY KEY (report_date, contract_code)
);

-- 信号回测记录（用于 Platt 校准）
CREATE TABLE cot_signal_outcomes (
    signal_id        BIGSERIAL PRIMARY KEY,
    signal_type      VARCHAR(32) NOT NULL,
    contract_code    VARCHAR(16) NOT NULL,
    issued_at        TIMESTAMPTZ NOT NULL,
    raw_confidence   DOUBLE PRECISION,
    return_1w        DOUBLE PRECISION,   -- 1 周后 return
    return_2w        DOUBLE PRECISION,
    return_4w        DOUBLE PRECISION,
    win              BOOLEAN,
    evaluated_at     TIMESTAMPTZ
);
CREATE INDEX idx_signal_type_outcome ON cot_signal_outcomes (signal_type, win);

-- 信号校准参数
CREATE TABLE cot_calibration (
    signal_type      VARCHAR(32) PRIMARY KEY,
    platt_a          DOUBLE PRECISION,
    platt_b          DOUBLE PRECISION,
    win_rate         DOUBLE PRECISION,
    sample_size      INTEGER,
    updated_at       TIMESTAMPTZ
);
```

### 2.4 Redis 缓存

| Key | 类型 | TTL | 内容 |
|-----|------|-----|------|
| `cache:cot:latest:{contract}` | JSON | 1h | 最新分析结果 |
| `cache:cot:all` | JSON | 1h | 全市场快照 |
| `cache:cot:usd_agg` | JSON | 1h | USD 合成信号 |
| `cache:cot:regime` | JSON | 1h | Risk-On/Off |
| `cache:cot:seasonal:{contract}` | JSON | 1d | 季节性 |
| `pubsub:cot:release` | Pub/Sub | - | CFTC 发布广播 |

### 2.5 调度集成

| 任务 | 频率 | 说明 |
|------|------|------|
| `cot-release-watch` | 每 15 min (周五 15:30-17:00 ET) | 快速探测 CFTC 发布 |
| `cot-full-sync` | 每周六 01:00 UTC | 全量回溯 1 周，校正数据 |
| `cot-analyze-all` | 发布后触发 | 重算所有 analysis + USD 合成 |
| `cot-calibrate` | 每日 03:00 UTC | 扫描过去 1 年信号，更新 Platt 参数 |
| `cot-signal-evaluate` | 每日 | 为 1/2/4 周前的信号评估 outcome |

### 2.6 优化与提升

| 维度 | ark-intelligent | AntClaw |
|------|----------------|---------|
| 持久化 | BadgerDB KV | TimescaleDB，支持 SQL 回溯分析 |
| 信号校准 | 单进程内存 | DB 持久化 + Redis 广播 |
| 跨合约合成 | 每次扫描 | 预计算缓存 |
| 历史 outcome | 无独立表 | 专表 `cot_signal_outcomes`，可审计 |
| Confluence | 实时计算 | 核心信号预计算 + 用户请求时组合 |
| 发布探测 | 周期轮询 | 事件驱动（周五窗口内高频探测）|

---

## 三、参考文件

- ark-intelligent：`internal/service/cot/{fetcher,analyzer,index,confluence,confluence_score,recalibrated_detector,category_zscore,regime,seasonal,signals,usd_aggregate,thresholds}.go`
- AntClaw proto：尚需在 proto 中补充 `COTService`（或纳入 `SentimentService`）
- AntClaw service：待新建 `backend/internal/service/cot/`
