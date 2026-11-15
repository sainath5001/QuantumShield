"""Educational Shor-style period finding for tiny integers (toy only).

This module demonstrates the *structure* of Shor's factoring approach on N=15.
It is NOT a practical attack on secp256k1 / Ethereum private keys.
"""

from __future__ import annotations

import logging
import math
from fractions import Fraction
from typing import Any

from qiskit import QuantumCircuit, QuantumRegister, ClassicalRegister, transpile
from qiskit_aer import AerSimulator

try:
    from qiskit.circuit.library import QFTGate

    def _iqft_circuit(n: int) -> QuantumCircuit:
        qc = QuantumCircuit(n, name="iqft")
        qc.append(QFTGate(num_qubits=n).inverse(), range(n))
        return qc

except ImportError:  # pragma: no cover - older Qiskit 1.x
    from qiskit.circuit.library import QFT

    def _iqft_circuit(n: int) -> QuantumCircuit:
        return QFT(num_qubits=n, do_swaps=True, inverse=True).decompose()


from qguard_quantum import LIMITATIONS_BANNER
from qguard_quantum.validation import ValidationError, require_odd_composite, require_seed, require_shots
from qguard_quantum.visualization.circuits import circuit_metadata, counts_to_sorted_dict

logger = logging.getLogger("qguard_quantum.shor")

# Only these (N, a) pairs use the textbook modular-multiplication gadgets.
SUPPORTED_N = {15}
SUPPORTED_BASES_FOR_15 = {2, 4, 7, 8, 11, 13}


def _c_amod15(a: int, power: int) -> QuantumCircuit:
    """Controlled multiplication by a^power mod 15 (Qiskit textbook gadget)."""
    if a not in SUPPORTED_BASES_FOR_15:
        raise ValidationError(f"Unsupported base a={a} for N=15 educational circuit")
    U = QuantumCircuit(4, name=f"{a}^{power} mod 15")
    for _ in range(power):
        if a in (2, 13):
            U.swap(0, 1)
            U.swap(1, 2)
            U.swap(2, 3)
        if a in (7, 8):
            U.swap(2, 3)
            U.swap(1, 2)
            U.swap(0, 1)
        if a in (4, 11):
            U.swap(1, 3)
            U.swap(0, 2)
        if a in (7, 11, 13):
            for q in range(4):
                U.x(q)
    return U.control(1)


def build_period_finding_circuit(n: int, a: int, *, counting_qubits: int = 8) -> QuantumCircuit:
    """Build an educational order-finding circuit for N=15."""
    require_odd_composite(n, allowed=SUPPORTED_N)
    if a not in SUPPORTED_BASES_FOR_15:
        raise ValidationError(
            f"Unsupported a={a}. For N=15 educational demo use one of {sorted(SUPPORTED_BASES_FOR_15)}"
        )
    if math.gcd(a, n) != 1:
        raise ValidationError(f"a={a} shares a factor with n={n}; choose a coprime base")
    if counting_qubits not in (4, 8):
        raise ValidationError("counting_qubits must be 4 or 8 for this educational demo")

    counting = QuantumRegister(counting_qubits, "counting")
    work = QuantumRegister(4, "work")
    creg = ClassicalRegister(counting_qubits, "c")
    qc = QuantumCircuit(counting, work, creg)

    for q in range(counting_qubits):
        qc.h(counting[q])
    qc.x(work[0])  # |1> in work register

    for q in range(counting_qubits):
        qc.append(_c_amod15(a, 2**q), [counting[q], work[0], work[1], work[2], work[3]])

    qc.append(_iqft_circuit(counting_qubits), counting)
    qc.measure(counting, creg)
    return qc


