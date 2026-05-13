# API 字段参考

## 阅读说明

本文件按 Service/RPC 展开请求与响应字段。字段来自 `proto/antclaw/v1/*.proto` 自动扫描。

- **是否必填**：proto3 本身不表达 required；本文件将其标记为“业务必填以服务端校验为准”。
- **字段约束**：优先采集 proto 注释；无注释时需以 handler 校验和业务文档为准。
- **JSON 示例**：用于客户端理解结构，不表示真实业务数据。

## AdminService

### AdminService/ListUsers

- **URL**：`POST https://api.alfq.org/antclaw.v1.AdminService/ListUsers`
- **请求消息**：`ListUsersRequest`
- **响应消息**：`ListUsersResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `cursor` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `page_size` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `email_filter` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `role_filter` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `banned_only` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `users` | `repeated User` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `next_cursor` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `total` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "cursor": "<string>",
  "page_size": 0,
  "email_filter": "<string>",
  "role_filter": "<string>",
  "banned_only": false
}
```

#### 响应示例

```json
{
  "users": [
    {
      "user_id": "<string>",
      "email": "<string>",
      "username": "<string>",
      "display_name": "<string>",
      "locale": "LOCALE_UNSPECIFIED",
      "timezone": "<string>",
      "roles": [
        "<string>"
      ],
      "email_verified": false,
      "created_at": 0,
      "updated_at": 0,
      "code_id": "<string>"
    }
  ],
  "next_cursor": "<string>",
  "total": 0
}
```

### AdminService/SetRole

- **URL**：`POST https://api.alfq.org/antclaw.v1.AdminService/SetRole`
- **请求消息**：`SetRoleRequest`
- **响应消息**：`SetRoleResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `user_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `roles` | `repeated string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `user` | `User` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "user_id": "<string>",
  "roles": [
    "<string>"
  ]
}
```

#### 响应示例

```json
{
  "user": {
    "user_id": "<string>",
    "email": "<string>",
    "username": "<string>",
    "display_name": "<string>",
    "locale": "LOCALE_UNSPECIFIED",
    "timezone": "<string>",
    "roles": [
      "<string>"
    ],
    "email_verified": false,
    "created_at": 0,
    "updated_at": 0,
    "code_id": "<string>"
  }
}
```

### AdminService/Ban

- **URL**：`POST https://api.alfq.org/antclaw.v1.AdminService/Ban`
- **请求消息**：`BanRequest`
- **响应消息**：`BanResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `user_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `reason` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `expires_at` | `int64` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 0 = permanent |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{
  "user_id": "<string>",
  "reason": "<string>",
  "expires_at": 0
}
```

#### 响应示例

```json
{}
```

### AdminService/Unban

- **URL**：`POST https://api.alfq.org/antclaw.v1.AdminService/Unban`
- **请求消息**：`UnbanRequest`
- **响应消息**：`UnbanResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `user_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{
  "user_id": "<string>"
}
```

#### 响应示例

```json
{}
```

### AdminService/RunJob

- **URL**：`POST https://api.alfq.org/antclaw.v1.AdminService/RunJob`
- **请求消息**：`RunJobRequest`
- **响应消息**：`RunJobResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `job_name` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `job_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `status` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "job_name": "<string>"
}
```

#### 响应示例

```json
{
  "job_id": "<string>",
  "status": "<string>"
}
```

### AdminService/ListJobs

- **URL**：`POST https://api.alfq.org/antclaw.v1.AdminService/ListJobs`
- **请求消息**：`ListJobsRequest`
- **响应消息**：`ListJobsResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `status_filter` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `jobs` | `repeated Job` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "status_filter": "<string>"
}
```

#### 响应示例

```json
{
  "jobs": [
    {
      "job_id": "<string>",
      "job_name": "<string>",
      "status": "<string>",
      "last_run": "<string>",
      "next_run": "<string>",
      "enabled": false,
      "last_error": "<string>"
    }
  ]
}
```

### AdminService/SetJobEnabled

- **URL**：`POST https://api.alfq.org/antclaw.v1.AdminService/SetJobEnabled`
- **请求消息**：`SetJobEnabledRequest`
- **响应消息**：`SetJobEnabledResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `job_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `enabled` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `job_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `enabled` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "job_id": "<string>",
  "enabled": false
}
```

#### 响应示例

```json
{
  "job_id": "<string>",
  "enabled": false
}
```

### AdminService/ListAuditLogs

- **URL**：`POST https://api.alfq.org/antclaw.v1.AdminService/ListAuditLogs`
- **请求消息**：`ListAuditLogsRequest`
- **响应消息**：`ListAuditLogsResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `cursor` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `page_size` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `user_id_filter` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `action_filter` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `time_range` | `TimeRange` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `entries` | `repeated AuditLogEntry` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `next_cursor` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "cursor": "<string>",
  "page_size": 0,
  "user_id_filter": "<string>",
  "action_filter": "<string>",
  "time_range": {
    "start": "<string>",
    "end": "<string>"
  }
}
```

#### 响应示例

```json
{
  "entries": [
    {
      "log_id": "<string>",
      "user_id": "<string>",
      "action": "<string>",
      "resource": "<string>",
      "details": "<string>",
      "created_at": 0,
      "ip_address": "<string>"
    }
  ],
  "next_cursor": "<string>"
}
```

### AdminService/ListWebhookDeliveries

- **URL**：`POST https://api.alfq.org/antclaw.v1.AdminService/ListWebhookDeliveries`
- **请求消息**：`ListWebhookDeliveriesRequest`
- **响应消息**：`ListWebhookDeliveriesResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `cursor` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `page_size` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `webhook_id_filter` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `deliveries` | `repeated WebhookDelivery` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `next_cursor` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "cursor": "<string>",
  "page_size": 0,
  "webhook_id_filter": "<string>"
}
```

#### 响应示例

```json
{
  "deliveries": [
    {
      "delivery_id": "<string>",
      "webhook_id": "<string>",
      "event_type": "<string>",
      "status_code": 0,
      "success": false,
      "created_at": 0
    }
  ],
  "next_cursor": "<string>"
}
```

### AdminService/ForceLogout

- **URL**：`POST https://api.alfq.org/antclaw.v1.AdminService/ForceLogout`
- **请求消息**：`ForceLogoutRequest`
- **响应消息**：`ForceLogoutResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `user_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{
  "user_id": "<string>"
}
```

#### 响应示例

```json
{}
```

### AdminService/AdminResetUserPassword

- **URL**：`POST https://api.alfq.org/antclaw.v1.AdminService/AdminResetUserPassword`
- **请求消息**：`AdminResetUserPasswordRequest`
- **响应消息**：`AdminResetUserPasswordResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `user_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `new_password` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{
  "user_id": "<string>",
  "new_password": "<string>"
}
```

#### 响应示例

```json
{}
```

### AdminService/SetUserCodeID

- **URL**：`POST https://api.alfq.org/antclaw.v1.AdminService/SetUserCodeID`
- **请求消息**：`SetUserCodeIDRequest`
- **响应消息**：`SetUserCodeIDResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `user_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `code_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `code_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "user_id": "<string>",
  "code_id": "<string>"
}
```

#### 响应示例

```json
{
  "code_id": "<string>"
}
```

## AdminDataService

### AdminDataService/GetDataSummary

- **URL**：`POST https://api.alfq.org/antclaw.v1.AdminDataService/GetDataSummary`
- **请求消息**：`GetDataSummaryRequest`
- **响应消息**：`GetDataSummaryResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `items` | `repeated DataSummaryItem` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `updated_at` | `int64` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{}
```

#### 响应示例

```json
{
  "items": [
    {
      "job_id": "<string>",
      "name": "<string>",
      "table": "<string>",
      "count": 0,
      "latest_time": 0,
      "error": "<string>"
    }
  ],
  "updated_at": 0
}
```

### AdminDataService/GetDataPreview

- **URL**：`POST https://api.alfq.org/antclaw.v1.AdminDataService/GetDataPreview`
- **请求消息**：`GetDataPreviewRequest`
- **响应消息**：`GetDataPreviewResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `job_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `limit` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `job_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `table` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `time_col` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `columns` | `repeated string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `rows_json` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `total_sampled` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "job_id": "<string>",
  "limit": 0
}
```

#### 响应示例

```json
{
  "job_id": "<string>",
  "table": "<string>",
  "time_col": "<string>",
  "columns": [
    "<string>"
  ],
  "rows_json": "<string>",
  "total_sampled": 0
}
```

## AIService

### AIService/Chat

- **URL**：`POST https://api.alfq.org/antclaw.v1.AIService/Chat`
- **请求消息**：`stream ChatRequest`
- **响应消息**：`ChatResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

未在 proto 中解析到 `stream ChatRequest` 字段。

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `session_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `chunk` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `done` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `full_message` | `ChatMessage` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{}
```

#### 响应示例

```json
{
  "session_id": "<string>",
  "chunk": "<string>",
  "done": false,
  "full_message": {
    "role": "<string>",
    "content": "<string>",
    "timestamp": 0
  }
}
```

### AIService/Interpret

- **URL**：`POST https://api.alfq.org/antclaw.v1.AIService/Interpret`
- **请求消息**：`InterpretRequest`
- **响应消息**：`InterpretResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `data_type` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | price, cot, macro, etc. |
| `raw_data` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `question` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `locale` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `interpretation` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `key_points` | `repeated string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `confidence` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `cache_hit` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 是否命中 Redis 缓存 |

#### 请求示例

```json
{
  "data_type": "<string>",
  "raw_data": "<string>",
  "question": "<string>",
  "locale": "<string>"
}
```

#### 响应示例

```json
{
  "interpretation": "<string>",
  "key_points": [
    "<string>"
  ],
  "confidence": 0.0,
  "cache_hit": false
}
```

### AIService/Outlook

- **URL**：`POST https://api.alfq.org/antclaw.v1.AIService/Outlook`
- **请求消息**：`OutlookRequest`
- **响应消息**：`OutlookResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `timeframe` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `locale` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `summary` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `bullish_case` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `bearish_case` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `key_levels` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `generated_at` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "pair": "<string>",
  "timeframe": "<string>",
  "locale": "<string>"
}
```

#### 响应示例

```json
{
  "pair": "<string>",
  "summary": "<string>",
  "bullish_case": "<string>",
  "bearish_case": "<string>",
  "key_levels": "<string>",
  "generated_at": "<string>"
}
```

### AIService/BuildContext

- **URL**：`POST https://api.alfq.org/antclaw.v1.AIService/BuildContext`
- **请求消息**：`BuildContextRequest`
- **响应消息**：`BuildContextResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `asset` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | BTC / ETH / EURUSD / ... |
| `scope` | `repeated string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `locale` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `asset` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `prompt_ready` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `generated_at` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "asset": "<string>",
  "scope": [
    "<string>"
  ],
  "locale": "<string>"
}
```

#### 响应示例

```json
{
  "asset": "<string>",
  "prompt_ready": "<string>",
  "generated_at": "<string>"
}
```

### AIService/RememberFact

- **URL**：`POST https://api.alfq.org/antclaw.v1.AIService/RememberFact`
- **请求消息**：`RememberFactRequest`
- **响应消息**：`RememberFactResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `user_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `scope` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 'global' / 'thread' |
| `key` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `value` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `ttl_seconds` | `int64` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 0 = 永久 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{
  "user_id": "<string>",
  "scope": "<string>",
  "key": "<string>",
  "value": "<string>",
  "ttl_seconds": 0
}
```

#### 响应示例

```json
{}
```

### AIService/RecallFact

- **URL**：`POST https://api.alfq.org/antclaw.v1.AIService/RecallFact`
- **请求消息**：`RecallFactRequest`
- **响应消息**：`RecallFactResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `user_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `scope` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `key` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `value` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `created_at` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "user_id": "<string>",
  "scope": "<string>",
  "key": "<string>"
}
```

#### 响应示例

```json
{
  "id": "<string>",
  "value": "<string>",
  "created_at": "<string>"
}
```

### AIService/SearchMemory

- **URL**：`POST https://api.alfq.org/antclaw.v1.AIService/SearchMemory`
- **请求消息**：`SearchMemoryRequest`
- **响应消息**：`SearchMemoryResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `user_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `query` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `limit` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{
  "user_id": "<string>",
  "query": "<string>",
  "limit": 0
}
```

#### 响应示例

```json
{}
```

### AIService/CheckRateLimit

- **URL**：`POST https://api.alfq.org/antclaw.v1.AIService/CheckRateLimit`
- **请求消息**：`CheckRateLimitRequest`
- **响应消息**：`CheckRateLimitResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `user_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `provider` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 留空表示总配额 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `used_today` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `max_per_day` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `remaining` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `allowed` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "user_id": "<string>",
  "provider": "<string>"
}
```

#### 响应示例

```json
{
  "used_today": 0,
  "max_per_day": 0,
  "remaining": 0,
  "allowed": false
}
```

### AIService/RunWithTools

- **URL**：`POST https://api.alfq.org/antclaw.v1.AIService/RunWithTools`
- **请求消息**：`RunWithToolsRequest`
- **响应消息**：`RunWithToolsResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `user_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `thread_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 可空，自动创建 |
| `message` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `tools` | `repeated string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `max_hops` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 默认 5 |
| `model` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `provider_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `thread_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `answer` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `calls` | `repeated ToolCall` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `cache_hit` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `model` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `provider_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `prompt_tokens` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `completion_tokens` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `total_tokens` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "user_id": "<string>",
  "thread_id": "<string>",
  "message": "<string>",
  "tools": [
    "<string>"
  ],
  "max_hops": 0,
  "model": "<string>",
  "provider_id": "<string>"
}
```

#### 响应示例

```json
{
  "thread_id": "<string>",
  "answer": "<string>",
  "calls": [
    {
      "name": "<string>",
      "args_json": "<string>",
      "result_json": "<string>",
      "error": "<string>"
    }
  ],
  "cache_hit": false,
  "model": "<string>",
  "provider_id": "<string>",
  "prompt_tokens": 0,
  "completion_tokens": 0,
  "total_tokens": 0
}
```

## AlertService

### AlertService/ListSubscriptions

- **URL**：`POST https://api.alfq.org/antclaw.v1.AlertService/ListSubscriptions`
- **请求消息**：`ListSubscriptionsRequest`
- **响应消息**：`ListSubscriptionsResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `alert_type_filter` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `active_only` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `subscriptions` | `repeated AlertSubscription` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "alert_type_filter": "<string>",
  "active_only": false
}
```

#### 响应示例

```json
{
  "subscriptions": [
    {
      "subscription_id": "<string>",
      "alert_type": "<string>",
      "pair": "<string>",
      "condition": "<string>",
      "threshold": "<string>",
      "notification_method": "<string>",
      "active": false,
      "created_at": 0
    }
  ]
}
```

### AlertService/Subscribe

- **URL**：`POST https://api.alfq.org/antclaw.v1.AlertService/Subscribe`
- **请求消息**：`SubscribeRequest`
- **响应消息**：`SubscribeResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `alert_type` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `condition` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `threshold` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `notification_method` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `subscription` | `AlertSubscription` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "alert_type": "<string>",
  "pair": "<string>",
  "condition": "<string>",
  "threshold": "<string>",
  "notification_method": "<string>"
}
```

#### 响应示例

```json
{
  "subscription": {
    "subscription_id": "<string>",
    "alert_type": "<string>",
    "pair": "<string>",
    "condition": "<string>",
    "threshold": "<string>",
    "notification_method": "<string>",
    "active": false,
    "created_at": 0
  }
}
```

### AlertService/Unsubscribe

- **URL**：`POST https://api.alfq.org/antclaw.v1.AlertService/Unsubscribe`
- **请求消息**：`UnsubscribeRequest`
- **响应消息**：`UnsubscribeResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `subscription_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{
  "subscription_id": "<string>"
}
```

#### 响应示例

```json
{}
```

### AlertService/RegisterWebhook

- **URL**：`POST https://api.alfq.org/antclaw.v1.AlertService/RegisterWebhook`
- **请求消息**：`RegisterWebhookRequest`
- **响应消息**：`RegisterWebhookResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `url` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `secret` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `event_types` | `repeated string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `webhook` | `WebhookConfig` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "url": "<string>",
  "secret": "<string>",
  "event_types": [
    "<string>"
  ]
}
```

#### 响应示例

```json
{
  "webhook": {
    "webhook_id": "<string>",
    "url": "<string>",
    "secret": "<string>",
    "event_types": [
      "<string>"
    ],
    "active": false,
    "created_at": 0
  }
}
```

### AlertService/ListWebhooks

- **URL**：`POST https://api.alfq.org/antclaw.v1.AlertService/ListWebhooks`
- **请求消息**：`ListWebhooksRequest`
- **响应消息**：`ListWebhooksResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `webhooks` | `repeated WebhookConfig` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{}
```

#### 响应示例

```json
{
  "webhooks": [
    {
      "webhook_id": "<string>",
      "url": "<string>",
      "secret": "<string>",
      "event_types": [
        "<string>"
      ],
      "active": false,
      "created_at": 0
    }
  ]
}
```

### AlertService/CreateAlert

