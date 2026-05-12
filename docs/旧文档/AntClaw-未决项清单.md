# AntClaw · 未决项清单（总表）

> **2026-04-24 · 全部 26 条已拍板并回写各源文档。** 本文保留作为决议索引与审计入口；后续如再出现新未决项，继续追加到本文并按 §十三 模板记录决议。

## 使用说明

- 每条包含：**出处**、**问题**、**决议**、**同步位置**。
- 任何新增未决项（代码实现中出现的边界问题）必须先回本文登记，再由用户拍板；AI 助手**不得**跳过本流程（宪章 1、2）。

---

## 一、用户系统与鉴权（出处：`AntClaw-用户系统与鉴权.md §十五`）

- [x] **U1 · 异地登录邮件判定口径**
  - **决议**：按 **IP 国家级**触发安全通知邮件；子网级不启用。
  - 同步：`AntClaw-用户系统与鉴权.md §6.2 / §十五`

- [x] **U2 · `premium` 订阅与支付入口**
  - **决议**：MVP 仅展示当前层级；**升级按钮置灰**；支付方式本期不集成。
  - 同步：`AntClaw-用户系统与鉴权.md §十五`、`AntClaw-功能清单.md §二`

- [x] **U3 · 管理员账户 2FA**
  - **决议**：**不启用** 2FA（管理员也不启用）；字段保留但全局开关禁用。
  - 同步：`AntClaw-用户系统与鉴权.md §4.4 / §十五`

## 二、订阅与 SSE（出处：`AntClaw-订阅与SSE.md §十一`）

- [x] **S1 · `briefing` 首登补发**
  - **决议**：**不补发**；历史简报由页面按需拉取。
  - 同步：`AntClaw-订阅与SSE.md §十一`

- [x] **S2 · `price_ticks` free 用户 symbol 上限**
  - **决议**：**无上限**；仅按连接数与 RPM 控制。
  - 同步：`AntClaw-订阅与SSE.md §十一`、`AntClaw-用户系统与鉴权.md §7.2`

## 三、国际化（出处：`AntClaw-国际化规范.md §十一`）

- [x] **I1 · `zh-TW` 机翻占位**
  - **决议**：**不启用**；首发仅 `zh-CN` + `en-US`。
  - 同步：`AntClaw-国际化规范.md §十一`

- [x] **I2 · 货币显示**
  - **决议**：**跟随业务**；后端返回 `Money{currency}`，前端不做汇率折算。
  - 同步：`AntClaw-国际化规范.md §十一`

## 四、Bot 接入（出处：`AntClaw-Bot接入规范.md §十三`）

- [x] **B1 · 同平台多 chat 绑定**
  - **决议**：**单绑定**；活跃绑定唯一约束保留。
  - 同步：`AntClaw-Bot接入规范.md §十三`

- [x] **B2 · 管理员代发消息**
  - **决议**：**不提供**；不新增该接口。
  - 同步：`AntClaw-Bot接入规范.md §十三`

## 五、前端架构（出处：`AntClaw-前端架构.md §十五`）

- [x] **F1 · Service Worker 离线壳**
  - **决议**：**启用**；仅缓存 app shell；业务 API 与 SSE 不走 SW；`vite-plugin-pwa`；`prompt to reload` 策略。
  - 同步：`AntClaw-前端架构.md §十五`

- [x] **F2 · 图表主题与 dark mode**
  - **决议**：**手动切换**；在设置页提供 light/dark 开关；不随系统自动切换。
  - 同步：`AntClaw-前端架构.md §十五`

## 六、管理员控制台（出处：`AntClaw-管理员控制台.md §九`）

- [x] **A1 · Admin access TTL**
  - **决议**：**Admin 5 分钟** / 用户端 15 分钟；Refresh 两端 30 天；Admin 过期前 60s 静默刷新。
  - 同步：`AntClaw-管理员控制台.md §九`、`AntClaw-用户系统与鉴权.md §5.1`（实施时同步修订）

