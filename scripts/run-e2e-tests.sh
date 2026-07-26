#!/usr/bin/env bash
# run-e2e-tests.sh — Build the azfloci binary, then run the e2e test suite.
# Usage: ./scripts/run-e2e-tests.sh [extra go test flags...]
#
# Skips Azure-dependent tests automatically when not authenticated.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "==> Building azfloci binary..."
cd "$REPO_ROOT"
go build -race -o bin/azfloci ./cmd/azfloci
echo "    Built: bin/azfloci"

echo "==> Running e2e tests..."
go test -race -tags=e2e -timeout 300s -v ./tests/e2e/... "$@"

echo "==> Done."
