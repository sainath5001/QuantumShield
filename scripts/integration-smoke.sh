#!/usr/bin/env bash
# Contract smoke checks against a running local stack (Go :8080, Quantum :8090).
set -euo pipefail

API="${QGUARD_API_BASE_URL:-http://127.0.0.1:8080}"
QAPI="${QGUARD_QUANTUM_API_BASE_URL:-http://127.0.0.1:8090}"

echo "== Go health =="
curl -fsS "$API/api/v1/health" | head -c 400; echo

echo "== Quantum health =="
curl -fsS "$QAPI/health" | head -c 400; echo

echo "== Quantum demos list =="
curl -fsS "$QAPI/demos" | head -c 400; echo

echo "== Quantum Shor (toy) =="
curl -fsS -X POST "$QAPI/demos/shor" \
  -H 'Content-Type: application/json' \
  -d '{"n":15,"a":7,"shots":128,"seed":42}' | head -c 500; echo

echo "== Scan create (requires configured RPC/Etherscan) =="
SCAN_JSON="$(curl -fsS -X POST "$API/api/v1/scan" \
  -H 'Content-Type: application/json' \
  -d '{"address":"0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0"}" || true)"
echo "$SCAN_JSON" | head -c 500; echo

if command -v python3 >/dev/null && [[ -n "$SCAN_JSON" ]]; then
  SCAN_ID="$(python3 -c 'import json,sys; print(json.load(sys.stdin).get("id",""))' <<<"$SCAN_JSON" || true)"
  if [[ -n "$SCAN_ID" ]]; then
    echo "== Poll scan $SCAN_ID (up to ~60s) =="
    for _ in $(seq 1 30); do
      JOB="$(curl -fsS "$API/api/v1/scan/$SCAN_ID")"
      STATUS="$(python3 -c 'import json,sys; print(json.load(sys.stdin).get("status",""))' <<<"$JOB")"
      echo "status=$STATUS"
      if [[ "$STATUS" == "completed" || "$STATUS" == "failed" ]]; then
        echo "$JOB" | head -c 600; echo
        break
      fi
      sleep 2
    done
  fi
fi

echo "Smoke finished."
