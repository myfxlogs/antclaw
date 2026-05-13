# SSE 实时通知接口

## 个人通知 SSE

- **URL**：`GET https://api.alfq.org/sse/notifications`
- **认证**：必须携带 `Authorization: Bearer <access_token>`
- **未登录响应**：`401 unauthorized`
- **Content-Type**：`text/event-stream`
- **服务端来源**：Redis Pub/Sub `user:<userID>:notifications`

## 事件格式

```text
event: notification
data: {...notification json...}

```

心跳可能为注释行：

```text
: ping

```

## Android 客户端策略

1. 登录成功后获取 token。
2. App 进入前台时调用 `NotificationService/UnreadCount` 和 `NotificationService/ListUnread` 补拉。
3. 使用 OkHttp SSE 连接 `https://api.alfq.org/sse/notifications`。
4. 收到 `notification` 事件后写入 Room 缓存并刷新未读数。
5. App 进入后台时主动关闭 SSE。
6. 断线后指数退避重连：`1s → 2s → 5s → 15s → 30s`。
7. 收到 `401` 时停止重连，清理 token 并跳转登录。

## 其它 SSE 端点

| 端点 | 主要用途 | Android 是否默认接入 |
|---|---|---|
| `/sse/jobs` | 后台任务事件 | 否，管理端用途 |
| `/sse/audit` | 审计事件 | 否，管理端用途 |
| `/sse/macro_alerts` | 宏观告警流 | 可按发现页需求接入 |
| `/sse/options_alerts` | 期权告警流 | 可按发现页需求接入 |
| `/sse/signals_alerts` | 信号告警流 | 可按信号页需求接入 |
| `/sse/notifications` | 个人通知 | 是，Android 必接 |

## Cloudflare Tunnel 注意事项

- `api.alfq.org` 不启用 Cloudflare Access。
- 长连接可能被移动网络或代理中断，Android 必须实现重连与前台补拉。
- 客户端禁止直连 `8082`。
