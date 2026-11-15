# Post-quantum cryptography overview

Educational material from `qguard_quantum.pqc`. This module explains NIST PQC families; it does **not** implement production ML-DSA / ML-KEM / SLH-DSA cryptography and does **not** label wallets quantum-resistant.

## Why quantum-resistant algorithms matter

Fault-tolerant quantum computers threaten widely deployed public-key schemes (RSA via Shor; elliptic-curve schemes used by Ethereum EOAs). PQC algorithms rely on different hardness assumptions (lattices, hashes, …) believed to resist known classical and quantum attacks.

## ML-DSA

**Module-Lattice-Based Digital Signature Algorithm** (FIPS 204, Dilithium lineage).

Role: digital signatures — authenticate messages / authorize actions.

Relevance: a candidate class if Ethereum or wallets eventually migrate signature schemes. Migration is protocol-wide, not a library swap inside one scanner.

## ML-KEM

**Module-Lattice-Based Key-Encapsulation Mechanism** (FIPS 203, Kyber lineage).

Role: establish a shared secret for symmetric encryption.

### Why ML-KEM is not a digital signature scheme

KEMs encapsulate keys. They do **not** provide the “sign with private key / verify with public key” authorization model EOAs use for transactions. Calling a KEM experiment “quantum-safe Ethereum signatures” is incorrect.

## SLH-DSA

**Stateless Hash-Based Digital Signature Algorithm** (FIPS 205, SPHINCS+ lineage).

Role: hash-based signatures with conservative assumptions, typically larger signatures. Still needs wallet and protocol support to replace EOAs.

## Replacing Ethereum EOA signatures

Ethereum EOAs today assume secp256k1 keys and ECDSA-compatible transaction formats, tooling, hardware wallets, and address derivation. Replacing EOA signatures requires **broader protocol and wallet support** (new tx types, account abstraction patterns, address schemes, ecosystem migration)—not merely demonstrating ML-DSA in a research notebook.

## Scanner boundary

Q-Guard’s Go backend assesses classical public-key exposure on Mainnet. This Python module teaches quantum algorithms and PQC concepts separately. A successful toy Shor/Grover run or a local PQC demo must never become a “wallet is quantum-safe” claim.

```bash
qguard-quantum pqc
qguard-quantum pqc --algorithm ML-KEM
```
