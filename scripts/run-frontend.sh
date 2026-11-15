#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT/frontend"

if [[ ! -d node_modules ]]; then
  echo "Installing frontend dependencies…"
  npm install --no-audit --no-fund
fi

if [[ ! -f .env.local ]]; then
  cp .env.example .env.local
  echo "Created frontend/.env.local from .env.example"
fi

export NEXT_PUBLIC_QGUARD_API_BASE_URL="${NEXT_PUBLIC_QGUARD_API_BASE_URL:-http://127.0.0.1:8080}"
export NEXT_PUBLIC_QGUARD_QUANTUM_API_BASE_URL="${NEXT_PUBLIC_QGUARD_QUANTUM_API_BASE_URL:-http://127.0.0.1:8090}"

echo "Q-Guard frontend → http://127.0.0.1:3000"
echo "  Go API:      ${NEXT_PUBLIC_QGUARD_API_BASE_URL}"
echo "  Quantum API: ${NEXT_PUBLIC_QGUARD_QUANTUM_API_BASE_URL}"
exec npm run dev -- --hostname 127.0.0.1 --port 3000
