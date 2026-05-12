# 02 · Vendor 薄客户端规范

> 所有第三方数据源必须以「薄 client」形式落入 `backend/internal/infra/apiclient/<vendor>/`。
> 业务派生指标住 `backend/internal/service/<capability>/`，不得在 vendor 包内承载业务语义。
> 本文给出统一接口契约、目录结构、限流/重试/断路器/密钥/指标规范，以及 18 家 vendor 的落位计划。

---

## 1. apiclient.Source 接口契约

### 1.1 接口定义（必须先实现）

**文件**：`backend/internal/infra/apiclient/source.go`

```go
package apiclient

import (
    "context"
    "net/http"
    "time"
)

// Source 表示一个数据源（vendor）的统一接入抽象。
// 所有 vendor 子包必须经由 Source 中间件链发起请求。
type Source interface {
    // Vendor 返回数据源标识（小写无空格，例如 "deribit"）。
    Vendor() string

    // Do 发起一次 HTTP 调用，自动应用：限流、重试、断路器、超时、指标埋点。
    // 调用方仅负责构造 *http.Request 与解析响应。
    Do(ctx context.Context, req *http.Request) (*http.Response, error)

    // Healthz 返回最近一次成功调用时间与错误率，用于 /datasources 面板展示。
    Healthz(ctx context.Context) HealthSnapshot
}

type HealthSnapshot struct {
    Vendor          string    `json:"vendor"`
    LastSuccessAt   time.Time `json:"last_success_at"`
    LastError       string    `json:"last_error,omitempty"`
    Window          string    `json:"window"`             // 例如 "5m"
    RequestsTotal   int64     `json:"requests_total"`
    ErrorsTotal     int64     `json:"errors_total"`
    P95LatencyMs    int64     `json:"p95_latency_ms"`
}
```

### 1.2 中间件链（按顺序）

`Source.Do` 内部按以下顺序处理：

1. **超时控制**：每次请求强制 `ctx, cancel := context.WithTimeout(ctx, vendorTimeout)`；默认 15s，可在 `data_sources` 表 `timeout_ms` 字段覆盖。
2. **限流**：使用 `internal/infra/redis/rate_limiter.go` token bucket，key 为 `rate:<vendor>:<endpoint>`。每个 vendor 单独配置 RPS。
3. **断路器**：使用 `internal/infra/redis/circuit_breaker.go`，连续失败阈值 5，熔断窗口 60s。
4. **指标**：每次调用记录 `apiclient_requests_total{vendor,endpoint,status}` 与 `apiclient_latency_ms`。
5. **重试**：使用 `pkg/retry`（若不存在则在 `internal/infra/apiclient/retry.go` 新建），仅对 5xx / 网络错误 / 429 重试，最多 3 次，指数退避 1s/2s/4s。
6. **指标尾巴**：成功更新 `LastSuccessAt`；失败写入 `LastError`。

### 1.3 实现位置

**文件**：`backend/internal/infra/apiclient/middleware.go`

提供 `func New(vendor string, opts Options) Source`，所有 vendor 子包通过其构造：

```go
type Options struct {
    BaseURL       string
    Timeout       time.Duration
    RPS           float64        // 每秒请求上限
    Burst         int            // 突发桶大小
    KeyResolver   func(ctx context.Context) (string, error) // 密钥获取
    HTTPClient    *http.Client   // 默认 http.DefaultClient
    UserAgent     string
}
```

---

## 2. Vendor 子包目录骨架

每个 vendor 子包**必须**遵循以下骨架：

```
backend/internal/infra/apiclient/<vendor>/
├── client.go         # 构造函数 New()，返回 *Client，内部持有 apiclient.Source
├── endpoints.go      # 各端点常量（路径、查询参数键）
├── types.go          # 请求/响应字段结构体（与第三方 API 字段命名一致）
├── parse.go          # 字段解析、单位换算（不含业务派生）
├── errors.go         # 错误码到 internal/errs 的映射
└── client_test.go    # 单元测试，使用 httptest.Server mock
```

### 2.1 client.go 模板

```go
package <vendor>

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"

    "github.com/antclaw/antclaw/internal/infra/apiclient"
)

type Client struct {
    src apiclient.Source
}

func New(src apiclient.Source) *Client {
    return &Client{src: src}
}

// 端点方法举例
func (c *Client) GetXYZ(ctx context.Context, params XYZParams) (*XYZResponse, error) {
    u, _ := url.Parse(endpointXYZ)
    q := u.Query()
    q.Set("param", params.Value)
    u.RawQuery = q.Encode()

    req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
    if err != nil {
        return nil, fmt.Errorf("<vendor> build req: %w", err)
    }
    resp, err := c.src.Do(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("<vendor> do: %w", err)
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return nil, parseError(resp)
    }
    var out XYZResponse
    if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
        return nil, fmt.Errorf("<vendor> decode: %w", err)
    }
    return &out, nil
}
```

