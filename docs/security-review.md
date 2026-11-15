# Q-Guard Security & Reliability Review

**Date:** 2026-11-15  
**Scope:** Prompt 9 — Go API, cryptography, blockchain data honesty, Next.js frontend, Python/Qiskit educational sidecar, API hardening.  
**Reviewer role:** Application security / Ethereum protocol / QA automation.

**Verdict:** Confirmed medium-severity issues found during review were remediated before MVP readiness. Cryptographic recovery path correctly requires sender-address verification. History coverage is labeled explicitly (never silently treated as complete). Residual items below are design limitations or lower-severity improvements, not hidden critical flaws.

---

## Executive summary

| Area | Assessment |
|------|------------|
| ECDSA recovery / digests / EIP-155 | Sound; sender verified before confirmed exposure |
| History / Etherscan honesty | Coverage statuses + truncation flags present |
| API DoS / concurrency | Inbound IP rate limit + concurrent-scan semaphore + store capacity |
| Error leakage | Client-facing messages sanitized (handlers, scan job errors, `/ready`, quantum 500s) |
| Frontend XSS / secrets | React text rendering; only `NEXT_PUBLIC_*` URLs; no private keys |
| Quantum sidecar | Input caps (qubits/shots/N); educational only; CORS localhost |

---

## Confirmed vulnerabilities (remediated)

### V-01 — Latest-by-address indexed incomplete scans

| Field | Value |
|-------|--------|
| **Severity** | Medium |
| **Affected component** | `backend/internal/repository` (`byAddr`) |
| **Status** | **Fixed** |

**Description:** `CreateScan` / `UpdateScan` always wrote the latest scan ID into `byAddr`, including `queued` / `running` / `failed` jobs. Wallet “latest exposure/risk” lookups could therefore miss a prior completed assessment while a new job was in flight, or surface non-completed state incorrectly depending on callers.

**Reproduction:**
1. Complete a scan for address A (status `completed`).
2. Start a second scan for A (status `queued`/`running`).
3. Call `GET /api/v1/wallet/:address/risk` (or exposure via completed-scan path).
4. Before fix: `LatestScanByAddress` returned the in-flight job → risk not found / incomplete evidence path.

**Impact:** Incomplete or missing risk/exposure presented for wallets that already had a finished assessment; undermines reliability of “latest evidence.”

**Remediation:** `byAddr` only updates when status is `ScanCompleted`. `LatestScanByAddress` requires completed status. Covered by `repository_test.go`.

---

### V-02 — Unauthenticated scan resource exhaustion

| Field | Value |
|-------|--------|
| **Severity** | Medium |
| **Affected component** | `ScanService`, HTTP middleware, config |
| **Status** | **Fixed** (MVP mitigations) |

**Description:** Anyone who can reach the API could enqueue unlimited concurrent scans (each hitting RPC + Etherscan within scan timeout), exhausting provider quotas, memory (scan payloads), and CPU. Outbound Etherscan token-bucket alone did not bound inbound fan-out.

**Reproduction:**
1. With API listening and credentials configured, script many parallel `POST /api/v1/scan` with distinct valid addresses.
2. Observe unbounded goroutines / provider pressure before mitigations.

**Impact:** Denial of service against the local API and quota burn on shared RPC/Etherscan keys.

**Remediation:**
- Process-wide concurrent-scan semaphore (`MaxConcurrent` default 4) → `429 RATE_LIMITED`.
- Inbound per-IP HTTP rate limiter (`QGUARD_HTTP_RATE_LIMIT_RPS` / `BURST`, default 20/40).
- Existing in-memory capacity (`MaxScans`/`MaxReports`) and request timeouts retained.
- 1 MiB request body limit middleware.

**Remaining residual:** Still no authentication; suitable for local/trusted MVP, not public Internet without an edge control plane.

---

### V-03 — Upstream / internal error string leakage

| Field | Value |
|-------|--------|
| **Severity** | Low |
| **Affected component** | `handler.mapDomainError`, `ScanService.run`, `MetaService.Ready`, quantum `http_api.py` |
| **Status** | **Fixed** |

**Description:** Raw RPC, Etherscan, and exception strings were returned in JSON error bodies, failed scan `error` fields, readiness check messages, and FastAPI 500 `detail`.

