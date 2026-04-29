#!/usr/bin/env bash
# SC-16 AI Cached Interpret（第二次调 cache_hit=true）
set -uo pipefail
source "$(dirname "$0")/_lib.sh"
body='{"dataType":"price","rawData":"BTC=68000","question":"trend","locale":"zh"}'
out1=$(call AIService Interpret "$body")
if is_error "$out1"; then
  echo "  首次调用失败：$(echo "$out1" | jq -r '.message')"
  exit 1
fi
out2=$(call AIService Interpret "$body")
hit=$(echo "$out2" | jq -r '.cacheHit // .cache_hit')
echo "  cache_hit=$hit"
[[ "$hit" == "true" ]]
