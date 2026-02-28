#!/usr/bin/env bash
set -euo pipefail

HOOK_SRC="scripts/pre-commit"
HOOK_DST=".git/hooks/pre-commit"

if [[ ! -f "$HOOK_SRC" ]]; then
  echo "error: $HOOK_SRC not found"
  exit 1
fi

cp "$HOOK_SRC" "$HOOK_DST"
chmod +x "$HOOK_DST"

echo "pre-commit hook installed at $HOOK_DST"
