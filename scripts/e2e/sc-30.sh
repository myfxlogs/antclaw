#!/usr/bin/env bash
# SC-30 终极验收：关键 DB 表 count > 0
set -uo pipefail
docker exec antclaw-postgres psql -U antclaw -d antclaw -tAc "
  SELECT t,c FROM (
    SELECT 'backtest_trades' AS t, COUNT(*) AS c FROM backtest_trades UNION ALL
    SELECT 'backtest_metrics_by_regime', COUNT(*) FROM backtest_metrics_by_regime UNION ALL
    SELECT 'signal_calibrations', COUNT(*) FROM signal_calibrations UNION ALL
    SELECT 'alert_log', COUNT(*) FROM alert_log UNION ALL
    SELECT 'ai_memories', COUNT(*) FROM ai_memories
  ) x;" | while IFS='|' read -r tbl cnt; do
  echo "  $tbl=$cnt"
  if [[ "${cnt:-0}" -le 0 ]]; then
    echo "  $tbl is empty (0)"
    exit 1
  fi
done
echo "  SC-30 OK"
