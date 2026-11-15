"""Fundamentals circuit and explanation tests."""

from __future__ import annotations

from qguard_quantum.fundamentals.demos import (
    entanglement_bell_demo,
    explain_fundamentals,
    gate_tour,
    superposition_demo,
)
from qguard_quantum.validation import ValidationError
import pytest


def test_explain_fundamentals_topics() -> None:
    data = explain_fundamentals()
    assert "qubit" in data["topics"]
    assert "entanglement" in data["topics"]
    assert "limitations" in data


def test_superposition_demo_structure() -> None:
    out = superposition_demo(shots=256, seed=42)
    assert out["experiment"] == "superposition_hadamard"
    assert "counts" in out
    assert out["circuit"]["num_qubits"] == 1
    assert set(out["counts"].keys()) <= {"0", "1"}
    assert sum(out["counts"].values()) == 256


def test_bell_demo_prefers_correlated_outcomes() -> None:
    out = entanglement_bell_demo(shots=512, seed=42)
    assert out["circuit"]["num_qubits"] == 2
    correlated = out["counts"].get("00", 0) + out["counts"].get("11", 0)
    assert correlated / 512 > 0.9


def test_gate_tour_and_invalid() -> None:
    out = gate_tour(gate="x")
    assert out["gate"] == "x"
    assert "draw_text" in out["circuit"]
    with pytest.raises(ValidationError):
        gate_tour(gate="cx")
