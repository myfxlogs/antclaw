# AntClaw · 测试策略

> 本文是代码实现阶段**测试工作**的唯一规范。CI 门槛、测试金字塔、黄金文件、压测脚本均以此为依据。违反者 CI 不得放行。

## 〇、与当前仓库同步（2026-04）

- **清场**：占位目录与脚本已删除，包括 `backend/test/`、`scripts/loadtest/sse.js`、`tests/e2e/` 等；事实清单见 `docs/精简重构收口清单.md`「仓库清场」。
- **CI 以仓库为准**：`.github/workflows/ci.yml`（命名检查、`go build`、`buf lint`、前端 lint）、`.github/workflows/strategy-tests.yml`（`go test ./internal/service/... ./internal/engine/...`）、`.github/workflows/similarity-check.yml`（`scripts/ci/similarity-guard.py`）。
- **下文路径**：除明确标注「当前已有」外，均为**规划约定**；落地集成/契约/E2E/k6 目录时，应同步更新本文与对应 workflow。

## 一、测试金字塔

```
           ▲
           │  5% 人工/探索
           │
           │  15% 端到端（Playwright + k6）
           │
           │  30% 集成（真实 PG/Redis/MinIO）
           │
           │  50% 单元（纯 Go / Vitest）
           ▼
```

- **单元**：最多、最快，纯函数 + 值对象 + 聚合根不变式。
- **集成**：启用 `testcontainers-go` 真实起 PG（含 TimescaleDB）/ Redis / MinIO。
- **端到端**：Playwright 驱动 Web 与 Admin；k6 跑 SSE 压测与限流；Connect client 脚本验证 RPC 链。
- **人工**：回归 checklist，不进 CI。

---

## 二、覆盖率门槛（CI 硬约束）

| 范围 | 覆盖率最低要求 | 备注 |
|------|----------------|------|
| `internal/domain/**` | **85%** | 聚合根不变式必覆盖 |
| `internal/service/**` | 80% | happy + error path |
| `internal/auth/**` | 85% | 包含 Argon2id、JWT、session、审计 |
| `internal/byok/**` | 90% | 加密/解密/版本轮换 |
| `internal/notify/**` | 80% | 站内信、推送桥接 |
| `internal/adapter/**` | 60% | 外部系统 adapter |
| `internal/pkg/**` | 70% | 基础工具 |
| `frontend/web/src/**` | 60% | 重点是工具函数与 hooks |
| `frontend/admin/src/**` | 60% | 同上 |
| 全局平均 | 75% | 低于即阻塞合并 |

工具：后端 `go test -cover -coverprofile=cover.out` + `go tool cover`；前端 `vitest --coverage`。

CI 任务 `coverage-gate` 读取覆盖报告，低于阈值 **fail**。

---

## 三、测试分类与命名

### 3.1 单元测试

- 文件：`*_test.go`（Go）、`*.test.ts(x)`（前端）；
- 测试函数：`TestXxx_场景_预期`，如 `TestUser_Register_RejectsEmptyEmail`；
- 前端：`describe('UserForm', () => { it('rejects empty email', ...) })`。

### 3.2 契约测试

- 工具：`bufbuild/connect-go` client + 真实后端；
- 位置（规划）：`backend/test/contract/<service>_test.go`（当前仓库尚未创建 `backend/test/`；清场见 §〇）；
- 每个 Proto 方法至少一个：正常输入 + 至少一个错误码验证；
- 破坏性变更：`buf breaking --against '.git#branch=main'` 作为 CI 任务 `proto-breaking-check`。

### 3.3 集成测试

- 工具：`testcontainers-go`（PG/Redis/MinIO）；
- 位置（规划）：`backend/test/integration/<topic>_test.go`；
- Go build tag：`//go:build integration`；
- 通过 `go test -tags=integration ./backend/test/integration/...` 执行（**当前无该目录**；重建后接入 CI）；
- CI 单独 job（并行跑，时间限额 10 分钟）— 待 workflow 增补。

### 3.4 端到端测试（前端）

- 工具：Playwright；
- 位置（规划）：`frontend/e2e/` 或独立 Playwright 包（当前仓库未创建；清场见 §〇）；
- 场景：
  - 注册 → 登录 → 查看 Dashboard；
  - 切换 locale（en-US / zh-CN）；
  - 提交反馈 → 管理员看见；
  - 站内信实时推送；
  - 密码输入框默认明文；
  - 策略/回测入口显示"后期提供"。

### 3.5 负载与 SLA 测试

- 工具：`k6`；
- 位置（规划）：`scripts/loadtest/`（原占位 `sse.js` 已清场删除；重建脚本时恢复目录）；
- 关键脚本（规划）：
  - `sse.js`：1K 并发 SSE，P95 ≤ 500ms（应对齐当前 SSE 路径，如 `/sse/jobs`）；
  - `rpc_login.js`：登录 RPS 限流验证；
  - `notify_fanout.js`：1W 用户 fanout 吞吐。
