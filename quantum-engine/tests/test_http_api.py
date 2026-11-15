"""FastAPI sidecar tests (simulator only; no hardware)."""

from __future__ import annotations

import pytest

fastapi = pytest.importorskip("fastapi")
httpx = pytest.importorskip("httpx")

from fastapi.testclient import TestClient  # noqa: E402

from qguard_quantum.http_api import app  # noqa: E402


@pytest.fixture
def client() -> TestClient:
    return TestClient(app)


def test_health_and_demos_list(client: TestClient) -> None:
    h = client.get("/health")
    assert h.status_code == 200
    body = h.json()
    assert body["status"] == "ok"
    assert body["feeds_production_scan_evidence"] is False

    demos = client.get("/demos")
    assert demos.status_code == 200
    ids = {d["id"] for d in demos.json()["demos"]}
    assert {"shor", "grover", "pqc", "fundamentals"} <= ids


def test_shor_success_compact(client: TestClient) -> None:
    res = client.post("/demos/shor", json={"n": 15, "a": 7, "shots": 256, "seed": 42})
    assert res.status_code == 200
    data = res.json()
    assert data["success"] is True
    assert data["discovered_factors"]["p"] * data["discovered_factors"]["q"] == 15
    assert "qasm" not in data.get("circuit", {})


def test_shor_validation_failure(client: TestClient) -> None:
    res = client.post("/demos/shor", json={"n": 21, "a": 2, "shots": 64, "seed": 1})
    assert res.status_code == 400


def test_grover_and_pqc(client: TestClient) -> None:
    g = client.post(
        "/demos/grover",
        json={"n_qubits": 2, "target": "11", "shots": 256, "seed": 42},
    )
    assert g.status_code == 200
    assert g.json()["most_probable"] == "11"

    p = client.get("/explainers/post-quantum")
    assert p.status_code == 200
    assert "ML-KEM" in p.json()["algorithms"]
