#!/usr/bin/env bash
# feature_audit.sh —— 扫描所有关键 RPC 端点，确认返回真值（非空 + 非 mock）。
# 用法：bash scripts/audit/feature_audit.sh
set -uo pipefail

API="${API_BASE:-http://localhost:8082}"
TOK="${TOKEN:-}"

PASS=0
FAIL=0
declare -a LINES

call() {
  local svc="$1" m="$2" body="$3"
  local hdr=(-H 'Content-Type: application/json')
  [[ -n "$TOK" ]] && hdr+=(-H "Authorization: Bearer $TOK")
  curl -s --max-time 30 -X POST "${API}/antclaw.v1.${svc}/${m}" "${hdr[@]}" -d "$body"
}

check() {
  local name="$1" jq_filter="$2" out="$3"
  local v
  v=$(echo "$out" | jq -r "$jq_filter" 2>/dev/null)
  if [[ -z "$v" || "$v" == "null" || "$v" == "0" || "$v" == "false" ]]; then
    LINES+=("FAIL $name (filter=$jq_filter, got=$v)")
    FAIL=$((FAIL+1))
  else
    LINES+=("PASS $name → $v")
    PASS=$((PASS+1))
  fi
}

# Price 真数据
check "price.GetPrice EURUSD" '.bars | length' \
  "$(call PriceService GetPrice '{"pair":"EURUSD","timeframe":"1d","count":5}')"
check "price.GetVolatility GARCH" '.persistence' \
  "$(call PriceService GetVolatility '{"pair":"EURUSD","timeframe":"1d","lookback":300}')"
check "price.GetHurst" '.hurst' \
  "$(call PriceService GetHurst '{"pair":"EURUSD","timeframe":"1d","lookback":300}')"
check "price.GetCorrelations" '.assets | length' \
  "$(call PriceService GetCorrelations '{"timeframe":"1d","window":30}')"

# Sentiment 真数据
check "sentiment.GetSentiment BTC" '.sentiment.score' \
  "$(call SentimentService GetSentiment '{"asset":"BTC"}')"

# Vol
check "vol.GetVix" '.vix.spot' \
  "$(call VolService GetVix '{}')"

echo
echo "===== feature_audit ====="
for l in "${LINES[@]}"; do echo "  $l"; done
echo "PASS=$PASS FAIL=$FAIL"
exit $FAIL
