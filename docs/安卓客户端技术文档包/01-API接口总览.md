# API 接口总览

## 统一调用规则

- **Base URL**：`https://api.alfq.org/`
- **Connect-RPC URL 格式**：`POST https://api.alfq.org/antclaw.v1.<Service>/<Rpc>`
- **Content-Type**：Android Connect 客户端自动处理；curl 调试可使用 `application/json`。
- **认证方式**：除 `AuthService.Register`、`AuthService.Login`、`AuthService.Refresh`、`SystemService.Healthz` 等公共接口外，业务接口默认应携带 `Authorization: Bearer <access_token>`。
- **错误格式**：Connect-RPC 错误，HTTP 状态与 Connect code 共同表达错误。

## 通用请求头

| Header | 必填 | 说明 |
|---|---|---|
| `Content-Type` | 是 | `application/json` 或 Connect 客户端默认编码 |
| `Authorization` | 业务接口是 | `Bearer <access_token>` |
| `Connect-Protocol-Version` | 客户端自动 | Connect-RPC 协议头 |

## 通用错误码

| HTTP/Connect | 含义 | Android 行为 |
|---|---|---|
| `400` / `invalid_argument` | 参数错误 | 标记字段错误，不重试 |
| `401` / `unauthenticated` | 未登录或 token 失效 | 清理 token，跳转登录 |
| `403` / `permission_denied` | 无权限 | 显示无权限 |
| `404` / `not_found` | 资源不存在 | 显示空态或不存在 |
| `409` / `already_exists` | 重复创建 | 显示冲突提示 |
| `429` / `resource_exhausted` | 限流 | 退避重试 |
| `500` / `internal` | 服务端错误 | 显示错误并允许重试 |
| 网络失败 | DNS/隧道/移动网络异常 | 使用缓存，显示重试入口 |

## AdminService

- **proto 文件**：`proto/antclaw/v1/admin.proto`
- **范围**：管理端
- **API 入口注册**：是

| RPC | 方法 | 端点 URL | 请求 | 响应 | 流式 |
|---|---|---|---|---|---|
| `ListUsers` | `POST` | `https://api.alfq.org/antclaw.v1.AdminService/ListUsers` | `ListUsersRequest` | `ListUsersResponse` | 否 |
| `SetRole` | `POST` | `https://api.alfq.org/antclaw.v1.AdminService/SetRole` | `SetRoleRequest` | `SetRoleResponse` | 否 |
| `Ban` | `POST` | `https://api.alfq.org/antclaw.v1.AdminService/Ban` | `BanRequest` | `BanResponse` | 否 |
| `Unban` | `POST` | `https://api.alfq.org/antclaw.v1.AdminService/Unban` | `UnbanRequest` | `UnbanResponse` | 否 |
| `RunJob` | `POST` | `https://api.alfq.org/antclaw.v1.AdminService/RunJob` | `RunJobRequest` | `RunJobResponse` | 否 |
| `ListJobs` | `POST` | `https://api.alfq.org/antclaw.v1.AdminService/ListJobs` | `ListJobsRequest` | `ListJobsResponse` | 否 |
| `SetJobEnabled` | `POST` | `https://api.alfq.org/antclaw.v1.AdminService/SetJobEnabled` | `SetJobEnabledRequest` | `SetJobEnabledResponse` | 否 |
| `ListAuditLogs` | `POST` | `https://api.alfq.org/antclaw.v1.AdminService/ListAuditLogs` | `ListAuditLogsRequest` | `ListAuditLogsResponse` | 否 |
| `ListWebhookDeliveries` | `POST` | `https://api.alfq.org/antclaw.v1.AdminService/ListWebhookDeliveries` | `ListWebhookDeliveriesRequest` | `ListWebhookDeliveriesResponse` | 否 |
| `ForceLogout` | `POST` | `https://api.alfq.org/antclaw.v1.AdminService/ForceLogout` | `ForceLogoutRequest` | `ForceLogoutResponse` | 否 |
| `AdminResetUserPassword` | `POST` | `https://api.alfq.org/antclaw.v1.AdminService/AdminResetUserPassword` | `AdminResetUserPasswordRequest` | `AdminResetUserPasswordResponse` | 否 |
| `SetUserCodeID` | `POST` | `https://api.alfq.org/antclaw.v1.AdminService/SetUserCodeID` | `SetUserCodeIDRequest` | `SetUserCodeIDResponse` | 否 |

## AdminDataService

- **proto 文件**：`proto/antclaw/v1/admin_data.proto`
- **范围**：管理端
- **API 入口注册**：是

| RPC | 方法 | 端点 URL | 请求 | 响应 | 流式 |
|---|---|---|---|---|---|
| `GetDataSummary` | `POST` | `https://api.alfq.org/antclaw.v1.AdminDataService/GetDataSummary` | `GetDataSummaryRequest` | `GetDataSummaryResponse` | 否 |
| `GetDataPreview` | `POST` | `https://api.alfq.org/antclaw.v1.AdminDataService/GetDataPreview` | `GetDataPreviewRequest` | `GetDataPreviewResponse` | 否 |

## AIService

- **proto 文件**：`proto/antclaw/v1/ai.proto`
- **范围**：客户端/通用
- **API 入口注册**：是

| RPC | 方法 | 端点 URL | 请求 | 响应 | 流式 |
|---|---|---|---|---|---|
| `Chat` | `POST` | `https://api.alfq.org/antclaw.v1.AIService/Chat` | `stream ChatRequest` | `ChatResponse` | 是 |
| `Interpret` | `POST` | `https://api.alfq.org/antclaw.v1.AIService/Interpret` | `InterpretRequest` | `InterpretResponse` | 否 |
| `Outlook` | `POST` | `https://api.alfq.org/antclaw.v1.AIService/Outlook` | `OutlookRequest` | `OutlookResponse` | 否 |
| `BuildContext` | `POST` | `https://api.alfq.org/antclaw.v1.AIService/BuildContext` | `BuildContextRequest` | `BuildContextResponse` | 否 |
| `RememberFact` | `POST` | `https://api.alfq.org/antclaw.v1.AIService/RememberFact` | `RememberFactRequest` | `RememberFactResponse` | 否 |
| `RecallFact` | `POST` | `https://api.alfq.org/antclaw.v1.AIService/RecallFact` | `RecallFactRequest` | `RecallFactResponse` | 否 |
| `SearchMemory` | `POST` | `https://api.alfq.org/antclaw.v1.AIService/SearchMemory` | `SearchMemoryRequest` | `SearchMemoryResponse` | 否 |
| `CheckRateLimit` | `POST` | `https://api.alfq.org/antclaw.v1.AIService/CheckRateLimit` | `CheckRateLimitRequest` | `CheckRateLimitResponse` | 否 |
| `RunWithTools` | `POST` | `https://api.alfq.org/antclaw.v1.AIService/RunWithTools` | `RunWithToolsRequest` | `RunWithToolsResponse` | 否 |

## AlertService

- **proto 文件**：`proto/antclaw/v1/alerts.proto`
- **范围**：客户端/通用
- **API 入口注册**：是

