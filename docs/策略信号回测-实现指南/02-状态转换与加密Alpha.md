# 02 · 状态转换与加密 Alpha（transition / cryptoalpha）

> 替换 `signals_handler.go` 中两个 placeholder（返回空数组）。基于 `regime_transitions` 与链上数据计算真实信号。

---

## 1. GetTransition — 宏观状态转换概率

### 1.1 目标
对给定 `pair, current_state` 返回未来 N 天迁移到其他状态的概率分布（基于历史一阶马尔可夫频率）。

### 1.2 算法
```
1. history = RegimeProvider.GetHistory(pair, "1d", lookback=730)   # 2 年
   if len(history) < 60: DATA_INSUFFICIENT

2. 构建状态序列 labels = [h.UnifiedLabel for h in history]
   状态域 = {STRONG_BULL, BULL, NEUTRAL, BEAR, STRONG_BEAR}

3. 统计转移矩阵 M[i][j] = count(labels[t]==i AND labels[t+1]==j)
   行归一化：P[i][j] = M[i][j] / sum_j(M[i][j])

4. current = req.CurrentState (若空则取 history[0].UnifiedLabel)
5. 对每个 to_state j：
     transitions.append({
        from_state: current,
        to_state: j,
        probability: P[current][j],
     })
   按概率降序输出
```

### 1.3 持久化
为加速，每日 worker 任务写一次到新表 `regime_transition_matrix`：

```sql
CREATE TABLE IF NOT EXISTS regime_transition_matrix (
    asof_date DATE NOT NULL,
    symbol    VARCHAR(32) NOT NULL,
    timeframe VARCHAR(8)  NOT NULL,
    from_label VARCHAR(16) NOT NULL,
    to_label   VARCHAR(16) NOT NULL,
    probability DOUBLE PRECISION NOT NULL,
    sample_size INT NOT NULL,
    PRIMARY KEY (asof_date, symbol, timeframe, from_label, to_label)
);
```

Worker 任务 `transition_matrix.go`：每日凌晨 02:30 重算所有主要标的（`Categories.majors + crypto`）的 5×5 矩阵。

### 1.4 修改清单
| 文件 | 动作 |
|------|------|
| `backend/internal/adapter/storage/postgres/ensure_schema.go` | 新增 `regime_transition_matrix` |
| `backend/cmd/antclaw-worker/transition_matrix.go` | 新建任务 |
| `backend/internal/service/signals/service.go` | 实现 `GetTransition` |
| `backend/internal/service/signals/transition.go` | 矩阵构建纯函数 |
| `backend/internal/service/signals/transition_test.go` | 单测 |

### 1.5 验证
```bash
# 触发一次矩阵计算（手动）
docker compose exec -T worker /app/antclaw-worker --once=transition_matrix

# 查询持久化
docker compose exec -T postgres psql -U antclaw -d antclaw -c \
  "SELECT from_label, to_label, probability, sample_size
   FROM regime_transition_matrix
   WHERE symbol='EURUSD' AND asof_date=CURRENT_DATE
   ORDER BY from_label, probability DESC;"

# 调用 RPC
curl -s http://localhost:8082/antclaw.v1.SignalsService/GetTransition \
  -d '{"pair":"EURUSD","current_state":"NEUTRAL"}' \
  -H 'Content-Type: application/json' | jq .
```

预期：返回 5 条 transitions，概率和 ≈ 1.0。

---

## 2. GetCryptoAlpha — 加密资产专用宏观信号

### 2.1 目标
对加密资产（BTC/ETH/SOL/...）输出 4 类信号：`accumulation`, `distribution`, `momentum_breakout`, `mean_reversion`。

### 2.2 数据依赖
| 数据 | 来源表 | 用途 |
|------|--------|------|
| 价格 | `price_daily` | 动量、均值回归 |
| 链上 | `onchain_metrics` (新建) | 累积/分发判断 |
| 资金流 | `flow_divergence_history` | BTC↔ETH 背离 |
| 情绪 | `crypto_sentiment` (已有 `internal/service/sentiment`) | 恐惧贪婪 |
| 期权 | `crypto_iv_surface` (已有 `internal/service/vol/`) | 偏度告警 |

新建表 `onchain_metrics`（与 `internal/service/onchain` 配合，已有 `OnChainClient` 拉取）：
```sql
CREATE TABLE IF NOT EXISTS onchain_metrics (
    time   TIMESTAMPTZ NOT NULL,
    asset  VARCHAR(16)  NOT NULL,
    metric VARCHAR(32)  NOT NULL,   -- 'exchange_netflow','active_addresses','mvrv','sopr','funding_rate'
    value  DOUBLE PRECISION NOT NULL,
    source VARCHAR(32),
    PRIMARY KEY (time, asset, metric)
);
SELECT create_hypertable('onchain_metrics','time', chunk_time_interval=>INTERVAL '30 days');
```

