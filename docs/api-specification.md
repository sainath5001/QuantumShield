# Q-Guard API Specification (MVP Draft)

Base URL (local): `http://localhost:8080`  
API prefix: `/api/v1`  
Content type: `application/json`

This specification is a design contract for the Go backend. It may be refined during implementation but semantic fields for assessments are normative per `business-logic.md`.

---

## 1. Conventions

### 1.1 Address format

- Input: `0x` + 40 hex characters.
- Recommendation: accept non-checksummed; normalize to EIP-55 in responses when feasible.
- Reject invalid hex / wrong length with `400`.

### 1.2 Scan identifiers

- `scan_id`: opaque string (ULID or UUID).

### 1.3 Status values

| `status` | Meaning |
|---|---|
| `queued` | Accepted, not started |
| `running` | In progress |
| `completed` | Finished with assessment |
| `failed` | Unrecoverable failure |

### 1.4 Error envelope

```json
{
  "error": {
    "code": "INVALID_ADDRESS",
    "message": "Address must be a 20-byte hex string with 0x prefix.",
    "details": {}
  }
}
```

Common codes: `INVALID_ADDRESS`, `UNSUPPORTED_CHAIN`, `SCAN_NOT_FOUND`, `RPC_UNAVAILABLE`, `HISTORY_PROVIDER_ERROR`, `RATE_LIMITED`, `INTERNAL_ERROR`.

---

## 2. Endpoints

### 2.1 Health (implemented)

`GET /api/v1/health`

Liveness only. Does not probe Ethereum RPC or Etherscan.

```json
{
  "status": "ok",
  "app": "Q-Guard",
  "version": "0.1.0-foundation",
  "risk_model_version": "qguard-risk-v1",
  "chain_id": 1,
  "environment": "development"
}
```

### 2.1b Ready (implemented)

`GET /api/v1/ready`

Probes Ethereum RPC (chain ID + block number) when a client is attached. Returns `503` when RPC is missing or mismatched.

**Does not** probe live Etherscan connectivity. Etherscan remains `configured` only — never reported as healthy from `/ready`.

```json
{
  "status": "ready",
  "ready": true,
  "checks": {
    "config": {"status": "ok", "message": "Configuration loaded and validated"},
    "ethereum_rpc": {"status": "ok", "message": "RPC reachable; chain ID matches configured mainnet; latest block observed"},
    "etherscan": {"status": "configured", "message": "API key and base URL present; live Etherscan connectivity is not probed by /ready"}
  },
  "message": "Essential dependencies passed readiness checks."
}
```

### 2.1c Info (implemented)

`GET /api/v1/info`

Returns mainnet chain metadata and capabilities. `history_retrieval` and `account_lookup` are true; `address_scan_implemented` and `public_key_recovery` remain false.

### 2.1d Account lookup (implemented)

`GET /api/v1/accounts/{address}`

Uses Ethereum RPC `eth_getCode` / `eth_getBalance` for an initial classification signal.

```json
{
  "address": "0x2449EcEF5012f0A0e153B278Ef4Fcc9625BC4C78",
  "kind": "likely_eoa",
  "has_code": false,
  "balance_wei": "123",
  "chain_id": 1,
  "block_number": 21000000,
  "limitations": [
    "Classification uses current eth_getCode as an initial signal, not absolute proof of wallet type."
  ]
}
```

`kind`: `likely_eoa` | `contract_account` | `unknown`

### 2.1e Transaction history (implemented)

`GET /api/v1/accounts/{address}/transactions`

Fetches **normal** transactions via Etherscan API v2 `account/txlist` with pagination, caps, and an explicit coverage object.

Does **not** fetch token transfers or internal transactions in this milestone. Does **not** perform signature recovery.

Coverage statuses:

