#!/usr/bin/env bash
# SC-25 M-D cta（Donchian breakout）
set -uo pipefail
source "$(dirname "$0")/_lib.sh"
out=$(call BacktestService RunCtaBt '{"config":{"pair":"EURUSD","strategy":"donchian","lookback":20}}')
task=$(echo "$out" | jq -r '.taskId')
status=$(echo "$out" | jq -r '.status')
echo "  task=$task status=$status"
[[ "$status" == "done" ]] || { echo "cta not done"; exit 1; }
trades=$(call BacktestService GetTrades "{\"jobId\":\"$task\"}")
nt=$(echo "$trades" | jq '.trades | length')
echo "  cta trades=$nt"
ge "$nt" 1 || exit 1
echo "  SC-25 OK"
