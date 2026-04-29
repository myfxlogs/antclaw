#!/usr/bin/env bash
# SC-22 M-C 概率校准（Platt + Isotonic）
set -uo pipefail
source "$(dirname "$0")/_lib.sh"

# 合成 200 对偏置样本：score 越高越倾向 outcome=true
build_synth() {
python3 - <<'PY'
import json, math, random
random.seed(7)
scores=[]; outcomes=[]
for _ in range(200):
    s=random.gauss(0,1)
    p=1/(1+math.exp(-(2*s-0.5)))
    scores.append(s)
    outcomes.append(random.random()<p)
print(json.dumps({"scores":scores,"outcomes":outcomes}))
PY
}
synth=$(build_synth)
scores=$(echo "$synth" | jq -c '.scores')
outcomes=$(echo "$synth" | jq -c '.outcomes')

# Platt
out=$(call SignalsService FitCalibration "{\"modelId\":\"sc22_platt\",\"type\":\"platt\",\"scores\":$scores,\"outcomes\":$outcomes}")
brier_p=$(echo "$out" | jq -r '.brier // 0.5')
echo "  Platt brier=$brier_p"
awk -v v="$brier_p" 'BEGIN{exit !(v+0<0.25 && v+0>0)}' || { echo "Platt brier out of range"; exit 1; }

# Isotonic
out=$(call SignalsService FitCalibration "{\"modelId\":\"sc22_iso\",\"type\":\"isotonic\",\"scores\":$scores,\"outcomes\":$outcomes}")
brier_i=$(echo "$out" | jq -r '.brier // 0.5')
echo "  Isotonic brier=$brier_i"
awk -v v="$brier_i" 'BEGIN{exit !(v+0<0.30 && v+0>0)}' || { echo "Isotonic brier out of range"; exit 1; }

# Predict：score=2 应该映射到 ~高概率
pred=$(call SignalsService PredictCalibrated '{"modelId":"sc22_platt","score":2.0}')
v=$(echo "$pred" | jq -r '.calibrated // 0')
echo "  Predict(score=2) → $v"
awk -v v="$v" 'BEGIN{exit !(v+0>0 && v+0<1)}' || { echo "calibrated out of (0,1)"; exit 1; }

# List
list=$(call SignalsService ListCalibrations '{}')
n=$(echo "$list" | jq '.items | length')
echo "  List items=$n"
ge "$n" 2 || exit 1

echo "  SC-22 OK"
