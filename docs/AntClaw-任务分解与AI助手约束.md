# AntClaw · 任务分解与 AI 助手约束

> 本文档是 **代码实现阶段** 的唯一工作手册。AI 助手（以下简称「助手」）必须严格遵守本文的约束。任何本文未覆盖或产生歧义的情形，助手**必须停下并提问**，不得自由发挥。
>
> 本文档与 `AntClaw-重构解决方案.md` 同为单一真相来源：重构方案决定"做什么"，本文决定"如何做以及不做什么"。

## 一、适用范围

- 对象：任何承担 AntClaw 代码实现的 AI 助手（Cascade、Claude Code、Cursor Agent 等）。
- 覆盖：`backend/` Go 代码、`frontend/` TypeScript/TSX 代码、`proto/` 契约、`deploy/` Docker/Compose、CI 配置、中文 Markdown 文档。
- 不覆盖：设计决策变更、需求增删、文档内容改写（见 §二 硬约束 1）。

---

## 二、AI 助手硬约束（十条宪章）

### 宪章 1 · 文档为王，助手不得改文档

- 助手**禁止修改** `docs/` 下任何已存在的 `.md` 文件，除非用户显式要求。
- 助手**禁止**在代码注释、Git 提交信息、PR 描述中擅自"澄清"或"补充"设计决策。任何分歧一律回到用户处确认。
- 发现文档内部矛盾或缺失，助手应以**问题清单**形式返回，由用户决定。

### 宪章 2 · 先分析，再实现

每个任务开始前，助手必须给出：

1. 我读懂的需求（一句话）；
2. 涉及的文档章节引用（到小节号）；
3. 拟修改/新增的文件清单；
4. 潜在风险与替代方案。

在用户或任务卡确认后再动手。**禁止在模糊需求下先写代码后补解释**。

### 宪章 3 · 最小改动原则（仅限 AntClaw 仓内已交付代码）

- 本宪章适用范围：本次重写完成后的**常规迭代维护**阶段，或同一 PR 内对**本仓已存在 AntClaw 代码**的调整。
- **不适用**于 P0–P7 重写阶段对**参照项目 `ark-intelligent` 仓库**的任何代码；见宪章 11。
- 在适用范围内：优先修改而非新建；优先复用 AntClaw 自有抽象而非重复构造。
- 不删除已有测试、不弱化断言。修 bug 必须**先写回归测试**。
- 不引入未经授权的依赖（见 `docs/AntClaw-重构解决方案.md` 附录 A）。

### 宪章 4 · 单文件 ≤ 800 行

- 任何代码、YAML、JSON、SQL、脚本文件（**含生成物**）超过 800 行必须拆分。
- Markdown、CHANGELOG、锁文件（`go.sum`、`pnpm-lock.yaml`）不受此限。
- 拆分按职责拆，不按行数机械切割。

### 宪章 5 · 所有业务计算在后端

- 前端只做 RPC 调用、响应渲染、UI 状态管理、表单校验（语法层），**严禁**策略/聚合/筛选/排序/指标计算。
- 见示例与反例见 §六·前端代码红线。

### 宪章 6 · 密码输入明文显示

- 所有前端密码输入框**统一不用 `type="password"` 黑点遮罩**。
- 仅 TLS + Argon2id 承担密码安全。

### 宪章 7 · 严格的租户隔离

- 任何从 `ctx` 提取的 `user_id` 必须传达到数据访问层；所有 SQL/Redis/对象存储查询**默认以 `user_id` 过滤**，除非是管理员操作。
- 违反此约束的 PR 不得合并。CI 应有 lint 规则检测裸露的无过滤查询。

### 宪章 8 · 先读生成代码，再读手写代码

- `gen/` 目录是 **buf 自动生成**的产物，助手**不得手工修改**其内任何文件。
- 修改接口请改 `.proto` 文件再 `buf generate`。

### 宪章 9 · 错误码与事件 key 注册在册

- 新增错误码必须先更新 `docs/AntClaw-重构解决方案.md` 附录 B.4（由用户改）与 `common.proto` 枚举。
- 新增流事件 channel 同理（附录 B.1 + `stream.proto`）。
- 助手若需新增但尚未注册，应**提问**，不得私自定义。

