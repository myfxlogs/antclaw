# AntClaw · 管理员控制台（Admin Console）

> 对应 `frontend/admin/` + `backend/internal/service/admin/` + `AdminService`。本文是重构方案 §五 的落地细则。管理端默认 locale `zh-CN`（见《国际化规范》§一）。

## 一、定位

- 面向 AntClaw 内部运营/运维，**独立部署**在同域 `/admin` 路径。
- 覆盖：用户 / 权限与配额 / 通知 / 任务调度 / 数据源健康 / 审计 / 数据清理 / 对象存储 / i18n / 反馈 / 系统指标。
- 仅 `role=admin` 可登录；普通用户访问 `/admin` 返回 403。

## 二、技术栈

- 与 `frontend/web` 同栈（React 18 + Vite + TS + TanStack Router/Query + shadcn/ui + Tailwind）。
- 独立 entry、独立产物目录 `frontend/admin/dist/`、独立 Sentry 项目。
- **共享 `packages/rpc-client`、`packages/ui`、`packages/i18n`**；不得复制一份。

## 三、功能模块

### 3.1 用户管理 `/admin/users`

- 列表：`email`、`role`、`status`、`created_at`、`last_login_at`、`locale`；支持模糊搜索、角色/状态过滤、分页。
- 操作：
  - `SetRole(user_id, role)`
  - `Ban(user_id, reason)` / `Unban(user_id)`
  - `ForceLogout(user_id)`（吊销所有会话）
  - `ResetPassword(user_id)`（后端生成一次性密码并邮件发送；返回值**仅提示已发**，不回显明文）
  - 查看该用户：会话清单、审计、订阅、BYOK 指纹（非密文）、AI 用量。
- 所有写操作触发二次确认 + 必填 `reason` 字段 → 写审计。

### 3.2 权限与配额 `/admin/quota`

- 角色模板：`free` / `premium` 的默认配额值（Redis 令牌桶策略、AI RPM/日额、告警订阅上限、SSE 并发上限）。
- 单用户覆盖：`user_quota_overrides(user_id, key, value, expires_at)`。
- 本页**不包含**「开启平台共享 AI 密钥」开关（BYOK 硬约束）。
- 2FA 全局开关：预留 UI，本期禁用（只读展示 + 提示「后期提供」）。

### 3.3 通知中心 `/admin/notify`

- 站内信模板 CRUD（`inapp_templates`，字段：`key`、`i18n 多语种 body`、`level`）。
- 移动推送配置：FCM / APNs / HMS 凭据下拉选择（密文不回显，仅允许替换）。
- 邮件：SMTP 配置 + 模板（`password_reset`、`security_notice` 等）。
- 广播：向指定角色/标签发送一次性公告 → 走 `CHANNEL_SYSTEM_NOTICE`。

### 3.4 任务与调度 `/admin/jobs`

- 列出 `backend/internal/scheduler` 注册任务：名称、cron、最近运行、下一次运行、平均耗时、失败率。
- 手动触发：`AdminService.RunJob(name, args_json)`（写审计）。
- 运行历史：`job_runs` 表，可查看日志尾部与产出结果 URI。
- 取消：仅允许取消「支持 context cancel」的任务（由注册元数据标记）。

### 3.5 数据源配置 `/datasources`（前端路由；历史文档曾写 `/admin/datasources`）

- 外部数据源列表（CFTC、FRED、ECB、SNB、OECD、Eurostat、BIS、TradingEconomics、DTCC、SEC13F、TreasuryAuctions、FedWatch、WorldBank、IMF、价格源…）。
- 每源展示：最近拉取时间、错误率（1h/24h）、降级标志、限流状态。
- 人工降级开关：置「降级」后 worker 跳过本源并走缓存/上一次成功结果；写审计。

### 3.6 审计日志 `/admin/audit`

- 筛选：`actor`、`action`、`target`、`time range`、`severity`。
- 只读：页面不提供删除/编辑；后端 UPDATE/DELETE 被 PG 拒绝。
- 导出：下载 CSV / JSONL（管理员操作本身写审计）。
- 哈希链校验：按钮「校验最近 7 天链完整性」→ 调 `scripts/audit_verify.go`（后台任务），结果在页面显示。

### 3.7 数据清理 `/admin/cleanup`

- 业务域：COT / Calendar / Macro / Price / Signals / Backtest / AI Usage…
- 按类型 + 时间范围**手动**清理；禁止定时清理（决策 #16 保留永久）。
- 双重确认 + 必填 `reason` + 干运行预估行数 → 确认后执行并写审计。

### 3.8 对象存储 `/admin/objects`

- 浏览 MinIO（或 R2）bucket：`backtest-exports`、`stream-snapshots`、`audit-worm`、`backups`。
- `audit-worm` bucket 只读（Object Lock），禁用删除按钮。
- 文件预览：文本 / JSON / 图片内联；其他类型仅下载。
- 预签名 URL TTL 最大 1h。

### 3.9 i18n 资源 `/admin/i18n`

- `i18n_strings` 表的 CRUD 界面（见《国际化规范》§六）。
- 缺失键看板：两语种 key 集合 diff。
- 导入/导出 JSON（按 locale）。
- 任意编辑 → 写审计 → 调 `POST /admin/i18n/reload` 触发后端内存刷新 → 前端热拉取新资源。

