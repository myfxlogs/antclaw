# AntClaw 策略管理引擎方案

> **版本**：v1.0  
> **对应 ark-intelligent 模块**：`internal/service/strategy/` (5 files)  
> **对应 AntClaw Proto**：`StrategyService`

---

## 一、ark-intelligent 方法分析

### 1.1 核心职责
`strategy/engine.go` 是 ark-intelligent 的 **regime-aware playbook 引擎**：综合多因子排名 + 宏观状态 + COT + 价格状态 → 生成具体的可执行交易计划（Playbook）。

`strategy/risk_parity_sizer.go` 提供组合层面的风险平价仓位管理。

### 1.2 Engine 输入

```go
type Input struct {
    Ranking         *factors.RankingResult     // 因子引擎排名
    MacroRegime     string                      // EXPANSION / SLOWDOWN / RECESSION / RECOVERY
    COTBias         map[string]string           // 每个合约的 COT 偏向
    VolRegime       map[string]string           // EXPANDING / CONTRACTING / NORMAL
    CarryBps        map[string]float64          // 利率差
    TransitionProb  float64                     // 宏观状态切换概率
    TransitionFrom, TransitionTo string
}
```

### 1.3 Playbook 生成

**Algorithm**：
1. 取 factor ranking 前 N 名 → 候选做多
2. 取 factor ranking 后 N 名 → 候选做空
3. 对每个候选：
   - 校验 COT 偏向是否与因子方向一致 → 若一致提升 conviction，相反降低
   - 校验当前宏观 regime 是否利好该方向（EXPANSION 利多 risk asset，RECESSION 利多 safe haven）
   - VolRegime EXPANDING → conviction × 0.7（降低仓位）
   - 处于状态转换期 → `IsTransition=true` → 进一步减仓
   - 加入 carry：高正 carry 长仓加分
4. 输出 `PlaybookEntry[]`，包含 Direction / Conviction / ConvLevel / 完整支持证据

### 1.4 Risk Parity Sizer

**目标**：把每个交易的 ATR 仓位整合到组合层面，控制总风险。

**Algorithm**：
1. 输入：planned positions + account balance + max heat (e.g. 6%)
2. 每个 position 的初始 risk = (entry-stop) × shares
3. 计算总风险 = Σ(risk_i)
4. 若 total > max heat → 等比缩放
5. **Kelly fraction**：用 winRate 和 avgWinLoss 计算 Kelly 比例
   ```
   kelly = (winRate × avgWinLoss - (1-winRate)) / avgWinLoss
   final_risk_per_trade = kelly × max_heat
   ```
6. 波动率调整：EXPANDING regime → 仓位 × 0.7

### 1.5 ConvictionLevel
- `HIGH`：因子前 3 + COT 一致 + Regime fit
- `MEDIUM`：因子前 5 + 部分支撑
- `LOW`：因子靠后或证据不足
- `AVOID`：转换期 + 高波动

---

## 二、AntClaw 设计方案

### 2.1 架构

```
StrategyService (Proto, 已存在)
  ├── service/strategy/playbook    (Playbook 生成器)
  ├── service/strategy/risk_parity (组合风险管理)
  ├── service/strategy/regime      (regime overlay，参见《宏观状态检测方案》)
  └── service/strategy/sizing      (单笔 ATR 仓位)
      ↓
  调用 SignalService / FactorService / MacroService / COTService 的 RPC
```

### 2.2 核心接口

```go
type PlaybookEngine interface {
    Generate(ctx, opts PlaybookOpts) (*Playbook, error)
    GetActive(ctx, userID string) ([]Playbook, error)
    SaveDecision(ctx, decision PlaybookDecision) error
}

type RiskParitySizer interface {
    Allocate(ctx, positions []PlannedPosition, account AccountInfo) (*Allocation, error)
}

type Playbook struct {
    GeneratedAt time.Time
    Regime      string
    Entries     []PlaybookEntry
    GlobalRisk  GlobalRiskAssessment
}

type PlaybookEntry struct {
    Symbol, Direction string
    Conviction        float64  // 0..1
    ConvLevel         string   // HIGH/MEDIUM/LOW/AVOID
    Entry, Stop, Take float64
    PositionSize      float64
    Evidence          map[string]any  // 各子系统证据
    IsTransition      bool
}
```

### 2.3 Schema

```sql
CREATE TABLE playbooks (
    id           BIGSERIAL PRIMARY KEY,
    generated_at TIMESTAMPTZ NOT NULL,
    user_id      VARCHAR(64),
    regime       VARCHAR(32),
    entries      JSONB,
    global_risk  JSONB,
    weights      JSONB         -- 本次使用的权重快照
);
SELECT create_hypertable('playbooks', 'generated_at', chunk_time_interval => INTERVAL '30 days');

CREATE TABLE playbook_decisions (
    id            BIGSERIAL PRIMARY KEY,
    playbook_id   BIGINT REFERENCES playbooks(id),
    user_id       VARCHAR(64),
    symbol        VARCHAR(32),
    decision      VARCHAR(16),  -- 'TAKEN','SKIPPED','MODIFIED'
    actual_entry  DOUBLE PRECISION,
    actual_stop   DOUBLE PRECISION,
    notes         TEXT,
    decided_at    TIMESTAMPTZ
);
```

### 2.4 Redis
| Key | TTL |
|-----|-----|
| `cache:playbook:active:{user_id}` | 1h |
| `cache:strategy:risk:{user_id}` | 30m |

### 2.5 调度
- **被动**：用户调用 `GeneratePlaybook` RPC → 同步生成
- **主动**：每周一 09:00 UTC 为活跃用户预生成本周 playbook（如有订阅）
- **状态切换告警**：MacroRegime 切换 → 自动重生成所有 active playbook

### 2.6 优化与提升

| 维度 | ark-intelligent | AntClaw |
|------|----------------|---------|
| Playbook 持久化 | 内存 | DB，支持回溯审计 |
| 决策反馈 | 无记录 | `playbook_decisions` 表 |
| 多用户 | 单 bot | 用户隔离 + 个性化 |
| Risk parity | 单调用 | 持续监控 + 超限报警 |
| Kelly 参数 | 输入硬编码 | 从用户回测/历史 outcome 自动算 |

---

## 三、参考文件

- ark-intelligent：`internal/service/strategy/{engine,risk_parity_sizer,types}.go`
- AntClaw proto：`proto/antclaw/v1/strategy.proto`
- AntClaw service：`backend/internal/service/strategy/`
