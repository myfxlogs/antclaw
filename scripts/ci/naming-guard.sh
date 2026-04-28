#!/bin/bash
# naming-guard.sh - Zero tolerance check for ark/ARK strings
# Excludes: docs/ARK-Intelligent-功能清单.md (reference document)

set -e

echo "=== AntClaw Naming Guard ==="
echo "Checking for forbidden 'ark'/'ARK' strings..."

# Define excluded paths
EXCLUDE_PATHS=(
    "docs/ARK-Intelligent-功能清单.md"
    "Emulator/ark-intelligent"
    ".git"
    "vendor"
    "node_modules"
    "gen"
)

# Build exclude pattern for grep
EXCLUDE_PATTERN=""
for path in "${EXCLUDE_PATHS[@]}"; do
    if [ -z "$EXCLUDE_PATTERN" ]; then
        EXCLUDE_PATTERN="$path"
    else
        EXCLUDE_PATTERN="$EXCLUDE_PATTERN|$path"
    fi
done

# Check for forbidden strings
# Pattern: whole word 'ark' or any 'ARK' (case sensitive for uppercase)
# Excluded: docs/ (except in paths), Emulator/, .git/, vendor/, node_modules/, gen/
FOUND=$(grep -r -E "\bark\b|ARK" \
    --include="*.go" \
    --include="*.ts" \
    --include="*.tsx" \
    --include="*.js" \
    --include="*.jsx" \
    --include="*.proto" \
    --include="*.yaml" \
    --include="*.yml" \
    --include="*.json" \
    --include="*.md" \
    --include="Dockerfile*" \
    --include="Makefile" \
    --include="*.sh" \
    --include="*.py" \
    --exclude-dir=Emulator \
    --exclude-dir=.git \
    --exclude-dir=vendor \
    --exclude-dir=node_modules \
    --exclude-dir=gen \
    --exclude-dir=.github \
    --exclude-dir=bak0428 \
    --exclude="ARK-Intelligent-功能清单.md" \
    --exclude="naming-guard.sh" \
    --exclude="similarity-guard.py" \
    . 2>/dev/null | grep -v "^./docs/" || true)

if [ -n "$FOUND" ]; then
    echo "ERROR: Forbidden 'ark'/'ARK' strings detected!"
    echo ""
    echo "Violations:"
    echo "$FOUND"
    echo ""
    echo "AntClaw CI has zero tolerance for 'ark'/'ARK' naming."
    echo "See: docs/AntClaw-重构解决方案.md §2.1 Naming Baseline"
    exit 1
fi

echo "✓ Naming check passed. No forbidden strings found."
exit 0
