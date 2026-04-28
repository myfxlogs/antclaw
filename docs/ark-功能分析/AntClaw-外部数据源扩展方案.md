# AntClaw 外部数据源扩展方案

> **版本**：v1.0  
> **对应 ark-intelligent 模块**：`bis/`、`fed/`、`sec/`、`treasury/`、`worldbank/`、`imf/`、`macro/`（多客户端）  
> **对应 AntClaw Proto**：`MacroService` / `SentimentService`

---

## 一、ark-intelligent 方法分析

汇总 ark-intelligent 中所有"外部权威数据源"的客户端实现。

### 1.1 BIS（国际清算银行）— `bis/`

3 个客户端，从 BIS 公开 SDMX API 拉取：

| 文件 | 数据 | 用途 |
|------|------|------|
| `cbpol.go` | 各国央行政策利率 | 跨国利率比较，FX 套息 |
| `creditgap.go` | 信贷缺口 | 系统性风险预警（缺口 > 10% → 危机预兆）|
| `reer.go` | 实际有效汇率 | 跨货币估值（高 REER → 货币高估）|

### 1.2 Fed — `fed/fedwatch.go`
- **CME FedWatch 工具**：从 CME 期货价格反推下次 FOMC 加息/降息概率
- 抓取 CME `https://www.cmegroup.com/...fedwatch` 公开数据
- 输出：未来 3 次会议每次的加息/降息/不变概率

### 1.3 SEC 13F — `sec/`
- `client.go`：调用 SEC EDGAR API（无 auth，需 User-Agent，10 req/s 限流）
- `parser.go`：解析 13F XML（informationTable 结构）
- `analyzer.go`：聚合 N 个机构（Berkshire、Bridgewater、Renaissance、Citadel）持仓 → 排名前 N
- 缓存 7 天（季报频率）

### 1.4 Treasury 拍卖 — `treasury/`
- `client.go`：TreasuryDirect API（公开，无 key）
- `analyzer.go`：bid-to-cover ratio + 间接竞标占比
  - BTC ≥ 2.5 → 强需求；< 2.0 → 弱需求
  - 弱拍卖 → USD 走弱、收益率上行
- 缓存 12 小时

### 1.5 World Bank — `worldbank/`
- 抓取宏观基本面：GDP 增长、经常账户、CPI、外汇储备
- 按国家（USD-USA, EUR-EMU, GBP-GBR, JPY-JPN, AUD-AUS, NZD-NZL, CAD-CAN, CHF-CHE）
- 年度数据，缓存 24 小时

### 1.6 IMF WEO — `imf/weo.go`
- IMF DataMapper API
- 前瞻性数据：GDP 增长预测、通胀预测、经常账户预测
- 一年两次发布（4 月 + 10 月）

### 1.7 macro/ 多源客户端

| 文件 | 数据源 | 数据 |
|------|--------|------|
| `ecb_client.go` | 欧央行 SDW | EUR M3、HICP、政策利率 |
| `oecd_client.go` | OECD | 综合领先指标（CLI），周期前瞻 |
| `eurostat_client.go` | Eurostat JSON-stat | EU HICP / 失业率 / GDP |
| `snb_client.go` | 瑞士国家银行 | 资产负债表、FX 干预代理 |
| `dtcc_client.go` | DTCC | 外汇互换机构流（机构 USD 资金流向）|
| `tradingeconomics_client.go` | TradingEconomics（Firecrawl 抓）| G10 国 GDP/CPI/PMI/失业 |
| `treasury_client.go` | Treasury.gov 备用 | 与 FRED 互补 |

### 1.8 共性模式
所有外部数据源遵循统一模式：
- 单例 client + sync.Once
- 内存 sync.Map 缓存（TTL 4-24h）
- circuit breaker（失败 N 次冷却）
- 优雅降级：API key 缺失 → `Available=false`，不报错
- 单元测试覆盖响应解析