| RPC | 方法 | 端点 URL | 请求 | 响应 | 流式 |
|---|---|---|---|---|---|
| `ListSubscriptions` | `POST` | `https://api.alfq.org/antclaw.v1.AlertService/ListSubscriptions` | `ListSubscriptionsRequest` | `ListSubscriptionsResponse` | 否 |
| `Subscribe` | `POST` | `https://api.alfq.org/antclaw.v1.AlertService/Subscribe` | `SubscribeRequest` | `SubscribeResponse` | 否 |
| `Unsubscribe` | `POST` | `https://api.alfq.org/antclaw.v1.AlertService/Unsubscribe` | `UnsubscribeRequest` | `UnsubscribeResponse` | 否 |
| `RegisterWebhook` | `POST` | `https://api.alfq.org/antclaw.v1.AlertService/RegisterWebhook` | `RegisterWebhookRequest` | `RegisterWebhookResponse` | 否 |
| `ListWebhooks` | `POST` | `https://api.alfq.org/antclaw.v1.AlertService/ListWebhooks` | `ListWebhooksRequest` | `ListWebhooksResponse` | 否 |
| `CreateAlert` | `POST` | `https://api.alfq.org/antclaw.v1.AlertService/CreateAlert` | `CreateAlertRequest` | `CreateAlertResponse` | 否 |
| `ListAlerts` | `POST` | `https://api.alfq.org/antclaw.v1.AlertService/ListAlerts` | `ListAlertsRequest` | `ListAlertsResponse` | 否 |
| `UpdateAlert` | `POST` | `https://api.alfq.org/antclaw.v1.AlertService/UpdateAlert` | `UpdateAlertRequest` | `UpdateAlertResponse` | 否 |
| `DeleteAlert` | `POST` | `https://api.alfq.org/antclaw.v1.AlertService/DeleteAlert` | `DeleteAlertRequest` | `DeleteAlertResponse` | 否 |
| `ToggleAlert` | `POST` | `https://api.alfq.org/antclaw.v1.AlertService/ToggleAlert` | `ToggleAlertRequest` | `ToggleAlertResponse` | 否 |
| `DecideAlert` | `POST` | `https://api.alfq.org/antclaw.v1.AlertService/DecideAlert` | `DecideAlertRequest` | `DecideAlertResponse` | 否 |
| `GetPreferences` | `POST` | `https://api.alfq.org/antclaw.v1.AlertService/GetPreferences` | `GetPreferencesRequest` | `GetPreferencesResponse` | 否 |
| `UpdatePreferences` | `POST` | `https://api.alfq.org/antclaw.v1.AlertService/UpdatePreferences` | `UpdatePreferencesRequest` | `UpdatePreferencesResponse` | 否 |
| `SetUserTier` | `POST` | `https://api.alfq.org/antclaw.v1.AlertService/SetUserTier` | `SetUserTierRequest` | `SetUserTierResponse` | 否 |
| `GetAlertHistory` | `POST` | `https://api.alfq.org/antclaw.v1.AlertService/GetAlertHistory` | `GetAlertHistoryRequest` | `GetAlertHistoryResponse` | 否 |

## ChatService

- **proto 文件**：`proto/antclaw/v1/alfq_chat.proto`
- **范围**：客户端/通用
- **API 入口注册**：是

| RPC | 方法 | 端点 URL | 请求 | 响应 | 流式 |
|---|---|---|---|---|---|
| `SendMessage` | `POST` | `https://api.alfq.org/antclaw.v1.ChatService/SendMessage` | `SendMessageRequest` | `Message` | 否 |
| `GetConversation` | `POST` | `https://api.alfq.org/antclaw.v1.ChatService/GetConversation` | `GetConversationRequest` | `Conversation` | 否 |
| `ListConversations` | `POST` | `https://api.alfq.org/antclaw.v1.ChatService/ListConversations` | `ListConversationsRequest` | `ConversationList` | 否 |
| `MarkRead` | `POST` | `https://api.alfq.org/antclaw.v1.ChatService/MarkRead` | `ChatMarkReadRequest` | `ChatMarkReadResponse` | 否 |

## CircleService

- **proto 文件**：`proto/antclaw/v1/alfq_circle.proto`
- **范围**：客户端/通用
- **API 入口注册**：是

| RPC | 方法 | 端点 URL | 请求 | 响应 | 流式 |
|---|---|---|---|---|---|
| `CreateCircle` | `POST` | `https://api.alfq.org/antclaw.v1.CircleService/CreateCircle` | `CreateCircleRequest` | `Circle` | 否 |
| `JoinCircle` | `POST` | `https://api.alfq.org/antclaw.v1.CircleService/JoinCircle` | `JoinCircleRequest` | `Circle` | 否 |
| `LeaveCircle` | `POST` | `https://api.alfq.org/antclaw.v1.CircleService/LeaveCircle` | `LeaveCircleRequest` | `LeaveCircleResponse` | 否 |
| `GetCircleFeed` | `POST` | `https://api.alfq.org/antclaw.v1.CircleService/GetCircleFeed` | `GetCircleFeedRequest` | `CircleFeedResponse` | 否 |
| `ListCircles` | `POST` | `https://api.alfq.org/antclaw.v1.CircleService/ListCircles` | `ListCirclesRequest` | `CircleList` | 否 |

## FeedService

- **proto 文件**：`proto/antclaw/v1/alfq_feed.proto`
- **范围**：客户端/通用
- **API 入口注册**：是

| RPC | 方法 | 端点 URL | 请求 | 响应 | 流式 |
|---|---|---|---|---|---|
| `CreatePost` | `POST` | `https://api.alfq.org/antclaw.v1.FeedService/CreatePost` | `CreatePostRequest` | `Post` | 否 |
| `GetFeed` | `POST` | `https://api.alfq.org/antclaw.v1.FeedService/GetFeed` | `GetFeedRequest` | `FeedResponse` | 否 |
| `GetPost` | `POST` | `https://api.alfq.org/antclaw.v1.FeedService/GetPost` | `GetPostRequest` | `Post` | 否 |
| `LikePost` | `POST` | `https://api.alfq.org/antclaw.v1.FeedService/LikePost` | `LikePostRequest` | `Post` | 否 |
| `UnlikePost` | `POST` | `https://api.alfq.org/antclaw.v1.FeedService/UnlikePost` | `UnlikePostRequest` | `Post` | 否 |
| `CommentOnPost` | `POST` | `https://api.alfq.org/antclaw.v1.FeedService/CommentOnPost` | `CommentRequest` | `Comment` | 否 |
| `SharePost` | `POST` | `https://api.alfq.org/antclaw.v1.FeedService/SharePost` | `SharePostRequest` | `Post` | 否 |

## MarketplaceService

- **proto 文件**：`proto/antclaw/v1/alfq_marketplace.proto`
- **范围**：客户端/通用
- **API 入口注册**：是

