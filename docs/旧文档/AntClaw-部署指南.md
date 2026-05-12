# AntClaw · 部署指南（单机 Docker Compose）

> 本文是重构方案 §十四 的可执行细则。终态为**单机单套 Compose**；后续硬化增益通过 Compose profile 叠加。

## 一、拓扑

```
                          ┌─────────────────────────────────────┐
                          │             Caddy（TLS）             │
                          │   /api → antclaw-api (8080)         │
                          │   /sse → antclaw-api (SSE 同端口)   │
                          │   /admin → antclaw-admin (静态)     │
                          │   /    → antclaw-web (静态)         │
                          └──────────────┬──────────────────────┘
                                         │
      ┌────────────┬────────────┬────────┴────────┬────────────┬────────────┐
      ▼            ▼            ▼                 ▼            ▼            ▼
 antclaw-api  antclaw-worker  postgres (TSDB)   redis     minio      prometheus / grafana / jaeger
```

**MVP 必起服务**：`caddy`、`postgres`、`redis`、`minio`、`antclaw-api`、`antclaw-worker`、`antclaw-web`、`antclaw-admin`、`prometheus`、`grafana`、`jaeger`。

**可选 profile**（硬化/后期）：

| profile | 新增服务 | 何时启用 |
|---|---|---|
| `replica` | `postgres-replica` | 读写分离压力期 |
| `runner` | `antclaw-backtest-runner` | 策略/回测上线后 |
| `mt` | `antclaw-mt-gateway` | MT4/MT5 等价层上线后 |

## 二、目录布局（部署侧）

```
deploy/
├── docker-compose.yaml
├── Dockerfile.backend
├── Dockerfile.web
├── Dockerfile.admin
├── Caddyfile
├── prometheus/
│   └── prometheus.yml
├── grafana/
│   └── provisioning/ (datasources, dashboards)
├── postgres/
│   └── init.sql            # 仅启用扩展；schema 由 migrations 负责
└── minio/
    └── init.sh             # 创建 buckets + Object Lock
```

## 三、镜像

- `antclaw-api`、`antclaw-worker`、`antclaw-backtest-runner`、`antclaw-mt-gateway` 共用 `Dockerfile.backend`（多 stage），差异由 entrypoint `ANTCLAW_CMD=api|worker|runner|mt-gateway` 决定。
- `antclaw-web` / `antclaw-admin` 各自多 stage 构建 → Nginx alpine **不使用**（Caddy 直接提供静态文件）。改为构建期产物拷贝到 Caddy 容器共享卷。
- 镜像标签：`antclaw/<component>:<git_sha>`；`latest` 仅指向最新 tag。

## 四、Compose 核心片段（节选，完整见 `deploy/docker-compose.yaml`）

