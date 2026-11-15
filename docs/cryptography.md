# Q-Guard Cryptography Reference

This document records the cryptographic foundations that Q-Guard relies on for analysis, messaging, and educational content. It is normative for product language: assessments must not contradict these definitions.

**Risk-model version reference:** concepts here underpin `risk_model_version` (see `business-logic.md`).

---

## 1. Ethereum Externally Owned Accounts (EOAs)

An EOA is an account controlled by a private key, not by contract code.

| Property | Value |
|---|---|
| Curve | secp256k1 (Koblitz curve over a 256-bit prime field) |
| Signature scheme | ECDSA (Elliptic Curve Digital Signature Algorithm) |
| Address size | 20 bytes (160 bits) |
| Private key | 256-bit scalar \(d\) in \([1, n-1]\) where \(n\) is the curve order |
| Public key | Point \(Q = d \cdot G\) on secp256k1 |

EOAs pay gas and authorize transactions by producing ECDSA signatures over a transaction signing digest. The Ethereum protocol does **not** store the public key on-chain as account state; only the address appears in the account trie.

---

## 2. Private key → public key → address

### 2.1 Key generation

1. Sample private key \(d\) uniformly from valid scalars.
2. Compute uncompressed public key \(Q = d \cdot G\).
3. Serialize \(Q\) as 64 bytes: \(x \,\|\, y\) (no `0x04` prefix for hashing; the `0x04` prefix is used in some encodings but **not** hashed for address derivation).

### 2.2 Address derivation

```
address = last_20_bytes( Keccak-256( uncompressed_pubkey_x_y ) )
```

Where `uncompressed_pubkey_x_y` is the 64-byte concatenation of the 32-byte big-endian \(x\) and \(y\) coordinates.

**Implications for Q-Guard:**

- Knowing only an address does **not** reveal the public key.
- Knowing the public key (64-byte or compressed form) allows recomputation and verification of the address.
- Recovering a public key from a signature and verifying it hashes to the claimed sender address is the primary evidence path for “key exposure.”

---

## 3. ECDSA signatures on Ethereum

An ECDSA signature is the pair \((r, s)\), with Ethereum additionally carrying a recovery identifier \(v\) (or \(yParity\) in typed transactions).

| Component | Role |
|---|---|
| \(r\) | \(x\)-coordinate related value of ephemeral point \(k \cdot G\) |
| \(s\) | \(s = k^{-1}(z + r \cdot d) \bmod n\) where \(z\) is the digest |
| \(v\) / `yParity` | Selects which of the candidate public keys is correct |

Given a signing digest \(z\) and \((r, s, v)\), Ethereum clients can run **public-key recovery** (`ecrecover`) to obtain the signer’s public key / address without prior knowledge of the key.

**Q-Guard uses recovery for analysis only.** The scanner never constructs signatures, never holds private keys, and never submits transactions.

---

## 4. Transaction types and signing digests

Ethereum has evolved multiple transaction encodings. The **signing digest** (the hash that is ECDSA-signed) differs by type. Incorrect digest construction yields wrong recovery and false negatives/positives.

### 4.1 Legacy (type 0)

- RLP list: `[nonce, gasPrice, gasLimit, to, value, data, v, r, s]` for the wire form.
- Pre-EIP-155 signing digest: `Keccak-256(RLP([nonce, gasPrice, gasLimit, to, value, data]))`.
- EIP-155 signing digest: `Keccak-256(RLP([nonce, gasPrice, gasLimit, to, value, data, chainId, 0, 0]))`.
- \(v\) encodes chain ID for EIP-155: \(v = \{27,28\}\) (pre-155) or \(v = chainId \times 2 + \{35,36\}\).

### 4.2 EIP-2930 Access List (type 1)

- Envelope: `0x01 \|\| RLP([chainId, nonce, gasPrice, gasLimit, to, value, data, accessList, yParity, r, s])`.
- Signing digest: `Keccak-256(0x01 \|\| RLP([chainId, nonce, gasPrice, gasLimit, to, value, data, accessList]))`.
- Recovery uses `yParity` ∈ {0,1} (not legacy \(v\)).