| Status | Meaning |
|---|---|
| `complete_within_scope` | Pagination finished within provider pages/caps (not absolute chain completeness) |
| `partial` | Truncated by page/tx caps or inconsistent block hashes |
| `rate_limited` | Stopped due to provider rate limiting |
| `failed` | Provider/RPC failure before usable completion |
| `unsupported` | Reserved for unsupported response shapes |
| `unknown` | Coverage could not be determined confidently |

HTTP 200 from Etherscan alone never implies `complete_within_scope`.

### 2.1f Public-key exposure analysis (implemented)

`GET /api/v1/accounts/{address}/exposure`

Runs classification + (for likely EOAs) history fetch + RPC signed-tx recovery.

- **Confirmed exposure** only when recovery succeeds and the derived address matches the analyzed wallet.
- Contract accounts → `finding: requires_manual_review` (no EOA recovery claims).
- Does **not** return a risk score.

Findings: `public_key_exposed` | `no_exposure_observed_in_scope` | `insufficient_evidence` | `requires_manual_review`

---

### 2.2 Create scan (implemented)

`POST /api/v1/scan`

**Request**

```json
{
  "address": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0"
}
```

| Field | Required | Notes |
|---|---|---|
| `address` | yes | Public address only |
| `chain_id` | no | Ignored in MVP; server uses configured mainnet chain id 1 |

**Rejected fields (must not be accepted):** `private_key`, `mnemonic`, `seed`, `password`, signed payloads for broadcasting.

**Response `202 Accepted`** — returns a `ScanJob` (`id`, `status`, `stage`, `progress`, …).

Storage is **in-memory** and lost on process restart.

---

### 2.3 Get scan (implemented)

`GET /api/v1/scan/{id}`

**Response `200` (completed example — elevated)**

```json
{
  "scan_id": "01JABC...",
  "status": "completed",
  "address": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
  "chain_id": 1,
  "created_at": "2026-11-15T02:00:00Z",
  "completed_at": "2026-11-15T02:00:08Z",
  "account": {
    "type": "EOA",
    "has_code": false
  },
  "assessment": {
    "classification": "ELEVATED_EXPOSURE",
    "evidence_quality": "HIGH",
    "scheme": "secp256k1-ecdsa",
    "risk_model_version": "qguard-risk-v1",
    "analyzed_tx_count": 42,
    "unsupported_tx_count": 0,
    "failed_tx_count": 0,
    "verified_exposure_count": 3,
    "history_coverage": "complete_within_caps",
    "evidence": [
      {
        "kind": "TX_PUBKEY_RECOVERY",
        "tx_hash": "0xabc...",
        "tx_type": 2,
        "verified": true,
        "detail": "Recovered secp256k1 public key matches transaction sender.",
        "pubkey_fingerprint": "keccak4:0x12ab..."
      }
    ],
    "limitations": [
      {
        "code": "SCOPE_LIMITED",
        "message": "Analysis covers only the configured chain and history provider responses within caps."
      },
      {
        "code": "OFFCHAIN_NOT_COVERED",
        "message": "Off-chain signatures and other networks are not examined."
      }
    ],
    "disclaimers": [
      "Public-key exposure does not mean the private key is known today.",
      "No observed exposure does not mean the wallet is quantum-safe.",
      "Q-Guard never requests private keys or seed phrases."
    ]
  },
  "progress": {
    "stage": "completed",
    "txs_fetched": 42,
    "txs_processed": 42
  }
}
```

**Response `200` (running)**

```json
{
  "scan_id": "01JABC...",
  "status": "running",
  "address": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
  "chain_id": 1,
  "created_at": "2026-11-15T02:00:00Z",
  "account": {
    "type": "EOA",
    "has_code": false
  },
  "assessment": null,
  "progress": {
    "stage": "analyze_transactions",
    "txs_fetched": 100,
    "txs_processed": 40
  }
}
```

---

### 2.4 Wallet routes (implemented)

