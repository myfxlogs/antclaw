# AntClaw · 用户系统与鉴权

> 本文是 `AntClaw-重构解决方案.md` §四、§十一 的**可实现细化**。助手在实现 `auth`、`user`、`admin` 相关代码时，**以本文为唯一依据**；本文未覆盖部分回到用户处确认，禁止自由发挥（见《任务分解与 AI 助手约束》宪章 1、2）。

## 一、适用范围

- Proto：`proto/antclaw/v1/auth.proto`、`user.proto`、`admin.proto`。
- 后端：`backend/internal/auth/`、`backend/internal/service/user/`、`backend/internal/service/admin/`、`backend/internal/adapter/storage/postgres/`（相关查询）。
- 前端：`frontend/web/src/features/auth/`、`frontend/admin/src/features/auth/`（登录/注册/重置/会话管理）。
- 数据库：`users`、`sessions`、`refresh_tokens`、`password_resets`、`email_tokens`、`audit_logs`、`user_ai_keys`。

## 二、核心概念

- **账户主键**：`user_id`（UUIDv7），全局唯一；外键均以此为准。
- **身份证明**：`email`（小写、唯一、加 CITEXT）+ `password_hash`（Argon2id）。
- **用户名**：`username` 允许空，空时显示回退到 `email` 本地部分；全局唯一（不区分大小写）。
- **会话**：`sessions` 记录设备级活跃会话，`refresh_tokens` 持有 jti 黑白名单。
- **JWT**：Access（15 分钟）+ Refresh（30 天）。

## 三、数据模型

### 3.1 `users`

| 列 | 类型 | 约束 | 说明 |
|---|---|---|---|
| `id` | `uuid` | PK，默认 `uuidv7()` | — |
| `email` | `citext` | UNIQUE，NOT NULL | 小写；允许 `+` |
| `email_verified_at` | `timestamptz` | 可空 | 免激活时仍记录首次点击时间 |
| `username` | `citext` | UNIQUE NULLS NOT DISTINCT | 可空 |
| `display_name` | `text` | 可空 | UI 展示名，≤ 64 字符 |
| `password_hash` | `text` | NOT NULL | Argon2id 编码串，含盐与参数 |
| `password_version` | `int` | NOT NULL DEFAULT 1 | 改密后 +1，使旧 JWT 失效 |
| `role` | `text` | NOT NULL DEFAULT `'free'` | `free` / `premium` / `admin` |
| `status` | `text` | NOT NULL DEFAULT `'active'` | `active` / `banned` / `deleted` |
| `locale` | `text` | NOT NULL DEFAULT `'zh-CN'` | BCP-47 |
| `timezone` | `text` | NOT NULL DEFAULT `'Asia/Shanghai'` | IANA |
| `totp_secret_enc` | `bytea` | 可空 | AES-GCM，主密钥 `ANTCLAW_BYOK_MASTER_KEY` |
| `totp_enabled` | `boolean` | DEFAULT FALSE | 本期不投产 |
| `created_at` / `updated_at` | `timestamptz` | NOT NULL | 触发器维护 |
| `deleted_at` | `timestamptz` | 可空 | 软删（管理员操作审计） |

索引：`UNIQUE(email)`、`UNIQUE(username)`、`INDEX(status, role)`、`INDEX(created_at DESC)`。

### 3.2 `sessions`

| 列 | 类型 | 说明 |
|---|---|---|
| `id` | `uuid` | 会话 id，等同 JWT `sid` |
| `user_id` | `uuid` | FK |
| `user_agent` | `text` | — |
| `ip` | `inet` | 最近一次 IP |
| `created_at` / `last_seen_at` | `timestamptz` | — |
| `revoked_at` | `timestamptz` | 被管理员或用户吊销 |

索引：`INDEX(user_id, last_seen_at DESC)`、`PARTIAL INDEX WHERE revoked_at IS NULL`。

### 3.3 `refresh_tokens`

| 列 | 类型 | 说明 |
|---|---|---|
| `jti` | `uuid` | PK；JWT `jti` |
| `session_id` | `uuid` | FK |
| `user_id` | `uuid` | FK |
| `issued_at` | `timestamptz` | — |
| `expires_at` | `timestamptz` | 发行 + 30d |
| `revoked_at` | `timestamptz` | 可空 |
| `rotated_to` | `uuid` | 可空；轮换的下一个 jti（用于复用检测） |

