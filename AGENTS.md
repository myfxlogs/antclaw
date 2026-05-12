# AGENTS.md — AntClaw AI Agent 工作规范

> 本项目由 AI Agent 全权开发。本文件定义铁律，Agent 每次修改代码前必须遵守。
> 违反本规范的代码不得合并。

---

## 一、文件组织原则（AI 优化）

> 本项目全部代码由 AI Agent 完成。以下规则基于 AI 的最佳理解能力制定，**不追求人类可读的行数阈值**。

### 核心原则：单一职责 > 行数

| 原则 | 说明 |
|---|---|
| **一个文件 = 一个职责** | 文件内所有代码服务于同一个概念（一个 Service / 一个页面 / 一个 proto message 组） |
| **逻辑内聚优先** | 高相关性的代码放在一起比拆散更好——AI 一次读取就能理解完整逻辑 |
| **拆分信号：职责变化** | 当文件开始包含"和文件名不直接相关的逻辑"时拆分，而不是达到某个行数 |
| **AI 上下文友好** | 单文件应能在一个 AI 对话回合中被完整读取和理解 |

### 参考软上限（非硬性，仅作警示）

| 语言 | 建议上限 | 超过时的检查项 |
|---|---|---|
| Go (`.go`) | 1000 行 | 是否承担了多个不相关的职责？是→拆分；否→保留 |
| TypeScript/TSX (`.ts`, `.tsx`) | 600 行 | 是否混合了业务逻辑+渲染+类型？是→拆为 hooks/view/model |
| Protobuf (`.proto`) | 400 行 | 是否包含了多个 Service？是→按领域拆文件 |
| Markdown (`.md`) | 600 行 | 是否涵盖了多个里程碑？是→按里程碑拆分子文档 |

> 不存在硬性行数 CI 检查。拆分决策由 AI 基于职责内聚性判断，而非机械计数。

---

## 二、模块边界（Go 后端）

```
backend/
├── cmd/                     # 可执行入口（仅 main + wire）
│   ├── antclaw-api/         # API 服务（Connect-RPC + SSE）
│   ├── antclaw-worker/      # 后台采集/分析 Worker
│   └── antclaw-cli/         # CLI 工具
├── internal/
│   ├── domain/              # 领域模型（纯结构体，零依赖）
│   ├── service/             # 业务逻辑层（接口 + 实现）
│   │   ├── signals/         # 信号计算
│   │   ├── backtest/        # 回测引擎
│   │   ├── price/           # 价格查询
│   │   └── ...
│   ├── infra/               # 基础设施
│   │   ├── postgres/        # DB 仓库（含 Repository 接口）
│   │   ├── redis/           # Redis 客户端
│   │   └── apiclient/       # 第三方 API 客户端
│   │       ├── <vendor>/    # 每家供应商一个子包
│   │       └── source.go    # 统一中间件（限流/断路器/重试）
│   ├── adapter/             # 传输层适配
│   │   ├── rpc/             # Connect-RPC Handler
│   │   ├── sse/             # SSE 推送
│   │   └── storage/         # 存储适配（postgres/redis/minio）
│   ├── auth/                # 认证（JWT/密码/限流）
│   ├── crypto/              # 加密（RSA/SecretBox/Sign）
│   ├── config/              # 配置加载
│   └── notify/              # 通知推送
```

### 铁律

- **Handler 不直接访问 DB**：`adapter/rpc/*` → `service/*` → `infra/postgres/*`
- **Service 定义接口，Infra 实现接口**：Repository 接口在 `infra/postgres/` 中定义
- **禁止循环依赖**：`domain` ← `infra` ← `service` ← `adapter`（单向依赖链）
- **每个 apiclient vendor 子包 ≤ 3 文件**：`client.go` + `types.go` + `client_test.go`

---

## 三、模块边界（TypeScript 前端）

```
frontend/<app>/src/
├── features/                # 业务功能模块
│   ├── _shared/             # 共享工具（transport/AsyncView/JsonView）
│   │   ├── transport.ts     # Connect-RPC 客户端 + JWT 注入
│   │   └── AsyncView.tsx    # 统一四态渲染
│   ├── signals/             # 信号模块
│   │   ├── SignalsPage.tsx  # 薄入口（≤ 60 行）
│   │   ├── hooks.ts         # 业务状态与副作用
│   │   ├── view.tsx         # 纯渲染
│   │   └── model.ts         # 类型定义
│   └── options/
│       ├── GEXPage.tsx
│       ├── IVSurfacePage.tsx
│       ├── api.ts           # 模块内 API 调用封装
│       └── ...
├── pages/                   # 非功能页面（Login/Settings 等）
├── components/              # 通用 UI 组件
└── hooks/                   # 全局 hooks
```

### 铁律