### 4.3 EIP-1559 Dynamic Fee (type 2)

- Envelope: `0x02 \|\| RLP([chainId, nonce, maxPriorityFeePerGas, maxFeePerGas, gasLimit, to, value, data, accessList, yParity, r, s])`.
- Signing digest: `Keccak-256(0x02 \|\| RLP([chainId, nonce, maxPriorityFeePerGas, maxFeePerGas, gasLimit, to, value, data, accessList]))`.

### 4.4 EIP-4844 Blob transactions (type 3)

- Envelope includes blob-related fields (`maxFeePerBlobGas`, `blobVersionedHashes`).
- Signing digest covers the typed payload **excluding** the signature fields, with type prefix `0x03`.
- Public-key recovery still applies for EOA-signed type-3 txs if digest construction is correct.

### 4.5 Unsupported / future types

Any transaction type not explicitly implemented in the decoder must be counted as **unsupported**, not silently skipped as “no exposure.” Future types (e.g., additional EIP envelopes) require explicit support before recovery claims.

### 4.6 Implementation rule

For each examined transaction:

1. Detect type.
2. Rebuild the exact signing digest per the EIP.
3. Recover public key from `(digest, r, s, yParity/v)`.
4. Derive address from recovered key; compare to `tx.from`.
5. Record match / mismatch / decode failure with evidence.

---

## 5. When is a public key “exposed”?

A secp256k1 public key associated with an Ethereum address is considered **observed as exposed** when Q-Guard can recover or otherwise obtain a public key that:

1. Verifies against a valid signature over a correctly reconstructed digest, **and**
2. Derives to the expected sender address (or is otherwise cryptographically bound to that address via verified evidence).

**Sources of exposure (in scope for MVP):**

- On-chain EOA transaction signatures (primary).

**Sources of exposure (out of scope for MVP, must be disclosed as limitations):**

- Off-chain signatures (SIWE, Permit, Permit2, personal_sign, EIP-712).
- Same key reused on other EVM chains or non-EVM networks.
- Historical archives not returned by the configured history provider.
- Multisig / smart-account internal signature verification paths that do not surface standard `ecrecover`-style EOA signatures in the examined tx.

**Critical product language:**

> Public-key exposure ≠ compromised wallet.  
> Exposure means a future cryptanalytically relevant quantum adversary *might* target that key. It does **not** mean a present-day classical attacker can derive the private key from the public key alone, and it does **not** mean the wallet has been “quantum-hacked.”

---

## 6. Contract wallets and smart accounts

Contract accounts have code. They do not use a single EOA secp256k1 private key as their account identity in the same way EOAs do.

| Mechanism | Implication for Q-Guard |
|---|---|
| ERC-4337 Account Abstraction | UserOperations may be validated by custom `validateUserOp` logic |
| Safe / multisig | On-chain signature checks may recover multiple EOAs or use contract signatures |
| EIP-1271 `isValidSignature` | Validity is defined by contract code, not by a single secp256k1 key |
| EIP-712 typed data | Digests differ from transaction digests |

**MVP policy:** If `eth_getCode(address)` is non-empty, classify as contract/smart account → `REQUIRES_REVIEW`. Do not claim EOA key exposure solely from “from” fields that are contracts. Optionally (post-MVP) analyze known Safe/4337 patterns with explicit confidence bounds.

---

## 7. Shor’s algorithm and elliptic-curve cryptography

Shor’s algorithm (extended to elliptic-curve discrete logarithms) provides a **polynomial-time** quantum algorithm for computing discrete logarithms on abelian groups, including secp256k1.

**Threat narrative Q-Guard may use:**

- Given a known public key \(Q = d \cdot G\), a sufficiently large fault-tolerant quantum computer running an ECC discrete-log instance of Shor’s algorithm could recover \(d\).
- Addresses whose public keys have **never** been revealed retain a one-way barrier (address = hash of pubkey). Breaking that still requires either key exposure elsewhere or attacking the hash / other surfaces—not a direct ECC-DL on an unknown point.