---

## 二、AntClaw 设计方案

### 2.1 架构

```
service/macro_extra/        (统一外部数据源服务)
  ├── bis/
  │     ├── cbpol.go
  │     ├── creditgap.go
  │     └── reer.go
  ├── fed/
  │     └── fedwatch.go
  ├── sec/                   (13F)
  ├── treasury/              (拍卖)
  ├── worldbank/
  ├── imf/                   (WEO)
  └── macro/
        ├── ecb.go
        ├── oecd.go
        ├── eurostat.go
        ├── snb.go
        ├── dtcc.go
        ├── trading_economics.go
        └── treasury_alt.go

infra/apiclient/             (HTTP 客户端，独立每个源)
  ├── bis_client.go
  ├── cme_fedwatch_client.go
  ├── sec_edgar_client.go
  ├── treasury_direct_client.go
  ├── worldbank_client.go
  ├── imf_client.go
  ├── ecb_client.go
  ├── oecd_client.go
  ├── eurostat_client.go
  ├── snb_client.go
  ├── dtcc_client.go
  ├── tradingeconomics_client.go
  └── firecrawl_client.go    (共用 web scraping)

infra/postgres/macro_extra_repo.go
```

**统一抽象**：所有外部数据源实现 `ExternalDataSource` 接口

### 2.2 核心接口

```go
type ExternalDataSource interface {
    Name() string
    Fetch(ctx context.Context) (*DataSnapshot, error)
    LastFetched() time.Time
    Available() bool
    Healthcheck(ctx context.Context) error
}

type ExternalDataRepository interface {
    SaveSnapshot(ctx, source string, payload []byte, fetchedAt time.Time) error
    GetLatest(ctx, source string) (*DataSnapshot, error)
    GetHistory(ctx, source string, since time.Time) ([]DataSnapshot, error)
}
```

### 2.3 Schema

```sql
-- 通用外部数据快照表（已在采集调度方案中定义为 data_snapshots）
-- 此处复用，按 source 字段区分

-- 各源细粒度表（高频访问）
CREATE TABLE bis_policy_rates (
    date DATE NOT NULL, currency VARCHAR(8) NOT NULL,
    rate DOUBLE PRECISION,
    PRIMARY KEY (date, currency)
);

CREATE TABLE bis_credit_gap (
    date DATE NOT NULL, country VARCHAR(8) NOT NULL,
    gap_pct DOUBLE PRECISION,
    risk_label VARCHAR(16),    -- 'NORMAL','WARN','CRITICAL'
    PRIMARY KEY (date, country)
);

CREATE TABLE bis_reer (
    date DATE NOT NULL, currency VARCHAR(8) NOT NULL,
    reer DOUBLE PRECISION,
    z_score DOUBLE PRECISION,   -- 历史百分位
    PRIMARY KEY (date, currency)
);

CREATE TABLE fed_watch_probabilities (
    snapshot_at TIMESTAMPTZ NOT NULL,
    meeting_date DATE NOT NULL,
    rate_change_bps INT NOT NULL,    -- -50, -25, 0, +25, +50
    probability DOUBLE PRECISION,
    PRIMARY KEY (snapshot_at, meeting_date, rate_change_bps)
);

CREATE TABLE sec_13f_holdings (
    institution_cik VARCHAR(16),
    quarter         VARCHAR(8),       -- '2026-Q1'
    issuer          VARCHAR(256),
    cusip           VARCHAR(16),
    value_usd       BIGINT,
    shares          BIGINT,
    fetched_at      TIMESTAMPTZ,
    PRIMARY KEY (institution_cik, quarter, cusip)
);

CREATE TABLE treasury_auctions (
    cusip           VARCHAR(16) PRIMARY KEY,
    security_type   VARCHAR(16),
    security_term   VARCHAR(16),
    auction_date    DATE,
    high_yield      DOUBLE PRECISION,
    bid_to_cover    DOUBLE PRECISION,
    indirect_pct    DOUBLE PRECISION,
    demand_label    VARCHAR(16)       -- 'STRONG','NEUTRAL','WEAK'
);

CREATE TABLE worldbank_macro (
    country         VARCHAR(8),
    currency        VARCHAR(8),
    year            INT,
    gdp_growth      DOUBLE PRECISION,
    current_account DOUBLE PRECISION,
    cpi             DOUBLE PRECISION,
    fx_reserves     DOUBLE PRECISION,
    PRIMARY KEY (country, year)
);

CREATE TABLE imf_weo (
    country         VARCHAR(8),
    indicator       VARCHAR(32),       -- 'GDP_GROWTH','CPI','CA'
    year            INT,
    is_forecast     BOOLEAN,
    value           DOUBLE PRECISION,
    vintage         VARCHAR(16),       -- '2026-04','2026-10'
    PRIMARY KEY (country, indicator, year, vintage)
);

-- ECB / OECD / Eurostat 等使用通用 macro_indicators 表
CREATE TABLE macro_indicators (
    source          VARCHAR(32) NOT NULL,    -- 'ECB','OECD','EUROSTAT','SNB','DTCC','TE'
    indicator       VARCHAR(64) NOT NULL,
    region          VARCHAR(16),
    period          DATE,
    value           DOUBLE PRECISION,
    metadata        JSONB,
    fetched_at      TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (source, indicator, region, period)
);
```

