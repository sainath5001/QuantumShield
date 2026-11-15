# Q-Guard local setup guide

Reliable local demonstration of the full stack (Go API + Quantum sidecar + Next.js UI).

## Prerequisites

| Tool | Notes |
| --- | --- |
| Go 1.22+ | `go version` |
| Node.js 20+ / npm | Frontend |
| Python 3.10+ | Quantum engine |
| curl (optional) | Smoke checks |
| Ethereum Mainnet RPC URL | Infura, Alchemy, public endpoint, etc. |
| Etherscan API key | [etherscan.io/apis](https://etherscan.io/apis) |

## 1. Clone and configure secrets

```bash
cd QuantumShield   # repository root
cp .env.example .env
```

Edit `.env` and set at least:

```bash
QGUARD_ETH_RPC_URL=https://…          # Mainnet
QGUARD_ETHERSCAN_API_KEY=…            # your key
QGUARD_CHAIN_ID=1
QGUARD_CORS_ORIGINS=http://localhost:3000
```

**Never** put private keys, seed phrases, or wallet passwords in `.env`.

Frontend public URLs (no secrets):

```bash
cp frontend/.env.example frontend/.env.local
# defaults point at 127.0.0.1:8080 and :8090
```

Or run the bootstrap script:

```bash
./scripts/setup-local.sh
```

## 2. Install dependencies

### Go

```bash
cd backend
go mod download
cd ..
```

### Python (quantum)

```bash
cd quantum-engine
python3 -m venv .venv
source .venv/bin/activate
pip install -U pip
pip install -e ".[dev,api]"
deactivate
cd ..
```

(`scripts/run-quantum.sh` creates the venv and installs if missing.)

### Frontend

```bash
cd frontend
npm install
cd ..
```

(`scripts/run-frontend.sh` runs `npm install` if `node_modules` is missing.)

## 3. Start services (three terminals)

```bash
./scripts/run-local.sh    # prints the checklist

# Terminal 1 — secrets stay here
./scripts/run-api.sh      # http://127.0.0.1:8080

# Terminal 2 — educational only
./scripts/run-quantum.sh  # http://127.0.0.1:8090

# Terminal 3 — UI
./scripts/run-frontend.sh # http://127.0.0.1:3000
```

Open **http://127.0.0.1:3000**. Keep `NEXT_PUBLIC_QGUARD_USE_MOCKS=false` for real scans.

## 4. Smoke checks

```bash
curl -s http://127.0.0.1:8080/api/v1/health | jq .
curl -s http://127.0.0.1:8080/api/v1/ready | jq .
curl -s http://127.0.0.1:8090/health | jq .
./scripts/integration-smoke.sh
```

Demo scan (any public address; results depend on live chain/indexer):

```bash
curl -s -X POST http://127.0.0.1:8080/api/v1/scan \
  -H 'Content-Type: application/json' \
  -d '{"address":"0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0"}' | jq .
```

Poll `GET /api/v1/scan/{id}` until `completed` or `failed`.

## 5. Run tests

```bash
./scripts/test-all.sh
```

Or individually:

```bash
./scripts/test-backend.sh
cd quantum-engine && source .venv/bin/activate && pytest && deactivate
cd frontend && npm test && npm run typecheck && npm run lint
```

## 6. Stop services

There is no process manager. Stop each terminal with **Ctrl+C**.

- Backend: in-memory scans/reports are discarded on exit.
- Frontend / quantum: no durable local state beyond browser `sessionStorage`.

## Demo talking points

1. Scan a public address from `/scanner` — real RPC + Etherscan data.
2. Open Analysis — show risk classification, history coverage, evidence list.
3. Open Report — limitations and disclaimers are mandatory reading.
4. Quantum Lab — run toy Shor (N=15) or Grover; emphasize educational boundary.
5. Explicitly state: **no private keys**, **exposure ≠ compromise**, **no exposure ≠ quantum-safe**.
