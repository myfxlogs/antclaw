# AntClaw 波动率与 GEX 方案

> **版本**：v1.0  
> **对应 ark-intelligent 模块**：`vix/`、`dvol/`、`gex/`  
> **对应 AntClaw Proto**：`VolService`

---

## 一、ark-intelligent 方法分析

### 1.1 VIX 体系（vix/）

**`fetcher.go`**：从 CBOE EOD CSV 拉取
- VX 期货曲线（M1, M2, M3 + Spot）
- VIX9D、VIX、VIX3M、VIX6M（短中长 IV）
- VVIX（vol-of-vol）

**`vol_suite.go`**：扩展指标
- SKEW（尾部风险指数）
- OVX（原油 VIX）、GVZ（黄金 VIX）、RVX（罗素 2000 VIX）
- COR3M（3 月隐含相关性）

**`move.go`**：MOVE 债券波动率（来自 Yahoo `^MOVE`）

**`cross_vol.go`**：跨资产波动率分类
```
CrossVolRegime ∈ {NORMAL, ENERGY_RISK, BROAD_RISK_OFF, SMALL_CAP_STRESS, SYSTEMIC}
```
基于 OVX/VIX, RVX/VIX, MOVE/VIX 比率检测异常资产类风险。

**`skew_vix_alert.go`**：SKEW/VIX 比率历史百分位 → 尾部风险预警

**`types.go`**：派生信号
- Contango / Backwardation（M1>Spot 且 M2>M1 / 反之）
- M1-M3 spread（结构紧张度）

**缓存**：进程内 12 小时（CBOE 每日收盘更新）

### 1.2 DVOL 模块（dvol/）

**加密期权波动率**（Deribit 提供）：
- DVOL Current（年化 IV）
- 24h Change（spike 检测，> 20% → alert）
- IV-HV Spread（implied vs realized 溢价/折价）
- 跨资产对比（DVOL vs CBOE VIX）

**核心：BTC + ETH 各自计算**

### 1.3 GEX 模块（gex/）

**Gamma Exposure 全套分析**（Deribit 期权数据）：

**`calculator.go`**：
- 每个 strike 的 Call GEX、Put GEX、Net GEX
- Total GEX：正 = 做市商做空 gamma → 价格阻尼；负 = 做市商做多 gamma → 价格放大
- **GEX Flip Level**：累积 GEX 变号的价格点（gamma neutral）

**`engine.go`**：
- 关键价位识别：最大 Call wall（阻力）、最大 Put wall（支撑）
- Spot vs Flip Level 关系

**`iv_surface.go`**：
- 拉取 Deribit 全市场期权 instrument
- 构建 IV 曲面（strike × expiry × IV）
- 缓存 30 分钟

**`skew.go`**：
- 5 点 moneyness smile（0.80/0.90/1.00/1.10/1.20）
- Put/Call IV ratio
- Skew slope（线性回归斜率）
- **Skew Flip 检测**：bearish→bullish 或反之
- ATM IV term structure slope

---

## 二、AntClaw 设计方案

### 2.1 架构

```
VolService (Proto)
  ├── service/vol/vix      (CBOE VIX 体系)
  ├── service/vol/dvol     (Deribit 加密)
  ├── service/vol/gex      (期权 Gamma Exposure)
  ├── service/vol/skew     (IV Skew/Smile)
  ├── service/vol/move     (债券波动率)
  └── service/vol/cross    (跨资产波动率分类)
      ↓
  infra/apiclient/{cboe,deribit,yahoo}_client.go
  infra/postgres/vol_repo.go
```

### 2.2 核心接口

```go
type VIXFetcher interface {
    FetchTermStructure(ctx) (*VIXTermStructure, error)
    FetchVolSuite(ctx) (*VolSuite, error)   // SKEW/OVX/GVZ/RVX/...
    FetchMOVE(ctx) (*MOVEData, error)
}

type DVOLFetcher interface {
    FetchDVOL(ctx, currency string) (*CurrencyDVOL, error)
}

type GEXEngine interface {
    Compute(ctx, symbol string) (*GEXResult, error)
    AnalyzeIVSurface(ctx, symbol string) (*IVSurface, error)
    AnalyzeSkew(ctx, symbol string) (*SkewResult, error)
}
```