Redis 层：`refresh:revoked:<jti>` → TTL = 剩余有效期；任何 Refresh 先查 Redis 黑名单，再查 PG 兜底。

### 3.4 `password_resets` / `email_tokens`

| 列 | 类型 | 说明 |
|---|---|---|
| `token_hash` | `bytea` | PK；SHA-256(token) |
| `user_id` | `uuid` | FK |
| `purpose` | `text` | `password_reset` / `email_verify` |
| `issued_at` | `timestamptz` | — |
| `expires_at` | `timestamptz` | 15 分钟后 |
| `consumed_at` | `timestamptz` | 使用后写入；同 token 不可复用 |

### 3.5 `user_ai_keys`

| 列 | 类型 | 说明 |
|---|---|---|
| `user_id` | `uuid` | FK |
| `provider` | `text` | `gemini` / `claude` |
| `key_enc` | `bytea` | AES-GCM 密文，含前缀 `v<n>:` |
| `key_fingerprint` | `text` | SHA-256(key) 前 12 位，便于展示与审计，不可逆推 |
| `last_verified_at` | `timestamptz` | 探针成功时间 |
| `last_error` | `text` | 探针最近错误信息 |
| `created_at` / `updated_at` | `timestamptz` | — |

主键：`(user_id, provider)`。

## 四、密码策略

- **算法**：Argon2id，参数默认：`memory=64MiB`、`iterations=3`、`parallelism=2`、`saltLen=16B`、`keyLen=32B`；可经环境变量覆盖。
- **编码格式**：标准 PHC 串 `$argon2id$v=19$m=65536,t=3,p=2$<salt>$<hash>`。
- **强度校验**：长度 ≥ 10；不少于两类字符（小写/大写/数字/符号）；服务端用 `github.com/trustelem/zxcvbn` 评分 ≥ 2；**拒绝**常见密码表前 10 万条。
- **改密**：改密成功后 `password_version += 1`，并吊销该用户全部 `refresh_tokens`、`sessions`（除当前发起会话）。
- **存储**：`password_hash` 不出后端；日志中必须以 `***` 占位；禁止出现在 trace / sentry 附件。
- **前端**：密码输入框**明文显示**（宪章 6），禁止 `type="password"`。

## 五、JWT

### 5.1 算法与密钥

- 算法：**EdDSA / Ed25519**；密钥对经 `ANTCLAW_JWT_PRIVATE_KEY` / `ANTCLAW_JWT_PUBLIC_KEY` 注入（PEM）。
- 支持 **双 kid 轮换**：`kid=current` 签发，`kid=previous` 仅校验，过渡期 7 天。
- JWKS 端点：`GET /.well-known/jwks.json`（仅暴露公钥，管理端不走 JWKS）。

### 5.2 Claims

```json
{
  "iss": "antclaw",
  "sub": "<user_id>",
  "aud": "antclaw-api",
  "iat": 1700000000,
  "exp": 1700000900,
  "nbf": 1700000000,
  "jti": "<uuid>",
  "sid": "<session_id>",
  "typ": "access",           // access | refresh
  "role": "free",
  "pv": 1,                    // password_version
  "locale": "zh-CN"
}
```

- `typ` 为 `access` 才允许调用业务 RPC；`refresh` 仅可调用 `AuthService.Refresh`、`Logout`。
- 校验顺序：签名 → `exp/nbf/iss/aud` → `pv` 与 DB `users.password_version` 一致 → `sid` 未在 `sessions.revoked_at` → Redis 黑名单未命中。

### 5.3 传输

- Web：Access 放 **HttpOnly + Secure + SameSite=Lax** Cookie（`antclaw_at`）；Refresh 同样 Cookie（`antclaw_rt`，`Path=/api/auth`）。
- 移动端与服务端：`Authorization: Bearer <token>`。
- **CSRF**：Cookie 模式下 double-submit；前端在 header 附 `X-AntClaw-CSRF: <random>`，服务端比对 `antclaw_csrf` cookie 一致。

## 六、AuthService 契约（节选）

> 完整字段以 `auth.proto` 为准；本节约束行为、错误码与幂等要求。

