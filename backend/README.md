# Q-Guard Backend (Go API)

Classical Ethereum Mainnet scanner: account classification, Etherscan history, secp256k1 public-key recovery with sender verification, `qguard-risk-v1` risk labels, and ephemeral security reports.

## Role in the monorepo

| Owns | Does not own |
| --- | --- |
| RPC + Etherscan secrets | Browser UI |
| Exposure evidence + risk | Qiskit simulations |
| `/api/v1/*` REST contract | Persistent multi-tenant storage |

Educational quantum demos live in `../quantum-engine/` and must never feed scan evidence.

## Requirements

- Go 1.22+ (repo developed against a newer toolchain)
- Ethereum Mainnet JSON-RPC URL
- Etherscan API key (v2)

## Configure

```bash
cp ../.env.example ../.env
# set QGUARD_ETH_RPC_URL and QGUARD_ETHERSCAN_API_KEY
```

Important variables (full list in root `.env.example`):

| Variable | Purpose |
| --- | --- |
| `QGUARD_ETH_RPC_URL` | Mainnet RPC |
| `QGUARD_ETHERSCAN_API_KEY` | History provider |
| `QGUARD_CHAIN_ID` | Must be `1` |
| `QGUARD_CORS_ORIGINS` | Frontend origin allow-list |
| `QGUARD_SCAN_TIMEOUT` | Per-scan wall clock |
| `QGUARD_SCAN_MAX_TXS` / `QGUARD_ETHERSCAN_MAX_PAGES` | History caps |
| `QGUARD_HTTP_RATE_LIMIT_*` | Inbound HTTP rate limit |
| `QGUARD_RATE_LIMIT_*` | Outbound Etherscan token bucket |

## Run

```bash
../scripts/run-api.sh
# or: go run ./cmd/api
```

Default listen: `http://127.0.0.1:8080`.

## Package map

```text
cmd/api/                 process entry
internal/config/         env load + validation
internal/server/         Gin engine + timeouts
internal/handler/        HTTP adapters
internal/service/        meta, account, scan orchestration
internal/blockchain/     RPC client, address parse, account classify
internal/etherscan/      v2 txlist + pagination + coverage
internal/history/        fetch + optional block-hash checks
internal/cryptography/   canonical digests + Ecrecover + sender verify
internal/exposure/       wallet evidence aggregation
internal/risk/           qguard-risk-v1
internal/report/         security report builder
internal/repository/     in-memory scan/report store
internal/middleware/     request ID, log, CORS, rate limit, body size, timeout
```

## API surface (MVP)

| Method | Path | Notes |
| --- | --- | --- |
| GET | `/api/v1/health` | Liveness |
| GET | `/api/v1/ready` | RPC probe; Etherscan = configured only |
| GET | `/api/v1/info` | Capabilities + limitations |
| GET | `/api/v1/accounts/:address` | Balance + code signal |
| GET | `/api/v1/accounts/:address/transactions` | History + coverage |
| GET | `/api/v1/accounts/:address/exposure` | Live exposure analysis |
| POST | `/api/v1/scan` | Async scan job |
| GET | `/api/v1/scan/:id` | Job status + nested results |
| GET | `/api/v1/reports/:id` | Security report |
| GET | `/api/v1/wallet/:address` | Account alias |
| GET | `/api/v1/wallet/:address/exposure` | Prefer latest completed scan |
| GET | `/api/v1/wallet/:address/risk` | Latest completed risk |

See [`../docs/api-specification.md`](../docs/api-specification.md). Runtime scan jobs use field `id` (not `scan_id` on the job object).

## Tests

```bash
../scripts/test-backend.sh
# or
go test ./...
go test -race ./...
go vet ./...
```

## Security notes

- Never accept private keys, mnemonics, or seed phrases.
- Confirmed exposure requires recovery **and** sender-address verification.
- Incomplete history is labeled (`partial`, `rate_limited`, …) — never silently “complete.”
- Absence of exposure is **not** a quantum-safe claim.
- Scan/report store is in-memory (lost on restart); suitable for local demo only.
