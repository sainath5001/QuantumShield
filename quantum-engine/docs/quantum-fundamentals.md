# Quantum fundamentals guide

Educational overview used by `qguard_quantum.fundamentals`.

## Classical bits versus qubits

A **classical bit** is either `0` or `1`. Reading it does not destroy the value beyond ordinary noise.

A **qubit** is described by a state α|0⟩ + β|1⟩ with |α|² + |β|² = 1. Measurement returns a classical bit `0` or `1` with those probabilities and collapses the state.

## Superposition

After a Hadamard (`H`) gate on |0⟩, the state is (|0⟩ + |1⟩)/√2. That is not a classical “both at once” bit; it is an amplitude description of measurement chances. The `superposition` demo measures this circuit many times to show an empirical ~50/50 distribution.

## Measurement

Measurement projects onto the computational basis. Repeated identical preparations produce a histogram (`counts`). Fixed simulator seeds make histograms reproducible for tests and teaching.

## Quantum gates

Gates are unitary operations. Common single-qubit examples:

- `X` — bit flip
- `Z` — phase flip
- `H` — create/remove equal superposition (in the usual basis)

Multi-qubit example: `CNOT` / `CX` entangles when applied after `H` on the control.

## Entanglement

The Bell demo prepares (|00⟩ + |11⟩)/√2. Ideal measurements concentrate on `00` and `11`. The joint state cannot be written as an independent product of two single-qubit states.

## Quantum circuits

A circuit is a sequence of gates and measurements on registers. This module extracts:

- qubit / classical-bit counts
- depth and size
- gate tallies
- text diagram (`draw_text`)
- OpenQASM when export is available

## CLI / API

```bash
qguard-quantum fundamentals --demo explain
qguard-quantum fundamentals --demo superposition --seed 42
qguard-quantum fundamentals --demo bell --seed 42
qguard-quantum fundamentals --demo gate --gate h
```
