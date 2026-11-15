# Q-Guard

**Ethereum Quantum Threat Scanner & Post-Quantum Readiness Platform**

Q-Guard analyzes **public** Ethereum Mainnet addresses for **secp256k1 public-key exposure** that would matter under a future cryptographically relevant quantum computer. It never requests private keys, never signs transactions, and never claims a wallet is quantum-safe merely because no exposure was observed in scope.

Repository folder: `QuantumShield`. Product name: **Q-Guard**.

> **Status:** Local MVP — demonstrable for developers, auditors, and quantum researchers.  
> **Not claimed:** Production multi-tenant readiness or public Internet deployment without additional controls.  
> See [`docs/release-checklist.md`](docs/release-checklist.md) and [`docs/security-review.md`](docs/security-review.md).

## Monorepo layout

| Path | Role |
|---|---|
| `backend/` | Go API (Gin, go-ethereum) — scanner & assessment owner |
| `frontend/` | Next.js App Router UI |
| `quantum-engine/` | Python/Qiskit education (toy Shor/Grover + PQC); off evidence path |
| `docs/` | Architecture, threat model, API, crypto, security, setup |
| `scripts/` | Local setup, run, and test helpers |

## Quick start (local demo)

```bash
./scripts/setup-local.sh          # .env files + Go/Python/npm deps
# edit .env — set QGUARD_ETH_RPC_URL and QGUARD_ETHERSCAN_API_KEY

./scripts/run-local.sh            # prints three-terminal instructions
./scripts/run-api.sh              # :8080  (secrets stay here)
./scripts/run-quantum.sh          # :8090  (educational)
./scripts/run-frontend.sh         # :3000  (UI)

./scripts/test-all.sh             # Go + pytest + frontend
```

Full guide: [`docs/local-setup.md`](docs/local-setup.md) · Troubleshooting: [`docs/troubleshooting.md`](docs/troubleshooting.md)

## What the MVP implements

- Mainnet-only account classification (`eth_getCode` / balance)
- Etherscan v2 normal-tx history with explicit **coverage** (never silent completeness)
- secp256k1 recovery for legacy / EIP-2930 / EIP-1559 / EIP-4844 with **sender verification**
- `qguard-risk-v1` classifications + security reports (limitations required)
- Async scan jobs (`POST /scan` → poll → report); in-memory ephemeral store
- Next.js scanner, analysis, reports, Quantum Lab, Learn pages
- Educational Qiskit CLI + FastAPI sidecar (resource-capped demos)

**Not included:** persistent DB, auth, scan cancellation, SSE, non-Mainnet chains, WalletConnect, private-key intake.

## Core product rules

1. **Exposure ≠ compromise** — public-key recovery is not proof of present-day theft.
2. **No exposure ≠ quantum-safe** — scoped history can miss evidence.
3. **Private keys never collected** — API and UI reject key/mnemonic fields by design.
4. **Quantum demos ≠ scan evidence** — Qiskit stays educational and decoupled.
5. **Incomplete data stays visible** — coverage statuses and unknown findings are first-class.

## Documentation index

| Doc | Purpose |
|---|---|
| [Local setup](docs/local-setup.md) | Install, start, stop, demo flow |
| [Troubleshooting](docs/troubleshooting.md) | Common failures |
| [Release checklist](docs/release-checklist.md) | Demo / OSS readiness gates |
| [MVP final report](docs/mvp-final-report.md) | Prompt 10 summary for demos |
| [Architecture](docs/architecture.md) | System design |
| [Business logic](docs/business-logic.md) | Risk semantics |
| [API specification](docs/api-specification.md) | REST contract |
| [Cryptography](docs/cryptography.md) | Digests, recovery, verification |
| [Quantum threat model](docs/quantum-threat-model.md) | Threat framing |
| [Security review](docs/security-review.md) | Prompt 9 findings + remediations |
| [Integration](docs/integration.md) | Frontend ↔ APIs |
| [Development roadmap](docs/development-roadmap.md) | Phases & next milestones |
| [Backend README](backend/README.md) | Go API |
| [Frontend README](frontend/README.md) | Next.js UI |
| [Quantum engine README](quantum-engine/README.md) | Qiskit module |

## API endpoints (summary)

| Method | Path |
|---|---|
| GET | `/api/v1/health`, `/ready`, `/info` |
| GET | `/api/v1/accounts/:address` (+ `/transactions`, `/exposure`) |
| POST | `/api/v1/scan` |
| GET | `/api/v1/scan/:id`, `/api/v1/reports/:id` |
| GET | `/api/v1/wallet/:address` (+ `/exposure`, `/risk`) |

## Frontend pages

| Route | Purpose |
|---|---|
| `/` | Overview + session stats + quick scan |
| `/scanner` | Full scanner with progress polling |
| `/wallet/[address]` | Analysis (risk, exposure, evidence, coverage) |
| `/reports`, `/reports/[id]` | Session report index + detail/export |
| `/quantum` | Educational Quantum Lab |
| `/learn` | Methodology & threat education |

## Quantum experiments

| Experiment | Scope |
|---|---|
| Fundamentals | Superposition, Bell, gate explainers |
| Shor (educational) | Period finding for **N=15** (not secp256k1) |
| Grover (educational) | ≤4 qubits unstructured search |
| PQC explainers | ML-DSA, ML-KEM, SLH-DSA overview |

## Security notes

- Secrets (RPC URL, Etherscan key) live only in root `.env` for the Go process.
- Browser env may expose only `NEXT_PUBLIC_*` base URLs.
- Local bind defaults to `127.0.0.1`. Do not expose unauthenticated APIs publicly.
- See security review for remediated medium findings (address index, rate limits, error sanitization).
