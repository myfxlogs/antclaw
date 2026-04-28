# AntClaw

> 金融数据分析平台 - 新一代架构重构版

## 项目结构

```
antclaw/
├── backend/          # Go 后端服务
├── frontend/         # 前端应用
│   ├── web/         # 用户 Web 端
│   ├── admin/       # 管理控制台
│   ├── mobile-rn/   # React Native 移动端
│   └── packages/    # 共享包
├── proto/           # Protobuf 契约定义
├── gen/             # 代码生成产物
├── deploy/          # 部署配置
├── scripts/         # 工具脚本
└── docs/            # 项目文档
```

## 开发规范

- **后端**: Go 1.22+, Connect-RPC, PostgreSQL, Redis
- **前端**: React + Vite + TypeScript + shadcn/ui + Tailwind
- **协议**: Protobuf + Connect (HTTP/1.1, HTTP/2, gRPC-Web)
- **部署**: Docker Compose, Caddy, MinIO

## 快速开始

```bash
# 启动开发环境
docker-compose -f deploy/docker-compose.yaml up -d

# 后端开发
cd backend && go run ./cmd/antclaw-api

# 前端开发
cd frontend/web && pnpm dev
```

## 文档

详见 `docs/` 目录下各专题文档。

- 开发环境与工具路径声明：`docs/开发环境与工具路径声明.md`

---

**注意**: 本项目处于重构阶段，API 与架构可能变动。
