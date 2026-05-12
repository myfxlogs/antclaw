# 07 · M4 链上 / DeFi / SEC / 情绪扩展（2 周）

> 目标：补齐 ARK 链上分析（Coinmetrics、Blockchain.com）、DeFi 分析（Defillama）、SEC EDGAR 全链路、以及情绪扩展（CBOE/MyFXBook/OpenInsider/CryptoCompare/Finviz）。

---

## 1. 范围

| 任务 | 输出 |
|---|---|
| T1 链上薄 client：Coinmetrics、Blockchain.com、CryptoCompare | `apiclient/{coinmetrics,blockchain,cryptocompare}/` |
| T2 service/onchain 增强 | analyzer.go + analysis_test.go |
| T3 DeFiService + Defillama | `apiclient/defillama/` + `service/defi/` |
| T4 SECService + SEC EDGAR | `apiclient/sec/` + `service/sec/` |
| T5 SentimentService 扩展 | CBOE/MyFXBook/OpenInsider/Finviz/CryptoCompare social |
| T6 Worker Job 调度 | 链上/DeFi/SEC/情绪共 12 个新 Job |
| T7 前端模块 | features/{onchain,defi,sec,sentiment} |

---

## 2. T1 链上薄 client

### 2.1 Coinmetrics

```
backend/internal/infra/apiclient/coinmetrics/
├── client.go
├── endpoints.go
├── types.go
└── client_test.go
```

端点：

| 方法 | 路径 |
|---|---|
| `GetMetricsTimeseries(ctx, asset, metrics, start, end)` | `/v4/timeseries/asset-metrics` |
| `GetExchangeFlow(ctx, asset, exchange, start, end)` | `/v4/timeseries/exchange-asset-metrics` |

常用 metrics：`AdrActCnt`、`TxCnt`、`FlowInExNtv`、`FlowOutExNtv`、`CapMVRVCur`、`SplyAct1yr`。

### 2.2 Blockchain.com

```
backend/internal/infra/apiclient/blockchain/
```

端点：

| 方法 | 路径 |
|---|---|
| `GetChartTimeseries(ctx, chart, timespan)` | `/charts/{chart}?format=json&timespan=...` |

常用 chart：`hash-rate`、`difficulty`、`mempool-size`、`n-transactions`、`n-unique-addresses`、`miners-revenue`、`mempool-state-by-fee-level`、`utxo-count`。

### 2.3 CryptoCompare

```
backend/internal/infra/apiclient/cryptocompare/
```

端点：

| 方法 | 路径 |
|---|---|
| `GetHistoday(ctx, fsym, tsym, limit)` | `/data/v2/histoday` |
| `GetSocialHistoday(ctx, coinId, limit)` | `/data/social/coin/histo/day` |
| `GetTop(ctx, by, limit)` | `/data/top/totalvolfull` |

需要 api_key（X-CCAGG-Apikey 或 query 参数 `api_key`）。

---

## 3. T2 service/onchain 增强

### 3.1 文件

```
backend/internal/service/onchain/
├── service.go            # 已有（M1 创建）
├── analyzer.go           # 链上 regime 分析（迁移 ARK onchain/analyzer.go）
├── coinmetrics_sync.go   # Coinmetrics 抓取 + 入库
├── blockchain_sync.go    # Blockchain.com 抓取 + 入库
└── *_test.go
```

### 3.2 Regime 判定（参照 ARK）

```
mvrv = market_cap / realized_cap
sopr_ratio = sopr_value
nupl = (market_cap - realized_cap) / market_cap

if mvrv > 3 and nupl > 0.6:    regime = euphoria
elif mvrv < 1 and nupl < 0:    regime = capitulation
elif mvrv > 2 and exchange_outflow_strong:  regime = distribution
else:                          regime = accumulation
```

`confidence` 由 z-score 累加；`narrative` 来自模板（可被 AIService 增强）。

### 3.3 数据库表

```sql
CREATE TABLE onchain_metrics_extended (
  asset TEXT NOT NULL,
  time TIMESTAMPTZ NOT NULL,
  active_addresses DOUBLE PRECISION,
  tx_count DOUBLE PRECISION,
  exchange_netflow DOUBLE PRECISION,
  mvrv DOUBLE PRECISION,
  sopr DOUBLE PRECISION,
  nupl DOUBLE PRECISION,
  realized_cap DOUBLE PRECISION,
  PRIMARY KEY (asset, time)
);

CREATE TABLE onchain_regime_history (
  asset TEXT NOT NULL,
  date DATE NOT NULL,
  regime TEXT NOT NULL,
  confidence DOUBLE PRECISION,
  narrative TEXT,
  PRIMARY KEY (asset, date)
);
```

