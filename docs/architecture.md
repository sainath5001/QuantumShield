# Q-Guard Architecture Proposal

**Product:** Q-Guard — Ethereum Quantum Threat Scanner & Post-Quantum Readiness Platform  
**Phase:** Backend foundation implemented (Prompt 2). Scanning not yet implemented.  
**Workspace inspected:** `/home/hp/QuantumShield` was empty at proposal time; no existing files overwritten.

---

## 1. System overview

Q-Guard is a local-first monorepo with three logical systems:

| System | Stack | Owns |
|---|---|---|
| Backend scanner & API | Go, Gin, go-ethereum | RPC, history, decode, recover, classify, score, REST |
| Quantum education module | Python, Qiskit | Circuits, toy Shor/Grover demos, explanations |
| Frontend | Next.js App Router, TS, Tailwind, shadcn/ui, Recharts, Lucide | UX, scan UI, reports, education pages |

```
┌─────────────────────────────────────────────────────────────┐
│  frontend/ (Next.js) — not scaffolded yet                   │
│  Scan UI · Reports · Education · PQ explainer               │
└───────────────┬─────────────────────────────┬───────────────┘
                │ REST /api/v1/*              │ (later) quantum demos
                ▼                             ▼
┌───────────────────────────────┐   ┌─────────────────────────┐
│  backend/ (Go/Gin)            │   │  quantum-engine/ (Py)   │
│  config · handler · service   │   │  demos · explainers     │
│  blockchain · cryptography …  │   │  (not implemented yet)  │
└───────────┬───────────────────┘   └─────────────────────────┘
            │
     ┌──────┴──────┐
     ▼             ▼
 Ethereum RPC   Etherscan API
```

**Hard separation:** Public-key exposure detection is **classical** (Go). Qiskit is **never** on the evidence path for scans.

---

## 2. Monorepo folder structure (adopted)

Prompt 2 finalized top-level names `backend/`, `frontend/`, `quantum-engine/` (replacing the earlier `services/api`, `apps/web`, `modules/quantum` proposal).

```
QuantumShield/                 # repo root (product name: Q-Guard)
├── docs/
├── backend/
│   ├── cmd/api/               # HTTP server main
│   ├── internal/
│   │   ├── apierror/          # structured JSON errors
│   │   ├── blockchain/        # RPC + address helpers (scan client TBD)
│   │   ├── config/            # env loading + validation
│   │   ├── cryptography/      # pubkey recovery (TBD)
│   │   ├── exposure/          # exposure evidence aggregation (TBD)
│   │   ├── handler/           # Gin handlers
│   │   ├── middleware/        # request ID, log, CORS, recovery, timeout
│   │   ├── quantum/           # classical threat-copy helpers (TBD; not Qiskit)
│   │   ├── report/            # assessment DTOs (TBD)
│   │   ├── repository/        # ScanRepository + memory impl (TBD)
│   │   ├── risk/              # qguard-risk-v1 (TBD)
│   │   ├── server/            # router + http.Server wiring
│   │   ├── service/           # use-cases (meta/health/info today)
│   │   ├── transaction/       # typed tx decode + digests (TBD)
│   │   └── version/           # app / risk-model version strings
│   ├── testdata/
│   └── go.mod
├── frontend/                  # Next.js (placeholder only)
├── quantum-engine/            # Python/Qiskit (README only)
├── scripts/
│   ├── run-api.sh
│   └── test-backend.sh
├── .env.example
├── .gitignore
└── README.md
```

**Naming note:** Repository folder is `QuantumShield`; product brand is **Q-Guard**.

### Implementation status

| Area | Status |
|---|---|
| Config + validation | Implemented (mainnet + Etherscan v2) |
| Health / ready / info | Implemented (`/ready` probes RPC) |
| Middleware + JSON errors | Implemented |
| RPC client + account classification | Implemented |
| Etherscan v2 history + coverage model | Implemented |
| Account / history HTTP endpoints | Implemented |
| Signature recovery / risk / scan jobs | **Not implemented** |
| Frontend / quantum-engine | Reserved directories only |