### 6.1 `Register`

- 输入：`email`、`password`、可选 `display_name`、`locale`、`timezone`、`client` 元信息。
- 流程：邮箱规范化 → 强度校验 → Argon2id 哈希 → 插入 `users`（并发冲突返回 `ALREADY_EXISTS`）→ 发送「注册成功通知邮件」（异步，失败不阻塞）→ 签发 access + refresh。
- 错误码：`INVALID_EMAIL`、`WEAK_PASSWORD`、`EMAIL_TAKEN`、`RATE_LIMITED`。
- 幂等：接受 `idempotency_key`（24h TTL）；重复相同邮箱 + idem key 直接返回首个结果。

### 6.2 `Login`

- 输入：`email` / `username`、`password`、`client`。
- 风控：同 email 连续失败 5 次后 Redis `login:fail:<email>` 退避 `[1s, 2s, 4s, 8s, 16s, 30s]`；失败响应必须 **恒定延迟** 200 ms，防止 timing 攻击。
- 成功：写 `sessions`、`refresh_tokens`；异地登录（IP 段变化 + UA 指纹变化）发送安全通知邮件。
- 错误码：`INVALID_CREDENTIALS`、`ACCOUNT_BANNED`、`RATE_LIMITED`。

### 6.3 `Refresh`

- 输入：Refresh JWT（Cookie 或 Header）。
- 流程：校验签名与 `typ=refresh` → 查 `refresh_tokens` 未吊销 → **立即吊销旧 jti 并生成新 jti**（Refresh Rotation）→ 返回新 access + refresh。
- **复用检测**：若 `rotated_to` 已存在但当前 jti 再次被使用 → 吊销整条会话（所有 refresh）、强制重登。
- 错误码：`REFRESH_EXPIRED`、`REFRESH_REUSED`、`SESSION_REVOKED`。

### 6.4 `Logout`

- 吊销当前 `session_id` 与 `refresh_tokens.jti`；Cookie 返回 `Max-Age=0`。

### 6.5 `RequestPasswordReset` / `ResetPassword`

- `RequestPasswordReset`：无论邮箱是否存在，响应 **恒定**；存在则发送 15min 有效 token 邮件，写 `password_resets`。
- `ResetPassword`：token 单次消费；成功后按 §四·改密流程吊销其他会话。

### 6.6 `VerifyEmail`（免激活期可选）

- 保留端点与表结构；注册后同时生成一条 `purpose=email_verify`；用户点击链接写 `email_verified_at`；不阻塞业务访问。

## 七、RBAC 与配额

### 7.1 角色

| 角色 | 可访问 | 备注 |
|---|---|---|
| `free` | 业务 RPC 子集 + 仅 USD/高影响告警 | 默认 |
| `premium` | 业务 RPC 全集 + 全告警 | 由 admin 设置 |
| `admin` | 全部 + `AdminService` | 由 DB 初始化或 admin 委任 |

### 7.2 配额（Redis 令牌桶）

| 维度 | free | premium | admin |
|---|---|---|---|
| AI 调用 RPM | 10 | 60 | 600 |
| AI 调用日额 | 50 | 1000 | 不限 |
| 告警订阅上限 | 10 | 200 | 不限 |
| SSE 并发连接 | 2 | 5 | 10 |

- 实现：`pkg/ratelimit/token_bucket.lua`；key 形如 `rl:<bucket>:<user_id>`。
- 超限响应：`RESOURCE_EXHAUSTED`，`retry_after_ms` 在 `ErrorDetail` 中返回。

## 八、会话管理（用户可见）

- `UserService.ListSessions`：列出本人全部活跃会话（UA、IP、最后活跃）。
- `UserService.RevokeSession(session_id)`：吊销指定会话。
- `UserService.RevokeAllOtherSessions`：保留当前会话，吊销其余。

## 九、管理员操作（AdminService）

- `ListUsers(filter, pagination)`：支持 `email`、`role`、`status`、`created_at` 过滤。
- `SetRole(user_id, role)`：写审计。
- `Ban(user_id, reason)` / `Unban(user_id)`：`status` 切换；Ban 时立刻吊销全部会话。
- `ForceLogout(user_id)`：只吊销会话，不改 status。
- `ResetPassword(user_id)`：重置为一次性随机密码并发邮件，同时 `password_version += 1`。
- `ListAuditLogs(filter)`：只读；**UPDATE / DELETE 被 PG 拒绝**（见 §十一）。

