#!/usr/bin/env bash
# SC-19 M-A 价格深度：GARCH 波动率 / Hurst / 相关矩阵 / 背离 / HMM 状态
# 全部基于 price_daily 真历史 K 线（FRED 注入），不允许伪造数据。
set -uo pipefail
source "$(dirname "$0")/_lib.sh"

PAIR="${SC19_PAIR:-EURUSD}"
TF="${SC19_TF:-1d}"

# 1) GARCH 条件波动率：参数有效；序列非空；最后一根 ≠ 0
out=$(call PriceService GetVolatility "{\"pair\":\"$PAIR\",\"timeframe\":\"$TF\",\"lookback\":500}")
if is_error "$out"; then
  echo "  FAIL GetVolatility error: $out" >&2; exit 1
fi
omega=$(echo "$out" | jq -r '.omega // 0')
alpha=$(echo "$out" | jq -r '.alpha // 0')
beta=$(echo "$out" | jq -r '.beta // 0')
n=$(echo "$out" | jq '.series | length')
last=$(echo "$out" | jq -r '.series[-1].conditionalVol // 0')
echo "  GARCH omega=$omega alpha=$alpha beta=$beta series=$n last=$last"
ge "$alpha" 0 || { echo "alpha < 0"; exit 1; }
ge "$beta" 0 || { echo "beta < 0"; exit 1; }
ge "$n" 50 || { echo "series too short ($n)"; exit 1; }
awk -v v="$last" 'BEGIN{exit !(v+0>0)}' || { echo "last vol == 0"; exit 1; }

# persistence < 1
awk -v a="$alpha" -v b="$beta" 'BEGIN{exit !(a+b<1.0)}' || { echo "persistence>=1"; exit 1; }

# 2) Hurst：H ∈ (0,1)，sample_size 合理
out=$(call PriceService GetHurst "{\"pair\":\"$PAIR\",\"timeframe\":\"$TF\",\"lookback\":500}")
if is_error "$out"; then echo "  FAIL GetHurst: $out" >&2; exit 1; fi
H=$(echo "$out" | jq -r '.hurst')
interp=$(echo "$out" | jq -r '.interpretation')
echo "  Hurst H=$H ($interp)"
awk -v v="$H" 'BEGIN{exit !(v+0>0 && v+0<1)}' || { echo "H out of range"; exit 1; }

# 3) GetCorrelations：默认资产；矩阵对角线全部 1.0
out=$(call PriceService GetCorrelations "{\"timeframe\":\"$TF\",\"window\":30}")
if is_error "$out"; then echo "  FAIL GetCorrelations: $out" >&2; exit 1; fi
assets=$(echo "$out" | jq -r '.assets | length')
echo "  Correlations assets=$assets"
ge "$assets" 2 || { echo "assets <2"; exit 1; }
diag_bad=$(echo "$out" | jq '[.matrix[] | select(.assetA == .assetB) | (.value - 1) | fabs] | max // 0')
awk -v v="$diag_bad" 'BEGIN{exit !(v+0<1e-9)}' || { echo "diagonal not 1.0 (max dev=$diag_bad)"; exit 1; }

# 4) GetDivergences：能跑通即可（无背离也是 OK）
out=$(call PriceService GetDivergences "{\"pair\":\"$PAIR\",\"timeframe\":\"$TF\",\"lookback\":200}")
if is_error "$out"; then echo "  FAIL GetDivergences: $out" >&2; exit 1; fi
echo "  Divergences events=$(echo "$out" | jq '.events // [] | length')"

# 5) HMM regime：engine=hmm；置信度 ≥ 0.5；engineUsed=hmm 或 adx_fallback
out=$(call PriceService GetRegime "{\"pair\":\"$PAIR\",\"timeframe\":\"$TF\",\"engine\":\"hmm\",\"nStates\":2}")
if is_error "$out"; then echo "  FAIL GetRegime hmm: $out" >&2; exit 1; fi
eng=$(echo "$out" | jq -r '.engineUsed // ""')
conf=$(echo "$out" | jq -r '.regime.confidence // 0')
regime=$(echo "$out" | jq -r '.regime.regime // ""')
echo "  Regime engine=$eng regime=$regime conf=$conf"
[[ "$eng" == "hmm" || "$eng" == "adx_fallback" ]] || { echo "unknown engine $eng"; exit 1; }
awk -v v="$conf" 'BEGIN{exit !(v+0>=0.5)}' || { echo "confidence<0.5"; exit 1; }
[[ -n "$regime" ]] || { echo "regime empty"; exit 1; }

echo "  SC-19 OK"