- **URL**：`POST https://api.alfq.org/antclaw.v1.AlertService/CreateAlert`
- **请求消息**：`CreateAlertRequest`
- **响应消息**：`CreateAlertResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `alert_type` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `symbol` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `params_json` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `cooldown_seconds` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{
  "alert_type": "<string>",
  "symbol": "<string>",
  "params_json": "<string>",
  "cooldown_seconds": 0
}
```

#### 响应示例

```json
{}
```

### AlertService/ListAlerts

- **URL**：`POST https://api.alfq.org/antclaw.v1.AlertService/ListAlerts`
- **请求消息**：`ListAlertsRequest`
- **响应消息**：`ListAlertsResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{}
```

#### 响应示例

```json
{}
```

### AlertService/UpdateAlert

- **URL**：`POST https://api.alfq.org/antclaw.v1.AlertService/UpdateAlert`
- **请求消息**：`UpdateAlertRequest`
- **响应消息**：`UpdateAlertResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `id` | `int64` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `params_json` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `cooldown_seconds` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{
  "id": 0,
  "params_json": "<string>",
  "cooldown_seconds": 0
}
```

#### 响应示例

```json
{}
```

### AlertService/DeleteAlert

- **URL**：`POST https://api.alfq.org/antclaw.v1.AlertService/DeleteAlert`
- **请求消息**：`DeleteAlertRequest`
- **响应消息**：`DeleteAlertResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{}
```

#### 响应示例

```json
{}
```

### AlertService/ToggleAlert

- **URL**：`POST https://api.alfq.org/antclaw.v1.AlertService/ToggleAlert`
- **请求消息**：`ToggleAlertRequest`
- **响应消息**：`ToggleAlertResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `id` | `int64` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `enabled` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{
  "id": 0,
  "enabled": false
}
```

#### 响应示例

```json
{}
```

### AlertService/DecideAlert

- **URL**：`POST https://api.alfq.org/antclaw.v1.AlertService/DecideAlert`
- **请求消息**：`DecideAlertRequest`
- **响应消息**：`DecideAlertResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `user_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `alert_type` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `severity` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | low / medium / high / critical |
| `pairs` | `repeated string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `send` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `reason` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | tier_blocked / quiet_hours / cooldown / unsubscribed_pair / ok |

#### 请求示例

```json
{
  "user_id": "<string>",
  "alert_type": "<string>",
  "severity": "<string>",
  "pairs": [
    "<string>"
  ]
}
```

#### 响应示例

```json
{
  "send": false,
  "reason": "<string>"
}
```

### AlertService/GetPreferences

- **URL**：`POST https://api.alfq.org/antclaw.v1.AlertService/GetPreferences`
- **请求消息**：`GetPreferencesRequest`
- **响应消息**：`GetPreferencesResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `user_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `pairs` | `repeated string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `high_impact_only` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `quiet_hours_start` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `quiet_hours_end` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `timezone` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{}
```

#### 响应示例

```json
{
  "user_id": "<string>",
  "pairs": [
    "<string>"
  ],
  "high_impact_only": false,
  "quiet_hours_start": 0,
  "quiet_hours_end": 0,
  "timezone": "<string>"
}
```

### AlertService/UpdatePreferences

- **URL**：`POST https://api.alfq.org/antclaw.v1.AlertService/UpdatePreferences`
- **请求消息**：`UpdatePreferencesRequest`
- **响应消息**：`UpdatePreferencesResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `user_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `pairs` | `repeated string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `high_impact_only` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `quiet_hours_start` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `quiet_hours_end` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `timezone` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{
  "user_id": "<string>",
  "pairs": [
    "<string>"
  ],
  "high_impact_only": false,
  "quiet_hours_start": 0,
  "quiet_hours_end": 0,
  "timezone": "<string>"
}
```

#### 响应示例

```json
{}
```

### AlertService/SetUserTier

- **URL**：`POST https://api.alfq.org/antclaw.v1.AlertService/SetUserTier`
- **请求消息**：`SetUserTierRequest`
- **响应消息**：`SetUserTierResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{}
```

#### 响应示例

```json
{}
```

### AlertService/GetAlertHistory

- **URL**：`POST https://api.alfq.org/antclaw.v1.AlertService/GetAlertHistory`
- **请求消息**：`GetAlertHistoryRequest`
- **响应消息**：`GetAlertHistoryResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{}
```

#### 响应示例

```json
{}
```

## ChatService

### ChatService/SendMessage

- **URL**：`POST https://api.alfq.org/antclaw.v1.ChatService/SendMessage`
- **请求消息**：`SendMessageRequest`
- **响应消息**：`Message`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `conversation_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `content` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `message_type` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `signal_data` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `conversation_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `sender_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `sender_name` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `content` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `message_type` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | text / signal_share / chart_share |
| `signal_data` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | JSON, for signal_share |
| `created_at` | `int64` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "conversation_id": "<string>",
  "content": "<string>",
  "message_type": "<string>",
  "signal_data": "<string>"
}
```

#### 响应示例

```json
{
  "id": "<string>",
  "conversation_id": "<string>",
  "sender_id": "<string>",
  "sender_name": "<string>",
  "content": "<string>",
  "message_type": "<string>",
  "signal_data": "<string>",
  "created_at": 0
}
```

### ChatService/GetConversation

- **URL**：`POST https://api.alfq.org/antclaw.v1.ChatService/GetConversation`
- **请求消息**：`GetConversationRequest`
- **响应消息**：`Conversation`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `name` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 对方用户名 / 群名 |
| `is_group` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `last_message` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `last_message_at` | `int64` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `unread_count` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `preview` | `Message` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{}
```

#### 响应示例

```json
{
  "id": "<string>",
  "name": "<string>",
  "is_group": false,
  "last_message": "<string>",
  "last_message_at": 0,
  "unread_count": 0,
  "preview": {
    "id": "<string>",
    "conversation_id": "<string>",
    "sender_id": "<string>",
    "sender_name": "<string>",
    "content": "<string>",
    "message_type": "<string>",
    "signal_data": "<string>",
    "created_at": 0
  }
}
```

### ChatService/ListConversations

- **URL**：`POST https://api.alfq.org/antclaw.v1.ChatService/ListConversations`
- **请求消息**：`ListConversationsRequest`
- **响应消息**：`ConversationList`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{}
```

#### 响应示例

```json
{}
```

### ChatService/MarkRead

- **URL**：`POST https://api.alfq.org/antclaw.v1.ChatService/MarkRead`
- **请求消息**：`ChatMarkReadRequest`
- **响应消息**：`ChatMarkReadResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{}
```

#### 响应示例

```json
{}
```

## CircleService

### CircleService/CreateCircle

- **URL**：`POST https://api.alfq.org/antclaw.v1.CircleService/CreateCircle`
- **请求消息**：`CreateCircleRequest`
- **响应消息**：`Circle`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `name` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `description` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `symbol` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `name` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `description` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `symbol` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | associated symbol |
| `member_count` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `is_member` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `created_at` | `int64` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "name": "<string>",
  "description": "<string>",
  "symbol": "<string>"
}
```

#### 响应示例

```json
{
  "id": "<string>",
  "name": "<string>",
  "description": "<string>",
  "symbol": "<string>",
  "member_count": 0,
  "is_member": false,
  "created_at": 0
}
```

### CircleService/JoinCircle

- **URL**：`POST https://api.alfq.org/antclaw.v1.CircleService/JoinCircle`
- **请求消息**：`JoinCircleRequest`
- **响应消息**：`Circle`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `name` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `description` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `symbol` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | associated symbol |
| `member_count` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `is_member` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `created_at` | `int64` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{}
```

#### 响应示例

```json
{
  "id": "<string>",
  "name": "<string>",
  "description": "<string>",
  "symbol": "<string>",
  "member_count": 0,
  "is_member": false,
  "created_at": 0
}
```

### CircleService/LeaveCircle

- **URL**：`POST https://api.alfq.org/antclaw.v1.CircleService/LeaveCircle`
- **请求消息**：`LeaveCircleRequest`
- **响应消息**：`LeaveCircleResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{}
```

#### 响应示例

```json
{}
```

### CircleService/GetCircleFeed

- **URL**：`POST https://api.alfq.org/antclaw.v1.CircleService/GetCircleFeed`
- **请求消息**：`GetCircleFeedRequest`
- **响应消息**：`CircleFeedResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `posts` | `repeated CirclePost` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `next_cursor` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{}
```

#### 响应示例

```json
{
  "posts": [
    {
      "id": "<string>",
      "circle_id": "<string>",
      "author_name": "<string>",
      "content": "<string>",
      "created_at": 0
    }
  ],
  "next_cursor": "<string>"
}
```

### CircleService/ListCircles

- **URL**：`POST https://api.alfq.org/antclaw.v1.CircleService/ListCircles`
- **请求消息**：`ListCirclesRequest`
- **响应消息**：`CircleList`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{}
```

#### 响应示例

```json
{}
```

## FeedService

### FeedService/CreatePost

- **URL**：`POST https://api.alfq.org/antclaw.v1.FeedService/CreatePost`
- **请求消息**：`CreatePostRequest`
- **响应消息**：`Post`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `content` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `post_type` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `signal_pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `signal_direction` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `signal_confidence` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `visibility` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `circle_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `author_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `author_name` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `content` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `post_type` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | text / signal_card / chart_share |
| `signal_pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | optional, for signal_card |
| `signal_direction` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `signal_confidence` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `visibility` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | public / circle / followers |
| `circle_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | if visibility=circle |
| `like_count` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `comment_count` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `share_count` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `liked_by` | `repeated string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `created_at` | `int64` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "content": "<string>",
  "post_type": "<string>",
  "signal_pair": "<string>",
  "signal_direction": "<string>",
  "signal_confidence": 0,
  "visibility": "<string>",
  "circle_id": "<string>"
}
```

#### 响应示例

```json
{
  "id": "<string>",
  "author_id": "<string>",
  "author_name": "<string>",
  "content": "<string>",
  "post_type": "<string>",
  "signal_pair": "<string>",
  "signal_direction": "<string>",
  "signal_confidence": 0,
  "visibility": "<string>",
  "circle_id": "<string>",
  "like_count": 0,
  "comment_count": 0
}
```

### FeedService/GetFeed

- **URL**：`POST https://api.alfq.org/antclaw.v1.FeedService/GetFeed`
- **请求消息**：`GetFeedRequest`
- **响应消息**：`FeedResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `cursor` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `page_size` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | default 20 |
| `filter` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | all / signals_only / posts_only |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `posts` | `repeated Post` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `next_cursor` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "cursor": "<string>",
  "page_size": 0,
  "filter": "<string>"
}
```

#### 响应示例

```json
{
  "posts": [
    {
      "id": "<string>",
      "author_id": "<string>",
      "author_name": "<string>",
      "content": "<string>",
      "post_type": "<string>",
      "signal_pair": "<string>",
      "signal_direction": "<string>",
      "signal_confidence": 0,
      "visibility": "<string>",
      "circle_id": "<string>",
      "like_count": 0,
      "comment_count": 0
    }
  ],
  "next_cursor": "<string>"
}
```

### FeedService/GetPost

- **URL**：`POST https://api.alfq.org/antclaw.v1.FeedService/GetPost`
- **请求消息**：`GetPostRequest`
- **响应消息**：`Post`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `author_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `author_name` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `content` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `post_type` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | text / signal_card / chart_share |
| `signal_pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | optional, for signal_card |
| `signal_direction` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `signal_confidence` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `visibility` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | public / circle / followers |
| `circle_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | if visibility=circle |
| `like_count` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `comment_count` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `share_count` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `liked_by` | `repeated string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `created_at` | `int64` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{}
```

#### 响应示例

```json
{
  "id": "<string>",
  "author_id": "<string>",
  "author_name": "<string>",
  "content": "<string>",
  "post_type": "<string>",
  "signal_pair": "<string>",
  "signal_direction": "<string>",
  "signal_confidence": 0,
  "visibility": "<string>",
  "circle_id": "<string>",
  "like_count": 0,
  "comment_count": 0
}
```

### FeedService/LikePost

- **URL**：`POST https://api.alfq.org/antclaw.v1.FeedService/LikePost`
- **请求消息**：`LikePostRequest`
- **响应消息**：`Post`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `author_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `author_name` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `content` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `post_type` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | text / signal_card / chart_share |
| `signal_pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | optional, for signal_card |
| `signal_direction` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `signal_confidence` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `visibility` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | public / circle / followers |
| `circle_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | if visibility=circle |
| `like_count` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `comment_count` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `share_count` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `liked_by` | `repeated string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `created_at` | `int64` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{}
```

#### 响应示例

```json
{
  "id": "<string>",
  "author_id": "<string>",
  "author_name": "<string>",
  "content": "<string>",
  "post_type": "<string>",
  "signal_pair": "<string>",
  "signal_direction": "<string>",
  "signal_confidence": 0,
  "visibility": "<string>",
  "circle_id": "<string>",
  "like_count": 0,
  "comment_count": 0
}
```

### FeedService/UnlikePost

- **URL**：`POST https://api.alfq.org/antclaw.v1.FeedService/UnlikePost`
- **请求消息**：`UnlikePostRequest`
- **响应消息**：`Post`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `author_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `author_name` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `content` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `post_type` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | text / signal_card / chart_share |
| `signal_pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | optional, for signal_card |
| `signal_direction` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `signal_confidence` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `visibility` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | public / circle / followers |
| `circle_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | if visibility=circle |
| `like_count` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `comment_count` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `share_count` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `liked_by` | `repeated string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `created_at` | `int64` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{}
```

#### 响应示例

```json
{
  "id": "<string>",
  "author_id": "<string>",
  "author_name": "<string>",
  "content": "<string>",
  "post_type": "<string>",
  "signal_pair": "<string>",
  "signal_direction": "<string>",
  "signal_confidence": 0,
  "visibility": "<string>",
  "circle_id": "<string>",
  "like_count": 0,
  "comment_count": 0
}
```

### FeedService/CommentOnPost

- **URL**：`POST https://api.alfq.org/antclaw.v1.FeedService/CommentOnPost`
- **请求消息**：`CommentRequest`
- **响应消息**：`Comment`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `post_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `content` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `post_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `author_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `author_name` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `content` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `created_at` | `int64` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "post_id": "<string>",
  "content": "<string>"
}
```

#### 响应示例

```json
{
  "id": "<string>",
  "post_id": "<string>",
  "author_id": "<string>",
  "author_name": "<string>",
  "content": "<string>",
  "created_at": 0
}
```

### FeedService/SharePost

- **URL**：`POST https://api.alfq.org/antclaw.v1.FeedService/SharePost`
- **请求消息**：`SharePostRequest`
- **响应消息**：`Post`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `post_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `comment` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | optional quote |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `author_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `author_name` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `content` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `post_type` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | text / signal_card / chart_share |
| `signal_pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | optional, for signal_card |
| `signal_direction` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `signal_confidence` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `visibility` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | public / circle / followers |
| `circle_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | if visibility=circle |
| `like_count` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `comment_count` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `share_count` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `liked_by` | `repeated string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `created_at` | `int64` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "post_id": "<string>",
  "comment": "<string>"
}
```

#### 响应示例

```json
{
  "id": "<string>",
  "author_id": "<string>",
  "author_name": "<string>",
  "content": "<string>",
  "post_type": "<string>",
  "signal_pair": "<string>",
  "signal_direction": "<string>",
  "signal_confidence": 0,
  "visibility": "<string>",
  "circle_id": "<string>",
  "like_count": 0,
  "comment_count": 0
}
```

## MarketplaceService

### MarketplaceService/ListProducts

- **URL**：`POST https://api.alfq.org/antclaw.v1.MarketplaceService/ListProducts`
- **请求消息**：`ListProductsRequest`
- **响应消息**：`ProductList`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{}
```

#### 响应示例

```json
{}
```

### MarketplaceService/PublishProduct

- **URL**：`POST https://api.alfq.org/antclaw.v1.MarketplaceService/PublishProduct`
- **请求消息**：`PublishProductRequest`
- **响应消息**：`Product`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `name` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `category` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `description` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `symbol` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `purchase_type` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `price` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `trial_days` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `author_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `author_name` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `name` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `category` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | strategy / indicator / signal / ea |
| `description` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `symbol` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `purchase_type` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | subscription / one_time |
| `price` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `trial_days` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `rating` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `purchase_count` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `created_at` | `int64` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "name": "<string>",
  "category": "<string>",
  "description": "<string>",
  "symbol": "<string>",
  "purchase_type": "<string>",
  "price": 0.0,
  "trial_days": 0
}
```

#### 响应示例

```json
{
  "id": "<string>",
  "author_id": "<string>",
  "author_name": "<string>",
  "name": "<string>",
  "category": "<string>",
  "description": "<string>",
  "symbol": "<string>",
  "purchase_type": "<string>",
  "price": 0.0,
  "trial_days": 0,
  "rating": 0.0,
  "purchase_count": 0
}
```

### MarketplaceService/PurchaseProduct