**What Q-Guard must not claim:**

- That such a machine exists today for 256-bit ECC.
- A calendar date when Ethereum keys will fall.
- That absence of observed exposure means quantum-safe.

**Educational demos (Python/Qiskit):** Use tiny toy moduli / small elliptic or integer factoring instances. Explicitly state that simulator demos are **not** attacks on secp256k1.

---

## 8. Grover’s algorithm

Grover’s algorithm provides a quadratic speedup for unstructured search.

| Classical | Grover-adjusted intuition |
|---|---|
| \(n\)-bit brute force ~ \(2^n\) | ~ \(2^{n/2}\) oracle queries (idealized) |

**Implications:**

- Symmetric keys and hashes need larger security margins under a quantum threat model (e.g., AES-128 → often discussed as needing AES-256 for long-term; hash preimage resistance similarly).
- Grover does **not** provide the same catastrophic break of ECDSA that Shor does for discrete log.
- Keccak-256 address truncation to 160 bits has its own collision/preimage considerations; Q-Guard’s MVP focus remains **ECC public-key exposure under Shor**, with Grover covered educationally.

---

## 9. Post-quantum standards (NIST)

| Standard | Name | Role |
|---|---|---|
| FIPS 203 | ML-KEM (Kyber family) | Key encapsulation / key establishment |
| FIPS 204 | ML-DSA (Dilithium family) | Digital signatures |
| FIPS 205 | SLH-DSA (SPHINCS+) | Stateless hash-based signatures |

**Ethereum migration challenges (product education, not scanner claims):**

- Account model assumes secp256k1 ECDSA for EOAs.
- Signature size / gas / verification cost changes for PQ schemes.
- Need for account abstraction, new tx types, or hybrid signatures.
- Ecosystem-wide coordination (wallets, L2s, bridges, hardware wallets).
- Harvest-now-decrypt-later / harvest-now-sign-later style risks for exposed keys.

Q-Guard’s scanner remains classical analysis of current ECDSA exposure; PQ content is educational and roadmap-oriented unless separately scoped.

---

## 10. Libraries and implementation notes (Go)

Planned stack (no application code in this phase):

| Concern | Approach |
|---|---|
| RPC / types | `go-ethereum` (`ethclient`, `types.Transaction`, `crypto`) |
| Keccak-256 | `go-ethereum/crypto` (or equivalent audited Keccak) |
| ECDSA recover | `crypto.SigToPub` / equivalent with correct `v`/`yParity` normalization |
| Address check | `crypto.PubkeyToAddress` vs `tx.From()` / recovered sender |

**Cryptographic implementation risks (must be mitigated in tests):**

1. Wrong signing hash for typed transactions → false “no recovery” or wrong key.
2. Incorrect `v` / `yParity` normalization (27/28 vs 0/1).
3. Compressed vs uncompressed pubkey handling in hashing.
4. Treating contract `from` as EOA.
5. Trusting indexed API `from` without verifying recovery.
6. Claiming global non-exposure from partial history.
7. Confusing signature malleability / `s` normalization with exposure status.

---

## 11. Glossary (product-facing)

| Term | Definition |
|---|---|
| Public-key exposure | Public key observed/recoverable and bound to the address under analysis |
| Quantum-vulnerable scheme | Cryptosystem for which a known quantum algorithm (e.g., Shor) breaks the hardness assumption |
| Evidence quality | Strength of observed cryptographic evidence for the assessment |
| Quantum-safe (wallet) | **Not** asserted by MVP scanner based on scan results alone |
| Classical compromise | Private key known/stolen by conventional means—out of scope |

---

## 12. References (non-exhaustive)

- Yellow Paper / Execution layer specs; EIPs 155, 2718, 2930, 1559, 4844.
- NIST FIPS 203, 204, 205.
- Shor (1994); elliptic-curve discrete log quantum algorithms literature.
- Grover (1996).
- Ethereum account abstraction / ERC-4337 documentation for smart accounts.
