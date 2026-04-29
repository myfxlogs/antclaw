#!/usr/bin/env bash
# SC-10 MacroExtras WorldBank 美国 GDP（points>0）
set -uo pipefail
source "$(dirname "$0")/_lib.sh"
out=$(call MacroExtrasService GetSeries '{"source":"worldbank","seriesId":"USA/NY.GDP.MKTP.CD"}')
n=$(echo "$out" | jq '.points | length')
echo "  points=$n"
ge "$n" 1
