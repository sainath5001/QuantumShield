"""Educational Grover demo tests (simulator only)."""

from __future__ import annotations

import pytest

from qguard_quantum.grover.demo import (
    MAX_QUBITS,
    build_grover_circuit,
    optimal_iterations,
    run_grover_demo,
    validate_unsupported_size,
)
from qguard_quantum.validation import ValidationError


def test_circuit_construction() -> None:
    qc = build_grover_circuit(2, "11", iterations=1)
    assert qc.num_qubits == 2
    assert qc.num_clbits == 2
    assert qc.depth() > 0


def test_optimal_iterations() -> None:
    assert optimal_iterations(1) >= 1
    assert optimal_iterations(2) == 1


def test_run_grover_finds_target_with_high_probability() -> None:
    out = run_grover_demo(n_qubits=2, target="11", shots=1024, seed=42)
    assert out["input"]["search_space_size"] == 4
    assert out["input"]["target"] == "11"
    assert sum(out["counts"].values()) == 1024
    assert out["target_probability_estimate"] > 0.7
    assert out["most_probable"] == "11"
    assert "classical_comparison" in out
    assert "Ethereum" in out["disclaimer"] or "wallet" in out["disclaimer"].lower()


def test_invalid_target_width() -> None:
    with pytest.raises(ValidationError):
        build_grover_circuit(2, "1")
    with pytest.raises(ValidationError):
        run_grover_demo(n_qubits=2, target="abc")


def test_unsupported_size() -> None:
    with pytest.raises(ValidationError):
        require = validate_unsupported_size
        require(MAX_QUBITS + 1)
    with pytest.raises(ValidationError):
        run_grover_demo(n_qubits=MAX_QUBITS + 1, target="0" * (MAX_QUBITS + 1))
