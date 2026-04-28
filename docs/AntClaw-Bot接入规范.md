# AntClaw · Bot 接入规范（预留接口）

> 本文是 `AntClaw-重构解决方案.md` §十 的实现细则。**本期不交付任何具体 Bot 实现**（决策：Telegram / WeChat / Feishu 全部延后）。任务卡 P7 的范围**仅限接口、数据表、空实现与单测**；实际接入平台的工作在后续阶段由独立任务卡领取。
>
> AI 助手在本期**禁止**引入 Telegram / WeChat / Feishu 任何 SDK 或做具体平台的协议实现（宪章 3）。

## 一、目标

1. 定义统一的 `BotAdapter` 端口，使未来任何 Bot（Telegram、WeChat OA、Feishu、Slack、Discord…）均可平滑接入。
2. 让 **账号、订阅、告警、权限、审计、i18n、配额** 体系一次实现，所有 Bot 共享，无需重复建设。
3. 本期交付空骨架，保证添加新适配器时无需改动核心业务代码。

## 二、架构概览

```
                 ┌─────────────────────────────────────────────┐
                 │                核心业务层                    │
                 │  user / alert / notify / audit / quota / i18n │
                 └──────────────▲─────────────────▲────────────┘
                                │                 │
                ┌───────────────┘                 └───────────────┐
                │                                                 │
       ┌────────┴────────┐                              ┌─────────┴────────┐
       │  BotRouter      │  ←── ports/bot.go            │  Outbound Sender  │
       │  (inbound)      │                              │  (outbound)       │
       └────────▲────────┘                              └─────────▲────────┘
                │                                                 │
      ┌─────────┼─────────┐                              ┌────────┼────────┐
      │         │         │                              │        │        │
 ┌────┴───┐ ┌───┴────┐ ┌──┴────┐                    ┌────┴───┐ ┌──┴─────┐ ...
 │Telegram│ │WeChat  │ │Feishu │   ← 未来新增        │  同上  │ │   ...  │
 │Adapter │ │Adapter │ │Adapter│                    │        │ │        │
 └────────┘ └────────┘ └───────┘                    └────────┘ └────────┘
```

- **BotRouter**：核心进程侧，将入站 `BotMessage` 映射为业务意图（登录绑定、查询、订阅命令），复用业务层。
- **Adapter**：各平台协议实现，仅做协议层翻译。

## 三、接口定义（`backend/internal/ports/bot.go`）

### 3.1 平台与消息模型

```go
package ports

import "context"

type BotPlatform string

const (
    PlatformTelegram BotPlatform = "telegram"
    PlatformWeChat   BotPlatform = "wechat"
    PlatformFeishu   BotPlatform = "feishu"
    PlatformSlack    BotPlatform = "slack"
    PlatformDiscord  BotPlatform = "discord"
)

type BotMessage struct {
    Platform   BotPlatform       // 来源平台
    ExternalID string            // 平台侧消息 id（去重）
    ChatID     string            // 群/私聊 id
    FromUserID string            // 平台侧用户 id
    UserID     string            // 已绑定时对应的 AntClaw user_id；未绑定为空
    Kind       MessageKind       // text / command / callback / media / event
    Text       string            // 文本正文
    Command    string            // 解析后的命令（/cot、/price 等）
    Args       []string          // 命令参数
    Metadata   map[string]string // 扩展字段：locale、timezone、client_ver…
    ReceivedAt int64             // epoch ms
}

type BotOutbound struct {
    Platform   BotPlatform
    ChatID     string                 // 发送目标
    Kind       OutboundKind           // text / markdown / card / file / image
    MessageKey string                 // i18n key（复用统一 i18n）
    Args       map[string]interface{} // ICU 参数
    Locale     string                 // 明确渲染语种，空时取用户 locale
    Idem       string                 // 幂等 id，防止重发
    Payload    []byte                 // kind 非 text 时承载结构化内容
}

type MessageKind string
type OutboundKind string
```

### 3.2 `BotAdapter` 接口

