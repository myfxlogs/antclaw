#!/usr/bin/env bash
# SC-13 Vol DVOL（Deribit 真数据）+ VIX 真值（CBOE）
set -uo pipefail
source "$(dirname "$0")/_lib.sh"
out=$(call VolService GetDvol '{"asset":"BTC"}')
v=$(echo "$out" | jq -r '.dvol.value')
echo "  DVOL BTC=$v"
[[ -n "$v" && "$v" != "null" && "$v" != "0" ]] || exit 1
out2=$(call VolService GetVix '{}')
spot=$(echo "$out2" | jq -r '.vix.spot')
echo "  VIX=$spot"
[[ -n "$spot" && "$spot" != "null" && "$spot" != "0" ]]
