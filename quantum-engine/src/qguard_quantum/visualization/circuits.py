"""Circuit metadata helpers for reproducible educational outputs."""

from __future__ import annotations

from typing import Any

from qiskit import QuantumCircuit


def circuit_metadata(circuit: QuantumCircuit) -> dict[str, Any]:
    """Extract reproducible metadata from a QuantumCircuit."""
    gate_counts: dict[str, int] = {}
    for inst in circuit.data:
        name = inst.operation.name
        gate_counts[name] = gate_counts.get(name, 0) + 1

    depth = circuit.depth()
    try:
        qasm = circuit.qasm(formatted=True)
    except Exception:
        # Qiskit 1.x may use different exporters depending on version.
        try:
            from qiskit.qasm2 import dumps

            qasm = dumps(circuit)
        except Exception:
            qasm = None

    return {
        "num_qubits": circuit.num_qubits,
        "num_clbits": circuit.num_clbits,
        "depth": depth,
        "size": circuit.size(),
        "gate_counts": dict(sorted(gate_counts.items())),
        "qasm": qasm,
        "draw_text": str(circuit.draw(output="text")),
    }


def counts_to_sorted_dict(counts: dict[str, int]) -> dict[str, int]:
    """Sort measurement counts by bitstring for stable JSON."""
    return dict(sorted(counts.items(), key=lambda kv: (-kv[1], kv[0])))
