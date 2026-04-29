#!/usr/bin/env bash
# SC-4 Walk-Forward 回测（windows≥5）
set -uo pipefail
source "$(dirname "$0")/_lib.sh"
# strategy=sma_crossover 是实际调试路径，symbols 使用 worker 采集过的 EURUSD。
body='{"strategy":"sma_crossover","symbols":["EURUSD"],"fromDate":"2025-04-28","toDate":"2026-04-29","folds":4,"trainRatio":0.7}'
out=$(call BacktestService RunWalkforward "$body")
if is_error "$out"; then
  echo "  RPC 错误：$(echo "$out" | jq -r '.message')"
  exit 1
fi
jid=$(echo "$out" | jq -r '.jobId')
status=$(echo "$out" | jq -r '.status')
echo "  jobId=$jid status=$status"
[[ "$status" == "done" ]] || exit 1
out2=$(call BacktestService GetWalkforwardResult "{\"jobId\":\"$jid\"}")
n=$(echo "$out2" | jq '.folds | length')
echo "  folds=$n"
ge "$n" 4