**Reproduction:** Trigger RPC failure or history provider error; inspect API JSON / scan job / `/ready` / quantum 500 response for internal exception text.

**Impact:** Information disclosure useful for reconnaissance (URLs, provider messages, stack-ish detail).

**Remediation:** Stable client messages; full errors remain in server logs. Quantum 500 returns `"Experiment failed."` Validation errors (400) still return validation detail by design.

---

## Design improvements & limitations (not treated as confirmed vulns)

### L-01 — No authentication / authorization

| Field | Value |
|-------|--------|
| **Severity** | Limitation (accepted for local MVP) |
| **Component** | API + quantum sidecar |
| **Status** | Open (by design for MVP) |

Any client that can reach the bind address can start scans and fetch scan/report IDs. UUIDs reduce casual enumeration but are not access control. Bind default `127.0.0.1`.

---

### L-02 — In-memory ephemeral storage

| Field | Value |
|-------|--------|
| **Severity** | Limitation |
| **Component** | `repository.Memory` |
| **Status** | Documented |

Restart clears all scans/reports. TTL eviction (24h in `main`) can drop evidence. Clients must not treat storage as durable evidence vault.

---

### L-03 — Indexer trust & reorg surface

| Field | Value |
|-------|--------|
| **Severity** | Limitation |
| **Component** | Etherscan client, `history.Fetcher` |
| **Status** | Mitigated / documented |

History depends on Etherscan v2. Coverage statuses (`complete_within_scope`, `partial`, `rate_limited`, `failed`, …), pagination caps, duplicate drops, and optional RPC block-hash spot checks reduce silent false completeness. Full reorg resistance is not claimed; hash mismatches force `partial`.

---

### L-04 — Unsupported / typed-tx edge cases

| Field | Value |
|-------|--------|
| **Severity** | Limitation |
| **Component** | `cryptography`, exposure |
| **Status** | Handled explicitly |

Unsupported envelopes return `unsupported_type` recovery status and are counted in coverage — not confirmed exposure. Set-code / future types remain out of scope until implemented.

---

### L-05 — Quantum educational API resource use

| Field | Value |
|-------|--------|
| **Severity** | Low / improvement |
| **Component** | `quantum-engine` |
| **Status** | Partially mitigated |

Pydantic + validation cap qubits (≤4), shots (≤4096), Shor N educational set, counting qubits 4|8. No inbound rate limit on `:8090` beyond process isolation. Acceptable for localhost educational use; do not expose publicly without limits.

---

### L-06 — Frontend dependency advisories

| Field | Value |
|-------|--------|
| **Severity** | Medium (supply-chain; not app-logic RCE confirmed) |
| **Component** | `frontend` npm tree (`next@15.1.0` → `postcss`/`sharp`, `zod`) |
| **Status** | Open — track upgrades |

`npm audit --omit=dev` reported 4 issues (3 high, 1 critical) primarily via Next/sharp/postcss/zod. Fixes via `npm audit fix --force` pull Next/zod major bumps outside current pin. **Not force-upgraded in this review** to avoid unverified breaking changes mid-MVP; schedule a deliberate dependency bump + retest.

Exploitability for this UI (text-only React, no attacker-controlled CSS pipeline in production request path) is limited, but pins should still be refreshed.

---

### L-07 — Client IP rate-limit trust

| Field | Value |
|-------|--------|
| **Severity** | Improvement |
| **Component** | `middleware.IPRateLimiter` + Gin `ClientIP` |
| **Status** | Open |

Behind misconfigured reverse proxies, `X-Forwarded-For` can dilute or spoof IP buckets. Prefer edge rate limiting in production.

---

### L-08 — Scan/report ID knowledge access

| Field | Value |
|-------|--------|
| **Severity** | Limitation |
| **Component** | `GET /scan/:id`, `GET /reports/:id` |
| **Status** | Accepted for MVP |

Knowing a UUID grants read access to that job/report (including recovered public keys for confirmed evidence). Treat IDs as sensitive capability URLs.

---

## Cryptographic correctness review