### 2.4 Redis

| Key | TTL |
|-----|-----|
| `cache:bis:cbpol` | 24h |
| `cache:bis:creditgap` | 24h |
| `cache:bis:reer` | 24h |
| `cache:fed:watch` | 1h |
| `cache:sec:13f:{cik}:{quarter}` | 7d |
| `cache:treasury:auctions` | 12h |
| `cache:worldbank:{currency}` | 24h |
| `cache:imf:weo` | 7d |
| `cache:ecb:rates` | 4h |
| `cache:te:{country}` | 6h |

### 2.5 调度

| 任务 | 频率 |
|------|------|
| `bis-cbpol-fetch` | 每日 06:00 UTC |
| `bis-creditgap-fetch` | 每周日 |
| `bis-reer-fetch` | 每周日 |
| `fed-watch-fetch` | 每小时 |
| `sec-13f-fetch` | 每周（季报发布期更频繁）|
| `treasury-auction-fetch` | 每日 22:00 |
| `worldbank-fetch` | 每月 1 日 |
| `imf-weo-fetch` | 每日（命中即缓存 7d）|
| `ecb-fetch` | 每日 |
| `oecd-fetch` | 每周 |
| `eurostat-fetch` | 每日 |
| `snb-fetch` | 每周三（瑞士 SNB 公布日）|
| `dtcc-fetch` | 每日 |
| `te-fetch` | 每 6 小时 |

### 2.6 优化与提升

| 维度 | ark-intelligent | AntClaw |
|------|----------------|---------|
| 持久化 | BadgerDB KV | PostgreSQL，结构化查询 |
| 接口统一 | 各包独立 | `ExternalDataSource` 统一接口 |
| 健康检查 | 隐式（fetch 失败）| 显式 healthcheck，配合管理员面板 |
| 缓存 | 进程内 sync.Map | Redis，多 worker 共享 |
| Schema | 各源散落 | 通用 `data_snapshots` + 高频专表 |
| 源管理 | 硬编码 | 配置驱动，可禁用 / 调频 |
| 故障告警 | 日志 | 走 `AlertsService`，管理员可见 |

---

## 三、参考文件

- ark-intelligent：`internal/service/{bis,fed,sec,treasury,worldbank,imf,macro}/*.go`
- AntClaw proto：`proto/antclaw/v1/macro.proto`
- AntClaw service：待新建 `backend/internal/service/macro_extra/`
