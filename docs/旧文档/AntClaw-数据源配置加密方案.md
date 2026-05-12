# AntClaw 数据源配置加密方案

本文档描述 Admin 后台中 **数据源 API key / 端点配置** 的加密方案，涵盖：

- **存储侧**：Argon2id 派生 + AES-256-GCM 认证加密
- **传输侧**：浏览器 → API 的 RSA-OAEP + AES-GCM 混合加密
- **请求侧**：HMAC-SHA256 签名 + 时间戳 + nonce 三重防篡改 / 防重放

按数据源粒度（每个 `source_id` 独立记录）。

> **同步（2026-04）**：公钥获取与数据源列表/更新已统一为 **Connect**（`CryptoService`、`DataSourceService`）；管理端 nginx 反代 `/antclaw.*`。下文对 **RSA hybrid + HMAC** 的说明保留为设计参考；**当前前端** `DataSources` 经 `lib/api.ts` 的 `updateDataSource` 提交明文字段，由 **后端 SecretBox** 负责落库加密（与旧「浏览器 hybrid + PUT `/admin/datasource`」路径不同）。

---

## 1. 总体架构

```
+----------+   1. Connect: CryptoService/GetCryptoPublicKey（可选，用于未来 hybrid 扩展）
|  浏览器  | <-------------------------------------|     |
| (Admin)  |                                       | API |
|          |   2. Connect: DataSourceService/UpdateDataSource（TLS + JWT；后端再 Argon2id+AES-GCM 落库）
|          | ------------------------------------> |     |    AES-256-GCM(dk)         +----+
|          |                                       |     | -- Argon2id KDF --------> | DB |
+----------+                                       +-----+    存储 (ct, salt, nonce) +----+
                                                      ^
                                                      | 3. Worker.GetSecret(source_id) → plaintext
                                                      |    （仅服务端，调用外部 API 用）
                                                      |
                                                  +--------+
                                                  | Worker |
                                                  +--------+
```

| 角色 | 算法 | 作用 |
| --- | --- | --- |
| 浏览器 → API | RSA-OAEP(SHA-256) + AES-256-GCM | 让 key 不在 HTTPS 之外的任何中间环节裸出现 |
| 浏览器 → API | HMAC-SHA256 + nonce + timestamp | 防篡改 / 防重放 |
| API → DB | Argon2id (KDF) + AES-256-GCM | 即便数据库泄露也无法直接拿到 key |
| Worker | Argon2id KDF + AES-GCM open | 内部业务可解密拿到明文 key 调用外部 API |

---

## 2. 存储侧：Argon2id + AES-256-GCM

### 2.1 表结构

```sql
CREATE TABLE data_source_configs (
    source_id        TEXT PRIMARY KEY,
    name             TEXT NOT NULL,
    kind             TEXT NOT NULL,                  -- api_key / endpoint / custom_json
    endpoint         TEXT NOT NULL DEFAULT '',
    secret_ciphertext BYTEA,
    secret_salt       BYTEA,                         -- 16 bytes
    secret_nonce      BYTEA,                         -- 12 bytes
    has_secret       BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by       TEXT NOT NULL DEFAULT ''
);
```

### 2.2 加解密流程

`backend/internal/crypto/secretbox.go`

```
master_key  := base64decode(ENV.ANTCLAW_SECRET_MASTER_KEY)   // ≥32 bytes
salt        := random(16)
nonce       := random(12)
dk          := Argon2id(master_key, salt, time=3, memory=64MiB, threads=4, keyLen=32)
ciphertext  := AES-256-GCM(dk, nonce).Seal(plaintext)
```

- **per-record salt**：彩虹表攻击不可行；
- **AES-GCM 认证**：篡改任一字段都会触发 Open 失败；
- **master key 进程内常驻**：DB dump 没有它就解不开；
- **解密透明**：业务层 `dataSourceSvc.GetSecret(ctx, sourceID)` 直接返回明文 string。

### 2.3 master key 管理

部署时通过环境变量注入：

```yaml
ANTCLAW_SECRET_MASTER_KEY: <base64, ≥32 bytes>
```

- 启动时 `NewSecretBox` 校验长度，缺失或不足直接 fatal exit；
- master key 轮换：未来若需要，可遍历表用旧 key 解密 → 用新 key 重新 Seal。本期不实现。

---

## 3. 传输侧：RSA + AES 混合加密

### 3.1 RSA 密钥对

`backend/internal/crypto/rsa.go`

- 启动时从 `ANTCLAW_RSA_KEY_PATH`（默认 `/data/rsa_private.pem`）加载，不存在则生成 2048 位密钥并写入；
- 文件权限 `0600`；docker-compose 用命名 volume `api_data` 持久化；
- `CryptoService/GetCryptoPublicKey` 返回公钥 PEM（`getCryptoPublicKeyPem()`），浏览器可用 `crypto.subtle.importKey('spki', …, RSA-OAEP/SHA-256)` 加载（供 hybrid 等扩展使用）。

