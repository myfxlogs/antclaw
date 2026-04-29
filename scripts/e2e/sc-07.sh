#!/usr/bin/env bash
# SC-7 SEC EDGAR 文件列表（Apple CIK 0000320193 至少 1 条）
set -uo pipefail
source "$(dirname "$0")/_lib.sh"
out=$(call SECService ListFilings '{"cik":"0000320193","limit":10}')
n=$(echo "$out" | jq '(.items // .filings) | length')
echo "  filings=$n"
ge "$n" 1
