# AntClaw 宏观状态检测方案

> **版本**：v1.0  
> **对应 ark-intelligent 模块**：`regime/`（统一状态引擎）+ 跨多模块（fred/cot/price 都有 regime）  
> **对应 AntClaw Proto**：`MacroService` / `StrategyService`

---

## 一、ark-intelligent 方法分析

### 1.1 三种 regime 视角

ark-intelligent 中"regime（市场/宏观状态）"出现在多处，**含义不同**：

| 模块 | 视角 | 输出 |
|------|------|------|
| `fred/regime.go` | 宏观经济周期 | INFLATIONARY/GOLDILOCKS/STAGFLATION/DEFLATION/STRESS/NEUTRAL |
| `cot/regime.go` | 风险偏好（机构持仓视角）| RISK-ON/RISK-OFF/UNCERTAINTY |
| `price/regime.go` + `hmm_regime.go` | 价格自身状态（HMM）| BULL/BEAR/CRISIS（含转移概率）|
| `regime/overlay_engine.go` | **统一融合** | UnifiedScore（-100~+100）|

### 1.2 OverlayEngine（核心）

**`overlay_engine.go`** 是 ark-intelligent 中**最高级的 regime 引擎**，融合四种子模型：

```
UnifiedScore = HMM(30%) + GARCH(25%) + ADX(25%) + COT(20%)
```

每个子模型贡献 **-100 ~ +100** 的局部分数，加权求和。

#### 1.2.1 HMM 贡献（30%）
- 输入：日线收益率
- 输出：当前 P(BULL/BEAR/CRISIS) + Viterbi 解码状态
- 映射：BULL→+score, BEAR→-score, CRISIS→-score×1.5
- HMMConfidence：最高状态概率

#### 1.2.2 GARCH 贡献（25%）
- 当前 σ² / 长期 σ² → VolRatio
- VolRatio < 0.7 → CONTRACTING（+score，volatility crush 利好趋势）
- VolRatio > 1.3 → EXPANDING（-score，风险上升）
- 否则 → NORMAL（0）

#### 1.2.3 ADX 贡献（25%）
- ADX > 25 → STRONG（趋势）
- ADX 15-25 → MODERATE
- ADX < 15 → WEAK（盘整）
- 与 +DI/-DI 方向结合 → 给定方向分数

#### 1.2.4 COT 贡献（20%）
- 从 `COTRepository` 取最新分析
- COTIndex 偏多 → +score；偏空 → -score

#### 1.2.5 Graceful Degradation
- 任一子模型失败（数据不足、训练失败）→ 该子模型权重置 0，其他权重按比例放大
- **永不返回错误**，至少返回部分模型的结果

### 1.3 缓存策略
- per `symbol:timeframe` 缓存
- TTL：日线 1 小时，4H 30 分钟
- 触发刷新：新蜡烛收盘 / 用户显式 refresh

---

## 二、AntClaw 设计方案

### 2.1 架构

**保留三层视角，但提升为独立子服务**：

```
StrategyService (Proto, 已存在)
  └── service/strategy/regime
        ├── overlay_engine.go    (融合引擎)
        ├── hmm_adapter.go       (从 PriceService 拉 HMM)
        ├── garch_adapter.go     (从 PriceService 拉 GARCH)
        ├── adx_adapter.go       (从 TAService 拉 ADX)
        └── cot_adapter.go       (从 COTService 拉 COT)

MacroService (Proto, 已存在)
  └── service/macro/regime
        └── fred_regime.go       (经济周期分类，已在《FRED 宏观方案》)
```

**设计原则**：
- `overlay_engine` 不直接计算子模型，而是**调用各子服务的 RPC**（或本地接口）
- 子服务对外暴露纯计算接口，可独立测试

### 2.2 核心接口

