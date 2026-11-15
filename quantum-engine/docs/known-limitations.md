# Known limitations

## Scope

- Educational and experimental only.
- Not connected to wallet private keys.
- Not a secp256k1 / Ethereum attack toolkit.
- Outputs must never be used as production scan evidence.

## Shor demo

- Only `N = 15` with a fixed set of bases that have textbook modular-multiplication gadgets.
- Counting register limited to 4 or 8 qubits.
- Success on factoring 15 does **not** imply ability to derive Ethereum keys.

## Grover demo

- At most 4 qubits (search space ≤ 16).
- Demonstrates amplitude amplification pedagogy, not cryptographically large key search.
- Does not imply Grover directly breaks Ethereum wallets.

## Simulation

- Uses **Qiskit Aer** local simulator only.
- No quantum hardware backends.
- Fixed seeds make tests reproducible; without a seed, counts vary.

## PQC content

- Explanatory text only — no production signature/KEM implementation.
- Demonstrating a PQC algorithm in this repo does **not** make any Ethereum wallet quantum-resistant.

## Integration

- No FastAPI/microservice in MVP.
- Interface is Python functions + CLI for later Go/frontend consumption.
- Classical Mainnet scanning remains in the Go `backend/`.

## Fabrication policy

- Measurement counts come from the simulator.
- Tests and demos must not invent fake histograms or fake factors.
