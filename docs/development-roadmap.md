# Q-Guard Development Roadmap

---

## 1. Answers to mandatory architecture questions

(Unchanged from Prompt 1 — see git history / prior revision for the full Q&A.)

**Summary:** Q-Guard classically detects verified secp256k1 public-key exposure within a scoped on-chain history. It cannot prove quantum-safety, detect classical theft, or run real 256-bit Shor attacks. Backend separates acquisition → crypto facts → versioned risk → API. Python/Qiskit stays educational and off the evidence path.

---

## 2. Implementation phases

### Phase 0 — Design ✅

- Documentation set under `docs/`.

### Phase 1a — Backend foundation ✅ (Prompt 2)

- Monorepo folders: `backend/`, `frontend/`, `quantum-engine/`, `docs/`, `scripts/`.
- Go module with Gin, go-ethereum, structured logging, env config.
- Endpoints: `GET /api/v1/health`, `/ready`, `/info`.
- Middleware: request ID, logging, recovery, CORS, timeouts.
- Tests for config, health/ready, middleware, errors.
- `.env.example`, `README.md`, scripts.

### Phase 1b — RPC + Etherscan history ✅ (Prompt 3)

- Mainnet-only RPC client with chain ID enforcement.
- Account classification via `eth_getCode` / balance.
- Etherscan API **v2** `txlist` client: pagination, rate limits, retries, coverage model.
- Endpoints: `GET /api/v1/accounts/:address`, `.../transactions`.
- Mocked unit tests + documented fixtures.
- **Deferred:** signature recovery, risk scoring, scan jobs, frontend, Qiskit.

### Phase 1c — Scanner core (next)

- In-memory `ScanRepository`.
- Tx decode + recovery + verification.
- Risk engine `qguard-risk-v1`.
- `POST /api/v1/scans` + `GET /api/v1/scans/{id}`.
- Unit/integration tests with fixtures.

### Phase 2 — Frontend MVP

- Next.js App Router scaffold in `frontend/`.
- Scan flow + report page + methodology/disclaimers.

### Phase 3 — Quantum education module

- Python package in `quantum-engine/` with toy Shor/Grover demos.
- Optional FastAPI sidecar later.

### Phase 4 — Hardening & expansion

- OpenAPI package, optional persistence, multi-network, smart-account playbooks.

---

## 3. Prompt 2 completion notes

**Design choices:**

- Adopted Prompt 2 top-level layout (`backend/`, `quantum-engine/`, `frontend/`) over the earlier `services/api` / `modules/quantum` / `apps/web` names.
- Mapped architecture concerns onto Prompt 2 packages (`blockchain`, `transaction`, `cryptography`, …).
- Reserved unimplemented scanner packages with `doc.go` only — no fake scan APIs.
- `/ready` reports dependencies as `configured`, never as connectivity-verified.
- `/info` sets `address_scan_implemented: false` and all tx types `implemented: false`.

**Exact local commands:**

```bash
cp .env.example .env
# edit QGUARD_ETH_RPC_URL and QGUARD_ETHERSCAN_API_KEY

./scripts/run-api.sh
# API: http://127.0.0.1:8080

./scripts/test-backend.sh
# or: cd backend && go test ./...
```

---

## 4. Testing milestones

| Milestone | Gate | Status |
|---|---|---|
| M0 | Config + health/ready/info + middleware tests | ✅ Prompt 2 |
| M1 | Golden digest tests for supported types | Pending |
| M2 | Recovery+verify fixtures | Pending |
| M3 | Risk table tests for all four labels | Pending |
| M4 | End-to-end mocked scan → API assessment | Pending |
| M5 | Frontend renders disclaimers for every classification | Pending |
| M6 | Quantum demos refuse overselling in README/tests | Pending |

---

## 5. Remaining tasks after Prompt 2

1. Implement scan lifecycle + repository (memory).
2. Implement blockchain RPC + Etherscan history clients.
3. Implement transaction decode + cryptography recovery.
4. Implement exposure aggregation + risk + report.
5. Add scan REST endpoints per `api-specification.md`.
6. Scaffold frontend and quantum-engine when authorized.

---

## 6. Success criteria for a credible MVP

- A security-literate reviewer can understand evidence vs assumptions.
- Classification is reproducible for the same fixture set and risk model version.
- No private key surfaces exist in API or UI.
- Limitations are impossible to miss on the report page.
- Quantum demos are clearly educational and decoupled.