| RPC | 方法 | 端点 URL | 请求 | 响应 | 流式 |
|---|---|---|---|---|---|
| `ListProducts` | `POST` | `https://api.alfq.org/antclaw.v1.MarketplaceService/ListProducts` | `ListProductsRequest` | `ProductList` | 否 |
| `PublishProduct` | `POST` | `https://api.alfq.org/antclaw.v1.MarketplaceService/PublishProduct` | `PublishProductRequest` | `Product` | 否 |
| `PurchaseProduct` | `POST` | `https://api.alfq.org/antclaw.v1.MarketplaceService/PurchaseProduct` | `PurchaseProductRequest` | `Purchase` | 否 |
| `GetMyProducts` | `POST` | `https://api.alfq.org/antclaw.v1.MarketplaceService/GetMyProducts` | `GetMyProductsRequest` | `ProductList` | 否 |
| `GetMyPurchases` | `POST` | `https://api.alfq.org/antclaw.v1.MarketplaceService/GetMyPurchases` | `GetMyPurchasesRequest` | `PurchaseList` | 否 |

## TraderService

- **proto 文件**：`proto/antclaw/v1/alfq_trader.proto`
- **范围**：客户端/通用
- **API 入口注册**：是

| RPC | 方法 | 端点 URL | 请求 | 响应 | 流式 |
|---|---|---|---|---|---|
| `GetProfile` | `POST` | `https://api.alfq.org/antclaw.v1.TraderService/GetProfile` | `GetTraderProfileRequest` | `TraderProfile` | 否 |
| `UpdateProfile` | `POST` | `https://api.alfq.org/antclaw.v1.TraderService/UpdateProfile` | `UpdateTraderProfileRequest` | `TraderProfile` | 否 |
| `Follow` | `POST` | `https://api.alfq.org/antclaw.v1.TraderService/Follow` | `FollowRequest` | `FollowResponse` | 否 |
| `Unfollow` | `POST` | `https://api.alfq.org/antclaw.v1.TraderService/Unfollow` | `UnfollowRequest` | `FollowResponse` | 否 |
| `GetFollowers` | `POST` | `https://api.alfq.org/antclaw.v1.TraderService/GetFollowers` | `GetFollowersRequest` | `UserList` | 否 |
| `GetFollowing` | `POST` | `https://api.alfq.org/antclaw.v1.TraderService/GetFollowing` | `GetFollowingRequest` | `UserList` | 否 |

## AuthService

- **proto 文件**：`proto/antclaw/v1/auth.proto`
- **范围**：客户端/通用
- **API 入口注册**：是

| RPC | 方法 | 端点 URL | 请求 | 响应 | 流式 |
|---|---|---|---|---|---|
| `Register` | `POST` | `https://api.alfq.org/antclaw.v1.AuthService/Register` | `RegisterRequest` | `RegisterResponse` | 否 |
| `Login` | `POST` | `https://api.alfq.org/antclaw.v1.AuthService/Login` | `LoginRequest` | `LoginResponse` | 否 |
| `Refresh` | `POST` | `https://api.alfq.org/antclaw.v1.AuthService/Refresh` | `RefreshRequest` | `RefreshResponse` | 否 |
| `Logout` | `POST` | `https://api.alfq.org/antclaw.v1.AuthService/Logout` | `LogoutRequest` | `LogoutResponse` | 否 |
| `RequestPasswordReset` | `POST` | `https://api.alfq.org/antclaw.v1.AuthService/RequestPasswordReset` | `RequestPasswordResetRequest` | `RequestPasswordResetResponse` | 否 |
| `ResetPassword` | `POST` | `https://api.alfq.org/antclaw.v1.AuthService/ResetPassword` | `ResetPasswordRequest` | `ResetPasswordResponse` | 否 |
| `VerifyEmail` | `POST` | `https://api.alfq.org/antclaw.v1.AuthService/VerifyEmail` | `VerifyEmailRequest` | `VerifyEmailResponse` | 否 |

## BacktestService

- **proto 文件**：`proto/antclaw/v1/backtest.proto`
- **范围**：客户端/通用
- **API 入口注册**：是

| RPC | 方法 | 端点 URL | 请求 | 响应 | 流式 |
|---|---|---|---|---|---|
| `RunBacktest` | `POST` | `https://api.alfq.org/antclaw.v1.BacktestService/RunBacktest` | `RunBacktestRequest` | `RunBacktestResponse` | 否 |
| `GetBacktest` | `POST` | `https://api.alfq.org/antclaw.v1.BacktestService/GetBacktest` | `GetBacktestRequest` | `GetBacktestResponse` | 否 |
| `GetAccuracy` | `POST` | `https://api.alfq.org/antclaw.v1.BacktestService/GetAccuracy` | `GetAccuracyRequest` | `GetAccuracyResponse` | 否 |
| `RunQuantBt` | `POST` | `https://api.alfq.org/antclaw.v1.BacktestService/RunQuantBt` | `RunQuantBtRequest` | `RunQuantBtResponse` | 否 |
| `RunVpBt` | `POST` | `https://api.alfq.org/antclaw.v1.BacktestService/RunVpBt` | `RunVpBtRequest` | `RunVpBtResponse` | 否 |
| `RunCtaBt` | `POST` | `https://api.alfq.org/antclaw.v1.BacktestService/RunCtaBt` | `RunCtaBtRequest` | `RunCtaBtResponse` | 否 |
| `RunWalkforward` | `POST` | `https://api.alfq.org/antclaw.v1.BacktestService/RunWalkforward` | `RunWalkforwardRequest` | `RunWalkforwardResponse` | 否 |
| `GetWalkforwardResult` | `POST` | `https://api.alfq.org/antclaw.v1.BacktestService/GetWalkforwardResult` | `GetWalkforwardResultRequest` | `GetWalkforwardResultResponse` | 否 |
| `RunBootstrap` | `POST` | `https://api.alfq.org/antclaw.v1.BacktestService/RunBootstrap` | `RunBootstrapRequest` | `RunBootstrapResponse` | 否 |
| `RunMonteCarlo` | `POST` | `https://api.alfq.org/antclaw.v1.BacktestService/RunMonteCarlo` | `RunMonteCarloRequest` | `RunMonteCarloResponse` | 否 |
| `GetTrades` | `POST` | `https://api.alfq.org/antclaw.v1.BacktestService/GetTrades` | `GetTradesRequest` | `GetTradesResponse` | 否 |
| `GetMetricsByRegime` | `POST` | `https://api.alfq.org/antclaw.v1.BacktestService/GetMetricsByRegime` | `GetMetricsByRegimeRequest` | `GetMetricsByRegimeResponse` | 否 |

## CalendarService

- **proto 文件**：`proto/antclaw/v1/calendar.proto`
- **范围**：客户端/通用
- **API 入口注册**：是

| RPC | 方法 | 端点 URL | 请求 | 响应 | 流式 |
|---|---|---|---|---|---|
| `ListEvents` | `POST` | `https://api.alfq.org/antclaw.v1.CalendarService/ListEvents` | `ListEventsRequest` | `ListEventsResponse` | 否 |
| `GetEvent` | `POST` | `https://api.alfq.org/antclaw.v1.CalendarService/GetEvent` | `GetEventRequest` | `GetEventResponse` | 否 |
| `GetImpact` | `POST` | `https://api.alfq.org/antclaw.v1.CalendarService/GetImpact` | `GetImpactRequest` | `GetImpactResponse` | 否 |
| `GetImpactHistory` | `POST` | `https://api.alfq.org/antclaw.v1.CalendarService/GetImpactHistory` | `GetImpactHistoryRequest` | `GetImpactHistoryResponse` | 否 |

