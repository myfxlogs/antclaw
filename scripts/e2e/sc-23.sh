#!/usr/bin/env bash
# SC-23 M-D quantbt（TSMOM）
set -uo pipefail
source "$(dirname "$0")/_lib.sh"
out=$(call BacktestService RunQuantBt '{"config":{"pair":"EURUSD","strategyName":"tsmom"}}')
task=$(echo "$out" | jq -r '.taskId')
status=$(echo "$out" | jq -r '.status')
echo "  task=$task status=$status"
[[ "$status" == "done" ]] || { echo "quantbt not done"; exit 1; }
trades=$(call BacktestService GetTrades "{\"jobId\":\"$task\"}")
nt=$(echo "$trades" | jq '.trades | length')
echo "  quantbt trades=$nt"
ge "$nt" 1 || exit 1
echo "  SC-23 OK"
