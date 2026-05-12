# AntClaw 全面收口 · 完成报告

- 完成时间：2026-05-12
- 执行范围：合成数据清除 + 旧代码清理 + 前端补齐 + 质量门禁

## 一、阶段1：合成数据清除

| 文件 | 变更 | 验证 |
|---|---|---|
| `engine/signals/provider.go` | 删除 `randFloat()`、`getSyntheticData()`、`generateSyntheticBars()`；`GetMarketData` 失败直接返回 error，不再吞错返回假数据 | ✅ `go vet` 通过 |
| `service/signals/unified_signal_service.go` | TA → 读 `price_daily` 计算 SMA(5)/SMA(20) 交叉；Sentiment → 读 `sentiment_snapshots` 最新 `fear_greed`；Macro → 读 `macro_regime_history`；`GetSignalHistory` → 读 `signals_history` 表 | ✅ `go vet` 通过；无 unimplemented 残留 |
| `service/backtest/service.go` | 删除 `simulateBacktest`（硬编码 3 条假 trade）；`RunBacktest` 改为调用真实 `runEngine`（CTA Donchian 引擎） | ✅ `go vet` 通过；无 simulateBacktest 残留 |

## 二、阶段2：旧代码清理

| 删除项 | 大小 | 状态 |
|---|---|---|
| `Emulator/` | 59MB / 525 文件 | ✅ 已删除 |
| `bak0428/` | 30MB | ✅ 已删除 |
| `backend/antclaw-worker` 二进制 | — | ✅ 已删除 |
| `internal/adapter/bot/` | Bot 骨架 | ✅ 已删除 |
| `internal/adapter/sandbox/` | 沙箱骨架 | ✅ 已删除 |
| `internal/service/bot/` | Bot 路由骨架 | ✅ 已删除 |

仓库体积：131MB → 40MB（-69%），Go 文件：274 → 267。

## 三、阶段3：前端补齐

| 变更 | 文件 |
|---|---|
| Web 端基础设施 | 新增 `features/_shared/transport.ts`（Connect-RPC 客户端 + JWT 注入） |
| Web 端信号页 | 新 `features/signals/SignalsPage.tsx`（真实 RPC 调用 `SignalsService.getBias`，5 对） |
| Web 端路由 | `App.tsx` 导入新 SignalsPage 替换旧的硬编码 demo 页 |
| Web 端依赖 | `package.json` 新增 `@antclaw/proto`、`@bufbuild/protobuf`、升级 `@connectrpc/connect` 到 v2 |

## 四、阶段4：质量门禁

- `go vet ./...` → ✅ 全通过
- `go test ./internal/...` → ✅ 5/5 包通过（auth、apiclient、backtest、calibration、quant）
- `lint-filesize.sh` → ✅ 所有文件 ≤ 800 行
- `lint-protocol.sh` → ✅ 后端协议合规

## 五、后续建议

1. **`backend/cmd/antclaw-api/main.go`** 需更新 `NewUnifiedSignalService` 调用点，传入新的 `MacroRepository` 和 `*pgxpool.Pool` 参数（当前该构造未被 main.go 调用，无编译影响）
2. **前端 `pnpm install`**：Web 端新增了 workspace 依赖，需在部署前执行
3. **`signals_history` 表**：`persistSignal` 写入 `signals_history` 表，需确保 migration 包含此表（当前为幂等 INSERT）
4. **`MT4Handler` / `MT5Handler`**：P6c 占位保留（所有方法返回 UNIMPLEMENTED），属于按规范的后期交付项

## 六、验收结论

- [x] 0 处合成数据 / 硬编码残留
- [x] 旧代码 ~89MB 已清除
- [x] Web 端从硬编码 demo 切换到真实 RPC
- [x] 编译 / vet / lint / 测试全绿
- [x] 所有信号组件（TA/Sentiment/Macro）接入真实数据源

**收口完成。**
