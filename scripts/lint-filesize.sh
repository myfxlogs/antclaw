#!/usr/bin/env bash
# 800 行硬上限检查：仅 Markdown 文档例外。
# 覆盖：Go 源码（含测试）、Proto、前端 TS/TSX/Vue、Shell、Python、SQL。
# 排除：gen/ 下 buf 生成的机械产物（pb.go / *_pb.ts / *_connect.ts），node_modules、dist、build。
set -euo pipefail
THRESHOLD=800
DIRS=(backend frontend proto scripts deploy)
FILES=$(
  for d in "${DIRS[@]}"; do
    [ -d "$d" ] || continue
    find "$d" -type f \
      \( -name "*.go" -o -name "*.proto" \
         -o -name "*.ts" -o -name "*.tsx" -o -name "*.vue" \
         -o -name "*.sh" -o -name "*.py" -o -name "*.sql" \) \
      ! -path "*/node_modules/*" ! -path "*/dist/*" ! -path "*/build/*"
  done
)
VIOLATIONS=$(echo "$FILES" | xargs -r wc -l 2>/dev/null | awk -v t=$THRESHOLD '$1 > t && $2 != "total" {print $0}')
if [ -n "$VIOLATIONS" ]; then
  echo "ERROR: 以下文件超过 $THRESHOLD 行（仅 Markdown 文档例外）："
  echo "$VIOLATIONS"
  exit 1
fi
echo "OK: 所有受检文件 ≤ $THRESHOLD 行"
