# AI Agent 执行指南

## 一、阅读顺序

1. `00-文档索引.md`
2. `04-客户端开发指导.md`
3. 按功能阅读 `01-API接口总览.md` 中对应 Service
4. 阅读 `02-API字段参考.md` 中对应 RPC 字段
5. 如涉及通知，阅读 `03-SSE实时通知接口.md`
6. `07-文档质量评审与落地边界.md`，确认接口优先级和禁止实现范围

## 二、统一术语

| 术语 | 定义 |
|---|---|
| API 根地址 | `https://api.alfq.org/` |
| RPC 端点 | `POST https://api.alfq.org/antclaw.v1.<Service>/<Rpc>` |
| SSE 通知 | `GET https://api.alfq.org/sse/notifications` |
| access token | `Authorization: Bearer <token>` 使用的短期令牌 |
| refresh token | 用于刷新 access token 的长期令牌 |
| UiState | `idle/loading/success/error` 四态模型 |

## 三、开发硬约束

- 不猜接口，先查 `01` 和 `02`。
- 不访问 `ad.alfq.org`。
- 不实现管理端接口：`AdminService`、`AdminDataService`、`DataSourceService`、`SystemAIService`。
- 不直连 `8082`。
- 不使用硬编码假数据 fallback。
- 不手动编辑生成代码。
- API 失败时展示错误状态。
- 修改客户端网络层后必须验证登录、Healthz、SSE 401。

## 四、接口实现模板

每新增一个 Android API 调用，应包含：

```text
Rpc 封装函数
Repository 方法
ViewModel UiState
Screen 四态渲染
错误处理
必要的本地缓存
```

## 五、调试命令

健康检查：

```bash
curl -i https://api.alfq.org/antclaw.v1.SystemService/Healthz \
  -H 'Content-Type: application/json' \
  --data '{}'
```

SSE 未登录检查：

```bash
curl -i https://api.alfq.org/sse/notifications
```

期望返回 `401 unauthorized`。