### 2.2 不得在 vendor 包内做的事

- 派生计算（IV Surface、GEX、TVL 排名、相关性矩阵...）
- 调用其他业务服务
- 直接读写 Postgres
- 调用 Redis（除限流/断路器，且必须经由 Source 中间件）
- 业务级日志（应在调用方业务服务里加上下文 log）

---

## 3. 18 家 Vendor 落位详情

> 每家 vendor 的优先级、典型端点、限流参数、密钥来源以下方式给出。
> 「能力归属」列表示派生指标住哪个 service 包。

### 3.1 已存在 / 已迁移

| Vendor | 路径 | 端点 | RPS 上限 | 密钥 | 能力归属 | 状态 |
|---|---|---|---|---|---|---|
| MQL5 Calendar | `apiclient/mql5/` | `/api/calendar/widget` | 1 | 无 | `service/calendar` | ✅ |
| FRED | `apiclient/fred/` | `/series/observations` | 2 | `data_sources.fred.api_key` | `service/macro` | ✅ |

### 3.2 需要抽出/重构（M1 范围）

| Vendor | 当前位置 | 目标位置 | 端点 | RPS | 密钥 | 能力归属 |
|---|---|---|---|---|---|---|
| CoinGecko | `worker/collector_onchain.go` 内联 HTTP | `apiclient/coingecko/` | `/coins/{id}/market_chart` 等 | 0.5（free 50/min） | 无（公开） | `service/onchain`、`service/price` |
| CFTC Socrata | `worker/collector_cot.go` 内联 HTTP | `apiclient/cftc/` | `/resource/jun7-fc8e.json` | 1 | 可选 token | `service/cot` |
| TwelveData | `service/price` 内联 | `apiclient/twelvedata/` | `/time_series` | 0.5（free 8/min） | api_key | `service/price` |
| AlphaVantage | `service/price` 内联 | `apiclient/alphavantage/` | `/query` | 0.083（free 5/min） | api_key | `service/price` |
| Yahoo | `service/price` 内联 | `apiclient/yahoo/` | `/v8/finance/chart/{symbol}` | 1 | 无 | `service/price` |

### 3.3 新建薄 client（M2~M4 范围）

#### M2 期权 / 加密衍生品

| Vendor | 端点（关键） | RPS | 密钥 | 能力归属 |
|---|---|---|---|---|
| **Bybit** | `/v5/market/kline`、`/v5/market/orderbook`、`/v5/market/funding/history` | 5 | 可选 api_key | `service/microstructure`、`service/orderflow` |
| **Deribit** | `/api/v2/public/get_book_summary_by_currency`、`/public/get_volatility_index_data`、`/public/get_instruments` | 10 | 无（公共端点） | `service/options`、`service/vol` |

#### M3 宏观

| Vendor | 端点 | RPS | 密钥 | 能力归属 |
|---|---|---|---|---|
| **CME FedWatch** | `https://www.cmegroup.com/services/...`（HTML / JSON） | 0.2 | 无 | `service/fedwatch` |
| **DTCC** | `https://www.dtcc.com/...` 公开下载 | 0.1 | 无 | `service/macro_extras` |
| **ECB** | `https://sdw-wsrest.ecb.europa.eu/service/data/{flow}/{key}` | 1 | 无 | `service/macro_extras` |
| **Eurostat** | `https://ec.europa.eu/eurostat/api/dissemination/statistics/1.0/data/{dataset}` | 1 | 无 | `service/macro_extras` |
| **OECD** | `https://stats.oecd.org/sdmx-json/data/{dataset}` | 0.5 | 无 | `service/macro_extras` |
| **SNB** | `https://data.snb.ch/api/cube/{cube}/data/csv` | 0.5 | 无 | `service/macro_extras` |
| **TradingEconomics** | `https://api.tradingeconomics.com/...` | 0.2 | api_key | `service/macro_extras` |
| **US Treasury** | `https://api.fiscaldata.treasury.gov/services/api/fiscal_service/{ds}` | 1 | 无 | `service/treasury` |
| **BIS** | `https://stats.bis.org/api/v2/...` | 0.2 | 无 | `service/bis` |
| **IMF** | `http://dataservices.imf.org/REST/SDMX_JSON.svc/{dataset}` | 0.2 | 无 | `service/imf` |
| **World Bank** | `https://api.worldbank.org/v2/country/{c}/indicator/{i}` | 1 | 无 | `service/worldbank` |

