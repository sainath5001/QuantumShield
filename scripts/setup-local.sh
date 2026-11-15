#!/usr/bin/env bash
# One-shot local dependency bootstrap (does not start services).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

echo "==> Q-Guard local setup"

if [[ ! -f "$ROOT/.env" ]]; then
  cp "$ROOT/.env.example" "$ROOT/.env"
  echo "Created .env from .env.example — edit QGUARD_ETH_RPC_URL and QGUARD_ETHERSCAN_API_KEY"
else
  echo ".env already present"
fi

if [[ ! -f "$ROOT/frontend/.env.local" ]]; then
  cp "$ROOT/frontend/.env.example" "$ROOT/frontend/.env.local"
  echo "Created frontend/.env.local"
else
  echo "frontend/.env.local already present"
fi

echo "==> Go modules"
(cd "$ROOT/backend" && go mod download)

echo "==> Python venv + quantum deps"
if [[ ! -d "$ROOT/quantum-engine/.venv" ]]; then
  python3 -m venv "$ROOT/quantum-engine/.venv"
fi
# shellcheck disable=SC1091
source "$ROOT/quantum-engine/.venv/bin/activate"
python -m pip install -q --upgrade pip
pip install -q -e "$ROOT/quantum-engine[dev,api]"
deactivate

echo "==> Frontend npm install"
(cd "$ROOT/frontend" && npm install --no-audit --no-fund)

cat <<EOF

Setup complete.

Next:
  ./scripts/run-local.sh          # how to start three terminals
  # edit .env if you have not set RPC URL + Etherscan key
  ./scripts/run-api.sh
  ./scripts/run-quantum.sh
  ./scripts/run-frontend.sh
  ./scripts/test-all.sh

Docs: docs/local-setup.md
EOF
