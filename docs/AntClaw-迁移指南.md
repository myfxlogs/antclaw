# AntClaw · 迁移指南（BadgerDB → PostgreSQL + MinIO）

> 本文是重构方案 §7.4 的实现细则，对应任务卡 **P4a**。目标：把旧 ARK Intelligent 的 BadgerDB 数据一次性导入 AntClaw 的 PG + MinIO，**幂等**、**可恢复**、**可审计**。

## 一、范围

**in scope**：

- 旧 Badger 中的用户相关数据（如果存在，本期从邮箱+密码体系新建，详见 §六 用户策略）。
- 业务领域数据（可映射到新库的部分）：COT 快照、Calendar 事件、Macro 指标快照、Price 历史、Signals 历史、回测历史、反馈。
- 附件文件（若原来存在）：截图、导出文件 → MinIO。

**out of scope**：

- 原 Telegram Bot 绑定关系（Bot 本期不交付，绑定表空置；旧绑定**不迁移**，由用户上线后重新 `/bind`）。
- 原运行期缓存（Redis / 内存缓存）：不迁移。
- 日志、调试数据：不迁移（决策 #16）。

## 二、工具

- `backend/cmd/antclaw-migrate`：一个独立 Go CLI；**首次部署必跑**。
- 子命令：

```
antclaw-migrate init          # 仅执行 db migrations + 初始化对象存储 bucket（audit-worm + Object Lock）
antclaw-migrate import        # 从 Badger 导入业务数据
antclaw-migrate verify        # 对比源/目标计数 + 采样校验
antclaw-migrate bootstrap-admin  # 创建初始 admin 用户
antclaw-migrate rotate-byok   # 轮换主密钥（运维用）
```

- 入参：

```
--source-badger <path>        # 旧 Badger 数据目录
--domain <name>               # 仅迁移指定域（cot|calendar|macro|price|signals|backtest|feedback|files|all）
--batch-size 500              # 每批行数
--resume                      # 断点续跑（默认开启）
--dry-run                     # 不写目标库
--concurrency 4               # 并发 worker 数
```

## 三、幂等与断点

### 3.1 迁移游标表

```sql
CREATE TABLE migration_cursors (
  domain       text PRIMARY KEY,
  last_key     bytea NOT NULL,          -- Badger key
  rows_done    bigint NOT NULL DEFAULT 0,
  finished_at  timestamptz,
  updated_at   timestamptz NOT NULL DEFAULT now()
);
```

- CLI 每处理一批 → 事务内 `INSERT ... ON CONFLICT UPDATE` 游标；目标表写入与游标更新在同一事务。
- 重跑时按 `last_key` 继续；完成后写 `finished_at`。
- 支持 `--force-restart <domain>`：清游标并重扫（仅 dry-run 默认允许；生产需 admin 确认）。

### 3.2 目标表去重

- 所有目标表都具备 **业务自然键** 唯一约束（例：`cot_snapshots(source, symbol, report_date)`）。
- 插入语句统一使用 `INSERT ... ON CONFLICT (natural_key) DO NOTHING` 或 `DO UPDATE SET ...`（按域策略）。
- **禁止**依赖自增 id 去重。

### 3.3 失败处理

- 任一批次失败 → 事务回滚 → 记录错误到 `migration_errors(domain, key, error, occurred_at)` → 继续下一批（fail-soft）。
- CLI 退出码：0 完全成功；10 有错误行（需人工审阅）；20 致命错误（游标不变）。

## 四、域迁移映射

### 4.1 COT → `cot_snapshots` / `cot_signals`

- Badger key 模式：`cot:<source>:<symbol>:<report_date>`。
- 源记录包含：持仓明细 JSON。
- 目标：
  - 持仓明细进 `cot_snapshots`（`source`、`symbol`、`report_date`、`payload JSONB`）。
  - 派生信号重算后入 `cot_signals`（**迁移期不重算历史信号**，仅导入持仓；信号由后续调度任务从头跑）。

### 4.2 Calendar → `calendar_events` / `calendar_impacts`

