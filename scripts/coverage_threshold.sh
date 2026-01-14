#!/usr/bin/env bash
set -euo pipefail

THRESHOLD="${1:-80}"
COVER_FILE="coverage/coverage.out"

if [[ ! -f "$COVER_FILE" ]]; then
  echo "coverage file not found: $COVER_FILE"
  echo "run: ./scripts/coverage.sh"
  exit 2
fi

total=$(go tool cover -func="$COVER_FILE" | awk '/^total:/ {print $3}' | tr -d '%')
# total may be like 81.2
total_int=$(printf "%.0f" "$total")

echo "Total coverage: ${total}% (threshold: ${THRESHOLD}%)"

if (( total_int < THRESHOLD )); then
  echo "Coverage threshold not met."
  exit 1
fi