### 宪章 10 · 不留「TODO」、不留「简化版」

- 禁止提交带有 `TODO`、`FIXME`、`XXX`、`simplified for now`、`quick hack`、`临时`、`待完善` 之类的代码。
- 功能未完成则本任务不算完成；把未完项回报给用户。

### 宪章 11 · 全面重写边界（硬约束）

- AntClaw 为**全新工程**，与参照项目 `ark-intelligent` **无代码继承关系**。
- **严禁**以任何形式搬运参照项目代码，包括但不限于：
  - 复制/粘贴 Go、TypeScript、SQL、YAML、Dockerfile、Shell 源文件；
  - 逐行改名/重命名后提交（语义上等同于搬运）；
  - 采用参照项目的函数/方法/结构体字面名称作为新代码标识符（除非由 `.proto` 契约或领域术语确定）；
  - 将参照项目文件目录布局整体导入 AntClaw 仓。
- **允许参考**的维度仅限于：
  - 功能清单与业务语义（通过 `docs/AntClaw-功能清单.md` 与 `ARK-Intelligent-功能清单.md` 对齐）；
  - 数据迁移对接（`docs/AntClaw-迁移指南.md` 重程 BadgerDB 数据，仅读取数据格式，不复用其代码）。
- AntClaw 采用**自有代码风格与命名规范**：Go 包/文件/类型/函数命名以 `docs/AntClaw-领域模型.md` 为准，TS 以 `docs/AntClaw-前端架构.md` 为准，**不以参照项目命名为准**。
- CI 硬检查：任意提交包含 `ark` / `ARK` 字样（除 `docs/ARK-Intelligent-功能清单.md` 参照文档外），或检出源文件与参照项目高位相似度（调查策略由运维于 P0 阶段配置），均阻断合并。

---

## 三、工作流

### 3.1 任务来源

- 任务由用户依照「路线图阶段」下发（P0、P1、…）；
- 每个阶段对应一个或多个**任务卡**（见 §五）；
- 任务卡以 GitHub Issue / chat 片段形式给出，至少包含：目标、输入文档引用、验收点。

### 3.2 助手工作环节

```
接单 → 分析（宪章 2）→ 计划回报 → 用户批准 → 实现 →
编译/测试/lint → 自检对照任务卡 → 报告完成度 → 用户验收
```

### 3.3 工具使用

- **编译检查**：`go build ./...` + `pnpm -r build` + `buf build` 必须通过。
- **静态检查**：`go vet ./...` + `golangci-lint run` + `pnpm -r lint` + `buf lint`。
- **测试**：`go test ./... -race` + `pnpm -r test`。
- **契约兼容**：`buf breaking --against '.git#branch=main'`。
- **依赖审计**：`govulncheck ./...` + `pnpm audit --prod`。

以上任一未通过，任务不得报完成。

### 3.4 产物规范

每次提交必须包含：

1. 代码变更（尽量一个 PR 一件事）；
2. 相应单测 / 契约测试 / e2e 测试；
3. CHANGELOG 段落（只在根 `CHANGELOG.md`）；
4. 若变更 proto：`buf generate` 后提交 `gen/`；
5. 若变更数据库：`db/migrations/<序号>_<描述>.sql` + `sqlc generate`。

---

## 四、任务分解矩阵（与路线图对齐）

下表中「输入文档」指助手必须先读完且作为唯一依据的文档章节。「产物」是本任务完成后仓库应出现的具体文件/变更。

### P0 · 准备（1 天）

| 项 | 内容 |
|----|------|
| 目标 | 新建 AntClaw 仓库与基线分支 `main`；落实命名基线（§2.1）与 CI lint（零容忍 `ark` / `ARK` 字样）；建立 monorepo 空骨架 |
| 输入文档 | `docs/AntClaw-重构解决方案.md §一、§2`、本文 §二 宪章 11 |
| 产物 | 空 `backend/` `frontend/` `proto/` `deploy/` `docs/` `scripts/` 目录；`go.work`、`pnpm-workspace.yaml`、`buf.yaml`、`buf.gen.yaml`、根 `README.md`（中文）、`.gitignore`、`.editorconfig`；CI 配置 `ark` 关键词阻断与相似度检查 |
| 验收 | `grep -rniI "ark" -- . ':!docs/ARK-Intelligent-功能清单.md'` 结果为 0；仓内零行来自 `ark-intelligent` 的源代码 |
| 禁止 | 克隆/拷贝 `ark-intelligent` 仓库任何源代码到本仓；写任何业务代码 |

