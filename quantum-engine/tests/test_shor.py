"""Educational Shor demo tests (simulator only)."""

from __future__ import annotations

import pytest

from qguard_quantum.shor.demo import build_period_finding_circuit, classical_postprocess, run_shor_demo
from qguard_quantum.validation import ValidationError


def test_circuit_construction_metadata() -> None:
    qc = build_period_finding_circuit(15, 7, counting_qubits=8)
    assert qc.num_qubits == 12  # 8 counting + 4 work
    assert qc.num_clbits == 8
    assert qc.depth() > 0


def test_unsupported_n_and_base() -> None:
    with pytest.raises(ValidationError):
        build_period_finding_circuit(21, 2)
    with pytest.raises(ValidationError):
        build_period_finding_circuit(15, 3)
    with pytest.raises(ValidationError):
        build_period_finding_circuit(15, 7, counting_qubits=6)


def test_run_shor_demo_reproducible_structure() -> None:
    out = run_shor_demo(n=15, a=7, counting_qubits=8, shots=512, seed=42)
    assert out["experiment"] == "shor_period_finding_toy"
    assert out["input"]["n"] == 15
    assert "counts" in out
    assert isinstance(out["counts"], dict)
    assert sum(out["counts"].values()) == 512
    assert "circuit" in out
    assert out["circuit"]["num_qubits"] == 12
    assert "transpiled_depth" in out
    assert "classical_postprocessing" in out
    assert "disclaimer" in out
    assert "secp256k1" in out["disclaimer"].lower() or "Ethereum" in out["disclaimer"]
    # With textbook N=15 / a=7 and a fixed seed, factoring should succeed.
    assert out["success"] is True
    factors = out["discovered_factors"]
    assert factors is not None
    assert factors["p"] * factors["q"] == 15


def test_classical_postprocess_empty_counts() -> None:
    post = classical_postprocess({}, a=7, n=15, counting_qubits=8)
    assert post["success"] is False
    assert post["factors"] is None