---

## 4. T3 DeFiService + Defillama

### 4.1 Defillama 薄 client

```
backend/internal/infra/apiclient/defillama/
├── client.go
├── endpoints.go
├── types.go
└── client_test.go
```

端点：

| 方法 | 路径 |
|---|---|
| `GetProtocols(ctx)` | `/protocols` |
| `GetProtocolTVL(ctx, slug)` | `/protocol/{slug}` |
| `GetChainTVL(ctx, chain)` | `/v2/historicalChainTvl/{chain}` |
| `GetGlobalTVL(ctx)` | `/charts` |

无需鉴权。

### 4.2 service/defi

```
backend/internal/service/defi/
├── service.go
├── analyzer.go           # 迁移 ARK defi/analyzer.go
├── tvl_sync.go           # 拉取 + 入库
└── *_test.go
```

### 4.3 数据库表

```sql
CREATE TABLE defi_protocols (
  slug TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  category TEXT,
  chain TEXT,
  meta JSONB,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE defi_tvl_history (
  slug TEXT NOT NULL,
  time TIMESTAMPTZ NOT NULL,
  tvl_usd DOUBLE PRECISION NOT NULL,
  PRIMARY KEY (slug, time)
);

CREATE TABLE defi_chain_tvl (
  chain TEXT NOT NULL,
  time TIMESTAMPTZ NOT NULL,
  tvl_usd DOUBLE PRECISION NOT NULL,
  PRIMARY KEY (chain, time)
);
```

---

## 5. T4 SECService + SEC EDGAR

### 5.1 SEC 薄 client

```
backend/internal/infra/apiclient/sec/
├── client.go
├── endpoints.go
├── types.go
├── parse.go              # 提取 10-K/10-Q/8-K 关键章节
└── client_test.go
```

端点：

| 方法 | 路径 |
|---|---|
| `GetSubmissions(ctx, cik)` | `https://data.sec.gov/submissions/CIK{cik}.json` |
| `GetCompanyFacts(ctx, cik)` | `https://data.sec.gov/api/xbrl/companyfacts/CIK{cik}.json` |
| `GetFilingDocument(ctx, accession, filename)` | `https://www.sec.gov/Archives/edgar/data/{cik}/{acc_no_dashes}/{filename}` |

**关键约束**：
- SEC 要求 `User-Agent: <Name> <Email>`，否则 403
- 限速：0.1 RPS（10 次/秒以下）
- 必须缓存到本地 Postgres，避免重复抓取

### 5.2 service/sec

```
backend/internal/service/sec/
├── service.go
├── analyzer.go           # 风险评分 / 哨兵摘要
├── parser.go             # 文档解析（HTML/XBRL）
└── *_test.go
```

### 5.3 数据库表

```sql
CREATE TABLE sec_filings (
  accession_number TEXT PRIMARY KEY,
  cik TEXT NOT NULL,
  ticker TEXT,
  form_type TEXT NOT NULL,
  filed_at TIMESTAMPTZ NOT NULL,
  company_name TEXT,
  url TEXT,
  raw_excerpt TEXT,
  meta JSONB
);
CREATE INDEX idx_sec_ticker_filed ON sec_filings (ticker, filed_at DESC);

CREATE TABLE sec_company_facts (
  cik TEXT PRIMARY KEY,
  ticker TEXT,
  facts JSONB NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE sec_analysis_history (
  ticker TEXT NOT NULL,
  analyzed_at TIMESTAMPTZ NOT NULL,
  sentiment TEXT,
  risk_score DOUBLE PRECISION,
  highlights TEXT,
  PRIMARY KEY (ticker, analyzed_at)
);
```

---

## 6. T5 SentimentService 扩展

### 6.1 五个新薄 client

```
backend/internal/infra/apiclient/cboe/
backend/internal/infra/apiclient/myfxbook/
backend/internal/infra/apiclient/openinsider/
backend/internal/infra/apiclient/finviz/
backend/internal/infra/apiclient/cryptocompare_social/   # 复用 cryptocompare 子包
```

### 6.2 service/sentiment 扩展

新增文件：

