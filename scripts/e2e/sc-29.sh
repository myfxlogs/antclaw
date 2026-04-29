#!/usr/bin/env bash
# SC-29 M-G Python 量化引擎 + feature_audit
set -uo pipefail
source "$(dirname "$0")/_lib.sh"

# 用合成 CSV（单调上升 + 噪声）测试 quant_engine.py
csv=$(python3 - <<'PY'
import random, math
random.seed(7)
print("date,close")
v = 100.0
for i in range(252):
    v *= 1 + (random.gauss(0,1) * 0.01)
    print(f"d{i},{v:.4f}")
PY
)
out=$(echo "$csv" | python3 /opt/antclaw/scripts/quant/quant_engine.py --csv -)
echo "  quant_engine: $out"
sharpe=$(echo "$out" | jq -r '.sharpe // 0')
n=$(echo "$out" | jq -r '.n_bars // 0')
ge "$n" 100 || exit 1

# feature_audit
bash /opt/antclaw/scripts/audit/feature_audit.sh > /tmp/feature_audit.log 2>&1
fa_exit=$?
tail -5 /tmp/feature_audit.log
[[ $fa_exit -eq 0 ]] || { echo "feature_audit failed"; exit 1; }

echo "  SC-29 OK"
