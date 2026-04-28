# 02. endpoint 字段落地到 collector

## 目标

让 `data_source_configs.endpoint` **真正生效**：管理员改完 endpoint，对应 collector 下次调用就用新地址。

## 背景（当前状态）

- endpoint 字段已能存储 / 通过管理端修改；
- 但各 collector（FRED / CoinGecko / DefiLlama / MQL5 / …）内部 **URL 硬编码**，或从 ENV 取；
- 表 `data_source_configs` 的 endpoint 当前只是"展示用"。

## 设计决策

| 决策点 | 选择 |
| --- | --- |
| 读取方式 | 通过 01 号文档引入的 `CredentialResolver.GetEndpoint(sourceID)` |
| 构造 client | 每次任务 tick 前 `client.SetBaseURL(resolver.GetEndpoint(...))` |
| 空值回退 | `GetEndpoint` 返回空串时 → 用 client 内置默认 URL（即现在硬编码的值） |
| 兼容 ENV | 保留旧的 `ANTCLAW_FRED_API_URL` 等环境变量作为第三级回退 |
| 只改"有意义改动"的 client | Yahoo/Stooq/DefiLlama 等公开 API 本就少变更，但统一接入便于运维切换备用镜像 |

## 数据模型

无新表。`data_source_configs.endpoint` 已有。

## 接口契约

### 每个 client 统一增加两个方法

```go
// internal/infra/apiclient/<provider>.go
func (c *<Provider>Client) SetBaseURL(url string) { /* 空串不动 */ }
func (c *<Provider>Client) SetKey(key string)     { /* 空串不动 */ }  // 无 key 的 client 省略
```

**空值保护**：`SetBaseURL("")` 应**保留当前值**，避免 resolver 返回空把客户端打坏。

### Resolver 回调注册（扩展 01 号文档）

```go
// 让 client 订阅某个 sourceID 的变化，resolver 在 Reload 时回调
resolver.OnChange("fred", func(secret, endpoint string) {
    fredClient.SetKey(secret)
    fredClient.SetBaseURL(endpoint)
})
```

`OnChange(sourceID, cb)` 内部将 cb 追加到 `callbacks[sourceID]`。启动预热和 pub/sub 触发都会调用。

## 14 个数据源接入清单

| source_id | client 文件 | 默认 URL（现硬编码） | 有 key? | 优先级 |
| --- | --- | --- | --- | --- |
| fred | `apiclient/fred.go` | `https://api.stlouisfed.org/fred` | ✅ | 高 |
| cftc_socrata | `apiclient/cftc.go`（如存在） | `https://publicreporting.cftc.gov` | 可选 | 高 |
| mql5 | `apiclient/mql5.go` | `https://www.mql5.com/...` | ❌ | 高 |
| yahoo | `apiclient/yahoo.go` | `https://query1.finance.yahoo.com` | ❌ | 中 |
| stooq | `apiclient/stooq.go` | `https://stooq.com` | ❌ | 中 |
| coingecko | `apiclient/coingecko.go` | `https://api.coingecko.com/api/v3` | 可选 Pro | 中 |
| defillama | `apiclient/defillama.go` | `https://api.llama.fi` | ❌ | 中 |
| alternative_me | `apiclient/alternativeme.go` | `https://api.alternative.me` | ❌ | 低 |
| deribit | `apiclient/deribit.go` | `https://www.deribit.com/api/v2` | ❌ | 低 |
| binance | - | `https://api.binance.com` | 可选 | 低 |
| bybit | - | - | 可选 | 低 |
| 其他 3 个 | 按实际 | - | - | 低 |

> 对不存在的 client 文件，不要为接入而临时新建 — 先确认 worker 里实际有用到，再改。

## 关键流程

### Worker 启动（完整形态）

```go
resolver := datasource.NewCredentialResolver(pool, box, envFallback, logger)
_ = resolver.ReloadAll(ctx)

// 构造所有 client，它们内部使用默认 URL
fredClient := apiclient.NewFredClient("")
mql5Fetcher := apiclient.NewMQL5Fetcher()
// ...

// 注册 reload 回调
resolver.OnChange("fred", func(secret, endpoint string) {
    fredClient.SetKey(secret)
    fredClient.SetBaseURL(endpoint)
})
resolver.OnChange("mql5", func(_, endpoint string) {
    mql5Fetcher.SetBaseURL(endpoint)
})
// 其余 sourceID 类推

// 触发一次回调（把 ReloadAll 已读到的值推入 client）
resolver.FireAll()

// 启动订阅
stop := resolver.StartSubscriber(ctx, rdb); defer stop()
```

### `FireAll`

```go
func (r *CredentialResolver) FireAll() {
    r.mu.RLock()
    defer r.mu.RUnlock()
    for id, cbs := range r.callbacks {
        secret := r.secrets[id]
        endpoint := r.endpoints[id]
        for _, cb := range cbs {
            cb(secret, endpoint)
        }
    }
}
```

### SetBaseURL 细节

**必须处理尾斜杠**：管理员可能填 `https://api.stlouisfed.org/fred/` 也可能 `.../fred`。client 内部拼路径前 `strings.TrimRight(url, "/")`。

## 修改清单

| 文件 | 变更 |
| --- | --- |
| `backend/internal/service/datasource/resolver.go` | 增加 `OnChange`、`FireAll`、回调存储 |
| 各 `apiclient/*.go` | 统一增加 `SetBaseURL` / `SetKey`；将硬编码 URL 改为 `baseURL` 字段 |
| `backend/cmd/antclaw-worker/main.go` | 替换硬编码构造逻辑为"构造 + 注册 + FireAll + StartSubscriber" |
| `backend/internal/adapter/storage/postgres/ensure_schema.go` | 数据源种子的 endpoint 兜底值是否准确（已经 OK）|
| `docs/AntClaw-数据源配置加密方案.md` | §7 注明 endpoint 现在也生效 |

## 验证步骤

```bash
# 1. 把 FRED endpoint 改成错误域名
ENDPOINT=https://totally-wrong.example.com SECRET=xxx python3 /tmp/test_secure_put.py
# （需要支持 ENDPOINT env 的脚本改造，或手写 curl）

# 2. 不重启 worker，等 macro tick 或手动触发 → 应失败，日志里 base URL 是新的
docker logs --since 10s antclaw-worker 2>&1 | grep -iE "fred|endpoint"

# 3. 改回正确 endpoint → 下一 tick 恢复 200
```

## 注意事项

- **不要给 collector 增加全局状态**：`SetBaseURL` 必须加 mutex；`Do(req)` 读 baseURL 时也要加读锁
- **尾斜杠**：见"SetBaseURL 细节"
- **错误隔离**：一个 collector 挂了不能拖垮其他；现有 `runWithEvent` 已有 recover 机制
- **Worker 启动顺序**：ReloadAll → 构造 client（它们用默认 URL） → 注册 OnChange → FireAll（此时才把 DB 值推进去）。乱序会造成"client 已开始工作但仍用默认 URL"

## 风险与回退

- 风险：运维错填 endpoint 导致采集失败 → 管理端一键恢复（置空 endpoint 回退默认）
- 回退：所有 `SetBaseURL("")` 都保持原值；endpoint 字段删空 = 保持默认
