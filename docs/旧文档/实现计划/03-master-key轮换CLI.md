# 03. Master Key 轮换 CLI

## 目标

在**不丢失任何密钥明文**的前提下，把 `data_source_configs` 和 `system_ai_configs`（见 06 文档）里所有 secret 从旧 master key 重加密为新 master key。

支持：

- `--dry-run`：只扫描，不写
- `--old <base64>` / `--new <base64>`：显式指定两把 master key
- 分批事务：失败可续跑
- 输出简明报告：成功 / 失败 / 跳过计数

## 背景

- master key 由 `ANTCLAW_SECRET_MASTER_KEY` 注入；
- 存储格式：`(ciphertext, salt, nonce)`；解密 = `SecretBox.Open(ct, salt, nonce)`；
- 轮换 = 用旧 key 拿明文 → 用新 key 重新 Seal → update 行。

## 设计决策

| 决策点 | 选择 |
| --- | --- |
| 程序形态 | 独立可执行 `cmd/antclaw-cli`（子命令模式） |
| 数据库连接 | 复用相同 `ANTCLAW_DB_*` 环境变量 |
| 事务粒度 | 每行一个事务（失败只影响当前行） |
| 幂等 | 记录"已轮换"通过行的 `updated_at` 变化；工具本身不加状态列 |
| 并发 | 串行（避免事务冲突，每行耗时 <1ms，总量 <100 行） |
| 运行环境 | 一次性容器：`docker compose run --rm api antclaw-cli rotate-master-key --old=$OLD --new=$NEW` |

## CLI 接口

```
antclaw-cli rotate-master-key \
  --old  <base64-32B>   # 必填
  --new  <base64-32B>   # 必填
  --dry-run             # 可选，仅扫描
  --tables data_source_configs,system_ai_configs   # 可选，默认两张都处理

Output (human-readable):
  [dry-run] would rotate 7 rows in data_source_configs
  [dry-run] would rotate 4 rows in system_ai_configs
  OK  fred
  OK  coingecko
  SKIP cftc_socrata (has_secret=false)
  OK  openai
  FAIL deepseek: decrypt failed: cipher: message authentication failed
  ---
  Total: 10 OK, 1 FAIL, 1 SKIP, 0 missing
```

## 数据模型

无新表。

## 关键流程

### 伪代码（核心）

```go
func rotate(ctx context.Context, pool *pgxpool.Pool, oldBox, newBox *crypto.SecretBox, table string, dryRun bool) (stats, error) {
    rows, err := pool.Query(ctx,
        `SELECT id_col, secret_ciphertext, secret_salt, secret_nonce FROM `+table+`
         WHERE has_secret = TRUE`)
    for rows.Next() {
        var id string; var ct, salt, nonce []byte
        rows.Scan(&id, &ct, &salt, &nonce)

        plaintext, err := oldBox.Open(ct, salt, nonce)
        if err != nil { stats.fail++; log("FAIL", id, err); continue }

        if dryRun { stats.wouldRotate++; continue }

        newCT, newSalt, newNonce, err := newBox.Seal(plaintext)
        if err != nil { stats.fail++; continue }

        _, err = pool.Exec(ctx,
            `UPDATE `+table+` SET secret_ciphertext=$1, secret_salt=$2, secret_nonce=$3, updated_at=NOW() WHERE id_col=$4`,
            newCT, newSalt, newNonce, id)
        if err != nil { stats.fail++; continue }
        stats.ok++
    }
    return stats, nil
}
```

> `id_col` 在 `data_source_configs` 是 `source_id`，在 `system_ai_configs` 是 `provider_id`（见 06 号文档）。做成 switch / 配置表。

### 轮换后端核心

```go
type RotatePlan struct {
    OldMaster, NewMaster string
    Tables               []string
    DryRun               bool
}

func Run(ctx context.Context, plan RotatePlan, pool *pgxpool.Pool, logger *slog.Logger) (Report, error)
```

### Report 结构

