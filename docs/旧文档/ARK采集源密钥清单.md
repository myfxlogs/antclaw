# ARK-Intelligent 数据采集源 API Key 清单

> 来源代码库：`Emulator/ark-intelligent/internal/service/**` 与 `internal/adapter/**`
>
> 整理范围：所有出站 HTTP/HTTPS 采集端点；按是否需要 API Key 分类，并给出对应环境变量、源码位置与用途。
>
> 生成时间：2026-04-27

---

## 一、需要 API Key 的采集源

下列数据源**必须配置环境变量**才能正常采集；未配置时通常自动降级或跳过对应功能。

| 序号 | 数据源 | 用途 | Endpoint | 环境变量 | 源码位置 |
|---|---|---|---|---|---|
| 1 | TwelveData | 周线 / 日内 K 线（FX、指数、商品） | `https://api.twelvedata.com/time_series` | `TWELVE_DATA_API_KEYS`（逗号分隔，多 Key 轮询）；兼容 `TWELVE_DATA_API_KEY` | `service/price/fetcher.go`、`service/price/intraday_fetcher.go`、`service/price/daily_fetcher.go` |
| 2 | Alpha Vantage | FX 周线、油 / 黄金 / 国债收益率备份 | `https://www.alphavantage.co/query?function=...&apikey=...` | `ALPHAVANTAGE_API_KEY`（多源回退链中调用） | `service/price/fetcher.go` |
| 3 | EIA（美国能源部） | 原油 / RBOB / ULSD 库存与炼厂利用率 | `https://api.eia.gov/v2/seriesid/{id}?api_key=...` | `EIA_API_KEY` | `service/price/eia.go` |
| 4 | CoinGecko Demo | BTC 主导率、altcoin 总市值、加密情绪 | `https://api.coingecko.com/api/v3/...`（公共端点）<br>`https://pro-api.coingecko.com/api/v3/...`（Demo Key） | `COINGECKO_API_KEY`（无 key 时降级到公共匿名端点，但限速更严） | `service/price/fetcher.go`、`service/marketdata/coingecko/client.go` |
| 5 | FRED（圣路易斯联储） | 收益率曲线、就业、通胀、TGA、Fed 资产负债表等宏观序列 | `https://api.stlouisfed.org/fred/series/observations` | `FRED_API_KEY` | `service/fred/fetcher.go` |
| 6 | Firecrawl | 网页爬取的统一代理（驱动以下所有抓取页面） | `https://api.firecrawl.dev/v1/scrape` | `FIRECRAWL_API_KEY` | 见下方 §三 |
| 7 | Massive（原 Polygon.io） | 历史面板研究、衍生市场数据 | `https://api.massive.com` | `MASSIVE_API_KEYS`（多 Key 轮询）；兼容 `MASSIVE_API_KEY` | `service/marketdata/massive/client.go` |
| 8 | Massive S3 Flat Files | 历史 panel 数据下载（可选） | S3 兼容协议 | `MASSIVE_S3_ACCESS_KEY` / `MASSIVE_S3_SECRET_KEY` | 配置项（按需启用） |
| 9 | Bybit（私有端点） | 资金费率历史、账户级数据（仅当走私有 API 时需要） | `https://api.bybit.com`（需 `Authorization`/签名） | `BYBIT_API_KEY` / `BYBIT_API_SECRET` | `service/marketdata/bybit/client.go` |
| 10 | Telegram Bot API | Bot 收发消息 / 文件下载（非"市场采集"，但属对外 HTTPS 出站） | `https://api.telegram.org/bot{token}` | `BOT_TOKEN` | `adapter/telegram/wiring.go`、`adapter/telegram/api.go` |
| 11 | Google Gemini | AI 报告生成 / 自然语言分析 | Google AI Studio SDK | `GEMINI_API_KEY` | `service/ai/gemini.go` |
| 12 | Anthropic Claude（自建网关） | AI Chatbot（通过 `marketriskmonitor.com` 代理转发） | `CLAUDE_ENDPOINT`（默认 `https://marketriskmonitor.com/api/analyze`） | `CLAUDE_ENDPOINT`（内部鉴权由网关自身完成） | `service/ai/claude.go` |

