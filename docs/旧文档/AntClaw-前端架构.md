# AntClaw · 前端架构（Web 与 Admin）

> 本文覆盖 `frontend/web/`（用户 Web）与 `frontend/admin/`（管理员 Web）。移动端见 `AntClaw-移动端架构.md`（P3）。本文与《任务分解与 AI 助手约束》宪章 5、6 强绑定：**前端零业务计算 + 密码明文输入**。

## 一、技术栈（强制）

| 能力 | 选型 | 备注 |
|---|---|---|
| 框架 | React 18 | 严格模式启用 |
| 构建 | Vite 5 | monorepo + pnpm workspaces |
| 语言 | TypeScript 5（`strict: true`） | 禁止 `any`、`@ts-ignore`（除接入三方类型的边界文件） |
| 路由 | TanStack Router | 文件路由 + 类型安全 |
| 数据 | TanStack Query + `@connectrpc/connect-query` | 统一缓存 |
| RPC | `@connectrpc/connect-web` | 只用 Connect 协议；禁用 REST 调用 |
| 样式 | Tailwind CSS + shadcn/ui | 禁止业务组件直接写 CSS 模块 |
| 表单 | React Hook Form + Zod | 语法校验；业务校验靠后端 |
| 图表 | ECharts（重度） + Recharts（轻度） | 统一封装在 `packages/ui/charts` |
| 国际化 | `react-i18next` + `i18next-icu` | 见《国际化规范》 |
| 状态 | Zustand（仅 UI state） | **禁止**把业务数据塞入 Zustand |
| 测试 | Vitest + React Testing Library + Playwright | E2E 在 CI 跑关键路径 |

**禁止引入**：Redux / MobX / Recoil / axios / fetch 直连后端 / WebSocket / jQuery / lodash 全量导入（可按方法树摇）。

## 二、Monorepo 结构

```
frontend/
├── web/                             # 用户 Web
│   ├── src/
│   │   ├── app/                     # 根装配：router、providers、i18n、query client
│   │   ├── routes/                  # TanStack Router 文件路由
│   │   ├── features/                # 业务模块（见 §四）
│   │   ├── api/                     # RPC client 装配 + 拦截器
│   │   ├── shared/                  # 应用内通用 hooks/utils（非可复用 UI）
│   │   └── main.tsx
│   ├── index.html
│   ├── vite.config.ts
│   └── tsconfig.json
├── admin/                           # 管理员 Web（结构同 web/，默认 locale zh-CN）
├── packages/
│   ├── ui/                          # shadcn/ui 封装、Design Tokens、图表
│   ├── i18n/                        # zh-CN、en-US 资源
│   ├── rpc-client/                  # @connectrpc 客户端封装 + 重试/鉴权
│   └── eslint-config/               # 共用 ESLint/Prettier/Stylelint
├── pnpm-workspace.yaml
└── tsconfig.base.json
```

- 每个包有独立 `package.json`；`web`、`admin` 仅依赖 `packages/*`，**不得**互相依赖。
- `packages/ui` 是唯一出口提供可见组件；`web/admin` 禁止直接写样式覆盖。

## 三、分层职责

```
routes/   ← URL 映射；只做组合
  └─ features/<domain>/
       ├─ pages/        ← 页面级组件（容器）
       ├─ components/   ← 业务组件（展示）
       ├─ hooks/        ← `useXxxQuery` / `useXxxMutation`（基于 connect-query）
       └─ types.ts      ← 仅 UI 本地类型；RPC 类型一律从 gen/ts 导入
api/
  ├─ clients.ts         ← 服务 Client 装配
  ├─ interceptors.ts    ← 鉴权、i18n 头、错误统一
  └─ error.ts           ← renderRpcError
shared/
  ├─ auth/              ← 会话状态（仅 UI），受保护路由守卫
  ├─ i18n/              ← react-i18next 初始化、locale switcher
  └─ ui/                ← 仅组合性小组件
```

**强制边界**：

- `features/*` 不得 `import 'api/clients'` 以外的 RPC 细节；统一经由 `api/`。
- `routes/*` 不得实现业务逻辑。
- `components/*` 不得发起 RPC（由 `hooks/` 调用 → props 注入）。

