# Q-Guard Quantum Engine

Educational Python research module using **Qiskit** to demonstrate quantum algorithms relevant to blockchain cryptography teaching.

This module is **independent** of the Go scanner. It must never connect to real wallet private keys, never attempt secp256k1 attacks, and never feed production scan evidence.

## What this is / is not

| Is | Is not |
| --- | --- |
| Toy Shor period-finding for **N=15** | A practical factoring attack |
| Tiny Grover search (≤4 qubits) | An Ethereum wallet cracker |
| PQC education (ML-DSA, ML-KEM, SLH-DSA) | A claim that any scanned wallet is quantum-safe |
| AerSimulator experiments with fixed seeds | A substitute for Mainnet exposure analysis |

## Package layout

```text
quantum-engine/
  pyproject.toml
  README.md
  docs/
  src/qguard_quantum/
    fundamentals/   # qubits, gates, entanglement demos
    shor/           # educational N=15 period finding
    grover/         # tiny unstructured search
    pqc/            # post-quantum explainers
    visualization/  # circuit metadata helpers
    interface.py    # stable callable API for later consumers
    cli.py          # qguard-quantum CLI
  tests/
```

## Documented dependency versions

Pinned ranges in `pyproject.toml`:

| Package | Range | Role |
| --- | --- | --- |
| Python | ≥3.10 | Runtime |
| qiskit | ≥1.2.0,<2.0.0 | Circuits / gates |
| qiskit-aer | ≥0.14.0,<0.16.0 | Local simulator (no hardware) |
| numpy | ≥1.24.0,<3 | Numeric support for Qiskit |
| pytest (dev) | ≥8.0.0 | Tests |

Stable interfaces: Python functions (`qguard_quantum.interface`), CLI (`qguard-quantum`), and a **thin FastAPI sidecar** for frontend integration (`qguard-quantum-api`, default `http://127.0.0.1:8090`). The sidecar only wraps the same educational payloads — it does not talk to Ethereum or the Go scan path.

```bash
# from quantum-engine with venv activated and extras installed:
pip install -e ".[api,dev]"
../scripts/run-quantum.sh
# GET /health  POST /demos/shor  POST /demos/grover  GET /explainers/post-quantum
```

## Local setup (virtual environment)

From repo root: `../scripts/setup-local.sh` (or `../scripts/run-quantum.sh`, which creates the venv if missing).

Manual:

```bash
cd quantum-engine
python3 -m venv .venv
source .venv/bin/activate
python -m pip install --upgrade pip
pip install -e ".[dev,api]"
```

Record installed versions:

```bash
python -c "import qiskit, qiskit_aer, numpy; print('qiskit', qiskit.__version__); print('qiskit-aer', qiskit_aer.__version__); print('numpy', numpy.__version__)"
```

## CLI usage

```bash
qguard-quantum info
qguard-quantum fundamentals --demo explain
qguard-quantum fundamentals --demo superposition --shots 512 --seed 42
qguard-quantum fundamentals --demo bell
qguard-quantum shor --n 15 --a 7 --counting-qubits 8 --shots 1024 --seed 42
qguard-quantum grover --qubits 2 --target 11 --shots 1024 --seed 42
qguard-quantum pqc
qguard-quantum pqc --algorithm ML-KEM

# Include full OpenQASM + text circuit diagrams (large for Shor):
qguard-quantum --full shor --n 15 --a 7 --seed 42
```

Equivalent module entrypoint:

```bash
python -m qguard_quantum info
```

## Python interface (for a future Go/frontend bridge)

```python
from qguard_quantum.interface import module_info, run_experiment

module_info()
run_experiment("shor", n=15, a=7, shots=512, seed=42)
run_experiment("grover", n_qubits=2, target="11", seed=42)
run_experiment("pqc")
```

Keep integration simple: import or subprocess the CLI. Do not invent microservices for the MVP.

## Tests

Tests use **AerSimulator** only (deterministic seeds). They do not call quantum hardware.

```bash
cd quantum-engine
source .venv/bin/activate
pytest
```

## Documentation

- [Quantum fundamentals](docs/quantum-fundamentals.md)
- [Shor demonstration](docs/shor-demonstration.md)
- [Grover demonstration](docs/grover-demonstration.md)
- [PQC overview](docs/pqc-overview.md)
- [Known limitations](docs/known-limitations.md)

## Boundary with the Go backend

Classical Mainnet exposure / risk scoring lives in `backend/`. This package teaches quantum algorithms and PQC concepts. Do not merge a successful toy factoring run into a wallet risk report.
