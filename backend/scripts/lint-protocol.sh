#!/usr/bin/env bash
set -euo pipefail

VIOLATIONS=$(grep -rnE 'mux\.HandleFunc\(' backend/cmd backend/internal 2>/dev/null \
  | grep -vE 'sse/' || true)

if [ -n "$VIOLATIONS" ]; then
  echo "ERROR: 禁止新增 REST handler，仅允许 Connect / SSE。违规："
  echo "$VIOLATIONS"
  exit 1
fi

WS=$(grep -rnE 'gorilla/websocket|nhooyr.io/websocket' backend 2>/dev/null | grep -v 'backend/scripts/lint-protocol.sh' || true)
if [ -n "$WS" ]; then
  echo "ERROR: 禁止 WebSocket。违规："
  echo "$WS"
  exit 1
fi

echo "OK: 后端协议合规"
