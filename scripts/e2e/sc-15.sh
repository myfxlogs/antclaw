#!/usr/bin/env bash
# SC-15 AI Chat（zhipu glm-4.5）
# 注意 Chat 是 bidi 流式，需用 buf curl 或 grpcurl；这里用 unary Outlook 替代验证 LLM 通路。
set -uo pipefail
source "$(dirname "$0")/_lib.sh"
out=$(call AIService Outlook '{"pair":"BTCUSD","timeframe":"1d","locale":"zh"}')
if is_error "$out"; then
  msg=$(echo "$out" | jq -r '.message')
  echo "  AI 通路错误：$msg"
  if echo "$msg" | grep -q "余额不足\|insufficient\|429"; then
    echo "  → 凭据/账户余额问题，非代码缺陷"
  fi
  exit 1
fi
v=$(echo "$out" | jq -r '.summary')
echo "  summary 长度=${#v}"
[[ ${#v} -gt 10 ]]