```yaml
services:
  caddy:
    image: caddy:2
    restart: unless-stopped
    ports: ["80:80", "443:443"]
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile:ro
      - caddy_data:/data
      - caddy_config:/config
      - web_dist:/srv/web:ro
      - admin_dist:/srv/admin:ro
    depends_on: [antclaw-api]

  postgres:
    image: timescale/timescaledb:2.16-pg16
    restart: unless-stopped
    environment:
      POSTGRES_DB: antclaw
      POSTGRES_USER: antclaw
      POSTGRES_PASSWORD: ${ANTCLAW_PG_PASSWORD}
    volumes:
      - pgdata:/var/lib/postgresql/data
      - ./postgres/init.sql:/docker-entrypoint-initdb.d/00-init.sql:ro
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U antclaw"]
      interval: 10s
      retries: 10

  redis:
    image: redis:7-alpine
    restart: unless-stopped
    command: ["redis-server", "--appendonly", "yes", "--appendfsync", "everysec"]
    volumes: [redisdata:/data]

  minio:
    image: minio/minio:latest
    restart: unless-stopped
    command: ["server", "/data", "--console-address", ":9001"]
    environment:
      MINIO_ROOT_USER: ${ANTCLAW_MINIO_ROOT_USER}
      MINIO_ROOT_PASSWORD: ${ANTCLAW_MINIO_ROOT_PASSWORD}
    volumes: [miniodata:/data]

  antclaw-api:
    image: antclaw/backend:${IMAGE_TAG}
    restart: unless-stopped
    environment:
      ANTCLAW_CMD: api
      ANTCLAW_PG_DSN: postgres://antclaw:${ANTCLAW_PG_PASSWORD}@postgres:5432/antclaw?sslmode=disable
      ANTCLAW_REDIS_ADDR: redis:6379
      ANTCLAW_S3_ENDPOINT: http://minio:9000
      ANTCLAW_S3_ACCESS_KEY: ${ANTCLAW_MINIO_ROOT_USER}
      ANTCLAW_S3_SECRET_KEY: ${ANTCLAW_MINIO_ROOT_PASSWORD}
      ANTCLAW_JWT_PRIVATE_KEY: ${ANTCLAW_JWT_PRIVATE_KEY}
      ANTCLAW_JWT_PUBLIC_KEY:  ${ANTCLAW_JWT_PUBLIC_KEY}
      ANTCLAW_BYOK_MASTER_KEY: ${ANTCLAW_BYOK_MASTER_KEY}
      SENTRY_DSN_BACKEND: ${SENTRY_DSN_BACKEND}
      OTEL_EXPORTER_OTLP_ENDPOINT: http://jaeger:4318
    depends_on:
      postgres: { condition: service_healthy }
      redis:    { condition: service_started }

  antclaw-worker:
    image: antclaw/backend:${IMAGE_TAG}
    environment:
      ANTCLAW_CMD: worker
      # 与 api 共享大部分环境变量
    depends_on: [antclaw-api]

  antclaw-web:      # 仅构建拷贝产物到 web_dist 卷
    image: antclaw/web:${IMAGE_TAG}
    command: ["sh", "-c", "cp -r /dist/. /srv/web/"]
    volumes: [web_dist:/srv/web]

  antclaw-admin:
    image: antclaw/admin:${IMAGE_TAG}
    command: ["sh", "-c", "cp -r /dist/. /srv/admin/"]
    volumes: [admin_dist:/srv/admin]

  prometheus:
    image: prom/prometheus:latest
    volumes: [./prometheus/prometheus.yml:/etc/prometheus/prometheus.yml:ro]
    ports: ["9090:9090"]

  grafana:
    image: grafana/grafana:latest
    volumes: [./grafana/provisioning:/etc/grafana/provisioning:ro, grafanadata:/var/lib/grafana]
    environment:
      GF_SECURITY_ADMIN_PASSWORD: ${ANTCLAW_GRAFANA_PASSWORD}
    ports: ["3000:3000"]

  jaeger:
    image: jaegertracing/all-in-one:latest
    environment:
      COLLECTOR_OTLP_ENABLED: "true"
    ports: ["16686:16686"]

  # --- profile: replica ---
  postgres-replica:
    image: timescale/timescaledb:2.16-pg16
    profiles: [replica]
    # 流复制配置...

  # --- profile: runner ---
  antclaw-backtest-runner:
    image: antclaw/backend:${IMAGE_TAG}
    profiles: [runner]
    environment:
      ANTCLAW_CMD: runner
    # cpu_quota / mem_limit 在硬化阶段设置

  # --- profile: mt ---
  antclaw-mt-gateway:
    image: antclaw/backend:${IMAGE_TAG}
    profiles: [mt]
    environment:
      ANTCLAW_CMD: mt-gateway

volumes:
  pgdata:
  redisdata:
  miniodata:
  caddy_data:
  caddy_config:
  grafanadata:
  web_dist:
  admin_dist:
```

## 五、Caddyfile

```caddy
{$ANTCLAW_DOMAIN} {
    encode zstd gzip

    @api   path /api/* /sse/*
    reverse_proxy @api antclaw-api:8080 {
        flush_interval -1       # SSE 禁缓冲
    }

    handle_path /admin/* {
        root * /srv/admin
        try_files {path} /index.html
        file_server
    }

    handle {
        root * /srv/web
        try_files {path} /index.html
        file_server
    }

    log {
        output stdout
        format json
    }
}
```

## 六、环境变量清单

