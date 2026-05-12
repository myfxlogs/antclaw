# 策略 / 信号 / 回测 — 真实化实现指南

> 本指南是**代码实现的强制依据**。所有文档以"目标 + 算法 + 数据 + 接口 + 修改清单 + 验证"五段式编写，AI 助手必须严格遵循。
>
> **核心原则：禁止任何硬编码假数据、禁止 `placeholder`、禁止 `TODO` / 简化版。所有计算必须基于真实数据库表中的真实数据。**

---

## 0. 当前问题对照（截止 checkpoint 3）

参考 `@docs/ARK-Intelligent-功能清单.md §2.7`，AntClaw 当前完成度约 20-25%。本指南目标是把完成度提升到 **≥ 95%**。

| 类别 | 现状 | 目标 |
|------|------|------|
| `signals_handler.go` 的 21 个 RPC | 多数返回硬编码假数据或空数组 | 全部基于 `price_daily` / `cot_analyses` / `unified_signals` 等真实表计算 |
| `RunQuantBt` / `RunVpBt` | 只返回 `task_id`，无引擎 | 完整异步引擎 + `backtest_jobs` / `backtest_results` 持久化 |
| `GetAccuracy` | 硬编码 0.68 | 基于 `signal_outcomes` 表统计 |
| Walk-forward / Bootstrap / MFE-MAE / 成本模型 / 校准 | 全部缺失 | 全部实现，写入 `walkforward_history` / `signal_calibration` |
| `/setalert` / `/report` / `/ctabt` | 完全缺失 | 新建表 + RPC + Worker 评估 |

---

## 1. 文档索引（实施顺序 = 优先级）

| 序号 | 文档 | 主题 | 估时 | 依赖 |
|------|------|------|------|------|
| 00 | `00-架构与共享契约.md` | 接口 / DI / Provider 契约 | 0.5d | — |
| 01 | `01-信号服务真实化.md` | bias / rank / xfactors / radar / intensity / unified | 3d | 00 |
| 02 | `02-状态转换与加密Alpha.md` | transition / cryptoalpha | 1.5d | 00 |
| 03 | `03-告警订阅SetAlert.md` | 用户级 COT / 信号告警 | 1d | 00 |
| 04 | `04-准确率统计Accuracy.md` | signal_outcomes 评估 + GetAccuracy | 1.5d | 01 |
| 05 | `05-综合报告Report.md` | /report 多维聚合输出 | 1d | 01,04 |
| 06 | `06-量化与CTA信号.md` | quant / cta 真实计算 | 2d | 00 |
| 07 | `07-回测引擎QuantBT.md` | 完整异步回测引擎 | 4d | 00,06 |
| 08 | `08-成交量轮廓回测VpBT.md` | VP 回测引擎 | 2d | 07 |
| 09 | `09-CTA回测.md` | CTA 回测 | 1.5d | 07 |
| 10 | `10-高级回测能力.md` | walk-forward / bootstrap / MC / MFE-MAE / 成本 / 校准 | 4d | 07 |
| 11 | `11-因子库扩展.md` | 因子库（动量 / 低波 / 趋势 / Carry / 拥挤度 / 残差反转）| 3d | 00 |
| 12 | `12-简报与展望.md` | briefing / outlook + AI 文本生成 | 2d | 01,04 |
| 13 | `13-验证与测试规范.md` | 单测 / 集成 / E2E 验证规范 | 持续 | 全部 |

**总估时：约 27 人天**（理想串行）。

---

## 2. 全局约束（适用于所有子任务）

### 2.1 数据真实性
- **禁止**任何形式的硬编码返回值。所有数值必须来自数据库或基于数据库数据的实时计算。
- 数据不足时返回 `code=DATA_INSUFFICIENT` 错误，**不要**返回伪造数据。
- Mock / 假数据**仅允许**在 `*_test.go` 单测文件中。

### 2.2 计算位置（宪章 5）
- 所有策略 / 聚合 / 筛选 / 排序 / 指标计算必须在后端 Service 层。
- 前端只能调用 RPC 并渲染。

### 2.3 租户隔离（宪章 7）
- 涉及用户数据的查询（playbooks、setalert、回测任务）必须按 `user_id` 过滤。
- 系统级查询（COT/价格/宏观）无 `user_id`，但 RPC 必须验证 JWT。

