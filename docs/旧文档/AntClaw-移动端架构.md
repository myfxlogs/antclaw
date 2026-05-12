# AntClaw · 移动端架构（RN + Expo · ArkTS）

> 本文是重构方案 §6.2 的实现细则。移动端整体排在 Web 与后端稳定后交付：
>
> - **首发**：`frontend/mobile-rn/` — React Native + Expo（iOS / Android，App Store + Google Play）。
> - **后期**：`frontend/mobile-arkts/` — HarmonyOS ArkTS。
>
> 本文对 AI 助手的约束与 `AntClaw-前端架构.md` 一致：**零业务计算、密码明文、i18n 全覆盖、禁用 WebSocket**。

## 一、总体原则

1. **共用 Proto**：移动端与 Web/Admin 共用 `gen/ts`（RN）/ `gen/` 生成的 ArkTS 绑定（后期）；不单独实现 HTTP 请求。
2. **实时通道**：移动端使用 **gRPC server-streaming** 而非 SSE（省电、原生长连接）；同 `stream.proto`。
3. **通知通道**：站内信 + 原生推送（FCM / APNs / HMS）；不依赖 SSE 在后台运行。
4. **权限最小化**：仅申请必要权限（通知、网络）；不申请位置、通讯录、相机（图表分享若需截图再单独申请）。
5. **端到端一致**：所有业务从 `XxxService.*` RPC 取数；移动端**不做**任何指标/聚合/过滤/排序逻辑。

## 二、RN + Expo 首发

### 2.1 技术栈

| 能力 | 选型 | 说明 |
|---|---|---|
| 框架 | React Native + **Expo SDK**（最新 LTS） | EAS Build / EAS Update |
| 语言 | TypeScript 5（`strict`） | 同 Web |
| 导航 | **Expo Router**（文件路由） | 与 TanStack Router 心智一致 |
| 数据 | TanStack Query + `@connectrpc/connect-query` | 共享 hooks 设计 |
| RPC | `@connectrpc/connect`（gRPC/Connect） + `@connectrpc/connect-node`（长连接） | 统一 transport |
| 状态 | Zustand（UI only） | 禁止放业务数据 |
| 样式 | **NativeWind**（Tailwind for RN） + 自定义 tokens | 与 Web 设计 token 映射 |
| 组件 | `react-native-paper` 或自建 + `packages/ui` native 子集 | 基础组件本地化 |
| 图表 | `victory-native` 或 `react-native-skia` | 大图场景走 Skia |
| 表单 | React Hook Form + Zod | 语法层校验 |
| i18n | `i18next` + `react-i18next` | 资源从 `packages/i18n` 打包 |
| 存储 | `expo-secure-store`（凭据）+ `@react-native-async-storage/async-storage`（非敏感） | — |
| 推送 | `expo-notifications`（iOS/Android） | FCM/APNs 由 Expo 中转或自配 |
| 观测 | Sentry RN SDK + OTel RN（可选） | DSN 来自 `SENTRY_DSN_MOBILE` |

**禁止引入**：Redux / MobX / Realm / WebSocket / WebView 承载业务页面。

### 2.2 目录结构

```
frontend/mobile-rn/
├── app/                         # Expo Router 文件路由
│   ├── _layout.tsx
│   ├── (auth)/                  # 登录/注册/重置
│   ├── (tabs)/                  # 主 Tab：Dashboard / Alerts / AI / Settings
│   └── detail/<domain>/[id].tsx
├── src/
│   ├── api/                     # 同 Web：clients + interceptors
│   ├── features/                # 与 Web features 对齐（见 §2.5）
│   ├── shared/
│   │   ├── auth/                # 会话守卫
│   │   ├── i18n/                # i18next 初始化
│   │   ├── push/                # 推送注册、token 同步
│   │   └── stream/              # gRPC server-streaming 封装
│   └── theme/                   # NativeWind tokens
├── assets/
├── app.config.ts                # Expo 配置（含 scheme、bundleId）
├── eas.json                     # EAS Build/Submit 配置
└── tsconfig.json
```

