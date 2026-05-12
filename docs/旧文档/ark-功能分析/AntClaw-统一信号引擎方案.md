# AntClaw 统一信号引擎方案

> **版本**：v1.0  
> **对应 ark-intelligent 模块**：`internal/service/analysis/` (3 files, ~17KB)  
> **对应 AntClaw Proto**：`StrategyService` / 新增 `SignalService`

---

## 一、ark-intelligent 方法分析

### 1.1 核心职责
`analysis/unified_signal_engine.go` 是 ark-intelligent 的**最高层信号融合器**：把五大子系统（COT / CTA / Quant / Sentiment / Seasonal）的独立信号融合为单一方向性建议。

### 1.2 加权方案

```
defaultWeights = {
    COT:       0.30,   // CFTC 持仓
    CTA:       0.30,   // ta 模块的多周期 confluence
    Quant:     0.20,   // price.regime / HMM / GARCH
    Sentiment: 0.15,   // sentiment 模块
    Seasonal:  0.05    // 季节性
}
// 权重总和必须 = 1.0
```

### 1.3 计算流程

```
For each currency/contract:
  1. 从各子系统拉取 raw_score（-100 ~ +100）
  2. 归一化到 [-1, +1]：normalized = raw / 100
  3. 投票方向：normalized > +0.2 → LONG; < -0.2 → SHORT; else NEUTRAL
  4. 加权求和：unified = Σ(weight_i × normalized_i)
  5. 推荐分类：
     unified > +0.6  → STRONG_LONG
     unified > +0.2  → LONG
     unified ∈ [-0.2, +0.2] → NEUTRAL
     unified < -0.2  → SHORT
     unified < -0.6  → STRONG_SHORT
  6. 置信度：投票一致比例 + 信号强度
```

### 1.4 ComponentScore 结构

每个子系统贡献：
```go
ComponentScore {
    Name:           "COT" / "CTA" / ...
    RawScore:       -100..+100
    NormalizedScore: -1..+1
    Vote:           LONG/NEUTRAL/SHORT
    Weight:         本次实际权重（降级时可能 < default）
    Available:      子系统是否可用
}
```

### 1.5 Graceful Degradation

任一子系统不可用：
- `Available = false`
- 该系统权重置 0
- **重新归一化**剩余权重，使其和 = 1.0
- 标注 `MissingSubsystems`

例：Sentiment 不可用 → COT/CTA/Quant/Seasonal 权重按比例放大

### 1.6 输出
```
UnifiedSignalResult {
    Currency, Recommendation, Score, Confidence,
    Components []ComponentScore,
    BullishVotes, BearishVotes, NeutralVotes,
    MissingSubsystems []string,
    ComputedAt time.Time
}
```

---

## 二、AntClaw 设计方案

### 2.1 架构

**新增 `SignalService`（Proto）**，职责单一：信号融合。

```
SignalService (Proto, 新增)
  └── service/signal/
      ├── engine.go             // UnifiedEngine 主逻辑
      ├── weights.go            // 权重配置 + 动态调整
      ├── voting.go             // 投票算法
      └── adapter/
          ├── cot_adapter.go    // 调 COTService
          ├── ta_adapter.go     // 调 TAService
          ├── regime_adapter.go // 调 RegimeService
          ├── sentiment_adapter.go
          └── seasonal_adapter.go
```

**架构亮点**：通过 adapter 调用子服务，**信号引擎本身无业务逻辑**，只做加权融合。

### 2.2 核心接口

```go
type UnifiedEngine interface {
    Compute(ctx, symbol string, opts UnifyOpts) (*UnifiedSignal, error)
    ComputeAll(ctx context.Context) ([]UnifiedSignal, error)
    GetWeights(ctx) (*WeightConfig, error)
    UpdateWeights(ctx, w WeightConfig) error  // admin 可调
}

type SignalSource interface {
    Name() string
    Score(ctx, symbol string) (*ComponentScore, error)  // 返回 -100~+100 + Available
}

type WeightConfig struct {
    COT, CTA, Quant, Sentiment, Seasonal float64
    UpdatedBy   string
    UpdatedAt   time.Time
}
```

### 2.3 Schema

```sql
-- 信号历史（用于回测和准确率统计）
CREATE TABLE unified_signals (
    id              BIGSERIAL PRIMARY KEY,
    symbol          VARCHAR(32) NOT NULL,
    issued_at       TIMESTAMPTZ NOT NULL,
    recommendation  VARCHAR(16),     -- STRONG_LONG/LONG/NEUTRAL/SHORT/STRONG_SHORT
    unified_score   DOUBLE PRECISION,    -- -1..+1
    confidence      DOUBLE PRECISION,
    components      JSONB,
    missing_subsys  TEXT[],
    weights_used    JSONB
);
SELECT create_hypertable('unified_signals', 'issued_at', chunk_time_interval => INTERVAL '90 days');
CREATE INDEX idx_signals_symbol_time ON unified_signals (symbol, issued_at DESC);

-- 权重配置（可热更新）
CREATE TABLE signal_weight_config (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(64) UNIQUE,    -- 'default','aggressive','conservative'
    weights     JSONB,
    is_active   BOOLEAN,
    updated_at  TIMESTAMPTZ
);

-- 信号准确率统计（回测）
CREATE TABLE signal_outcomes (
    signal_id       BIGINT REFERENCES unified_signals(id),
    horizon         VARCHAR(8),       -- '1D','1W','2W','1M'
    return_pct      DOUBLE PRECISION,
    direction_match BOOLEAN,
    evaluated_at    TIMESTAMPTZ
);
```

### 2.4 Redis

| Key | TTL |
|-----|-----|
| `cache:signal:{symbol}` | 5m |
| `cache:signal:weights` | 永久（变更时显式失效）|
| `pubsub:signal:updated` | - |

### 2.5 调度
- **触发刷新**：任一子系统数据更新（COT 发布、TA 重算等）→ pubsub 通知 → 重算受影响 symbol
- **定期回评**：每日扫描 1D/1W/2W 前的信号，计算 outcome → 写入 `signal_outcomes`
- **权重自适应**（高级）：基于过去 90 天 outcome，自动微调子系统权重（贝叶斯优化）

### 2.6 优化与提升

| 维度 | ark-intelligent | AntClaw |
|------|----------------|---------|
| 子系统耦合 | import 多个 service 包 | 通过 adapter 接口隔离 |
| 权重 | 硬编码 default | DB 配置 + 多套预设 + 热更新 |
| 历史 | 无 | 完整信号 + outcome 表，可回测 |
| 实时性 | 用户请求触发 | pubsub 触发，前端 SSE 推 |
| 准确率反馈 | 无 | outcome 评估 + 权重自适应 |

---

## 三、参考文件

- ark-intelligent：`internal/service/analysis/unified_signal_engine.go`, `analysis/types.go`
- AntClaw proto：需新增 `proto/antclaw/v1/signal.proto`
- AntClaw service：待新建 `backend/internal/service/signal/`