---

## 3. Backend package responsibilities

| Package | Responsibility | Must not |
|---|---|---|
| `config` | Env loading, validation | Hardcode secrets |
| `blockchain` | Ethereum JSON-RPC, address helpers | Business risk rules; fabricate chain data |
| `transaction` | Typed tx decode + signing digests | Assign risk tiers |
| `cryptography` | Pubkey recover + address verify | Accept private keys; call external APIs |
| `exposure` | Aggregate verified exposure evidence | Invent exposures |
| `risk` | Pure classification (`qguard-risk-v1`) | I/O |
| `service` | Orchestration / use-cases | Embed low-level Gin details unnecessarily |
| `repository` | `ScanRepository` + memory impl | Leak into crypto |
| `report` | Assessment DTOs + disclaimers | Invent tx data |
| `handler` / `middleware` / `server` | REST, CORS, timeouts, errors | Inline crypto |
| `quantum` | Classical threat messaging helpers | Run Qiskit or affect scan evidence |

### 3.1 Repository interfaces (persistence-ready)

```text
ScanRepository
  Create(scan) error
  Get(id) (scan, error)
  Update(scan) error
  ListByAddress(addr, limit) ([]scan, error)   // optional MVP

// MVP: MemoryScanRepository (process-local, ephemeral)
// Later: SQL/other without changing scan/risk packages
```

No PostgreSQL in MVP.

---

## 4. Python module responsibilities

| Area | Contents |
|---|---|
| Shor demos | Small integer factoring / toy discrete-log style circuits; limitation banners |
| Grover demos | Small search spaces; explain quadratic speedup intuition |
| Explainers | Markdown/JSON lesson payloads for frontend consumption (later) |
| PQ notes | ML-KEM / ML-DSA / SLH-DSA summaries; Ethereum migration challenges |
| Tests | Deterministic circuit smoke tests; no “attack Ethereum” claims |

**Exposure to frontend (recommended path):**

1. **MVP:** Static educational content in Next.js + local Python CLI/notebooks for researchers; no coupling required.
2. **Phase 2:** Thin FastAPI (or similar) sidecar under `modules/quantum` serving `/demos/*` read-only JSON + optional circuit diagrams (SVG/base64).
3. **Phase 3:** Next.js BFF proxies to sidecar; never merge demo outputs into Go scan reports.

Do **not** call Qiskit from Go for scanning.

---

## 5. Frontend page structure (App Router)

| Route | Purpose |
|---|---|
| `/` | Product landing: brand, one headline, one CTA to scan (design rules apply at impl) |
| `/scan` | Address input, network select, start scan, progress |
| `/scan/[id]` | Live/completed report: classification, evidence, limitations |
| `/learn` | Hub for quantum & Ethereum crypto education |
| `/learn/shor` | Shor vs ECC explanation + link to demos |
| `/learn/grover` | Grover & security margins |
| `/learn/post-quantum` | NIST PQ + Ethereum migration challenges |
| `/learn/methodology` | What Q-Guard detects / does not detect |
| `/about` | Trust, non-custodial, no keys ever |

**UI libraries:** Tailwind, shadcn/ui, Lucide, Recharts (e.g., coverage breakdown, tx type mix—not fake “hack probability” gauges).

---

## 6. Scan lifecycle