def _continued_fraction_candidates(phase: float, *, max_denominator: int) -> list[int]:
    """Return candidate periods from continued-fraction expansion of a phase."""
    if phase <= 0:
        return []
    frac = Fraction(phase).limit_denominator(max_denominator)
    candidates = {frac.denominator}
    # Also consider nearby convergents via Fraction internals
    for lim in range(1, max_denominator + 1):
        f = Fraction(phase).limit_denominator(lim)
        if f.denominator > 0:
            candidates.add(f.denominator)
    return sorted(d for d in candidates if 1 < d <= max_denominator)


def _factors_from_period(a: int, n: int, r: int) -> tuple[int, int] | None:
    if r % 2 == 1:
        return None
    x = pow(a, r // 2, n)
    if x in (0, n - 1):
        return None
    p = math.gcd(x - 1, n)
    q = math.gcd(x + 1, n)
    if p * q == n and p > 1 and q > 1:
        return (min(p, q), max(p, q))
    return None


def classical_postprocess(
    counts: dict[str, int],
    *,
    a: int,
    n: int,
    counting_qubits: int,
) -> dict[str, Any]:
    """Map measurement bitstrings to candidate factors."""
    total = sum(counts.values()) or 1
    ranked = sorted(counts.items(), key=lambda kv: (-kv[1], kv[0]))
    attempts: list[dict[str, Any]] = []
    factors: tuple[int, int] | None = None

    for bitstr, count in ranked[:12]:
        # Qiskit bitstrings are little-endian by default for ClassicalRegister measures.
        value = int(bitstr, 2)
        phase = value / (2**counting_qubits)
        candidates = _continued_fraction_candidates(phase, max_denominator=n)
        entry = {
            "bitstring": bitstr,
            "value": value,
            "phase": phase,
            "probability": count / total,
            "count": count,
            "period_candidates": candidates,
            "factors": None,
        }
        for r in candidates:
            found = _factors_from_period(a, n, r)
            if found:
                entry["factors"] = {"r": r, "p": found[0], "q": found[1]}
                if factors is None:
                    factors = found
                break
        attempts.append(entry)

    success = factors is not None and factors[0] * factors[1] == n
    return {
        "success": success,
        "factors": {"p": factors[0], "q": factors[1]} if factors else None,
        "attempts": attempts,
    }


def run_shor_demo(
    n: int = 15,
    a: int = 7,
    *,
    counting_qubits: int = 8,
    shots: int = 1024,
    seed: int | None = 42,
) -> dict[str, Any]:
    """Run the educational Shor-style demo and return structured results."""
    require_odd_composite(n, allowed=SUPPORTED_N)
    shots = require_shots(shots)
    seed = require_seed(seed)

    logger.info(
        "Running educational Shor demo n=%s a=%s counting_qubits=%s shots=%s (NOT a secp256k1 attack)",
        n,
        a,
        counting_qubits,
        shots,
    )

    qc = build_period_finding_circuit(n, a, counting_qubits=counting_qubits)
    sim = AerSimulator(seed_simulator=seed) if seed is not None else AerSimulator()
    tqc = transpile(qc, sim)
    result = sim.run(tqc, shots=shots).result()
    counts = {k: int(v) for k, v in result.get_counts().items()}
    post = classical_postprocess(counts, a=a, n=n, counting_qubits=counting_qubits)

    return {
        "experiment": "shor_period_finding_toy",
        "limitations": LIMITATIONS_BANNER,
        "disclaimer": (
            "Factoring a tiny integer such as 15 is not equivalent to deriving an Ethereum private key. "
            "secp256k1 discrete-log instances require cryptographically large parameters far beyond this simulator demo."
        ),
        "input": {"n": n, "a": a, "counting_qubits": counting_qubits, "shots": shots, "seed": seed},
        "circuit": circuit_metadata(qc),
        "transpiled_depth": tqc.depth(),
        "transpiled_size": tqc.size(),
        "counts": counts_to_sorted_dict(counts),
        "classical_postprocessing": post,
        "success": post["success"],
        "discovered_factors": post["factors"],
    }