- **URL**：`POST https://api.alfq.org/antclaw.v1.MarketplaceService/PurchaseProduct`
- **请求消息**：`PurchaseProductRequest`
- **响应消息**：`Purchase`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `product_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `start_trial` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `product_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `product_name` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `is_trial` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `expires_at` | `int64` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `created_at` | `int64` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "product_id": "<string>",
  "start_trial": false
}
```

#### 响应示例

```json
{
  "id": "<string>",
  "product_id": "<string>",
  "product_name": "<string>",
  "is_trial": false,
  "expires_at": 0,
  "created_at": 0
}
```

### MarketplaceService/GetMyProducts

- **URL**：`POST https://api.alfq.org/antclaw.v1.MarketplaceService/GetMyProducts`
- **请求消息**：`GetMyProductsRequest`
- **响应消息**：`ProductList`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{}
```

#### 响应示例

```json
{}
```

### MarketplaceService/GetMyPurchases

- **URL**：`POST https://api.alfq.org/antclaw.v1.MarketplaceService/GetMyPurchases`
- **请求消息**：`GetMyPurchasesRequest`
- **响应消息**：`PurchaseList`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{}
```

#### 响应示例

```json
{}
```

## TraderService

### TraderService/GetProfile

- **URL**：`POST https://api.alfq.org/antclaw.v1.TraderService/GetProfile`
- **请求消息**：`GetTraderProfileRequest`
- **响应消息**：`TraderProfile`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `user_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `display_name` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `bio` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `tier` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | normal / verified / elite |
| `show_win_rate` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `show_profit_factor` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `show_sharpe` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `show_total_trades` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `win_rate` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `profit_factor` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `sharpe_ratio` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `total_trades` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `follower_count` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `following_count` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `created_at` | `int64` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{}
```

#### 响应示例

```json
{
  "user_id": "<string>",
  "display_name": "<string>",
  "bio": "<string>",
  "tier": "<string>",
  "show_win_rate": false,
  "show_profit_factor": false,
  "show_sharpe": false,
  "show_total_trades": false,
  "win_rate": 0.0,
  "profit_factor": 0.0,
  "sharpe_ratio": 0.0,
  "total_trades": 0
}
```

### TraderService/UpdateProfile

- **URL**：`POST https://api.alfq.org/antclaw.v1.TraderService/UpdateProfile`
- **请求消息**：`UpdateTraderProfileRequest`
- **响应消息**：`TraderProfile`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `display_name` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `bio` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `show_win_rate` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `show_profit_factor` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `show_sharpe` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `show_total_trades` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `user_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `display_name` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `bio` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `tier` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | normal / verified / elite |
| `show_win_rate` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `show_profit_factor` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `show_sharpe` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `show_total_trades` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `win_rate` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `profit_factor` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `sharpe_ratio` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `total_trades` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `follower_count` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `following_count` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `created_at` | `int64` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "display_name": "<string>",
  "bio": "<string>",
  "show_win_rate": false,
  "show_profit_factor": false,
  "show_sharpe": false,
  "show_total_trades": false
}
```

#### 响应示例

```json
{
  "user_id": "<string>",
  "display_name": "<string>",
  "bio": "<string>",
  "tier": "<string>",
  "show_win_rate": false,
  "show_profit_factor": false,
  "show_sharpe": false,
  "show_total_trades": false,
  "win_rate": 0.0,
  "profit_factor": 0.0,
  "sharpe_ratio": 0.0,
  "total_trades": 0
}
```

### TraderService/Follow

- **URL**：`POST https://api.alfq.org/antclaw.v1.TraderService/Follow`
- **请求消息**：`FollowRequest`
- **响应消息**：`FollowResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{}
```

#### 响应示例

```json
{}
```

### TraderService/Unfollow

- **URL**：`POST https://api.alfq.org/antclaw.v1.TraderService/Unfollow`
- **请求消息**：`UnfollowRequest`
- **响应消息**：`FollowResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{}
```

#### 响应示例

```json
{}
```

### TraderService/GetFollowers

- **URL**：`POST https://api.alfq.org/antclaw.v1.TraderService/GetFollowers`
- **请求消息**：`GetFollowersRequest`
- **响应消息**：`UserList`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `users` | `repeated UserInfo` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `next_cursor` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{}
```

#### 响应示例

```json
{
  "users": [
    {
      "user_id": "<string>",
      "display_name": "<string>",
      "tier": "<string>",
      "follower_count": 0
    }
  ],
  "next_cursor": "<string>"
}
```

### TraderService/GetFollowing

- **URL**：`POST https://api.alfq.org/antclaw.v1.TraderService/GetFollowing`
- **请求消息**：`GetFollowingRequest`
- **响应消息**：`UserList`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `users` | `repeated UserInfo` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `next_cursor` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{}
```

#### 响应示例

```json
{
  "users": [
    {
      "user_id": "<string>",
      "display_name": "<string>",
      "tier": "<string>",
      "follower_count": 0
    }
  ],
  "next_cursor": "<string>"
}
```

## AuthService

### AuthService/Register

- **URL**：`POST https://api.alfq.org/antclaw.v1.AuthService/Register`
- **请求消息**：`RegisterRequest`
- **响应消息**：`RegisterResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `email` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `password` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `display_name` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `locale` | `Locale` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `timezone` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `client` | `ClientInfo` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `idempotency_key` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `user_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `access_token` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `refresh_token` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `expires_at` | `int64` | 否；proto3 默认值语义，业务必填以服务端校验为准 | Unix timestamp |
| `code_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "email": "<string>",
  "password": "<string>",
  "display_name": "<string>",
  "locale": "LOCALE_UNSPECIFIED",
  "timezone": "<string>",
  "client": {
    "user_agent": "<string>",
    "ip_address": "<string>",
    "device_id": "<string>"
  },
  "idempotency_key": "<string>"
}
```

#### 响应示例

```json
{
  "user_id": "<string>",
  "access_token": "<string>",
  "refresh_token": "<string>",
  "expires_at": 0,
  "code_id": "<string>"
}
```

### AuthService/Login

- **URL**：`POST https://api.alfq.org/antclaw.v1.AuthService/Login`
- **请求消息**：`LoginRequest`
- **响应消息**：`LoginResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `email` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `password` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `client` | `ClientInfo` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `user_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `access_token` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `refresh_token` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `expires_at` | `int64` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `code_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "email": "<string>",
  "password": "<string>",
  "client": {
    "user_agent": "<string>",
    "ip_address": "<string>",
    "device_id": "<string>"
  }
}
```

#### 响应示例

```json
{
  "user_id": "<string>",
  "access_token": "<string>",
  "refresh_token": "<string>",
  "expires_at": 0,
  "code_id": "<string>"
}
```

### AuthService/Refresh

- **URL**：`POST https://api.alfq.org/antclaw.v1.AuthService/Refresh`
- **请求消息**：`RefreshRequest`
- **响应消息**：`RefreshResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `refresh_token` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `access_token` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `refresh_token` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `expires_at` | `int64` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "refresh_token": "<string>"
}
```

#### 响应示例

```json
{
  "access_token": "<string>",
  "refresh_token": "<string>",
  "expires_at": 0
}
```

### AuthService/Logout

- **URL**：`POST https://api.alfq.org/antclaw.v1.AuthService/Logout`
- **请求消息**：`LogoutRequest`
- **响应消息**：`LogoutResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `refresh_token` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `all_devices` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | If true, logout all sessions |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{
  "refresh_token": "<string>",
  "all_devices": false
}
```

#### 响应示例

```json
{}
```

### AuthService/RequestPasswordReset

- **URL**：`POST https://api.alfq.org/antclaw.v1.AuthService/RequestPasswordReset`
- **请求消息**：`RequestPasswordResetRequest`
- **响应消息**：`RequestPasswordResetResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `email` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `sent` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | Always true to prevent email enumeration |

#### 请求示例

```json
{
  "email": "<string>"
}
```

#### 响应示例

```json
{
  "sent": false
}
```

### AuthService/ResetPassword

- **URL**：`POST https://api.alfq.org/antclaw.v1.AuthService/ResetPassword`
- **请求消息**：`ResetPasswordRequest`
- **响应消息**：`ResetPasswordResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `token` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `new_password` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `user_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "token": "<string>",
  "new_password": "<string>"
}
```

#### 响应示例

```json
{
  "user_id": "<string>"
}
```

### AuthService/VerifyEmail

- **URL**：`POST https://api.alfq.org/antclaw.v1.AuthService/VerifyEmail`
- **请求消息**：`VerifyEmailRequest`
- **响应消息**：`VerifyEmailResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `token` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `user_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `verified` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "token": "<string>"
}
```

#### 响应示例

```json
{
  "user_id": "<string>",
  "verified": false
}
```

## BacktestService

### BacktestService/RunBacktest

- **URL**：`POST https://api.alfq.org/antclaw.v1.BacktestService/RunBacktest`
- **请求消息**：`RunBacktestRequest`
- **响应消息**：`RunBacktestResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `config` | `BacktestConfig` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `idempotency_key` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `task_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `status` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | pending, running, completed, failed |

#### 请求示例

```json
{
  "config": {
    "strategy_id": "<string>",
    "pair": "<string>",
    "period": {
      "start": "<string>",
      "end": "<string>"
    },
    "timeframe": "<string>",
    "initial_balance": {
      "amount": "<string>",
      "currency": "<string>"
    },
    "max_position_size": 0.0,
    "stop_loss_pct": 0.0,
    "take_profit_pct": 0.0
  },
  "idempotency_key": "<string>"
}
```

#### 响应示例

```json
{
  "task_id": "<string>",
  "status": "<string>"
}
```

### BacktestService/GetBacktest

- **URL**：`POST https://api.alfq.org/antclaw.v1.BacktestService/GetBacktest`
- **请求消息**：`GetBacktestRequest`
- **响应消息**：`GetBacktestResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `task_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `task_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `status` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `config` | `BacktestConfig` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `metrics` | `BacktestMetrics` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `trades` | `repeated TradeRecord` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "task_id": "<string>"
}
```

#### 响应示例

```json
{
  "task_id": "<string>",
  "status": "<string>",
  "config": {
    "strategy_id": "<string>",
    "pair": "<string>",
    "period": {
      "start": "<string>",
      "end": "<string>"
    },
    "timeframe": "<string>",
    "initial_balance": {
      "amount": "<string>",
      "currency": "<string>"
    },
    "max_position_size": 0.0,
    "stop_loss_pct": 0.0,
    "take_profit_pct": 0.0
  },
  "metrics": {
    "total_return": "<string>",
    "total_return_pct": "<string>",
    "sharpe_ratio": 0.0,
    "max_drawdown": 0.0,
    "win_rate": 0.0,
    "total_trades": 0,
    "profit_factor": 0.0
  },
  "trades": [
    {
      "trade_id": "<string>",
      "entry_time": "<string>",
      "exit_time": "<string>",
      "direction": "<string>",
      "entry_price": "<string>",
      "exit_price": "<string>",
      "pnl": "<string>",
      "pnl_pct": "<string>"
    }
  ]
}
```

### BacktestService/GetAccuracy

- **URL**：`POST https://api.alfq.org/antclaw.v1.BacktestService/GetAccuracy`
- **请求消息**：`GetAccuracyRequest`
- **响应消息**：`GetAccuracyResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `strategy_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `period` | `TimeRange` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `strategy_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `metrics` | `AccuracyMetrics` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "strategy_id": "<string>",
  "period": {
    "start": "<string>",
    "end": "<string>"
  }
}
```

#### 响应示例

```json
{
  "strategy_id": "<string>",
  "metrics": {
    "directional_accuracy": 0.0,
    "avg_return": 0.0,
    "hit_rate": 0.0,
    "sharpe": 0.0,
    "sortino": 0.0,
    "sample_size": 0,
    "std_dev": 0.0
  }
}
```

### BacktestService/RunQuantBt

- **URL**：`POST https://api.alfq.org/antclaw.v1.BacktestService/RunQuantBt`
- **请求消息**：`RunQuantBtRequest`
- **响应消息**：`RunQuantBtResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `config` | `QuantBtConfig` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `task_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `status` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "config": {
    "pair": "<string>",
    "strategy_name": "<string>",
    "period": {
      "start": "<string>",
      "end": "<string>"
    }
  }
}
```

#### 响应示例

```json
{
  "task_id": "<string>",
  "status": "<string>"
}
```

### BacktestService/RunVpBt

- **URL**：`POST https://api.alfq.org/antclaw.v1.BacktestService/RunVpBt`
- **请求消息**：`RunVpBtRequest`
- **响应消息**：`RunVpBtResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `config` | `VpBtConfig` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `task_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `status` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "config": {
    "pair": "<string>",
    "period": {
      "start": "<string>",
      "end": "<string>"
    },
    "num_bins": 0
  }
}
```

#### 响应示例

```json
{
  "task_id": "<string>",
  "status": "<string>"
}
```

### BacktestService/RunCtaBt

- **URL**：`POST https://api.alfq.org/antclaw.v1.BacktestService/RunCtaBt`
- **请求消息**：`RunCtaBtRequest`
- **响应消息**：`RunCtaBtResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `config` | `CtaBtConfig` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `task_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `status` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "config": {
    "pair": "<string>",
    "period": {
      "start": "<string>",
      "end": "<string>"
    },
    "lookback": 0,
    "strategy": "<string>",
    "symbols": [
      "<string>"
    ],
    "timeframe": "<string>",
    "secondary_timeframe": "<string>",
    "target_vol": 0.0
  }
}
```

#### 响应示例

```json
{
  "task_id": "<string>",
  "status": "<string>"
}
```

### BacktestService/RunWalkforward

- **URL**：`POST https://api.alfq.org/antclaw.v1.BacktestService/RunWalkforward`
- **请求消息**：`RunWalkforwardRequest`
- **响应消息**：`RunWalkforwardResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `strategy` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `symbols` | `repeated string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `from_date` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `to_date` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `folds` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `train_ratio` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `job_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `status` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "strategy": "<string>",
  "symbols": [
    "<string>"
  ],
  "from_date": "<string>",
  "to_date": "<string>",
  "folds": 0,
  "train_ratio": 0.0
}
```

#### 响应示例

```json
{
  "job_id": "<string>",
  "status": "<string>"
}
```

### BacktestService/GetWalkforwardResult

- **URL**：`POST https://api.alfq.org/antclaw.v1.BacktestService/GetWalkforwardResult`
- **请求消息**：`GetWalkforwardResultRequest`
- **响应消息**：`GetWalkforwardResultResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `job_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `job_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `folds` | `repeated WalkforwardFold` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "job_id": "<string>"
}
```

#### 响应示例

```json
{
  "job_id": "<string>",
  "folds": [
    {
      "fold_idx": 0,
      "train_from": "<string>",
      "train_to": "<string>",
      "test_from": "<string>",
      "test_to": "<string>",
      "in_sample_sharpe": 0.0,
      "oos_sharpe": 0.0
    }
  ]
}
```

### BacktestService/RunBootstrap

- **URL**：`POST https://api.alfq.org/antclaw.v1.BacktestService/RunBootstrap`
- **请求消息**：`RunBootstrapRequest`
- **响应消息**：`RunBootstrapResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `base_job_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `iterations` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `random_seed` | `uint64` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `sharpe_p5` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `sharpe_p50` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `sharpe_p95` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `maxdd_p5` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `maxdd_p50` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `maxdd_p95` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `iterations` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "base_job_id": "<string>",
  "iterations": 0,
  "random_seed": 0
}
```

#### 响应示例

```json
{
  "sharpe_p5": 0.0,
  "sharpe_p50": 0.0,
  "sharpe_p95": 0.0,
  "maxdd_p5": 0.0,
  "maxdd_p50": 0.0,
  "maxdd_p95": 0.0,
  "iterations": 0
}
```

### BacktestService/RunMonteCarlo

- **URL**：`POST https://api.alfq.org/antclaw.v1.BacktestService/RunMonteCarlo`
- **请求消息**：`RunMonteCarloRequest`
- **响应消息**：`RunMonteCarloResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `timeframe` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `paths` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 默认 1000；上限 10000 |
| `horizon_bars` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 默认 20 |
| `random_seed` | `uint64` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `lookback` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 用于 GARCH 拟合的历史长度，默认 500 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `paths` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `horizon_bars` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `terminal_p05` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `terminal_p50` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `terminal_p95` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `quantile_paths` | `repeated MCPath` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `garch_omega` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `garch_alpha` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `garch_beta` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "pair": "<string>",
  "timeframe": "<string>",
  "paths": 0,
  "horizon_bars": 0,
  "random_seed": 0,
  "lookback": 0
}
```

#### 响应示例

```json
{
  "pair": "<string>",
  "paths": 0,
  "horizon_bars": 0,
  "terminal_p05": 0.0,
  "terminal_p50": 0.0,
  "terminal_p95": 0.0,
  "quantile_paths": [
    {
      "label": "<string>",
      "values": [
        0.0
      ]
    }
  ],
  "garch_omega": 0.0,
  "garch_alpha": 0.0,
  "garch_beta": 0.0
}
```

### BacktestService/GetTrades

- **URL**：`POST https://api.alfq.org/antclaw.v1.BacktestService/GetTrades`
- **请求消息**：`GetTradesRequest`
- **响应消息**：`GetTradesResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `job_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `job_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `trades` | `repeated TradeDetail` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "job_id": "<string>"
}
```

