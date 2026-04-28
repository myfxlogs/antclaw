# Bot Adapter 目录

> **P7 范围**：仅接口、数据表、空实现与单测。实际接入平台的工作在后续阶段（LP-B）完成。

## 目录结构

```
internal/adapter/bot/
├── README.md          # 本文档
├── stub/              # Stub 实现（P7）
└── [future]/          # 未来平台实现（延后）
    ├── telegram/      # Telegram Bot（LP-B）
    ├── wechat/        # 微信公众号（LP-B）
    └── feishu/        # 飞书 Bot（LP-B）
```

## 实现步骤（未来新增 Adapter 时参考）

1. **创建目录**: `internal/adapter/bot/{platform}/`

2. **实现接口**: 实现 `ports.BotPort` 接口
   - `Send()`: 调用平台 API 发送消息
   - `ParseCommand()`: 解析命令格式

3. **Webhook 处理**: 创建 HTTP handler 接收平台消息
   - 验证签名/Token
   - 解析为 `ports.InboundMessage`
   - 调用 `BotRouter.Route()`

4. **注册 Router**: 在 `cmd/antclaw-api/main.go` 中注册命令处理器

5. **配置**: 添加 `ANTCLAW_BOT_{PLATFORM}_TOKEN` 等环境变量

## 本期禁止事项

- 禁止引入 Telegram/WeChat/Feishu SDK
- 禁止实现具体平台的协议细节
- 仅保留 `stub` 空实现通过单测

## 相关文档

- `docs/AntClaw-Bot接入规范.md` - 完整 Bot 接入规范
- `internal/ports/bot.go` - 端口接口定义