## 四、业务模块划分（与 Proto 对齐）

| 模块目录 | 对应 Service | 主要页面 |
|---|---|---|
| `features/auth` | `AuthService` | 登录 / 注册 / 重置密码 / 会话管理 |
| `features/dashboard` | 聚合多 Service | 首页总览（仅渲染后端聚合结果） |
| `features/cot` | `CotService` | 持仓总览 / 对比 / 信号 |
| `features/calendar` | `CalendarService` | 财经日历 / 事件详情 |
| `features/macro` | `MacroService` | 宏观各子源 |
| `features/price` | `PriceService` | 行情 / 情景 / 制度 |
| `features/vol` | `VolService` | VIX / GEX / IV / Skew |
| `features/signals` | `SignalService` | 偏好 / 排名 / X 因子 / 雷达 |
| `features/ta` | `TaService` | 技术面（Wyckoff / Elliott / ICT…） |
| `features/sentiment` | `SentimentService` | 情绪 / 链上 / DeFi / Carry |
| `features/ai` | `AiService` | 聊天（server-stream） / 解读 / 展望 |
| `features/alerts` | `AlertService` + Stream | 订阅管理 + 实时告警 |
| `features/backtest` | `BacktestService`（占位） | 页面显示「后期提供」 |
| `features/strategy` | `StrategyService`（占位） | 页面显示「后期提供」 |
| `features/settings` | `UserService` | 偏好 / BYOK / 会话 / 语言时区 |

**占位页面**强制要求：不做任何 RPC 调用，仅渲染 `ui.placeholder.coming_later` 文案；不得在前端实现简化版业务。

## 五、RPC 与数据流

### 5.1 Client 装配

```ts
// packages/rpc-client/src/transport.ts
import { createConnectTransport } from '@connectrpc/connect-web';
import { createPromiseClient } from '@connectrpc/connect';

export const transport = createConnectTransport({
  baseUrl: '/api',
  credentials: 'include',          // 携带 HttpOnly Cookie
  interceptors: [authInterceptor, i18nInterceptor, traceInterceptor, errorInterceptor],
});

export const authClient = createPromiseClient(AuthService, transport);
// ...其余 Service 同理
```

### 5.2 拦截器

- **authInterceptor**：401 触发 `AuthService.Refresh` → 成功后重放原请求；连续失败跳登录页。
- **i18nInterceptor**：注入 `Accept-Language`；响应 `Content-Language` 写入调试日志。
- **traceInterceptor**：透传 `traceparent`（从 OTel web SDK 取）。
- **errorInterceptor**：统一把 Connect `Code` + `ErrorDetail.message_key` 包装为 `AntClawError`。

### 5.3 Query 策略

- 默认 `staleTime=30s`、`gcTime=5m`；接口级在 hook 覆盖。
- **不缓存敏感数据**：`cacheTime=0` 用于会话列表、BYOK 密钥指纹。
- **不在客户端做跨接口聚合**：Dashboard 由后端 `PriceService.GetMarketOverview` 聚合返回。
- Mutation 成功后 `invalidateQueries` 目标 key；禁止手动拼接 cache。

### 5.4 SSE 订阅

- 组件层 `useStreamChannel(channel, filter)`：
  - 内部维护 `EventSource`；断线指数退避（1s, 2s, 4s…, 上限 30s）。
  - `Last-Event-ID` 由浏览器自动携带；组件消费成功后调用 `postAck(channel, event_id)`。
  - 事件按 `event_id` 去重（LRU 256）。
- 全局连接单例：同一用户只开**一条** SSE 连接，多订阅复用；组件 mount/unmount 仅加减订阅清单，不开新连接。

## 六、鉴权与路由守卫

- Cookie-based；前端**不存** JWT。
- `<ProtectedRoute roles={['free','premium','admin']}>`：`users/me` 查询失败或角色不符时重定向。
- 管理端独立域前缀 `/admin`；**两套应用**分别部署（同域不同路径，由 Caddy 路由）。
- `/admin/*` 路由守卫仅放行 `admin` 角色，非 admin 显示 403 页。

## 七、国际化 UI