```go
type BotAdapter interface {
    Platform() BotPlatform
    // Start 启动长连接 / webhook；将入站消息写入 in channel。
    Start(ctx context.Context, in chan<- BotMessage) error
    // Send 向平台发送一条消息；必须幂等（利用 Outbound.Idem）。
    Send(ctx context.Context, out BotOutbound) error
    // Close 优雅关闭。
    Close() error
}
```

### 3.3 注册与装配

- 每个 Adapter 提供工厂 `func New<Platform>(cfg Config, deps Deps) BotAdapter`。
- 主进程 `app.WireBots()` 根据配置 `ANTCLAW_BOT_<PLATFORM>_ENABLED=true/false` 决定是否装配；本期全部为 false。
- **Adapter 不直接依赖业务包**；通过 `BotRouter` 与业务层解耦。

## 四、BotRouter（入站路由）

- 职责：`BotMessage` → 业务调用。
- 流程：
  1. **去重**：以 `(platform, external_id)` 查 Redis `bot:dedup:<platform>:<ext_id>`（TTL 24h），命中即丢弃。
  2. **身份绑定**：
     - 已绑定（`bot_bindings` 查到）→ 填充 `UserID`，注入 `context`。
     - 未绑定 → 仅允许白名单命令（如 `/bind <token>`、`/help`）。
  3. **命令分发**：映射到业务服务（`UserService`、`CotService` 等），调用方式与 RPC 一致，共用鉴权与配额。
  4. **响应**：业务返回 `message_key` + `args` → 写入出站队列 `XADD ev:bot_out:<platform>` → Adapter `Send`。
  5. **审计**：入站与出站各写一条 `audit_logs`，字段 `actor=bot:<platform>:<ext_user>`。
- **禁止**：在 Router 中内嵌平台协议细节（Markdown 语法、keyboard 结构等）。平台差异全部在 Adapter 侧处理。

## 五、数据模型

### 5.1 `bot_bindings`

| 列 | 类型 | 说明 |
|---|---|---|
| `id` | `uuid` | PK |
| `user_id` | `uuid` | FK |
| `platform` | `text` | `telegram` / `wechat` / ... |
| `external_user_id` | `text` | 平台侧 user id |
| `external_chat_id` | `text` | 默认聊天（用于主动推送） |
| `display_name` | `text` | 平台侧昵称（仅展示） |
| `locale` | `text` | 平台侧识别 locale，缺失回落到 `users.locale` |
| `bound_at` | `timestamptz` | — |
| `unbound_at` | `timestamptz` | 可空；软解绑 |

唯一约束：`UNIQUE (platform, external_user_id) WHERE unbound_at IS NULL`。

### 5.2 `bot_bind_tokens`

一次性绑定 token（用户在 Web 生成 → 在 Bot 中发 `/bind <token>`）：

| 列 | 类型 | 说明 |
|---|---|---|
| `token_hash` | `bytea` | PK；SHA-256(token) |
| `user_id` | `uuid` | FK |
| `issued_at` | `timestamptz` | — |
| `expires_at` | `timestamptz` | +10 分钟 |
| `consumed_at` | `timestamptz` | 可空 |

## 六、命令注册表

- 路由表位于 `backend/internal/bot/commands.go`，形如：

```go
var commandMap = map[string]CommandHandler{
    "/help":  help.Handle,
    "/bind":  bind.Handle,    // 本期仅此命令的骨架有意义
    "/cot":   cot.Handle,     // 本期未暴露
    ...
}
```

- 命令文本以 `ARK-Intelligent-功能清单.md` 作为**功能参照**对齐命名与语义（仅对照业务能力，不复用其代码；参见《任务分解与 AI 助手约束》宪章 11），与 `AntClaw-重构解决方案.md §3.3` 对照列保持一致；未来激活任一 Bot 时，逐条映射到对应 `XxxService` RPC。
- 未知命令响应 `err.bot.unknown_command`（走 i18n）。

## 七、出站与 i18n