### P2 · Monorepo 骨架（1 天）

| 项 | 内容 |
|----|------|
| 目标 | 按 §2.3 建立完整目录，空实现占位 |
| 输入文档 | §2.3、附录 A |
| 产物 | 按目录树创建所有 package（允许文件只包含 `package` 声明+注释）；`Dockerfile.backend/web/admin` 空壳；`docker-compose.yaml` 最小可跑 |
| 禁止 | 擅自新增未列目录 |

### P3 · 契约先行（2 天）

| 项 | 内容 |
|----|------|
| 目标 | 按 `AntClaw-重构解决方案.md` 对应章节写全 `.proto` 文件 |
| 输入文档 | §三、附录 B（错误码与 channel）、`docs/AntClaw-领域模型.md`（各实体字段） |
| 产物 | `proto/antclaw/v1/*.proto`（见目录树完整列表）；`buf build` + `buf lint` + `buf breaking`（相对 `main`）通过；`gen/` 代码生成无冲突 |
| 约束 | 错误码集中在 `common.proto`；事件枚举集中在 `stream.proto`；不允许 `google.protobuf.Any` |

### P4 · 存储与鉴权（3 天）

| 项 | 内容 |
|----|------|
| 目标 | PG + TimescaleDB 就绪；用户系统、Argon2id、Ed25519 JWT、会话；审计 append-only |
| 输入文档 | §7.1、§十一、`docs/AntClaw-用户系统与鉴权.md` |
| 产物 | `db/migrations/` 全量建表；`sqlc` 生成；`auth` 包：注册/登录/会话/刷新/登出/密码找回；`audit` 服务；审计触发器 SQL |
| 验收 | `AuthService` 所有方法契约测试通过；审计 `UPDATE/DELETE` 被 PG 拒绝；hash 链通过校验测试 |

### P4a · BadgerDB 迁移（1 天）

| 项 | 内容 |
|----|------|
| 目标 | `backend/cmd/antclaw-migrate` 首次部署必跑 |
| 输入文档 | §7.4、`docs/AntClaw-迁移指南.md` |
| 产物 | CLI 程序；按领域读旧 Badger → 写 PG；幂等；日志详尽 |
| 验收 | 对相同 Badger 输入重跑，PG 行数稳定；失败可 resume |

### P5 · 业务服务接入（7-10 天）

拆分子任务，每个服务一卡：

- P5.1 CotService、P5.2 CalendarService、P5.3 MacroService、P5.4 PriceService、P5.5 VolService、P5.6 SignalsService、P5.7 BacktestService（**仅契约，RPC 返回 `UNIMPLEMENTED`**）、P5.8 TAService、P5.9 SentimentService、P5.10 AIService、P5.11 AlertsService、P5.12 AdminService、P5.13 UserService。

每个子任务产物：`internal/service/<name>/`、`adapter/rpc/<name>_handler.go`、契约测试、单测 ≥ 80%。

### P6 · 异步与 SSE（3 天）

| 项 | 内容 |
|----|------|
| 目标 | `stream.proto` 统一契约 + Redis Streams + SSE 网关 + `Last-Event-ID` |
| 输入文档 | §十三A、`docs/AntClaw-订阅与SSE.md` |
| 产物 | `adapter/sse`、`adapter/storage/redis` Streams 客户端、`Last-Event-ID` 游标表、k6 压测脚本（规划目录 `scripts/loadtest/`；原占位 `sse.js` 已清场，见收口清单） |

### P6b · 审计 WORM（0.5 天）

| 项 | 内容 |
|----|------|
| 目标 | 审计双写 MinIO `audit-worm` bucket + object lock |
| 输入文档 | §7.1 审计硬约束、§十一 |
| 产物 | `audit` 服务的 MinIO 双写实现；`audit-worm` bucket 初始化脚本；单测覆盖双写失败处理 |

### P6c · 策略/MT 结构预留（0.5 天）

