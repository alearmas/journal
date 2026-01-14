#!/usr/bin/env bash
set -euo pipefail

echo "==> go test (all packages)"
go test ./... -count=1

echo "==> go test -race"
go test ./... -race -count=1
