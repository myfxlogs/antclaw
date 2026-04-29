#!/usr/bin/env bash
# SC-11 IV Surface（Deribit BTC，points>50）
set -uo pipefail
source "$(dirname "$0")/_lib.sh"
out=$(call OptionsService GetIVSurface '{"asset":"BTC"}')
n=$(echo "$out" | jq '.points | length')
echo "  points=$n"
ge "$n" 50