## COTService

- **proto 文件**：`proto/antclaw/v1/cot.proto`
- **范围**：客户端/通用
- **API 入口注册**：是

| RPC | 方法 | 端点 URL | 请求 | 响应 | 流式 |
|---|---|---|---|---|---|
| `GetSummary` | `POST` | `https://api.alfq.org/antclaw.v1.COTService/GetSummary` | `GetSummaryRequest` | `GetSummaryResponse` | 否 |
| `Compare` | `POST` | `https://api.alfq.org/antclaw.v1.COTService/Compare` | `CompareRequest` | `CompareResponse` | 否 |
| `GetSignals` | `POST` | `https://api.alfq.org/antclaw.v1.COTService/GetSignals` | `GetSignalsRequest` | `GetSignalsResponse` | 否 |
| `GetHistory` | `POST` | `https://api.alfq.org/antclaw.v1.COTService/GetHistory` | `COTServiceGetHistoryRequest` | `COTServiceGetHistoryResponse` | 否 |
| `SubscribePairAlert` | `POST` | `https://api.alfq.org/antclaw.v1.COTService/SubscribePairAlert` | `SubscribePairAlertRequest` | `SubscribePairAlertResponse` | 否 |

## CryptoService

- **proto 文件**：`proto/antclaw/v1/crypto.proto`
- **范围**：客户端/通用
- **API 入口注册**：是

| RPC | 方法 | 端点 URL | 请求 | 响应 | 流式 |
|---|---|---|---|---|---|
| `GetCryptoPublicKey` | `POST` | `https://api.alfq.org/antclaw.v1.CryptoService/GetCryptoPublicKey` | `GetCryptoPublicKeyRequest` | `GetCryptoPublicKeyResponse` | 否 |
| `PostEnvelope` | `POST` | `https://api.alfq.org/antclaw.v1.CryptoService/PostEnvelope` | `PostEnvelopeRequest` | `PostEnvelopeResponse` | 否 |

## DataSourceService

- **proto 文件**：`proto/antclaw/v1/datasource.proto`
- **范围**：管理端
- **API 入口注册**：是

| RPC | 方法 | 端点 URL | 请求 | 响应 | 流式 |
|---|---|---|---|---|---|
| `ListDataSources` | `POST` | `https://api.alfq.org/antclaw.v1.DataSourceService/ListDataSources` | `ListDataSourcesRequest` | `ListDataSourcesResponse` | 否 |
| `UpdateDataSource` | `POST` | `https://api.alfq.org/antclaw.v1.DataSourceService/UpdateDataSource` | `UpdateDataSourceRequest` | `UpdateDataSourceResponse` | 否 |

## DeFiService

- **proto 文件**：`proto/antclaw/v1/defi.proto`
- **范围**：客户端/通用
- **API 入口注册**：是

| RPC | 方法 | 端点 URL | 请求 | 响应 | 流式 |
|---|---|---|---|---|---|
| `GetTopProtocols` | `POST` | `https://api.alfq.org/antclaw.v1.DeFiService/GetTopProtocols` | `GetTopProtocolsRequest` | `GetTopProtocolsResponse` | 否 |
| `GetProtocolTVL` | `POST` | `https://api.alfq.org/antclaw.v1.DeFiService/GetProtocolTVL` | `GetProtocolTVLRequest` | `GetProtocolTVLResponse` | 否 |
| `GetAnalysis` | `POST` | `https://api.alfq.org/antclaw.v1.DeFiService/GetAnalysis` | `DeFiServiceGetAnalysisRequest` | `DeFiServiceGetAnalysisResponse` | 否 |

## FedWatchService

- **proto 文件**：`proto/antclaw/v1/fedwatch.proto`
- **范围**：客户端/通用
- **API 入口注册**：是

| RPC | 方法 | 端点 URL | 请求 | 响应 | 流式 |
|---|---|---|---|---|---|
| `GetFOMCProbabilities` | `POST` | `https://api.alfq.org/antclaw.v1.FedWatchService/GetFOMCProbabilities` | `GetFOMCProbabilitiesRequest` | `GetFOMCProbabilitiesResponse` | 否 |

## MacroService

- **proto 文件**：`proto/antclaw/v1/macro.proto`
- **范围**：客户端/通用
- **API 入口注册**：是

| RPC | 方法 | 端点 URL | 请求 | 响应 | 流式 |
|---|---|---|---|---|---|
| `GetFred` | `POST` | `https://api.alfq.org/antclaw.v1.MacroService/GetFred` | `GetFredRequest` | `GetFredResponse` | 否 |
| `GetEcb` | `POST` | `https://api.alfq.org/antclaw.v1.MacroService/GetEcb` | `GetEcbRequest` | `GetEcbResponse` | 否 |
| `GetSnb` | `POST` | `https://api.alfq.org/antclaw.v1.MacroService/GetSnb` | `GetSnbRequest` | `GetSnbResponse` | 否 |
| `GetOecdLeading` | `POST` | `https://api.alfq.org/antclaw.v1.MacroService/GetOecdLeading` | `GetOecdLeadingRequest` | `GetOecdLeadingResponse` | 否 |
| `GetEurostat` | `POST` | `https://api.alfq.org/antclaw.v1.MacroService/GetEurostat` | `GetEurostatRequest` | `GetEurostatResponse` | 否 |
| `GetBis` | `POST` | `https://api.alfq.org/antclaw.v1.MacroService/GetBis` | `GetBisRequest` | `GetBisResponse` | 否 |
| `GetTradingEconomics` | `POST` | `https://api.alfq.org/antclaw.v1.MacroService/GetTradingEconomics` | `GetTradingEconomicsRequest` | `GetTradingEconomicsResponse` | 否 |
| `GetDtccSwaps` | `POST` | `https://api.alfq.org/antclaw.v1.MacroService/GetDtccSwaps` | `GetDtccSwapsRequest` | `GetDtccSwapsResponse` | 否 |
| `GetSec13f` | `POST` | `https://api.alfq.org/antclaw.v1.MacroService/GetSec13f` | `GetSec13fRequest` | `GetSec13fResponse` | 否 |
| `GetTreasuryAuctions` | `POST` | `https://api.alfq.org/antclaw.v1.MacroService/GetTreasuryAuctions` | `GetTreasuryAuctionsRequest` | `GetTreasuryAuctionsResponse` | 否 |
| `GetFedWatch` | `POST` | `https://api.alfq.org/antclaw.v1.MacroService/GetFedWatch` | `GetFedWatchRequest` | `GetFedWatchResponse` | 否 |
| `GetWorldBank` | `POST` | `https://api.alfq.org/antclaw.v1.MacroService/GetWorldBank` | `GetWorldBankRequest` | `GetWorldBankResponse` | 否 |
| `GetImfWeo` | `POST` | `https://api.alfq.org/antclaw.v1.MacroService/GetImfWeo` | `GetImfWeoRequest` | `GetImfWeoResponse` | 否 |