> 变量名以 `ANTCLAW_*` 前缀（决策 #1）；业务外的第三方变量（Sentry/OTel）保留其官方名。

| 变量 | 说明 | 示例 |
|---|---|---|
| `ANTCLAW_DOMAIN` | 对外域名 | `app.antclaw.example.com` |
| `ANTCLAW_PG_PASSWORD` | PG 密码 | — |
| `ANTCLAW_REDIS_PASSWORD` | Redis 密码（启用后） | — |
| `ANTCLAW_MINIO_ROOT_USER/PASSWORD` | MinIO 凭据 | — |
| `ANTCLAW_JWT_PRIVATE_KEY` / `ANTCLAW_JWT_PUBLIC_KEY` | Ed25519 PEM | — |
| `ANTCLAW_JWT_PREV_PUBLIC_KEY` | 旧公钥，用于轮换期校验 | 可空 |
| `ANTCLAW_BYOK_MASTER_KEY` | BYOK 主密钥（`v<n>:<base64>` 多版本） | — |
| `ANTCLAW_ARGON2_MEMORY/ITERS/PARALLELISM` | 覆盖 Argon2id 默认 | 可空 |
| `ANTCLAW_SMTP_*` | 注册/重置邮件 SMTP | — |
| `ANTCLAW_S3_ENDPOINT/ACCESS_KEY/SECRET_KEY/BUCKET/REGION` | 对象存储 | — |
| `SENTRY_DSN_BACKEND` | 后端 Sentry | — |
| `SENTRY_DSN_FRONTEND` | 前端 Sentry（注入 Vite 构建期） | — |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OTel collector | `http://jaeger:4318` |
| `ANTCLAW_GRAFANA_PASSWORD` | Grafana admin 密码 | — |
| `ANTCLAW_BOT_*_ENABLED` | Bot 开关（本期全 false） | `false` |
| `IMAGE_TAG` | 镜像 tag（git sha） | — |

`.env.example` 保存模板；`.env` 加入 `.gitignore`。

## 七、初始化顺序

1. `make build`：构建 backend、web、admin 镜像。
2. `docker compose up -d postgres redis minio`；等待健康检查通过。
3. `docker compose run --rm antclaw-api antclaw-migrate`：执行 `db/migrations` + sqlc 生成物 + 初始化 `audit-worm` bucket + 创建初始 admin 用户（由 CLI prompt 或 `ADMIN_EMAIL` / `ADMIN_PASSWORD` 环境变量）。
4. `docker compose up -d`：起全部服务。
5. 访问 `https://<ANTCLAW_DOMAIN>/admin` 用初始 admin 登录，立刻修改密码。

## 八、备份与恢复

- **PG**：`worker` 内置 cron 任务 `backup_postgres`，每日 02:00 `pg_dump -Fc` → 上传 MinIO `backups/postgres/<yyyy-mm-dd>.dump`。
- **MinIO audit-worm**：对象锁保障，禁止本地备份覆盖；按季度由运维从 MinIO `mc mirror` 至异地 R2。
- **Redis**：AOF + 每日 RDB 快照自动留存于 `redisdata` 卷；不跨机复制（本期单机）。
- **恢复演练**：每季度一次；流程：新实例 → 还原 `pgdata` + `miniodata` → 跑 `antclaw-migrate --verify` → 起 api/worker → 人工验收。

## 九、观测部署

- Prometheus 刮取：`antclaw-api`、`antclaw-worker`、`postgres_exporter`、`redis_exporter`、`minio`、`node_exporter`。
- Grafana dashboard provisioning 预置：
  - API 延迟 / QPS / 错误率
  - SSE 连接数 / 投递延迟
  - Redis Streams 深度 / 任务 lag
  - PG 连接 / 慢查询 / 副本延迟（启用 replica 后）
  - BYOK 健康 / AI Token 用量
- Jaeger：OTLP 4318；api/worker 默认开启（采样 10%，错误全采）。
- Sentry：SaaS；两个 DSN 分别注入 backend/web。

## 十、升级与回滚

