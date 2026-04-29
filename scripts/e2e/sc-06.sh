#!/usr/bin/env bash
# SC-6 DeFi TVL 协议清单（>=50）
set -uo pipefail
source "$(dirname "$0")/_lib.sh"
out=$(call DeFiService GetTopProtocols '{"limit":50}')
n=$(echo "$out" | jq '(.items // .protocols) | length')
echo "  protocols=$n"
ge "$n" 50