```go
type RegimeOverlayEngine interface {
    Compute(ctx, symbol, timeframe string) (*RegimeOverlay, error)
    Subscribe(ctx, symbol string) (<-chan RegimeOverlay, error) // SSE 流
}

type SubModelProvider interface {
    HMM(ctx, symbol, tf string) (*HMMState, error)
    GARCH(ctx, symbol, tf string) (*GARCHRegime, error)
    ADX(ctx, symbol, tf string) (*ADXResult, error)
    COT(ctx, contract string) (*COTAnalysis, error)
}

type RegimeOverlay struct {
    Symbol, Timeframe string
    UnifiedScore      float64  // -100..+100
    UnifiedLabel      string   // 'STRONG_BULL','BULL','NEUTRAL','BEAR','STRONG_BEAR'
    HMMScore          float64
    GARCHScore        float64
    ADXScore          float64
    COTScore          float64
    AvailableModels   []string // 实际成功参与的子模型
    ComputedAt        time.Time
}
```

### 2.3 Schema

```sql
CREATE TABLE regime_overlay_history (
    time            TIMESTAMPTZ NOT NULL,
    symbol          VARCHAR(32) NOT NULL,
    timeframe       VARCHAR(8)  NOT NULL,
    unified_score   DOUBLE PRECISION,
    unified_label   VARCHAR(16),
    hmm_state       VARCHAR(16),
    hmm_confidence  DOUBLE PRECISION,
    hmm_score       DOUBLE PRECISION,
    garch_regime    VARCHAR(16),
    vol_ratio       DOUBLE PRECISION,
    garch_score     DOUBLE PRECISION,
    adx_strength    VARCHAR(16),
    adx_value       DOUBLE PRECISION,
    adx_score       DOUBLE PRECISION,
    cot_score       DOUBLE PRECISION,
    available_models JSONB,
    PRIMARY KEY (time, symbol, timeframe)
);
SELECT create_hypertable('regime_overlay_history', 'time', chunk_time_interval => INTERVAL '90 days');

-- 状态转换事件（用于报警和回测）
CREATE TABLE regime_transitions (
    id          BIGSERIAL PRIMARY KEY,
    time        TIMESTAMPTZ NOT NULL,
    symbol      VARCHAR(32),
    timeframe   VARCHAR(8),
    from_label  VARCHAR(16),
    to_label    VARCHAR(16),
    from_score  DOUBLE PRECISION,
    to_score    DOUBLE PRECISION,
    severity    VARCHAR(8)            -- 'INFO','WARN','CRITICAL'
);
```

### 2.4 Redis

| Key | TTL |
|-----|-----|
| `cache:regime:overlay:{symbol}:{tf}` | 30m-1h |
| `pubsub:regime:transition` | - |

### 2.5 调度
- **被动触发**：新蜡烛收盘事件 → 重算受影响 symbol/tf
- **主动刷新**：每 30 分钟扫描热门品种，确保缓存不过期
- **状态切换检测**：对比上次结果，UnifiedLabel 变化 → 写入 `regime_transitions` + 推送

### 2.6 优化与提升

| 维度 | ark-intelligent | AntClaw |
|------|----------------|---------|
| 子模型隔离 | 同进程紧耦合 | 子服务接口隔离，可独立部署 |
| 历史回溯 | 仅最新缓存 | TimescaleDB 全量历史 |
| 状态转换 | 重算时检测 | 独立事件表 + SSE 推送 |
| 降级 | 子模型失败置 0 | 同左 + 显式标注 `AvailableModels` |
| 多品种 | 同步逐个计算 | 异步 worker pool 并行 |

---

## 三、参考文件

- ark-intelligent：`internal/service/regime/overlay_engine.go`, `regime/types.go`, 关联 `fred/regime.go`, `cot/regime.go`, `price/hmm_regime.go`
- AntClaw proto：`proto/antclaw/v1/strategy.proto`, `macro.proto`
- AntClaw service：待新建 `backend/internal/service/strategy/regime/`