## MacroExtrasService

- **proto 文件**：`proto/antclaw/v1/macro_extras.proto`
- **范围**：客户端/通用
- **API 入口注册**：是

| RPC | 方法 | 端点 URL | 请求 | 响应 | 流式 |
|---|---|---|---|---|---|
| `GetSeries` | `POST` | `https://api.alfq.org/antclaw.v1.MacroExtrasService/GetSeries` | `MacroExtrasServiceGetSeriesRequest` | `MacroExtrasServiceGetSeriesResponse` | 否 |
| `ListAvailableSeries` | `POST` | `https://api.alfq.org/antclaw.v1.MacroExtrasService/ListAvailableSeries` | `ListAvailableSeriesRequest` | `ListAvailableSeriesResponse` | 否 |

## MT4Service

- **proto 文件**：`proto/antclaw/v1/mt4.proto`
- **范围**：客户端/通用
- **API 入口注册**：是

| RPC | 方法 | 端点 URL | 请求 | 响应 | 流式 |
|---|---|---|---|---|---|
| `AddAccount` | `POST` | `https://api.alfq.org/antclaw.v1.MT4Service/AddAccount` | `AddMT4AccountRequest` | `MT4Account` | 否 |
| `RemoveAccount` | `POST` | `https://api.alfq.org/antclaw.v1.MT4Service/RemoveAccount` | `RemoveMT4AccountRequest` | `RemoveMT4AccountResponse` | 否 |
| `GetAccountInfo` | `POST` | `https://api.alfq.org/antclaw.v1.MT4Service/GetAccountInfo` | `GetMT4AccountInfoRequest` | `MT4AccountInfo` | 否 |
| `GetPositions` | `POST` | `https://api.alfq.org/antclaw.v1.MT4Service/GetPositions` | `GetMT4PositionsRequest` | `MT4PositionsResponse` | 否 |
| `GetHistory` | `POST` | `https://api.alfq.org/antclaw.v1.MT4Service/GetHistory` | `GetMT4HistoryRequest` | `MT4HistoryResponse` | 否 |

## MT5Service

- **proto 文件**：`proto/antclaw/v1/mt5.proto`
- **范围**：客户端/通用
- **API 入口注册**：是

| RPC | 方法 | 端点 URL | 请求 | 响应 | 流式 |
|---|---|---|---|---|---|
| `AddAccount` | `POST` | `https://api.alfq.org/antclaw.v1.MT5Service/AddAccount` | `AddMT5AccountRequest` | `MT5Account` | 否 |
| `RemoveAccount` | `POST` | `https://api.alfq.org/antclaw.v1.MT5Service/RemoveAccount` | `RemoveMT5AccountRequest` | `RemoveMT5AccountResponse` | 否 |
| `GetAccountInfo` | `POST` | `https://api.alfq.org/antclaw.v1.MT5Service/GetAccountInfo` | `GetMT5AccountInfoRequest` | `MT5AccountInfo` | 否 |
| `GetPositions` | `POST` | `https://api.alfq.org/antclaw.v1.MT5Service/GetPositions` | `GetMT5PositionsRequest` | `MT5PositionsResponse` | 否 |
| `GetHistory` | `POST` | `https://api.alfq.org/antclaw.v1.MT5Service/GetHistory` | `GetMT5HistoryRequest` | `MT5HistoryResponse` | 否 |

## NotificationService

- **proto 文件**：`proto/antclaw/v1/notification.proto`
- **范围**：客户端/通用
- **API 入口注册**：是

| RPC | 方法 | 端点 URL | 请求 | 响应 | 流式 |
|---|---|---|---|---|---|
| `ListUnread` | `POST` | `https://api.alfq.org/antclaw.v1.NotificationService/ListUnread` | `ListUnreadRequest` | `ListUnreadResponse` | 否 |
| `ListHistory` | `POST` | `https://api.alfq.org/antclaw.v1.NotificationService/ListHistory` | `ListHistoryRequest` | `ListHistoryResponse` | 否 |
| `UnreadCount` | `POST` | `https://api.alfq.org/antclaw.v1.NotificationService/UnreadCount` | `UnreadCountRequest` | `UnreadCountResponse` | 否 |
| `MarkRead` | `POST` | `https://api.alfq.org/antclaw.v1.NotificationService/MarkRead` | `MarkReadRequest` | `MarkReadResponse` | 否 |
| `MarkAllRead` | `POST` | `https://api.alfq.org/antclaw.v1.NotificationService/MarkAllRead` | `MarkAllReadRequest` | `MarkAllReadResponse` | 否 |
| `GetPrefs` | `POST` | `https://api.alfq.org/antclaw.v1.NotificationService/GetPrefs` | `GetPrefsRequest` | `GetPrefsResponse` | 否 |
| `UpdatePrefs` | `POST` | `https://api.alfq.org/antclaw.v1.NotificationService/UpdatePrefs` | `UpdatePrefsRequest` | `UpdatePrefsResponse` | 否 |
| `GetAlertPrefs` | `POST` | `https://api.alfq.org/antclaw.v1.NotificationService/GetAlertPrefs` | `GetAlertPrefsRequest` | `GetAlertPrefsResponse` | 否 |
| `UpdateAlertPrefs` | `POST` | `https://api.alfq.org/antclaw.v1.NotificationService/UpdateAlertPrefs` | `UpdateAlertPrefsRequest` | `UpdateAlertPrefsResponse` | 否 |

## OnchainService

- **proto 文件**：`proto/antclaw/v1/onchain.proto`
- **范围**：客户端/通用
- **API 入口注册**：是

| RPC | 方法 | 端点 URL | 请求 | 响应 | 流式 |
|---|---|---|---|---|---|
| `GetMetrics` | `POST` | `https://api.alfq.org/antclaw.v1.OnchainService/GetMetrics` | `OnchainServiceGetMetricsRequest` | `OnchainServiceGetMetricsResponse` | 否 |
| `GetAnalysis` | `POST` | `https://api.alfq.org/antclaw.v1.OnchainService/GetAnalysis` | `OnchainServiceGetAnalysisRequest` | `OnchainServiceGetAnalysisResponse` | 否 |

## OptionsService

- **proto 文件**：`proto/antclaw/v1/options.proto`
- **范围**：客户端/通用
- **API 入口注册**：是

| RPC | 方法 | 端点 URL | 请求 | 响应 | 流式 |
|---|---|---|---|---|---|
| `GetGEX` | `POST` | `https://api.alfq.org/antclaw.v1.OptionsService/GetGEX` | `GetGEXRequest` | `GetGEXResponse` | 否 |
| `GetIVSurface` | `POST` | `https://api.alfq.org/antclaw.v1.OptionsService/GetIVSurface` | `GetIVSurfaceRequest` | `GetIVSurfaceResponse` | 否 |
| `GetOptionsSkew` | `POST` | `https://api.alfq.org/antclaw.v1.OptionsService/GetOptionsSkew` | `GetOptionsSkewRequest` | `GetOptionsSkewResponse` | 否 |
| `GetIVAlerts` | `POST` | `https://api.alfq.org/antclaw.v1.OptionsService/GetIVAlerts` | `GetIVAlertsRequest` | `GetIVAlertsResponse` | 否 |

