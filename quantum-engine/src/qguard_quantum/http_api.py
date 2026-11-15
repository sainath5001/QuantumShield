"""Thin FastAPI sidecar exposing qguard_quantum.interface over localhost JSON.

Justified for Prompt 8 frontend integration only. Does not duplicate Go scan logic
and must never feed wallet-scan evidence.
"""

from __future__ import annotations

from typing import Any

from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel, Field

from qguard_quantum import LIMITATIONS_BANNER, __version__
from qguard_quantum.interface import module_info, run_experiment
from qguard_quantum.validation import ValidationError

app = FastAPI(
    title="Q-Guard Quantum Research API",
    version=__version__,
    description=(
        "Educational quantum experiments (Qiskit). "
        "Not connected to Ethereum private keys or scan evidence."
    ),
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=[
        "http://localhost:3000",
        "http://127.0.0.1:3000",
    ],
    allow_credentials=False,
    allow_methods=["GET", "POST", "OPTIONS"],
    allow_headers=["*"],
)

_BULKY = ("qasm", "draw_text")


def _compact(payload: Any) -> Any:
    if isinstance(payload, dict):
        return {k: _compact(v) for k, v in payload.items() if k not in _BULKY}
    if isinstance(payload, list):
        return [_compact(x) for x in payload]
    return payload


class ShorRequest(BaseModel):
    n: int = 15
    a: int = 7
    counting_qubits: int = Field(default=8, ge=4, le=8)
    shots: int = Field(default=512, ge=1, le=4096)
    seed: int | None = 42
    full: bool = False


class GroverRequest(BaseModel):
    n_qubits: int = Field(default=2, ge=1, le=4)
    target: str = "11"
    iterations: int | None = None
    shots: int = Field(default=512, ge=1, le=4096)
    seed: int | None = 42
    full: bool = False


@app.get("/health")
def health() -> dict[str, Any]:
    return {
        "status": "ok",
        "service": "qguard-quantum",
        "version": __version__,
        "feeds_production_scan_evidence": False,
        "limitations": LIMITATIONS_BANNER,
    }


@app.get("/info")
def info() -> dict[str, Any]:
    return module_info()


@app.get("/demos")
def list_demos() -> dict[str, Any]:
    return {
        "limitations": LIMITATIONS_BANNER,
        "feeds_production_scan_evidence": False,
        "demos": [
            {
                "id": "shor",
                "name": "Shor period-finding toy",
                "path": "POST /demos/shor",
                "notes": "N=15 educational only; not a secp256k1 attack.",
            },
            {
                "id": "grover",
                "name": "Grover small search",
                "path": "POST /demos/grover",
                "notes": "≤4 qubits; not an Ethereum wallet cracker.",
            },
            {
                "id": "fundamentals",
                "name": "Quantum fundamentals explainer",
                "path": "GET /demos/fundamentals",
            },
            {
                "id": "pqc",
                "name": "Post-quantum cryptography overview",
                "path": "GET /explainers/post-quantum",
            },
        ],
    }


@app.get("/demos/fundamentals")
def fundamentals() -> dict[str, Any]:
    return run_experiment("fundamentals")


@app.post("/demos/shor")
def shor(req: ShorRequest) -> dict[str, Any]:
    try:
        result = run_experiment(
            "shor",
            n=req.n,
            a=req.a,
            counting_qubits=req.counting_qubits,
            shots=req.shots,
            seed=req.seed,
        )
    except ValidationError as exc:
        raise HTTPException(status_code=400, detail=str(exc)) from exc
    except Exception as exc:  # noqa: BLE001
        raise HTTPException(status_code=500, detail=f"Experiment failed: {exc}") from exc
    return result if req.full else _compact(result)


@app.post("/demos/grover")
def grover(req: GroverRequest) -> dict[str, Any]:
    kwargs: dict[str, Any] = {
        "n_qubits": req.n_qubits,
        "target": req.target,
        "shots": req.shots,
        "seed": req.seed,
    }
    if req.iterations is not None:
        kwargs["iterations"] = req.iterations
    try:
        result = run_experiment("grover", **kwargs)
    except ValidationError as exc:
        raise HTTPException(status_code=400, detail=str(exc)) from exc
    except Exception as exc:  # noqa: BLE001
        raise HTTPException(status_code=500, detail=f"Experiment failed: {exc}") from exc
    return result if req.full else _compact(result)


@app.get("/explainers/post-quantum")
def pqc_explainer() -> dict[str, Any]:
    return run_experiment("pqc")


def create_app() -> FastAPI:
    return app
