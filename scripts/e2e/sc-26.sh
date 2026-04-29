#!/usr/bin/env bash
# SC-26 M-E AlertGate：free 用户对 critical 告警应被 tier_blocked
set -uo pipefail
source "$(dirname "$0")/_lib.sh"

uid="sc26_free_user"
# 设置 free 等级
call AlertService SetUserTier "{\"userId\":\"$uid\",\"tier\":\"free\",\"aiMaxPerDay\":20}" >/dev/null

# 触发 critical 告警
out=$(call AlertService DecideAlert "{\"userId\":\"$uid\",\"alertType\":\"sc26_test\",\"severity\":\"critical\",\"pairs\":[\"EURUSD\"]}")
send=$(echo "$out" | jq -r '.send // false')
reason=$(echo "$out" | jq -r '.reason')
echo "  send=$send reason=$reason"
[[ "$send" == "false" && "$reason" == "tier_blocked" ]] || { echo "free critical 应被拦截"; exit 1; }

# 升级 premium 后应放行
call AlertService SetUserTier "{\"userId\":\"$uid\",\"tier\":\"premium\",\"aiMaxPerDay\":200}" >/dev/null
out=$(call AlertService DecideAlert "{\"userId\":\"$uid\",\"alertType\":\"sc26_test_premium\",\"severity\":\"critical\",\"pairs\":[\"EURUSD\"]}")
send=$(echo "$out" | jq -r '.send')
echo "  premium send=$send"
[[ "$send" == "true" ]] || { echo "premium critical 应放行"; exit 1; }

# 历史里至少有 2 条
hist=$(call AlertService GetAlertHistory "{\"userId\":\"$uid\",\"limit\":10}")
n=$(echo "$hist" | jq '.items | length')
echo "  history=$n"
ge "$n" 2 || exit 1
echo "  SC-26 OK"
