"""Post-quantum cryptography education (explanatory only).

Does not implement ML-DSA / ML-KEM / SLH-DSA cryptosystems and does not
label any Ethereum wallet as quantum-resistant.
"""

from __future__ import annotations

from typing import Any

from qguard_quantum import LIMITATIONS_BANNER


def pqc_overview() -> dict[str, Any]:
    """Structured PQC education content for docs and CLI."""
    return {
        "limitations": LIMITATIONS_BANNER,
        "disclaimer": (
            "Using a post-quantum algorithm in an unrelated experiment does not make an "
            "Ethereum externally owned account (EOA) quantum-resistant."
        ),
        "why_quantum_resistant_matters": (
            "Large-scale fault-tolerant quantum computers could threaten widely deployed "
            "public-key schemes (for example RSA via Shor, and elliptic-curve schemes used "
            "by Ethereum EOAs). Post-quantum cryptography (PQC) aims to provide algorithms "
            "believed to resist both classical and known quantum attacks, based on different "
            "hardness assumptions (lattices, hashes, etc.)."
        ),
        "algorithms": {
            "ML-DSA": {
                "full_name": "Module-Lattice-Based Digital Signature Algorithm",
                "nist_status": "FIPS 204 (Dilithium lineage)",
                "role": "Digital signatures (authenticate messages / transactions).",
                "relevance": (
                    "A candidate class for eventual signature schemes if Ethereum (or wallets) "
                    "migrate away from secp256k1 ECDSA/EdDSA-style EOAs. Migration is a protocol "
                    "and ecosystem change, not a single library swap."
                ),
            },
            "ML-KEM": {
                "full_name": "Module-Lattice-Based Key-Encapsulation Mechanism",
                "nist_status": "FIPS 203 (Kyber lineage)",
                "role": "Key encapsulation / establishing shared secrets.",
                "relevance": (
                    "Used to establish keys for symmetric encryption. It is not a drop-in "
                    "replacement for Ethereum transaction signatures."
                ),
                "not_a_signature": (
                    "ML-KEM encapsulates a shared secret; it does not produce a public verification "
                    "scheme for authorizing state transitions the way ECDSA signatures do for EOAs. "
                    "Confusing KEMs with signatures is a common educational mistake."
                ),
            },
            "SLH-DSA": {
                "full_name": "Stateless Hash-Based Digital Signature Algorithm",
                "nist_status": "FIPS 205 (SPHINCS+ lineage)",
                "role": "Hash-based digital signatures with conservative assumptions.",
                "relevance": (
                    "Attractive when implementers want hash-based assumptions, typically with "
                    "larger signatures. Still requires wallet and protocol support to replace EOAs."
                ),
            },
        },
        "ethereum_eoa_migration": {
            "why_not_just_swap_crypto": (
                "Ethereum EOAs today are defined around secp256k1 keys and ECDSA signatures "
                "in transaction formats, account addressing, RPC tooling, hardware wallets, "
                "and smart-contract recovery patterns. Replacing EOA signatures requires "
                "broader protocol and wallet support (account abstraction / new tx types / "
                "address schemes), not merely running ML-DSA in a research notebook."
            ),
            "scanner_boundary": (
                "Q-Guard's Go backend assesses classical public-key exposure on Mainnet. "
                "This Python module teaches quantum algorithms and PQC concepts separately. "
                "They must not be conflated into a 'wallet is quantum-safe' claim."
            ),
        },
    }


def explain_algorithm(name: str) -> dict[str, Any]:
    """Return one algorithm section; raises KeyError-style ValidationError via lookup."""
    from qguard_quantum.validation import ValidationError

    data = pqc_overview()
    key = name.strip().upper().replace(" ", "-")
    aliases = {
        "DILITHIUM": "ML-DSA",
        "KYBER": "ML-KEM",
        "SPHINCS+": "SLH-DSA",
        "SPHINCS": "SLH-DSA",
    }
    key = aliases.get(key, key)
    alg = data["algorithms"].get(key)
    if alg is None:
        raise ValidationError(f"Unknown PQC topic {name!r}. Use ML-DSA, ML-KEM, or SLH-DSA.")
    return {"name": key, "limitations": LIMITATIONS_BANNER, **alg}