| Method | Path | Behavior |
|---|---|---|
| GET | `/api/v1/wallet/{address}` | Account classification + balance (alias of `/accounts/{address}`) |
| GET | `/api/v1/wallet/{address}/exposure` | Latest completed scan exposure, else live analysis |
| GET | `/api/v1/wallet/{address}/risk` | Latest completed `RiskAssessment` (`404` if none) |

### 2.5 Reports (implemented)

`GET /api/v1/reports/{id}`

Returns a complete `SecurityReport` (executive summary, wallet, cryptography, exposure, evidence, quantum threat explanation, risk, coverage, next steps, limitations, disclaimers).

No numerical compromise probability is included.

### 2.6 Meta: risk model

Risk model version is exposed on `/api/v1/info` and every assessment/report as `risk_model_version: qguard-risk-v1`.

---

## 3. Data models (domain ↔ API)

### 3.1 ScanJob

| Field | Type | Notes |
|---|---|---|
| `id` | string | |
| `address` | string | |
| `chain_id` | int64 | |
| `status` | enum | |
| `account_type` | `EOA` \| `CONTRACT` \| `UNKNOWN` | |
| `progress` | object | |
| `assessment` | object \| null | |
| `error` | object \| null | On `failed` |
| `created_at` | RFC3339 | |
| `completed_at` | RFC3339 \| null | |

### 3.2 Assessment

Matches mandatory fields in `business-logic.md` §7.

### 3.3 Classification enum

`LOWER_EXPOSURE` | `ELEVATED_EXPOSURE` | `UNKNOWN` | `REQUIRES_REVIEW`

### 3.4 History coverage enum

`complete_within_scope` | `partial` | `rate_limited` | `failed` | `unsupported` | `unknown`

---

## 4. Frontend ↔ backend interaction

```
POST /api/v1/scan  →  poll GET /api/v1/scan/{id} every 1–2s until completed|failed
GET /api/v1/reports/{report_id}  →  full SecurityReport when scan completed
```

Optional later: SSE/WebSocket; not required for MVP.

---

## 5. Quantum module API (Phase 2 proposal — not MVP-blocking)

Base: `http://localhost:8090` (Python sidecar)

| Method | Path | Purpose |
|---|---|---|
| GET | `/health` | Liveness |
| GET | `/demos` | List demos |
| GET | `/demos/shor/small-factor` | Toy Shor explanation + result metadata |
| GET | `/demos/grover/small-search` | Toy Grover explanation + result metadata |
| GET | `/explainers/post-quantum` | Static structured explainer JSON |

Responses must include `limitations` stating these are educational and unrelated to any Ethereum address scan.

The Go scan API must **not** call these endpoints when producing assessments.

---

## 6. Security requirements for the API

1. No endpoints that accept private keys or mnemonics.
2. No transaction broadcast endpoints in MVP.
3. No wallet-connect session endpoints for scanning.
4. Do not return fabricated chain data.
5. CORS restricted via `QGUARD_CORS_ORIGINS`.
6. Log addresses and scan IDs; never log secrets (none should exist).

---

## 7. OpenAPI

During implementation, publish `packages/openapi/qguard-v1.yaml` mirroring this document for frontend codegen / contract tests.

---

## 8. RPC and Etherscan limitations (Prompt 3)

| Area | Limitation |
|---|---|
| Chain | Ethereum Mainnet (`chain_id=1`) only; RPC chain mismatch aborts use of that client |
| Etherscan API | **v2** (`https://api.etherscan.io/v2/api` + `chainid`) |
| History kind | Normal `txlist` only; no token transfers / internal txs in this milestone |
| Completeness | Indexed + paginated + capped; never absolute proof of full chain history |
| Signatures | v2 `txlist` often omits `type` / `v,r,s`; left empty until RPC enrichment / recovery |
| `/ready` | Probes RPC only; does not claim Etherscan is reachable |
| Classification | `eth_getCode` is an initial signal (AA / proxies / history not fully modeled) |
| Reorgs | Optional block-hash spot checks vs RPC; mismatches downgrade coverage to `partial` |