## PriceService

- **proto 文件**：`proto/antclaw/v1/price.proto`
- **范围**：客户端/通用
- **API 入口注册**：是

| RPC | 方法 | 端点 URL | 请求 | 响应 | 流式 |
|---|---|---|---|---|---|
| `GetPrice` | `POST` | `https://api.alfq.org/antclaw.v1.PriceService/GetPrice` | `GetPriceRequest` | `GetPriceResponse` | 否 |
| `GetLevels` | `POST` | `https://api.alfq.org/antclaw.v1.PriceService/GetLevels` | `GetLevelsRequest` | `GetLevelsResponse` | 否 |
| `GetMarketOverview` | `POST` | `https://api.alfq.org/antclaw.v1.PriceService/GetMarketOverview` | `GetMarketOverviewRequest` | `GetMarketOverviewResponse` | 否 |
| `GetSession` | `POST` | `https://api.alfq.org/antclaw.v1.PriceService/GetSession` | `GetSessionRequest` | `GetSessionResponse` | 否 |
| `RunScenario` | `POST` | `https://api.alfq.org/antclaw.v1.PriceService/RunScenario` | `RunScenarioRequest` | `RunScenarioResponse` | 否 |
| `GetRegime` | `POST` | `https://api.alfq.org/antclaw.v1.PriceService/GetRegime` | `GetRegimeRequest` | `GetRegimeResponse` | 否 |
| `GetSeasonal` | `POST` | `https://api.alfq.org/antclaw.v1.PriceService/GetSeasonal` | `GetSeasonalRequest` | `GetSeasonalResponse` | 否 |
| `GetVolatility` | `POST` | `https://api.alfq.org/antclaw.v1.PriceService/GetVolatility` | `GetVolatilityRequest` | `GetVolatilityResponse` | 否 |
| `GetHurst` | `POST` | `https://api.alfq.org/antclaw.v1.PriceService/GetHurst` | `GetHurstRequest` | `GetHurstResponse` | 否 |
| `GetCorrelations` | `POST` | `https://api.alfq.org/antclaw.v1.PriceService/GetCorrelations` | `GetCorrelationsRequest` | `GetCorrelationsResponse` | 否 |
| `GetDivergences` | `POST` | `https://api.alfq.org/antclaw.v1.PriceService/GetDivergences` | `GetDivergencesRequest` | `GetDivergencesResponse` | 否 |

## RegimeService

- **proto 文件**：`proto/antclaw/v1/regime.proto`
- **范围**：客户端/通用
- **API 入口注册**：是

| RPC | 方法 | 端点 URL | 请求 | 响应 | 流式 |
|---|---|---|---|---|---|
| `GetOverlay` | `POST` | `https://api.alfq.org/antclaw.v1.RegimeService/GetOverlay` | `GetOverlayRequest` | `GetOverlayResponse` | 否 |
| `ListRecent` | `POST` | `https://api.alfq.org/antclaw.v1.RegimeService/ListRecent` | `ListRecentRequest` | `ListRecentResponse` | 否 |

## ReportService

- **proto 文件**：`proto/antclaw/v1/report.proto`
- **范围**：客户端/通用
- **API 入口注册**：是

| RPC | 方法 | 端点 URL | 请求 | 响应 | 流式 |
|---|---|---|---|---|---|
| `GetReport` | `POST` | `https://api.alfq.org/antclaw.v1.ReportService/GetReport` | `GetReportRequest` | `GetReportResponse` | 否 |

## SECService

- **proto 文件**：`proto/antclaw/v1/sec.proto`
- **范围**：客户端/通用
- **API 入口注册**：是

| RPC | 方法 | 端点 URL | 请求 | 响应 | 流式 |
|---|---|---|---|---|---|
| `ListFilings` | `POST` | `https://api.alfq.org/antclaw.v1.SECService/ListFilings` | `ListFilingsRequest` | `ListFilingsResponse` | 否 |
| `GetFiling` | `POST` | `https://api.alfq.org/antclaw.v1.SECService/GetFiling` | `GetFilingRequest` | `GetFilingResponse` | 否 |
| `GetAnalysis` | `POST` | `https://api.alfq.org/antclaw.v1.SECService/GetAnalysis` | `SECServiceGetAnalysisRequest` | `SECServiceGetAnalysisResponse` | 否 |

## SentimentService

- **proto 文件**：`proto/antclaw/v1/sentiment.proto`
- **范围**：客户端/通用
- **API 入口注册**：是

| RPC | 方法 | 端点 URL | 请求 | 响应 | 流式 |
|---|---|---|---|---|---|
| `GetSentiment` | `POST` | `https://api.alfq.org/antclaw.v1.SentimentService/GetSentiment` | `GetSentimentRequest` | `GetSentimentResponse` | 否 |
| `GetOnchain` | `POST` | `https://api.alfq.org/antclaw.v1.SentimentService/GetOnchain` | `GetOnchainRequest` | `GetOnchainResponse` | 否 |
| `GetDefiHealth` | `POST` | `https://api.alfq.org/antclaw.v1.SentimentService/GetDefiHealth` | `GetDefiHealthRequest` | `GetDefiHealthResponse` | 否 |
| `GetCarryMonitor` | `POST` | `https://api.alfq.org/antclaw.v1.SentimentService/GetCarryMonitor` | `GetCarryMonitorRequest` | `GetCarryMonitorResponse` | 否 |

## SentimentExtrasService

- **proto 文件**：`proto/antclaw/v1/sentiment_extras.proto`
- **范围**：客户端/通用
- **API 入口注册**：是

| RPC | 方法 | 端点 URL | 请求 | 响应 | 流式 |
|---|---|---|---|---|---|
| `GetCBOEPutCall` | `POST` | `https://api.alfq.org/antclaw.v1.SentimentExtrasService/GetCBOEPutCall` | `GetCBOEPutCallRequest` | `GetCBOEPutCallResponse` | 否 |
| `GetMyFXBookPositions` | `POST` | `https://api.alfq.org/antclaw.v1.SentimentExtrasService/GetMyFXBookPositions` | `GetMyFXBookPositionsRequest` | `GetMyFXBookPositionsResponse` | 否 |
| `GetInsiderTrades` | `POST` | `https://api.alfq.org/antclaw.v1.SentimentExtrasService/GetInsiderTrades` | `GetInsiderTradesRequest` | `GetInsiderTradesResponse` | 否 |
| `GetCryptoSocial` | `POST` | `https://api.alfq.org/antclaw.v1.SentimentExtrasService/GetCryptoSocial` | `GetCryptoSocialRequest` | `GetCryptoSocialResponse` | 否 |
| `GetFinvizMetrics` | `POST` | `https://api.alfq.org/antclaw.v1.SentimentExtrasService/GetFinvizMetrics` | `GetFinvizMetricsRequest` | `GetFinvizMetricsResponse` | 否 |

## SignalsService

- **proto 文件**：`proto/antclaw/v1/signals.proto`
- **范围**：客户端/通用
- **API 入口注册**：是