| 项 | 内容 |
|----|------|
| 目标 | 仅骨架，**不实现执行** |
| 输入文档 | §十A（开头的 MVP 说明）、§2.3 proto 目录 |
| 产物 | `user_strategies` 迁移；`strategy.proto` / `mt4.proto` / `mt5.proto` 空定义 + 服务 RPC 返回 `codes.Unimplemented`；`cmd/antclaw-backtest-runner/` 目录骨架（`main.go` 仅打印日志后退出）；`internal/adapter/sandbox/` 目录骨架（空实现）；前端路由占位页面显示「后期提供」 |
| 禁止 | 引入 `go.starlark.net/starlark` 依赖；启动 runner 容器；实现任何 built-in 函数 |

### P7 · Bot 接口预留（0.5 天）

| 项 | 内容 |
|----|------|
| 目标 | `BotAdapter` 端口 + `bot_bindings` 表 + 空实现 |
| 输入文档 | §十、`docs/AntClaw-Bot接入规范.md` |
| 产物 | `internal/ports/bot.go`；`adapter/bot/` 目录；单测通过 |
| 禁止 | 引入 Telegram / WeChat 等 SDK |

### P8 · i18n 基础设施（2 天）

| 项 | 内容 |
|----|------|
| 目标 | 后端 go-i18n + 前端 react-i18next + 资源目录 + CI i18n-check |
| 输入文档 | §十二、`docs/AntClaw-国际化规范.md` |
| 产物 | `frontend/packages/i18n/locales/{zh-CN,en-US}/`；后端按 `audience` 做回退；邮件/告警模板双语；CI `scripts/i18n-check.sh` |

### P9 · 通知与站内信（2 天）

| 项 | 内容 |
|----|------|
| 目标 | `notifications` 表 + `notify` 服务 + SSE 推送；过渡期不走邮件 |
| 输入文档 | §八、§7.1 表定义 |
| 产物 | `internal/notify/`；站内信模板；SSE 频道 `notify.new`；FCM/APNs/HMS 桥接骨架（不投产） |

### P9a · AI 用量与密钥探针（0.5 天）

| 项 | 内容 |
|----|------|
| 目标 | `ai_usage` 写入 + 前端用量展示 + BYOK 每日探针 |
| 输入文档 | §4.5、§7.1 |
| 产物 | AI 调用中间件写 `ai_usage`；用户设置页"本日/本月 token 开销"卡片；`worker` 中的 `JobByokHealthCheck` |

### P10 · 前端 Web（6-8 天）

| 子任务 | 目标 |
|--------|------|
| P10.1 基础骨架 | Vite + Tailwind + shadcn/ui + 路由 + query + connect-web |
| P10.2 登录/注册/设置 | 含密码明文、BYOK 配置、时区 |
| P10.3 Dashboard | COT、Calendar、Macro 只读面板 |
| P10.4 Price/Vol/Signals | 只读列表，不做 K 线 |
| P10.5 AI Chat | 与 BYOK 对接 |
| P10.6 Alerts + 站内信 | SSE 订阅；通知抽屉 |
| P10.7 策略/回测页面占位 | 展示「后期提供」 |

### P11 · Admin 前端（5 天）

按 §五 管理员控制台模块拆分子任务。

### P12 · 对象存储接入（1 天）

| 项 | 内容 |
|----|------|
| 目标 | MinIO SDK 封装 + 回测/附件上传下载 RPC |
| 输入文档 | §7.3 |
| 产物 | `adapter/storage/objectstore/`；`ObjectService.Presign` / `Download`；前端附件上传组件 |

### P13 · 硬化（3 天）

| 项 | 内容 |
|----|------|
| 目标 | 压测、Sentry 接入、Caddy 生产配置、文档 review |
| 输入文档 | §十三、§十四、`docs/AntClaw-部署指南.md` |
| 产物 | `deploy/caddy/Caddyfile`、`deploy/seccomp-*.json`、`scripts/loadtest/`（规划；占位已清场）、`.env.example`、Sentry 初始化代码 |

### P14 / P15 · 移动端（后期）

按 `docs/AntClaw-移动端架构.md` 再行拆分。

---

## 五、任务卡模板（每个 PR 对应一张）

