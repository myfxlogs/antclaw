# AntClaw 技术指标引擎方案

> **版本**：v1.0  
> **对应 ark-intelligent 模块**：`internal/service/ta/` (36 files)  
> **对应 AntClaw Proto**：`TAService`

---

## 一、ark-intelligent 方法分析

### 1.1 核心职责
TA 模块是 ark-intelligent **最庞大的服务**，实现了从经典指标到 ICT/Wyckoff/AMT/Elliott 等多种交易学派的结构化分析，并集成多周期 confluence 引擎与 walk-forward 回测。

### 1.2 模块矩阵

| 子模块 | 文件 | 核心算法 / 输出 |
|--------|------|-----------------|
| **engine.go** | TA 主引擎 | 接受 OHLCV，分发至各 analyzer，输出统一 `TAResult` |
| **indicators.go** | 27 KB 指标库 | RSI、MACD、Bollinger、ATR、ADX、CCI、Stoch、ROC、SMA/EMA/WMA、MFI、OBV…… |
| **patterns.go** | K 线形态 | Doji、Hammer、Engulfing、Morning Star、Three White Soldiers… |
| **divergence.go** | 背离 | RSI / MACD 与价格的常规 + 隐性背离 |
| **fibonacci.go** | 斐波那契 | 回撤 / 扩展 / 时间窗 |
| **ichimoku.go** | 一目均衡表 | Tenkan / Kijun / Senkou A/B / Chikou + 信号 |
| **vwap.go** | VWAP | 日 VWAP + 1σ/2σ 带 |
| **supertrend.go** | SuperTrend | 多周期 SuperTrend 趋势线 |
| **zones.go** | 供需区 | 主动 OB-Style 供需区检测 |
| **session_analyzer.go** | 交易时段 | 亚 / 欧 / 美 时段统计 |
| **killzone.go** | ICT Killzone | London / NY 关键时段标记 |
| **delta.go** | 委托价差 | tick-rule delta，累计 delta 估算 |
| **mtf.go** | 多周期 | 日 + 4H + 1H confluence 聚合 |
| **confluence.go** | 共振引擎 | 综合所有指标 + 模式 + ICT，输出 GRADE A/B/C |
| **elliott.go** | 艾略特 | 摆动检测 → 1-5 浪标记 + 规则验证 |
| **ict.go** | ICT | FVG / OrderBlock / BreakerBlock / LiquidityLevel / Killzone |
| **wyckoff.go** | Wyckoff | 5 阶段（Accumulation/Markup/Distribution/Markdown/Transition）|
| **amt_daytype.go** | AMT 1 | Dalton 6 种日类型（Normal/Trend/Double Distribution …）|
| **amt_opening.go** | AMT 2 | Dalton 4 种开盘类型（OpenDrive/OpenTestDrive/OpenRejection/OpenAuction）|
| **amt_rotation.go** | AMT 3 | Rotation Factor（VAH ↔ VAL 半旋转计数）|
| **amt_close.go** | AMT 4 | 收盘位置（AboveVAH/AtPOC/InsideVA/BelowVAL）+ 跟随率 |
| **amt_migration.go** | AMT 5 | POC/VA 多日迁移（UP/DOWN/OVERLAP）|
| **backtest.go** | Walk-forward | CTA confluence 回测，含 GRADE 过滤 |

### 1.3 多周期 Confluence（mtf.go + confluence.go）

**步骤**：
1. 对每个周期（D / 4H / 1H）独立计算 indicators + patterns + ICT
2. 给每个 timeframe 评分（趋势方向 + 强度）
3. 加权求和（D 50% / 4H 30% / 1H 20%）
4. 输出 `TAGrade` (A=高度共振 / B=部分共振 / C=噪声)

### 1.4 ICT 体系（ict.go + killzone.go）

- **FVG**：连续 3 根 K 线，中间 K 线高低形成"价格真空"
- **OrderBlock**：突破前最后一根反向蜡烛
- **BreakerBlock**：被突破后的反向 OB 极性翻转
- **LiquidityLevel**：等高/等低（流动性目标）
- **Killzone**：London Open (07:00 GMT)、NY Open (12:00 GMT) 等高波动时段

### 1.5 AMT（Auction Market Theory）

完整实现 Dalton 五大模块：
- **DayType**：基于 IB（Initial Balance）/ 全日 range 比例分类
- **OpeningType**：开盘相对昨日 Value Area 的位置
- **RotationFactor**：VAH ↔ VAL 半旋转次数
- **CloseLocation**：收盘相对 VA 位置 + 历史跟随率
- **Migration**：多日 POC 迁移方向

### 1.6 Walk-Forward Backtest（backtest.go）

- 以 confluence GRADE 为信号入场
- ATR 止损 + 固定风险百分比
- 走步法（warmup → in-sample → out-sample 滑动）
- 输出 Sharpe / Sortino / max drawdown / win rate / profit factor