---

## 二、不需要 API Key 的采集源（公共开放数据）

下列源**完全无需鉴权**，仅依赖 `User-Agent` 等基础 HTTP 头即可访问，多为政府/官方/开放数据 API 与公开 CSV/JSON 端点。

### 2.1 持仓与衍生品

| 数据源 | 用途 | Endpoint | 源码位置 |
|---|---|---|---|
| CFTC Socrata（TFF Combined） | 交易者持仓周报 | `https://publicreporting.cftc.gov/resource/yw9f-hn96.json` | `service/cot/fetcher.go` |
| CFTC Socrata（Disaggregated Combined） | 交易者持仓周报（细分） | `https://publicreporting.cftc.gov/resource/kh3c-gbw2.json` | `service/cot/fetcher.go` |
| CFTC Socrata（TFF Futures-Only） | TFF 期货专用 | `https://publicreporting.cftc.gov/resource/gpe5-46if.json` | `service/cot/fetcher.go` |
| CFTC Socrata（Disagg Futures-Only） | Disaggregated 期货专用 | `https://publicreporting.cftc.gov/resource/72hh-3qpy.json` | `service/cot/fetcher.go` |
| CFTC Legacy CSV | COT 历史 CSV 兜底 | `https://www.cftc.gov/dea/newcot/deafut.txt` | `service/cot/fetcher.go` |
| DTCC PPD | FX 远期/掉期/期权累计名义量 | `https://pddata.dtcc.com/ppd/api/report/cumulative/CFTC/FOREX` | `service/macro/dtcc_client.go` |

### 2.2 经济数据/政府公开 API

| 数据源 | 用途 | Endpoint | 源码位置 |
|---|---|---|---|
| World Bank | 各国 GDP、经常账户、CPI | `https://api.worldbank.org/v2/country/{code}/indicator/{series}` | `service/worldbank/client.go`、`service/fred/worldbank.go` |
| ECB Statistical Data Warehouse | 欧元区货币政策指标 | `https://data-api.ecb.europa.eu/service/data` | `service/macro/ecb_client.go` |
| Eurostat | 欧元区经济统计 | `https://ec.europa.eu/eurostat/api/dissemination/statistics/1.0/data` | `service/macro/eurostat_client.go` |
| OECD SDMX | 综合领先指数（CLI） | `https://sdmx.oecd.org/public/rest/data` | `service/macro/oecd_client.go` |
| SNB（瑞士央行） | 瑞郎流动性 / 干预迹象 | `https://data.snb.ch/api/cube/snbbipo/data/csv/en` | `service/macro/snb_client.go` |
| US Treasury | 每日国债收益率曲线（实际/名义）CSV | `https://home.treasury.gov/resource-center/data-chart-center/interest-rates/daily-treasury-rates.csv/...` | `service/macro/treasury_client.go` |
| TreasuryDirect | 国债拍卖结果 | `https://www.treasurydirect.gov/TA_WS/securities/search` | `service/treasury/client.go` |
| BIS Stats（Credit Gap） | 信贷缺口（早期金融危机预警） | `https://stats.bis.org/api/v2/data/BIS,WS_CREDIT_GAP,1.0` | `service/bis/creditgap.go` |
| BIS Stats（REER/NEER） | 实际/名义有效汇率 | `https://stats.bis.org/api/v2/data/BIS,WS_EER,1.0` | `service/bis/reer.go` |
| BIS Stats（CBPOL） | 各国央行政策利率 | `https://stats.bis.org/api/v2/data/BIS,WS_CBPOL,1.0` | `service/bis/cbpol.go` |
| IMF DataMapper | WEO GDP 与经常账户预测 | `https://www.imf.org/external/datamapper/api/v1/{indicator}/{countries}` | `service/imf/weo.go` |
| Federal Reserve RSS（讲话） | Fed 讲话信息流 | `https://www.federalreserve.gov/feeds/speeches.xml` | `service/news/fed_rss.go` |
| Federal Reserve RSS（FOMC 新闻稿） | FOMC 货币政策新闻 | `https://www.federalreserve.gov/feeds/press_monetary.xml` | `service/news/fed_rss.go` |
| MQL5 Economic Calendar | 财经日历事件（隐藏 POST API） | `https://www.mql5.com/en/economic-calendar/content` | `service/news/fetcher.go` |
| SEC EDGAR | 13F 机构持仓（仅需 `User-Agent`） | `https://data.sec.gov/submissions/CIK*.json`、`https://www.sec.gov/Archives/edgar/data` | `service/sec/client.go` |

