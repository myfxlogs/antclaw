#!/usr/bin/env bash
# SC-28 M-F AI 记忆 + 工具调用 + 限流
set -uo pipefail
source "$(dirname "$0")/_lib.sh"

uid="sc28_user"

# 写一条记忆
out=$(call AIService RememberFact "{\"userId\":\"$uid\",\"scope\":\"global\",\"key\":\"risk_pref\",\"value\":\"moderate\"}")
mid=$(echo "$out" | jq -r '.id')
echo "  memory id=$mid"
[[ -n "$mid" && "$mid" != "null" ]] || { echo "remember failed"; exit 1; }

# 取回
out=$(call AIService RecallFact "{\"userId\":\"$uid\",\"scope\":\"global\",\"key\":\"risk_pref\"}")
val=$(echo "$out" | jq -r '.value')
echo "  recalled value=$val"
[[ "$val" == "moderate" ]] || { echo "recall mismatch"; exit 1; }

# 通过工具执行器读取记忆
out=$(call AIService RunWithTools "{\"userId\":\"$uid\",\"message\":\"What is my risk pref?\"}")
ans=$(echo "$out" | jq -r '.answer')
calls=$(echo "$out" | jq '.calls | length')
echo "  tools answer=\"$ans\" calls=$calls"
[[ "$ans" == *"moderate"* ]] || { echo "tool answer should mention moderate"; exit 1; }
ge "$calls" 1 || { echo "no tool call recorded"; exit 1; }

# 限流：check 状态返回有效字段
rl=$(call AIService CheckRateLimit "{\"userId\":\"$uid\"}")
maxN=$(echo "$rl" | jq -r '.maxPerDay // 0')
used=$(echo "$rl" | jq -r '.usedToday // 0')
echo "  ratelimit used=$used max=$maxN"
ge "$maxN" 1 || exit 1

echo "  SC-28 OK"