### 2.3 Schema

```sql
-- VIX 时间序列
CREATE TABLE vix_term_structure (
    time           TIMESTAMPTZ PRIMARY KEY,
    spot           DOUBLE PRECISION,
    m1 DOUBLE PRECISION, m2 DOUBLE PRECISION, m3 DOUBLE PRECISION,
    vvix           DOUBLE PRECISION,
    vix9d          DOUBLE PRECISION, vix3m DOUBLE PRECISION,
    skew DOUBLE PRECISION, ovx DOUBLE PRECISION, gvz DOUBLE PRECISION,
    rvx  DOUBLE PRECISION, move DOUBLE PRECISION,
    contango       BOOLEAN,
    cross_regime   VARCHAR(32),
    raw            JSONB
);
SELECT create_hypertable('vix_term_structure', 'time', chunk_time_interval => INTERVAL '90 days');

-- DVOL
CREATE TABLE dvol_snapshots (
    time           TIMESTAMPTZ NOT NULL,
    currency       VARCHAR(8) NOT NULL,
    current_iv     DOUBLE PRECISION,
    change_24h_pct DOUBLE PRECISION,
    iv_hv_spread   DOUBLE PRECISION,
    iv_hv_ratio    DOUBLE PRECISION,
    spike          BOOLEAN,
    PRIMARY KEY (time, currency)
);
SELECT create_hypertable('dvol_snapshots', 'time', chunk_time_interval => INTERVAL '30 days');

-- GEX 快照
CREATE TABLE gex_snapshots (
    time           TIMESTAMPTZ NOT NULL,
    symbol         VARCHAR(8)  NOT NULL,
    spot_price     DOUBLE PRECISION,
    total_gex      DOUBLE PRECISION,
    flip_level     DOUBLE PRECISION,
    max_call_wall  DOUBLE PRECISION,
    max_put_wall   DOUBLE PRECISION,
    levels         JSONB,
    PRIMARY KEY (time, symbol)
);

-- IV 曲面与 Skew 历史
CREATE TABLE iv_skew_history (
    time         TIMESTAMPTZ NOT NULL,
    symbol       VARCHAR(8)  NOT NULL,
    pc_iv_ratio  DOUBLE PRECISION,
    skew_slope   DOUBLE PRECISION,
    smile        JSONB,             -- 5-point moneyness IV
    term_slope   DOUBLE PRECISION,
    flip_event   VARCHAR(32),        -- 'BEAR_TO_BULL','BULL_TO_BEAR' or NULL
    PRIMARY KEY (time, symbol)
);
```

### 2.4 Redis

| Key | TTL |
|-----|-----|
| `cache:vix:term` | 12h |
| `cache:vix:suite` | 12h |
| `cache:dvol:{currency}` | 1h |
| `cache:gex:{symbol}` | 30m |
| `cache:gex:iv_surface:{symbol}` | 30m |
| `cache:gex:skew:{symbol}` | 30m |

### 2.5 调度

| 任务 | 频率 |
|------|------|
| `vix-eod-fetch` | 每日 22:00 UTC（CBOE 收盘后）|
| `move-fetch` | 每 4 小时 |
| `dvol-btc/eth` | 每小时 |
| `gex-compute` | 每 30 分钟（盘中）|
| `iv-surface` | 每 30 分钟 |
| `skew-analyze` | 每 30 分钟 |

### 2.6 优化与提升

| 维度 | ark-intelligent | AntClaw |
|------|----------------|---------|
| 历史回溯 | 内存 + 即时拉历史 CSV | TimescaleDB 持久化全量 |
| Skew Flip 检测 | 单次比较 | 持久化历史，可回测胜率 |
| GEX 计算 | 每次重算 | 缓存 30m，pubsub 触发刷新 |
| 跨资产 regime | 实时分类 | 历史回溯 + 统计 |
| 报警 | 进程级 | 集成 `AlertsService` |

---

## 三、参考文件

- ark-intelligent：`internal/service/vix/*`, `dvol/*`, `gex/*`
- AntClaw proto：`proto/antclaw/v1/vol.proto`
- AntClaw service：`backend/internal/service/vol/service.go`