- 所有可见字符串走 `useTranslation`；CI `i18n-check` 阻断中文硬编码。
- 语言切换：`features/settings/LocaleSwitcher` 调用 `UserService.UpdateSettings` 并 `i18n.changeLanguage`。
- 时间/数字：统一封装 `shared/format/{date,number,money}.ts`，内部基于 `Intl.*` + 用户 `timezone`。

## 八、密码输入组件（硬约束）

```tsx
// packages/ui/src/PasswordInput.tsx
export function PasswordInput(props: InputProps) {
  return <input type="text" autoComplete="new-password" {...props} />;
}
```

- 项目根 ESLint 规则：禁止 `<input type="password">`；Admin 与 Web 共用。
- 表单层禁止自行渲染 `*` 遮罩。

## 九、错误与 Toast

- `renderRpcError(err, t)` 是唯一渲染入口。
- Toast 使用 `packages/ui/src/toast`；错误 toast 显示 `t(message_key, args)`，附 `traceId` 小字用于排查。
- 致命错误（`INTERNAL` / 网络断开 > 10s）切换为全屏错误页而非 toast。

## 十、可观测

- 集成 **Sentry**（DSN 从 `import.meta.env.VITE_SENTRY_DSN` 读取；对应 `SENTRY_DSN_FRONTEND`）。
- 集成 **OTel Web SDK**（仅在生产环境开启）；span 关联 `traceparent`。
- 不采集个人数据；禁止 `replay` 功能记录密码页。

## 十一、构建与产物

- Web 产物目录：`frontend/web/dist/`；Admin：`frontend/admin/dist/`。
- Vite 构建 `base` 为相对路径；由 Caddy 以 `/` 与 `/admin` 提供。
- 产物缓存：`assets/*.<hash>.js` 强缓存；`index.html` 不缓存。
- 构建时注入环境变量：`VITE_API_BASE`、`VITE_SSE_BASE`、`VITE_SENTRY_DSN`、`VITE_BUILD_SHA`。

## 十二、CI 要求（前端）

- `pnpm -r lint`：ESLint + Prettier + Stylelint。
- `pnpm -r typecheck`：`tsc --noEmit`。
- `pnpm -r test`：Vitest 单测（覆盖率目标：业务组件 hooks ≥ 70%）。
- `pnpm --filter web e2e` / `pnpm --filter admin e2e`：Playwright 关键路径。
- `pnpm i18n:check`：见《国际化规范》§七。
- `pnpm -r build`：产物存档用于部署。

## 十三、性能基线

- Web 初次加载 gzip 后 ≤ 400 KiB（不含图表 chunk）。
- 首屏 LCP（4G）≤ 2.5s。
- 路由切换 TTI ≤ 300ms（已登录）。
- 单页面同时 Query 数 ≤ 8；超过需提供聚合接口。

## 十四、验收清单（对照任务卡 P10）

- [ ] `features/*` 目录层次一致；禁用 `any` CI 通过。
- [ ] 无任何组件直接 `fetch` / `axios` / `WebSocket`。
- [ ] `PasswordInput` 在登录/注册/重置密码、BYOK、改密所有场景使用；ESLint 规则通过。
- [ ] SSE 全局单连接复用验证（devtools 只见一条 `text/event-stream`）。
- [ ] Lighthouse 跑分 Perf ≥ 85、A11y ≥ 90。
- [ ] Sentry 生产环境上报 `release = VITE_BUILD_SHA`。
- [ ] Playwright：登录、订阅告警、接收实时事件、切换 locale 四路全绿。

## 十五、已决事项（2026-04-24）

- **F1 · Service Worker 离线壳**：**启用**；Web 使用 `vite-plugin-pwa` 生成 SW，仅缓存 app shell（HTML/JS/CSS）与静态资源；**业务 API 与 SSE 不走 SW**，保证鉴权与实时性；SW 更新采用 `prompt to reload` 策略，版本号绑定 `VITE_BUILD_SHA`。
- **F2 · 图表主题与 dark mode**：**手动切换**；在 `features/settings` 提供主题选项（`light` / `dark`），图表按用户偏好应用；**不**跟随系统 `prefers-color-scheme` 自动切换。