- 验收硬指标：见 `AntClaw-重构解决方案.md` §十三A SLA 表。

### 3.6 金融 Golden 测试

- 位置：`backend/internal/domain/<子域>/testdata/*.golden`；
- 要求：
  - 输入数据定期存档（CSV / JSON）；
  - 输出字段 `numeric(20,10)` 严格文本对比；
  - 任何指标算法（EMA / RSI / Sharpe / MaxDrawdown / …）必须有 golden；
  - 更新 golden 需 PR 明确说明原因，Reviewer 双签。

示例：

```
backend/internal/domain/signal/testdata/
  ema_20_eurusd_h1.input.csv
  ema_20_eurusd_h1.golden
```

### 3.7 回测↔实盘一致性

- 集成测试（规划）`backend/test/integration/strategy_consistency_test.go`（目录清场后待重建再激活）；
- 断言：同一 `SignalEvaluator` 实例被回测引擎与实盘信号生成器调用；
- 验证：相同历史片段下，两条路径产生的 `Signal` 序列完全等价。

---

## 四、CI 流水线

### 4.1 并行 Job 列表

| Job | 命令 | 时长预期 |
|-----|------|----------|
| `buf-lint` | `buf lint` | < 30s |
| `buf-breaking` | `buf breaking --against '.git#branch=main'` | < 30s |
| `go-build` | `go build ./...` | < 2min |
| `go-vet` | `go vet ./...` | < 30s |
| `golangci-lint` | `golangci-lint run` | < 2min |
| `go-test-unit` | `go test -race -cover ./...` | < 5min |
| `go-test-integration` | （未配置）清场后无 `backend/test/integration`；重建目录后再启用 `go test -tags=integration ./backend/test/integration/...` | — |
| `coverage-gate` | 聚合覆盖率 | < 30s |
| `pnpm-lint` | `pnpm -r lint` | < 2min |
| `pnpm-test` | `pnpm -r test` | < 3min |
| `pnpm-build` | `pnpm -r build` | < 3min |
| `playwright-e2e` | 启动 compose + 运行 E2E | < 15min |
| `i18n-check` | `scripts/i18n-check.sh`（**未落地**；脚本不存在前勿写入 CI） | — |
| `dependency-audit` | `govulncheck ./... && pnpm audit --prod` | < 2min |

单台 runner 总时长预期 ≤ 30 分钟；启用 matrix 并行后 ≤ 15 分钟。

### 4.2 门槛策略

- 所有 job green 才能合并；
- 覆盖率任一模块低于阈值 fail；
- `buf breaking` 失败 fail；
- E2E flaky 允许最多重跑 1 次，三次内必须稳定。

---

## 五、测试数据约定

### 5.1 夹具位置

- Go（规划）：`backend/test/fixtures/<子域>/*.json`（随 `backend/test/` 一并重建时启用；清场见 §〇）；
- 前端：`frontend/<app>/src/test/fixtures/`；
- Golden：`<包>/testdata/*.golden`。

### 5.2 时间与随机源

- 测试禁用 `time.Now()` / `math/rand.Int()` 直接调用；
- 注入 `clock.Clock`（`github.com/benbjohnson/clock`）与 `crypto/rand` 包装；
- 集成测试用 `time.FixedZone` 固定时区。

### 5.3 数据库隔离

- 每个集成测试用独立 schema（`test_<uuid>`）或独立容器；
- 测试结束 **tear-down** 释放资源；
- 严禁共享 PG 实例跨测试；
- TimescaleDB hypertable 每次测试重建，不共享历史数据。

---

## 六、禁止事项

| 禁止 | 理由 |
|------|------|
| 删除失败的测试 | 直接破坏质量门槛 |
| 将 `Equal` 改为 `NotNil` | 弱化断言 |
| 在测试里 sleep 固定时间 | flaky；用 channel / poll |
| 用真实外部 API 做 CI 测试 | 不可靠、慢、泄露 key |
| 用生产数据做测试 | 隐私与合规 |
| 跳过 race 检测 | 并发 bug 逃逸 |
| 使用 `t.Skip` 绕过未实现功能 | 该任务应报未完成，不是 skip |

---

## 七、Mock / Stub 策略

- 领域层：**不 mock**；直接构造聚合根对象；
- Service 层：
  - 对 **端口接口**用 `gomock` 或手写 fake；
  - 对 **外部系统**用 testcontainers（集成）或 HTTP 测试服务（单元）；
- 前端：
  - Hooks 测试 mock Connect client；
  - 组件测试用 `msw` 拦截请求；