### 3.10 反馈工单 `/admin/feedback`

- 列出 `UserService.SubmitFeedback` 的记录；状态：`open` / `in_review` / `closed`。
- 管理员可添加内部备注（不回显给用户）。
- 无外部工单系统对接（后期再说）。

### 3.11 系统指标 `/admin/metrics`

- 内嵌 Grafana（同域 iframe，登录态由 Caddy 反代 + 管理端 Cookie 授权）。
- 预置仪表板：API 延迟、SSE 连接、Redis 队列深度、PG 慢查询、BYOK 失败率、AI Token 用量。
- Jaeger/Tempo 链路：从错误列表点击 `traceId` 跳转。

## 四、API 契约（`AdminService`）

> 完整字段以 `proto/antclaw/v1/admin.proto` 为准；本节只列出方法与行为约束。

| 方法 | 说明 | 幂等 | 审计 |
|---|---|---|---|
| `ListUsers` | 分页 + 过滤 | 是 | 否 |
| `GetUser` | 含会话/订阅/用量摘要 | 是 | 否 |
| `SetRole` | 角色变更 | 按 `(user_id, role)` | **是** |
| `Ban` / `Unban` | 必填 `reason` | 否 | **是** |
| `ForceLogout` | 吊销所有会话 | 否 | **是** |
| `ResetPassword` | 后端重置 + 邮件 | 否 | **是** |
| `ListJobs` / `GetJob` | — | 是 | 否 |
| `RunJob` | 手动触发 | 按 `idempotency_key` | **是** |
| `ListAuditLogs` | 只读分页 | 是 | 否 |
| `ListWebhookDeliveries` | 只读 | 是 | 否 |
| `UpdateDatasourceHealth` | 降级开关 | 按 `(source, state)` | **是** |
| `CleanupData` | 按域 + 时间范围 | 否 | **是**（含预估行数与实际行数） |
| `UpdateI18nString` / `ReloadI18n` | i18n 编辑与重载 | 否 | **是** |
| `ListFeedback` / `UpdateFeedback` | 反馈管理 | 否 | 仅状态变更审计 |

所有写方法必须：

1. `ctx.user.role == admin` 硬校验。
2. 必填 `reason`（字符串，≥ 4 字）。
3. 写 `audit_logs`，`action = admin.<方法名小写>`。

## 五、前端路由与权限

```
/admin/login                 公开
/admin/                      仪表板（admin）
/admin/users                 admin
/admin/users/:id             admin
/admin/quota                 admin
/admin/notify                admin
/admin/jobs                  admin
/admin/datasources（或根挂载下的 `/datasources`） admin
/admin/audit                 admin
/admin/cleanup               admin
/admin/objects               admin
/admin/i18n                  admin
/admin/feedback              admin
/admin/metrics               admin（iframe Grafana）
```

- 路由守卫：`users/me` 查询失败或 role != admin → 跳 `/admin/login`，非 admin 角色显示 403。
- 会话独立：Admin 端会话 Cookie 名 `antclaw_admin_at`；**不与用户端共用**，即使同一浏览器同一账号也要二次登录。

## 六、UI 约束

- 所有破坏性操作（ban、reset、cleanup、降级）使用 `packages/ui/ConfirmDialog`，强制二次输入（如输入用户 email 或关键字确认）。
- 列表统一使用 `packages/ui/DataTable`，支持 URL 同步的过滤 + 分页 + 排序。
- 语言切换：Admin 独立 `admin_locale` cookie（默认 `zh-CN`），与用户端解耦。
- 时区：管理员资料 timezone；时间列悬浮显示 UTC 原值。

## 七、安全

- Admin 端所有写操作返回最小必要信息；禁止在响应中回显密码、token、BYOK 密文。
- 管理员 2FA：本期不强制，决策 #12 延后；`users` 表已预留字段。
- Admin 前端禁用 Sentry `replay`（隐私）。
- CSRF double-submit 同用户端，但 header 名为 `X-AntClaw-Admin-CSRF`。

## 八、验收清单（对照任务卡 P11）

- [ ] Admin 独立产物 + 独立会话 cookie。
- [ ] 所有写操作必填 `reason` 且写审计（契约测试覆盖）。
- [ ] `audit-worm` bucket UI 只读；删除按钮不存在。
- [ ] 数据清理页的预估与实际行数记录进审计。
- [ ] Grafana iframe 在非 admin 会话下 401。
- [ ] Playwright：登录 → 改角色 → 封禁 → 强制登出 → 查看审计链路全绿。

## 九、已决事项（2026-04-24）

- **A1 · Admin access TTL**：Admin 端 access JWT **5 分钟**；用户端保持 **15 分钟**；Refresh 两端同为 30 天。Admin 前端自动在过期前 60s 静默刷新；刷新失败跳 `/admin/login`。
- **A2 · admin 查看明文 BYOK**：**永远不允许**；Admin UI 仅展示 `key_fingerprint` 与健康状态；服务端无任何 RPC/CLI 可解出用户明文 BYOK。这是硬安全边界（宪章 3），任何后续需求需用户书面变更本条。
