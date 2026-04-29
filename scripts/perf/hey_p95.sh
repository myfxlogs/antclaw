#!/usr/bin/env bash
# hey_p95.sh —— 关键 RPC P50/P95 压测；hey 不在 PATH 时自动用 docker 镜像跑。
# 用法：bash scripts/perf/hey_p95.sh [并发=10] [总请求=200]
set -uo pipefail

C="${1:-10}"
N="${2:-200}"
API="${API_BASE:-http://host.docker.internal:8082}"
HOST_API="${HOST_API_BASE:-http://localhost:8082}"

# Endpoints to probe（POST + 空 body 即可触发，业务校验由服务自行）
declare -a TARGETS=(
  "antclaw.v1.SystemService/Healthz {}"
  "antclaw.v1.PriceService/GetPrice {\"pair\":\"EURUSD\",\"timeframe\":\"1d\",\"count\":30}"
  "antclaw.v1.OptionsService/GetGEX {\"asset\":\"BTC\"}"
  "antclaw.v1.TreasuryService/GetCurve {}"
  "antclaw.v1.SentimentService/GetSentiment {\"asset\":\"BTC\"}"
  "antclaw.v1.VolService/GetVix {}"
  "antclaw.v1.BacktestService/RunQuantBt {\"config\":{\"pair\":\"EURUSD\",\"strategyName\":\"tsmom\"}}"
)

have_hey() { command -v hey >/dev/null 2>&1; }

run_hey() {
  local path="$1" body="$2"
  if have_hey; then
    hey -n "$N" -c "$C" -m POST -T 'application/json' -d "$body" "${HOST_API}/${path}" 2>/dev/null
  else
    docker run --rm --network=host -i ghcr.io/rakyll/hey:latest \
      hey -n "$N" -c "$C" -m POST -T 'application/json' -d "$body" "${API}/${path}" 2>/dev/null \
      || docker run --rm --add-host=host.docker.internal:host-gateway williamyeh/hey \
           hey -n "$N" -c "$C" -m POST -T 'application/json' -d "$body" "${API}/${path}" 2>/dev/null
  fi
}

extract() {
  awk '
    /Average:/    { avg = $2 }
    /50%% in /    { p50 = $3 }
    /95%% in /    { p95 = $3 }
    /Total:/      { tot = $2 }
    END { printf "avg=%s p50=%s p95=%s total=%s", avg, p50, p95, tot }
  '
}

echo "并发=$C 请求=$N target=${HOST_API}"
printf "%-55s %s\n" "RPC" "Latency (s)"
echo "------------------------------------------------------------------"
for line in "${TARGETS[@]}"; do
  path=$(echo "$line" | awk '{print $1}')
  body=$(echo "$line" | cut -d' ' -f2-)
  out=$(run_hey "$path" "$body")
  m=$(echo "$out" | extract)
  printf "%-55s %s\n" "$path" "$m"
done