#### 响应示例

```json
{
  "job_id": "<string>",
  "trades": [
    {
      "seq": 0,
      "opened_at": "<string>",
      "closed_at": "<string>",
      "side": "<string>",
      "entry": 0.0,
      "exit": 0.0,
      "pnl": 0.0,
      "pnl_pct": 0.0,
      "mfe": 0.0,
      "mae": 0.0,
      "cost": 0.0,
      "regime": "<string>"
    }
  ]
}
```

### BacktestService/GetMetricsByRegime

- **URL**：`POST https://api.alfq.org/antclaw.v1.BacktestService/GetMetricsByRegime`
- **请求消息**：`GetMetricsByRegimeRequest`
- **响应消息**：`GetMetricsByRegimeResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `job_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `job_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `metrics` | `repeated RegimeMetrics` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "job_id": "<string>"
}
```

#### 响应示例

```json
{
  "job_id": "<string>",
  "metrics": [
    {
      "regime": "<string>",
      "n_trades": 0,
      "sharpe": 0.0,
      "sortino": 0.0,
      "max_drawdown": 0.0,
      "win_rate": 0.0
    }
  ]
}
```

## CalendarService

### CalendarService/ListEvents

- **URL**：`POST https://api.alfq.org/antclaw.v1.CalendarService/ListEvents`
- **请求消息**：`ListEventsRequest`
- **响应消息**：`ListEventsResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `date` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | YYYY-MM-DD |
| `currency_filter` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `min_impact` | `ImpactLevel` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `range` | `TimeRange` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `events` | `repeated CalendarEvent` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "date": "<string>",
  "currency_filter": "<string>",
  "min_impact": "IMPACT_LEVEL_UNSPECIFIED",
  "range": {
    "start": "<string>",
    "end": "<string>"
  }
}
```

#### 响应示例

```json
{
  "events": [
    {
      "event_id": "<string>",
      "title": "<string>",
      "country": "<string>",
      "currency": "<string>",
      "impact": "IMPACT_LEVEL_UNSPECIFIED",
      "scheduled_at": "<string>",
      "previous": "<string>",
      "forecast": "<string>",
      "actual": "<string>"
    }
  ]
}
```

### CalendarService/GetEvent

- **URL**：`POST https://api.alfq.org/antclaw.v1.CalendarService/GetEvent`
- **请求消息**：`GetEventRequest`
- **响应消息**：`GetEventResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `event_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `event` | `CalendarEvent` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `description` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "event_id": "<string>"
}
```

#### 响应示例

```json
{
  "event": {
    "event_id": "<string>",
    "title": "<string>",
    "country": "<string>",
    "currency": "<string>",
    "impact": "IMPACT_LEVEL_UNSPECIFIED",
    "scheduled_at": "<string>",
    "previous": "<string>",
    "forecast": "<string>",
    "actual": "<string>"
  },
  "description": "<string>"
}
```

### CalendarService/GetImpact

- **URL**：`POST https://api.alfq.org/antclaw.v1.CalendarService/GetImpact`
- **请求消息**：`GetImpactRequest`
- **响应消息**：`GetImpactResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `event_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `analysis` | `ImpactAnalysis` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "event_id": "<string>"
}
```

#### 响应示例

```json
{
  "analysis": {
    "event_id": "<string>",
    "affected_pairs": [
      "<string>"
    ],
    "expected_direction": "<string>",
    "confidence": 0.0
  }
}
```

### CalendarService/GetImpactHistory

- **URL**：`POST https://api.alfq.org/antclaw.v1.CalendarService/GetImpactHistory`
- **请求消息**：`GetImpactHistoryRequest`
- **响应消息**：`GetImpactHistoryResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `event_type` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `count` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `impacts` | `repeated HistoricalImpact` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "event_type": "<string>",
  "pair": "<string>",
  "count": 0
}
```

#### 响应示例

```json
{
  "impacts": [
    {
      "event_id": "<string>",
      "date": "<string>",
      "pair": "<string>",
      "price_change": 0.0,
      "volatility": 0.0
    }
  ]
}
```

## COTService

### COTService/GetSummary

- **URL**：`POST https://api.alfq.org/antclaw.v1.COTService/GetSummary`
- **请求消息**：`GetSummaryRequest`
- **响应消息**：`GetSummaryResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `latest` | `COTEntry` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `history` | `repeated COTEntry` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "pair": "<string>"
}
```

#### 响应示例

```json
{
  "latest": {
    "pair": "<string>",
    "date": "<string>",
    "non_comm_long": 0,
    "non_comm_short": 0,
    "comm_long": 0,
    "comm_short": 0,
    "non_rep_long": 0,
    "non_rep_short": 0
  },
  "history": [
    {
      "pair": "<string>",
      "date": "<string>",
      "non_comm_long": 0,
      "non_comm_short": 0,
      "comm_long": 0,
      "comm_short": 0,
      "non_rep_long": 0,
      "non_rep_short": 0
    }
  ]
}
```

### COTService/Compare

- **URL**：`POST https://api.alfq.org/antclaw.v1.COTService/Compare`
- **请求消息**：`CompareRequest`
- **响应消息**：`CompareResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `date_a` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `date_b` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `entry_a` | `COTEntry` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `entry_b` | `COTEntry` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `change_summary` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "pair": "<string>",
  "date_a": "<string>",
  "date_b": "<string>"
}
```

#### 响应示例

```json
{
  "entry_a": {
    "pair": "<string>",
    "date": "<string>",
    "non_comm_long": 0,
    "non_comm_short": 0,
    "comm_long": 0,
    "comm_short": 0,
    "non_rep_long": 0,
    "non_rep_short": 0
  },
  "entry_b": {
    "pair": "<string>",
    "date": "<string>",
    "non_comm_long": 0,
    "non_comm_short": 0,
    "comm_long": 0,
    "comm_short": 0,
    "non_rep_long": 0,
    "non_rep_short": 0
  },
  "change_summary": "<string>"
}
```

### COTService/GetSignals

- **URL**：`POST https://api.alfq.org/antclaw.v1.COTService/GetSignals`
- **请求消息**：`GetSignalsRequest`
- **响应消息**：`GetSignalsResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair_filter` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `signals` | `repeated COTSignal` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "pair_filter": "<string>"
}
```

#### 响应示例

```json
{
  "signals": [
    {
      "pair": "<string>",
      "signal_type": "<string>",
      "direction": "<string>",
      "strength": 0.0,
      "created_at": "<string>"
    }
  ]
}
```

### COTService/GetHistory

- **URL**：`POST https://api.alfq.org/antclaw.v1.COTService/GetHistory`
- **请求消息**：`COTServiceGetHistoryRequest`
- **响应消息**：`COTServiceGetHistoryResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `range` | `TimeRange` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `entries` | `repeated COTEntry` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "pair": "<string>",
  "range": {
    "start": "<string>",
    "end": "<string>"
  }
}
```

#### 响应示例

```json
{
  "entries": [
    {
      "pair": "<string>",
      "date": "<string>",
      "non_comm_long": 0,
      "non_comm_short": 0,
      "comm_long": 0,
      "comm_short": 0,
      "non_rep_long": 0,
      "non_rep_short": 0
    }
  ]
}
```

### COTService/SubscribePairAlert

- **URL**：`POST https://api.alfq.org/antclaw.v1.COTService/SubscribePairAlert`
- **请求消息**：`SubscribePairAlertRequest`
- **响应消息**：`SubscribePairAlertResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `threshold` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `subscription_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "pair": "<string>",
  "threshold": 0.0
}
```

#### 响应示例

```json
{
  "subscription_id": "<string>"
}
```

## CryptoService

### CryptoService/GetCryptoPublicKey

- **URL**：`POST https://api.alfq.org/antclaw.v1.CryptoService/GetCryptoPublicKey`
- **请求消息**：`GetCryptoPublicKeyRequest`
- **响应消息**：`GetCryptoPublicKeyResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pem` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{}
```

#### 响应示例

```json
{
  "pem": "<string>"
}
```

### CryptoService/PostEnvelope

- **URL**：`POST https://api.alfq.org/antclaw.v1.CryptoService/PostEnvelope`
- **请求消息**：`PostEnvelopeRequest`
- **响应消息**：`PostEnvelopeResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `body_b64` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `ts` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `nonce` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `sig` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | hex 编码 |
| `target_path` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `target_body_b64` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `body_b64` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "body_b64": "<string>",
  "ts": "<string>",
  "nonce": "<string>",
  "sig": "<string>",
  "target_path": "<string>",
  "target_body_b64": "<string>"
}
```

#### 响应示例

```json
{
  "body_b64": "<string>"
}
```

## DataSourceService

### DataSourceService/ListDataSources

- **URL**：`POST https://api.alfq.org/antclaw.v1.DataSourceService/ListDataSources`
- **请求消息**：`ListDataSourcesRequest`
- **响应消息**：`ListDataSourcesResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `items` | `repeated DataSourceConfig` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{}
```

#### 响应示例

```json
{
  "items": [
    {
      "source_id": "<string>",
      "name": "<string>",
      "kind": "<string>",
      "endpoint": "<string>",
      "has_secret": false,
      "updated_at": "<string>",
      "updated_by": "<string>"
    }
  ]
}
```

### DataSourceService/UpdateDataSource

- **URL**：`POST https://api.alfq.org/antclaw.v1.DataSourceService/UpdateDataSource`
- **请求消息**：`UpdateDataSourceRequest`
- **响应消息**：`UpdateDataSourceResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `source_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `clear_secret` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `item` | `DataSourceConfig` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "source_id": "<string>",
  "clear_secret": false
}
```

#### 响应示例

```json
{
  "item": {
    "source_id": "<string>",
    "name": "<string>",
    "kind": "<string>",
    "endpoint": "<string>",
    "has_secret": false,
    "updated_at": "<string>",
    "updated_by": "<string>"
  }
}
```

## DeFiService

### DeFiService/GetTopProtocols

- **URL**：`POST https://api.alfq.org/antclaw.v1.DeFiService/GetTopProtocols`
- **请求消息**：`GetTopProtocolsRequest`
- **响应消息**：`GetTopProtocolsResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{}
```

#### 响应示例

```json
{}
```

### DeFiService/GetProtocolTVL

- **URL**：`POST https://api.alfq.org/antclaw.v1.DeFiService/GetProtocolTVL`
- **请求消息**：`GetProtocolTVLRequest`
- **响应消息**：`GetProtocolTVLResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{}
```

#### 响应示例

```json
{}
```

### DeFiService/GetAnalysis

- **URL**：`POST https://api.alfq.org/antclaw.v1.DeFiService/GetAnalysis`
- **请求消息**：`DeFiServiceGetAnalysisRequest`
- **响应消息**：`DeFiServiceGetAnalysisResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `total_tvl` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `tvl_change_7d` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `regime` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `narrative` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{}
```

#### 响应示例

```json
{
  "total_tvl": 0.0,
  "tvl_change_7d": 0.0,
  "regime": "<string>",
  "narrative": "<string>"
}
```

## FedWatchService

### FedWatchService/GetFOMCProbabilities

- **URL**：`POST https://api.alfq.org/antclaw.v1.FedWatchService/GetFOMCProbabilities`
- **请求消息**：`GetFOMCProbabilitiesRequest`
- **响应消息**：`GetFOMCProbabilitiesResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `meeting_date` | `google.protobuf.Timestamp` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `probabilities` | `repeated RateProbability` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{}
```

#### 响应示例

```json
{
  "meeting_date": {},
  "probabilities": [
    {
      "rate_low": 0.0,
      "rate_high": 0.0,
      "probability": 0.0
    }
  ]
}
```

## MacroService

### MacroService/GetFred

- **URL**：`POST https://api.alfq.org/antclaw.v1.MacroService/GetFred`
- **请求消息**：`GetFredRequest`
- **响应消息**：`GetFredResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `series_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `range` | `TimeRange` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `series_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `series_name` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `data` | `repeated MacroDataPoint` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "series_id": "<string>",
  "range": {
    "start": "<string>",
    "end": "<string>"
  }
}
```

#### 响应示例

```json
{
  "series_id": "<string>",
  "series_name": "<string>",
  "data": [
    {
      "date": "<string>",
      "value": "<string>",
      "unit": "<string>"
    }
  ]
}
```

### MacroService/GetEcb

- **URL**：`POST https://api.alfq.org/antclaw.v1.MacroService/GetEcb`
- **请求消息**：`GetEcbRequest`
- **响应消息**：`GetEcbResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `series_key` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `range` | `TimeRange` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `series_key` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `description` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `data` | `repeated MacroDataPoint` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "series_key": "<string>",
  "range": {
    "start": "<string>",
    "end": "<string>"
  }
}
```

#### 响应示例

```json
{
  "series_key": "<string>",
  "description": "<string>",
  "data": [
    {
      "date": "<string>",
      "value": "<string>",
      "unit": "<string>"
    }
  ]
}
```

### MacroService/GetSnb

- **URL**：`POST https://api.alfq.org/antclaw.v1.MacroService/GetSnb`
- **请求消息**：`GetSnbRequest`
- **响应消息**：`GetSnbResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `indicator` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `range` | `TimeRange` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `indicator` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `data` | `repeated MacroDataPoint` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "indicator": "<string>",
  "range": {
    "start": "<string>",
    "end": "<string>"
  }
}
```

#### 响应示例

```json
{
  "indicator": "<string>",
  "data": [
    {
      "date": "<string>",
      "value": "<string>",
      "unit": "<string>"
    }
  ]
}
```

### MacroService/GetOecdLeading

- **URL**：`POST https://api.alfq.org/antclaw.v1.MacroService/GetOecdLeading`
- **请求消息**：`GetOecdLeadingRequest`
- **响应消息**：`GetOecdLeadingResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `country` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `indicators` | `repeated OecdIndicator` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "country": "<string>"
}
```

#### 响应示例

```json
{
  "indicators": [
    {
      "country": "<string>",
      "date": "<string>",
      "cli": 0.0,
      "phase": "<string>"
    }
  ]
}
```

### MacroService/GetEurostat

- **URL**：`POST https://api.alfq.org/antclaw.v1.MacroService/GetEurostat`
- **请求消息**：`GetEurostatRequest`
- **响应消息**：`GetEurostatResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `dataset` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `geo` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `range` | `TimeRange` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `dataset` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `data` | `repeated MacroDataPoint` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "dataset": "<string>",
  "geo": "<string>",
  "range": {
    "start": "<string>",
    "end": "<string>"
  }
}
```

#### 响应示例

```json
{
  "dataset": "<string>",
  "data": [
    {
      "date": "<string>",
      "value": "<string>",
      "unit": "<string>"
    }
  ]
}
```

### MacroService/GetBis

- **URL**：`POST https://api.alfq.org/antclaw.v1.MacroService/GetBis`
- **请求消息**：`GetBisRequest`
- **响应消息**：`GetBisResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `dataset` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | credit, debt_securities, etc. |
| `jurisdiction` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `dataset` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `data` | `repeated BisDataPoint` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "dataset": "<string>",
  "jurisdiction": "<string>"
}
```

#### 响应示例

```json
{
  "dataset": "<string>",
  "data": [
    {
      "date": "<string>",
      "value": "<string>",
      "currency": "<string>"
    }
  ]
}
```

### MacroService/GetTradingEconomics

- **URL**：`POST https://api.alfq.org/antclaw.v1.MacroService/GetTradingEconomics`
- **请求消息**：`GetTradingEconomicsRequest`
- **响应消息**：`GetTradingEconomicsResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `country` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `category` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `indicators` | `repeated TeIndicator` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "country": "<string>",
  "category": "<string>"
}
```

#### 响应示例

```json
{
  "indicators": [
    {
      "country": "<string>",
      "category": "<string>",
      "last": "<string>",
      "previous": "<string>",
      "frequency": "<string>"
    }
  ]
}
```

### MacroService/GetDtccSwaps

- **URL**：`POST https://api.alfq.org/antclaw.v1.MacroService/GetDtccSwaps`
- **请求消息**：`GetDtccSwapsRequest`
- **响应消息**：`GetDtccSwapsResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `tenor` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `swaps` | `repeated DtccSwap` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "pair": "<string>",
  "tenor": "<string>"
}
```

#### 响应示例

```json
{
  "swaps": [
    {
      "date": "<string>",
      "pair": "<string>",
      "tenor": "<string>",
      "volume": 0.0,
      "open_interest": 0.0
    }
  ]
}
```

### MacroService/GetSec13f

- **URL**：`POST https://api.alfq.org/antclaw.v1.MacroService/GetSec13f`
- **请求消息**：`GetSec13fRequest`
- **响应消息**：`GetSec13fResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `cik` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `quarter` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | YYYY-QN format |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `cik` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `filer_name` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `quarter` | `int64` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `holdings` | `repeated Holding13f` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "cik": "<string>",
  "quarter": "<string>"
}
```

#### 响应示例

```json
{
  "cik": "<string>",
  "filer_name": "<string>",
  "quarter": 0,
  "holdings": [
    {
      "cusip": "<string>",
      "issuer": "<string>",
      "class": "<string>",
      "shares": 0,
      "value": "<string>"
    }
  ]
}
```

