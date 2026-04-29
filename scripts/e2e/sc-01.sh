#!/usr/bin/env bash
# SC-1 COT 摘要（CFTC Socrata 真数据）
set -uo pipefail
source "$(dirname "$0")/_lib.sh"
out=$(call COTService GetSummary '{"pair":"EURUSD"}')
if is_error "$out"; then
  out=$(call COTService GetSummary '{"asset":"EUR"}')
fi
# 响应含 latest.nonCommLong/nonCommShort 即表示真数据已返回
v=$(echo "$out" | jq -r '.latest.nonCommLong // empty')
echo "  latest.nonCommLong=$v"
[[ -n "$v" && "$v" != "null" && "$v" != "0" ]]