```
## 任务卡 · <编号> <标题>

### 1. 目标（一句话）
<...>

### 2. 输入文档
- docs/AntClaw-重构解决方案.md §x.y
- docs/<其他>.md §z

### 3. 改动范围
- 新增：<文件清单>
- 修改：<文件清单>
- 删除：<文件清单>

### 4. 验收点
- [ ] 单测覆盖率 ≥ 80%
- [ ] 契约测试通过
- [ ] `buf breaking` 通过
- [ ] CI 全绿
- [ ] 任务卡目标在验收环境可复现

### 5. 预期风险
<...>

### 6. 回滚方案
<...>
```

助手在开工前**必须**提交此模板的草稿给用户过目。

---

## 六、代码风格规范

### 6.1 后端 Go

- 模块：`github.com/antclaw/antclaw`，`go 1.22+`。
- 格式：`gofmt -s`；import 分三组：标准库、第三方、本地。
- 命名：包名小写无连字符；接口名`<能力>er`（如 `Notifier`）；错误变量 `ErrXxx`；常量 `UpperCamel`。
- 错误处理：
  - 外层始终用 `fmt.Errorf("%s: %w", op, err)` 加 `op = "pkg.Func"`；
  - RPC 返回 `connect.NewError(connect.CodeX, ...)`，`code` 取自 `common.proto` 枚举映射；
  - 禁止 `panic` 用于流控；`panic` 仅用于**程序员错误**（不变式破坏）。
- 并发：
  - 每个 goroutine 必须有明确的退出条件与 `ctx` 传递；
  - 禁用 `time.After` 在循环里（泄漏）；用 `time.NewTimer`。
- 日志：`zerolog.Ctx(ctx).Info().Str("user_id", ...).Msg("...")`；不打印密码、token、API key；错误等级使用 `Err(err)`。

### 6.2 前端 TypeScript

- 严格模式：`strict: true`、`noUncheckedIndexedAccess: true`。
- 路由：TanStack Router；数据：TanStack Query；请求：Connect-Web。**禁用** `axios/redux/mobx`。
- 组件：优先 shadcn/ui 基础组件 + 业务组件层薄封装。
- 样式：Tailwind 原子类；避免 `style={{...}}`；颜色只用 Design Tokens。
- 国际化：文案统一走 `t('key')`；**禁止**硬编码中文/英文。
- 密码字段：`<Input type="text" autoComplete="current-password">`，默认明文；提供 `<Eye />` 切换按钮仅做 UX 美化（不改变遮罩）。
- 前端不得出现：
  - `sort()` 业务排序（由后端返回）；
  - `filter()` 业务筛选（同上）；
  - 聚合运算（`reduce` 求总和等）；
  - 金额 / 收益率计算（调后端计算 RPC）。

### 6.3 SQL / 迁移

- 迁移文件命名：`<14位时间戳>_<snake_case>.sql`，成对 `up/down`。
- 所有业务表必须包含 `created_at`、`updated_at`；金额/价格字段 `numeric(20,10)`。
- 时间戳用 `timestamptz` 并存储 UTC。
- 每个时间序列表建立后立即 `SELECT create_hypertable(...)`。

### 6.4 Proto

- 文件头统一 `syntax = "proto3";` + `package antclaw.v1;` + `option go_package = "github.com/antclaw/antclaw/gen/go/antclaw/v1;antclawv1";`。
- 字段 `snake_case`；消息 `PascalCase`；枚举值 `UPPER_SNAKE`，首项 `<NAME>_UNSPECIFIED = 0`。
- 强制 `protovalidate` 校验标签（`[(buf.validate.field) = {...}]`）。
- 破坏性变更必须过 `buf breaking`；不兼容情形请先 bump minor 版本路径（`antclaw/v2/`）。

### 6.5 Docker / Compose

- 所有镜像基于 `distroless`（后端）或 `nginx:alpine`（前端）；
- 所有服务写 `healthcheck`；
- 生产变量从 `.env` 注入，不硬编码；
- 示例文件放 `.env.example`；真实 `.env` 入 `.gitignore`。

---

## 七、依赖、错误码、i18n、测试

本节是对其他文档的**指针**，助手不得以不读为由偏离：