```
1. ACCEPT
   - Validate address checksum/hex
   - Create ScanJob (status=queued), return scan_id

2. CLASSIFY_ACCOUNT
   - eth_getCode → EOA | CONTRACT
   - If CONTRACT → may short-circuit to REQUIRES_REVIEW (still may fetch history for context)

3. FETCH_HISTORY
   - Paginate Etherscan (or configured indexer)
   - Track pages, rate limits, truncation flags
   - Optionally enrich via RPC getTransaction for raw payloads

4. ANALYZE_TRANSACTIONS
   - For each tx: decode type → digest → recover → verify vs from
   - Bucket: analyzed_ok | unsupported | decode_error | verify_mismatch

5. ASSESS
   - risk package applies rules (business-logic.md)
   - Attach evidence, confidence, counts, limitations, risk_model_version

6. PERSIST & COMPLETE
   - store.Update(status=completed|failed)
   - Frontend polls GET /scans/{id}

7. FAIL PATHS
   - Provider errors, invalid address, chain mismatch → structured API errors
   - Partial success allowed with UNKNOWN / coverage flags
```

**Idempotency:** Same address may create multiple scans; MVP does not require deduplication. Optional cache key later.

**Reorgs:** MVP treats provider responses as point-in-time. Document that deep reorgs can theoretically alter history; optional block-hash anchoring post-MVP.

---

## 7. Layering: scanning vs cryptography vs risk

```
scan (orchestration)
  │ uses
  ├─ ethereum + etherscan        → data acquisition
  ├─ txdecode + cryptoanalyze    → cryptographic facts (evidence)
  ├─ wallet                      → account class facts
  └─ risk                        → interpretation (classification)
        └─ report                → API presentation
```

**Rules:**

- Cryptography packages emit **facts** (recovered key hex, match bool), never risk labels.
- Risk package is pure and versioned (`risk_model_version`).
- Scan package is the only orchestrator that talks to I/O + risk together.

---

## 8. Configuration and environment variables

See repository root `.env.example`. Validated at process start by `internal/config`.

| Variable | Component | Purpose |
|---|---|---|
| `QGUARD_APP_ENV` | Go | `development` \| `test` \| `production` |
| `QGUARD_LOG_LEVEL` | Go | `debug` \| `info` \| `warn` \| `error` |
| `QGUARD_SERVER_HOST` | Go | Bind host (default `127.0.0.1`) |
| `QGUARD_SERVER_PORT` | Go | Bind port (default `8080`) |
| `QGUARD_ETH_RPC_URL` | Go | Ethereum JSON-RPC endpoint (**required**) |
| `QGUARD_CHAIN_ID` | Go | Expected chain ID (default `1`) |
| `QGUARD_ETHERSCAN_API_URL` | Go | History API base URL |
| `QGUARD_ETHERSCAN_API_VERSION` | Go | Provider API version label |
| `QGUARD_ETHERSCAN_API_KEY` | Go | API key (**required**) |
| `QGUARD_ETHERSCAN_MAX_PAGES` | Go | Hard cap on pagination |
| `QGUARD_MAX_HISTORY_PAGE_SIZE` | Go | Page size |
| `QGUARD_SCAN_MAX_TXS` | Go | Max txs analyzed per scan |
| `QGUARD_REQUEST_TIMEOUT` | Go | Per-HTTP-request timeout |
| `QGUARD_SCAN_TIMEOUT` | Go | Per-scan deadline (future) |
| `QGUARD_RATE_LIMIT_RPS` | Go | Provider client rate limit |
| `QGUARD_RATE_LIMIT_BURST` | Go | Provider client burst |
| `QGUARD_CORS_ORIGINS` | Go | Allowed frontend origins (CSV) |
| `NEXT_PUBLIC_API_BASE_URL` | Web (later) | Backend base URL |
| `QGUARD_QUANTUM_API_URL` | Web (later) | Python sidecar URL |

**Security:** Never accept private keys via env or API. API keys are for third-party indexers only.

### Local run commands

```bash
cp .env.example .env   # set RPC URL + Etherscan key
./scripts/run-api.sh
./scripts/test-backend.sh
```

---

## 9. External API dependencies

| Dependency | Use | Failure mode |
|---|---|---|
| Ethereum JSON-RPC | Account code, tx bodies, chain ID | Scan fails or degrades; mark incomplete |
| Etherscan API (or Etherscan-compatible) | Indexed outbound/inbound tx lists | Pagination/rate-limit → coverage flags; possibly `UNKNOWN` |

