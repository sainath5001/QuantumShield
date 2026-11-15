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

### Phase 1c — Signature recovery + exposure ✅ (Prompt 4)

- `internal/cryptography`: canonical digests + Ecrecover + sender verification.
- `internal/exposure`: wallet-level evidence aggregation; contract gating.
- Endpoint: `GET /api/v1/accounts/:address/exposure`.
- Fixture tests for legacy / 2930 / 1559 / 4844.
- **Deferred:** risk scoring, scan jobs, frontend.

### Phase 1d — Risk engine + reports + scan workflow ✅ (Prompt 5)

- Deterministic `qguard-risk-v1` engine (`internal/risk`).
- Security report builder (`internal/report`).
- In-memory scan/report repository with TTL/capacity.
- Endpoints: `POST /scan`, `GET /scan/:id`, `GET /reports/:id`, wallet aliases.
- Full Go test coverage for classifications and scan lifecycle.

### Phase 2 — Frontend MVP (next)

- Next.js App Router scaffold in `frontend/`.
- Scan flow + report page + methodology/disclaimers.

### Phase 2 — Frontend MVP ✅

- Next.js App Router app in `frontend/` (overview, scanner, analysis, reports, quantum lab, learn).
- Typed API client against Go `/api/v1`; session-local scan index; no fake live metrics.

### Phase 3 — Quantum education module ✅

- Python package in `quantum-engine/` with toy Shor/Grover demos, fundamentals, PQC education, CLI, pytest.
- Optional FastAPI sidecar later (not required; interface is functions + CLI).

### Phase 3.5 — Local integration ✅

- Frontend ↔ Go scan polling (2s / 180s timeout).
- Quantum FastAPI sidecar + Quantum Lab experiment runner.
- `scripts/run-local.sh`, `run-quantum.sh`, `run-frontend.sh`, `integration-smoke.sh`.

### Phase 4 — Security review & local release prep ✅ (Prompts 9–10)

- Security/reliability review (`docs/security-review.md`); medium findings remediated.
- Inbound rate limit, concurrent-scan cap, error sanitization, completed-only address index.
- Docs: local setup, troubleshooting, release checklist; `setup-local.sh`, `test-all.sh`.
- Root/backend/frontend/quantum READMEs aligned for demo.

### Phase 5 — Hardening & expansion (next)

- OpenAPI artifact; durable storage + auth for shared demos.
- Dependency upgrades (Next/zod) after regression suite.
- Smart-account / AA playbooks; optional SSE progress; CI on `test-all.sh`.
- Broader history sources with coverage honesty preserved.

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
| M0 | Config + health/ready/info + middleware tests | ✅ |
| M1 | Golden digest tests for supported types | ✅ |
| M2 | Recovery+verify fixtures | ✅ |
| M3 | Risk table tests for all four labels | ✅ |
| M4 | End-to-end mocked scan → API assessment | ✅ |
| M5 | Frontend renders disclaimers for every classification | ✅ |
| M6 | Quantum demos refuse overselling in README/tests | ✅ |
| M7 | Security review + local release docs | ✅ Prompt 9–10 |

---

## 5. Suggested next development milestones

1. Generate OpenAPI from Go types; keep docs/API clients in sync.
2. Add durable scan/report storage and authentication for shared demos.
3. Upgrade frontend dependency pins after full regression (`npm audit` L-06).
4. Enrich contract / AA “manual review” guidance without false EOA claims.
5. Optional SSE for scan progress; keep polling as fallback.
6. CI: run `scripts/test-all.sh` on every PR.

---

## 6. Success criteria for a credible local MVP

- A security-literate reviewer can understand evidence vs assumptions.
- Classification is reproducible for the same fixture set and risk model version.
- No private key surfaces exist in API or UI.
- Limitations are impossible to miss on the report page.
- Quantum demos are clearly educational and decoupled.
- Local startup is documented and scripted; incomplete history is never sold as complete.
- Production / public Internet readiness is **not** claimed without auth + durable storage + edge controls.
