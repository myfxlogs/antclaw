#!/usr/bin/env bash
# SC-2 SSE jobs 心跳（30s 内至少 1 条事件 / 心跳）
set -uo pipefail
source "$(dirname "$0")/_lib.sh"
TMP=$(mktemp)
( curl -s --max-time 30 -N "${API_BASE}/sse/jobs" >"$TMP" 2>&1 || true ) &
PID=$!
sleep 30
kill $PID 2>/dev/null || true
LINES=$(wc -l <"$TMP")
echo "  收到 $LINES 行 SSE"
ge "$LINES" 1
