#!/usr/bin/env bash
# SC-5 链上分析（CoinGecko/onchain_metrics 真数据）
set -uo pipefail
source "$(dirname "$0")/_lib.sh"
out=$(call OnchainService GetMetrics '{"asset":"BTC"}')
n=$(echo "$out" | jq '.points | length')
echo "  points=$n"
ge "$n" 7   # 至少 7 天数据
