# AntClaw 跨市场与微观结构方案

> **版本**：v1.0  
> **对应 ark-intelligent 模块**：`intermarket/`、`microstructure/`、`orderflow/`  
> **对应 AntClaw Proto**：`StrategyService` / `SentimentService`

---

## 一、ark-intelligent 方法分析

### 1.1 Intermarket 引擎（`intermarket/engine.go`, 11.7KB）

**职责**：跟踪资产间的相关性 / 比率 / 领先滞后关系，识别 risk-on/off 的多资产共振或分歧。

**核心逻辑**：
- 经典 intermarket pairs：
  - DXY ↔ EUR/USD（反向）
  - DXY ↔ Gold（反向）
  - SPX ↔ VIX（反向）
  - SPX ↔ Treasury 10Y（risk on/off）
  - Oil ↔ CAD（正向）
  - AUD ↔ Gold / Copper（正向）
  - Yen ↔ Risk（避险时 JPY 强）
- 计算 30 / 90 日滚动相关
- 识别"破坏关系"事件（rolling corr 偏离历史 mean > 2σ）
- 输出 `IntermarketSignal`：哪对正常，哪对异常，给出方向暗示

### 1.2 Microstructure 引擎（`microstructure/engine.go`, 11.9KB）

**职责**：从盘口和高频价格中提取微观信号（针对 crypto，使用 Bybit orderbook）。

**核心指标**：
- **Order Book Imbalance (OBI)**：top N levels 买/卖量比
- **Spread**：bid-ask spread + 历史百分位
- **Depth**：累计买卖深度
- **Stress Score**：极薄深度 + 宽 spread → 流动性风险
- **Slippage Estimator**：估算各档位的市价单滑点

输出建议：
- 加大仓位需谨慎（流动性差）
- 异常 OBI → 短期方向预判

### 1.3 Orderflow 引擎（`orderflow/engine.go`, 5.9KB）

**子模块**：
- **delta.go**：tick-rule 估算的买卖压差，累计 delta
- **absorption.go**：在某价位反复买卖但价格不动 → 大单吸筹/派发
- **poc.go**：Volume Profile 的 Point of Control（最大成交量价位）

**输出**：
- 当前 session 的 cum delta 趋势（持续 + → 多头压力）
- Absorption events 列表（位置 + 强度）
- POC 价位 + Value Area High/Low

---

## 二、AntClaw 设计方案

### 2.1 架构

```
service/intermarket/
  ├── engine.go        // 主引擎
  ├── pairs.go         // 经典对配置
  └── correlation.go   // 滚动相关 + 异常检测

service/microstructure/
  ├── engine.go        // 基于 orderbook 的指标
  ├── obi.go           // Order Book Imbalance
  ├── depth.go         // Depth + slippage
  └── stress.go        // Stress score

service/orderflow/
  ├── engine.go
  ├── delta.go         // tick-rule delta
  ├── absorption.go    // 吸筹检测
  └── volume_profile.go // POC + VA
```

### 2.2 核心接口

```go
type IntermarketEngine interface {
    Compute(ctx context.Context) (*IntermarketReport, error)
    Track(ctx, pair PairKey) ([]CorrSnapshot, error)
}

type MicrostructureEngine interface {
    Snapshot(ctx, symbol string) (*MicroResult, error)   // 单次快照
    Subscribe(ctx, symbol string) (<-chan MicroResult, error)  // 实时流
}

type OrderflowEngine interface {
    SessionDelta(ctx, symbol string, session SessionWindow) (*DeltaResult, error)
    Absorptions(ctx, symbol string, lookback time.Duration) ([]AbsorptionEvent, error)
    VolumeProfile(ctx, symbol string, period time.Duration) (*VolumeProfile, error)
}
```

### 2.3 Schema

```sql
-- Intermarket 相关性历史
CREATE TABLE intermarket_correlations (
    time         TIMESTAMPTZ NOT NULL,
    pair_a       VARCHAR(32),
    pair_b       VARCHAR(32),
    window_days  INT,
    correlation  DOUBLE PRECISION,
    historical_mean DOUBLE PRECISION,
    historical_std  DOUBLE PRECISION,
    z_score      DOUBLE PRECISION,
    is_break     BOOLEAN,
    PRIMARY KEY (time, pair_a, pair_b, window_days)
);
SELECT create_hypertable('intermarket_correlations', 'time', chunk_time_interval => INTERVAL '90 days');

-- Microstructure 快照（高频，注意压缩）
CREATE TABLE micro_snapshots (
    time         TIMESTAMPTZ NOT NULL,
    symbol       VARCHAR(32) NOT NULL,
    obi_top10    DOUBLE PRECISION,
    spread_bps   DOUBLE PRECISION,
    bid_depth    DOUBLE PRECISION,
    ask_depth    DOUBLE PRECISION,
    stress_score DOUBLE PRECISION,
    PRIMARY KEY (time, symbol)
);
SELECT create_hypertable('micro_snapshots', 'time', chunk_time_interval => INTERVAL '7 days');
SELECT add_compression_policy('micro_snapshots', INTERVAL '7 days');

-- Orderflow 事件
CREATE TABLE orderflow_absorptions (
    id          BIGSERIAL PRIMARY KEY,
    time        TIMESTAMPTZ NOT NULL,
    symbol      VARCHAR(32),
    price       DOUBLE PRECISION,
    direction   VARCHAR(8),    -- 'BUY','SELL'
    strength    DOUBLE PRECISION,
    volume      DOUBLE PRECISION
);
SELECT create_hypertable('orderflow_absorptions', 'time', chunk_time_interval => INTERVAL '30 days');

CREATE TABLE volume_profiles (
    time          TIMESTAMPTZ,
    symbol        VARCHAR(32),
    period        VARCHAR(16),     -- 'session','day','week'
    poc           DOUBLE PRECISION,
    vah           DOUBLE PRECISION,
    val           DOUBLE PRECISION,
    profile       JSONB,           -- 价格-成交量分布
    PRIMARY KEY (time, symbol, period)
);
```

### 2.4 Redis
| Key | TTL |
|-----|-----|
| `cache:intermarket:report` | 30m |
| `cache:micro:{symbol}` | 30s（高频）|
| `cache:orderflow:profile:{symbol}:{period}` | 5m |

### 2.5 调度
- **Intermarket**：每 30 分钟重算所有 pair correlation
- **Microstructure**：实时（WebSocket Bybit + Deribit），写入快照每 5 秒一条
- **Orderflow**：跟随交易 tick；session 收盘时整理 delta + VP

### 2.6 优化与提升

| 维度 | ark-intelligent | AntClaw |
|------|----------------|---------|
| Intermarket 历史 | 无 | hypertable + 异常事件表 |
| 微观数据流 | 非持久 | 持久化 + 压缩 + Continuous aggregates |
| 订单流 absorption | 单次扫描 | 长期事件库，可统计胜率 |
| Volume Profile | 实时算 | 缓存 + 跨 session 拼接 |
| WebSocket | 单连接 | 支持多 worker 分片订阅 |

---

## 三、参考文件

- ark-intelligent：`internal/service/intermarket/*`, `microstructure/*`, `orderflow/*`
- AntClaw proto：`proto/antclaw/v1/strategy.proto`, `sentiment.proto`
- AntClaw service：待新建 `backend/internal/service/{intermarket,microstructure,orderflow}/`
