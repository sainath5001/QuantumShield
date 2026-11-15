# Q-Guard Quantum Threat Model

This document defines what “quantum threat” means for Q-Guard, what evidence the product may cite, and what claims are forbidden. It is the normative threat model for risk scoring and UI copy.

**Related docs:** `cryptography.md`, `business-logic.md`.

---

## 1. Assets

| Asset | Description | Sensitivity |
|---|---|---|
| EOA private key \(d\) | Controls funds and authority for an EOA | Critical |
| secp256k1 public key \(Q\) | Enables discrete-log targeting once known | High (under CRQC assumptions) |
| Ethereum address | Public identifier; hash of pubkey | Public |
| Transaction history | Evidence of signatures / key use | Public (when available) |
| Scan reports | Derived assessments; must not invent chain data | Integrity-critical |

---

## 2. Adversaries

### 2.1 Present-day classical adversary

- Can read public chain data and indexed APIs.
- Cannot derive secp256k1 private keys from public keys with known practical classical attacks under standard assumptions.
- May already possess keys via phishing, malware, or operational compromise—**out of Q-Guard’s detection scope**.

### 2.2 Cryptographically relevant quantum computer (CRQC)

Hypothetical future adversary capable of running large-scale fault-tolerant quantum algorithms that solve elliptic-curve discrete logarithms at secp256k1 parameters (Shor-type ECC-DL).

**Assumptions Q-Guard makes explicit:**

- CRQC availability is **uncertain** in timeline and capability.
- Q-Guard does **not** estimate probability or year of CRQC arrival.
- Under a CRQC, **exposed** ECDSA public keys become high-value targets for key recovery; **unexposed** keys retain the address-hash barrier for that specific key material (with caveats in §5).

### 2.3 Harvest-now adversaries

Actors who collect exposed public keys and signatures today for potential future CRQC exploitation (“store now, attack later”). This motivates scanning for exposure **before** CRQC existence.

---

## 3. Trust boundaries

```
[User / Frontend]
       │  public address only
       ▼
[Go API] ──► in-memory scan store
       │
       ├──► Ethereum JSON-RPC (execution client / provider)
       └──► Etherscan (or compatible) history indexer

[Python/Qiskit module] ── educational only; not on the scan evidence path
```

**Trust rules:**

- Users supply **only** a public address (and optional chain/network selection).
- No wallet connection required for scanning.
- RPC and indexer data are **untrusted inputs** until cross-checked (signature recovery vs sender).
- Educational quantum outputs must never be mixed into production exposure evidence.

---

## 4. What Q-Guard reliably detects

Within a defined **analysis scope** (chain ID, history source, block/time window, supported tx types):

1. Whether the target is an EOA or has contract code (via `eth_getCode`).
2. Whether examined EOA transactions of **supported types** allow ECDSA public-key recovery.
3. Whether recovered keys derive to the claimed sender address.
4. Whether at least one verified public-key exposure event was observed in that scope.
5. Counts of analyzed vs unsupported vs failed-decode transactions.
6. History coverage status relative to provider responses (complete within returned pages vs truncated/errors/rate-limited).

---

## 5. What Q-Guard cannot detect

1. Whether a private key is already stolen classically.
2. Whether the same key was exposed on another chain, in an L2 archive not queried, or off-chain (EIP-712, SIWE, Permit, etc.).
3. Whether a key will be exposed in the future.
4. Whether a CRQC exists or when it will exist.
5. Whether an address with no observed exposure is “quantum-safe.”
6. Custom smart-account validation logic without specialized analyzers (MVP → `REQUIRES_REVIEW`).
7. Realistic 256-bit Shor breaks via Qiskit simulators (impossible / not attempted).
8. Guaranteed completeness of “all historical transactions ever” if the indexer paginates, rate-limits, or omits data.

---

## 6. Attack / risk scenarios (product framing)

### Scenario A — Exposed EOA key under future CRQC

1. User broadcasts EOA transactions; signatures allow pubkey recovery.
2. Adversary archives \(Q\).
3. Future CRQC solves ECC-DL for \(Q\), recovers \(d\).
4. Adversary drains account or impersonates signer.

**Q-Guard role:** Detect and evidence step 1 within scope; explain steps 2–4 as **conditional** on CRQC.

### Scenario B — Address-only EOA (no observed exposure)

1. Address known; pubkey not observed in scope.
2. CRQC cannot run ECC-DL on an unknown point tied to that address without additional exposure or other attacks.

**Q-Guard role:** Classify `LOWER_EXPOSURE` with strong limitations text—not a safety certificate.

### Scenario C — Contract / smart account

1. Funds controlled by code and possibly multiple signers.
2. Exposure may exist for underlying EOAs without a single “account public key.”

**Q-Guard role:** `REQUIRES_REVIEW`; avoid EOA-style conclusions.

### Scenario D — Incomplete history

1. Indexer returns partial data or unsupported types dominate.
2. True exposure may exist outside evidence.

**Q-Guard role:** Prefer `UNKNOWN` (or elevate uncertainty) rather than `LOWER_EXPOSURE` when coverage is insufficient.

---

## 7. Forbidden claims

The product, API, and educational copy must not:

- Request or imply need for private keys, seeds, or passwords.
- State that a wallet was hacked by a quantum computer.
- State that a wallet is quantum-safe because no exposure was found.
- Present fabricated transactions, signatures, or public keys in production scan results.
- Present precise probabilities or dates for CRQC attacks.
- Attribute Python/Qiskit demo outcomes as evidence about a scanned address.
- Equate “public key known” with “private key known today.”

---

## 8. Evidence vs assumption

| Kind | Examples |
|---|---|
| Evidence | Recovered pubkey matches sender; `getCode` result; tx hash; type; recovery success/failure |
| Assumption | CRQC will exist; indexer completeness; key unused elsewhere |
| Limitation | Unsupported tx type; rate limit; reorg risk; off-chain signatures |

Every assessment must separate these three in structured fields (see API / business logic).

---

## 9. Grover vs Shor in product messaging

| Algorithm | Role in Q-Guard |
|---|---|
| Shor (ECC-DL) | Primary reason exposed secp256k1 keys are concerning long-term |
| Grover | Educational; symmetric/hash security-margin discussion; not the MVP risk classifier driver |

Do not imply Grover alone “breaks Ethereum wallets” in the same way Shor threatens ECDSA with known pubkeys.

---

## 10. Post-quantum readiness (educational scope)

Q-Guard explains:

- NIST PQ standards (ML-KEM, ML-DSA, SLH-DSA).
- Why Ethereum EOAs are not drop-in PQ-safe.
- Migration challenges: gas, signature size, account abstraction, ecosystem coordination.

MVP does **not** score wallets as “PQ ready” based on on-chain PQ usage unless a later version explicitly adds detection of specific PQ/hybrid schemes with documented false-positive rates.

---

## 11. Residual risk statement (canonical)

> Q-Guard analyzes publicly available transaction data for cryptographic public-key exposure relevant to a hypothetical future cryptographically relevant quantum computer. It does not assess classical theft risk, does not prove absence of exposure outside the examined scope, and does not predict quantum-computing timelines.

This statement (or equivalent) should appear on reports and educational pages.