### 2.3 RPC 与实时

- **Transport**：Connect over HTTP/2；真机环境允许自签名证书仅用于 dev build。
- **流订阅**：`StreamService.Subscribe(channels, resume_from, symbols?, task_id?)` → 前台维持连接；后台切回用 `resume_from` 恢复。
- **后台策略**：进入后台 60s 主动断流；同时注册推送，由后端把关键事件（alerts/system_notice）同步经推送通道投递。
- **去重**：流事件 + 推送事件同 `event_id` 去重，防止重复提示。

### 2.4 鉴权

- Access + Refresh JWT 经 `expo-secure-store` 以 `AES-GCM` 由 OS Keystore/Keychain 保护。
- 启动时取 Access；过期自动走 `AuthService.Refresh`；两次失败清凭据回登录。
- **密码明文输入**（宪章 6）：`PasswordInput` RN 版本基于 `<TextInput secureTextEntry={false} />`，统一不遮罩。
- 生物识别（Face ID / Touch ID / 指纹）：可选开关，仅用于**本地解锁** App，不替代服务端鉴权。

### 2.5 业务模块（与 Web 对齐）

| 目录 | 说明 | 本期启用 |
|---|---|---|
| `features/auth` | 登录/注册/重置/会话管理 | ✅ |
| `features/dashboard` | 首页聚合 | ✅ |
| `features/alerts` | 告警中心 + 订阅管理 + 推送 | ✅ |
| `features/ai` | AI Chat（server-stream） | ✅ |
| `features/settings` | 语言/时区/BYOK/会话/推送偏好 | ✅ |
| `features/market` | Price / Vol / Signals / Sentiment 只读视图 | ✅ |
| `features/calendar` / `macro` / `cot` / `ta` | 详情只读 | ✅ |
| `features/backtest` / `strategy` | 占位页「后期提供」 | 占位 |

### 2.6 推送

- 注册：登录后调用 `UserService.RegisterPushToken(platform, token, app_version, locale)`。
- 服务端：`push_tokens(user_id, platform, token, expires_at, last_seen_at)`；后端 worker 在投递失败时清理失效 token。
- 通道与事件对应：
  - `CHANNEL_ALERTS` / `CHANNEL_SIGNALS` → 高优先推送（包含 `deep_link`）。
  - `CHANNEL_SYSTEM_NOTICE` → 普通推送。
  - 其他通道默认**不推送**，仅流通道投递。
- 文案：服务端走 i18n，`notification.title` / `notification.body` 从 `notify.push.*` key 渲染。

### 2.7 构建与分发

- **EAS Build**：iOS 与 Android profile 分 `development` / `preview` / `production`。
- **EAS Update**：仅修复非二进制层 bug；任何涉及原生模块/权限变更必须走商店审核。
- **版本**：`version` = `<年>.<周>.<patch>`；`runtimeVersion` = major（例 `2026.1`）。
- **签名与 ID**：
  - iOS bundleId：`com.antclaw.app`；App Store Connect 绑定。
  - Android applicationId：`com.antclaw.app`。
  - 用 Fastlane/EAS Submit 一键提审。

### 2.8 环境配置

- `app.config.ts` 读取：
  - `EXPO_PUBLIC_API_BASE`
  - `EXPO_PUBLIC_STREAM_BASE`
  - `EXPO_PUBLIC_SENTRY_DSN`
  - `EXPO_PUBLIC_BUILD_SHA`
- 环境分离：dev/staging/prod 三套；通过 EAS profile 注入。

### 2.9 可观测

- Sentry：错误 + 性能；禁用 `session replay`（隐私）。
- OTel RN（可选）：仅生产采样 10%，错误全采。
- 分析：不接入第三方统计 SDK；所有埋点走后端 `UserService.RecordEvent`（后期设计）。

### 2.10 性能基线

- 冷启动 → 登录页可交互（中端机，4G）≤ 3.0s。
- 主 Tab 首屏数据就绪（已登录） ≤ 1.5s。
- 低内存模式下长列表保持 60fps（使用 `FlashList`）。
- 安装包 iOS ≤ 60 MiB、Android ≤ 40 MiB（初次上架基线）。