```
service/sentiment/
├── service.go             # 已有
├── cboe.go                # CBOE Put/Call
├── myfxbook.go            # 散户多空比
├── openinsider.go         # 内幕交易
├── finviz.go              # 短期空单 / 机构持股
├── cryptosocial.go        # CryptoCompare social stats
├── dvol_integration.go    # DVOL 与情绪联动
└── *_test.go
```

### 6.3 数据库表

```sql
CREATE TABLE sentiment_cboe_pc (
  date DATE PRIMARY KEY,
  total_pc DOUBLE PRECISION,
  equity_pc DOUBLE PRECISION,
  index_pc DOUBLE PRECISION
);

CREATE TABLE sentiment_myfxbook (
  symbol TEXT NOT NULL,
  taken_at TIMESTAMPTZ NOT NULL,
  long_pct DOUBLE PRECISION,
  short_pct DOUBLE PRECISION,
  long_lots BIGINT,
  short_lots BIGINT,
  PRIMARY KEY (symbol, taken_at)
);

CREATE TABLE sentiment_insider_trades (
  ticker TEXT NOT NULL,
  filed_at TIMESTAMPTZ NOT NULL,
  insider TEXT,
  title TEXT,
  action TEXT,
  price DOUBLE PRECISION,
  quantity BIGINT,
  PRIMARY KEY (ticker, filed_at, insider, action)
);

CREATE TABLE sentiment_finviz (
  ticker TEXT NOT NULL,
  taken_at TIMESTAMPTZ NOT NULL,
  short_ratio DOUBLE PRECISION,
  short_pct_float DOUBLE PRECISION,
  inst_own_pct DOUBLE PRECISION,
  PRIMARY KEY (ticker, taken_at)
);

CREATE TABLE sentiment_crypto_social (
  asset TEXT NOT NULL,
  date DATE NOT NULL,
  twitter_followers_growth DOUBLE PRECISION,
  reddit_subscribers_growth DOUBLE PRECISION,
  sentiment_score DOUBLE PRECISION,
  PRIMARY KEY (asset, date)
);
```

---

## 7. T6 Worker Job 调度

新增 12 个 Job：

| Job ID | 周期 |
|---|---|
| `coinmetrics-sync` | 6h |
| `blockchain-sync` | 6h |
| `onchain-analysis` | 24h（在 Coinmetrics + Blockchain 完成后） |
| `defillama-sync` | 1h |
| `defi-analysis` | 6h |
| `sec-edgar-sync` | 24h |
| `sec-analysis` | 24h |
| `cboe-sync` | 24h |
| `myfxbook-sync` | 1h |
| `openinsider-sync` | 6h |
| `finviz-sync` | 24h |
| `crypto-social-sync` | 6h |

依赖关系：分析类 Job 依赖采集类 Job 当日数据已写入；通过 Redis key `jobs:status:<id>.last_success_at` 判断。

---

## 8. T7 前端模块

### 8.1 features/onchain（新建）

```
features/onchain/
├── pages/OnchainPage.tsx        # 选 asset → 显示 metrics 时序 + regime 卡片
├── components/MetricsChart.tsx  # 多指标叠加
├── components/RegimeCard.tsx
└── api.ts
```

### 8.2 features/defi（新建）

```
features/defi/
├── pages/DeFiPage.tsx           # Top protocols 排行 + chain TVL
├── pages/ProtocolDetailPage.tsx # 单 protocol TVL 历史
└── api.ts
```

### 8.3 features/sec（新建）

```
features/sec/
├── pages/SECPage.tsx            # 输入 ticker → list filings + analysis
├── components/FilingCard.tsx
└── api.ts
```

### 8.4 features/sentiment 扩展

新增页面：

- `CBOEPutCallPage.tsx`
- `MyFXBookPage.tsx`
- `InsiderTradesPage.tsx`
- `FinvizPage.tsx`
- `CryptoSocialPage.tsx`

---

## 9. M4 验收清单

- [ ] 8 家薄 client 全部上线，覆盖率 ≥ 70%
- [ ] OnchainService、DeFiService、SECService 全部端到端可调
- [ ] SentimentService 5 个新 RPC 可调
- [ ] 12 个新 Worker Job 启动播种成功
- [ ] 链上 regime 在 UI 可见，且 narrative 由 AI 生成（占位）
- [ ] DeFi top protocols 在 UI 可见
- [ ] SEC 输入 AAPL/MSFT 等热门 ticker 能返回 10 条 filings
- [ ] CBOE/MyFXBook/Insider/Finviz/CryptoSocial 数据均有近期记录
- [ ] 差距清单 §1.8、§1.9、§1.10 状态全部 ✅