- **不允许**：直接 mock 具体结构体（只 mock 接口）。

---

## 八、安全测试

- 登录限流测试：连续失败后阶梯退避生效；
- JWT 签名切换：Ed25519 公钥轮换下旧 token 拒绝；
- BYOK 密钥泄露回归：日志 / 追踪 / pprof 堆中**不出现**明文密钥；
- 审计不可篡改测试：
  - 直接 `UPDATE audit_logs` 应失败；
  - `DELETE` 应失败；
  - `TRUNCATE` 应失败；
  - hash 链被破坏时校验函数报错。
- 审计 WORM：写入 MinIO bucket 的对象 `DeleteObject` 应被拒绝。

---

## 九、国际化测试

- `i18n-check.sh`：
  - 校验所有 `t('key')` 使用的 key 都在 `locales/{zh-CN,en-US}/*.json` 中存在；
  - 占位符一致性（参数个数与类型一致）；
  - 未使用 key 警告（非 fail）；
- 每个关键消息有快照测试：
  - 后端 `go-i18n` 按 `audience=app, locale=en-US` 渲染 → 期望输出；
  - 前端 `react-i18next` 同上。
- 回退链测试：未知 locale（如 `de-DE`）经过回退后必须得到 `en-US`（app）或 `zh-CN`（admin）。

---

## 十、实时性测试

- k6 `sse.js`（规划脚本；占位已清场，见 §〇）：
  - 1K 并发 SSE 连接；
  - 每 100ms 触发一次 `alert.triggered`；
  - 客户端测量端到端延迟；
  - P95 ≤ 500ms 为硬指标；
  - 断连 10% 客户端，验证 `Last-Event-ID` 恢复无丢事件；
- Prometheus 断言：
  - `stream_delivery_total{result="ok"}` > 99.9% * 总发送；
  - `sse_connections` 稳定；
  - `rpc_duration_seconds{quantile=0.95}` < SLA。

---

## 十一、回滚测试

- 数据库迁移 `down` 脚本必须在 CI 测一次；
- Compose profile 切换（启用副本 / backtest-runner）必须有冒烟测试；
- 大版本升级：用户 session、BYOK 密文、审计 hash 链必须跨版本有效。

---

## 十二、测试工具矩阵

| 场景 | 工具 |
|------|------|
| Go 单元 | 标准 `testing` + `testify/assert`（可选） |
| Go mock | `go.uber.org/mock`（gomock） |
| Go 容器 | `github.com/testcontainers/testcontainers-go` |
| 前端单元 | `vitest` + `@testing-library/react` |
| 前端 mock | `msw` |
| E2E | `@playwright/test` |
| 负载 | `k6` |
| 契约 | `connect-go` + `buf` |
| 覆盖率 | `go tool cover`、`vitest --coverage`、`codecov`（可选私部） |
| 依赖审计 | `govulncheck`、`pnpm audit` |

以上工具已列入依赖白名单（`AntClaw-重构解决方案.md` 附录 A 的测试/开发部分），禁引入其他测试框架。

---

## 十三、失败处理

1. CI 失败 → 助手必须修复；禁止 `skip` 或 `xfail`；
2. 不能自修的故障 → 回报用户；
3. 集成测试 flaky → 首次允许重试；连续 flaky 必须 `t.Log` 复现信息并停下修根因；
4. Playwright flaky → 上报 `playwright-report`，截图与 trace 作为证据。

---

## 十四、测试数据伦理

- 禁止使用真实用户数据；
- 仿真数据的生成脚本放 `scripts/synth-data/`；
- 任何包含 PII 的 fixture 必须脱敏（`user-<序号>@test.local`）。

---

## 十五、验收清单（任务卡附带）

每张任务卡必须勾选：

- [ ] 新增/修改代码有对应单测；
- [ ] 新增 RPC 有契约测试；
- [ ] 覆盖率未回退；
- [ ] E2E 对受影响路径跑通；
- [ ] i18n key 已补齐；
- [ ] 错误码已注册；
- [ ] 压测脚本未失效（若涉及 SSE/RPC；脚本重建后纳入 `scripts/loadtest/`）；
- [ ] 审计事件已写入（写操作）；
- [ ] 回滚方案确认。

---

## 十六、与其他文档的联动

- 依赖：本文 §十二 工具矩阵引用 `AntClaw-重构解决方案.md` 附录 A；
- 命名：i18n key / 事件 / 错误码的命名见 `AntClaw-重构解决方案.md` 附录 B；
- 精度：所有数值比对用 `numeric(20,10)` 字符串等值，禁止容差浮点比较（除非显式声明 `EPS = 1e-12`）。

---

> 测试是质量的最后防线。助手若感测试过严，请用**沟通**替代**绕过**。
