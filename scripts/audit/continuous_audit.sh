#!/usr/bin/env bash
# continuous_audit.sh —— 跑 E2E + feature_audit，对比上次结果，差异化输出。
# 用法：bash scripts/audit/continuous_audit.sh
set -uo pipefail
cd "$(dirname "$0")/../.."

REPORT_DIR="${REPORT_DIR:-/tmp/antclaw_audit}"
mkdir -p "$REPORT_DIR"
ts=$(date +%Y%m%d_%H%M%S)
out="$REPORT_DIR/audit_$ts.log"

{
  echo "=== E2E ==="
  bash scripts/e2e/run_all.sh
  echo
  echo "=== feature_audit ==="
  bash scripts/audit/feature_audit.sh
} > "$out" 2>&1

echo "Report: $out"
tail -30 "$out"
