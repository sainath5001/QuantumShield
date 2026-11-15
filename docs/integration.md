# Q-Guard local integration (Prompt 8)

## Architecture boundary

| Component | Port | Role |
|---|---|---|
| Go backend | `8080` | Wallet scan, exposure, risk, reports (holds RPC + Etherscan secrets) |
| Quantum FastAPI | `8090` | Educational Shor/Grover/PQC demos only |
| Next.js frontend | `3000` | UI; public URLs only |

Quantum outputs **must never** become scan evidence or risk inputs.

## Startup

```bash
./scripts/run-local.sh          # prints the three-terminal instructions
./scripts/run-api.sh            # terminal 1
./scripts/run-quantum.sh        # terminal 2
./scripts/run-frontend.sh       # terminal 3
./scripts/integration-smoke.sh  # optional contract smoke (services must be up)
```

## Scan lifecycle (frontend)

1. Client validates `0x` + 40 hex.
2. `POST /api/v1/scan` → job `status=queued`.
3. Poll `GET /api/v1/scan/:id` every **2s** (timeout **180s**).
4. Terminal statuses: `completed` | `failed`.
5. On success, open `/wallet/:address?scan=…&report=…` and/or `/reports/:id`.

**Not supported:** scan cancellation API.

## Resolved / known integration issues

| Issue | Resolution |
|---|---|
| Docs showed `assessment` / `scan_id` on jobs | Frontend uses runtime Go fields (`id`, `exposure`, `risk`, `report_id`) |
| Quantum was CLI-only | Thin FastAPI sidecar wraps `qguard_quantum.interface` |
| Aggressive polling risk | Frontend polls every 2s with hard timeout |
| Missing report after restart | UI explains in-memory backend storage |
| API failure → fake success | Explicitly forbidden; errors surface via `humanizeApiError` |
| Secrets in browser | RPC/Etherscan keys only in root `.env` for Go |

## Environment

See root `.env.example` and `frontend/.env.example`.