**Assumption (unresolved):** Exact Etherscan plan/endpoints (Free vs Pro) and whether to support multiple networks in MVP (recommend **Ethereum mainnet only** for MVP).

---

## 10. Error-handling strategy

| Class | HTTP | Behavior |
|---|---|---|
| Invalid address | 400 | No scan created |
| Unsupported chain | 400 | Explicit message |
| Upstream rate limit | 200 + incomplete scan **or** 503 on start | Prefer completing with `history_coverage=partial` + `UNKNOWN` if evidence insufficient |
| Upstream timeout | scan `failed` or `completed_partial` | Structured error on job |
| Internal bug | 500 | Log; no fabricated evidence |

Domain errors use typed codes (`INVALID_ADDRESS`, `RPC_UNAVAILABLE`, `HISTORY_INCOMPLETE`, `TX_UNSUPPORTED`, …) in JSON bodies.

---

## 11. Testing strategy

| Layer | Focus |
|---|---|
| Unit (`txdecode`) | Golden signing digests per tx type (legacy EIP-155, 2930, 1559, 4844) |
| Unit (`cryptoanalyze`) | Known vectors: recover → address match; wrong v; malleability edge cases |
| Unit (`risk`) | Table-driven classification; version lock |
| Integration | Mock RPC + mock Etherscan; full scan lifecycle |
| Frontend | Component tests for disclaimer rendering; API client error states |
| Quantum module | Smoke tests for demos; assert limitation strings present |
| Manual | Real mainnet address fixtures with known exposure / fresh unused addresses |

**Forbidden in tests for production paths:** Fabricating scan results that look like live chain data without labeling fixtures as fixtures.

---

## 12. Known technical limitations (architecture-level)

1. In-memory store lost on process restart.
2. History completeness bounded by indexer and caps.
3. Off-chain / cross-chain exposure invisible.
4. Smart accounts need manual/specialized review.
5. Typed tx support must be explicitly maintained as Ethereum evolves.
6. No CRQC timeline; educational demos ≠ cryptanalysis of secp256k1.
7. Single-process MVP; no Redis/Kafka/K8s.
8. Reorg/consistency: point-in-time reads.

---

## 13. Assumptions

1. MVP targets **Ethereum mainnet** primarily.
2. Users paste addresses; **no** wallet connect for scanning.
3. Etherscan-compatible REST is acceptable as the history source.
4. English-only UI for MVP.
5. Python module may ship as CLI-first in MVP.
6. Product name **Q-Guard**; repo may remain `QuantumShield`.

---

## 14. Unresolved decisions (need approval)

1. **Networks:** Mainnet only vs Sepolia for demos?
2. **History direction:** Outgoing txs only (sufficient for sender pubkey recovery) vs full account history?
3. **Etherscan alternatives:** Allow Alchemy/Transfers APIs as fallback?
4. **Scan async model:** Always async+poll vs sync for small histories?
5. **Python exposure:** CLI-only in MVP vs FastAPI sidecar immediately?
6. **Address input:** Enforce EIP-55 checksum or accept any hex?
7. **Rate limiting API:** Per-IP limits on `/scans` in local MVP?
8. **Brand assets / design:** Confirm visual direction before frontend implementation.
9. **OpenAPI:** Generate from Go annotations vs hand-written `packages/openapi`?
10. **License / disclosure:** How strongly to watermark educational demos?

---

## 15. Security product principles (engineering)

- Never implement signing, keystore, or wallet-connect for the scanner path.
- Never claim quantum-safe from negative exposure results.
- Prefer `UNKNOWN` / `REQUIRES_REVIEW` over optimistic `LOWER_EXPOSURE` when evidence is weak.
- Verify recovered keys against sender; discard unverified recoveries as evidence of exposure.
- Version the risk model; show version on every report.
