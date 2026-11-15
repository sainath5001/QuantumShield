"""Stable callable interface for later Go backend / frontend consumption.

No HTTP microservice is required for the MVP. Prefer importing these functions
or invoking the `qguard-quantum` CLI. If an HTTP API is added later, FastAPI
would be justified only to expose these same payloads over localhost JSON—
not to invent a distributed architecture.
"""

from __future__ import annotations

from typing import Any

from qguard_quantum import LIMITATIONS_BANNER, __version__
from qguard_quantum.fundamentals.demos import (
    entanglement_bell_demo,
    explain_fundamentals,
    gate_tour,
    superposition_demo,
)
from qguard_quantum.grover.demo import run_grover_demo
from qguard_quantum.pqc.overview import explain_algorithm, pqc_overview
from qguard_quantum.shor.demo import run_shor_demo


def module_info() -> dict[str, Any]:
    return {
        "name": "qguard-quantum",
        "version": __version__,
        "role": "educational_quantum_research",
        "limitations": LIMITATIONS_BANNER,
        "feeds_production_scan_evidence": False,
        "entrypoints": {
            "python": "qguard_quantum.interface",
            "cli": "qguard-quantum",
        },
    }


def run_experiment(name: str, **kwargs: Any) -> dict[str, Any]:
    """Dispatch a named educational experiment.

    Supported names: fundamentals, superposition, bell, gate, shor, grover, pqc, pqc-algo.
    """
    key = name.strip().lower()
    if key in ("fundamentals", "explain"):
        return explain_fundamentals()
    if key == "superposition":
        return superposition_demo(**kwargs)
    if key in ("bell", "entanglement"):
        return entanglement_bell_demo(**kwargs)
    if key == "gate":
        return gate_tour(**kwargs)
    if key == "shor":
        return run_shor_demo(**kwargs)
    if key == "grover":
        return run_grover_demo(**kwargs)
    if key == "pqc":
        return pqc_overview()
    if key in ("pqc-algo", "pqc_algo"):
        algo = kwargs.pop("name", kwargs.pop("algorithm", "ML-DSA"))
        return explain_algorithm(str(algo))
    from qguard_quantum.validation import ValidationError

    raise ValidationError(f"Unknown experiment {name!r}")
