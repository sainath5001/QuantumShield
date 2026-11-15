"""Educational quantum fundamentals demonstrations."""

from __future__ import annotations

from typing import Any

from qiskit import QuantumCircuit, transpile
from qiskit_aer import AerSimulator

from qguard_quantum import LIMITATIONS_BANNER
from qguard_quantum.validation import ValidationError, require_seed, require_shots
from qguard_quantum.visualization.circuits import circuit_metadata, counts_to_sorted_dict


def explain_fundamentals() -> dict[str, Any]:
    """Return structured educational explanations (no simulation)."""
    return {
        "limitations": LIMITATIONS_BANNER,
        "topics": {
            "classical_bit": (
                "A classical bit is either 0 or 1. Copying and reading it does not disturb the value "
                "beyond ordinary noise."
            ),
            "qubit": (
                "A qubit has a quantum state that can be a superposition α|0⟩ + β|1⟩ with "
                "|α|² + |β|² = 1. Measurement yields 0 or 1 with those probabilities."
            ),
            "superposition": (
                "Superposition is not '0 and 1 at once' in a classical sense; it is a complex-amplitude "
                "description of possible measurement outcomes."
            ),
            "measurement": (
                "Measurement collapses the qubit into a computational-basis state. Repeating an "
                "identically prepared experiment builds an empirical distribution."
            ),
            "gates": (
                "Quantum gates are unitary operations (for example X, H, CNOT). They rearrange "
                "amplitudes without reading the state as a classical bit."
            ),
            "entanglement": (
                "Entangled multi-qubit states cannot be written as independent single-qubit states. "
                "A Bell pair is the simplest textbook example."
            ),
            "circuits": (
                "A quantum circuit is a sequence of gates and measurements on qubits, often "
                "followed by classical post-processing of bitstrings."
            ),
        },
    }


def superposition_demo(*, shots: int = 512, seed: int | None = 42) -> dict[str, Any]:
    """Hadamard on one qubit → roughly 50/50 measurement distribution."""
    shots = require_shots(shots)
    seed = require_seed(seed)

    qc = QuantumCircuit(1, 1)
    qc.h(0)
    qc.measure(0, 0)

    sim = AerSimulator(seed_simulator=seed) if seed is not None else AerSimulator()
    result = sim.run(transpile(qc, sim), shots=shots).result()
    counts = {k: int(v) for k, v in result.get_counts().items()}

    return {
        "experiment": "superposition_hadamard",
        "limitations": LIMITATIONS_BANNER,
        "description": "Apply H to |0⟩ and measure. Expect counts near 50/50 for large shots.",
        "shots": shots,
        "seed": seed,
        "circuit": circuit_metadata(qc),
        "counts": counts_to_sorted_dict(counts),
    }


def entanglement_bell_demo(*, shots: int = 512, seed: int | None = 42) -> dict[str, Any]:
    """Create and measure a Bell pair (|00⟩+|11⟩)/√2."""
    shots = require_shots(shots)
    seed = require_seed(seed)

    qc = QuantumCircuit(2, 2)
    qc.h(0)
    qc.cx(0, 1)
    qc.measure([0, 1], [0, 1])

    sim = AerSimulator(seed_simulator=seed) if seed is not None else AerSimulator()
    result = sim.run(transpile(qc, sim), shots=shots).result()
    counts = {k: int(v) for k, v in result.get_counts().items()}

    return {
        "experiment": "bell_pair",
        "limitations": LIMITATIONS_BANNER,
        "description": "Bell state preparation. Ideal outcomes concentrate on 00 and 11.",
        "shots": shots,
        "seed": seed,
        "circuit": circuit_metadata(qc),
        "counts": counts_to_sorted_dict(counts),
    }


def gate_tour(*, gate: str = "h") -> dict[str, Any]:
    """Return a tiny single-gate circuit for visualization."""
    gate = gate.lower().strip()
    qc = QuantumCircuit(1, 1)
    if gate == "h":
        qc.h(0)
    elif gate == "x":
        qc.x(0)
    elif gate == "z":
        qc.z(0)
    else:
        raise ValidationError("gate must be one of: h, x, z")
    qc.measure(0, 0)
    return {
        "experiment": "single_gate",
        "gate": gate,
        "limitations": LIMITATIONS_BANNER,
        "circuit": circuit_metadata(qc),
    }
