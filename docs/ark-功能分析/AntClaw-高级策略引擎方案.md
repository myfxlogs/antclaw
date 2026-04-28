# AntClaw 高级策略引擎方案（Elliott / Wyckoff / ICT）

> **版本**：v1.0  
> **对应 ark-intelligent 模块**：`elliott/`、`wyckoff/`、`ict/`（独立 service 包）+ `ta/elliott.go,wyckoff.go,ict.go`（ta 包内简版）  
> **对应 AntClaw Proto**：`StrategyService` / `TAService`

---

## 一、ark-intelligent 方法分析

ark-intelligent 把 Elliott / Wyckoff / ICT 三大经典学派**双重实现**：
- `ta/` 包内为简版（基础检测，用于 confluence）
- `elliott/`、`wyckoff/`、`ict/` 独立包为**完整版**（更深 + 更严格规则）

### 1.1 Elliott Wave（elliott/，6 文件）

**`engine.go`**：主引擎
- `MinRetracement`：最小回撤比例（默认 5%），用于过滤噪声 ZigZag
- 入口：`Engine.Analyze(bars []OHLCV) (*WaveCount, error)`

**`zigzag.go`**：ZigZag 摆动检测
- 输入 newest-first，内部反转为 oldest-first
- 严格的 reversal threshold（默认 5%）
- 输出 chronological `[]SwingPoint`

**`validator.go`**：三大铁律
1. Wave 2 不能回撤超过 Wave 1 的 100%
2. Wave 3 不能是最短浪（vs W1, W5）
3. Wave 4 不能进入 Wave 1 的价格区域

任何一条违反 → 取消计数，回到候选状态

**`projector.go`**：斐波那契目标
- Conservative：W5 Target = W1 Start + W1 Length × 1.0
- Aggressive：W5 Target = W1 Start + W1 Length × 1.618
- 下跌镜像

**`types.go`**：类型定义
- WaveType：IMPULSE / CORRECTIVE / DIAGONAL / ZIGZAG / FLAT / TRIANGLE
- Wave：单浪（start/end pivot + length + direction）
- WaveCount：完整计数（5 浪 + targets + confidence）

### 1.2 Wyckoff（wyckoff/，6 文件）

**`engine.go`**：主引擎
- 入口：`Analyze(bars) (*WyckoffReport, error)`

**`events.go`**：11 种事件检测
- **Phase A 事件**：PS（初步支撑）、SC（卖出高潮）、AR（自动反弹）、ST（二次测试）
- **Phase C 事件**：SPRING / UPTHRUST（破位回收）
- **Phase D 事件**：SOS（强势之征）、LPS（最后支撑点）、LPSY（最后供应点）
- **Distribution 镜像**：BC、ARDist、UP、UTAD、SOW
- 每个事件检测算法：基于 K 线模式 + 体积放大 + 摆动幅度

**`phase.go`**：阶段判定
- ACCUMULATION / DISTRIBUTION / TRANSITION
- 看 60 bar 内事件序列匹配度

**`classifier.go`**：分类置信度
- accumScore vs distScore，事件计数加权
- 输出 schematic + confidence 标签 + cause score

**`summary.go`**：人话总结

### 1.3 ICT（ict/，8 文件）

**`engine.go`**：主引擎
- 入口 `Engine.Analyze(bars, symbol, timeframe)`
- FVG / OB **委托给 ta.CalcICT**（避免重复实现）
- 本包专注：BOS / CHoCH / Liquidity Sweeps（ta 包没有）

**`structure.go`**：BOS / CHoCH 检测
- **BOS**（Break of Structure）：上升趋势中收盘突破上一摆动高 = 趋势延续
- **CHoCH**（Change of Character）：上升趋势中收盘跌破最近摆动低 = 反转
- 防泛滥：每次趋势转换只发 1 个 CHoCH

**`liquidity.go`**：流动性扫荡
- SWEEP_HIGH：bar 高 > 上一摆高，但收盘 < 摆高 → 流动性收割
- SWEEP_LOW：镜像
- `Reversed=true` 当下一根 K 线确认反向

**`swing.go`**：swing point 检测
- swingLookback = 5（左右各 5 根确认）
- 内部 chronological 处理

**`fvg.go`**：FVG 包装（委托 ta）
**`orderblock.go`**：OB 包装（委托 ta）

---

## 二、AntClaw 设计方案

### 2.1 架构

**保留双层结构，但严格分包**：

