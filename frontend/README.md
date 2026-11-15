# Q-Guard Frontend

Next.js App Router UI for the Q-Guard Ethereum quantum threat scanner.

## Stack

- Next.js 15 (App Router) + TypeScript
- Tailwind CSS + shadcn-style primitives
- Lucide icons, Recharts, React Hook Form, Zod
- Vitest + Testing Library

## Pages

| Route | Purpose |
| --- | --- |
| `/` | Overview dashboard (session stats, recent scans, quick scanner) |
| `/scanner` | Wallet scanner with progress polling |
| `/wallet/[address]` | Wallet analysis (risk, exposure, evidence) |
| `/reports` | Session report index |
| `/reports/[id]` | Full security report + JSON export |
| `/quantum` | Quantum Research Lab (educational; no fake sims) |
| `/learn` | Research & learn explainers |

## Setup

From repo root (recommended):

```bash
../scripts/setup-local.sh
../scripts/run-frontend.sh
```

Or manually:

```bash
cd frontend
cp .env.example .env.local
npm install
npm run dev -- --hostname 127.0.0.1 --port 3000
```

Open [http://127.0.0.1:3000](http://127.0.0.1:3000). Start the Go API (`../scripts/run-api.sh`) and optionally the quantum sidecar. Full stack: [`docs/local-setup.md`](../docs/local-setup.md).

### Environment

| Variable | Purpose |
| --- | --- |
| `NEXT_PUBLIC_QGUARD_API_BASE_URL` | Go API base URL (default `http://127.0.0.1:8080`) |
| `NEXT_PUBLIC_QGUARD_QUANTUM_API_BASE_URL` | Quantum FastAPI sidecar (default `http://127.0.0.1:8090`) |
| `NEXT_PUBLIC_QGUARD_USE_MOCKS` | `true` only for isolated UI work — keep `false` for real API flows |

**Never** put Etherscan or RPC API keys in frontend env vars.

## Scripts

```bash
npm run lint
npm run typecheck
npm test
npm run build
```

## API contract notes

Frontend types follow **runtime Go domain JSON** (Prompt 5), which differs slightly from older examples in `docs/api-specification.md`:

- Scan job id field: `id` (not `scan_id` on the job itself; reports use `scan_id`)
- Nested `exposure` + `risk` on completed jobs (not a single `assessment` object)
- Flow: `POST /api/v1/scan` → poll `GET /api/v1/scan/:id` → `GET /api/v1/reports/:report_id`

Session metrics / recent scans are stored in **browser sessionStorage** because the backend has no list-scans endpoint and storage is in-memory.

## Integration (Prompt 8)

See [`docs/integration.md`](../docs/integration.md). Local startup: `../scripts/run-local.sh`.

Scan polling: every 2s, 180s timeout. Quantum Lab calls the FastAPI sidecar (no fabricated histograms).

## Product quality (MVP)

- Empty states on overview/reports when the session has no scans
- Live API data when mocks are disabled (default)
- History coverage and unknown/manual-review findings shown without “safe” framing
- Quantum Lab surfaces educational limitations; no fabricated histograms
- Responsive layout (tables scroll on narrow viewports)

## Remaining follow-ups

- Optional SSE for scan progress (polling works for MVP)
- Persist session scan index across devices (requires backend list endpoint)
- Upgrade Next.js / zod after regression suite (`docs/security-review.md` L-06)

## Design

Dark charcoal surfaces, restrained cyan accents, subtle purple only for quantum-education surfaces. No wallet-connect, no fake live chain metrics, no fabricated simulation histograms.
