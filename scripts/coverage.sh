#!/usr/bin/env bash
set -euo pipefail

COVER_DIR="coverage"
mkdir -p "$COVER_DIR"

echo "==> Running tests with coverage (internal packages only)"
go test ./internal/... -count=1 -covermode=atomic -coverprofile="$COVER_DIR/coverage.out"

echo "==> Coverage summary"
go tool cover -func="$COVER_DIR/coverage.out" | tee "$COVER_DIR/coverage.txt"

echo "==> HTML report"
go tool cover -html="$COVER_DIR/coverage.out" -o "$COVER_DIR/coverage.html"

echo "Generated:"
echo " - $COVER_DIR/coverage.out"
echo " - $COVER_DIR/coverage.txt"
echo " - $COVER_DIR/coverage.html"