### MacroService/GetTreasuryAuctions

- **URL**：`POST https://api.alfq.org/antclaw.v1.MacroService/GetTreasuryAuctions`
- **请求消息**：`GetTreasuryAuctionsRequest`
- **响应消息**：`GetTreasuryAuctionsResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `security_type` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `count` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `auctions` | `repeated TreasuryAuction` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "security_type": "<string>",
  "count": 0
}
```

#### 响应示例

```json
{
  "auctions": [
    {
      "date": "<string>",
      "security_type": "<string>",
      "term": "<string>",
      "amount": "<string>",
      "high_yield": 0.0,
      "bid_to_cover": 0.0
    }
  ]
}
```

### MacroService/GetFedWatch

- **URL**：`POST https://api.alfq.org/antclaw.v1.MacroService/GetFedWatch`
- **请求消息**：`GetFedWatchRequest`
- **响应消息**：`GetFedWatchResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `current_target` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `probabilities` | `repeated FedWatchProb` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{}
```

#### 响应示例

```json
{
  "current_target": "<string>",
  "probabilities": [
    {
      "meeting_date": "<string>",
      "target_range": "<string>",
      "probability": 0.0
    }
  ]
}
```

### MacroService/GetWorldBank

- **URL**：`POST https://api.alfq.org/antclaw.v1.MacroService/GetWorldBank`
- **请求消息**：`GetWorldBankRequest`
- **响应消息**：`GetWorldBankResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `indicator` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `country` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `indicator` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `data` | `repeated WbIndicator` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "indicator": "<string>",
  "country": "<string>"
}
```

#### 响应示例

```json
{
  "indicator": "<string>",
  "data": [
    {
      "indicator_code": "<string>",
      "country": "<string>",
      "year": "<string>",
      "value": "<string>"
    }
  ]
}
```

### MacroService/GetImfWeo

- **URL**：`POST https://api.alfq.org/antclaw.v1.MacroService/GetImfWeo`
- **请求消息**：`GetImfWeoRequest`
- **响应消息**：`GetImfWeoResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `country` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `year` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `data` | `repeated ImfWeoData` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "country": "<string>",
  "year": "<string>"
}
```

#### 响应示例

```json
{
  "data": [
    {
      "country": "<string>",
      "year": "<string>",
      "gdp_growth": 0.0,
      "inflation": 0.0,
      "unemployment": 0.0
    }
  ]
}
```

## MacroExtrasService

### MacroExtrasService/GetSeries

- **URL**：`POST https://api.alfq.org/antclaw.v1.MacroExtrasService/GetSeries`
- **请求消息**：`MacroExtrasServiceGetSeriesRequest`
- **响应消息**：`MacroExtrasServiceGetSeriesResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `source` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `series_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `start` | `google.protobuf.Timestamp` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `end` | `google.protobuf.Timestamp` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `points` | `repeated MacroPoint` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `unit` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `frequency` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "source": "<string>",
  "series_id": "<string>",
  "start": {},
  "end": {}
}
```

#### 响应示例

```json
{
  "points": [
    {
      "time": {},
      "value": 0.0,
      "label": "<string>"
    }
  ],
  "unit": "<string>",
  "frequency": "<string>"
}
```

### MacroExtrasService/ListAvailableSeries

- **URL**：`POST https://api.alfq.org/antclaw.v1.MacroExtrasService/ListAvailableSeries`
- **请求消息**：`ListAvailableSeriesRequest`
- **响应消息**：`ListAvailableSeriesResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{}
```

#### 响应示例

```json
{}
```

## MT4Service

### MT4Service/AddAccount

- **URL**：`POST https://api.alfq.org/antclaw.v1.MT4Service/AddAccount`
- **请求消息**：`AddMT4AccountRequest`
- **响应消息**：`MT4Account`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `server` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `account` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `investor_password` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `label` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `is_demo` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `server` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `account` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `label` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `is_demo` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `connected` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `created_at` | `int64` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "server": "<string>",
  "account": "<string>",
  "investor_password": "<string>",
  "label": "<string>",
  "is_demo": false
}
```

#### 响应示例

```json
{
  "id": "<string>",
  "server": "<string>",
  "account": "<string>",
  "label": "<string>",
  "is_demo": false,
  "connected": false,
  "created_at": 0
}
```

### MT4Service/RemoveAccount

- **URL**：`POST https://api.alfq.org/antclaw.v1.MT4Service/RemoveAccount`
- **请求消息**：`RemoveMT4AccountRequest`
- **响应消息**：`RemoveMT4AccountResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{}
```

#### 响应示例

```json
{}
```

### MT4Service/GetAccountInfo

- **URL**：`POST https://api.alfq.org/antclaw.v1.MT4Service/GetAccountInfo`
- **请求消息**：`GetMT4AccountInfoRequest`
- **响应消息**：`MT4AccountInfo`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `balance` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `equity` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `margin` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `free_margin` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `margin_level` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `today_pnl` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `position_count` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `updated_at` | `int64` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{}
```

#### 响应示例

```json
{
  "id": "<string>",
  "balance": 0.0,
  "equity": 0.0,
  "margin": 0.0,
  "free_margin": 0.0,
  "margin_level": 0.0,
  "today_pnl": 0.0,
  "position_count": 0,
  "updated_at": 0
}
```

### MT4Service/GetPositions

- **URL**：`POST https://api.alfq.org/antclaw.v1.MT4Service/GetPositions`
- **请求消息**：`GetMT4PositionsRequest`
- **响应消息**：`MT4PositionsResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{}
```

#### 响应示例

```json
{}
```

### MT4Service/GetHistory

- **URL**：`POST https://api.alfq.org/antclaw.v1.MT4Service/GetHistory`
- **请求消息**：`GetMT4HistoryRequest`
- **响应消息**：`MT4HistoryResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `from_time` | `int64` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `max_count` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `history_orders` | `repeated MT4Order` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `total_count` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "id": "<string>",
  "from_time": 0,
  "max_count": 0
}
```

#### 响应示例

```json
{
  "history_orders": [
    {
      "ticket": 0,
      "symbol": "<string>",
      "type": "<string>",
      "volume": 0.0,
      "open_price": 0.0,
      "close_price": 0.0,
      "profit": 0.0,
      "swap": 0.0,
      "commission": 0.0,
      "open_time": 0,
      "close_time": 0,
      "comment": "<string>"
    }
  ],
  "total_count": 0
}
```

## MT5Service

### MT5Service/AddAccount

- **URL**：`POST https://api.alfq.org/antclaw.v1.MT5Service/AddAccount`
- **请求消息**：`AddMT5AccountRequest`
- **响应消息**：`MT5Account`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `server` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | e.g. "ICMarkets-Demo" |
| `account` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | e.g. "88005522" |
| `investor_password` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 只读密码（强制要求，不能是主密码） |
| `label` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 用户自定义标签（可选） |
| `is_demo` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 模拟盘/实盘标记 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 服务端生成的 UUID |
| `server` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `account` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `label` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `is_demo` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `connected` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 当前是否可连接 |
| `created_at` | `int64` | 否；proto3 默认值语义，业务必填以服务端校验为准 | unix timestamp |

#### 请求示例

```json
{
  "server": "<string>",
  "account": "<string>",
  "investor_password": "<string>",
  "label": "<string>",
  "is_demo": false
}
```

#### 响应示例

```json
{
  "id": "<string>",
  "server": "<string>",
  "account": "<string>",
  "label": "<string>",
  "is_demo": false,
  "connected": false,
  "created_at": 0
}
```

### MT5Service/RemoveAccount

- **URL**：`POST https://api.alfq.org/antclaw.v1.MT5Service/RemoveAccount`
- **请求消息**：`RemoveMT5AccountRequest`
- **响应消息**：`RemoveMT5AccountResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 账号 UUID |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `success` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "id": "<string>"
}
```

#### 响应示例

```json
{
  "success": false
}
```

### MT5Service/GetAccountInfo

- **URL**：`POST https://api.alfq.org/antclaw.v1.MT5Service/GetAccountInfo`
- **请求消息**：`GetMT5AccountInfoRequest`
- **响应消息**：`MT5AccountInfo`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 账号 UUID |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `balance` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 余额 |
| `equity` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 净值 |
| `margin` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 已用保证金 |
| `free_margin` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 可用保证金 |
| `margin_level` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 保证金水平 (%) |
| `today_pnl` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 今日浮动盈亏 |
| `position_count` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 持仓数 |
| `updated_at` | `int64` | 否；proto3 默认值语义，业务必填以服务端校验为准 | unix timestamp |

#### 请求示例

```json
{
  "id": "<string>"
}
```

#### 响应示例

```json
{
  "id": "<string>",
  "balance": 0.0,
  "equity": 0.0,
  "margin": 0.0,
  "free_margin": 0.0,
  "margin_level": 0.0,
  "today_pnl": 0.0,
  "position_count": 0,
  "updated_at": 0
}
```

### MT5Service/GetPositions

- **URL**：`POST https://api.alfq.org/antclaw.v1.MT5Service/GetPositions`
- **请求消息**：`GetMT5PositionsRequest`
- **响应消息**：`MT5PositionsResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 账号 UUID |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `positions` | `repeated MT5Position` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "id": "<string>"
}
```

#### 响应示例

```json
{
  "positions": [
    {
      "ticket": 0,
      "symbol": "<string>",
      "type": "<string>",
      "volume": 0.0,
      "open_price": 0.0,
      "current_price": 0.0,
      "stop_loss": 0.0,
      "take_profit": 0.0,
      "profit": 0.0,
      "swap": 0.0,
      "open_time": 0
    }
  ]
}
```

### MT5Service/GetHistory

- **URL**：`POST https://api.alfq.org/antclaw.v1.MT5Service/GetHistory`
- **请求消息**：`GetMT5HistoryRequest`
- **响应消息**：`MT5HistoryResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 账号 UUID |
| `from_time` | `int64` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 起始时间（unix timestamp），默认最近 90 天 |
| `max_count` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 最大条数，默认 200 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `history_orders` | `repeated MT5Order` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `total_count` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "id": "<string>",
  "from_time": 0,
  "max_count": 0
}
```

#### 响应示例

```json
{
  "history_orders": [
    {
      "ticket": 0,
      "symbol": "<string>",
      "type": "<string>",
      "volume": 0.0,
      "open_price": 0.0,
      "close_price": 0.0,
      "profit": 0.0,
      "swap": 0.0,
      "commission": 0.0,
      "open_time": 0,
      "close_time": 0,
      "comment": "<string>"
    }
  ],
  "total_count": 0
}
```

## NotificationService

### NotificationService/ListUnread

- **URL**：`POST https://api.alfq.org/antclaw.v1.NotificationService/ListUnread`
- **请求消息**：`ListUnreadRequest`
- **响应消息**：`ListUnreadResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{}
```

#### 响应示例

```json
{}
```

### NotificationService/ListHistory

- **URL**：`POST https://api.alfq.org/antclaw.v1.NotificationService/ListHistory`
- **请求消息**：`ListHistoryRequest`
- **响应消息**：`ListHistoryResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{}
```

#### 响应示例

```json
{}
```

### NotificationService/UnreadCount

- **URL**：`POST https://api.alfq.org/antclaw.v1.NotificationService/UnreadCount`
- **请求消息**：`UnreadCountRequest`
- **响应消息**：`UnreadCountResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{}
```

#### 响应示例

```json
{}
```

### NotificationService/MarkRead

- **URL**：`POST https://api.alfq.org/antclaw.v1.NotificationService/MarkRead`
- **请求消息**：`MarkReadRequest`
- **响应消息**：`MarkReadResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{}
```

#### 响应示例

```json
{}
```

### NotificationService/MarkAllRead

- **URL**：`POST https://api.alfq.org/antclaw.v1.NotificationService/MarkAllRead`
- **请求消息**：`MarkAllReadRequest`
- **响应消息**：`MarkAllReadResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{}
```

#### 响应示例

```json
{}
```

### NotificationService/GetPrefs

- **URL**：`POST https://api.alfq.org/antclaw.v1.NotificationService/GetPrefs`
- **请求消息**：`GetPrefsRequest`
- **响应消息**：`GetPrefsResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{}
```

#### 响应示例

```json
{}
```

### NotificationService/UpdatePrefs

- **URL**：`POST https://api.alfq.org/antclaw.v1.NotificationService/UpdatePrefs`
- **请求消息**：`UpdatePrefsRequest`
- **响应消息**：`UpdatePrefsResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{}
```

#### 响应示例

```json
{}
```

### NotificationService/GetAlertPrefs

- **URL**：`POST https://api.alfq.org/antclaw.v1.NotificationService/GetAlertPrefs`
- **请求消息**：`GetAlertPrefsRequest`
- **响应消息**：`GetAlertPrefsResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{}
```

#### 响应示例

```json
{}
```

### NotificationService/UpdateAlertPrefs

- **URL**：`POST https://api.alfq.org/antclaw.v1.NotificationService/UpdateAlertPrefs`
- **请求消息**：`UpdateAlertPrefsRequest`
- **响应消息**：`UpdateAlertPrefsResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{}
```

#### 响应示例

```json
{}
```

## OnchainService

### OnchainService/GetMetrics

- **URL**：`POST https://api.alfq.org/antclaw.v1.OnchainService/GetMetrics`
- **请求消息**：`OnchainServiceGetMetricsRequest`
- **响应消息**：`OnchainServiceGetMetricsResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `asset` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `start` | `google.protobuf.Timestamp` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `end` | `google.protobuf.Timestamp` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{
  "asset": "<string>",
  "start": {},
  "end": {}
}
```

#### 响应示例

```json
{}
```

### OnchainService/GetAnalysis

- **URL**：`POST https://api.alfq.org/antclaw.v1.OnchainService/GetAnalysis`
- **请求消息**：`OnchainServiceGetAnalysisRequest`
- **响应消息**：`OnchainServiceGetAnalysisResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `regime` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `confidence` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `narrative` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{}
```

#### 响应示例

```json
{
  "regime": "<string>",
  "confidence": 0.0,
  "narrative": "<string>"
}
```

## OptionsService

### OptionsService/GetGEX

- **URL**：`POST https://api.alfq.org/antclaw.v1.OptionsService/GetGEX`
- **请求消息**：`GetGEXRequest`
- **响应消息**：`GetGEXResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `asset` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `venue` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `total_gex` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `zero_gamma` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `strikes` | `repeated GEXBucket` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "asset": "<string>",
  "venue": "<string>"
}
```

#### 响应示例

```json
{
  "total_gex": 0.0,
  "zero_gamma": 0.0,
  "strikes": [
    {
      "strike": 0.0,
      "call_gex": 0.0,
      "put_gex": 0.0,
      "net_gex": 0.0
    }
  ]
}
```

### OptionsService/GetIVSurface

- **URL**：`POST https://api.alfq.org/antclaw.v1.OptionsService/GetIVSurface`
- **请求消息**：`GetIVSurfaceRequest`
- **响应消息**：`GetIVSurfaceResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{}
```

#### 响应示例

```json
{}
```

### OptionsService/GetOptionsSkew

- **URL**：`POST https://api.alfq.org/antclaw.v1.OptionsService/GetOptionsSkew`
- **请求消息**：`GetOptionsSkewRequest`
- **响应消息**：`GetOptionsSkewResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `rr_25d` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `bf_25d` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `atm_iv` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `skew_z` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{}
```

#### 响应示例

```json
{
  "rr_25d": 0.0,
  "bf_25d": 0.0,
  "atm_iv": 0.0,
  "skew_z": 0.0
}
```

### OptionsService/GetIVAlerts

- **URL**：`POST https://api.alfq.org/antclaw.v1.OptionsService/GetIVAlerts`
- **请求消息**：`GetIVAlertsRequest`
- **响应消息**：`GetIVAlertsResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{}
```

#### 响应示例

```json
{}
```

## PriceService

### PriceService/GetPrice

- **URL**：`POST https://api.alfq.org/antclaw.v1.PriceService/GetPrice`
- **请求消息**：`GetPriceRequest`
- **响应消息**：`GetPriceResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `timeframe` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | e.g., "1m", "5m", "1h", "4h", "1d" |
| `count` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `current` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `change_24h` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `change_pct_24h` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `bars` | `repeated PriceBar` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "pair": "<string>",
  "timeframe": "<string>",
  "count": 0
}
```

#### 响应示例

```json
{
  "pair": "<string>",
  "current": "<string>",
  "change_24h": "<string>",
  "change_pct_24h": "<string>",
  "bars": [
    {
      "timestamp": "<string>",
      "open": "<string>",
      "high": "<string>",
      "low": "<string>",
      "close": "<string>",
      "volume": 0
    }
  ]
}
```

### PriceService/GetLevels

- **URL**：`POST https://api.alfq.org/antclaw.v1.PriceService/GetLevels`
- **请求消息**：`GetLevelsRequest`
- **响应消息**：`GetLevelsResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `timeframe` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `levels` | `repeated PriceLevel` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "pair": "<string>",
  "timeframe": "<string>"
}
```

