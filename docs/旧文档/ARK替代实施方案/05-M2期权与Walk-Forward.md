# 05 · M2 期权与 Walk-Forward 回测（2~3 周）

> 目标：补齐 ARK 的 D 类核心能力——GEX、IV Surface、Skew、IV Alerts、Walk-Forward 回测、Bootstrap 显著性。
> 同步补强 Wyckoff、ICT、Microstructure、Orderflow、Elliott、Vol（VIX/MOVE/CrossVol/TermStructure）。

---

## 1. 任务依赖图

```
T1 Deribit 薄 client                ─┐
T2 Bybit 薄 client                  ─┤
                                     ├─→ T5 OptionsService（GEX/IV/Skew/Alerts）
T3 service/options 业务服务         ─┤
T4 service/vol 扩展 (MOVE/CrossVol) ─┘

T6 BacktestService 扩展（WalkForward/Bootstrap/Cost/Trend）
T7 Wyckoff/ICT/Microstructure/Orderflow/Elliott 补强
T8 前端模块：features/options、features/vol、features/backtest/walkforward
```

---

## 2. T1 Deribit 薄 client

### 2.1 文件

```
backend/internal/infra/apiclient/deribit/
├── client.go
├── endpoints.go
├── types.go
├── parse.go
├── errors.go
└── client_test.go
```

### 2.2 端点（仅公共，无需鉴权）

| 方法 | 端点 | 用途 |
|---|---|---|
| `GetBookSummaryByCurrency(ctx, currency)` | `/api/v2/public/get_book_summary_by_currency` | 取所有期权快照 |
| `GetVolatilityIndexData(ctx, currency, start, end, resolution)` | `/api/v2/public/get_volatility_index_data` | DVOL 指数序列 |
| `GetInstruments(ctx, currency, kind, expired)` | `/api/v2/public/get_instruments` | 合约清单（用于 IV Surface 网格） |
| `GetIndexPrice(ctx, indexName)` | `/api/v2/public/get_index_price` | 现货指数 |

### 2.3 字段（关键示例）

```go
type BookSummary struct {
    InstrumentName     string  `json:"instrument_name"`
    MarkIV             float64 `json:"mark_iv"`            // %（除以 100 转小数）
    MarkPrice          float64 `json:"mark_price"`
    UnderlyingPrice    float64 `json:"underlying_price"`
    OpenInterest       float64 `json:"open_interest"`
    Volume             float64 `json:"volume"`
    InterestRate       float64 `json:"interest_rate"`
    Bid                float64 `json:"bid_price"`
    Ask                float64 `json:"ask_price"`
}
```

### 2.4 测试

`client_test.go` 必须包含：

- 用 ARK 仓库 `Emulator/ark-intelligent/internal/service/marketdata/deribit/client_test.go` 的 fixture 做对照
- 失败用例：429 限流后退避；500 重试；403 不重试

---

## 3. T2 Bybit 薄 client

### 3.1 文件

```
backend/internal/infra/apiclient/bybit/
```

### 3.2 端点

| 方法 | 端点 | 用途 |
|---|---|---|
| `GetKline(ctx, category, symbol, interval, start, end, limit)` | `/v5/market/kline` | OHLC |
| `GetOrderbook(ctx, category, symbol, limit)` | `/v5/market/orderbook` | 实时盘口（仅快照） |
| `GetFundingHistory(ctx, category, symbol, start, end, limit)` | `/v5/market/funding/history` | Funding rate |
| `GetOpenInterest(ctx, category, symbol, intervalTime)` | `/v5/market/open-interest` | OI 时序 |

### 3.3 注意

- Bybit `category` 必填：`linear` / `inverse` / `spot` / `option`
- 限速：5 RPS（公共端点，鉴权后更高），保守起见配置 RPS=3 Burst=6

---

## 4. T3 service/options 业务服务

### 4.1 文件

```
backend/internal/service/options/
├── service.go            # NewService(deribit, bybit, ...) *Service
├── gex.go                # GEX 计算（参考 ARK gex/calculator.go）
├── iv_surface.go         # IV 曲面构建
├── skew.go               # 25Δ RR / BF 计算
├── alerts.go             # IV 告警（skew_extreme / iv_spike / term_inversion）
├── types.go
└── service_test.go
```

### 4.2 GEX 算法（核心）

```
Gamma per option = N'(d1) / (S * sigma * sqrt(T))    // Black-Scholes
Dealer Gamma = Σ over options [ -Sign * OI * ContractMultiplier * S^2 * 0.01 * Gamma ]
   - Calls 的 dealer 净 Gamma 通常为正（buy-write 结构假设，可参考 ARK 注释）
   - Puts 的 dealer 净 Gamma 通常为负
   - Sign 由 ARK 公式给定，迁移时严格对照 calculator.go
```

