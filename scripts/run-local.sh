#!/usr/bin/env bash
# Print how to start / stop the local demo stack (three terminals).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

cat <<EOF
Q-Guard local MVP stack
=======================

Prerequisites: Go, Node/npm, Python 3.10+, Mainnet RPC URL, Etherscan API key.
First-time bootstrap:  $ROOT/scripts/setup-local.sh
Full guide:            $ROOT/docs/local-setup.md
Troubleshooting:       $ROOT/docs/troubleshooting.md
Release checklist:     $ROOT/docs/release-checklist.md

1) Configure (once)
   cp -n "$ROOT/.env.example" "$ROOT/.env"
   # edit QGUARD_ETH_RPC_URL and QGUARD_ETHERSCAN_API_KEY
   cp -n "$ROOT/frontend/.env.example" "$ROOT/frontend/.env.local"

2) Terminal A — Go API (holds RPC + Etherscan secrets)
   "$ROOT/scripts/run-api.sh"
   → http://127.0.0.1:8080

3) Terminal B — Quantum research sidecar (educational only)
   "$ROOT/scripts/run-quantum.sh"
   → http://127.0.0.1:8090

4) Terminal C — Next.js UI
   "$ROOT/scripts/run-frontend.sh"
   → http://127.0.0.1:3000

Stop: Ctrl+C in each terminal (no daemon). Backend in-memory scans are discarded on stop.

Smoke:
   curl -s http://127.0.0.1:8080/api/v1/health
   curl -s http://127.0.0.1:8080/api/v1/ready
   curl -s http://127.0.0.1:8090/health
   "$ROOT/scripts/integration-smoke.sh"

Tests:
   "$ROOT/scripts/test-all.sh"

Demo reminders:
- Keep NEXT_PUBLIC_QGUARD_USE_MOCKS=false
- Quantum results never feed wallet risk/exposure
- Incomplete history must stay visible (coverage status)
- No private keys; exposure ≠ compromise; no exposure ≠ quantum-safe
EOF