```go
type Report struct {
    PerTable map[string]TableStat
    OK, Fail, Skip, WouldRotate int
}
type TableStat struct { OK, Fail, Skip int; FailedRows []string }
```

## 修改清单

| 文件 | 变更 |
| --- | --- |
| `backend/cmd/antclaw-cli/main.go`（新） | 子命令 dispatch；只处理参数解析与最终退出码 |
| `backend/cmd/antclaw-cli/rotate_master_key.go`（新） | 核心轮换逻辑（本文档的伪代码） |
| `backend/internal/crypto/secretbox.go` | 若尚未有 `func NewSecretBoxFromRawB64(s string)` 可独立，方便 CLI 构造 |
| `deploy/Dockerfile.backend` | 增加 `go build -o ../bin/antclaw-cli ./cmd/antclaw-cli` |
| `docs/AntClaw-数据源配置加密方案.md` | §7 或单独小节链接到本文档 |
| `docs/AntClaw-部署指南.md` | 增补"master key 轮换操作流程" |

## 运维流程（运行手册）

```bash
# 0. 生成新 master key（32B base64）
NEW_KEY=$(openssl rand -base64 32)
OLD_KEY=$(grep ANTCLAW_SECRET_MASTER_KEY deploy/docker-compose.yaml | head -1 | awk -F: '{print $2}' | tr -d ' ')

# 1. 先 dry-run
docker compose -f deploy/docker-compose.yaml run --rm api \
  antclaw-cli rotate-master-key --old="$OLD_KEY" --new="$NEW_KEY" --dry-run

# 2. 确认 dry-run 的 "would rotate N rows" 符合预期，执行真实轮换
docker compose -f deploy/docker-compose.yaml run --rm api \
  antclaw-cli rotate-master-key --old="$OLD_KEY" --new="$NEW_KEY"

# 3. 切换 api/worker 环境变量到新 key
sed -i "s#ANTCLAW_SECRET_MASTER_KEY:.*#ANTCLAW_SECRET_MASTER_KEY: $NEW_KEY#" deploy/docker-compose.yaml
docker compose -f deploy/docker-compose.yaml up -d api worker

# 4. 验证
docker logs antclaw-worker 2>&1 | grep "key resolved"   # 应仍为 data_source_configs (encrypted)
```

## 验证步骤

### 自动化单测（建议）

`backend/cmd/antclaw-cli/rotate_master_key_test.go`：

1. 起内存 PostgreSQL 或用 testcontainers
2. 用 oldBox Seal 插入 3 行
3. 运行 Rotate(plan, dryRun=true)：OK=0 WouldRotate=3
4. 运行 Rotate(plan, dryRun=false)：OK=3
5. 再用 newBox Open 每一行 → 得到原明文

### 集成验证

```bash
# 故意用错 old key → 全部 FAIL，不应改 DB
antclaw-cli rotate-master-key --old=bogus-b64 --new=$NEW_KEY --dry-run
# 期望：FAIL count == 行数
```

## 注意事项

- **备份先行**：任何轮换前运维侧先 `pg_dump data_source_configs system_ai_configs`
- **不要在 web 请求触发**：纯 CLI；避免 HTTP 接口被误调
- **日志脱敏**：绝不打印 plaintext；失败行只打印 id_col + 错误类型
- **返回码**：`Fail>0` 时退出码非 0，便于 CI/脚本检测
- **同 key 轮换检测**：如果 old == new，直接拒绝并提示"nothing to do"
- **部分失败策略**：单行失败不阻塞其余；最终退出码非 0，报告清晰列出失败行

## 风险与回退

- 风险：运行中 api/worker 仍持旧 key，DB 已轮换 → 解密失败
  - 缓解：运行手册规定"轮换完成后 5 秒内切环境变量并重启"；或做成滚动（两 key 兼容窗）
- 风险：中途断电，一半轮换一半没轮 → 下次重跑，old key 解失败的行跳过（那些已是 new key）
  - 改进：加 `--tolerate-old-fails` 开关，首次 FAIL 的行用 new key 再试一次 Open，通过则 SKIP