**单元测试要求**：拷贝 ARK `gex/calculator_test.go` 的输入/期望输出，结果偏差 ≤ 1e-6。

### 4.3 IV Surface 构建

- 抓 `GetInstruments` + `GetBookSummaryByCurrency`
- 按 `(strike, dte, type)` 三元组聚合
- 缓存到 Postgres `iv_surface_snapshots(asset, taken_at, strike, dte, option_type, iv)`
- 返回最新一份

### 4.4 Skew 计算

- 从 IV Surface 中线性插值得到 25Δ Call / 25Δ Put / ATM 的 IV
- `RR_25d = IV_call_25d - IV_put_25d`
- `BF_25d = (IV_call_25d + IV_put_25d) / 2 - IV_ATM`
- `Skew_Z` 用过去 90 天历史均值/标准差归一

### 4.5 Alerts

- `skew_extreme`：`abs(Skew_Z) > 2`
- `iv_spike`：ATM IV 当日变化 > 30%
- `term_inversion`：30D IV > 90D IV

告警通过 SSE `/sse/options_alerts`（新增 handler）推送。

---

## 5. T4 VolService 扩展（MOVE/CrossVol/TermStructure）

### 5.1 文件

扩展现有 `backend/internal/service/vol/`：

```
service/vol/
├── service.go            # 已有
├── move.go               # MOVE 指数（来源：FRED MOVE 系列或 finviz/yahoo 抓取）
├── cross_vol.go          # 综合 VIX+MOVE+DVOL_BTC+DVOL_ETH
├── term_structure.go     # 期限结构（IV by DTE）
└── *_test.go
```

### 5.2 数据流

- MOVE：FRED 没有官方 series，使用 ICE 公开数据或 Yahoo `^MOVE`，落 `vol_move_history` 表
- CrossVol：聚合内存计算，无需独立表
- TermStructure：复用 `service/options/iv_surface.go` 输出

---

## 6. T5 OptionsService Connect Handler

### 6.1 文件

```
backend/internal/adapter/rpc/options_handler.go
```

实现 `proto/antclaw/v1/options.proto` 中所有 RPC。

### 6.2 注册

```go
// cmd/antclaw-api/main.go
optionsHandler := rpc.NewOptionsHandler(optionsSvc)
mux.Handle(antclawv1connect.NewOptionsServiceHandler(optionsHandler))
```

---

## 7. T6 BacktestService 扩展

### 7.1 Walk-Forward

**文件**：`backend/internal/service/backtest/walkforward.go`

算法：

```
windows = []
t = start
while t + train_days + test_days <= end:
    train_window = [t, t + train_days)
    test_window  = [t + train_days, t + train_days + test_days)
    in_sample  = run_strategy(train_window)
    out_sample = run_strategy(test_window)
    windows.append({train_window, test_window, in_sample, out_sample})
    t += test_days  // rolling forward
return aggregate(windows.out_sample)
```

聚合 PerfMetrics：每个窗口的 out-of-sample 平均。

### 7.2 Bootstrap 显著性

**文件**：`backend/internal/service/backtest/bootstrap.go`

```
returns = strategy_daily_returns
N = iterations  // 默认 1000
for i in 0..N:
    resampled = sample_with_replacement(returns, len(returns))
    sharpe_i = compute_sharpe(resampled)
p_value = count(sharpe_i <= 0) / N
```

### 7.3 Cost Model

**文件**：扩展 `backend/internal/service/backtest/costs.go`（如不存在则新建）

每笔交易扣减：

```
gross_pnl = (exit - entry) * size
commission = abs(gross_pnl) * commission_bps / 1e4
slippage   = abs(gross_pnl) * slippage_bps / 1e4
borrow     = (holding_days / 365) * notional * borrow_bps / 1e4
net_pnl    = gross_pnl - commission - slippage - borrow
```

### 7.4 Daily Trend Filter

**文件**：`backend/internal/service/backtest/trend_filter.go`

接收 `filter` 参数（`ema200|sma50|adx14`）：

- 仅在过滤通过的交易日开仓
- 返回过滤前后 PerfMetrics 对比

---

## 8. T7 业务服务补强

### 8.1 Wyckoff（对齐 ARK 916 行实现）

```
service/wyckoff/
├── engine.go             # 现 272 行 → 拆分
├── classifier.go         # ARK：阶段分类器
├── phase.go              # ARK：四阶段判别（accumulation/markup/distribution/markdown）
├── events.go             # ARK：spring/upthrust/sign_of_strength/sign_of_weakness 等事件
├── summary.go            # ARK：汇总
└── *_test.go
```