#### 响应示例

```json
{
  "pair": "<string>",
  "levels": [
    {
      "price": "<string>",
      "type": "<string>",
      "strength": 0.0
    }
  ]
}
```

### PriceService/GetMarketOverview

- **URL**：`POST https://api.alfq.org/antclaw.v1.PriceService/GetMarketOverview`
- **请求消息**：`GetMarketOverviewRequest`
- **响应消息**：`GetMarketOverviewResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `category` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | majors, crosses, all |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `items` | `repeated MarketOverviewItem` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "category": "<string>"
}
```

#### 响应示例

```json
{
  "items": [
    {
      "pair": "<string>",
      "price": "<string>",
      "change_24h": "<string>",
      "trend": "<string>"
    }
  ]
}
```

### PriceService/GetSession

- **URL**：`POST https://api.alfq.org/antclaw.v1.PriceService/GetSession`
- **请求消息**：`GetSessionRequest`
- **响应消息**：`GetSessionResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `sessions` | `repeated SessionInfo` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `current_session` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "pair": "<string>"
}
```

#### 响应示例

```json
{
  "pair": "<string>",
  "sessions": [
    {
      "session": "<string>",
      "is_open": false,
      "opens_at": "<string>",
      "closes_at": "<string>",
      "volatility_index": 0.0
    }
  ],
  "current_session": "<string>"
}
```

### PriceService/RunScenario

- **URL**：`POST https://api.alfq.org/antclaw.v1.PriceService/RunScenario`
- **请求消息**：`RunScenarioRequest`
- **响应消息**：`RunScenarioResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `event_type` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `results` | `repeated ScenarioResult` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "pair": "<string>",
  "event_type": "<string>"
}
```

#### 响应示例

```json
{
  "results": [
    {
      "scenario_name": "<string>",
      "outcome": "<string>",
      "probability": 0.0
    }
  ]
}
```

### PriceService/GetRegime

- **URL**：`POST https://api.alfq.org/antclaw.v1.PriceService/GetRegime`
- **请求消息**：`GetRegimeRequest`
- **响应消息**：`GetRegimeResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `timeframe` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `engine` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `n_states` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `regime` | `MarketRegime` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `recent_regimes` | `repeated string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `engine_used` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "pair": "<string>",
  "timeframe": "<string>",
  "engine": "<string>",
  "n_states": 0
}
```

#### 响应示例

```json
{
  "regime": {
    "regime": "<string>",
    "confidence": 0.0,
    "since": "<string>"
  },
  "recent_regimes": [
    "<string>"
  ],
  "engine_used": "<string>"
}
```

### PriceService/GetSeasonal

- **URL**：`POST https://api.alfq.org/antclaw.v1.PriceService/GetSeasonal`
- **请求消息**：`GetSeasonalRequest`
- **响应消息**：`GetSeasonalResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `years` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `data` | `repeated SeasonalDataPoint` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "pair": "<string>",
  "years": 0
}
```

#### 响应示例

```json
{
  "pair": "<string>",
  "data": [
    {
      "month": "<string>",
      "avg_return": 0.0,
      "win_rate": 0.0
    }
  ]
}
```

### PriceService/GetVolatility

- **URL**：`POST https://api.alfq.org/antclaw.v1.PriceService/GetVolatility`
- **请求消息**：`GetVolatilityRequest`
- **响应消息**：`GetVolatilityResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `timeframe` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 默认 1d |
| `lookback` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 用最近 N 根 K 线，默认 500，最大 5000 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `omega` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `alpha` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `beta` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `persistence` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | alpha + beta |
| `unconditional_vol` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | sqrt(omega/(1-alpha-beta)) 年化 |
| `next_step_forecast` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 1 步预测，年化标准差 |
| `series` | `repeated VolatilityPoint` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "pair": "<string>",
  "timeframe": "<string>",
  "lookback": 0
}
```

#### 响应示例

```json
{
  "pair": "<string>",
  "omega": 0.0,
  "alpha": 0.0,
  "beta": 0.0,
  "persistence": 0.0,
  "unconditional_vol": 0.0,
  "next_step_forecast": 0.0,
  "series": [
    {
      "timestamp": "<string>",
      "conditional_vol": 0.0
    }
  ]
}
```

### PriceService/GetHurst

- **URL**：`POST https://api.alfq.org/antclaw.v1.PriceService/GetHurst`
- **请求消息**：`GetHurstRequest`
- **响应消息**：`GetHurstResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `timeframe` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `lookback` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 默认 500 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `hurst` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | [0, 1] |
| `interpretation` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | "trending" / "mean_reverting" / "random_walk" |
| `sample_size` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "pair": "<string>",
  "timeframe": "<string>",
  "lookback": 0
}
```

#### 响应示例

```json
{
  "pair": "<string>",
  "hurst": 0.0,
  "interpretation": "<string>",
  "sample_size": 0
}
```

### PriceService/GetCorrelations

- **URL**：`POST https://api.alfq.org/antclaw.v1.PriceService/GetCorrelations`
- **请求消息**：`GetCorrelationsRequest`
- **响应消息**：`GetCorrelationsResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `assets` | `repeated string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 资产列表；空时使用默认 8 主流货币对 |
| `timeframe` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `window` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 滚动窗 K 线数；默认 30 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `assets` | `repeated string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `window` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `matrix` | `repeated CorrelationCell` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "assets": [
    "<string>"
  ],
  "timeframe": "<string>",
  "window": 0
}
```

#### 响应示例

```json
{
  "assets": [
    "<string>"
  ],
  "window": 0,
  "matrix": [
    {
      "asset_a": "<string>",
      "asset_b": "<string>",
      "value": 0.0
    }
  ]
}
```

### PriceService/GetDivergences

- **URL**：`POST https://api.alfq.org/antclaw.v1.PriceService/GetDivergences`
- **请求消息**：`GetDivergencesRequest`
- **响应消息**：`GetDivergencesResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `timeframe` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `lookback` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 默认 200 |
| `indicators` | `repeated string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `events` | `repeated DivergenceEvent` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "pair": "<string>",
  "timeframe": "<string>",
  "lookback": 0,
  "indicators": [
    "<string>"
  ]
}
```

#### 响应示例

```json
{
  "pair": "<string>",
  "events": [
    {
      "indicator": "<string>",
      "kind": "<string>",
      "detected_at": "<string>",
      "price_pivot": 0.0,
      "indicator_pivot": 0.0,
      "note": "<string>"
    }
  ]
}
```

## RegimeService

### RegimeService/GetOverlay

- **URL**：`POST https://api.alfq.org/antclaw.v1.RegimeService/GetOverlay`
- **请求消息**：`GetOverlayRequest`
- **响应消息**：`GetOverlayResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `symbol` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 例如 "EURUSD"、"BTCUSDT" |
| `timeframe` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | "D"、"4H"、"1H" |
| `contract_code` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | CFTC 合约代码（COT 子模型用），可选 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `symbol` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `timeframe` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `computed_at` | `google.protobuf.Timestamp` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `unified_score` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 加权融合分数 -100..+100 |
| `unified_label` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | STRONG_BULL / BULL / NEUTRAL / BEAR / STRONG_BEAR |
| `hmm` | `RegimeSubModel` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `garch` | `RegimeSubModel` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `adx` | `RegimeSubModel` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `cot` | `RegimeSubModel` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `available_models` | `repeated string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 实际参与融合的模型名 |

#### 请求示例

```json
{
  "symbol": "<string>",
  "timeframe": "<string>",
  "contract_code": "<string>"
}
```

#### 响应示例

```json
{
  "symbol": "<string>",
  "timeframe": "<string>",
  "computed_at": {},
  "unified_score": 0.0,
  "unified_label": "<string>",
  "hmm": {
    "name": "<string>",
    "available": false,
    "score": 0.0,
    "weight": 0.0,
    "state": "<string>",
    "confidence": 0.0
  },
  "garch": {
    "name": "<string>",
    "available": false,
    "score": 0.0,
    "weight": 0.0,
    "state": "<string>",
    "confidence": 0.0
  },
  "adx": {
    "name": "<string>",
    "available": false,
    "score": 0.0,
    "weight": 0.0,
    "state": "<string>",
    "confidence": 0.0
  },
  "cot": {
    "name": "<string>",
    "available": false,
    "score": 0.0,
    "weight": 0.0,
    "state": "<string>",
    "confidence": 0.0
  },
  "available_models": [
    "<string>"
  ]
}
```

### RegimeService/ListRecent

- **URL**：`POST https://api.alfq.org/antclaw.v1.RegimeService/ListRecent`
- **请求消息**：`ListRecentRequest`
- **响应消息**：`ListRecentResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `symbol` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `timeframe` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `days` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 默认 30 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `items` | `repeated OverlaySnapshot` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "symbol": "<string>",
  "timeframe": "<string>",
  "days": 0
}
```

#### 响应示例

```json
{
  "items": [
    {
      "time": {},
      "unified_score": 0.0,
      "unified_label": "<string>"
    }
  ]
}
```

## ReportService

### ReportService/GetReport

- **URL**：`POST https://api.alfq.org/antclaw.v1.ReportService/GetReport`
- **请求消息**：`GetReportRequest`
- **响应消息**：`GetReportResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `symbol` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `sections` | `repeated string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `with_ai_summary` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `symbol` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `generated_at` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `bias` | `ReportBias` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `unified` | `ReportUnified` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `accuracy_1w` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `accuracy_1m` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `missing_sections` | `repeated string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "symbol": "<string>",
  "sections": [
    "<string>"
  ],
  "with_ai_summary": false
}
```

#### 响应示例

```json
{
  "symbol": "<string>",
  "generated_at": "<string>",
  "bias": {
    "pair": "<string>",
    "direction": "<string>",
    "confidence": 0.0,
    "timeframe": "<string>"
  },
  "unified": {
    "symbol": "<string>",
    "issued_at": "<string>",
    "recommendation": "<string>",
    "unified_score": 0.0,
    "confidence": 0.0
  },
  "accuracy_1w": 0.0,
  "accuracy_1m": 0.0,
  "missing_sections": [
    "<string>"
  ]
}
```

## SECService

### SECService/ListFilings

- **URL**：`POST https://api.alfq.org/antclaw.v1.SECService/ListFilings`
- **请求消息**：`ListFilingsRequest`
- **响应消息**：`ListFilingsResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `cik` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `form_type` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `limit` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{
  "cik": "<string>",
  "form_type": "<string>",
  "limit": 0
}
```

#### 响应示例

```json
{}
```

### SECService/GetFiling

- **URL**：`POST https://api.alfq.org/antclaw.v1.SECService/GetFiling`
- **请求消息**：`GetFilingRequest`
- **响应消息**：`GetFilingResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `filing` | `SECFiling` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `raw_text_excerpt` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{}
```

#### 响应示例

```json
{
  "filing": {
    "accession_number": "<string>",
    "form_type": "<string>",
    "filed_at": {},
    "company_name": "<string>",
    "url": "<string>"
  },
  "raw_text_excerpt": "<string>"
}
```

### SECService/GetAnalysis

- **URL**：`POST https://api.alfq.org/antclaw.v1.SECService/GetAnalysis`
- **请求消息**：`SECServiceGetAnalysisRequest`
- **响应消息**：`SECServiceGetAnalysisResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `sentiment` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `risk_score` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `highlights` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{}
```

#### 响应示例

```json
{
  "sentiment": "<string>",
  "risk_score": 0.0,
  "highlights": "<string>"
}
```

## SentimentService

### SentimentService/GetSentiment

- **URL**：`POST https://api.alfq.org/antclaw.v1.SentimentService/GetSentiment`
- **请求消息**：`GetSentimentRequest`
- **响应消息**：`GetSentimentResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `asset` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `sentiment` | `SentimentData` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `components` | `repeated SentimentData` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "asset": "<string>"
}
```

#### 响应示例

```json
{
  "sentiment": {
    "asset": "<string>",
    "score": 0.0,
    "label": "<string>",
    "source": "<string>",
    "timestamp": "<string>"
  },
  "components": [
    {
      "asset": "<string>",
      "score": 0.0,
      "label": "<string>",
      "source": "<string>",
      "timestamp": "<string>"
    }
  ]
}
```

### SentimentService/GetOnchain

- **URL**：`POST https://api.alfq.org/antclaw.v1.SentimentService/GetOnchain`
- **请求消息**：`GetOnchainRequest`
- **响应消息**：`GetOnchainResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `asset` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `asset` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `metrics` | `repeated OnchainMetric` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `signal` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "asset": "<string>"
}
```

#### 响应示例

```json
{
  "asset": "<string>",
  "metrics": [
    {
      "name": "<string>",
      "value": 0.0,
      "trend": "<string>"
    }
  ],
  "signal": "<string>"
}
```

### SentimentService/GetDefiHealth

- **URL**：`POST https://api.alfq.org/antclaw.v1.SentimentService/GetDefiHealth`
- **请求消息**：`GetDefiHealthRequest`
- **响应消息**：`GetDefiHealthResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `chain` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `chain` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `protocols` | `repeated DefiMetric` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `overall_health` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "chain": "<string>"
}
```

#### 响应示例

```json
{
  "chain": "<string>",
  "protocols": [
    {
      "protocol": "<string>",
      "tvl": "<string>",
      "tvl_change_24h": "<string>",
      "utilization_rate": 0.0,
      "health_score": "<string>"
    }
  ],
  "overall_health": "<string>"
}
```

### SentimentService/GetCarryMonitor

- **URL**：`POST https://api.alfq.org/antclaw.v1.SentimentService/GetCarryMonitor`
- **请求消息**：`GetCarryMonitorRequest`
- **响应消息**：`GetCarryMonitorResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `category` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | fx, crypto |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `carries` | `repeated CarryData` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "category": "<string>"
}
```

#### 响应示例

```json
{
  "carries": [
    {
      "pair": "<string>",
      "spot": "<string>",
      "futures": "<string>",
      "basis": "<string>",
      "annualized_yield": "<string>"
    }
  ]
}
```

## SentimentExtrasService

### SentimentExtrasService/GetCBOEPutCall

- **URL**：`POST https://api.alfq.org/antclaw.v1.SentimentExtrasService/GetCBOEPutCall`
- **请求消息**：`GetCBOEPutCallRequest`
- **响应消息**：`GetCBOEPutCallResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `date` | `google.protobuf.Timestamp` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `total_pc` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `equity_pc` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `index_pc` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{}
```

#### 响应示例

```json
{
  "date": {},
  "total_pc": 0.0,
  "equity_pc": 0.0,
  "index_pc": 0.0
}
```

### SentimentExtrasService/GetMyFXBookPositions

- **URL**：`POST https://api.alfq.org/antclaw.v1.SentimentExtrasService/GetMyFXBookPositions`
- **请求消息**：`GetMyFXBookPositionsRequest`
- **响应消息**：`GetMyFXBookPositionsResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `symbol` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `long_pct` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `short_pct` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `long_lots` | `int64` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `short_lots` | `int64` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{}
```

#### 响应示例

```json
{
  "symbol": "<string>",
  "long_pct": 0.0,
  "short_pct": 0.0,
  "long_lots": 0,
  "short_lots": 0
}
```

### SentimentExtrasService/GetInsiderTrades

- **URL**：`POST https://api.alfq.org/antclaw.v1.SentimentExtrasService/GetInsiderTrades`
- **请求消息**：`GetInsiderTradesRequest`
- **响应消息**：`GetInsiderTradesResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{}
```

#### 响应示例

```json
{}
```

### SentimentExtrasService/GetCryptoSocial

- **URL**：`POST https://api.alfq.org/antclaw.v1.SentimentExtrasService/GetCryptoSocial`
- **请求消息**：`GetCryptoSocialRequest`
- **响应消息**：`GetCryptoSocialResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `date` | `google.protobuf.Timestamp` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `twitter_followers_growth` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `reddit_subscribers_growth` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `sentiment_score` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{}
```

#### 响应示例

```json
{
  "date": {},
  "twitter_followers_growth": 0.0,
  "reddit_subscribers_growth": 0.0,
  "sentiment_score": 0.0
}
```

### SentimentExtrasService/GetFinvizMetrics

- **URL**：`POST https://api.alfq.org/antclaw.v1.SentimentExtrasService/GetFinvizMetrics`
- **请求消息**：`GetFinvizMetricsRequest`
- **响应消息**：`GetFinvizMetricsResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `short_ratio` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `short_pct_float` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `inst_own_pct` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{}
```

#### 响应示例

```json
{
  "short_ratio": 0.0,
  "short_pct_float": 0.0,
  "inst_own_pct": 0.0
}
```

## SignalsService

### SignalsService/GetBias

- **URL**：`POST https://api.alfq.org/antclaw.v1.SignalsService/GetBias`
- **请求消息**：`GetBiasRequest`
- **响应消息**：`GetBiasResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `timeframe` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `biases` | `repeated BiasData` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "pair": "<string>",
  "timeframe": "<string>"
}
```

#### 响应示例

```json
{
  "biases": [
    {
      "pair": "<string>",
      "direction": "<string>",
      "confidence": 0.0,
      "timeframe": "<string>"
    }
  ]
}
```