| RPC | 方法 | 端点 URL | 请求 | 响应 | 流式 |
|---|---|---|---|---|---|
| `GetBias` | `POST` | `https://api.alfq.org/antclaw.v1.SignalsService/GetBias` | `GetBiasRequest` | `GetBiasResponse` | 否 |
| `GetRank` | `POST` | `https://api.alfq.org/antclaw.v1.SignalsService/GetRank` | `GetRankRequest` | `GetRankResponse` | 否 |
| `GetXFactors` | `POST` | `https://api.alfq.org/antclaw.v1.SignalsService/GetXFactors` | `GetXFactorsRequest` | `GetXFactorsResponse` | 否 |
| `GetRadar` | `POST` | `https://api.alfq.org/antclaw.v1.SignalsService/GetRadar` | `GetRadarRequest` | `GetRadarResponse` | 否 |
| `GetIntensity` | `POST` | `https://api.alfq.org/antclaw.v1.SignalsService/GetIntensity` | `GetIntensityRequest` | `GetIntensityResponse` | 否 |
| `GetTransition` | `POST` | `https://api.alfq.org/antclaw.v1.SignalsService/GetTransition` | `GetTransitionRequest` | `GetTransitionResponse` | 否 |
| `GetCryptoAlpha` | `POST` | `https://api.alfq.org/antclaw.v1.SignalsService/GetCryptoAlpha` | `GetCryptoAlphaRequest` | `GetCryptoAlphaResponse` | 否 |
| `GetUnified` | `POST` | `https://api.alfq.org/antclaw.v1.SignalsService/GetUnified` | `GetUnifiedRequest` | `GetUnifiedResponse` | 否 |
| `GetQuant` | `POST` | `https://api.alfq.org/antclaw.v1.SignalsService/GetQuant` | `GetQuantRequest` | `GetQuantResponse` | 否 |
| `GetCta` | `POST` | `https://api.alfq.org/antclaw.v1.SignalsService/GetCta` | `GetCtaRequest` | `GetCtaResponse` | 否 |
| `GetBriefing` | `POST` | `https://api.alfq.org/antclaw.v1.SignalsService/GetBriefing` | `GetBriefingRequest` | `GetBriefingResponse` | 否 |
| `GetOutlook` | `POST` | `https://api.alfq.org/antclaw.v1.SignalsService/GetOutlook` | `GetOutlookRequest` | `GetOutlookResponse` | 否 |
| `FitCalibration` | `POST` | `https://api.alfq.org/antclaw.v1.SignalsService/FitCalibration` | `FitCalibrationRequest` | `FitCalibrationResponse` | 否 |
| `PredictCalibrated` | `POST` | `https://api.alfq.org/antclaw.v1.SignalsService/PredictCalibrated` | `PredictCalibratedRequest` | `PredictCalibratedResponse` | 否 |
| `ListCalibrations` | `POST` | `https://api.alfq.org/antclaw.v1.SignalsService/ListCalibrations` | `ListCalibrationsRequest` | `ListCalibrationsResponse` | 否 |

## StrategyService

- **proto 文件**：`proto/antclaw/v1/strategy.proto`
- **范围**：客户端/通用
- **API 入口注册**：是

| RPC | 方法 | 端点 URL | 请求 | 响应 | 流式 |
|---|---|---|---|---|---|
| `ListStrategies` | `POST` | `https://api.alfq.org/antclaw.v1.StrategyService/ListStrategies` | `ListStrategiesRequest` | `ListStrategiesResponse` | 否 |
| `GetStrategy` | `POST` | `https://api.alfq.org/antclaw.v1.StrategyService/GetStrategy` | `GetStrategyRequest` | `GetStrategyResponse` | 否 |
| `CreateStrategy` | `POST` | `https://api.alfq.org/antclaw.v1.StrategyService/CreateStrategy` | `CreateStrategyRequest` | `CreateStrategyResponse` | 否 |
| `UpdateStrategy` | `POST` | `https://api.alfq.org/antclaw.v1.StrategyService/UpdateStrategy` | `UpdateStrategyRequest` | `UpdateStrategyResponse` | 否 |
| `DeleteStrategy` | `POST` | `https://api.alfq.org/antclaw.v1.StrategyService/DeleteStrategy` | `DeleteStrategyRequest` | `DeleteStrategyResponse` | 否 |
| `EnableStrategy` | `POST` | `https://api.alfq.org/antclaw.v1.StrategyService/EnableStrategy` | `EnableStrategyRequest` | `EnableStrategyResponse` | 否 |
| `DisableStrategy` | `POST` | `https://api.alfq.org/antclaw.v1.StrategyService/DisableStrategy` | `DisableStrategyRequest` | `DisableStrategyResponse` | 否 |
| `RunStrategy` | `POST` | `https://api.alfq.org/antclaw.v1.StrategyService/RunStrategy` | `RunStrategyRequest` | `RunStrategyResponse` | 否 |
| `ListStrategyRuns` | `POST` | `https://api.alfq.org/antclaw.v1.StrategyService/ListStrategyRuns` | `ListStrategyRunsRequest` | `ListStrategyRunsResponse` | 否 |

## StreamService

- **proto 文件**：`proto/antclaw/v1/stream.proto`
- **范围**：客户端/通用
- **API 入口注册**：否/需确认

| RPC | 方法 | 端点 URL | 请求 | 响应 | 流式 |
|---|---|---|---|---|---|
| `SubscribeEvents` | `POST` | `https://api.alfq.org/antclaw.v1.StreamService/SubscribeEvents` | `SubscribeEventsRequest` | `SubscribeEventsResponse` | 是 |

## SystemService

- **proto 文件**：`proto/antclaw/v1/system.proto`
- **范围**：客户端/通用
- **API 入口注册**：是

| RPC | 方法 | 端点 URL | 请求 | 响应 | 流式 |
|---|---|---|---|---|---|
| `Healthz` | `POST` | `https://api.alfq.org/antclaw.v1.SystemService/Healthz` | `HealthzRequest` | `HealthzResponse` | 否 |
| `Readyz` | `POST` | `https://api.alfq.org/antclaw.v1.SystemService/Readyz` | `ReadyzRequest` | `ReadyzResponse` | 否 |
| `Info` | `POST` | `https://api.alfq.org/antclaw.v1.SystemService/Info` | `InfoRequest` | `InfoResponse` | 否 |

## SystemAIService

- **proto 文件**：`proto/antclaw/v1/system_ai.proto`
- **范围**：管理端
- **API 入口注册**：是

