# Q-Guard

Ethereum Quantum Threat Scanner & Post-Quantum Readiness Platform.

Q-Guard analyzes public Ethereum addresses for **secp256k1 public-key exposure** relevant to a hypothetical future cryptographically relevant quantum computer. It never requests private keys, never signs transactions, and never claims a wallet is quantum-safe merely because no exposure was observed.

Repository folder: `QuantumShield`. Product name: **Q-Guard**.

## Monorepo layout

| Path | Role |
|---|---|
| `backend/` | Go API (Gin, go-ethereum) — scanner & assessment owner |
| `frontend/` | Next.js UI (not scaffolded in foundation milestone) |
| `quantum-engine/` | Python/Qiskit education module (not implemented yet) |
| `docs/` | Architecture, threat model, API, cryptography |
| `scripts/` | Local helper scripts |

## Current milestone: backend foundation

Implemented:

- Configuration from environment (validated at startup)
- `GET /api/v1/health` — liveness + version
- `GET /api/v1/ready` — config readiness (**does not** probe RPC/Etherscan connectivity)
- `GET /api/v1/info` — chain/capability metadata (scan features marked not implemented)
- Middleware: request ID, structured logging, recovery, CORS, request timeout
- Structured JSON errors

Not implemented yet: address scanning, transaction decode, pubkey recovery, risk scoring, frontend, Qiskit demos.

## Prerequisites

- Go 1.22+ (developed with Go 1.27)
- Ethereum JSON-RPC URL and Etherscan API key for configuration validation (connectivity unused until scanning lands)

## Configure

```bash
cp .env.example .env
# edit .env — set QGUARD_ETH_RPC_URL and QGUARD_ETHERSCAN_API_KEY
```

## Run the API locally

```bash
# from repository root
./scripts/run-api.sh
```

Or:

```bash
cd backend
# ensure env vars are exported or ../.env exists
go run ./cmd/api
```

Server listens on `QGUARD_SERVER_HOST:QGUARD_SERVER_PORT` (default `127.0.0.1:8080`).

### Smoke checks

```bash
curl -s http://127.0.0.1:8080/api/v1/health | jq
curl -s http://127.0.0.1:8080/api/v1/ready | jq
curl -s http://127.0.0.1:8080/api/v1/info | jq
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
