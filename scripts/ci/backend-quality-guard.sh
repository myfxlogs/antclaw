#!/usr/bin/env bash
set -euo pipefail
# backend-quality-guard.sh — 扫描后端生产路径中的反模式
# 白名单：测试文件 / 文档 / 生成代码

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
ERR=0

scan() {
    local pattern="$1"
    local msg="$2"
    local matches
    matches=$(rg -n --include='*.go' "$pattern" "$ROOT/backend/internal" "$ROOT/backend/cmd" 2>/dev/null | grep -v '_test.go' | grep -v '/gen/' | grep -v '/.git/' || true)
    if [ -n "$matches" ]; then
        echo "❌ $msg:"
        echo "$matches"
        echo
        ERR=1
    fi
}

echo "=== Backend Quality Guard ==="

# 禁止在生产路径返回样例/合成数据
scan 'sample data|sampleData|getSample|fallback data' \
    'sample/fallback data reference in production code'

# 禁止 demo 用户或硬编码默认用户
scan 'demo@antclaw|user-1' \
    'hardcoded demo user reference'

# 禁止吞掉错误（_, _ =）
scan '_, _ = .*\.(Exec|Query|Scan|Send|Do|Encode|Decode)' \
    'silent error discard (_, _ =) in production code'

# 禁止生产代码中的 math/rand
scan '"math/rand"' \
    'math/rand import in production code (use crypto/rand)'

if [ $ERR -eq 0 ]; then
    echo "✅ All quality checks passed"
else
    echo "❌ Quality checks failed"
fi
exit $ERR
