# Q-Guard Business Logic

Normative rules for wallet classification, evidence gathering, risk classification, and report contents. Implementation must follow this document; UI copy must not contradict it.

**Risk model version:** `qguard-risk-v1` (increment on any semantic change to classification rules).

---

## 1. Product objective (operationalized)

Given a public Ethereum address and an analysis scope, Q-Guard:

1. Classifies the account type (EOA vs contract).
2. Retrieves available transaction history from configured sources.
3. Attempts ECDSA public-key recovery on supported EOA transaction signatures.
4. Verifies recovered keys against expected senders.
5. Emits a structured assessment with evidence, confidence, coverage, and limitations.
6. Educates users on quantum algorithms and post-quantum migration **separately** from scan evidence.

---

## 2. Analysis scope

Every scan defines a scope object:

| Field | Description |
|---|---|
| `chain_id` | e.g. `1` |
| `history_provider` | e.g. `etherscan` |
| `max_transactions` | Hard cap |
| `max_pages` | Pagination cap |
| `supported_tx_types` | e.g. `[0,1,2,3]` |
| `started_at` / `completed_at` | Timestamps |
| `address` | Normalized address |

Assessments are valid **only** within this scope. Cross-scope claims are forbidden.

---

## 3. Account classification

```
code = eth_getCode(address)
if code is empty or "0x":
    account_type = EOA
else:
    account_type = CONTRACT
```

| Account type | Default assessment path |
|---|---|
| `EOA` | Full signature exposure analysis |
| `CONTRACT` | `REQUIRES_REVIEW` (MVP); optional contextual history without EOA-style final claims |

**Note:** An EOA that later becomes a contract (rare/edge) is out of MVP sophistication; document if encountered.

---

## 4. History retrieval rules

1. Prefer transactions where the address is **`from`** (sender)—these are the primary source of recoverable EOA pubkeys for that address.
2. Paginate until: no more results, `max_pages`, `max_transactions`, or fatal provider error.
3. On rate limit: exponential backoff with jitter; if exhausted, set `history_coverage = partial` and continue with collected data.
4. Record provider cursors/pages for auditability in the scan record.
5. Do not invent transactions if the provider returns empty.

**Incoming-only transactions** do not expose the recipient’s public key via the recipient field. They must not be counted as exposure evidence for the scanned address’s key.

---

## 5. Transaction analysis pipeline

For each candidate transaction:

1. **Fetch raw tx** (from indexer fields and/or RPC) sufficient to rebuild the signing digest.
2. **Detect type** (`type` field / envelope).
3. If type ∉ `supported_tx_types` → `unsupported_tx_count++`; stop processing that tx.
4. Rebuild signing digest per EIP rules (`cryptography.md`).
5. Normalize signature (`r`, `s`, `v`/`yParity`).
6. Recover public key.
7. Derive address; compare to `tx.from` (case-insensitive hex).
8. Outcomes:
   - `verified_exposure` — recovery OK and address matches scanned address (when analyzing scanned address as sender)
   - `recovery_mismatch` — recovery OK but address ≠ expected (data inconsistency; not exposure evidence)
   - `recovery_failed` — cryptographic/decode failure
   - `unsupported` — type not supported

**Verification is mandatory** before counting exposure.

---

## 6. Risk classification model (`qguard-risk-v1`)

### 6.1 Labels

| Label | Meaning |
|---|---|
| `LOWER_EXPOSURE` | No verified public-key exposure observed **within scope**. Does **not** prove the key was never exposed. |
| `ELEVATED_EXPOSURE` | ≥1 verified public-key recovery bound to the scanned EOA address; secp256k1 ECDSA identified as the scheme. |
| `UNKNOWN` | Insufficient evidence: incomplete history, too many unsupported txs, RPC/indexer failures, or ambiguous results. |
| `REQUIRES_REVIEW` | Contract wallet / smart-account / custom verification path needs further analysis. |

**No probability percentages. No attack timelines.**

### 6.2 Decision rules (EOA)

Apply in order:

1. If scan failed before meaningful analysis → status failed; no optimistic label (API error / `UNKNOWN` if partial).
2. If `history_coverage` is `none` (no usable history fetch) → `UNKNOWN`.
3. If `history_coverage` is `partial` **and** no verified exposure → `UNKNOWN` (do not reward incomplete data with `LOWER_EXPOSURE`).
4. If verified exposure count ≥ 1 → `ELEVATED_EXPOSURE`.
5. If `unsupported_tx_count > 0` and `analyzed_ok_count == 0` and no verified exposure → `UNKNOWN`.
6. If `unsupported_tx_count` is high relative to total (recommended threshold: unsupported ≥ 50% of fetched sender txs) and no verified exposure → `UNKNOWN`.
7. If coverage `complete_within_caps`, analyzed supported sender txs ≥ 1 or zero sender txs with complete empty history, and verified exposure = 0 → `LOWER_EXPOSURE`.
8. If zero outbound txs ever returned and coverage complete → `LOWER_EXPOSURE` with explicit “no sender txs in scope” evidence (still not a safety certificate).

### 6.3 Decision rules (CONTRACT)

1. MVP: classification = `REQUIRES_REVIEW`.
2. May still attach contextual metrics (tx counts) as informational, not as EOA exposure proof for the contract address itself.
3. Post-MVP: specialized modules may flag known patterns without removing `REQUIRES_REVIEW` unless confidence criteria are explicitly versioned.

### 6.4 Confidence / evidence quality

| Level | When |
|---|---|
| `HIGH` | Verified exposure with matching address on supported types; or `LOWER_EXPOSURE` with complete coverage and only supported types |
| `MEDIUM` | Verified exposure present but mixed with some unsupported/decode errors; or lower exposure with minor gaps that still passed rule 7 carefully |
| `LOW` | `UNKNOWN` / partial coverage / high unsupported ratio / provider inconsistencies |
| `MANUAL` | `REQUIRES_REVIEW` |

Confidence describes **evidence quality**, not probability of quantum attack.

---

## 7. Mandatory report fields

Every completed assessment includes:

| Field | Description |
|---|---|
| `classification` | One of the four labels |
| `evidence` | List of evidence items (tx hashes, recovery summaries, account type) |
| `evidence_quality` | HIGH / MEDIUM / LOW / MANUAL |
| `analyzed_tx_count` | Successfully decoded & examined |
| `unsupported_tx_count` | Unsupported types |
| `failed_tx_count` | Decode/recovery failures |
| `verified_exposure_count` | Count of verified exposures |
| `history_coverage` | `complete_within_scope` \| `partial` \| `rate_limited` \| `failed` \| `unsupported` \| `unknown` |
| `limitations` | Human-readable + machine codes |
| `risk_model_version` | e.g. `qguard-risk-v1` |
| `scheme` | e.g. `secp256k1-ecdsa` when applicable |
| `disclaimers` | Fixed product disclaimers |

---

## 8. Evidence item schema (logical)

```text
EvidenceItem
  kind: ACCOUNT_CODE | TX_PUBKEY_RECOVERY | HISTORY_COVERAGE | PROVIDER_NOTE
  tx_hash?: string
  tx_type?: number
  detail: string
  pubkey_fingerprint?: string   # e.g. truncated hash of pubkey; avoid huge payloads in lists
  verified: bool
```

Full public keys may be included in detail endpoints but should be handled carefully in UI (copyable, not sensationalized).

---

## 9. Disclaimers (required)

At least the following concepts must appear on every report:

1. No private keys requested or used.
2. Exposure ≠ present-day private key compromise.
3. No observed exposure ≠ quantum-safe.
4. Scope-limited: other chains / off-chain signatures not covered.
5. No claim of quantum hacking having occurred.
6. Residual risk statement from `quantum-threat-model.md` §11.

---

## 10. Educational content vs scan results

| Content | May influence classification? |
|---|---|
| On-chain recovery evidence | Yes |
| Qiskit demo output | **No** |
| PQ standards articles | **No** |
| User-entered notes | **No** |

---

## 11. Non-goals (business)

- Wallet custody, swapping, portfolio tracking as a primary goal.
- Classical vulnerability scanning (approval phishing, etc.) in MVP.
- Insurance or compliance certification.
- Claiming migration completeness for Ethereum as a whole.

---

## 12. Change control

Any change to §6 rules requires:

1. Bump `risk_model_version`.
2. Update this document and API examples.
3. Add regression tests for old vs new behavior notes in changelog.
