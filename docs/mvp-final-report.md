# Q-Guard MVP final report (Prompt 10)

**Date:** 2026-11-15  
**Audience:** Blockchain developers, security auditors, quantum-computing researchers  
**Claim level:** Well-engineered **local demonstration MVP** — not production SaaS readiness.

---

## Final architecture summary

Three processes, hard separation of concerns:

1. **Go API (`:8080`)** — classical Mainnet scan owner: RPC, Etherscan history, ECDSA recovery with sender verification, `qguard-risk-v1`, ephemeral reports. Holds all chain secrets.
2. **Quantum FastAPI (`:8090`)** — educational Qiskit demos only; never feeds exposure/risk.
3. **Next.js UI (`:3000`)** — scanner, analysis, reports, Quantum Lab, Learn; public base URLs only.

Evidence path: address → classify → history (coverage) → recover+verify → expose → risk → report.  
Qiskit is explicitly off that path.

---

## Implemented features

- Env-validated Mainnet-only configuration
- Account classification (`likely_eoa` / `contract_account` / `unknown`)
- Etherscan v2 normal-tx pagination with coverage honesty
- Public-key recovery: legacy, EIP-2930, EIP-1559, EIP-4844
- Confirmed exposure only after sender verification
- Async scan jobs + security reports (in-memory)
- Concurrent-scan cap, inbound HTTP rate limit, body size limit, sanitized errors
- Next.js dashboard with empty states, live polling, session scan index
- Quantum fundamentals / Shor(N=15) / Grover(≤4 qubits) / PQC explainers

---

## API endpoint list

| Method | Path |
| --- | --- |
| GET | `/api/v1/health` |
| GET | `/api/v1/ready` |
| GET | `/api/v1/info` |
| GET | `/api/v1/accounts/:address` |
| GET | `/api/v1/accounts/:address/transactions` |
| GET | `/api/v1/accounts/:address/exposure` |
| POST | `/api/v1/scan` |
| GET | `/api/v1/scan/:id` |
| GET | `/api/v1/reports/:id` |
| GET | `/api/v1/wallet/:address` |
| GET | `/api/v1/wallet/:address/exposure` |
| GET | `/api/v1/wallet/:address/risk` |

Quantum sidecar: `GET /health`, `GET /info`, `POST /demos/shor`, `POST /demos/grover`, `GET /explainers/post-quantum` (+ related explainer routes as implemented).

---

## Frontend page list

| Route | Purpose |
| --- | --- |
| `/` | Overview, session stats, quick scan |
| `/scanner` | Scanner + progress |
| `/wallet/[address]` | Analysis |
| `/reports` | Session report index |
| `/reports/[id]` | Full report + export |
| `/quantum` | Quantum Lab |
| `/learn` | Education |

---

## Quantum experiments available

| ID | Description |
| --- | --- |
| Fundamentals | Superposition, Bell, gate explainers |
| `shor` | Educational period finding for N=15 |
| `grover` | Tiny unstructured search (≤4 qubits) |
| PQC | ML-DSA / ML-KEM / SLH-DSA explainers |

---

## Test results (Prompt 10 verification)

```text
./scripts/test-all.sh → test-all: OK

Go:     go vet ./... ; go test ./... ; go test -race ./...   (pass)
Python: pytest                                               (26 passed)
Frontend: npm test (14) ; typecheck ; lint                   (pass)
```

`golangci-lint` is not configured in-repo. Live provider smoke: `./scripts/integration-smoke.sh` (requires running services + credentials).

---

## Security-review summary

Confirmed medium issues remediated (Prompt 9):

- Latest-by-address only tracks **completed** scans
- Concurrent-scan semaphore + inbound IP rate limit + body size cap
- Client-facing error sanitization (API, scan jobs, ready, quantum 500s)

Accepted limitations: no auth, ephemeral storage, indexer trust, npm advisory follow-up (L-06).  
Details: [`security-review.md`](./security-review.md).

---

## Local startup instructions

```bash
./scripts/setup-local.sh
# edit .env (RPC + Etherscan key)
./scripts/run-api.sh
./scripts/run-quantum.sh
./scripts/run-frontend.sh
```

See [`local-setup.md`](./local-setup.md), [`troubleshooting.md`](./troubleshooting.md), [`release-checklist.md`](./release-checklist.md).

---

## Known issues

- In-memory scans/reports lost on API restart
- No scan cancellation / SSE
- Session report index is browser-local only
- History limited to normal txs within page/tx caps
- Contract / AA wallets → manual review, not EOA confirmed exposure
- Frontend dependency advisories on pinned Next/zod (track upgrade)
- Unauthenticated localhost API — do not expose publicly

---

## Suggested next milestones

1. OpenAPI generation + CI (`test-all.sh`)
2. Durable storage + authentication
3. Dependency upgrades after regression
4. Smart-account playbooks
5. Optional SSE progress; broader history sources with coverage honesty

---

## Readiness statement

**Local MVP demonstration: ready** when release checklist items are satisfied and live RPC/Etherscan credentials work.

**Production / public Internet deployment: not ready** — authentication, durable storage, and edge rate limiting remain unresolved on purpose for this milestone.
