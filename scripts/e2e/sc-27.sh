#!/usr/bin/env bash
# SC-27 M-E AlertGate：briefing 推送链路 alert_log 写入 sent=true
set -uo pipefail
source "$(dirname "$0")/_lib.sh"

uid="sc27_briefing_user"
call AlertService SetUserTier "{\"userId\":\"$uid\",\"tier\":\"premium\",\"aiMaxPerDay\":200}" >/dev/null
call AlertService UpdatePreferences "{\"userId\":\"$uid\",\"pairs\":[\"EURUSD\"],\"timezone\":\"UTC\"}" >/dev/null

# 模拟 scheduler 触发 briefing：直接调 DecideAlert（实际 scheduler 也走该闸门）
out=$(call AlertService DecideAlert "{\"userId\":\"$uid\",\"alertType\":\"briefing\",\"severity\":\"medium\",\"pairs\":[\"EURUSD\"]}")
send=$(echo "$out" | jq -r '.send')
reason=$(echo "$out" | jq -r '.reason')
echo "  briefing send=$send reason=$reason"
[[ "$send" == "true" && "$reason" == "ok" ]] || { echo "briefing 应放行"; exit 1; }

# 历史中至少 1 条 sent=true
hist=$(call AlertService GetAlertHistory "{\"userId\":\"$uid\",\"limit\":10}")
sent_n=$(echo "$hist" | jq '[.items[] | select(.sent==true)] | length')
echo "  sent_in_history=$sent_n"
ge "$sent_n" 1 || exit 1
echo "  SC-27 OK"