### SignalsService/GetRank

- **URL**：`POST https://api.alfq.org/antclaw.v1.SignalsService/GetRank`
- **请求消息**：`GetRankRequest`
- **响应消息**：`GetRankResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `category` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | majors, crosses, crypto, all |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `rankings` | `repeated RankItem` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "category": "<string>"
}
```

#### 响应示例

```json
{
  "rankings": [
    {
      "pair": "<string>",
      "rank": 0,
      "score": 0.0,
      "trend": "<string>"
    }
  ]
}
```

### SignalsService/GetXFactors

- **URL**：`POST https://api.alfq.org/antclaw.v1.SignalsService/GetXFactors`
- **请求消息**：`GetXFactorsRequest`
- **响应消息**：`GetXFactorsResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `factors` | `repeated XFactor` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `composite_signal` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "pair": "<string>"
}
```

#### 响应示例

```json
{
  "pair": "<string>",
  "factors": [
    {
      "name": "<string>",
      "weight": 0.0,
      "direction": "<string>",
      "description": "<string>"
    }
  ],
  "composite_signal": "<string>"
}
```

### SignalsService/GetRadar

- **URL**：`POST https://api.alfq.org/antclaw.v1.SignalsService/GetRadar`
- **请求消息**：`GetRadarRequest`
- **响应消息**：`GetRadarResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `category` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `points` | `repeated RadarDataPoint` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "category": "<string>"
}
```

#### 响应示例

```json
{
  "points": [
    {
      "pair": "<string>",
      "x": 0.0,
      "y": 0.0,
      "quadrant": "<string>",
      "strength": 0.0
    }
  ]
}
```

### SignalsService/GetIntensity

- **URL**：`POST https://api.alfq.org/antclaw.v1.SignalsService/GetIntensity`
- **请求消息**：`GetIntensityRequest`
- **响应消息**：`GetIntensityResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `intensity` | `IntensityData` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "pair": "<string>"
}
```

#### 响应示例

```json
{
  "intensity": {
    "pair": "<string>",
    "intensity": 0.0,
    "strength_label": "<string>",
    "percentile_30d": 0.0
  }
}
```

### SignalsService/GetTransition

- **URL**：`POST https://api.alfq.org/antclaw.v1.SignalsService/GetTransition`
- **请求消息**：`GetTransitionRequest`
- **响应消息**：`GetTransitionResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `current_state` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `transitions` | `repeated TransitionProb` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "pair": "<string>",
  "current_state": "<string>"
}
```

#### 响应示例

```json
{
  "pair": "<string>",
  "transitions": [
    {
      "from_state": "<string>",
      "to_state": "<string>",
      "probability": 0.0
    }
  ]
}
```

### SignalsService/GetCryptoAlpha

- **URL**：`POST https://api.alfq.org/antclaw.v1.SignalsService/GetCryptoAlpha`
- **请求消息**：`GetCryptoAlphaRequest`
- **响应消息**：`GetCryptoAlphaResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `asset_filter` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `signals` | `repeated CryptoAlphaSignal` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "asset_filter": "<string>"
}
```

#### 响应示例

```json
{
  "signals": [
    {
      "asset": "<string>",
      "signal_type": "<string>",
      "confidence": 0.0,
      "timeframe": "<string>"
    }
  ]
}
```

### SignalsService/GetUnified

- **URL**：`POST https://api.alfq.org/antclaw.v1.SignalsService/GetUnified`
- **请求消息**：`GetUnifiedRequest`
- **响应消息**：`GetUnifiedResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `signal` | `UnifiedSignal` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "pair": "<string>"
}
```

#### 响应示例

```json
{
  "signal": {
    "pair": "<string>",
    "direction": "<string>",
    "confidence": 0.0,
    "contributing_factors": [
      "<string>"
    ]
  }
}
```

### SignalsService/GetQuant

- **URL**：`POST https://api.alfq.org/antclaw.v1.SignalsService/GetQuant`
- **请求消息**：`GetQuantRequest`
- **响应消息**：`GetQuantResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `signals` | `repeated QuantSignal` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "pair": "<string>"
}
```

#### 响应示例

```json
{
  "signals": [
    {
      "pair": "<string>",
      "strategy": "<string>",
      "signal": "<string>",
      "sharpe": 0.0,
      "drawdown": 0.0
    }
  ]
}
```

### SignalsService/GetCta

- **URL**：`POST https://api.alfq.org/antclaw.v1.SignalsService/GetCta`
- **请求消息**：`GetCtaRequest`
- **响应消息**：`GetCtaResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `signal` | `CtaSignal` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "pair": "<string>"
}
```

#### 响应示例

```json
{
  "signal": {
    "pair": "<string>",
    "trend": "<string>",
    "momentum": 0.0,
    "regime": "<string>"
  }
}
```

### SignalsService/GetBriefing

- **URL**：`POST https://api.alfq.org/antclaw.v1.SignalsService/GetBriefing`
- **请求消息**：`GetBriefingRequest`
- **响应消息**：`GetBriefingResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `category` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | fx, crypto, macro, all |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `sections` | `repeated BriefingSection` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `generated_at` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "category": "<string>"
}
```

#### 响应示例

```json
{
  "sections": [
    {
      "title": "<string>",
      "content": "<string>",
      "priority": "<string>"
    }
  ],
  "generated_at": "<string>"
}
```

### SignalsService/GetOutlook

- **URL**：`POST https://api.alfq.org/antclaw.v1.SignalsService/GetOutlook`
- **请求消息**：`GetOutlookRequest`
- **响应消息**：`GetOutlookResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `horizon` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `outlooks` | `repeated OutlookData` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "pair": "<string>",
  "horizon": "<string>"
}
```

#### 响应示例

```json
{
  "outlooks": [
    {
      "pair": "<string>",
      "outlook": "<string>",
      "confidence": 0.0,
      "horizon": "<string>"
    }
  ]
}
```

### SignalsService/FitCalibration

- **URL**：`POST https://api.alfq.org/antclaw.v1.SignalsService/FitCalibration`
- **请求消息**：`FitCalibrationRequest`
- **响应消息**：`FitCalibrationResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `model_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 自定义标识 |
| `type` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 'platt' / 'isotonic' |
| `scores` | `repeated double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `outcomes` | `repeated bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `model_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `type` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `n_samples` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `brier` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "model_id": "<string>",
  "type": "<string>",
  "scores": [
    0.0
  ],
  "outcomes": [
    false
  ]
}
```

#### 响应示例

```json
{
  "model_id": "<string>",
  "type": "<string>",
  "n_samples": 0,
  "brier": 0.0
}
```

### SignalsService/PredictCalibrated

- **URL**：`POST https://api.alfq.org/antclaw.v1.SignalsService/PredictCalibrated`
- **请求消息**：`PredictCalibratedRequest`
- **响应消息**：`PredictCalibratedResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `model_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `score` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `model_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `calibrated` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "model_id": "<string>",
  "score": 0.0
}
```

#### 响应示例

```json
{
  "model_id": "<string>",
  "calibrated": 0.0
}
```

### SignalsService/ListCalibrations

- **URL**：`POST https://api.alfq.org/antclaw.v1.SignalsService/ListCalibrations`
- **请求消息**：`ListCalibrationsRequest`
- **响应消息**：`ListCalibrationsResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `items` | `repeated CalibrationSummary` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{}
```

#### 响应示例

```json
{
  "items": [
    {
      "model_id": "<string>",
      "type": "<string>",
      "n_samples": 0,
      "brier": 0.0,
      "fitted_at": "<string>"
    }
  ]
}
```

## StrategyService

### StrategyService/ListStrategies

- **URL**：`POST https://api.alfq.org/antclaw.v1.StrategyService/ListStrategies`
- **请求消息**：`ListStrategiesRequest`
- **响应消息**：`ListStrategiesResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `offset` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `limit` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `items` | `repeated Strategy` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `total` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "offset": 0,
  "limit": 0
}
```

#### 响应示例

```json
{
  "items": [
    {
      "id": "<string>",
      "name": "<string>",
      "kind": "<string>",
      "symbol": "<string>",
      "timeframe": "<string>",
      "params": {},
      "schedule_cron": "<string>",
      "enabled": false,
      "status": "<string>",
      "description": "<string>",
      "created_at": "<string>",
      "updated_at": "<string>"
    }
  ],
  "total": 0
}
```

### StrategyService/GetStrategy

- **URL**：`POST https://api.alfq.org/antclaw.v1.StrategyService/GetStrategy`
- **请求消息**：`GetStrategyRequest`
- **响应消息**：`GetStrategyResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `item` | `Strategy` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "id": "<string>"
}
```

#### 响应示例

```json
{
  "item": {
    "id": "<string>",
    "name": "<string>",
    "kind": "<string>",
    "symbol": "<string>",
    "timeframe": "<string>",
    "params": {},
    "schedule_cron": "<string>",
    "enabled": false,
    "status": "<string>",
    "description": "<string>",
    "created_at": "<string>",
    "updated_at": "<string>"
  }
}
```

### StrategyService/CreateStrategy

- **URL**：`POST https://api.alfq.org/antclaw.v1.StrategyService/CreateStrategy`
- **请求消息**：`CreateStrategyRequest`
- **响应消息**：`CreateStrategyResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `name` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `kind` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `symbol` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `timeframe` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `params` | `google.protobuf.Struct` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `schedule_cron` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `description` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `item` | `Strategy` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "name": "<string>",
  "kind": "<string>",
  "symbol": "<string>",
  "timeframe": "<string>",
  "params": {},
  "schedule_cron": "<string>",
  "description": "<string>"
}
```

#### 响应示例

```json
{
  "item": {
    "id": "<string>",
    "name": "<string>",
    "kind": "<string>",
    "symbol": "<string>",
    "timeframe": "<string>",
    "params": {},
    "schedule_cron": "<string>",
    "enabled": false,
    "status": "<string>",
    "description": "<string>",
    "created_at": "<string>",
    "updated_at": "<string>"
  }
}
```

### StrategyService/UpdateStrategy

- **URL**：`POST https://api.alfq.org/antclaw.v1.StrategyService/UpdateStrategy`
- **请求消息**：`UpdateStrategyRequest`
- **响应消息**：`UpdateStrategyResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `name` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `kind` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `symbol` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `timeframe` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `params` | `google.protobuf.Struct` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `schedule_cron` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `enabled` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `status` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `description` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "id": "<string>",
  "name": "<string>",
  "kind": "<string>",
  "symbol": "<string>",
  "timeframe": "<string>",
  "params": {},
  "schedule_cron": "<string>",
  "enabled": false,
  "status": "<string>",
  "description": "<string>"
}
```

#### 响应示例

```json
{
  "id": "<string>"
}
```

### StrategyService/DeleteStrategy

- **URL**：`POST https://api.alfq.org/antclaw.v1.StrategyService/DeleteStrategy`
- **请求消息**：`DeleteStrategyRequest`
- **响应消息**：`DeleteStrategyResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{
  "id": "<string>"
}
```

#### 响应示例

```json
{}
```

### StrategyService/EnableStrategy

- **URL**：`POST https://api.alfq.org/antclaw.v1.StrategyService/EnableStrategy`
- **请求消息**：`EnableStrategyRequest`
- **响应消息**：`EnableStrategyResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `enabled` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "id": "<string>"
}
```

#### 响应示例

```json
{
  "id": "<string>",
  "enabled": false
}
```

### StrategyService/DisableStrategy

- **URL**：`POST https://api.alfq.org/antclaw.v1.StrategyService/DisableStrategy`
- **请求消息**：`DisableStrategyRequest`
- **响应消息**：`DisableStrategyResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `enabled` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "id": "<string>"
}
```

#### 响应示例

```json
{
  "id": "<string>",
  "enabled": false
}
```

### StrategyService/RunStrategy

- **URL**：`POST https://api.alfq.org/antclaw.v1.StrategyService/RunStrategy`
- **请求消息**：`RunStrategyRequest`
- **响应消息**：`RunStrategyResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `item` | `StrategyRun` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "id": "<string>"
}
```

#### 响应示例

```json
{
  "item": {
    "run_id": "<string>",
    "strategy_id": "<string>",
    "started_at": "<string>",
    "finished_at": "<string>",
    "status": "<string>",
    "metrics": {},
    "mock": false,
    "error_message": "<string>"
  }
}
```

### StrategyService/ListStrategyRuns

- **URL**：`POST https://api.alfq.org/antclaw.v1.StrategyService/ListStrategyRuns`
- **请求消息**：`ListStrategyRunsRequest`
- **响应消息**：`ListStrategyRunsResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `limit` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `items` | `repeated StrategyRun` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "id": "<string>",
  "limit": 0
}
```

#### 响应示例

```json
{
  "items": [
    {
      "run_id": "<string>",
      "strategy_id": "<string>",
      "started_at": "<string>",
      "finished_at": "<string>",
      "status": "<string>",
      "metrics": {},
      "mock": false,
      "error_message": "<string>"
    }
  ]
}
```

## StreamService

### StreamService/SubscribeEvents

- **URL**：`POST https://api.alfq.org/antclaw.v1.StreamService/SubscribeEvents`
- **请求消息**：`SubscribeEventsRequest`
- **响应消息**：`SubscribeEventsResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `channel` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `last_event_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `type` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `payload` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `timestamp` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "channel": "<string>",
  "last_event_id": "<string>"
}
```

#### 响应示例

```json
{
  "id": "<string>",
  "type": "<string>",
  "payload": "<string>",
  "timestamp": "<string>"
}
```

## SystemService

### SystemService/Healthz

- **URL**：`POST https://api.alfq.org/antclaw.v1.SystemService/Healthz`
- **请求消息**：`HealthzRequest`
- **响应消息**：`HealthzResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `status` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | healthy | degraded | unhealthy |
| `checked_at` | `google.protobuf.Timestamp` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{}
```

#### 响应示例

```json
{
  "status": "<string>",
  "checked_at": {}
}
```

### SystemService/Readyz

- **URL**：`POST https://api.alfq.org/antclaw.v1.SystemService/Readyz`
- **请求消息**：`ReadyzRequest`
- **响应消息**：`ReadyzResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{}
```

#### 响应示例

```json
{}
```

### SystemService/Info

- **URL**：`POST https://api.alfq.org/antclaw.v1.SystemService/Info`
- **请求消息**：`InfoRequest`
- **响应消息**：`InfoResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `version` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `git_commit` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `built_at` | `google.protobuf.Timestamp` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{}
```

#### 响应示例

```json
{
  "version": "<string>",
  "git_commit": "<string>",
  "built_at": {}
}
```

## SystemAIService

### SystemAIService/ListConfigs

- **URL**：`POST https://api.alfq.org/antclaw.v1.SystemAIService/ListConfigs`
- **请求消息**：`ListSystemAIConfigsRequest`
- **响应消息**：`ListSystemAIConfigsResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `items` | `repeated SystemAIConfig` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{}
```

#### 响应示例

```json
{
  "items": [
    {
      "provider_id": "<string>",
      "name": "<string>",
      "base_url": "<string>",
      "organization": "<string>",
      "models": [
        "<string>"
      ],
      "default_model": "<string>",
      "temperature": 0.0,
      "timeout_seconds": 0,
      "max_tokens": 0,
      "purposes": [
        "<string>"
      ],
      "primary_for": [
        "<string>"
      ],
      "enabled": false
    }
  ]
}
```

### SystemAIService/GetConfig

- **URL**：`POST https://api.alfq.org/antclaw.v1.SystemAIService/GetConfig`
- **请求消息**：`GetSystemAIConfigRequest`
- **响应消息**：`GetSystemAIConfigResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `provider_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `item` | `SystemAIConfig` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "provider_id": "<string>"
}
```

#### 响应示例

```json
{
  "item": {
    "provider_id": "<string>",
    "name": "<string>",
    "base_url": "<string>",
    "organization": "<string>",
    "models": [
      "<string>"
    ],
    "default_model": "<string>",
    "temperature": 0.0,
    "timeout_seconds": 0,
    "max_tokens": 0,
    "purposes": [
      "<string>"
    ],
    "primary_for": [
      "<string>"
    ],
    "enabled": false
  }
}
```

### SystemAIService/UpdateConfig

- **URL**：`POST https://api.alfq.org/antclaw.v1.SystemAIService/UpdateConfig`
- **请求消息**：`UpdateSystemAIConfigRequest`
- **响应消息**：`UpdateSystemAIConfigResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `provider_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `name` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `base_url` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `organization` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `models` | `repeated string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `default_model` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `temperature` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `timeout_seconds` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `max_tokens` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `purposes` | `repeated string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `primary_for` | `repeated string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `enabled` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `provider_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "provider_id": "<string>",
  "name": "<string>",
  "base_url": "<string>",
  "organization": "<string>",
  "models": [
    "<string>"
  ],
  "default_model": "<string>",
  "temperature": 0.0,
  "timeout_seconds": 0,
  "max_tokens": 0,
  "purposes": [
    "<string>"
  ],
  "primary_for": [
    "<string>"
  ],
  "enabled": false
}
```

#### 响应示例

```json
{
  "provider_id": "<string>"
}
```

### SystemAIService/UpdateSecret

- **URL**：`POST https://api.alfq.org/antclaw.v1.SystemAIService/UpdateSecret`
- **请求消息**：`UpdateSystemAISecretRequest`
- **响应消息**：`UpdateSystemAISecretResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `provider_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `secret` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `provider_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `secret_updated` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "provider_id": "<string>",
  "secret": "<string>"
}
```