- [x] **A2 · admin 查看明文 BYOK**
  - **决议**：**永远不允许**；Admin UI 仅展示指纹与健康状态；无任何 RPC/CLI 解出明文。
  - 同步：`AntClaw-管理员控制台.md §九`、`AntClaw-用户系统与鉴权.md §十`

## 七、部署（出处：`AntClaw-部署指南.md §十四`）

- [x] **D1 · MinIO → Cloudflare R2**
  - **决议**：**暂不切换**；保持 MinIO；切换需书面变更。
  - 同步：`AntClaw-部署指南.md §十四`

- [x] **D2 · `pgbouncer`**
  - **决议**：**不主动引入**；以连接打满告警为触发条件（> 80% 持续 10 min）。
  - 同步：`AntClaw-部署指南.md §十四`

## 八、迁移（出处：`AntClaw-迁移指南.md §十四`）

- [x] **M1 · Price tick 历史**
  - **决议**：**不迁移**；接受真空窗口。
  - 同步：`AntClaw-迁移指南.md §4.4 / §十四`

- [x] **M2 · Calendar 英文标题**
  - **决议**：**源端直接拉多语言**；源有则写入 `title_i18n['en-US']`，源缺失则留空；**不机翻**。
  - 同步：`AntClaw-迁移指南.md §4.2 / §十四`

## 九、移动端（出处：`AntClaw-移动端架构.md §五`）

- [x] **MB1 · RN 图表库**
  - **决议**：**RN 启动任务卡之前定**；之前不依赖具体库。
  - 同步：`AntClaw-移动端架构.md §五`

- [x] **MB2 · RN 与 ArkTS 共用 DSL**
  - **决议**：本期**不共用**；**HarmonyOS 启动前复议**。
  - 同步：`AntClaw-移动端架构.md §五`

- [x] **MB3 · Pad / 平板布局**
  - **决议**：**支持 iPad / Android Tablet**；首发即提供自适应（断点 ≥ 768dp）；HarmonyOS 同等要求。
  - 同步：`AntClaw-移动端架构.md §五`

## 十、功能清单（出处：`AntClaw-功能清单.md §二十一`）

- [x] **FN1 · Pin 表**
  - **决议**：**不拆分**；统一 `user_pins(user_id, ref_type, ref_id, created_at)`。
  - 同步：`AntClaw-功能清单.md §二十一`、领域模型补充。

- [x] **FN2 · `GetHistory` 保留窗口**
  - **决议**：**永久**；仅 `ClearHistory` 手动清理。
  - 同步：`AntClaw-功能清单.md §二十一`

