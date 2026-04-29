#!/usr/bin/env bash
# SC-21 M-B Bootstrap + Monte Carlo
set -uo pipefail
source "$(dirname "$0")/_lib.sh"

# 复用 sc-20 的 wf jobId（如不存在则现跑一次）
if [[ -f /tmp/sc20_job.txt ]]; then
  job=$(cat /tmp/sc20_job.txt)
else
  to=$(date +%Y-%m-%d)
  from=$(date -d '-300 days' +%Y-%m-%d 2>/dev/null || date -v-300d +%Y-%m-%d)
  job=$(call BacktestService RunWalkforward "{\"strategy\":\"sma_crossover\",\"symbols\":[\"EURUSD\"],\"fromDate\":\"$from\",\"toDate\":\"$to\",\"folds\":3,\"trainRatio\":0.7}" | jq -r '.jobId')
fi

# Bootstrap：CI 排序合理（p5 <= p50 <= p95）
out=$(call BacktestService RunBootstrap "{\"baseJobId\":\"$job\",\"iterations\":300,\"randomSeed\":7}")
p5=$(echo "$out" | jq -r '.sharpeP5 // 0')
p50=$(echo "$out" | jq -r '.sharpeP50 // 0')
p95=$(echo "$out" | jq -r '.sharpeP95 // 0')
echo "  Bootstrap Sharpe p5=$p5 p50=$p50 p95=$p95"
awk -v a="$p5" -v b="$p50" -v c="$p95" 'BEGIN{exit !(a<=b && b<=c)}' || { echo "Sharpe CI not ordered"; exit 1; }

# Monte Carlo：路径维度正确；GARCH 参数合理
mc=$(call BacktestService RunMonteCarlo '{"pair":"EURUSD","timeframe":"1d","paths":500,"horizonBars":20,"randomSeed":7,"lookback":300}')
paths=$(echo "$mc" | jq -r '.paths // 0')
horizon=$(echo "$mc" | jq -r '.horizonBars // 0')
qp=$(echo "$mc" | jq '.quantilePaths | length')
echo "  MC paths=$paths horizon=$horizon quantilePaths=$qp"
ge "$paths" 100 || exit 1
ge "$horizon" 5 || exit 1
ge "$qp" 3 || exit 1
echo "  SC-21 OK"
