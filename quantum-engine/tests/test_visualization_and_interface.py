"""Visualization helpers, PQC education, and interface dispatch."""

from __future__ import annotations

import pytest
from qiskit import QuantumCircuit

from qguard_quantum.interface import module_info, run_experiment
from qguard_quantum.pqc.overview import explain_algorithm, pqc_overview
from qguard_quantum.validation import ValidationError
from qguard_quantum.visualization.circuits import circuit_metadata, counts_to_sorted_dict


def test_circuit_metadata_structure() -> None:
    qc = QuantumCircuit(1, 1)
    qc.h(0)
    qc.measure(0, 0)
    meta = circuit_metadata(qc)
    assert meta["num_qubits"] == 1
    assert meta["depth"] >= 1
    assert "h" in meta["gate_counts"]
    assert "measure" in meta["gate_counts"]
    assert isinstance(meta["draw_text"], str)
    assert meta["draw_text"]


def test_counts_sorted() -> None:
    assert list(counts_to_sorted_dict({"01": 1, "11": 5, "00": 5}).keys())[0] in ("11", "00")


def test_pqc_overview_and_ml_kem_not_signature() -> None:
    data = pqc_overview()
    assert "ML-KEM" in data["algorithms"]
    assert "not_a_signature" in data["algorithms"]["ML-KEM"]
    assert "ethereum_eoa_migration" in data
    kem = explain_algorithm("Kyber")
    assert kem["name"] == "ML-KEM"


def test_module_info_and_dispatch() -> None:
    info = module_info()
    assert info["feeds_production_scan_evidence"] is False
    fund = run_experiment("fundamentals")
    assert "topics" in fund
    pqc = run_experiment("pqc")
    assert "algorithms" in pqc
    with pytest.raises(ValidationError):
        run_experiment("not-a-real-experiment")