- **升级**：
  1. `docker compose pull`（或 `build`）；
  2. `docker compose run --rm antclaw-api antclaw-migrate --dry-run`；
  3. `docker compose run --rm antclaw-api antclaw-migrate`；
  4. `docker compose up -d --no-deps --build antclaw-api antclaw-worker antclaw-web antclaw-admin`；
  5. 健康检查通过 → 完成。
- **回滚**：
  - 代码回滚：切 `IMAGE_TAG` 到上一个 sha，重复 4；
  - 数据回滚：原则上不做 DB 向下迁移；仅依赖备份恢复；迁移文件必须支持 `down`，但**生产环境禁止**执行 down（由 `ANTCLAW_MIGRATE_ALLOW_DOWN=false` 强制）。

## 十一、安全加固

- 容器默认 `read_only: true` + 必要 tmpfs；`cap_drop: [ALL]`，按需 `cap_add`。
- 内部网络：`antclaw_backend`（api/worker/pg/redis/minio）与 `antclaw_edge`（caddy + 前端静态卷）分离。
- 系统用户 `antclaw` 非 root；所有业务镜像 `USER 10001`。
- 容器 `no-new-privileges: true`；seccomp 使用 docker 默认 profile。
- TLS：Caddy 自动 Let's Encrypt；HSTS 默认开启；禁止 HTTP 明文代理。
- 审计 WORM：MinIO `audit-worm` bucket 创建时 `mc retention set compliance 3650d` 保护期 10 年。
- **沙箱后期硬化（LP-A/B/C 阶段启用）**：`deploy/seccomp-strict.json` 占位文件（基于 Docker default 再关闭 `unshare/mount/ptrace` 等）；`antclaw-backtest-runner` 启用时配置 `security_opt: ["seccomp=deploy/seccomp-strict.json"]`，见《重构解决方案》§十A.4。

## 十二、故障处理速查

| 症状 | 排查 |
|---|---|
| SSE 客户端持续 5xx | 查 Caddy `flush_interval`，检查 api pod 内存 |
| PG 连接打满 | 查 `antclaw-api` pgx pool 配置；`pg_stat_activity` |
| BYOK 调用全部失败 | 先确认主密钥未轮换错误；查 `byok:health:*` |
| SSE 延迟超 SLA | 检查 Redis Streams lag + consumer group |
| 审计双写 MinIO 失败 | Sentry 告警 → 检查 MinIO Object Lock 状态 |

## 十三、验收清单（对照任务卡 P12）

- [ ] `docker compose up -d` 在空机可一次起成功。
- [ ] 首次 migrate 创建 admin 用户并写入审计首条记录。
- [ ] Caddy 自动 TLS 生效；HSTS 响应头齐全。
- [ ] Prometheus + Grafana 全部 dashboard 有数据。
- [ ] 备份任务次日可在 MinIO `backups/postgres/` 看到 dump。
- [ ] profile `replica` / `runner` / `mt` 可独立启停，不影响 MVP 服务。

## 十四、已决事项（2026-04-24）

- **D1 · MinIO → Cloudflare R2**：**暂不切换**；保持 MinIO；任何切换需用户书面变更本条，运维不得自行迁移。
- **D2 · `pgbouncer`**：**不主动引入**；以 `postgres` 连接池告警为触发条件（`pg_stat_activity.count > max_connections * 0.8` 持续 10 分钟），触发后开临时任务卡评估再决定。
- **X4 · Redis 单机故障降级**：**不实现显式降级模式**；依赖 Prometheus 告警 + Sentry 上报即可；Redis 不可用时 api/worker 快速失败，恢复后自愈。
- **X5 · Sentry 数据脱敏**：在后端与前端 Sentry SDK 各安装 `beforeSend` 钩子，按关键词表拦截**政治类**内容（关键词清单存 `deploy/sentry-scrub.yaml`，由运维维护），命中则整条丢弃；其他字段按默认规则上报。
- **X6 · `ANTCLAW_BYOK_MASTER_KEY` 托管**：**直接塞 `.env`**；`.env` 文件权限 `0600`，宿主机仅 `antclaw` 用户可读；轮换流程走 `antclaw-migrate rotate-byok` + 手工更新 `.env`。