### 2.11 验收清单（对应任务卡 P13）

- [ ] `gen/ts` 在 RN 工程零改动编译通过。
- [ ] `PasswordInput` 无遮罩；ESLint 规则通过。
- [ ] gRPC server-streaming 在 iOS/Android 均能保持连接 ≥ 10 分钟；后台切换可 `resume_from` 恢复。
- [ ] 推送通道：alert 从服务端到设备（P95）≤ 5s（FCM/APNs 常规）。
- [ ] 离线启动可显示上次缓存数据（Query persistent storage）；写操作离线排队失败并提示。
- [ ] Sentry 生产环境上报 `release = EXPO_PUBLIC_BUILD_SHA`。
- [ ] 商店上架素材齐全（截图两语种、隐私声明、权限解释）。

## 三、HarmonyOS ArkTS（后期）

### 3.1 范围

- `frontend/mobile-arkts/`：HarmonyOS 原生应用；**不与 RN 共用代码**，只共用 Proto 与 i18n 资源。
- 技术栈：ArkTS + ArkUI（声明式）；使用 DevEco Studio 构建。
- 发布渠道：华为应用市场。

### 3.2 关键点

- **Proto 绑定**：使用 `protoc` + `protoc-gen-grpc-arkts`（官方/社区插件）生成 ArkTS 客户端；若插件不成熟，暂用 Connect-over-HTTP + `protoc-gen-ts` 通过 @ohos/napi 封装。
- **通信**：优先 gRPC；HarmonyOS 原生 `@ohos.net.http` 支持 HTTP/2；流订阅使用 gRPC server-streaming，同 RN 策略。
- **推送**：HMS Push；`UserService.RegisterPushToken(platform="huawei", ...)`。
- **存储**：`@ohos.data.preferences`（非敏感） + `@ohos.security.huks`（凭据加密）。
- **i18n**：资源来自 `packages/i18n`，由构建脚本转换为 ArkTS `resources/*.json`。
- **设计 token**：从 `packages/ui` tokens 映射；ArkUI 主题定义统一。
- **编译/发布**：HAR / HAP；AppGallery Connect 提审。

### 3.3 模块优先级

- P3a（首版）：auth / dashboard / alerts / settings。
- P3b：market / calendar / macro / cot / ta / ai。
- P3c：策略/回测占位页（与 RN 一致显示「后期提供」）。

### 3.4 验收清单（对应任务卡 P14，后期）

- [ ] Proto 客户端在 HarmonyOS 真机可用。
- [ ] HMS Push 注册与投递链路通。
- [ ] i18n 资源从 `packages/i18n` 一键同步脚本 OK。
- [ ] 应用市场审核上架版本号与 `runtimeVersion` 一致。

## 四、跨端一致性约束

- **Feature parity**：RN 与 ArkTS 模块清单保持一致；任一端新增模块必须同步出任务卡给另一端。
- **共享规范**：
  - `PasswordInput` 明文（两端）。
  - 占位页文案 key：`ui.placeholder.coming_later`（两端）。
  - 推送文案 key：`notify.push.<channel>.<template>`（两端）。
- **禁止端特供业务逻辑**：若某功能需要"移动端简化版"，必须回到后端加专用 RPC，**不得**在客户端做妥协计算。

## 五、已决事项（2026-04-24）

- **MB1 · RN 图表库**：**择期定稿**；RN 任务卡启动前补做一次性能压测比选 `victory-native` 与 `react-native-skia`，在 RN 启动任务卡说明中登记最终选型；在此之前前端 mock 不依赖具体图表库。
- **MB2 · RN 与 ArkTS 共用 DSL**：本期**不共用**；HarmonyOS 启动前复议；期间双端以相同 i18n + 相同 Proto 保持能力一致。
- **MB3 · 平板 / Pad 布局**：**支持 iPad / Android Tablet 布局**；首发即提供自适应（断点 ≥ 768dp 切双栏）；HarmonyOS 后期版同等要求。
