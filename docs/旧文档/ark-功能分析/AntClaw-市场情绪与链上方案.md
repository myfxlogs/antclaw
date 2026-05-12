# AntClaw 市场情绪与链上分析方案

> **版本**：v1.0  
> **对应 ark-intelligent 模块**：`sentiment/`、`onchain/`、`defi/`  
> **对应 AntClaw Proto**：`SentimentService`

---

## 一、ark-intelligent 方法分析

### 1.1 Sentiment 模块（sentiment/）

**主入口 `sentiment.go`（30KB）整合多源情绪指标**：

| 子源 | 文件 | 数据 |
|------|------|------|
| **CBOE Put/Call** | `cboe.go` | 抓 CBOE EOD CSV → 当前 P/C ratio + 历史百分位 |
| **MyFXBook** | `myfxbook.go` | 散户多空持仓比（外汇）— 作为反向指标 |
| **OpenInsider** | `openinsider.go` | 美股内部人买卖（CEO/CFO 行为）|
| **DVOL Integration** | `dvol_integration.go` | 桥接加密 DVOL 数据进入 Sentiment |
| **Cache** | `cache.go` | 多源 TTL 缓存（4-12h）|

**SentimentData 输出**：
- VIX 当前 + 12h 变化、Term structure（contango/backwardation）
- MOVE 债券波动率、SKEW 尾部风险
- Fear & Greed Index（CNN 风格综合 0-100）
- DVOL BTC/ETH（加密 VIX）
- P/C ratio + 百分位
- 散户头寸占比
- 内部人净买卖

**评分逻辑**：
- 多维加权 → 单一 Sentiment Score（-100 极度恐惧 ~ +100 极度贪婪）
- Regime 标签：`COMPLACENCY`/`NORMAL`/`STRESS`/`PANIC`

### 1.2 OnChain 模块（onchain/）

**两大数据源**：
- `coinmetrics.go`：CoinMetrics community API → BTC/ETH 链上指标
  - Exchange in/out flow（交易所流入流出 → 抛压指标）
  - Active addresses（活跃地址）
  - Tx count（交易笔数）
- `blockchain_client.go`：blockchain.info → BTC 网络健康
  - Mempool 大小（拥堵 → 手续费 → 持有意愿）
  - Hash rate
  - Difficulty

**`analyzer.go`**：
- 7 日 / 30 日累计 net flow（负值 = 累积，正值 = 抛压）
- 活跃地址趋势（增长 = 网络健康）
- 综合 `OnChainScore`

### 1.3 DeFi 模块（defi/）

**单一数据源 DefiLlama**：
- Total TVL + 24h/7d 变化
- 主要协议 TVL 排名
- DEX 24h 交易量
- 链 TVL 分布
- 稳定币市值

**信号检测（analyzer.go）**：
- TVL 24h 跌幅 > 5% → `risk_off` alert
- TVL 升幅 > 5% → `risk_on` 信号
- 主要协议 TVL 异动 → 协议风险预警

---

## 二、AntClaw 设计方案

### 2.1 架构

```
SentimentService (Proto)
  ├── service/sentiment/aggregator   (多源整合)
  ├── service/sentiment/cboe
  ├── service/sentiment/myfxbook
  ├── service/sentiment/openinsider
  ├── service/sentiment/feargreed
  ├── service/sentiment/onchain      (CoinMetrics + blockchain.info)
  └── service/sentiment/defi         (DefiLlama)
      ↓
  infra/apiclient/{cboe,myfxbook,coinmetrics,defillama,blockchain_info}_client.go
  infra/postgres/sentiment_repo.go
```

### 2.2 接口