```
service/ta/        (基础版本，参与 confluence 评分)
  ├── ict.go       // 简版 FVG/OB（共用算法）
  ├── elliott.go   // 简版浪型
  └── wyckoff.go   // 简版阶段

service/elliott/   (完整版，独立调用)
  ├── engine.go
  ├── zigzag.go
  ├── validator.go
  ├── projector.go
  └── types.go

service/wyckoff/
  ├── engine.go
  ├── events.go    // 11 种事件
  ├── phase.go
  ├── classifier.go
  └── types.go

service/ict_advanced/   (避免与 service/ta 内 ict 命名冲突)
  ├── engine.go
  ├── structure.go (BOS/CHoCH)
  ├── liquidity.go (Sweeps)
  └── swing.go
```

**单一权威算法（DRY）**：FVG / OB 共享算法，定义在 `ta/internal/ict_core` 包，两边引用。

### 2.2 核心接口

```go
type ElliottEngine interface {
    Analyze(bars []OHLCV) (*WaveCount, error)
    Project(count *WaveCount) (conservative, aggressive float64)
}

type WyckoffEngine interface {
    Analyze(bars []OHLCV) (*WyckoffReport, error)
    DetectEvents(bars []OHLCV) ([]WyckoffEvent, error)
    ClassifyPhase(events []WyckoffEvent) (string, float64)
}

type ICTAdvancedEngine interface {
    Analyze(bars []OHLCV, symbol string, tf string) (*ICTAdvancedResult, error)
    DetectStructure(swings []SwingPoint) ([]StructureEvent, error)
    DetectSweeps(bars []OHLCV, swings []SwingPoint) ([]LiquiditySweep, error)
}
```

### 2.3 Schema

```sql
-- Elliott 浪型历史
CREATE TABLE elliott_counts (
    id          BIGSERIAL PRIMARY KEY,
    symbol      VARCHAR(32),
    timeframe   VARCHAR(8),
    detected_at TIMESTAMPTZ,
    wave_type   VARCHAR(16),
    waves       JSONB,
    valid       BOOLEAN,
    targets     JSONB,
    confidence  DOUBLE PRECISION
);

-- Wyckoff 事件
CREATE TABLE wyckoff_events (
    id          BIGSERIAL PRIMARY KEY,
    symbol      VARCHAR(32),
    timeframe   VARCHAR(8),
    event_name  VARCHAR(16),    -- PS,SC,AR,ST,SPRING,...
    bar_time    TIMESTAMPTZ,
    price       DOUBLE PRECISION,
    volume      DOUBLE PRECISION,
    confidence  DOUBLE PRECISION,
    raw         JSONB
);
SELECT create_hypertable('wyckoff_events', 'bar_time', chunk_time_interval => INTERVAL '90 days');

-- Wyckoff 阶段快照
CREATE TABLE wyckoff_phases (
    time         TIMESTAMPTZ NOT NULL,
    symbol       VARCHAR(32) NOT NULL,
    timeframe    VARCHAR(8)  NOT NULL,
    phase        VARCHAR(16),         -- ACCUMULATION/DISTRIBUTION/TRANSITION
    confidence   DOUBLE PRECISION,
    cause_score  DOUBLE PRECISION,
    PRIMARY KEY (time, symbol, timeframe)
);

-- ICT 结构事件（与 ta 包共用 ict_structures 表）
-- 使用《技术指标引擎方案》中已定义的 ict_structures 表
-- 新增 BOS/CHoCH/Sweep 类型记录
ALTER TABLE ict_structures ADD COLUMN IF NOT EXISTS reversed BOOLEAN;
```

### 2.4 Redis

| Key | TTL |
|-----|-----|
| `cache:elliott:{symbol}:{tf}` | 1h |
| `cache:wyckoff:{symbol}:{tf}` | 1h |
| `cache:ict_adv:{symbol}:{tf}` | 30m |

### 2.5 调度
- 蜡烛收盘事件触发对应周期重算
- 重大事件（BOS/CHoCH/SPRING）→ pubsub → SSE 推送 + 用户订阅通知

### 2.6 优化与提升

| 维度 | ark-intelligent | AntClaw |
|------|----------------|---------|
| 算法重复 | ta + elliott 各自实现 | DRY，共享 `ict_core` |
| 事件历史 | 内存仅最新 | hypertable 完整历史 |
| 失效检测 | 单次报告 | 持续追踪（CHoCH 可标记为 mitigated）|
| 浪型质量 | 单一计数 | 保留多候选 + 置信度排序 |
| 验证 | 即时验证 | 持续重验：每根新 K 线检查铁律是否仍成立 |

---

## 三、参考文件

- ark-intelligent：`internal/service/{elliott,wyckoff,ict}/*.go`，`ta/{elliott,wyckoff,ict}.go`
- AntClaw proto：`proto/antclaw/v1/strategy.proto`, `ta.proto`
- AntClaw service：待新建 `backend/internal/service/{elliott,wyckoff,ict_advanced}/`