### 2.3 算法

#### 2.3.1 累积/分发（accumulation / distribution）
```
exchange_netflow_30d = sum(value WHERE metric='exchange_netflow' AND time > now-30d)
mvrv = latest('mvrv')
funding = avg(value WHERE metric='funding_rate' AND time > now-7d)

if exchange_netflow_30d < -threshold_acc AND mvrv < 1.0:
    signal_type = "accumulation"
    confidence = sigmoid(-exchange_netflow_30d / threshold_acc)
elif exchange_netflow_30d > threshold_dist AND mvrv > 2.5:
    signal_type = "distribution"
    confidence = sigmoid(exchange_netflow_30d / threshold_dist)
```

阈值（可配置）：
- `threshold_acc = -50000` (BTC, 单位 BTC)
- `threshold_dist = 50000`

#### 2.3.2 动量突破（momentum_breakout）
```
bars = PriceProvider.GetDailyBars(asset, now-90d, now)
ret_30d = (bars[-1].close - bars[-30].close) / bars[-30].close
ret_90d = (bars[-1].close - bars[0].close) / bars[0].close
vol_30d = stddev([daily_return(b)] for last 30 bars)

if ret_30d > 2 * vol_30d * sqrt(30) AND ret_30d > 0.15:
    signal_type = "momentum_breakout"
    confidence = min(0.95, ret_30d / (3*vol_30d*sqrt(30)))
```

#### 2.3.3 均值回归（mean_reversion）
```
sma_50 = mean(bars[-50:].close)
deviation = (close - sma_50) / sma_50
funding_extreme = |funding| > 0.05% per 8h

if abs(deviation) > 0.20 AND funding_extreme:
    signal_type = "mean_reversion"
    direction = -sign(deviation)
    confidence = min(0.9, abs(deviation) / 0.4)
```

### 2.4 实现位置
| 文件 | 动作 |
|------|------|
| `backend/internal/service/cryptoalpha/service.go` | 新建独立 service |
| `backend/internal/service/cryptoalpha/types.go` | 新建 |
| `backend/internal/service/cryptoalpha/calc.go` | 三类算法纯函数 |
| `backend/internal/service/cryptoalpha/service_test.go` | 单测 |
| `backend/internal/service/signals/service.go` | `GetCryptoAlpha` 委托给 cryptoalpha service |
| `backend/cmd/antclaw-worker/onchain_collect.go` | 周期拉取链上指标写入 `onchain_metrics` |
| `backend/internal/adapter/storage/postgres/ensure_schema.go` | 新增 `onchain_metrics` 表 |

### 2.5 RPC 输出
```go
[]*signalsv1.CryptoAlphaSignal{
    {
        Asset: "BTC",
        SignalType: "accumulation",   // 见上 4 类
        Confidence: 0.82,
        Timeframe:  "1d",
    },
    ...
}
```

`asset_filter` 请求参数为空时返回所有 crypto Categories；非空则过滤。

### 2.6 边界
- `onchain_metrics` 数据缺失 → 该资产仅返回基于价格的信号（momentum/mean_reversion）
- 全部缺失 → DATA_INSUFFICIENT

### 2.7 验证
```bash
# 链上数据 worker 至少跑过一次
docker compose exec -T postgres psql -U antclaw -d antclaw -c \
  "SELECT asset, metric, count(*) FROM onchain_metrics
   WHERE time > now() - INTERVAL '7 days'
   GROUP BY asset, metric ORDER BY 1,2;"

# RPC
curl -s http://localhost:8082/antclaw.v1.SignalsService/GetCryptoAlpha \
  -d '{"asset_filter":""}' -H 'Content-Type: application/json' | jq .

# 期望返回 BTC/ETH/SOL 等至少 3 个 signal，confidence ∈ [0,1]
```

---

## 3. 完成判定

- [ ] `regime_transition_matrix` 表创建并由 worker 写入
- [ ] `onchain_metrics` 表创建并由 worker 写入（至少 BTC, ETH 的 5 个指标）
- [ ] `GetTransition` 返回真实概率
- [ ] `GetCryptoAlpha` 返回真实信号
- [ ] 单测 ≥ 80% 行覆盖

## 4. 实施记录

<!-- 完成后追加 -->