### 2.3 价格 / 量化指标（公共行情）

| 数据源 | 用途 | Endpoint | 源码位置 |
|---|---|---|---|
| Yahoo Finance | 主流 K 线兜底（FX、指数、商品、^MOVE） | `https://query1.finance.yahoo.com/v8/finance/chart/{symbol}`<br>`https://query2.finance.yahoo.com/v8/finance/chart/{symbol}` | `service/price/fetcher.go`、`service/price/daily_fetcher.go`、`service/price/intraday_fetcher.go`、`service/vix/move.go` |
| Stooq | FX 周线 CSV 历史 | `https://stooq.com/q/d/l/?s={pair}.fx&i=w` | `service/price/stooq.go` |
| CBOE CDN（VX/VIX/VVIX/SKEW/OVX/GVZ/RVX/VIX9D/COR3M） | 波动率指数 EOD CSV | `https://cdn.cboe.com/api/global/us_indices/daily_prices/{INDEX}_EOD.csv` | `service/vix/fetcher.go`、`service/vix/vol_suite.go` |
| CoinGecko 公共端点（无 Demo Key 时） | 加密总览 / 单币市值 | `https://api.coingecko.com/api/v3/global`、`/coins/{id}/market_chart` | `service/price/fetcher.go` |
| Coin Metrics Community API | 链上指标（活跃地址、转账等） | `https://community-api.coinmetrics.io/v4/timeseries/asset-metrics` | `service/onchain/coinmetrics.go` |
| Blockchain.info | BTC 算力、内存池、链上统计 | `https://api.blockchain.info/charts`、`https://api.blockchain.info/stats` | `service/onchain/blockchain_client.go` |
| DeFiLlama（Protocols） | DeFi 协议 TVL | `https://api.llama.fi/v2/protocols` | `service/defi/client.go` |
| DeFiLlama（Chains） | 各公链 TVL | `https://api.llama.fi/v2/chains` | `service/defi/client.go` |
| DeFiLlama（DEX） | DEX 交易量概览 | `https://api.llama.fi/overview/dexs` | `service/defi/client.go` |
| DeFiLlama（Stablecoins） | 稳定币市值 | `https://stablecoins.llama.fi/stablecoins` | `service/defi/client.go` |
| DeFiLlama（Historical TVL） | 历史 TVL 序列 | `https://api.llama.fi/v2/historicalChainTvl` | `service/marketdata/defillama/client.go` |
| CryptoCompare 公共 | 价格备用源 | `https://min-api.cryptocompare.com` | `service/marketdata/cryptocompare/client.go` |
| Deribit Public | 期权 IV / 行权链 | `https://www.deribit.com/api/v2/public` | `service/marketdata/deribit/client.go` |
| Bybit Public（行情） | 订单簿 / 资金费率（只读市场数据） | `https://api.bybit.com`（公共路径无需鉴权） | `service/marketdata/bybit/client.go` |

### 2.4 情绪 / 风险（直连公共 JSON）

| 数据源 | 用途 | Endpoint | 源码位置 |
|---|---|---|---|
| CNN Fear & Greed | 美股情绪指数 | `https://production.dataviz.cnn.io/index/fearandgreed/graphdata` | `service/sentiment/sentiment.go` |
| Alternative.me（Crypto F&G） | 加密贪婪/恐慌指数 | `https://api.alternative.me/fng/?limit=2` | `service/sentiment/sentiment.go` |
| Alternative.me（v2 全局） | 加密总市值 / Tickers | `https://api.alternative.me/v2/global/`、`https://api.alternative.me/v2/ticker/` | `service/sentiment/sentiment.go` |

