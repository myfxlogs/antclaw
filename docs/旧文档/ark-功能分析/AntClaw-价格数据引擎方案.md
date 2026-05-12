# AntClaw 价格数据引擎方案

> **版本**：v1.0  
> **对应 ark-intelligent 模块**：`internal/service/price/` (43 files)  
> **对应 AntClaw Proto**：`PriceService`, `VolService`（波动率相关下放至波动率服务）

---

## 一、ark-intelligent 方法分析

### 1.1 核心职责
价格模块是全系统**最底层的数据供应者与高级量化分析引擎**，既对接多家外部行情源，又内置 GARCH / HMM / Hurst / Monte Carlo / Wyckoff 等多种量化模型。

**三大子域**：
1. **数据采集层**：多源切换（TwelveData → AlphaVantage → Yahoo → Stooq → CoinGecko）
2. **上下文构建层**：把原始 OHLCV 转换为可直接喂给策略的 `PriceContext`
3. **量化分析层**：波动率、相关性、区间、季节性、微观结构

### 1.2 多源降级采集（fetcher.go）

**API Key 池化 + 轮询**：
- `twelveDataKeys []string` + 原子计数器 `tdKeyIndex` → 每次请求选一个 key，避免单 key 限流
- 每个源有独立 `circuitbreaker.Breaker`（阈值 3 次失败 / 5 分钟冷却）
- 失败链路：主源失败 → 切备源 → 全部失败才返回错误

**日线 vs 盘中**：
- 日线（`daily_fetcher.go`）：Yahoo 优先（无 key 稳定）→ TwelveData 备
- 盘中 4H（`intraday_fetcher.go`）：TwelveData 优先（原生 4h 间隔）→ Yahoo（1h 聚合模拟）

### 1.3 上下文构建（context.go / daily_context.go / intraday_context.go）

把原始 OHLCV 打包成 **`PriceContext`**：
- 趋势标签（UP/DOWN/FLAT，基于斜率 + 波动率调整）
- ATR、Normalized ATR（波动率占价格百分比）
- 52 周最高 / 最低、距高低点的百分比
- RSI(14)、均线关系
- 最近 4 周 return

### 1.4 量化模型矩阵

| 文件 | 模型 | 核心算法 | 输出 |
|------|------|---------|------|
| `garch.go` | GARCH(1,1) | 极大似然估计 ω/α/β 参数 | 当前 σ² + 预测 N 步波动 |
| `hmm_regime.go` | HMM (3 状态) | Baum-Welch 训练 + Viterbi 解码 | P(BULL/BEAR/CRISIS) + 转移预警 |
| `hurst.go` | R/S 分析 | Hurst 指数 0-1 | 趋势性（>0.5 持续，<0.5 均值回归）|
| `volatility.go` | 已实现波动率 | 年化 stddev(log return) | 20/60/120 日 RV |
| `vol_cone.go` | 波动率锥 | 历史分布 P5/P25/P50/P75/P95 | 当前 vol 相对分位 + 异常标志 |
| `correlation.go` | 滚动 Pearson | 配对相关矩阵 + 动态区间 | 相关性热图 + 变化点 |
| `regime_correlation.go` | 状态相依相关 | 按 FRED regime 分组统计 | 每个宏观状态下的相关性 |
| `levels.go` | 关键价位检测 | 枢轴点 / SR 聚类 / 斐波 | 最近支撑压力 + 距离 |
| `seasonal.go` | 季节性 | ISO 周均 return + Z-score | 周度季节性偏离 |
| `wyckoff.go` | Wyckoff 阶段 | 事件检测（SC/ST/SOS/SPRING）| ACCUMULATION/MARKUP/DISTRIBUTION/MARKDOWN |
| `montecarlo_scenario.go` | MC 路径 | GARCH 波动 + HMM 漂移 GBM | 价格分布百分位 + VaR |
| `divergence.go` | 价格-COT 背离 | 4 周价格趋势 vs COT 方向 | 背离标签 + 严重度 |
| `hmm_regime` `regime_alert.go` | 状态转换预警 | HMM 转换概率 + 冷却 | AMBER / RED 三级预警 |
| `position_size.go` | 仓位计算 | ATR 止损 + 风险比例 | 建议手数 + 风险金额 |

### 1.5 触发机制
- **批量拉取**：`FetchAllDaily` / `FetchAllIntraday` 一次请求所有合约（8-15 个品种）
- **上下文重算**：每次 COT 发布（周五）触发 `DailyContext`；盘中每 30 分钟重算 `IntradayContext`
- **量化模型**：按需计算（用户请求 `/regime` 才算 HMM），结果缓存 5-15 分钟

### 1.6 降级策略
- API 失败 → 切换源 → 全部失败 → 返回缓存数据 + 标记 `stale`
- GARCH 训练失败 → 回退到简单 rolling stddev
- HMM 未收敛 → 回退到基于波动率的规则分类

---

## 二、AntClaw 设计方案

### 2.1 架构调整

**Hexagonal 分层**：
```
PriceService (Proto) 
  → internal/service/price (business logic)
      ├── Fetcher (interface) → adapter/marketdata/{twelvedata, alphavantage, yahoo, stooq, coingecko}
      ├── Analyzer (pure functions): GARCH, HMM, Hurst, VolCone, ...
      └── Repository (interface) → adapter/postgres/price_repo
```