#### 响应示例

```json
{
  "provider_id": "<string>",
  "secret_updated": false
}
```

### SystemAIService/DiscoverModels

- **URL**：`POST https://api.alfq.org/antclaw.v1.SystemAIService/DiscoverModels`
- **请求消息**：`DiscoverSystemAIModelsRequest`
- **响应消息**：`DiscoverSystemAIModelsResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `provider_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `provider_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `models` | `repeated string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `default_model` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "provider_id": "<string>"
}
```

#### 响应示例

```json
{
  "provider_id": "<string>",
  "models": [
    "<string>"
  ],
  "default_model": "<string>"
}
```

### SystemAIService/ValidateConnection

- **URL**：`POST https://api.alfq.org/antclaw.v1.SystemAIService/ValidateConnection`
- **请求消息**：`ValidateSystemAIConnectionRequest`
- **响应消息**：`ValidateSystemAIConnectionResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `provider_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `provider_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `ok` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `model_count` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "provider_id": "<string>"
}
```

#### 响应示例

```json
{
  "provider_id": "<string>",
  "ok": false,
  "model_count": 0
}
```

## TAService

### TAService/GetIndicators

- **URL**：`POST https://api.alfq.org/antclaw.v1.TAService/GetIndicators`
- **请求消息**：`GetIndicatorsRequest`
- **响应消息**：`GetIndicatorsResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `timeframe` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `indicators` | `repeated string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `timeframe` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `values` | `repeated IndicatorValue` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "pair": "<string>",
  "timeframe": "<string>",
  "indicators": [
    "<string>"
  ]
}
```

#### 响应示例

```json
{
  "pair": "<string>",
  "timeframe": "<string>",
  "values": [
    {
      "name": "<string>",
      "value": 0.0,
      "signal": "<string>"
    }
  ]
}
```

### TAService/GetElliott

- **URL**：`POST https://api.alfq.org/antclaw.v1.TAService/GetElliott`
- **请求消息**：`GetElliottRequest`
- **响应消息**：`GetElliottResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `timeframe` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `current_count` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `waves` | `repeated ElliottWave` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `next_projection` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "pair": "<string>",
  "timeframe": "<string>"
}
```

#### 响应示例

```json
{
  "pair": "<string>",
  "current_count": "<string>",
  "waves": [
    {
      "wave_number": 0,
      "direction": "<string>",
      "price_start": 0.0,
      "price_end": 0.0
    }
  ],
  "next_projection": "<string>"
}
```

### TAService/GetWyckoff

- **URL**：`POST https://api.alfq.org/antclaw.v1.TAService/GetWyckoff`
- **请求消息**：`GetWyckoffRequest`
- **响应消息**：`GetWyckoffResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `phase` | `WyckoffPhase` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `structure` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "pair": "<string>"
}
```

#### 响应示例

```json
{
  "pair": "<string>",
  "phase": {
    "phase": "<string>",
    "start_price": 0.0,
    "current_price": 0.0,
    "evidence": "<string>"
  },
  "structure": "<string>"
}
```

### TAService/GetIct

- **URL**：`POST https://api.alfq.org/antclaw.v1.TAService/GetIct`
- **请求消息**：`GetIctRequest`
- **响应消息**：`GetIctResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `timeframe` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `levels` | `repeated IctLevel` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `bias` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "pair": "<string>",
  "timeframe": "<string>"
}
```

#### 响应示例

```json
{
  "pair": "<string>",
  "levels": [
    {
      "type": "<string>",
      "price": 0.0,
      "timeframe": "<string>"
    }
  ],
  "bias": "<string>"
}
```

### TAService/GetAmt

- **URL**：`POST https://api.alfq.org/antclaw.v1.TAService/GetAmt`
- **请求消息**：`GetAmtRequest`
- **响应消息**：`GetAmtResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `lookback_days` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `zones` | `repeated AmtZone` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `auction_context` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "pair": "<string>",
  "lookback_days": 0
}
```

#### 响应示例

```json
{
  "pair": "<string>",
  "zones": [
    {
      "type": "<string>",
      "price": 0.0,
      "volume": 0.0
    }
  ],
  "auction_context": "<string>"
}
```

### TAService/GetOrderflow

- **URL**：`POST https://api.alfq.org/antclaw.v1.TAService/GetOrderflow`
- **请求消息**：`GetOrderflowRequest`
- **响应消息**：`GetOrderflowResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `timeframe` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `imbalances` | `repeated Imbalance` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `cumulative_delta` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "pair": "<string>",
  "timeframe": "<string>"
}
```

#### 响应示例

```json
{
  "pair": "<string>",
  "imbalances": [
    {
      "type": "<string>",
      "price": 0.0,
      "delta": 0.0
    }
  ],
  "cumulative_delta": 0.0
}
```

### TAService/GetVolumeProfile

- **URL**：`POST https://api.alfq.org/antclaw.v1.TAService/GetVolumeProfile`
- **请求消息**：`GetVolumeProfileRequest`
- **响应消息**：`GetVolumeProfileResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `timeframe` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `num_bins` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `profile` | `repeated VpLevel` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `poc` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | point of control |
| `value_area_high` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `value_area_low` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "pair": "<string>",
  "timeframe": "<string>",
  "num_bins": 0
}
```

#### 响应示例

```json
{
  "pair": "<string>",
  "profile": [
    {
      "price": 0.0,
      "volume": 0.0
    }
  ],
  "poc": 0.0,
  "value_area_high": 0.0,
  "value_area_low": 0.0
}
```

### TAService/GetIntermarket

- **URL**：`POST https://api.alfq.org/antclaw.v1.TAService/GetIntermarket`
- **请求消息**：`GetIntermarketRequest`
- **响应消息**：`GetIntermarketResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `correlations` | `repeated Correlation` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `dominant_driver` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "pair": "<string>"
}
```

#### 响应示例

```json
{
  "pair": "<string>",
  "correlations": [
    {
      "asset_a": "<string>",
      "asset_b": "<string>",
      "correlation": 0.0,
      "timeframe": "<string>"
    }
  ],
  "dominant_driver": "<string>"
}
```

## TreasuryService

### TreasuryService/GetCurve

- **URL**：`POST https://api.alfq.org/antclaw.v1.TreasuryService/GetCurve`
- **请求消息**：`GetCurveRequest`
- **响应消息**：`GetCurveResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `date` | `google.protobuf.Timestamp` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `points` | `repeated YieldPoint` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{}
```

#### 响应示例

```json
{
  "date": {},
  "points": [
    {}
  ]
}
```

### TreasuryService/GetAnalysis

- **URL**：`POST https://api.alfq.org/antclaw.v1.TreasuryService/GetAnalysis`
- **请求消息**：`TreasuryServiceGetAnalysisRequest`
- **响应消息**：`TreasuryServiceGetAnalysisResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `curve_2s10s` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `curve_3m10y` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `regime` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{}
```

#### 响应示例

```json
{
  "curve_2s10s": 0.0,
  "curve_3m10y": 0.0,
  "regime": "<string>"
}
```

## UserService

### UserService/GetMe

- **URL**：`POST https://api.alfq.org/antclaw.v1.UserService/GetMe`
- **请求消息**：`GetMeRequest`
- **响应消息**：`GetMeResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `user` | `User` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{}
```

#### 响应示例

```json
{
  "user": {
    "user_id": "<string>",
    "email": "<string>",
    "username": "<string>",
    "display_name": "<string>",
    "locale": "LOCALE_UNSPECIFIED",
    "timezone": "<string>",
    "roles": [
      "<string>"
    ],
    "email_verified": false,
    "created_at": 0,
    "updated_at": 0,
    "code_id": "<string>"
  }
}
```

### UserService/UpdateSettings

- **URL**：`POST https://api.alfq.org/antclaw.v1.UserService/UpdateSettings`
- **请求消息**：`UpdateSettingsRequest`
- **响应消息**：`UpdateSettingsResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `display_name` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `locale` | `Locale` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `timezone` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `user` | `User` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "display_name": "<string>",
  "locale": "LOCALE_UNSPECIFIED",
  "timezone": "<string>"
}
```

#### 响应示例

```json
{
  "user": {
    "user_id": "<string>",
    "email": "<string>",
    "username": "<string>",
    "display_name": "<string>",
    "locale": "LOCALE_UNSPECIFIED",
    "timezone": "<string>",
    "roles": [
      "<string>"
    ],
    "email_verified": false,
    "created_at": 0,
    "updated_at": 0,
    "code_id": "<string>"
  }
}
```

### UserService/GetMembership

- **URL**：`POST https://api.alfq.org/antclaw.v1.UserService/GetMembership`
- **请求消息**：`GetMembershipRequest`
- **响应消息**：`GetMembershipResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `membership` | `Membership` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{}
```

#### 响应示例

```json
{
  "membership": {
    "tier": "MEMBERSHIP_TIER_UNSPECIFIED",
    "expires_at": 0,
    "quota_daily": 0,
    "quota_used_today": 0
  }
}
```

### UserService/StartOnboarding

- **URL**：`POST https://api.alfq.org/antclaw.v1.UserService/StartOnboarding`
- **请求消息**：`StartOnboardingRequest`
- **响应消息**：`StartOnboardingResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `onboarding_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `steps` | `repeated string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{}
```

#### 响应示例

```json
{
  "onboarding_id": "<string>",
  "steps": [
    "<string>"
  ]
}
```

### UserService/GetHistory

- **URL**：`POST https://api.alfq.org/antclaw.v1.UserService/GetHistory`
- **请求消息**：`GetHistoryRequest`
- **响应消息**：`GetHistoryResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `cursor` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `page_size` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `items` | `repeated HistoryItem` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `next_cursor` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "cursor": "<string>",
  "page_size": 0
}
```

#### 响应示例

```json
{
  "items": [
    {
      "item_id": "<string>",
      "type": "<string>",
      "title": "<string>",
      "payload": "<string>",
      "created_at": 0
    }
  ],
  "next_cursor": "<string>"
}
```

### UserService/ClearHistory

- **URL**：`POST https://api.alfq.org/antclaw.v1.UserService/ClearHistory`
- **请求消息**：`ClearHistoryRequest`
- **响应消息**：`ClearHistoryResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `all` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `types` | `repeated string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `cleared_count` | `int32` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "all": false,
  "types": [
    "<string>"
  ]
}
```

#### 响应示例

```json
{
  "cleared_count": 0
}
```

### UserService/ListPins

- **URL**：`POST https://api.alfq.org/antclaw.v1.UserService/ListPins`
- **请求消息**：`ListPinsRequest`
- **响应消息**：`ListPinsResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pins` | `repeated Pin` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{}
```

#### 响应示例

```json
{
  "pins": [
    {
      "pin_id": "<string>",
      "item_id": "<string>",
      "item_type": "<string>",
      "title": "<string>",
      "created_at": 0
    }
  ]
}
```

### UserService/Pin

- **URL**：`POST https://api.alfq.org/antclaw.v1.UserService/Pin`
- **请求消息**：`PinRequest`
- **响应消息**：`PinResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `item_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `item_type` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `title` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pin` | `Pin` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "item_id": "<string>",
  "item_type": "<string>",
  "title": "<string>"
}
```

#### 响应示例

```json
{
  "pin": {
    "pin_id": "<string>",
    "item_id": "<string>",
    "item_type": "<string>",
    "title": "<string>",
    "created_at": 0
  }
}
```

### UserService/Unpin

- **URL**：`POST https://api.alfq.org/antclaw.v1.UserService/Unpin`
- **请求消息**：`UnpinRequest`
- **响应消息**：`UnpinResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pin_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 请求示例

```json
{
  "pin_id": "<string>"
}
```

#### 响应示例

```json
{}
```

### UserService/SubmitFeedback

- **URL**：`POST https://api.alfq.org/antclaw.v1.UserService/SubmitFeedback`
- **请求消息**：`SubmitFeedbackRequest`
- **响应消息**：`SubmitFeedbackResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `category` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `content` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `contact` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `feedback_id` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "category": "<string>",
  "content": "<string>",
  "contact": "<string>"
}
```

#### 响应示例

```json
{
  "feedback_id": "<string>"
}
```

### UserService/SetAiKey

- **URL**：`POST https://api.alfq.org/antclaw.v1.UserService/SetAiKey`
- **请求消息**：`SetAiKeyRequest`
- **响应消息**：`SetAiKeyResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `provider` | `AiProvider` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `api_key` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `success` | `bool` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "provider": "AI_PROVIDER_UNSPECIFIED",
  "api_key": "<string>"
}
```

#### 响应示例

```json
{
  "success": false
}
```

## VolService

### VolService/GetVix

- **URL**：`POST https://api.alfq.org/antclaw.v1.VolService/GetVix`
- **请求消息**：`GetVixRequest`
- **响应消息**：`GetVixResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `vix` | `VixData` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `history` | `repeated VixData` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{}
```

#### 响应示例

```json
{
  "vix": {
    "timestamp": "<string>",
    "spot": 0.0,
    "term_structure": 0.0,
    "percentile_30d": 0.0,
    "regime": "<string>"
  },
  "history": [
    {
      "timestamp": "<string>",
      "spot": 0.0,
      "term_structure": 0.0,
      "percentile_30d": 0.0,
      "regime": "<string>"
    }
  ]
}
```

### VolService/GetMove

- **URL**：`POST https://api.alfq.org/antclaw.v1.VolService/GetMove`
- **请求消息**：`GetMoveRequest`
- **响应消息**：`GetMoveResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `move` | `MoveData` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{}
```

#### 响应示例

```json
{
  "move": {
    "timestamp": "<string>",
    "value": 0.0,
    "trend": "<string>"
  }
}
```

### VolService/GetDvol

- **URL**：`POST https://api.alfq.org/antclaw.v1.VolService/GetDvol`
- **请求消息**：`GetDvolRequest`
- **响应消息**：`GetDvolResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `asset` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | BTC, ETH, etc. |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `dvol` | `DvolData` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "asset": "<string>"
}
```

#### 响应示例

```json
{
  "dvol": {
    "timestamp": "<string>",
    "value": 0.0,
    "asset": "<string>"
  }
}
```

### VolService/GetGex

- **URL**：`POST https://api.alfq.org/antclaw.v1.VolService/GetGex`
- **请求消息**：`GetGexRequest`
- **响应消息**：`GetGexResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `gex` | `GexData` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "pair": "<string>"
}
```

#### 响应示例

```json
{
  "gex": {
    "timestamp": "<string>",
    "pair": "<string>",
    "net_gex": 0.0,
    "flip_point": 0.0,
    "wall": "<string>"
  }
}
```

### VolService/GetIvol

- **URL**：`POST https://api.alfq.org/antclaw.v1.VolService/GetIvol`
- **请求消息**：`GetIvolRequest`
- **响应消息**：`GetIvolResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `expiry` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `surface` | `repeated IvolPoint` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |
| `atm_iv` | `double` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "pair": "<string>",
  "expiry": "<string>"
}
```

#### 响应示例

```json
{
  "pair": "<string>",
  "surface": [
    {
      "strike": 0.0,
      "iv": 0.0,
      "delta": 0.0,
      "gamma": 0.0,
      "theta": 0.0,
      "vega": 0.0
    }
  ],
  "atm_iv": 0.0
}
```

### VolService/GetSkew

- **URL**：`POST https://api.alfq.org/antclaw.v1.VolService/GetSkew`
- **请求消息**：`GetSkewRequest`
- **响应消息**：`GetSkewResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `skew` | `SkewData` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "pair": "<string>"
}
```

#### 响应示例

```json
{
  "skew": {
    "timestamp": "<string>",
    "pair": "<string>",
    "risk_reversal": 0.0,
    "fly": 0.0,
    "term_structure": "<string>"
  }
}
```

### VolService/GetSkewVixAlert

- **URL**：`POST https://api.alfq.org/antclaw.v1.VolService/GetSkewVixAlert`
- **请求消息**：`GetSkewVixAlertRequest`
- **响应消息**：`GetSkewVixAlertResponse`
- **认证**：业务接口默认需要 `Authorization: Bearer <access_token>`；登录/健康检查等公共接口除外。

#### 请求字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `pair` | `string` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 响应字段

| 字段 | 类型 | 是否必填 | 约束/说明 |
|---|---|---|---|
| `alerts` | `repeated SkewVixAlert` | 否；proto3 默认值语义，业务必填以服务端校验为准 | 详见 proto 定义与业务校验 |

#### 请求示例

```json
{
  "pair": "<string>"
}
```

#### 响应示例

```json
{
  "alerts": [
    {
      "alert_id": "<string>",
      "timestamp": "<string>",
      "pair": "<string>",
      "signal": "<string>",
      "confidence": 0.0,
      "reason": "<string>"
    }
  ]
}
```

