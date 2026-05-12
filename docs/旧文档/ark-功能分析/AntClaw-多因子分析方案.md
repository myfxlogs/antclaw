# AntClaw 多因子分析方案

> **版本**：v1.0  
> **对应 ark-intelligent 模块**：`internal/service/factors/` (12 files)  
> **对应 AntClaw Proto**：`StrategyService`（factors 子域）/ 新增 `FactorService`

---

## 一、ark-intelligent 方法分析

### 1.1 核心职责
`factors/` 是**横截面因子排名引擎**：给每个资产打分（-1 ~ +1），跨资产 z-score 归一化，组合排名输出最强 / 最弱标的。

### 1.2 六大因子

| 因子 | 文件 | 算法 |
|------|------|------|
| **Momentum** | `momentum.go` | 多窗口收益率（1M/3M/6M/12M）跳过最近 1M（避免短期反转）|
| **Trend Quality** | `trend_quality.go` | ADX proxy + MA 对齐（20/50/200）+ 线性回归 R² + 连续涨跌天数 |
| **Carry-Adjusted Momentum** | `carry_adjusted.go` | momentum + α × normalized(carry)；FX 用利差，crypto 用资金费率 |
| **Low Vol** | `low_vol.go` | annualized return / annualized vol（夏普代理），低波动 + 正收益奖励 |
| **Residual Reversal** | `residual_reversal.go` | OLS 回归 vs 市场 → 残差大 = 异动，预期均值回归 |
| **Crowding Risk** | `crowding.go` | COTIndex + CrowdingIndex + spec momentum，过度拥挤 = 风险惩罚 |

### 1.3 Flow Divergence 检测（flow_divergence.go）

**TASK-162**：检测正常相关资产对解耦
- 滚动 20 bar Pearson 相关
- 60 bar 基线（mean + stddev）
- 偏离 z-score：`(curr_corr - base_mean) / base_std`
- Lead-lag 分析：在 ±1..±5 bar 偏移做交叉相关，找出谁领先

### 1.4 Profile Builder（profile_builder.go）
- 从 `DailyPriceStore` + `COTRepository` 组装 `AssetProfile`
- 包含：DailyCloses、COTIndex、SmartMoneyNet、CrowdingIndex、SpecMomentum4W

### 1.5 Engine（engine.go）

```
For each asset:
  raw_score = Σ(factor_weight × factor_raw_score)
Cross-sectional z-score normalize:
  norm_i = (raw_i - mean) / stddev
Rank by norm_i:
  Top N → 候选做多
  Bottom N → 候选做空
```

权重默认：momentum 0.25, trend_quality 0.20, carry 0.15, low_vol 0.15, reversal 0.10, crowding -0.15（惩罚）

### 1.6 输出
```go
RankingResult {
    Ranked []AssetRank {
        ContractCode, Currency, RawScore, NormScore, Rank, FactorBreakdown
    }
    Top, Bottom []AssetRank
    ComputedAt
}
```

---

## 二、AntClaw 设计方案

### 2.1 架构

```
service/factors/
  ├── engine.go              // 主引擎 + 权重 + 排名
  ├── profile_builder.go     // 组装 AssetProfile
  ├── flow_divergence.go     // 解耦检测
  └── score/                  // 各因子纯函数（拆分子包）
      ├── momentum.go
      ├── trend_quality.go
      ├── carry.go
      ├── low_vol.go
      ├── residual.go
      └── crowding.go
```

### 2.2 核心接口

```go
type FactorEngine interface {
    Rank(ctx context.Context) (*RankingResult, error)
    Detail(ctx, symbol string) (*AssetFactorBreakdown, error)
    UpdateWeights(ctx, w FactorWeights) error
}

type FactorScorer interface {
    Name() string
    Score(profile AssetProfile) float64       // raw score
    Weight() float64
}

type FlowDivergenceEngine interface {
    Compute(ctx, pairs [][2]string) (*FlowDivergenceResult, error)
}
```

### 2.3 Schema

```sql
-- 因子排名快照
CREATE TABLE factor_rankings (
    time           TIMESTAMPTZ NOT NULL,
    snapshot_id    BIGSERIAL,
    weights        JSONB,
    PRIMARY KEY (time, snapshot_id)
);

CREATE TABLE factor_ranking_entries (
    snapshot_id    BIGINT,
    symbol         VARCHAR(32),
    rank           INT,
    raw_score      DOUBLE PRECISION,
    norm_score     DOUBLE PRECISION,
    breakdown      JSONB,
    PRIMARY KEY (snapshot_id, symbol)
);

-- Flow divergence 历史
CREATE TABLE flow_divergence_history (
    time         TIMESTAMPTZ NOT NULL,
    pair_a       VARCHAR(32),
    pair_b       VARCHAR(32),
    corr         DOUBLE PRECISION,
    baseline_mean DOUBLE PRECISION,
    baseline_std  DOUBLE PRECISION,
    z_score      DOUBLE PRECISION,
    lead_lag     INT,
    PRIMARY KEY (time, pair_a, pair_b)
);
SELECT create_hypertable('flow_divergence_history', 'time', chunk_time_interval => INTERVAL '30 days');
```

### 2.4 Redis
| Key | TTL |
|-----|-----|
| `cache:factor:ranking:latest` | 1h |
| `cache:factor:flow_div:latest` | 30m |

### 2.5 调度
- **每日 23:00 UTC** 触发全市场因子排名（日线收盘后）
- **每 4 小时**重算 flow divergence
- 提供 `RecomputeNow` 管理员 API

### 2.6 优化与提升

| 维度 | ark-intelligent | AntClaw |
|------|----------------|---------|
| 因子定义 | 包内私有函数 | 显式 `FactorScorer` 接口，可插件化 |
| 权重 | 硬编码 | DB + 热更新 |
| 历史 | 无 | TimescaleDB 排名快照可回测 |
| Flow divergence | 周期重算 | 持久化时间序列 + 趋势分析 |
| 用户化 | 单一全局 | 支持用户自定义权重 |

---

## 三、参考文件

- ark-intelligent：`internal/service/factors/*.go`
- AntClaw proto：`proto/antclaw/v1/strategy.proto`（factor 子域）
- AntClaw service：待新建 `backend/internal/service/factors/`
