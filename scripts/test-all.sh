#!/usr/bin/env bash
# Run unit/integration-light suites for Go, Python, and frontend.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FAIL=0

echo "==> Go vet + test (+ race)"
(
  cd "$ROOT/backend"
  go vet ./...
  go test ./...
  go test -race ./...
) || FAIL=1

echo "==> Quantum pytest"
(
  cd "$ROOT/quantum-engine"
  if [[ -f .venv/bin/activate ]]; then
    # shellcheck disable=SC1091
    source .venv/bin/activate
  fi
  python -m pytest
) || FAIL=1

echo "==> Frontend test + typecheck + lint"
(
  cd "$ROOT/frontend"
  npm test
  npm run typecheck
  npm run lint
) || FAIL=1

if [[ "$FAIL" -ne 0 ]]; then
  echo "test-all: FAILED"
  exit 1
fi
echo "test-all: OK"