- 出站仅接受 `BotOutbound.MessageKey`；Adapter 在发送前调用 `i18n.Render(messageKey, args, locale)` 得到具体文本。
- 平台特有富文本（Markdown / Card）由 Adapter 定义 `render<Platform>(key, args, locale)` 扩展方法，本期**不实现**。
- 所有出站必须幂等：`bot:sent:<platform>:<idem>` Redis key（TTL 24h）；重试命中即跳过。

## 八、错误码

| key | 场景 |
|---|---|
| `err.bot.not_bound` | 用户未绑定任何账号 |
| `err.bot.bind_token_invalid` | 绑定 token 无效或已过期 |
| `err.bot.bind_already` | 已绑定同平台 |
| `err.bot.rate_limited` | Bot 侧配额（独立于 Web） |
| `err.bot.unknown_command` | 未知命令 |
| `err.bot.platform_disabled` | 管理员禁用该平台 |

## 九、安全

- **秘密管理**：平台 token / secret 放 `ANTCLAW_BOT_<PLATFORM>_TOKEN`，进程加载时只读一次，内存保存；禁止写日志。
- **签名校验**：每个 Adapter 必须实现入站签名/时间戳校验（如 Telegram secret_token、飞书 verification_token、微信签名）。
- **速率**：Bot 入站以 `external_user_id` 维度走令牌桶；绑定后叠加 `user_id` 维度（与 Web 共享配额池）。
- **PII**：不得把平台昵称 / 头像 URL 写入业务表（除 `display_name` 展示外）。

## 十、本期（P7）交付范围

1. `backend/internal/ports/bot.go`：接口 + 数据结构（本文 §三）。
2. `backend/internal/adapter/bot/README.md`：**空实现目录说明**，列出未来新增 Adapter 的步骤。
3. `backend/internal/adapter/bot/stub/stub.go`：合规的空 Adapter（用于编译与单测占位），所有方法返回 `ErrNotImplemented`。
4. `db/migrations/<序号>_bot_bindings.sql`：建 `bot_bindings` + `bot_bind_tokens`。
5. `backend/internal/bot/router.go`：仅接口 + 去重 + 绑定解析骨架，**不包含**任何具体命令实现（`commandMap` 为空，结构就绪）。
6. 单测：
   - 去重命中测试。
   - 未绑定用户仅放行白名单命令。
   - 空 Adapter 方法返回预期错误。
7. 文档：本文 + §一 列出的延后工作说明。

**禁止交付**：任何 `telegram-bot-api`、`wechat-sdk`、`feishu-sdk` 依赖；任何真实协议握手代码。

## 十一、未来阶段（非本期）

- P-Ext.1 · Telegram Adapter：接入 `go-telegram/bot`；对齐旧清单命令。
- P-Ext.2 · Feishu Adapter：机器人 + 卡片消息。
- P-Ext.3 · WeChat OA Adapter：公众号被动回复 + 客服消息。
- P-Ext.4 · Slack / Discord（视业务需要）。

每个未来任务必须：

- 读本文与《用户系统与鉴权》§七 配额章节；
- 不修改 `ports/bot.go` 接口（需要扩展时走用户确认）；
- 提交契约测试（入站消息样例 → Router 行为）与端到端沙箱测试（模拟平台）。

## 十二、验收清单（对照任务卡 P7）

- [ ] `ports/bot.go` 编译通过，接口与本文 §三 完全一致。
- [ ] `stub` Adapter 单测通过。
- [ ] `bot_bindings` 与 `bot_bind_tokens` 迁移脚本通过 `migrate up/down` 往返。
- [ ] 未引入任何具体 Bot 平台 SDK（`go.mod` diff 审查）。
- [ ] 文档与代码引用一致（接口签名、平台枚举、错误码 key）。

## 十三、已决事项（2026-04-24）

- **B1 · 同平台多 chat 绑定**：**单绑定**；保留 `UNIQUE(platform, external_user_id) WHERE unbound_at IS NULL`；同一用户在同一平台仅一条活跃绑定。
- **B2 · 管理员代发消息**：**不提供**；`AdminService` 不新增「代表用户在 Bot 侧发消息」接口。
