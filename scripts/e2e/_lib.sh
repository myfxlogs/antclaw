#!/usr/bin/env bash
# 公共工具：API_BASE / 调用 RPC / JSON 断言。
# 使用方式：source ./scripts/e2e/_lib.sh
set -uo pipefail

: "${API_BASE:=http://localhost:8082}"
: "${TOKEN:=}"

# call <Service> <Method> <json-body>
call() {
  local svc="$1" m="$2" body="$3"
  local hdr=(-H 'Content-Type: application/json')
  if [[ -n "$TOKEN" ]]; then
    hdr+=(-H "Authorization: Bearer $TOKEN")
  fi
  curl -s --max-time 60 -X POST "${API_BASE}/antclaw.v1.${svc}/${m}" "${hdr[@]}" -d "$body"
}

# expect <output> <jq filter> <description>
# 通过 jq 取值；非空且非 0/false 视为通过。
expect() {
  local out="$1" filter="$2" desc="$3"
  local v
  v=$(echo "$out" | jq -r "$filter" 2>/dev/null || echo "")
  if [[ -z "$v" || "$v" == "null" || "$v" == "false" || "$v" == "0" ]]; then
    echo "  FAIL: $desc (filter=$filter, value=$v)" >&2
    echo "$out" | head -c 400 >&2; echo >&2
    return 1
  fi
  echo "  OK: $desc → $v"
}

# ge <a> <b> 断言 a>=b（数字）
ge() {
  awk -v a="$1" -v b="$2" 'BEGIN{exit !(a+0>=b+0)}'
}

# 简单错误捕获：如果 RPC 返回 {"code":...,"message":...} 则视为错误
is_error() {
  echo "$1" | jq -e '.code' >/dev/null 2>&1
}

run_case() {
  local id="$1"; shift
  local desc="$1"; shift
  echo "[$id] $desc"
  if "$@"; then
    echo "  → $id PASS"
    return 0
  else
    echo "  → $id FAIL" >&2
    return 1
  fi
}