### 3.2 envelope 协议

```
{
  "key_enc":   base64(RSA-OAEP(server_pub).encrypt(aes_key)),
  "iv":        base64(random 12 bytes),
  "ciphertext":base64(AES-256-GCM(aes_key, iv).Seal(JSON.stringify(payload)))
}
```

`payload`（解密后）形如：

```json
{ "endpoint": "https://api.stlouisfed.org", "secret": "your-real-key" }
{ "clear_secret": true }
```

支持 `endpoint`、`secret`、`clear_secret` 三种字段任意组合。

### 3.3 浏览器实现

- **当前**：`frontend/admin/src/pages/DataSources.tsx` → `lib/api.ts` 的 `updateDataSource()` → Connect `DataSourceService/UpdateDataSource`；密钥由后端加密存储。
- **扩展（可选）**：`lib/crypto.ts` 仍提供 `hybridEncryptJSON` / `signRequest` / `sendSecurePut`，若未来对敏感字段恢复「浏览器侧 hybrid + 签名」再接入对应 RPC。
- WebCrypto：`SubtleCrypto`，无第三方依赖。

---

## 4. 请求侧：签名 + 时间戳 + nonce

`backend/internal/crypto/sign.go`

```
session_key = SHA-256(JWT access token)
to_sign     = `${ts}\n${nonce}\n${body}`
sig_hex     = HMAC-SHA256(session_key, to_sign)
```

校验顺序（严格按顺序短路）：

1. 取 `Authorization: Bearer` 中的 token；空则 401；
2. 校验 `X-Sign-Timestamp` 在 ±5 分钟漂移内；
3. 重算 HMAC，与 `X-Signature` 比较（`hmac.Equal`，恒定时间）；
4. **Redis SETNX `sigreplay:<nonce>` TTL 10min**：第二次发同 nonce → 401 `nonce already used`；
5. 都通过后才进入 RSA 解密。

> session_key 选择 SHA-256(token) 而非 token 本身：避免在客户端代码里把 token 直接当 HMAC key（防止误用），但语义上仍然是双方共享的 session 私有密钥。

---

## 5. 路由与 nginx

### Connect RPC（当前）

| 服务 | RPC | 说明 |
| --- | --- | --- |
| `CryptoService` | `GetCryptoPublicKey` | 公钥 PEM |
| `DataSourceService` | `ListDataSources` | 列出数据源（**永不返回密文**） |
| `DataSourceService` | `UpdateDataSource` | 更新 endpoint / secret / clear_secret |

### nginx (`deploy/nginx.admin.conf`)

```nginx
location ~ ^/antclaw\. { proxy_pass http://api:8080; ... }
```

避免被 SPA 兜底 `/index.html` 吃掉 Connect 路径。

---

## 6. 前端 UI

`frontend/admin/src/pages/DataSources.tsx`

- 路由：`/datasources`，侧栏"数据源密钥"
- 展示：每行一个数据源，含 endpoint 输入框 + password 类型 secret 输入框 + 操作按钮
- 操作：
  - **保存**：将 endpoint / secret 拼成 payload，hybrid 加密 + 签名提交
  - **清除密钥**：发送 `{"clear_secret": true}`，后端置空密文字段
- 视觉：已配置密钥显示绿色 Lock 图标，未配置显示灰色 Unlock；密钥框 placeholder 提示当前状态

---

## 7. 业务集成（Worker）

Worker 启动时使用 `resolveFredKey(ctx, pool, logger)` 解析凭据，按 **DB → ENV → 内置默认** 优先级：

```go
fredKey := resolveFredKey(context.Background(), dbpool, logger)
fredClient := apiclient.NewFredClient(fredKey)
macroSvc := macro.NewMacroService(macroRepo, fredKey, logger)
```

`resolveFredKey` 行为：

1. 若 `ANTCLAW_SECRET_MASTER_KEY` 存在，构造 `SecretBox` + `datasource.Service`，调用 `GetSecret("fred")`；
2. DB 解密失败 / 未配置 → 回退 `ANTCLAW_FRED_API_KEY`；
3. 仍无 → 内置默认值（仅开发用，Warn 级别日志）。

启动日志可观察来源（三选一）：

```
FRED key resolved from data_source_configs (encrypted)
FRED key resolved from ENV
FRED key resolved to built-in default (not for production)
```

> 当前实现仅启动时解析一次；通过管理端修改 key 后需重启 worker（`docker compose restart worker`）生效。运行时热加载留给后续。

其他 12 个数据源当前都是公开 API（无 key），后续需要时按相同模式扩展即可：在 `resolve<XYZ>Key()` 中增加同样的 DB-first 解析。

## 8. 审计

数据源更新动作会在 `audit_logs` 留痕（**绝不含明文密钥**）：

```sql
SELECT action, resource, details
FROM audit_logs
WHERE action = 'datasource.update'
ORDER BY id DESC LIMIT 5;

   action       |       resource           |                    details
----------------+--------------------------+-------------------------------------------------
 datasource.update | data_source_configs/fred | {"source_id":"fred","changed":"endpoint,secret"}
```

