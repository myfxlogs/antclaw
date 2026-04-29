#!/usr/bin/env bash
# SC-17 SystemService.Healthz（Connect 健康检查替代 REST /health）
set -uo pipefail
source "$(dirname "$0")/_lib.sh"
out=$(call SystemService Healthz '{}')
status=$(echo "$out" | jq -r '.status')
echo "  status=$status"
[[ "$status" == "healthy" ]]
