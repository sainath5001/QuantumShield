# Q-Guard troubleshooting

## Backend will not start

| Symptom | Likely cause | Fix |
| --- | --- | --- |
| `invalid configuration` / missing RPC or key | Empty `.env` | Copy `.env.example` → `.env` and set `QGUARD_ETH_RPC_URL`, `QGUARD_ETHERSCAN_API_KEY` |
| `QGUARD_CHAIN_ID must be 1` | Non-mainnet | Set `QGUARD_CHAIN_ID=1` |
| Port already in use | Another process on 8080 | Stop it or change `QGUARD_SERVER_PORT` |

## `/ready` returns not ready (503)

- RPC URL unreachable, wrong network, or chain ID ≠ 1.
- Fix RPC endpoint; check `curl` to your provider.
- Etherscan showing `configured` (not `ok`) is **expected** — `/ready` does not probe Etherscan.

## Scans fail or hang

| Symptom | Fix |
| --- | --- |
| `RATE_LIMITED` / too many concurrent scans | Wait; reduce parallel clients; check inbound + Etherscan limits |
| `CAPACITY_EXCEEDED` | Restart API or wait for TTL eviction |
| Timeout / provider error | Check RPC + Etherscan quotas; raise `QGUARD_SCAN_TIMEOUT` carefully |
| `partial` / `rate_limited` coverage | Expected under caps — not a silent full history |
| Report disappears after restart | In-memory store — by design |

## Frontend cannot reach API

1. Confirm Go API: `curl -s http://127.0.0.1:8080/api/v1/health`
2. Check `frontend/.env.local` → `NEXT_PUBLIC_QGUARD_API_BASE_URL=http://127.0.0.1:8080`
3. Restart `npm run dev` after changing env (Next inlines `NEXT_PUBLIC_*` at start)
4. CORS: `QGUARD_CORS_ORIGINS` must include the UI origin (`http://localhost:3000` or `http://127.0.0.1:3000`)
5. Ensure `NEXT_PUBLIC_QGUARD_USE_MOCKS=false` for live demos

## Quantum Lab unavailable

1. Start sidecar: `./scripts/run-quantum.sh`
2. `curl -s http://127.0.0.1:8090/health`
3. `NEXT_PUBLIC_QGUARD_QUANTUM_API_BASE_URL=http://127.0.0.1:8090`
4. First run may install Qiskit (slow); subsequent starts are faster
5. Validation errors (bad `n` / qubits / shots) return HTTP 400 — adjust inputs within educational caps

## “Unknown” / “requires manual review” findings

These are **valid** outcomes, not UI bugs:

- Contract accounts → manual review (EOA recovery rules not applied as if EOA)
- Incomplete history → insufficient evidence / unknown risk — **not** “safe”
- Unsupported tx types counted separately

## Go / Python / Node test failures

```bash
cd backend && go test ./... -count=1
cd quantum-engine && source .venv/bin/activate && pytest -vv
cd frontend && npm test
```

Network-dependent integration tests are not required for unit suites; live smoke uses `./scripts/integration-smoke.sh` with services up.

## Security reminders during demos

- Do not paste private keys into the UI or `.env`.
- Do not claim a wallet is quantum-safe from a negative scan.
- Do not treat public-key exposure as proof of present-day compromise.