---

## 三、需要 Firecrawl（`FIRECRAWL_API_KEY`）间接抓取的网页

下列页面本身没有公开 API；ark-intelligent 通过 Firecrawl 的 `POST https://api.firecrawl.dev/v1/scrape` 进行结构化提取。**因此使用这些数据的前提是已配置 `FIRECRAWL_API_KEY`**。

| 目标页面 | 数据用途 | 源码位置 |
|---|---|---|
| `https://www.aaii.com/sentimentsurvey` | AAII 个人投资者情绪调查 | `service/sentiment/sentiment.go` |
| `https://www.myfxbook.com/community/outlook` | 散户多空分布 | `service/sentiment/myfxbook.go` |
| `https://openinsider.com/latest-cluster-buys` | 内部人集中买入 | `service/sentiment/openinsider.go` |
| `https://www.cboe.com/us/options/market_statistics/daily/` | CBOE Put/Call Ratio | `service/sentiment/cboe.go` |
| `https://www.cmegroup.com/markets/interest-rates/cme-fedwatch-tool.html` | CME FedWatch 加息概率 | `service/fed/fedwatch.go` |
| `https://tradingeconomics.com/{country}/{indicator}` | TradingEconomics 国家级指标 | `service/macro/tradingeconomics_client.go` |
| `https://finviz.com/futures.ashx` | 期货热力图 | `service/marketdata/finviz/client.go` |
| `https://finviz.com/groups.ashx?g=sector&v=140` | 行业板块绩效 | `service/marketdata/finviz/client.go` |
| `https://www.federalreserve.gov/newsevents/speeches.htm` | Fed 讲话列表（鹰/鸽分类） | `service/fred/speeches.go` |

---

## 四、汇总速查表

### 必须配置的环境变量

```env
# Bot 与 AI（Bot 必需；AI 二选一或同时配置）
BOT_TOKEN=
GEMINI_API_KEY=
CLAUDE_ENDPOINT=

# 行情（强烈建议至少配 TwelveData）
TWELVE_DATA_API_KEYS=
ALPHAVANTAGE_API_KEY=

# 宏观/能源/情绪
FRED_API_KEY=
EIA_API_KEY=
COINGECKO_API_KEY=

# 网页爬取统一代理（用于 §三 列表）
FIRECRAWL_API_KEY=

# 可选 - 更深度数据
MASSIVE_API_KEYS=
MASSIVE_S3_ACCESS_KEY=
MASSIVE_S3_SECRET_KEY=
BYBIT_API_KEY=
BYBIT_API_SECRET=
```

### 关键行为约定

- **降级策略**：缺失非必需 Key 时仅停用对应模块（`config.go` 启动期日志会打印 `WARN`），核心功能不受影响。
- **多 Key 轮询**：`TWELVE_DATA_API_KEYS` 与 `MASSIVE_API_KEYS` 支持逗号分隔多 Key，按 round-robin 调度规避免费额度限制。
- **回退链**：行情采集顺序为 TwelveData → AlphaVantage → Yahoo → Stooq → CoinGecko，任意一层失败自动切换下一层。
- **Firecrawl 是网页爬取的"瓶颈"**：§三 中所有数据源都依赖单一 `FIRECRAWL_API_KEY`，建议为生产环境单独申请并监控配额。
- **公共数据源限速**：政府/公开 API（CFTC、World Bank、BIS、ECB、Eurostat、SEC EDGAR 等）虽然无需 Key，但都依赖 `User-Agent` 头与轻量限速；代码已使用 `httpclient` + `circuitbreaker` 做防御。

---

## 五、维护说明

- 新增数据源时，请在对应 §一 / §二 / §三 表格中追加一行，并补全：用途、Endpoint、环境变量、源码位置。
- 端点变更（含 base URL、参数格式）请同步本文件与代码注释，避免文档与实现漂移。
- 任何敏感 Key 一律使用环境变量或 KMS / sops 注入，**禁止**入库到 `.env` 真实值。