迁移规则：拷贝 ARK `wyckoff_test.go` 的所有用例 → AntClaw 必须全部通过。

### 8.2 ICT（对齐 ARK 622 行）

```
service/ict/
├── engine.go
├── fvg.go                # Fair Value Gap
├── liquidity.go          # Liquidity sweep
├── orderblock.go         # ARK 此文件仅 9 行占位，需独立实现完整
├── structure.go          # Market structure shift
├── swing.go              # Swing high/low
└── *_test.go
```

### 8.3 Microstructure / Orderflow / Elliott

| 包 | 文件清单 |
|---|---|
| `service/microstructure/` | engine.go, depth.go, imbalance.go, *_test.go |
| `service/orderflow/` | engine.go, delta.go, poc.go, absorption.go, *_test.go |
| `service/elliott/` | engine.go, zigzag.go, projector.go, validator.go, types.go, *_test.go |

每个文件必须 < 400 行，超过则按指标拆。

---

## 9. T8 前端模块

### 9.1 features/options（新建）

```
frontend/admin/src/features/options/
├── pages/
│   ├── GEXPage.tsx
│   ├── IVSurfacePage.tsx
│   ├── SkewPage.tsx
│   └── AlertsPage.tsx
├── components/
│   ├── GEXBucketChart.tsx
│   ├── IVSurface3D.tsx       # 使用 plotly.js 或 echarts-gl
│   └── SkewSeries.tsx
└── api.ts                     # 仅 connect-web 客户端调用
```

### 9.2 features/vol（扩展）

新增 `MOVEPage.tsx`、`CrossVolPage.tsx`、`TermStructurePage.tsx`。

### 9.3 features/backtest/walkforward（新建）

```
features/backtest/walkforward/
├── pages/WalkForwardPage.tsx
├── components/WindowsTable.tsx
└── api.ts
```

UI 元素：

- 输入：strategy_id、start、end、train_days、test_days、cost
- 输出：每个窗口 in-sample / out-sample 指标对比表 + 汇总卡片
- 触发：点击 "Run" → 调 `BacktestService.RunWalkForward` → 显示进度（来自 SSE `stream:backtest_progress`）

---

## 10. 数据库迁移

新增 / 扩展表（迁移文件位于 `backend/internal/postgres/migrations/`）：

```sql
-- iv_surface_snapshots
CREATE TABLE iv_surface_snapshots (
  asset TEXT NOT NULL,
  taken_at TIMESTAMPTZ NOT NULL,
  strike DOUBLE PRECISION NOT NULL,
  dte INT NOT NULL,
  option_type TEXT NOT NULL,
  iv DOUBLE PRECISION NOT NULL,
  PRIMARY KEY (asset, taken_at, strike, dte, option_type)
);
CREATE INDEX idx_iv_surface_asset_taken ON iv_surface_snapshots (asset, taken_at DESC);

-- vol_move_history
CREATE TABLE vol_move_history (
  time TIMESTAMPTZ NOT NULL PRIMARY KEY,
  value DOUBLE PRECISION NOT NULL
);

-- backtest_walkforward_runs
CREATE TABLE backtest_walkforward_runs (
  run_id UUID PRIMARY KEY,
  strategy_id TEXT NOT NULL,
  params JSONB NOT NULL,
  result JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- calibration_models
CREATE TABLE calibration_models (
  id UUID PRIMARY KEY,
  strategy_id TEXT NOT NULL,
  method TEXT NOT NULL,
  brier_before DOUBLE PRECISION,
  brier_after DOUBLE PRECISION,
  blob BYTEA,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

---

## 11. M2 验收清单

- [ ] Deribit、Bybit 薄 client 完成；测试 ≥ 70%
- [ ] OptionsService.GetGEX/GetIVSurface/GetSkew/GetIVAlerts 端到端可调
- [ ] VolService.GetMOVE/GetCrossVol/GetTermStructure 上线
- [ ] Wyckoff/ICT/Microstructure/Orderflow/Elliott 拷贝 ARK 测试全部通过
- [ ] BacktestService.RunWalkForward 支持 100 个窗口；可在 UI 看到结果
- [ ] BootstrapSignificance 返回的 p_value 与 ARK fixture 一致（容差 1e-3）
- [ ] CalibrateConfidence 通过 Platt + isotonic 两种方法落库
- [ ] features/options 与 features/backtest/walkforward 上线，**无 fetch/setInterval**
- [ ] 容器重启后健康，所有新 Job 在 18+N 范围内可见状态
- [ ] 差距清单 §1.7、§1.2、§1.11 状态全部 ✅
