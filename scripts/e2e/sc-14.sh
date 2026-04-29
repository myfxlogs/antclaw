#!/usr/bin/env bash
# SC-14 Insider/CryptoSocial 情绪 — 当前 Insider/CryptoSocial 仍待接入；用 Finviz 替代验证 firecrawl 通路
set -uo pipefail
source "$(dirname "$0")/_lib.sh"
out=$(call SentimentExtrasService GetFinvizMetrics '{"ticker":"AAPL"}')
v=$(echo "$out" | jq -r '.shortRatio // .short_ratio')
echo "  AAPL short_ratio=$v"
[[ -n "$v" && "$v" != "null" && "$v" != "0" ]]
