#!/usr/bin/env python3
"""
similarity-guard.py - Code similarity detection against reference project
Zero tolerance for high similarity with ark-intelligent codebase.
"""

import sys
import os
import hashlib
from pathlib import Path

# File extensions to check
CHECK_EXTENSIONS = {'.go', '.ts', '.tsx', '.js', '.jsx', '.proto', '.sql'}

# Paths to exclude
EXCLUDE_PATHS = {
    '.git', 'vendor', 'node_modules', 'gen', '.github', 'scripts/ci',
    'docs', 'Emulator'
}

def get_file_hash(filepath: Path) -> str:
    """Generate MD5 hash of file content (normalized)."""
    try:
        content = filepath.read_text(encoding='utf-8', errors='ignore')
        # Normalize: remove whitespace, lowercase
        normalized = ''.join(content.split()).lower()
        return hashlib.md5(normalized.encode()).hexdigest()[:16]
    except Exception:
        return ""

def scan_project(root: Path) -> dict:
    """Scan project and return file hashes."""
    files = {}
    for filepath in root.rglob('*'):
        if filepath.suffix not in CHECK_EXTENSIONS:
            continue
        if any(excl in str(filepath) for excl in EXCLUDE_PATHS):
            continue
        file_hash = get_file_hash(filepath)
        if file_hash:
            files[str(filepath.relative_to(root))] = file_hash
    return files

def check_similarity() -> bool:
    """Check for suspicious similarity patterns."""
    print("=== AntClaw Similarity Guard ===")
    print("Checking for code patterns that indicate copy-paste from reference project...")
    
    root = Path('.')
    project_files = scan_project(root)
    
    if not project_files:
        print("Warning: No source files found to check")
        return True
    
    # For now, just report stats
    # In production, this would compare against ark-intelligent hashes
    print(f"Scanned {len(project_files)} files")
    
    # Check for suspicious patterns in Go files
    suspicious_patterns = [
        b'github.com/arkcode369/ark-intelligent',  # Old import path
        b'ark-intelligent',
        b'ARK_Intelligent',
    ]
    
    violations = []
    for filepath in root.rglob('*.go'):
        if any(excl in str(filepath) for excl in EXCLUDE_PATHS):
            continue
        try:
            content = filepath.read_bytes()
            for pattern in suspicious_patterns:
                if pattern in content:
                    violations.append(f"{filepath}: contains '{pattern.decode()}'")
        except Exception:
            continue
    
    if violations:
        print("ERROR: Suspicious patterns detected!")
        for v in violations[:10]:  # Show first 10
            print(f"  - {v}")
        if len(violations) > 10:
            print(f"  ... and {len(violations) - 10} more")
        print("\nAntClaw requires original implementation. See docs/AntClaw-重构解决方案.md §一")
        return False
    
    print("✓ Similarity check passed")
    return True

if __name__ == '__main__':
    success = check_similarity()
    sys.exit(0 if success else 1)
