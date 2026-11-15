# Q-Guard local MVP release checklist

**Target:** Demonstrable local MVP for blockchain developers, security auditors, and quantum-computing researchers.  
**Not claimed:** Multi-tenant production SaaS readiness, authenticated public Internet deployment, or durable evidence vault.

Use this list before a public demo or first open-source tag.

---

## Pre-demo verification

- [ ] `.env` present locally (not committed) with working Mainnet RPC + Etherscan key
- [ ] `frontend/.env.local` has public API URLs only; `USE_MOCKS=false`
- [ ] `./scripts/setup-local.sh` (or manual install) completed
- [ ] `./scripts/run-api.sh` → `/health` ok, `/ready` ready when RPC healthy
- [ ] `./scripts/run-quantum.sh` → `/health` ok
- [ ] `./scripts/run-frontend.sh` → UI loads on desktop and narrow viewport
- [ ] End-to-end scan of a known public address completes; report shows evidence + coverage + disclaimers
- [ ] Quantum Lab Shor or Grover runs and shows limitations banner
- [ ] `./scripts/test-all.sh` passes
- [ ] No real API keys in git (`git status` / `git grep`)

---

## Required environment variables

### Backend (root `.env`)

| Variable | Required |
| --- | --- |
| `QGUARD_ETH_RPC_URL` | Yes |
| `QGUARD_ETHERSCAN_API_KEY` | Yes |
| `QGUARD_CHAIN_ID` | Yes (`1`) |
| `QGUARD_CORS_ORIGINS` | Yes |
| Timeouts, page/tx caps, rate limits | Defaults OK for demo |

### Frontend (`frontend/.env.local`)

| Variable | Required |
| --- | --- |
| `NEXT_PUBLIC_QGUARD_API_BASE_URL` | Yes (default localhost:8080) |
| `NEXT_PUBLIC_QGUARD_QUANTUM_API_BASE_URL` | Yes for Quantum Lab |
| `NEXT_PUBLIC_QGUARD_USE_MOCKS` | `false` for demos |

### Quantum

| Variable | Default |
| --- | --- |
| `QGUARD_QUANTUM_HOST` / `PORT` | `127.0.0.1` / `8090` |

---

## External API dependencies

| Dependency | Use |
| --- | --- |
| Ethereum JSON-RPC (Mainnet) | Account code/balance, signed tx fetch, chain ID |
| Etherscan API v2 `txlist` | Indexed normal transaction history |

Offline / mocked unit tests do not replace live provider dependency for demos.

---

## Supported Ethereum transaction formats

| Type | Name | Recovery |
| --- | --- | --- |
| 0 | Legacy (Homestead + EIP-155) | Yes |
| 1 | EIP-2930 access list | Yes |
| 2 | EIP-1559 dynamic fee | Yes |
| 3 | EIP-4844 blob | Yes |

Unsupported / future envelopes → explicit `unsupported_type` (not confirmed exposure).

---

## Supported account types

| Kind | Behavior |
| --- | --- |
| `likely_eoa` | EOA recovery + risk path |
| `contract_account` | Manual review; no EOA-style confirmed exposure claim |
| `unknown` | Insufficient evidence path |

Account-abstraction / EIP-1271 / multisig internals are **out of scope** for automated EOA recovery.

---

## Unsupported / out of scope (MVP)

- Non-Mainnet chains
- Token transfers / internal txs as history sources
- Scan cancellation / SSE progress
- Persistent DB / multi-user auth
- WalletConnect / private key intake
- Claiming quantum safety from negative scans
- Real cryptographically relevant Shor attacks on secp256k1
- Public unauthenticated Internet exposure without edge controls

---

## Security limitations (must disclose)

- No authentication; bind to localhost for demos
- In-memory scan/report store (ephemeral; UUID = capability URL)
- Indexer trust + pagination caps; reorg resistance is limited
- npm dependency advisories may exist on pinned Next/zod — see `docs/security-review.md` L-06
- Confirmed remediations from Prompt 9: address-index correctness, concurrent-scan + HTTP rate limits, error sanitization

Full findings: [`security-review.md`](./security-review.md).

---

## Known product limitations

- Session scan list is browser `sessionStorage` (API has no list endpoint)
- History `complete_within_scope` ≠ absolute chain completeness
- Risk labels are evidence classifications, not probabilities or timelines
- Quantum sidecar is educational and resource-capped (≤4 Grover qubits, educational Shor N)

---

## Future roadmap (suggested)

1. OpenAPI artifact generated from Go types; keep `docs/api-specification.md` in sync
2. Durable storage + authenticated access for shared demos
3. Dependency upgrades (Next.js / zod) after regression suite
4. Smart-account / AA playbooks (manual-review enrichment)
5. Optional SSE for scan progress
6. Broader history sources (internal txs) with coverage honesty preserved
7. CI pipeline: `test-all.sh` on every PR

See also [`development-roadmap.md`](./development-roadmap.md).

---

## Sign-off

| Role | Question | Pass? |
| --- | --- | --- |
| Engineering | Tests green; local stack starts from docs | |
| Security | No keys in repo; limitations disclosed in UI/docs | |
| Product | Empty states honest; no fake “safe” claims | |
| Research | Quantum demos clearly educational | |

**Local MVP demonstration: ready when checklist items above are checked.**  
**Production / public Internet deployment: not ready** without auth, durable storage, and edge rate limiting.
