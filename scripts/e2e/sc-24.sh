#!/usr/bin/env bash
# SC-24 M-D vpbt（POC 突破）
set -uo pipefail
source "$(dirname "$0")/_lib.sh"
out=$(call BacktestService RunVpBt '{"config":{"pair":"EURUSD","numBins":20}}')
task=$(echo "$out" | jq -r '.taskId')
status=$(echo "$out" | jq -r '.status')
echo "  task=$task status=$status"
[[ "$status" == "done" ]] || { echo "vpbt not done"; exit 1; }
trades=$(call BacktestService GetTrades "{\"jobId\":\"$task\"}")
nt=$(echo "$trades" | jq '.trades | length')
echo "  vpbt trades=$nt"
ge "$nt" 1 || exit 1
echo "  SC-24 OK"
