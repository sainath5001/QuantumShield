# Grover demonstration guide

## Purpose

Show Grover’s unstructured search on a **tiny** space (at most 4 qubits → 16 basis states).

## What this does not show

Grover gives a **quadratic** query speedup for unstructured search. It does **not** directly break Ethereum wallets the way a cryptographically large Shor-capable machine would threaten ECC with known public keys. Do not treat a 2-qubit demo as wallet compromise.

## Concepts shown

| Concept | Meaning in this demo |
| --- | --- |
| Search-space size | `2 ** n_qubits` |
| Target item | Bitstring marked by the oracle (MSB-left, matching Qiskit count keys) |
| Oracle | Phase flip on the marked basis state |
| Diffusion | Amplitude amplification operator |
| Iterations | ≈ `(π/4)√N` for one marked item |
| Measurement distribution | Histogram over bitstrings |
| Target probability | Empirical frequency of the marked string |

## Classical comparison (intuition only)

For space size `N`:

- Classical worst case ≈ `N` queries
- Classical average ≈ `(N+1)/2`
- Grover iterations ≈ `(π/4)√N`

Oracle cost and implementation constants are omitted—this is pedagogy, not a cryptanalytic benchmark.

## Run

```bash
qguard-quantum grover --qubits 2 --target 11 --shots 1024 --seed 42
```

```python
from qguard_quantum.grover import run_grover_demo
run_grover_demo(n_qubits=2, target="11", seed=42)
```

Inputs larger than 4 qubits are rejected.