```go
type SentimentSource interface {
    Name() string
    Fetch(ctx context.Context) (*SentimentSubResult, error)
}

type SentimentAggregator interface {
    Aggregate(ctx context.Context) (*SentimentData, error)
    GetScore(ctx context.Context) (float64, string, error)  // score + regime label
}

type OnChainFetcher interface {
    FetchExchangeFlows(ctx, asset string, days int) ([]ExchangeFlow, error)
    FetchActiveAddresses(ctx, asset string, days int) ([]ActiveAddrMetric, error)
    FetchNetworkHealth(ctx) (*BTCNetworkHealth, error)
}

type DeFiFetcher interface {
    FetchProtocols(ctx) ([]ProtocolTVL, error)
    FetchChains(ctx) ([]ChainTVL, error)
    FetchDEXVolume(ctx) (*DEXVolume, error)
    FetchStablecoins(ctx) (*StablecoinSnapshot, error)
}
```

### 2.3 Schema

```sql
-- 情绪快照
CREATE TABLE sentiment_snapshots (
    time          TIMESTAMPTZ PRIMARY KEY,
    score         DOUBLE PRECISION,        -- -100..+100
    regime        VARCHAR(16),
    pc_ratio      DOUBLE PRECISION,
    pc_percentile DOUBLE PRECISION,
    fear_greed    DOUBLE PRECISION,
    retail_long_pct DOUBLE PRECISION,
    insider_net   DOUBLE PRECISION,
    raw           JSONB
);
SELECT create_hypertable('sentiment_snapshots', 'time', chunk_time_interval => INTERVAL '30 days');

-- 链上指标
CREATE TABLE onchain_metrics (
    date            DATE NOT NULL,
    asset           VARCHAR(8) NOT NULL,    -- 'BTC','ETH'
    flow_in         DOUBLE PRECISION,
    flow_out        DOUBLE PRECISION,
    net_flow        DOUBLE PRECISION,
    active_addr     BIGINT,
    tx_count        BIGINT,
    onchain_score   DOUBLE PRECISION,
    PRIMARY KEY (date, asset)
);
SELECT create_hypertable('onchain_metrics', 'date', chunk_time_interval => INTERVAL '180 days');

-- DeFi 快照
CREATE TABLE defi_snapshots (
    time           TIMESTAMPTZ PRIMARY KEY,
    total_tvl      DOUBLE PRECISION,
    tvl_change_24h DOUBLE PRECISION,
    tvl_change_7d  DOUBLE PRECISION,
    dex_vol_24h    DOUBLE PRECISION,
    stablecoin_mc  DOUBLE PRECISION,
    raw            JSONB
);
```

### 2.4 Redis

| Key | TTL | 内容 |
|-----|-----|------|
| `cache:sentiment:agg` | 30m | 综合情绪结果 |
| `cache:sentiment:pc` | 4h | CBOE P/C |
| `cache:sentiment:fxretail` | 1h | MyFXBook |
| `cache:onchain:btc` | 6h | BTC 链上 |
| `cache:onchain:eth` | 6h | ETH 链上 |
| `cache:defi:total` | 4h | DeFi TVL |

### 2.5 调度

| 任务 | 频率 |
|------|------|
| `sentiment-cboe` | 每 4 小时 |
| `sentiment-myfxbook` | 每小时 |
| `sentiment-insider` | 每日 02:00 UTC |
| `sentiment-feargreed` | 每 30 分钟 |
| `onchain-coinmetrics` | 每 6 小时 |
| `onchain-blockchain-info` | 每 4 小时 |
| `defi-llama` | 每 4 小时 |

### 2.6 优化与提升

| 维度 | ark-intelligent | AntClaw |
|------|----------------|---------|
| 多源协调 | 串行抓取 | 并发 errgroup + 局部失败容忍 |
| 持久化 | BadgerDB | TimescaleDB（历史趋势查询）|
| Score 计算 | 实时 | 缓存 + 增量重算 |
| 报警 | 内嵌打印 | 走 `AlertsService` + SSE |
| 历史回溯 | 仅最新 | 完整时间序列，可叠加价格做相关分析 |

---

## 三、参考文件

- ark-intelligent：`internal/service/sentiment/*`, `onchain/*`, `defi/*`
- AntClaw proto：`proto/antclaw/v1/sentiment.proto`
- AntClaw service：`backend/internal/service/sentiment/service.go`
