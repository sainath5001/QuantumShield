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

Returns whether essential configuration is present. External dependency checks are `configured` — **not** live connectivity probes.

```json
{
  "status": "ready",
  "ready": true,
  "checks": {
    "config": {"status": "ok", "message": "Configuration loaded and validated"},
    "ethereum_rpc": {"status": "configured", "message": "RPC URL present; connectivity is not probed in the foundation milestone"},
    "etherscan": {"status": "configured", "message": "API key and base URL present; connectivity is not probed in the foundation milestone"}
  },
  "message": "Essential configuration is present. External dependency connectivity is not probed in this milestone."
}
```

### 2.1c Info (implemented)

`GET /api/v1/info`

Returns chain metadata, planned transaction types, and capability flags. In the foundation milestone, scan-related capabilities are explicitly `false` / `implemented: false`.

---

### 2.2 Create scan (not implemented yet)

`POST /api/v1/scans`

**Request**

```json
{
  "address": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
  "chain_id": 1
}
```

| Field | Required | Notes |
|---|---|---|
| `address` | yes | Public address only |
| `chain_id` | no | Default from server config |

**Rejected fields (must not be accepted):** `private_key`, `mnemonic`, `seed`, `password`, signed payloads for broadcasting.

**Response `202 Accepted`**

```json
{
  "scan_id": "01JABC...",
  "status": "queued",
  "address": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
  "chain_id": 1,
  "created_at": "2026-11-15T02:00:00Z"
}
```

---

### 2.3 Get scan

`GET /api/v1/scans/{scan_id}`

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

### 2.4 Get scan evidence detail (optional MVP+)

`GET /api/v1/scans/{scan_id}/exposures`

Returns expanded exposure events including full uncompressed public key hex for verified items only.

---

### 2.5 Meta: risk model

`GET /api/v1/meta/risk-model`

```json
{
  "version": "qguard-risk-v1",
  "classifications": [
    "LOWER_EXPOSURE",
    "ELEVATED_EXPOSURE",
    "UNKNOWN",
    "REQUIRES_REVIEW"
  ],
  "notes": "See docs/business-logic.md and docs/quantum-threat-model.md"
}
```

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

`complete_within_caps` | `partial` | `none`

---

## 4. Frontend ↔ backend interaction

```
POST /scans  →  poll GET /scans/{id} every 1–2s until completed|failed
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
