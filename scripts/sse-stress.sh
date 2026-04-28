#!/usr/bin/env bash
# SSE 压测：并发打开 N 路 SSE 连接、统计 M 秒内总事件数与稳定性。
# 用法：
#   API_BASE=http://localhost:8082 PATHS="/sse/jobs,/sse/audit" \
#   CONNECTIONS=10 DURATION=15 bash scripts/sse-stress.sh
#
# 输出：
#   - 每路连接的 HTTP 状态、收到的事件行数
#   - 总连接数 / 成功连接数 / 总事件行数 / 平均延迟
set -euo pipefail

BASE="${API_BASE:-http://localhost:8082}"
PATHS_CSV="${PATHS:-/sse/jobs}"
CONN="${CONNECTIONS:-5}"
DUR="${DURATION:-10}"
TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

IFS=',' read -ra PATHS <<<"$PATHS_CSV"

start_ts=$(date +%s)
for i in $(seq 1 "$CONN"); do
  ep="${PATHS[$(( (i-1) % ${#PATHS[@]} ))]}"
  url="$BASE$ep"
  # -N 禁用缓冲，让事件实时落到文件
  (
    code=$(curl -sS -N --max-time "$DUR" -o "$TMPDIR/conn-$i.body" -w "%{http_code}" "$url" 2>"$TMPDIR/conn-$i.err" || true)
    echo "$code" >"$TMPDIR/conn-$i.code"
  ) &
done
wait
end_ts=$(date +%s)
elapsed=$((end_ts - start_ts))

ok=0
events=0
for i in $(seq 1 "$CONN"); do
  code=$(cat "$TMPDIR/conn-$i.code" 2>/dev/null || echo 000)
  lines=$(wc -l <"$TMPDIR/conn-$i.body" 2>/dev/null || echo 0)
  if [ "$code" = "200" ]; then ok=$((ok+1)); fi
  events=$((events + lines))
  printf "  conn-%02d  http=%s  lines=%d\n" "$i" "$code" "$lines"
done

echo "---"
echo "总连接: $CONN, 成功(200): $ok, 失败: $((CONN-ok))"
echo "事件总行数: $events, 持续时间: ${elapsed}s"
if [ "$ok" -ne "$CONN" ]; then
  echo "FAIL: 部分连接未返回 200"
  exit 1
fi
echo "OK: SSE 压测通过"