`changed` 字段仅枚举哪些字段被改动（`endpoint` / `secret` / `clear_secret`），不包含值。

---

## 9. 端到端验证

### 9.1 列表 → 加密 PUT → 重新列表

```bash
# 登录获取 token（a@1.com / 12345678）
curl -s -XPOST -H "Content-Type: application/json" \
  -d '{"email":"a@1.com","password":"12345678"}' \
  http://localhost:8082/antclaw.v1.AuthService/Login

# 取公钥（Connect，无鉴权）
curl -s -X POST http://localhost:8082/antclaw.v1.CryptoService/GetCryptoPublicKey \
  -H "Content-Type: application/json" -d '{}' | jq -r .pem

# 列出数据源（Connect；不含明文）
curl -s -X POST http://localhost:8082/antclaw.v1.DataSourceService/ListDataSources \
  -H "Authorization: Bearer $T" -H "Content-Type: application/json" -d '{}' | jq .
```

### 9.2 自动化脚本

- `python3 /tmp/test_secure_put.py`：端到端跑一遍 hybrid 加密 + 签名 PUT，再列表
- `python3 /tmp/test_replay.py`：同 nonce 二次提交 → 应 401 `nonce already used`

实测结果：

| 验证项 | 结果 |
| --- | --- |
| `audit_logs` / `data_source_configs` / `sessions` / `refresh_tokens` 等表自动幂等创建 | ✅ |
| 数据源种子 9 条写入 | ✅ |
| `CryptoService/GetCryptoPublicKey` 返回有效 PEM | ✅ |
| 通过 Connect `DataSourceService/UpdateDataSource` 更新 FRED key → has_secret=true，密文 36B / salt 16B / nonce 12B（旧 hybrid HTTP PUT 路径已移除） | ✅ |
| 列表接口不回明文 | ✅ |
| 同 nonce 重放 → 401 `nonce already used` | ✅ |

---

### 9.3 Worker 闭环验证

```bash
# 重启 worker 让其重新从 DB 解密 key
docker compose -f deploy/docker-compose.yaml restart worker

# 观察解密来源
docker logs --since 30s antclaw-worker 2>&1 | grep "FRED key"
# → "FRED key resolved from data_source_configs (encrypted)"

# 真实 FRED 拉取
docker logs --since 30s antclaw-worker 2>&1 | grep "Macro sync"
# → "✓ Macro sync completed inserted=800"
```

实测：填入正确 key 后，Macro 一次同步插入 800 条；填入测试 key 时 8 个 series 全部 400 — 链路与解密都正确。

---

## 10. 安全要点速查

| 威胁 | 缓解措施 |
| --- | --- |
| 数据库被脱库 | master key 在 ENV，不在 DB；per-record salt + AES-GCM 认证 |
| TLS 终端被中间人观察 body | 浏览器侧 RSA-OAEP 二次加密 |
| Body 被篡改 | AES-GCM 认证 tag + HMAC 签名双重校验 |
| 抓包重放 | 时间戳 ±5min + Redis nonce SETNX 10min |
| Session token 泄露 | 仅当 token 有效期内能签名；下游仍需 token 本身做认证；登出后 token 被 jti 列表撤销 |
| key 泄露后回滚 | 前端"清除密钥"一键置空（后端不保留任何字段） |

---

## 11. 已完成里程碑

- ✅ 后端加密三件套（Argon2id+AES-GCM 存储 / RSA-OAEP 传输 / HMAC 签名 + nonce）
- ✅ `data_source_configs` 表自检 + 9 条种子；同时补齐 `sessions / refresh_tokens / password_resets / audit_logs` 启动自检（部署即可登录）
- ✅ Connect：`CryptoService/GetCryptoPublicKey`、`DataSourceService/ListDataSources` / `UpdateDataSource`；管理端 nginx 反代 `/antclaw.*`
- ✅ 浏览器 `lib/crypto.ts`（经 `getCryptoPublicKeyPem` 拉 PEM）+ `lib/api.ts` + DataSources 页面
- ✅ Worker 启动按 **DB → ENV → 默认** 解析 FRED key
- ✅ 数据源更新写审计日志（不含明文）
- ✅ 端到端 + Worker 闭环 + 重放保护实测

## 12. 后续可演进

- **运行时热加载**：API 更新数据源时通过 Redis Pub/Sub 通知 worker，避免重启；
- **更多有 key 的数据源**：按 `resolveFredKey` 的同款模式接入 CoinGecko Pro、CFTC Socrata（如启用 token）等；
- **master key 轮换 CLI**：`antclaw-cli rotate-master-key --old=<b64> --new=<b64>`，遍历表重加密；
- **endpoint 也按数据源粒度生效**：当前各 collector 的 endpoint 仍硬编码，下一步从 `data_source_configs.endpoint` 读取；
- **登录限频 / IP 黑白名单** 等接口签名层之上的纵深防御。
