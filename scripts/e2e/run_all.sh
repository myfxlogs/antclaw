#!/usr/bin/env bash
# 跑全部 18 个 SC 场景，输出 PASS/FAIL 矩阵。
# 用法：bash scripts/e2e/run_all.sh
set -uo pipefail
cd "$(dirname "$0")/../.."

source ./scripts/e2e/_lib.sh

PASS=0
FAIL=0
RESULTS=()

run_one() {
  local sc="$1"
  if bash "./scripts/e2e/${sc}.sh"; then
    PASS=$((PASS+1))
    RESULTS+=("$sc PASS")
  else
    FAIL=$((FAIL+1))
    RESULTS+=("$sc FAIL")
  fi
  echo
}

for sc in sc-01 sc-02 sc-03 sc-04 sc-05 sc-06 sc-07 sc-08 sc-09 \
          sc-10 sc-11 sc-12 sc-13 sc-14 sc-15 sc-16 sc-17 sc-18 sc-19 sc-20 sc-21 sc-22 sc-23 sc-24 sc-25 sc-26 sc-27 sc-28 sc-29 sc-30; do
  if [[ -f "./scripts/e2e/${sc}.sh" ]]; then
    run_one "$sc"
  fi
done

echo "===== E2E 总结 ====="
for r in "${RESULTS[@]}"; do echo "  $r"; done
echo "Total: PASS=$PASS FAIL=$FAIL"
exit $FAIL