| RPC | 方法 | 端点 URL | 请求 | 响应 | 流式 |
|---|---|---|---|---|---|
| `ListConfigs` | `POST` | `https://api.alfq.org/antclaw.v1.SystemAIService/ListConfigs` | `ListSystemAIConfigsRequest` | `ListSystemAIConfigsResponse` | 否 |
| `GetConfig` | `POST` | `https://api.alfq.org/antclaw.v1.SystemAIService/GetConfig` | `GetSystemAIConfigRequest` | `GetSystemAIConfigResponse` | 否 |
| `UpdateConfig` | `POST` | `https://api.alfq.org/antclaw.v1.SystemAIService/UpdateConfig` | `UpdateSystemAIConfigRequest` | `UpdateSystemAIConfigResponse` | 否 |
| `UpdateSecret` | `POST` | `https://api.alfq.org/antclaw.v1.SystemAIService/UpdateSecret` | `UpdateSystemAISecretRequest` | `UpdateSystemAISecretResponse` | 否 |
| `DiscoverModels` | `POST` | `https://api.alfq.org/antclaw.v1.SystemAIService/DiscoverModels` | `DiscoverSystemAIModelsRequest` | `DiscoverSystemAIModelsResponse` | 否 |
| `ValidateConnection` | `POST` | `https://api.alfq.org/antclaw.v1.SystemAIService/ValidateConnection` | `ValidateSystemAIConnectionRequest` | `ValidateSystemAIConnectionResponse` | 否 |

## TAService

- **proto 文件**：`proto/antclaw/v1/ta.proto`
- **范围**：客户端/通用
- **API 入口注册**：是

| RPC | 方法 | 端点 URL | 请求 | 响应 | 流式 |
|---|---|---|---|---|---|
| `GetIndicators` | `POST` | `https://api.alfq.org/antclaw.v1.TAService/GetIndicators` | `GetIndicatorsRequest` | `GetIndicatorsResponse` | 否 |
| `GetElliott` | `POST` | `https://api.alfq.org/antclaw.v1.TAService/GetElliott` | `GetElliottRequest` | `GetElliottResponse` | 否 |
| `GetWyckoff` | `POST` | `https://api.alfq.org/antclaw.v1.TAService/GetWyckoff` | `GetWyckoffRequest` | `GetWyckoffResponse` | 否 |
| `GetIct` | `POST` | `https://api.alfq.org/antclaw.v1.TAService/GetIct` | `GetIctRequest` | `GetIctResponse` | 否 |
| `GetAmt` | `POST` | `https://api.alfq.org/antclaw.v1.TAService/GetAmt` | `GetAmtRequest` | `GetAmtResponse` | 否 |
| `GetOrderflow` | `POST` | `https://api.alfq.org/antclaw.v1.TAService/GetOrderflow` | `GetOrderflowRequest` | `GetOrderflowResponse` | 否 |
| `GetVolumeProfile` | `POST` | `https://api.alfq.org/antclaw.v1.TAService/GetVolumeProfile` | `GetVolumeProfileRequest` | `GetVolumeProfileResponse` | 否 |
| `GetIntermarket` | `POST` | `https://api.alfq.org/antclaw.v1.TAService/GetIntermarket` | `GetIntermarketRequest` | `GetIntermarketResponse` | 否 |

## TreasuryService

- **proto 文件**：`proto/antclaw/v1/treasury.proto`
- **范围**：客户端/通用
- **API 入口注册**：是

| RPC | 方法 | 端点 URL | 请求 | 响应 | 流式 |
|---|---|---|---|---|---|
| `GetCurve` | `POST` | `https://api.alfq.org/antclaw.v1.TreasuryService/GetCurve` | `GetCurveRequest` | `GetCurveResponse` | 否 |
| `GetAnalysis` | `POST` | `https://api.alfq.org/antclaw.v1.TreasuryService/GetAnalysis` | `TreasuryServiceGetAnalysisRequest` | `TreasuryServiceGetAnalysisResponse` | 否 |

## UserService

- **proto 文件**：`proto/antclaw/v1/user.proto`
- **范围**：客户端/通用
- **API 入口注册**：是

| RPC | 方法 | 端点 URL | 请求 | 响应 | 流式 |
|---|---|---|---|---|---|
| `GetMe` | `POST` | `https://api.alfq.org/antclaw.v1.UserService/GetMe` | `GetMeRequest` | `GetMeResponse` | 否 |
| `UpdateSettings` | `POST` | `https://api.alfq.org/antclaw.v1.UserService/UpdateSettings` | `UpdateSettingsRequest` | `UpdateSettingsResponse` | 否 |
| `GetMembership` | `POST` | `https://api.alfq.org/antclaw.v1.UserService/GetMembership` | `GetMembershipRequest` | `GetMembershipResponse` | 否 |
| `StartOnboarding` | `POST` | `https://api.alfq.org/antclaw.v1.UserService/StartOnboarding` | `StartOnboardingRequest` | `StartOnboardingResponse` | 否 |
| `GetHistory` | `POST` | `https://api.alfq.org/antclaw.v1.UserService/GetHistory` | `GetHistoryRequest` | `GetHistoryResponse` | 否 |
| `ClearHistory` | `POST` | `https://api.alfq.org/antclaw.v1.UserService/ClearHistory` | `ClearHistoryRequest` | `ClearHistoryResponse` | 否 |
| `ListPins` | `POST` | `https://api.alfq.org/antclaw.v1.UserService/ListPins` | `ListPinsRequest` | `ListPinsResponse` | 否 |
| `Pin` | `POST` | `https://api.alfq.org/antclaw.v1.UserService/Pin` | `PinRequest` | `PinResponse` | 否 |
| `Unpin` | `POST` | `https://api.alfq.org/antclaw.v1.UserService/Unpin` | `UnpinRequest` | `UnpinResponse` | 否 |
| `SubmitFeedback` | `POST` | `https://api.alfq.org/antclaw.v1.UserService/SubmitFeedback` | `SubmitFeedbackRequest` | `SubmitFeedbackResponse` | 否 |
| `SetAiKey` | `POST` | `https://api.alfq.org/antclaw.v1.UserService/SetAiKey` | `SetAiKeyRequest` | `SetAiKeyResponse` | 否 |

## VolService

- **proto 文件**：`proto/antclaw/v1/vol.proto`
- **范围**：客户端/通用
- **API 入口注册**：是

| RPC | 方法 | 端点 URL | 请求 | 响应 | 流式 |
|---|---|---|---|---|---|
| `GetVix` | `POST` | `https://api.alfq.org/antclaw.v1.VolService/GetVix` | `GetVixRequest` | `GetVixResponse` | 否 |
| `GetMove` | `POST` | `https://api.alfq.org/antclaw.v1.VolService/GetMove` | `GetMoveRequest` | `GetMoveResponse` | 否 |
| `GetDvol` | `POST` | `https://api.alfq.org/antclaw.v1.VolService/GetDvol` | `GetDvolRequest` | `GetDvolResponse` | 否 |
| `GetGex` | `POST` | `https://api.alfq.org/antclaw.v1.VolService/GetGex` | `GetGexRequest` | `GetGexResponse` | 否 |
| `GetIvol` | `POST` | `https://api.alfq.org/antclaw.v1.VolService/GetIvol` | `GetIvolRequest` | `GetIvolResponse` | 否 |
| `GetSkew` | `POST` | `https://api.alfq.org/antclaw.v1.VolService/GetSkew` | `GetSkewRequest` | `GetSkewResponse` | 否 |
| `GetSkewVixAlert` | `POST` | `https://api.alfq.org/antclaw.v1.VolService/GetSkewVixAlert` | `GetSkewVixAlertRequest` | `GetSkewVixAlertResponse` | 否 |

