#!/usr/bin/env bash
# Start the educational quantum FastAPI sidecar (localhost:8090 by default).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT/quantum-engine"

if [[ -f "$ROOT/.env" ]]; then
  set -a
  # shellcheck disable=SC1091
  source "$ROOT/.env"
  set +a
fi

if [[ ! -d .venv ]]; then
  echo "Creating quantum-engine/.venv …"
  python3 -m venv .venv
  # shellcheck disable=SC1091
  source .venv/bin/activate
  python -m pip install -q --upgrade pip
  pip install -q -e ".[dev,api]"
else
  # shellcheck disable=SC1091
  source .venv/bin/activate
fi

export QGUARD_QUANTUM_HOST="${QGUARD_QUANTUM_HOST:-127.0.0.1}"
export QGUARD_QUANTUM_PORT="${QGUARD_QUANTUM_PORT:-8090}"

echo "Q-Guard quantum API → http://${QGUARD_QUANTUM_HOST}:${QGUARD_QUANTUM_PORT}"
exec qguard-quantum-api
