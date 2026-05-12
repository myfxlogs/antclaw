# AntClaw 实现计划（第二批：配置热加载 · 策略管理 · 系统 AI · 采集可视化）

本目录收录 6 份**可直接照做**的实现文档，给后续 AI 助手作为编码依据。
每份文档自包含：**目标 / 背景 / 数据模型 / 接口契约 / 关键流程 / 验证步骤 / 注意事项**。
如无强关系，文档之间不强制执行顺序；同批次内部的依赖会在文档开头明确写出。

## 批次与优先级

| 序号 | 标题 | 建议批次 | 依赖 |
| --- | --- | --- | --- |
| [01](./01-配置热加载.md) | 数据源配置热加载（API → Worker 通过 Redis Pub/Sub） | 批一 | 已有加密 + `data_source_configs` |
| [02](./02-endpoint落地到collector.md) | endpoint 字段落地到各 collector（14 个数据源） | 批三 | 01 建议先完成 |
| [03](./03-master-key轮换CLI.md) | Master Key 轮换工具 `antclaw-cli rotate-master-key` | 批一 | 已有 `SecretBox` |
| [04](./04-采集内容预览.md) | 采集数据明细预览接口 + 管理端页面 | 批二 | — |
| [05](./05-回测策略管理.md) | 回测策略 CRUD + 启停（执行引擎 mock） | 批二 | — |
| [06](./06-系统AI模型配置.md) | 系统级 AI 模型配置（独立表 `system_ai_configs`） | 批一 | 已有加密栈 |

> **批一**建议优先：纯后端 + 工具层面，风险低、让后续两批的开发环境更好用。
> **批二**管理面板相关，可并行推进。
> **批三**跨越 14 个 collector，风险最高，放最后。

## 全局约束（所有文档都必须遵守）

1. 禁止为了简化牺牲可读性 / 可维护性（对应 rules.md #1）
2. 禁止为了快速牺牲代码质量 / 稳定性（对应 rules.md #2）
3. 脚本 "先分析，再实现"；能短就不写长（对应 rules.md #3）
4. 单文件 > 800 行须拆分（rules.md #4）
5. **每次修改都必须编译 + 重启容器验证**（rules.md #5）
6. 所有文档放在 `docs/`，中文命名（rules.md #6）

## 通用工程约定

### 后端
- Go 1.22+，模块 `github.com/antclaw/antclaw`
- 服务入口：`cmd/antclaw-api`、`cmd/antclaw-worker`（新 CLI 放 `cmd/antclaw-cli`）
- 分层：`handler (adapter/rpc) → service → storage (adapter/storage/postgres)`
- 表结构自检：在 `internal/adapter/storage/postgres/ensure_schema.go` 追加 `CREATE TABLE IF NOT EXISTS`
- 加密：`internal/crypto`（SecretBox / RSAManager / VerifyRequestSignature）
- 审计：调用 `audit.AuditService.Log(ctx, AuditEntry{...})`，**任何写操作都要留痕**

### 前端（管理端）
- React + Vite + TypeScript + TailwindCSS + Lucide Icons
- 路由：`frontend/admin/src/App.tsx`；侧栏：`frontend/admin/src/components/Layout.tsx`
- 加密请求：`frontend/admin/src/lib/crypto.ts` 中的 `sendSecurePut(url, payload)`
- i18n：`frontend/admin/src/locales/{zh-CN,en-US}.json`
- 新页面放 `frontend/admin/src/pages/<PageName>.tsx`

### 部署
- `deploy/docker-compose.yaml` 两个后端容器：`api`、`worker`
- nginx：`deploy/nginx.admin.conf` 需要为新 HTTP JSON 路由加 `location` 反代
- 验证入口：`http://localhost:8082`（api 直连）、`http://localhost:8081`（admin SPA via nginx）

### 登录（便于脚本联调）
- 管理员账号：`a@1.com` / `12345678`
- 登录接口：`POST /antclaw.v1.AuthService/Login`，响应里 `accessToken` 即 JWT

## 已完成里程碑（参考上下文）

- 数据源加密栈完整（Argon2id + AES-GCM 存储 / RSA-OAEP 传输 / HMAC 签名 + nonce）
- 9 个数据源种子 + 管理页 `/datasources`
- Worker 启动时按 `DB → ENV → 默认` 解析 FRED key（这次文档 01 会扩展为运行时热加载）
- 审计写入对所有数据源 PUT 生效
- 详见 `@/home/mluser/code/antclaw/docs/AntClaw-数据源配置加密方案.md`