- **所有 API 调用必须走 `@connectrpc/connect-web`**：**禁止 `fetch()`、`axios`、`WebSocket`**
- **禁止 `setInterval` 轮询**：用 SSE `EventSource` 订阅实时数据
- **禁止合成/硬编码数据作为 fallback**：数据失败时显示错误状态，不伪造数据
- **页面组件 ≤ 60 行**：复杂页面必须拆为 `hooks.ts` + `view.tsx` + `model.ts`

---

## 四、协议层（Protobuf）

- 所有对外接口定义在 `proto/antclaw/v1/*.proto`
- 生成代码在 `gen/`（Go/TS/Kotlin/Swift/Dart），**禁止手动编辑**
- 修改 proto 后执行 `buf generate` 重新生成
- 新增 Service 必须在 `buf.yaml` 注册，在 `cmd/antclaw-api/main.go` 注册 Handler

---

## 五、代码风格

### Go
- `gofmt` + `go vet` 零告警
- 错误处理：**不吞错**，不用 `_` 忽略 error；用 `fmt.Errorf("context: %w", err)` 包装
- context 传递：所有 I/O 函数第一个参数是 `ctx context.Context`
- 不使用 `panic`（除 `init()` 和 `main()` 中的致命错误）
- 不返回合成数据：数据库无数据时返回 `ErrDataInsufficient`，**不返回假数据**

### TypeScript
- 函数组件 + hooks，不使用 class component
- 类型明确：不滥用 `any`，优先从 proto 生成的类型
- 异步状态统一用 `AsyncState<T>` 模式（idle/loading/success/error）

---

## 六、测试要求

- 每个 `service/*` 包至少 1 个 `_test.go`
- 关键路径必须有表驱动测试：`auth`、`signals`、`backtest`
- E2E 脚本在 `scripts/e2e/`，每新增 RPC 需配套 e2e case
- 冒烟测试：`bash scripts/smoke-rpc.sh`（22+ 端点 200）

---

## 七、禁止事项清单

| 禁止 | 替代方案 |
|---|---|
| 合成/随机数据（`randFloat`、硬编码 demo 等） | 数据缺失时返回明确错误 |
| REST handler（`HandleFunc` 业务路由） | Connect-RPC + SSE |
| 前端 `fetch()` / `axios` / `WebSocket` | `@connectrpc/connect-web` + `EventSource` |
| `setInterval` 轮询 | SSE 推送 |
| Go `panic`（非 init 场景） | error 返回 + 上层降级 |
| 跨层调用（Handler 直调 DB） | Handler → Service → Repository |
| 单文件超限 | 按行数上限拆分 |

---

## 八、构建与运行

```bash
# 后端编译 & 检查
cd backend && go vet ./... && go build ./...

# 全部测试
cd backend && go test ./internal/...

# 文件行数检查
bash scripts/lint-filesize.sh

# 协议合规检查
bash backend/scripts/lint-protocol.sh

# 容器构建与启动（完整栈）
docker compose -f deploy/docker-compose.yaml --project-name antclaw up -d --build api worker admin

# E2E 全量回归
bash scripts/e2e/run_all.sh
```

---

## 九、端口速查

| 服务 | 宿主机端口 | 容器内 | 协议 |
|---|---|---|---|
| antclaw-api | 8082 | 8080 | Connect-RPC + SSE |
| antclaw-admin | 8081 | 80 | HTTP → Nginx → API |
| antclaw-web | 8080 | 80 | HTTP |
| PostgreSQL | 5434 | 5432 | TCP |
| Redis | 6380 | 6379 | TCP |

---

## 十、文档约定

- 新功能完成后更新 `docs/AntClaw-功能清单.md`
- 新增数据源后更新 `docs/凭据就绪核对.md`
- 阶段验收产出 `docs/<阶段名>-完成报告.md`
- 所有文档使用中文，代码引用保持原文

---

## 十一、AI Agent 开发工作流

> 本项目全部代码由 AI Agent 完成。以下规则确保 AI 不走偏、不自行补全未定义细节。

- **开发前必读 `docs/` 下对应设计文档**。不确定的实现细节**先提问，不猜测**。
- 设计文档未定义的参数（分页条数 / 缓存 TTL / 去重窗口 / 重试间隔）**使用本文档已声明的默认值**，本文档未声明的**先问**。
- 每完成一个里程碑后，更新对应设计文档的 `[ ]` → `[x]`。
- 代码修改后必须跑 `go vet ./...` + `go test ./...` + `lint-filesize.sh` + `lint-protocol.sh` 全部通过。
- 新增 proto 后先执行 `buf generate` 再写 Handler。
- 禁止在未理解数据流的情况下写代码——先画调用链路（RPC → Service → Repository），再落键盘。
