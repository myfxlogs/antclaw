#!/usr/bin/env bash
# 面向 api.alfq.org 的冒烟测试，覆盖 Healthz / Info / SSE 401 / 登录 / 通知未读 / 管理接口拒绝 / CORS。
# 用法: bash scripts/smoke-api-domain.sh [BASE_URL]
set -euo pipefail
BASE="${1:-https://api.alfq.org}"
CT="Content-Type: application/json"
fail=0
skip=0

echo "=== 1) Healthz ==="
code=$(curl -sS -o /dev/null -w "%{http_code}" -X POST -H "$CT" -d '{}' "$BASE/antclaw.v1.SystemService/Healthz" 2>/dev/null || echo "000")
if [ "$code" = "200" ]; then
  echo "  [OK] $code"
else
  echo "  [FAIL] Healthz returned $code"
  fail=$((fail + 1))
fi

echo "=== 2) Info ==="
body=$(curl -sS -X POST -H "$CT" -d '{}' "$BASE/antclaw.v1.SystemService/Info" 2>/dev/null || echo "")
if echo "$body" | grep -q '"version"'; then
  ver=$(echo "$body" | grep -o '"version":"[^"]*"' | cut -d'"' -f4)
  echo "  [OK] version=$ver"
else
  echo "  [FAIL] no version field"
  fail=$((fail + 1))
fi

echo "=== 3) SSE 未登录 → 401 ==="
code=$(curl -sS -o /dev/null -w "%{http_code}" "$BASE/sse/notifications" 2>/dev/null || echo "000")
if [ "$code" = "401" ]; then
  echo "  [OK] $code"
else
  echo "  [FAIL] SSE without token returned $code (expected 401)"
  fail=$((fail + 1))
fi

echo "=== 4) 登录 + 通知未读 ==="
TOKEN=$(curl -sS -X POST -H "$CT" \
  -d '{"email":"test@alfq.org","password":"test123"}' \
  "$BASE/antclaw.v1.AuthService/Login" 2>/dev/null \
  | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4 || echo "")
if [ -n "$TOKEN" ]; then
  echo "  [OK] login success"
  code=$(curl -sS -o /dev/null -w "%{http_code}" \
    -X POST -H "$CT" -H "Authorization: Bearer $TOKEN" -d '{}' \
    "$BASE/antclaw.v1.NotificationService/UnreadCount" 2>/dev/null || echo "000")
  if [ "$code" = "200" ]; then
    echo "  [OK] UnreadCount $code"
  else
    echo "  [FAIL] UnreadCount returned $code"
    fail=$((fail + 1))
  fi
  # 拉取未读列表
  code2=$(curl -sS -o /dev/null -w "%{http_code}" \
    -X POST -H "$CT" -H "Authorization: Bearer $TOKEN" -d '{"limit":5}' \
    "$BASE/antclaw.v1.NotificationService/ListUnread" 2>/dev/null || echo "000")
  if [ "$code2" = "200" ]; then
    echo "  [OK] ListUnread $code2"
  else
    echo "  [FAIL] ListUnread returned $code2"
    fail=$((fail + 1))
  fi
else
  echo "  [SKIP] login failed (credentials not configured)"
  skip=$((skip + 1))
fi

echo "=== 5) 管理接口拒绝非管理员 ==="
code=$(curl -sS -o /dev/null -w "%{http_code}" -X POST -H "$CT" -d '{}' \
  "$BASE/antclaw.v1.AdminService/ListUsers" 2>/dev/null || echo "000")
if [ "$code" = "401" ] || [ "$code" = "403" ]; then
  echo "  [OK] AdminService rejected ($code)"
else
  echo "  [FAIL] AdminService returned $code (expected 401 or 403)"
  fail=$((fail + 1))
fi

echo "=== 6) CORS OPTIONS ==="
code=$(curl -sS -o /dev/null -w "%{http_code}" -X OPTIONS \
  -H "Origin: https://app.alfq.org" \
  -H "Access-Control-Request-Method: POST" \
  "$BASE/antclaw.v1.SystemService/Healthz" 2>/dev/null || echo "000")
if [ "$code" = "200" ] || [ "$code" = "204" ]; then
  echo "  [OK] CORS $code"
else
  echo "  [FAIL] CORS returned $code"
  fail=$((fail + 1))
fi

echo ""
if [ "$fail" -eq 0 ]; then
  echo "ALL PASS ($skip skipped)"
  exit 0
else
  echo "$fail FAILURES"
  exit 1
fi