- Badger key：`cal:<yyyymmdd>:<event_id>`。
- 原始结构直接映射；`title_i18n` 字段：迁移时只填 `{ "zh-CN": <原标题> }`；英文字段后由 worker 批量补译（可选）。

### 4.3 Macro → `macro_series` / `macro_points`

- 按系列分：`macro:<provider>:<series_id>` → 一条 `macro_series`；
- 数据点 `macro:<provider>:<series_id>:<date>` → `macro_points(series_id, ts, value)`，写 TimescaleDB hypertable。

### 4.4 Price → `price_bars` / `price_sessions`

- Price tick 本期**不迁移历史 tick**（数据量大、价值低）；仅迁移日/小时 bar。
- 目标 hypertable `price_bars(symbol, ts, tf, o, h, l, c, v)`，按 `(symbol, ts)` 分块。
- session/制度历史 → `price_sessions` / `price_regimes`。

### 4.5 Signals / TA / Vol / Sentiment → 各自快照表

- 仅导入原始快照；派生指标在 AntClaw 上线后以调度任务重算填补。

### 4.6 Backtest → `backtests`

- 原回测结果元数据 → `backtests(id, owner_user_id=NULL, params_json, metrics_json, created_at)`；
- `owner_user_id` 置空（旧体系无映射用户）；迁移后对用户可见性为「历史基准」，不可编辑删除。
- 附件（如 equity curve CSV）→ MinIO `backtest-exports/legacy/<id>/...`；`result_uri` 字段写入。

### 4.7 Feedback → `feedback`

- 直接复制；`user_id` 置空；状态初值 `closed`（历史留痕）。

### 4.8 Files → MinIO

- 原附件路径映射到 MinIO key；按 `legacy/<domain>/<original_relpath>` 前缀存放。
- 写入前计算 SHA-256；若目标 key 已存在且 hash 一致则跳过；不一致则 key 后追加 `.dup.<n>` 并记 `migration_errors`。

## 五、用户与鉴权策略

**本期不迁移旧用户**：

- 旧系统以 Telegram Bot 为入口，没有邮箱+密码概念；
- 新用户体系邮箱+Argon2id 从零开始注册；
- 若历史上有邮箱白名单，可由运维在 `antclaw-migrate bootstrap-admin` 之后，通过 Admin 控制台邀请（发一次性重置 token）。

**初始 admin**：

- `antclaw-migrate bootstrap-admin` 交互式 prompt 或通过环境变量：
  - `ADMIN_EMAIL`
  - `ADMIN_INITIAL_PASSWORD`（强度 ≥ §用户系统 §四）
- 创建 `users(role=admin, status=active)`；写入审计首条 `action=system.bootstrap_admin`。

## 六、审计链初始化

- 空库首次 `init` 时，`audit_logs` 写入一条「创世记录」：`action=system.genesis`，`hash_prev=NULL`，`hash_self=sha256(payload)`。
- 之后每条审计均续链；`antclaw-migrate verify --audit-chain` 用于抽样验证。

## 七、执行顺序（首次部署）

```
1. docker compose up -d postgres redis minio
2. antclaw-migrate init
      - goose up
      - 创建 MinIO buckets：backtest-exports / stream-snapshots / backups / audit-worm(Object Lock)
      - audit_logs 写创世
3. antclaw-migrate bootstrap-admin
4. antclaw-migrate import --source-badger /path --domain all
      - 逐域跑，使用游标
5. antclaw-migrate verify
      - 源/目标行数对比
      - 自然键抽样（每域 100 条）
      - 审计链完整性
6. docker compose up -d
```

- 步骤 4 可中断恢复；允许在线重跑（不影响已运行的 AntClaw 服务）。
- 步骤 4 完成前，AntClaw 前端可先上线，但后端某些接口会返回空结果（由任务调度后续填充）。

## 八、验证矩阵

