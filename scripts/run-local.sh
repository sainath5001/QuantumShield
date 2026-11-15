#!/usr/bin/env bash
# Print how to start all three local processes (separate terminals).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

cat <<EOF
Q-Guard local stack (three terminals)

1) Go backend (secrets stay here — RPC + Etherscan):
   cp -n "$ROOT/.env.example" "$ROOT/.env"   # then edit RPC/key
   "$ROOT/scripts/run-api.sh"
   → http://127.0.0.1:8080

2) Quantum research sidecar (educational only):
   "$ROOT/scripts/run-quantum.sh"
   → http://127.0.0.1:8090

3) Next.js frontend:
   "$ROOT/scripts/run-frontend.sh"
   → http://127.0.0.1:3000

Smoke checks:
   curl -s http://127.0.0.1:8080/api/v1/health | head
   curl -s http://127.0.0.1:8090/health | head
   curl -s -X POST http://127.0.0.1:8080/api/v1/scan -H 'Content-Type: application/json' \\
     -d '{"address":"0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0"}'

Notes:
- Scan cancellation is not supported by the Go API.
- Quantum results never feed wallet risk/exposure.
- Backend scan/report storage is in-memory (lost on restart).
EOF
