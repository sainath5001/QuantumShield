# Shor demonstration guide

## Purpose

Demonstrate the **structure** of Shor-style order finding on the textbook integer **N = 15**, using a small Qiskit circuit and classical continued-fraction post-processing.

## What this does not show

- Factoring `15` is **not** equivalent to deriving an Ethereum private key.
- This is **not** a practical attack against **secp256k1**.
- Cryptographically large discrete-log / factoring instances are far beyond this simulator demo.

## Inputs

| Parameter | Educational values |
| --- | --- |
| `n` | `15` only |
| `a` | Coprime bases with modular-multiplication gadgets: `2, 4, 7, 8, 11, 13` |
| `counting_qubits` | `4` or `8` |
| `shots` / `seed` | Simulator controls for reproducibility |

Unsupported `n` values raise a validation error (by design).

## Pipeline

1. Build period-finding circuit (Hadamards on counting register, controlled `a^{2^q} mod 15` on work qubits, inverse QFT, measure counting register).
2. Simulate with AerSimulator.
3. Map measured phases → candidate periods via continued fractions.
4. Classical check: if even period `r` yields non-trivial factors of `n`, report success.

## Output fields

- `input` — n, a, counting qubits, shots, seed
- `circuit` — qubit count, depth, gate info, text diagram
- `transpiled_depth` / `transpiled_size`
- `counts` — measurement histogram
- `classical_postprocessing` — candidate periods and factor attempts
- `success` / `discovered_factors`
- Explicit disclaimers

## Run

```bash
qguard-quantum shor --n 15 --a 7 --counting-qubits 8 --shots 1024 --seed 42
```

```python
from qguard_quantum.shor import run_shor_demo
run_shor_demo(n=15, a=7, seed=42)
```
