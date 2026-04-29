#!/usr/bin/env bash
# SC-3 GEX（Deribit BTC 真数据，strikes>10）
set -uo pipefail
source "$(dirname "$0")/_lib.sh"
out=$(call OptionsService GetGEX '{"asset":"BTC"}')
n=$(echo "$out" | jq '.strikes | length')
echo "  strikes 数量=$n"
ge "$n" 10
