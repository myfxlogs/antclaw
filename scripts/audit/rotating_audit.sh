#!/usr/bin/env bash
# rotating_audit.sh —— 每日定时：滚动保留最近 14 份审计报告。
# 用法：bash scripts/audit/rotating_audit.sh
set -uo pipefail
cd "$(dirname "$0")/../.."

REPORT_DIR="${REPORT_DIR:-/tmp/antclaw_audit}"
KEEP="${KEEP:-14}"
mkdir -p "$REPORT_DIR"

bash scripts/audit/continuous_audit.sh

# 保留最近 KEEP 份
ls -1t "$REPORT_DIR"/audit_*.log 2>/dev/null | tail -n +"$((KEEP+1))" | xargs -r rm -f
echo "Rotated; kept $KEEP latest in $REPORT_DIR"
