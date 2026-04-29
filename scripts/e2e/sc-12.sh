#!/usr/bin/env bash
# SC-12 Skew（rr_25d 字段返回真值）
set -uo pipefail
source "$(dirname "$0")/_lib.sh"
out=$(call OptionsService GetOptionsSkew '{"asset":"BTC"}')
v=$(echo "$out" | jq -r '.rr_25d // .rr25d')
echo "  rr_25d=$v"
[[ -n "$v" && "$v" != "null" ]]
