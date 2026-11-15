"""Educational Grover search on a tiny unstructured space.

Does not imply that Grover's algorithm directly breaks Ethereum wallets.
"""

from __future__ import annotations

import logging
import math
from typing import Any

from qiskit import QuantumCircuit, transpile
from qiskit_aer import AerSimulator

from qguard_quantum import LIMITATIONS_BANNER
from qguard_quantum.validation import (
    ValidationError,
    require_bit_string,
    require_positive_int,
    require_seed,
    require_shots,
)
from qguard_quantum.visualization.circuits import circuit_metadata, counts_to_sorted_dict

logger = logging.getLogger("qguard_quantum.grover")

MAX_QUBITS = 4  # educational simulator limit for this module


def _oracle(n_qubits: int, target: str) -> QuantumCircuit:
    """Phase oracle marking the target bitstring with a multi-controlled Z.

    ``target`` is written MSB-left (qubit n-1 ... qubit 0), matching Qiskit count keys.
    """
    qc = QuantumCircuit(n_qubits, name="oracle")
    for i, bit in enumerate(reversed(target)):
        if bit == "0":
            qc.x(i)
    if n_qubits == 1:
        qc.z(0)
    else:
        qc.h(n_qubits - 1)
        qc.mcx(list(range(n_qubits - 1)), n_qubits - 1)
        qc.h(n_qubits - 1)
    for i, bit in enumerate(reversed(target)):
        if bit == "0":
            qc.x(i)
    return qc


def _diffusion(n_qubits: int) -> QuantumCircuit:
    qc = QuantumCircuit(n_qubits, name="diffusion")
    qc.h(range(n_qubits))
    qc.x(range(n_qubits))
    if n_qubits == 1:
        qc.z(0)
    else:
        qc.h(n_qubits - 1)
        qc.mcx(list(range(n_qubits - 1)), n_qubits - 1)
        qc.h(n_qubits - 1)
    qc.x(range(n_qubits))
    qc.h(range(n_qubits))
    return qc


def optimal_iterations(n_qubits: int) -> int:
    """Approximate optimal Grover iteration count for one marked item.

    Uses floor(π/4 √N). Rounding can overshoot on tiny spaces (e.g. N=4).
    """
    n = 2**n_qubits
    return max(1, int(math.floor((math.pi / 4) * math.sqrt(n))))


def build_grover_circuit(n_qubits: int, target: str, iterations: int | None = None) -> QuantumCircuit:
    require_positive_int("n_qubits", n_qubits, maximum=MAX_QUBITS)
    require_bit_string("target", target, width=n_qubits)
    if iterations is None:
        iterations = optimal_iterations(n_qubits)
    require_positive_int("iterations", iterations, maximum=64)

    qc = QuantumCircuit(n_qubits, n_qubits)
    qc.h(range(n_qubits))
    for _ in range(iterations):
        qc.compose(_oracle(n_qubits, target), inplace=True)
        qc.compose(_diffusion(n_qubits), inplace=True)
    qc.measure(range(n_qubits), range(n_qubits))
    return qc


def run_grover_demo(
    n_qubits: int = 2,
    target: str = "11",
    *,
    iterations: int | None = None,
    shots: int = 1024,
    seed: int | None = 42,
) -> dict[str, Any]:
    """Run a tiny Grover search and return structured results."""
    require_positive_int("n_qubits", n_qubits, maximum=MAX_QUBITS)
    require_bit_string("target", target, width=n_qubits)
    shots = require_shots(shots)
    seed = require_seed(seed)
    if iterations is None:
        iterations = optimal_iterations(n_qubits)
    else:
        require_positive_int("iterations", iterations, maximum=64)

    search_space = 2**n_qubits
    classical_worst = search_space
    classical_avg = (search_space + 1) / 2

    logger.info(
        "Running educational Grover demo n_qubits=%s target=%s iterations=%s (NOT an Ethereum attack)",
        n_qubits,
        target,
        iterations,
    )

    qc = build_grover_circuit(n_qubits, target, iterations=iterations)
    sim = AerSimulator(seed_simulator=seed) if seed is not None else AerSimulator()
    tqc = transpile(qc, sim)
    result = sim.run(tqc, shots=shots).result()
    counts = {k: int(v) for k, v in result.get_counts().items()}
    total = sum(counts.values()) or 1
    p_target = counts.get(target, 0) / total

    return {
        "experiment": "grover_search_toy",
        "limitations": LIMITATIONS_BANNER,
        "disclaimer": (
            "Grover provides a quadratic speedup for unstructured search. "
            "It does not directly 'break Ethereum wallets' the way Shor threatens ECC with known public keys."
        ),
        "input": {
            "n_qubits": n_qubits,
            "search_space_size": search_space,
            "target": target,
            "iterations": iterations,
            "shots": shots,
            "seed": seed,
        },
        "oracle_concept": (
            "The oracle flips the phase of the marked basis state. Diffusion amplifies that amplitude "
            "over successive Grover iterations."
        ),
        "classical_comparison": {
            "worst_case_queries": classical_worst,
            "average_case_queries_approx": classical_avg,
            "grover_iterations": iterations,
            "note": "Query-complexity intuition only; constant factors and oracle cost omitted.",
        },
        "circuit": circuit_metadata(qc),
        "transpiled_depth": tqc.depth(),
        "transpiled_size": tqc.size(),
        "counts": counts_to_sorted_dict(counts),
        "target_probability_estimate": p_target,
        "most_probable": max(counts.items(), key=lambda kv: kv[1])[0] if counts else None,
    }


def validate_unsupported_size(n_qubits: int) -> None:
    """Explicit guard used by tests for oversized inputs."""
    if n_qubits > MAX_QUBITS:
        raise ValidationError(
            f"n_qubits={n_qubits} exceeds educational maximum {MAX_QUBITS}. "
            "This module does not attempt cryptographically large Grover instances."
        )