#### M4 链上 / DeFi / SEC / 情绪扩展

| Vendor | 端点 | RPS | 密钥 | 能力归属 |
|---|---|---|---|---|
| **CryptoCompare** | `/data/v2/histoday`、`/data/social/coin/histo/day` | 1 | api_key | `service/onchain`、`service/sentiment` |
| **Defillama** | `https://api.llama.fi/protocols`、`/tvl/{protocol}` | 1 | 无 | `service/defi` |
| **Coinmetrics** | `https://community-api.coinmetrics.io/v4/...` | 0.5 | 无 | `service/onchain` |
| **Blockchain.com** | `https://blockchain.info/charts/{metric}?format=json` | 0.5 | 无 | `service/onchain` |
| **SEC EDGAR** | `https://data.sec.gov/submissions/CIK{cik}.json`、`https://www.sec.gov/cgi-bin/browse-edgar` | 0.1（必须 User-Agent 标识） | 无 | `service/sec` |
| **Finviz** | `https://finviz.com/screener.ashx?...&ft=4` 等（HTML） | 0.2 | 无 | `service/sentiment` |
| **MyFXBook** | `https://www.myfxbook.com/community/outlook` | 0.2 | 无 | `service/sentiment` |
| **CBOE** | `https://www.cboe.com/us/options/market_statistics/...` | 0.2 | 无 | `service/sentiment` |
| **OpenInsider** | `http://openinsider.com/screener?...` | 0.1 | 无 | `service/sentiment` |

---

## 4. 数据源种子（datasource seed）

每新增一个 vendor 必须在 `backend/internal/adapter/storage/postgres/ensure_schema_seeds.go` 增加一条种子，包含：

```go
{
    ID:          "deribit",
    DisplayName: "Deribit",
    Category:    "options",
    Endpoint:    "https://www.deribit.com/api/v2",
    RequiresKey: false,
    Enabled:     true,
    Priority:    1,
    TimeoutMs:   15000,
    RPS:         10.0,
    Burst:       20,
}
```

> `Priority` 决定降级链顺序（同 `Category` 内值越小越先用）。

---

## 5. 密钥与配置热更新

- 密钥不写在 `.env`，全部存于 `data_sources.api_key`（加密）。
- `apiclient.New(vendor, Options{KeyResolver: resolver.GetSecret(vendor)})` 在每次请求时即时取最新密钥。
- 管理端在 `/datasources` 修改密钥后，`internal/service/datasource/service.go` 触发 Redis Pub/Sub `datasource:updated`，所有 worker 与 api 实例订阅并刷新内存中的 KeyResolver 缓存。

---

## 6. 错误码映射

**文件**：每个 vendor 子包的 `errors.go`

```go
package <vendor>

import (
    "fmt"
    "net/http"

    "github.com/antclaw/antclaw/internal/errs"
)

func parseError(resp *http.Response) error {
    switch resp.StatusCode {
    case http.StatusUnauthorized, http.StatusForbidden:
        return errs.Wrap(errs.AuthError, "<vendor> auth failed")
    case http.StatusTooManyRequests:
        return errs.Wrap(errs.RateLimited, "<vendor> rate limited")
    case http.StatusNotFound:
        return errs.Wrap(errs.NotFound, "<vendor> resource not found")
    default:
        return errs.Wrap(errs.External, fmt.Sprintf("<vendor> http %d", resp.StatusCode))
    }
}
```

业务服务层据此区分"是否可重试"、"是否提示用户"、"是否降级"。

---

## 7. 测试规范

每个 vendor 子包必须包含：

- **协议测试**：`httptest.NewServer` 拦截，断言请求 URL/Header/Body 符合预期
- **解析测试**：用真实 ARK 仓库内的 fixture（如 `Emulator/ark-intelligent/...test_data/...`）做断言
- **错误测试**：模拟 401/429/500 验证 `errors.go` 映射

覆盖率目标：≥ 70%（vendor 包属适配器层）。

---

## 8. 拆包阈值

- 单 `client.go` > 400 行 → 按端点拆为 `client_market.go`、`client_options.go` 等
- 端点常量、字段类型、解析函数务必分别住 `endpoints.go`、`types.go`、`parse.go`
- 永远不要让任何文件 > 800 行

---

## 9. 完成判据（每个 vendor）

- [ ] 子包目录骨架存在并编译通过
- [ ] 实现 `apiclient.Source` 中间件链
- [ ] `data_sources` 种子已加入
- [ ] 至少一个端到端业务服务消费此 client
- [ ] 单元测试覆盖率 ≥ 70%
- [ ] `/datasources` 页面可见此 vendor，密钥可热配