## 十、AI BYOK（严格隔离）

- 入参：`UserService.SetAiKey(provider, api_key)`；服务端立即以 AES-GCM + 主密钥加密，**绝不明文落库或落日志**。
- 读取：调用 AI 前在 **请求上下文中取当前 `user_id`**，据此读取 `user_ai_keys`；禁止任何跨用户缓存。
- 失败路径：`key_missing`、`key_invalid` 错误码；前端跳转设置页。**禁止**回落到平台密钥。
- 探针：worker 每 24h 调用 `models.list` 最轻量探针；仅通知用户本人；**不写持久日志**（见 §十二）。
- 轮换：`ANTCLAW_BYOK_MASTER_KEY` 支持 `v1,v2,...` 多版本；密文前缀 `v<n>:`；后台任务逐步解密+重加密写回。

## 十一、审计

- 所有写操作（`Register`、`Login` 失败/成功、`Refresh`、`Logout`、`ResetPassword`、`SetRole`、`Ban` 等）写 `audit_logs`。
- **Append-only 硬约束**（由 PG 实现，禁止业务代码绕过）：
  - 触发器 `audit_logs_no_update` 对 `UPDATE / DELETE` 抛异常。
  - 每条记录包含 `hash_prev` + `hash_self = sha256(prev || payload)`，形成哈希链。
  - 同步双写 MinIO `audit-worm` bucket（Object Lock = compliance 模式）；双写失败计入 Sentry + 告警，但**不回滚业务事务**（审计降级通道由 P6b 兜底重试）。
- 保留：**永久**；管理员可在 Admin 控制台查询、导出，不可删除。

## 十二、日志与隐私

- 不得记录：明文密码、`password_hash`、JWT 全文、BYOK 密钥、email token 明文。
- 可记录：`user_id`、`session_id`、`jti`（UUID）、`ip` 前 3 段（IPv4）或 `/64`（IPv6）、UA 哈希。
- BYOK 探针结果只写 Redis `byok:health:<user_id>:<provider>`（TTL 7d），**不入审计表**（隐私）。

## 十三、前端集成要点

- `frontend/web/src/api/auth.ts`：封装 `register/login/refresh/logout`，统一处理 401 → 自动 `Refresh` → 重试；连续两次 401 触发跳转登录。
- **密码输入框**：项目级 ESLint 规则禁止 `<input type="password">`；统一使用 `<PasswordInput>` 组件（明文展示）。
- **Token 存储**：Web 仅 Cookie；禁止 `localStorage`/`sessionStorage` 存 JWT。
- **CSRF**：`axios` 拦截器自动写入 `X-AntClaw-CSRF`。

## 十四、验收清单（对照任务卡 P4）

- [ ] `AuthService` 全方法契约测试通过（含错误码矩阵）。
- [ ] Argon2id 参数可环境变量覆盖，默认值与本文一致。
- [ ] JWT `kid` 轮换演练成功（两把公钥都可验签）。
- [ ] Refresh rotation + 复用检测可通过 e2e 测试复现。
- [ ] `audit_logs` 的 `UPDATE / DELETE` 被 PG 拒绝（集成测试验证）。
- [ ] 审计哈希链完整性校验脚本 `scripts/audit_verify.go` 通过。
- [ ] BYOK 密钥跨用户泄漏测试：用户 A 的请求能否读到用户 B 密钥 → 必须失败。
- [ ] Login 恒定延迟（成功/失败响应时间差 < 5 ms）。
- [ ] 密码策略 zxcvbn ≥ 2 不过关（含常用密码库）单测通过。

## 十五、已决事项（2026-04-24）

- **U1 · 异地登录判定**：按 **IP 国家级**触发安全通知邮件；子网级不启用。
- **U2 · `premium` 订阅入口**：MVP 仅展示当前层级；**升级按钮置灰**；支付方式本期不集成。
- **U3 · 管理员账户 2FA**：**不启用** 2FA；`users.totp_*` 字段保留但管理员不强制；全局开关预留禁用。

> 对应 §4.4 / §六 的行为以本节决议为准；实现时禁止再自行放宽。