| Check | Result |
|-------|--------|
| Signing digest ≠ tx identifying hash | Pass — `canonicalSigningDigest` / Homestead vs EIP-155 / typed signers |
| EIP-155 chain ID in V decoding | Pass — plain V derived; chain mismatch → inconclusive/invalid |
| Typed tx yParity {0,1} → 27/28 | Pass |
| `ValidateSignatureValues` | Pass |
| Uncompressed SEC1 (0x04\|\|X\|\|Y) | Pass |
| Keccak fingerprint of pubkey | Pass (display only) |
| Sender verification required for confirmed | Pass — mismatch → `sender_mismatch` / not confirmed |
| Unsupported types | Pass — explicit status, no false confirmed |

**Rule honored:** Recovered pubkey alone is never treated as confirmed exposure without derived address matching expected sender.

---

## Blockchain data reliability review

| Check | Result |
|-------|--------|
| Pagination completion flags | Present |
| Page/tx caps → `partial` | Present |
| Rate limit / failed statuses | Present |
| Duplicate tx handling | Counted / dropped |
| Block hash spot checks | Optional; mismatch → `partial` |
| Incomplete never labeled complete | `complete_within_scope` only when within configured caps |

---

## Frontend security review

| Check | Result |
|-------|--------|
| Address validation (Zod / hex) | Present |
| XSS / `dangerouslySetInnerHTML` | Not used in app source reviewed |
| Env exposure | Only public API base URLs via `NEXT_PUBLIC_*` |
| API errors | Humanized; no fabricated success |
| CORS | Backend allow-list (default localhost:3000) |
| sessionStorage | Scan index only (ids/addresses/status), not keys |
| Sensitive browser logs | No private-key logging paths |

---

## Python / quantum review

| Check | Result |
|-------|--------|
| Dependency pins | qiskit 1.x, aer, numpy; FastAPI optional |
| Input limits | Enforced in validation + Pydantic |
| Expensive sims | Bounded by qubit/shot/N caps |
| Error handling | 400 validation; 500 generic |
| Reproducibility | Seed parameters supported |
| Separation from scan evidence | Explicit; does not feed Go exposure |

---

## API security checklist

| Control | Status |
|---------|--------|
| Address validation | Yes |
| Request-size limit | Yes (1 MiB) |
| Inbound rate limit | Yes (configurable) |
| Request timeouts | Yes |
| CORS restrictions | Yes |
| Error-message leakage | Mitigated |
| Large-scan DoS | Mitigated (concurrency + caps) |
| In-memory storage limits | Yes (256/256 default) |
| Concurrent scan behavior | Semaphore |

---

## Test results (exact)

Commands run 2026-11-15 after remediations:

### Go

```text
$ go vet ./...
(exit 0, no output)

$ go test ./...
ok  ... (all packages with tests)
# packages without tests: cmd/api, domain, handler, quantum, transaction, version

$ go test -race ./...
ok  ... (all packages with tests; exit 0)
```

`golangci-lint`: **not installed / not configured** in this repository (no `.golangci*` present).

### Python (`quantum-engine`)

```text
$ python3 -m pytest
26 passed, 1 warning in ~10s
# warning: Starlette TestClient httpx deprecation
```

### Frontend

```text
$ npm test
Test Files  5 passed (5)
Tests       14 passed (14)

$ npm run typecheck
(exit 0)

$ npm run lint
✔ No ESLint warnings or errors

$ npm run build
✓ Compiled successfully
Linting and checking validity of types ...
✓ Generating static pages (9/9)
Route (app): /, /_not-found, /learn, /quantum, /reports, /reports/[id], /scanner, /wallet/[address]
(exit 0)
```

### Dependency audit (informational)

```text
$ npm audit --omit=dev
4 vulnerabilities (3 high, 1 critical)
# next/postcss/sharp/zod — tracked as L-06
```

`pip audit`: not available in the quantum venv (`pip audit` unknown command).

---

## MVP readiness statement

**MVP may proceed for local/trusted deployment** after V-01–V-03 fixes, with explicit acceptance of L-01–L-08 (no auth, ephemeral store, indexer trust, npm advisory follow-up).

Do **not** claim quantum safety from negative scans, do **not** treat exposure as compromise, and do **not** expose the API or quantum sidecar to the public Internet without authentication and edge rate limiting.