---

## 二、AntClaw 设计方案

### 2.1 架构调整

```
TAService (Proto)
  ├── service/ta/engine        (主引擎 + dispatcher)
  ├── service/ta/indicators    (经典指标，纯函数)
  ├── service/ta/patterns      (K 线形态)
  ├── service/ta/ict           (ICT 结构)
  ├── service/ta/wyckoff       (Wyckoff 阶段)
  ├── service/ta/amt           (AMT 五模块)
  ├── service/ta/elliott       (艾略特浪)
  ├── service/ta/confluence    (多周期共振 + GRADE)
  └── service/ta/backtest      (回测，与 BacktestService 复用)
      ↓
  infra/postgres/ta_repo.go    (信号 / GRADE 历史)
  infra/redis/ta_cache.go
```

**vs ark-intelligent 调整**：
- 36 个 文件压缩为分包结构，避免单 800 行限制
- 所有 analyzer **纯函数化**，输入 `[]OHLCV` 输出 result，无副作用
- 回测从 ta 包剥离到独立 `BacktestService`（避免循环依赖）

### 2.2 核心接口

```go
type TAEngine interface {
    Analyze(ctx, bars []OHLCV, opts AnalyzeOptions) (*TAResult, error)
    AnalyzeMTF(ctx, daily, h4, h1 []OHLCV) (*MTFResult, error)
}

type Analyzer[T any] interface {
    Name() string
    Analyze(bars []OHLCV) (*T, error)
}

// 每个子分析器为独立 Analyzer
// 例如：ICTAnalyzer, WyckoffAnalyzer, AMTAnalyzer, ElliottAnalyzer
```

### 2.3 持久化

```sql
-- TA 信号历史（用于回测和校准）
CREATE TABLE ta_signals (
    id          BIGSERIAL PRIMARY KEY,
    symbol      VARCHAR(32),
    timeframe   VARCHAR(8),       -- 'D','4H','1H'
    issued_at   TIMESTAMPTZ,
    signal_type VARCHAR(64),      -- 'FVG_LONG','OB_SHORT','WYCKOFF_SPRING'…
    grade       VARCHAR(2),       -- 'A','B','C'
    direction   VARCHAR(8),       -- 'LONG','SHORT'
    entry_price DOUBLE PRECISION,
    stop_loss   DOUBLE PRECISION,
    take_profit DOUBLE PRECISION,
    metadata    JSONB
);
SELECT create_hypertable('ta_signals', 'issued_at', chunk_time_interval => INTERVAL '90 days');

-- ICT 结构持久化（FVG/OB 长期跟踪）
CREATE TABLE ict_structures (
    id          BIGSERIAL PRIMARY KEY,
    symbol      VARCHAR(32),
    timeframe   VARCHAR(8),
    type        VARCHAR(16),      -- 'FVG','OB','BREAKER','LIQUIDITY'
    formed_at   TIMESTAMPTZ,
    high        DOUBLE PRECISION,
    low         DOUBLE PRECISION,
    direction   VARCHAR(8),
    mitigated   BOOLEAN DEFAULT FALSE,
    mitigated_at TIMESTAMPTZ,
    raw         JSONB
);
```

### 2.4 Redis 缓存

| Key | TTL | 内容 |
|-----|-----|------|
| `cache:ta:result:{symbol}:{tf}` | 5-15min | 完整 TAResult |
| `cache:ta:mtf:{symbol}` | 15min | 多周期共振 |
| `cache:ta:ict:{symbol}:{tf}` | 1h | 当前活跃 FVG/OB 列表 |
| `cache:ta:amt:{symbol}` | 1d | AMT 日类型 |

### 2.5 触发与调度
- **盘中蜡烛收盘事件**：4H 蜡烛收盘 → 触发该周期 TA 重算
- **日线收盘**：每日 22:00 UTC → 全市场 TA 重算 + AMT 多日迁移更新
- **用户请求**：API 命中 Redis 缓存 → 返回；缓存失效 → 同步重算

### 2.6 优化与提升

| 维度 | ark-intelligent | AntClaw |
|------|----------------|---------|
| 包结构 | 36 文件挤一包 | 8-10 个细分子包，符合 800 行规约 |
| 依赖 | 含 ta+backtest+ict 循环 | Hexagonal，回测剥离 |
| 信号持久化 | 每次重算 | 写入 hypertable，可历史回溯 |
| ICT 跟踪 | 单次查询 | 长期 mitigation 状态机 |
| 多周期对齐 | 单次同步计算 | 各周期独立缓存 + 时戳对齐 |
| 测试 | 集成测试为主 | 每个 analyzer 单元测试 ≥80% 覆盖 |

---

## 三、参考文件

- ark-intelligent：`internal/service/ta/*.go`（36 文件）
- AntClaw proto：`proto/antclaw/v1/ta.proto`
- AntClaw service：`backend/internal/service/ta/service.go`
