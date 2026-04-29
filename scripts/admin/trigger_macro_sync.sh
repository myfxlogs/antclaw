#!/usr/bin/env bash
# trigger_macro_sync.sh —— 通过 admin RPC 触发一次 macro-sync 作业，把 FRED 数据写入
# data_snapshots（依赖：valid FRED API key 已通过 /datasources 配置进系统）。
#
# 用法：
#   TOKEN=<admin-jwt> bash scripts/admin/trigger_macro_sync.sh
set -uo pipefail

API="${API_BASE:-http://127.0.0.1:8082}"
TOK="${TOKEN:-}"

if [[ -z "$TOK" ]]; then
  echo "请设置环境变量 TOKEN=<admin-jwt>"
  exit 2
fi

curl -fsS -X POST "${API}/antclaw.v1.AdminService/RunJob" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOK" \
  -d '{"job":"macro-sync"}' | jq .

echo
echo "==== data_snapshots 现状 ===="
docker exec antclaw-postgres psql -U antclaw -d antclaw -tAc "
  SELECT source, COUNT(*) AS rows, MAX(time) AS latest
    FROM data_snapshots GROUP BY source ORDER BY rows DESC;"
