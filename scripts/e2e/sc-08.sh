#!/usr/bin/env bash
# SC-8 FedWatch FOMC 概率（probabilities 数组非空 + 概率和≈1）
set -uo pipefail
source "$(dirname "$0")/_lib.sh"
out=$(call FedWatchService GetFOMCProbabilities '{}')
n=$(echo "$out" | jq '.probabilities | length')
echo "  probabilities=$n"
ge "$n" 1 || exit 1
sum=$(echo "$out" | jq '[.probabilities[].probability] | add')
echo "  sum=$sum"
awk -v s="$sum" 'BEGIN{exit !(s+0 >= 0.95 && s+0 <= 1.05 || s+0 >= 95 && s+0 <= 105)}'
