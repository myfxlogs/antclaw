#!/usr/bin/env bash
# SC-9 Treasury 收益率曲线（tenors≥10）
set -uo pipefail
source "$(dirname "$0")/_lib.sh"
out=$(call TreasuryService GetCurve '{}')
n=$(echo "$out" | jq '.points | length')
echo "  tenors=$n"
ge "$n" 10