- 依赖白名单：`docs/AntClaw-重构解决方案.md` 附录 A。**禁止**引入白名单外依赖。
- 错误码：附录 B.4，新增须先改 `common.proto`。
- 事件 channel：附录 B.1，新增须先改 `stream.proto`。
- i18n key 命名：附录 B.3，并见 `docs/AntClaw-国际化规范.md`。
- 测试硬规则：附录 C，并见 `docs/AntClaw-测试策略.md`。
- 精度 / 时区：附录 D。

---

## 八、提交与验收流程

### 8.1 分支策略

- 主干：`refactor/antclaw-v2`；
- 功能分支：`feat/p<阶段编号>-<简述>`，例 `feat/p4-auth-service`；
- 修复分支：`fix/<简述>`；
- 禁止直接推主干。

### 8.2 Commit 规范

- Conventional Commits：`feat|fix|chore|docs|test|refactor|perf|build|ci(scope): 描述`；
- 描述必须为中文；
- 引用任务卡编号作为 footer `Task: P<n>.<k>`。

### 8.3 PR 模板

PR 必须填写：

- 任务卡链接；
- 改动总结（中文）；
- 自测结果（命令 + 输出截图或日志）；
- 是否影响 proto / DB / 依赖；
- 回滚办法。

### 8.4 验收顺序

1. CI 全绿（构建 + lint + 测试 + `buf breaking` + 覆盖率门槛）；
2. 助手对照任务卡 **自检清单**打勾；
3. 用户人工验收；
4. 合并由**用户操作**，助手**不得自行合并**。

---

## 九、歧义处理

- 当文档互相冲突时，`AntClaw-重构解决方案.md` 优先，其次本文，其他文档再其次。
- 文档缺失：**停止实现，回报用户**。
- 技术选型有悖常识但文档要求：**执行文档**（文档是单一真相），同时在 PR 中指出质疑。

---

## 十、禁止事项清单（One-Stop Ban List）

| 类别 | 禁止 | 允许 |
|------|------|------|
| 文档 | 改 `docs/*.md` | 另开文件补说明（仅当用户要求） |
| 依赖 | 白名单外的运行时依赖 | 开发工具可讨论加入 |
| 实现方式 | `TODO` / `FIXME` / 简化版 | 完整实现或明确报告未完 |
| 密码 UI | `type="password"` 遮罩 | 明文 + 可选 Eye 切换装饰 |
| 前端计算 | 聚合/排序/过滤/金额计算 | 纯渲染 + 表单语法校验 |
| 代码脚本 | 超 800 行 | 超限前拆文件 |
| 数据库 | 裸 SELECT 无 `user_id` 过滤（非 admin） | 明确 admin 路径 + 审计 |
| 日志 | 打印 token/密码/key | 打印 `user_id / request_id / trace_id` |
| Proto | 手改 `gen/*` | 改 `.proto` 后 `buf generate` |
| Compose | 启用 `backtest-runner` / `mt-gateway` | 注释保留为后期 profile |
| Bot | 引入 Telegram / WeChat SDK | 仅保留端口接口 |
| 合并 | 助手直接 merge | 等用户人工合并 |

---

## 附录 A · 常见陷阱

1. **生成代码跑回 main**：助手常在 `gen/` 下手工修改类型。**禁止**；改 proto 再 generate。
2. **前端擅自加 axios**：任何"顺手加"的依赖一律挡回。
3. **测试被悄悄弱化**：Review 重点检查 assertion 是否被删除或从 `==` 降为 `notNil`。
4. **时间戳用 `now()` 写入 hypertable**：批量导入时可能产生乱序块；用显式 `ts` 参数。
5. **BYOK 解密后留在 struct 里**：必须 `memguard.Destroy()`，否则被 `pprof heap` 抓到。
6. **SSE 连接泄漏**：断连时要清理 Redis Streams 消费者组 offset；加 `defer g.Unregister(subID)`。
7. **Admin 与 App JWT 混用**：一律检查 `aud`，不匹配即 `AUTH_FORBIDDEN`。
8. **回测相关代码提前写实现**：按本文 P6c 严格只做骨架；越权直接回退 PR。

---

> **结语**：本文为代码实现阶段的操作圣经，任何违背都将回滚提交。助手若感到约束过紧，请用**提问**替代**先斩后奏**。
