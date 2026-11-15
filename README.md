# Q-Guard

Ethereum Quantum Threat Scanner & Post-Quantum Readiness Platform.

Q-Guard analyzes public Ethereum addresses for **secp256k1 public-key exposure** relevant to a hypothetical future cryptographically relevant quantum computer. It never requests private keys, never signs transactions, and never claims a wallet is quantum-safe merely because no exposure was observed.

Repository folder: `QuantumShield`. Product name: **Q-Guard**.

## Monorepo layout

| Path | Role |
|---|---|
| `backend/` | Go API (Gin, go-ethereum) — scanner & assessment owner |
| `frontend/` | Next.js App Router UI (scanner, analysis, reports, quantum lab, learn) |
| `quantum-engine/` | Python/Qiskit education module (toy Shor/Grover + PQC docs; off evidence path) |
| `docs/` | Architecture, threat model, API, cryptography |
| `scripts/` | Local helper scripts |

## Current milestone: RPC + history integration

Implemented:

- Configuration from environment (Mainnet + Etherscan API v2)
- `GET /api/v1/health` — liveness + version
- `GET /api/v1/ready` — probes Ethereum RPC chain ID; does **not** claim Etherscan is healthy
- `GET /api/v1/info` — capabilities (`history_retrieval` true; scan/recovery still false)
- `GET /api/v1/accounts/:address` — balance + code-based classification signal
- `GET /api/v1/accounts/:address/transactions` — Etherscan normal tx history + coverage
- Middleware: request ID, structured logging, recovery, CORS, request timeout

Also implemented (Prompt 4):

- secp256k1 public-key recovery for legacy / EIP-2930 / EIP-1559 / EIP-4844
- `GET /api/v1/accounts/:address/exposure` — verified exposure evidence (not a risk score)

Also implemented (Prompt 5):

- `qguard-risk-v1` risk engine + security reports
- `POST /api/v1/scan`, `GET /api/v1/scan/:id`, `GET /api/v1/reports/:id`
- `GET /api/v1/wallet/:address` (+ `/exposure`, `/risk`)
- Thread-safe in-memory scan/report store (ephemeral)

Also implemented (Prompt 6):

- `quantum-engine/` educational Qiskit package (`qguard-quantum` CLI): fundamentals, N=15 Shor toy, tiny Grover, PQC explainers
- Independent of scan evidence — see `quantum-engine/README.md`

Also implemented (Prompt 7–8):

- Next.js frontend against `/api/v1` (see `frontend/README.md`)
- Quantum FastAPI sidecar on `:8090` + local integration scripts (`docs/integration.md`)

Not implemented yet: persistent database; scan cancellation; SSE progress.

## Prerequisites

- Go 1.22+ (developed with Go 1.27)
- Ethereum JSON-RPC URL and Etherscan API key for configuration validation (connectivity unused until scanning lands)

## Configure

```bash
cp .env.example .env
# edit .env — set QGUARD_ETH_RPC_URL and QGUARD_ETHERSCAN_API_KEY
```

## Run the local stack

```bash
./scripts/run-local.sh          # prints three-terminal instructions
./scripts/run-api.sh            # Go API :8080 (needs .env with RPC + Etherscan)
./scripts/run-quantum.sh        # Quantum education API :8090
./scripts/run-frontend.sh       # Next.js :3000
./scripts/integration-smoke.sh  # optional contract smoke when services are up
```

Server listens on `QGUARD_SERVER_HOST:QGUARD_SERVER_PORT` (default `127.0.0.1:8080`).

### Smoke checks

```bash
curl -s http://127.0.0.1:8080/api/v1/health | jq
curl -s http://127.0.0.1:8080/api/v1/ready | jq
curl -s http://127.0.0.1:8080/api/v1/info | jq
curl -s http://127.0.0.1:8080/api/v1/accounts/0x0000000000000000000000000000000000000000 | jq
curl -s "http://127.0.0.1:8080/api/v1/accounts/0x0000000000000000000000000000000000000000/transactions" | jq
```

## Test

```bash
./scripts/test-backend.sh
# or
cd backend && go test ./...
```

## Documentation

- [Architecture](docs/architecture.md)
- [Business logic](docs/business-logic.md)
- [API specification](docs/api-specification.md)
- [Quantum threat model](docs/quantum-threat-model.md)
- [Cryptography](docs/cryptography.md)
- [Development roadmap](docs/development-roadmap.md)

## Security notes

- Never put private keys, seed phrases, or wallet passwords in `.env` or API requests.
- Do not treat `ready: true` as proof that Ethereum RPC or Etherscan are reachable.
- Educational quantum demos (future `quantum-engine/`) are not evidence for scan results.