### 2.4 编码约束
- 单文件 ≤ 800 行（宪章 4）；超过则按职责拆分。
- 错误码先在 `proto/antclaw/v1/common.proto` 注册，再使用（宪章 9）。
- 不修改 `gen/` 自动生成代码（宪章 8）。
- 修改 `.proto` 后必须 `buf generate` 并提交。

### 2.5 验证规范
每完成一个子任务必须执行：
1. `cd backend && go build ./...` 成功；
2. `go test ./internal/service/<module>/... -v -race` 通过；
3. `cd deploy && docker compose up -d --build api worker`；
4. 执行该任务文档"§验证步骤"中的 curl / SQL；
5. 在 `frontend/admin` 中验证 UI 数据正确显示（如适用）。

### 2.6 提交粒度
- 一个子任务一个 PR / commit；commit message 必须引用文档编号，例如：
  `feat(signals): real bias/rank/xfactors per docs/策略信号回测-实现指南/01`

---

## 3. 推荐并行批次

```
Batch A (基础)        : 00 → 01 → 11
Batch B (评估闭环)    : 04 → 05 → 13   依赖 Batch A
Batch C (回测引擎)    : 07 → 08, 09, 10  依赖 Batch A
Batch D (扩展信号)    : 02, 03, 06, 12   可独立或与 B/C 并行
```

---

## 4. 关键依赖数据表清单

实现时**必须使用以下既存表**，不得新建重复表：

| 数据域 | 表 | 主要字段 |
|-------|----|---------|
| 价格 | `price_daily`, `price_intraday`, `price_weekly` | `time, symbol, open, high, low, close, volume` |
| COT | `cot_records`, `cot_analyses` | `report_date, contract_code, cot_index, direction, zscore` |
| 宏观状态 | `macro_regime_history`, `regime_overlay_history`, `regime_transitions` | `unified_label, hmm_state, garch_regime` |
| 资金流 | `flow_divergence_history` | `pair_a, pair_b, z_score, lead_lag` |
| 微观结构 | `volume_profiles`, `orderflow_absorptions` | `poc, vah, val` |
| 信号 | `unified_signals`, `signal_outcomes`, `signal_weight_config` | `unified_score, recommendation, components` |
| 因子 | `factor_rankings`, `factor_ranking_entries` | `rank, raw_score, breakdown` |
| 回测 | `backtest_jobs`, `backtest_results`, `walkforward_history`, `signal_calibration` | — |
| Playbook | `playbooks`, `playbook_decisions` | `entries, global_risk, weights` |

---

## 5. 关键依赖代码模块

| 模块 | 路径 | 用途 |
|-----|------|------|
| 信号引擎 | `backend/internal/engine/signals/engine.go` | bias/radar/unified 计算根入口 |
| 因子引擎 | `backend/internal/service/factors/engine.go` | 多因子排名 |
| 资金流背离 | `backend/internal/service/factors/flow_divergence.go` | x-factors 输入 |
| Playbook | `backend/internal/service/strategy/playbook.go` | 策略组合输出 |
| Risk Parity | `backend/internal/service/strategy/risk_parity.go` | 仓位分配 |
| 回测服务 | `backend/internal/service/backtest/backtest_service.go` | 主回测引擎（待扩展）|
| Worker analyzer | `backend/cmd/antclaw-worker/analyzer.go` | 周期性分析任务 |

---

## 6. 风险与回退

- **新增表必须可幂等创建**：放入 `internal/adapter/storage/postgres/ensure_schema.go`，`CREATE TABLE IF NOT EXISTS`。
- **新增字段不向下兼容时**：先加列，旧代码忽略即可；不要 DROP 列。
- **回测/分析失败**：必须落库到 `backtest_jobs.status='failed'` + `error` 字段，不允许吞错。
- **AI 调用失败**：briefing/outlook 必须有"无 AI 兜底文本"，不允许抛错。

---

## 7. 助手工作纪律

读完本 README 后，AI 助手处理任一子任务时必须：

1. 先读对应编号文档（00 / 01 / ... / 13）；
2. 再读该文档列出的"依赖代码 / 表"；
3. 实施前在 chat 中复述："我将按 docs/策略信号回测-实现指南/XX 实现 ABC，修改 X / Y / Z 文件，写入 P / Q 表，验证步骤为 ..."；
4. 实施完成执行 §2.5 的验证；
5. 在文档末尾追加"实施记录"小节（日期、commit、验证结果）。

不得跳步、不得"等会儿再补"、不得引入未经批准的依赖。