| 项 | 手段 | 通过条件 |
|---|---|---|
| 行数一致（可统计域） | `verify --counts` | 源/目标差异 ≤ 0.01% |
| 自然键采样 | 每域 100 条 PK 反查 | 100% 命中 |
| TimescaleDB 分块 | `SELECT show_chunks(...)` | 每 hypertable 至少 1 chunk |
| 附件哈希 | MinIO key 全量拉 + sha256 | 与源 Badger 记录一致 |
| 审计链 | `verify --audit-chain` | 无断链 |
| 用户体系 | `SELECT count(*)` | 至少 1 个 admin，无游离记录 |

失败项：

- 必须阻断上线（stop-the-line），除非运维书面确认接受偏差。
- 任何阻断项需写入 `docs/migration-exceptions.md`（由人工维护），说明接受原因与补救计划。

## 九、性能与资源

- 迁移机器基线：8C16G，SSD。
- `concurrency` 默认 4；`batch-size` 默认 500。
- 预计耗时（经验值，按实际源数据量重估）：
  - Macro/Signals：每百万行 ≈ 5 分钟。
  - Price bar：每百万行 ≈ 8 分钟。
  - Files：受 IO 限制，按源总大小估算。
- PG 期间临时配置：`maintenance_work_mem=1GB`、`checkpoint_timeout=30min`、`max_wal_size=4GB`；迁移完成后恢复默认。
- TimescaleDB：建议迁移期**先不启用压缩**，完成后再 `alter_job` 开启 continuous aggregates + 压缩策略。

## 十、回滚

- 原则：迁移是**一次性**写入目标库；回滚不做 down migration。
- 若数据问题严重：
  1. 停 AntClaw 业务服务；
  2. drop 受影响域的表 + 游标；
  3. 修复 CLI 或数据；
  4. 重跑。
- 敏感域（audit_logs、users）**禁止** drop；问题需 case-by-case 修复，保留链完整。

## 十一、监控

- 迁移期间 CLI 打印进度 + 写 `migration_progress(domain, rows_done, eta_sec)` 表（worker 周期更新）。
- Admin 控制台新增「迁移」临时页面（仅部署期存在，由 `ANTCLAW_MIGRATION_VIEW=true` 开启），展示进度与错误数。
- 错误数 > 阈值触发 Sentry 告警。

## 十二、安全

- 迁移 CLI 必须与业务 api 共享 `ANTCLAW_BYOK_MASTER_KEY`；所有敏感字段（若存在）加密后入库，禁止明文落库或明文日志。
- CLI 日志级别默认 `info`，明确禁止 `debug` 模式在生产开启（可能泄漏源数据片段）。

## 十三、验收清单（对照任务卡 P4a）

- [ ] `antclaw-migrate` CLI 所有子命令通过单测与 dry-run。
- [ ] 对同一 Badger 源重跑 `import` → 目标库行数稳定（幂等）。
- [ ] `verify` 全过。
- [ ] 创世审计写入 + `bootstrap-admin` 审计写入。
- [ ] MinIO `audit-worm` Object Lock 生效，测试删除被拒绝。
- [ ] 迁移期进度表有数据；Admin 临时视图可读。

## 十四、已决事项（2026-04-24）

- **M1 · Price tick 历史迁移**：**不迁移**；仅迁移日/小时 bar（见 §4.4）；tick 历史接受真空窗口，上线后由 worker 重新累积。
- **M2 · Calendar 英文标题**：**源端直接拉多语言数据**；迁移时若源含英文字段即写入 `title_i18n['en-US']`；源缺失则字段留空，不做机翻；上线后 worker 以相同方式补拉。
- **X1 · `bootstrap-admin` 入参方式**：**交互式 prompt + 环境变量两者都支持**；首次登录**不强制改密**（由用户自行决定是否立即修改）。
- **X2 · JWT `kid` 轮换演练触发**：**手动**；由运维在 runbook 中按季度发起；不做自动轮换。
- **X3 · 审计 WORM 双写失败阈值**：累计 **5 次**双写失败后**阻断后续业务写事务**，直到运维人工确认恢复；失败计数与恢复动作本身写入本地审计（PG）。