- [x] **FN3 · `TranslateStrategy` 指令集**
  - **决议**：**策略沙箱启动前定**；MVP 返回 `UNIMPLEMENTED`。
  - 同步：`AntClaw-功能清单.md §二十一`、`AntClaw-重构解决方案.md §2.3 目录结构（`cmd/antclaw-backtest-runner/`、`internal/adapter/sandbox/`）

## 十一、跨文档隐含议题

- [x] **X1 · `bootstrap-admin` 入参方式**
  - **决议**：交互式 prompt + 环境变量**两者都支持**；首次登录**不强制改密**。
  - 同步：`AntClaw-迁移指南.md §十四`、`AntClaw-部署指南.md §七`

- [x] **X2 · JWT `kid` 轮换触发**
  - **决议**：**手动**；按季度运维 runbook。
  - 同步：`AntClaw-迁移指南.md §十四`、`AntClaw-部署指南.md §十一`

- [x] **X3 · 审计 WORM 双写失败阈值**
  - **决议**：累计 **5 次**失败后**阻断业务写事务**，需运维人工确认恢复。
  - 同步：`AntClaw-迁移指南.md §十四`、`AntClaw-用户系统与鉴权.md §十一`

- [x] **X4 · Redis 单机故障降级**
  - **决议**：**不实现显式降级**；依赖告警与上报，恢复后自愈。
  - 同步：`AntClaw-部署指南.md §十四`

- [x] **X5 · Sentry 数据脱敏**
  - **决议**：后端与前端各安装 `beforeSend` 钩子，按关键词表**拦截政治类内容**（清单 `deploy/sentry-scrub.yaml`，运维维护），命中整条丢弃；其他字段正常上报。
  - 同步：`AntClaw-部署指南.md §十四`

- [x] **X6 · `ANTCLAW_BYOK_MASTER_KEY` 托管**
  - **决议**：**Compose 直接塞 `.env`**；文件权限 `0600`；轮换走 `antclaw-migrate rotate-byok` + 手工更新。
  - 同步：`AntClaw-部署指南.md §十四`

---

## 十二、决议执行注意事项

下列决议在后续实现时需**显式体现**，不得被任何任务卡忽略：

1. **A1 双 TTL**：`AuthService.Login` 需按请求来源（Admin 前端独立 audience 或路径）签发不同 `exp`；Admin 访问 `/admin/*` 路径校验 audience = `antclaw-admin`。
2. **A2 硬边界**：Admin proto / Go service 禁止任何 `GetUserAiKeyPlaintext` 类方法；代码 review 作为红线检查项。
3. **X3 阻断机制**：`audit_writer` 组件内置连续失败计数；达到阈值后向主进程发熔断信号，其余业务写事务走 `FAILED_PRECONDITION` 错误码，响应体附 `X-AntClaw-Audit-Frozen: true`。
4. **X5 脱敏清单**：`deploy/sentry-scrub.yaml` 必须有初始种子版本（运维补全关键词）；CI 校验 YAML 合法性；前后端 Sentry SDK 启动时加载。
5. **F1 SW 边界**：SW 仅 scope `/`（用户端）与 `/admin`（管理端各自 scope），**不拦截** `/api/*`、`/sse/*`、`/admin/api/*` 路径；CI 添加测试确保 fetch `/api/*` 不命中缓存。
6. **MB3 自适应断点**：RN 与 ArkTS 均以 `768dp` 为分界；公共断点常量 `ResponsiveBreakpoints` 写入 `packages/ui` 供双端使用（ArkTS 侧以同名常量同步）。
7. **P6c 沙箱目录占位**：MVP 阶段 `cmd/antclaw-backtest-runner/` 目录存在但 `main.go` 仅打印日志后退出；`internal/adapter/sandbox/` 目录骨架存在但文件均为空实现或仅返回 `ErrNotImplemented`；**禁止**引入 `go.starlark.net/starlark` 依赖；`go.mod` 中该依赖不得出现（CI 检查）。
8. **后期沙箱启用路径**：LP-A/B/C 阶段再填充 `cmd/antclaw-backtest-runner/runner/` 与 `internal/adapter/sandbox/{starlark,builtin,validator}/` 实现；启用时必须配套容器硬化配置（`deploy/seccomp-strict.json`、`docker-compose.yaml` 资源限制）。
9. **全面重写边界（硬约束）**：AntClaw 与 `ark-intelligent` **无代码继承关系**；后者仅为**功能参照项目**（对齐 `docs/AntClaw-功能清单.md` 与数据迁移格式）。**严禁**搬运、粘贴、逐行改名其任何 Go/TS/SQL/YAML/Shell 代码。AntClaw 采用自有代码风格与命名规范（Go 以《领域模型》为准，TS 以《前端架构》为准）。P0 阶段即取消独立的"改名"阶段，命名基线在仓库创建时直接采用，CI 零容忍 `ark`/`ARK` 字样（`docs/ARK-Intelligent-功能清单.md` 参照文档除外）及与参照项目源码的高位相似度。详见：
   - 《AntClaw-重构解决方案.md》§一第 1 条、§2.1
   - 《AntClaw-任务分解与AI助手约束.md》宪章 11、P0 任务卡
   - 《AntClaw-功能清单.md》开篇声明
   - 《AntClaw-Bot接入规范.md》§六

## 十三、新增未决项追加模板

```
- [ ] <编号> · <主题>
  - 问题：...
  - 当前默认：...
  - 影响：...
  - 建议决策时机：P0/P1/P2
  - 出处：<文档 §章节>
```

决议后改为：

```
- [x] <编号> · <主题>
  - 决议：...
  - 同步：<文档 §章节>
```