**vs ark-intelligent 核心差异**：
1. 数据层与分析层严格分离（Hexagonal），分析纯函数，易测试
2. 日 K / 4H K 线持久化到 TimescaleDB hypertable，重启后立即可用
3. 量化模型结果缓存到 Redis（TTL 15 min，key 含数据指纹）
4. 多 API Key 从配置/Redis 读取，支持热更新和动态限流

### 2.2 核心接口

```go
// ports/price.go
type PriceFetcher interface {
    FetchDaily(ctx, symbol, days) ([]DailyBar, error)
    FetchIntraday(ctx, symbol, interval, bars) ([]IntradayBar, error)
    Healthcheck(ctx) error
}

type PriceRepository interface {
    UpsertDailyBars(ctx, bars []DailyBar) error
    UpsertIntradayBars(ctx, bars []IntradayBar) error
    GetDailyBars(ctx, symbol string, from, to time.Time) ([]DailyBar, error)
    GetLatest(ctx, symbol string) (*Bar, error)
}

// service/price/analyzer
type Analyzer struct {
    garch  *GARCHModel
    hmm    *HMMModel
    hurst  *HurstEstimator
    volCone *VolConeCalc
}
func (a *Analyzer) ComputeContext(bars []DailyBar) (*PriceContext, error)
func (a *Analyzer) DetectRegime(bars []DailyBar) (*RegimeResult, error)
func (a *Analyzer) MonteCarloScenario(bars []DailyBar, horizon int) (*MCScenario, error)
```

### 2.3 数据持久化

```sql
-- 日线 hypertable
CREATE TABLE price_daily (
    time      TIMESTAMPTZ NOT NULL,
    symbol    VARCHAR(32) NOT NULL,
    open      DOUBLE PRECISION,
    high      DOUBLE PRECISION,
    low       DOUBLE PRECISION,
    close     DOUBLE PRECISION,
    volume    DOUBLE PRECISION,
    source    VARCHAR(16),       -- 'twelvedata' | 'yahoo' | ...
    PRIMARY KEY (time, symbol)
);
SELECT create_hypertable('price_daily', 'time', chunk_time_interval => INTERVAL '90 days');
ALTER TABLE price_daily SET (timescaledb.compress, timescaledb.compress_segmentby = 'symbol');
SELECT add_compression_policy('price_daily', INTERVAL '90 days');

-- 盘中 4H hypertable
CREATE TABLE price_intraday (
    time      TIMESTAMPTZ NOT NULL,
    symbol    VARCHAR(32) NOT NULL,
    interval  VARCHAR(8)  NOT NULL,   -- '1h', '4h', '15m'
    open DOUBLE PRECISION, high DOUBLE PRECISION, low DOUBLE PRECISION,
    close DOUBLE PRECISION, volume DOUBLE PRECISION,
    PRIMARY KEY (time, symbol, interval)
);
SELECT create_hypertable('price_intraday', 'time', chunk_time_interval => INTERVAL '7 days');

-- 连续聚合：日线 → 周线 / 月线
CREATE MATERIALIZED VIEW price_weekly WITH (timescaledb.continuous) AS
SELECT time_bucket('1 week', time) AS week, symbol,
       first(open, time) AS open, max(high) AS high,
       min(low) AS low, last(close, time) AS close, sum(volume) AS volume
FROM price_daily GROUP BY week, symbol;
```

### 2.4 Redis 缓存

| Key | 类型 | TTL | 内容 |
|-----|------|-----|------|
| `cache:price:ctx:{symbol}` | JSON | 15m | 最新 PriceContext |
| `cache:price:regime:{symbol}` | JSON | 1h | HMM regime 结果 |
| `cache:price:mc:{symbol}:{horizon}` | JSON | 30m | Monte Carlo 场景 |
| `cache:price:levels:{symbol}` | JSON | 1h | 关键价位 |
| `ratelimit:twelvedata:{key_idx}:{min}` | Counter | 60s | 每 key 每分钟计数 |

### 2.5 调度集成

采集调度器新增：
- **日线作业**：每日 01:00 UTC 抓取前一交易日所有合约，增量 UPSERT
- **盘中作业**：每 4 小时抓取 4H bar（对齐蜡烛收盘时间）
- **模型刷新**：COT 发布后（周六）触发重算 HMM、Monte Carlo、季节性

### 2.6 优化与提升

| 维度 | ark-intelligent | AntClaw |
|------|----------------|---------|
| 历史数据 | 每次请求拉全量 | TimescaleDB 持久化，增量拉最新 |
| API Key 管理 | 启动时读 env | 配置中心 + Redis，热更换 |
| 模型结果 | 内存缓存 | Redis 分布式缓存 |
| 限流 | 无 | Redis 滑动窗口（每 key 每分钟）|
| 并发 | 串行 Foreach | Worker pool（4-8 并发）|
| 聚合查询 | 计算时扫全表 | continuous aggregates，O(1) |

---

## 三、参考文件

- ark-intelligent：`internal/service/price/{fetcher,daily_fetcher,intraday_fetcher,context,garch,hmm_regime,hurst,volatility,vol_cone,correlation,levels,seasonal,wyckoff,montecarlo_scenario,divergence,regime_alert,position_size}.go`
- AntClaw proto：`proto/antclaw/v1/price.proto`, `vol.proto`
- AntClaw service：`backend/internal/service/price/service.go`
