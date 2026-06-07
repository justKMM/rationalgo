# RationAlgo — Mission Control (Frontend)

A premium, fintech-grade mission-control UI for **RationAlgo**, an Algorand-native governance and transparency layer for **agentic commerce**. Every spend an AI agent proposes via x402 is reasoned, policy-checked, anchored on Algorand, executed, verified, and re-anchored. This frontend visualises that pipeline live, in a way a hackathon judge can grok in 30 seconds.

> Status: **connected to the Go backend**. Mission Control hydrates from `GET /api/state` and
> runs the hero demo via `POST /api/scenario/run` (SSE). Client-side mocks remain as a
> fallback when `VITE_USE_API=false`.

---

## 1. The Product Story (30-second version)

```
AI Agent  →  Policy Layer  →  Algorand Provenance  →  x402 Payment  →  Outcome Verification
   │             │                    │                    │                  │
 reasoning   guardrails            tx (pre)             settlement          tx (post)
```

Six discrete stages, every one observable, every one auditable on-chain. The UI is intentionally **not an analytics dashboard** — it is a SOC/operations console for agentic spend.

---

## 2. Tech Stack

| Concern | Choice | Reason |
|---|---|---|
| Framework | **TanStack Start v1** (React 19 + Vite 7) | File-based routing, SSR via Nitro; standard public npm packages only. |
| Styling | **Tailwind v4** via `src/styles.css` (`@theme`, native CSS vars) | No `tailwind.config.js`; design tokens are CSS variables in `oklch` / hex, easy to theme. |
| UI primitives | **shadcn/ui** (Radix) | Already vendored under `src/components/ui/`. We only wrap what we need (Sheet for the drawer). |
| State | **Zustand** (`src/hooks/useMissionStore.ts`) | Live mission state; `runScenario` streams backend SSE or falls back to mock timers. |
| Animation | **framer-motion** | Stage transitions, reasoning feed slide-ins, decision card swaps. |
| Icons | **lucide-react** | Consistent stroke weight matches the Linear/Stripe aesthetic. |
| Fonts | Inter (body), JetBrains Mono (IDs, hashes, timestamps, costs) | Mono everywhere a value is technical/auditable. |
| Package manager | **npm** (`package-lock.json`) | Public registry only; Node **≥ 20.19**. After install, run `npm approve-scripts esbuild` if prompted. |

### Why TanStack Start (and not plain Vite + React Router)
The product needs server-side things later: signature verification for webhooks, x402 callback receivers, and authenticated reads of decision history. TanStack Start gives us file-based routing, SSR, and server functions in one Vite project — configured with public packages (`@tanstack/react-start`, `nitro`, `@vitejs/plugin-react`) in `vite.config.ts`.

**Node:** TanStack Start 1.16x expects **Node ≥ 20.19** (22.12+ recommended). Upgrade if `npm run dev` fails engine checks.

---

## 3. Repo Layout

```
frontend/
  package.json                 # dependencies + scripts
  package-lock.json            # npm lockfile — commit this; use npm only
  vite.config.ts               # TanStack Start + Vite (public npm plugins)
src/
  routes/
    __root.tsx                 # shell, head, providers, <Outlet/>
    index.tsx                  # the entire Mission Control screen
    api/                       # (reserved) HTTP handlers — webhooks, x402 callbacks
  components/
    mission-control/
      TopBar.tsx               # brand, api-live dot, Execute Flow / Anomaly / Reset
      ReasoningFeed.tsx        # left column — chronological agent events (scrollable)
      ActiveDecisionCard.tsx   # center column — current decision, KV rows, policy checks
      TrustPipeline.tsx        # right column — 6-stage vertical ladder
      TrustMetrics.tsx         # 4 KPI tiles strip
      DecisionHistoryTable.tsx # dense audit table (fixed height, scrollable)
      DecisionDetailsDrawer.tsx# slide-out audit trail for one decision
      ConfidenceMeter.tsx      # 0–1 bar
      StatusPill.tsx           # approved / blocked / verified / pending / failed
      TxHash.tsx               # truncated mono hash + copy + explorer link
    ui/                        # shadcn primitives — Sheet used by drawer; rest vendored
  hooks/
    useMissionStore.ts         # Zustand store — SSE scenario orchestration + API hydrate
  lib/
    types.ts                   # DecisionRecord, ScenarioEvent, PIPELINE_STAGES
    api.ts                     # fetch helpers: health, state, reset, scenario SSE stream
    mapBackend.ts              # backend DecisionRecord / Decision → UI DecisionRecord
    mock/
      generators.ts            # seeded PRNG, algoTx(), shortId()
      decisions.ts             # seedHistory() — used when VITE_USE_API=false
      scenarios.ts             # buildScenario('normal' | 'anomaly') — mock fallback
    config.server.ts           # server-only env reads (for future createServerFn handlers)
  styles.css                   # Tailwind v4 @theme tokens, panel heights, hairline utilities
```

---

## 4. The Domain Model (source of truth: `src/lib/types.ts`)

```ts
DecisionRecord {
  id: string                                  // e.g. "DEC-A1B2C3"
  timestamp: string                           // ISO 8601, UTC
  task: string                                // human-readable intent
  vendor: string                              // counterparty (e.g. "OpenAI API")
  cost: number                                // USD
  confidence: number                          // 0..1
  policyStatus: "approved" | "blocked"
  outcomeStatus: "verified" | "failed" | "pending"
  reasoningSummary: string
  txPre?: string                              // Algorand tx hash, 52-char base32
  txOutcome?: string                          // Algorand tx hash, 52-char base32
  policyChecks?: PolicyCheck[]                // budget, allow-list, rate, risk
  predictedOutcome?: string
  actualOutcome?: string
}

ScenarioEvent {
  id, timestamp,
  type: "agent.thinking" | "decision.pending"
      | "policy.approved" | "policy.blocked"
      | "decision.committed" | "decision.blocked"
      | "payment.sent" | "decision.outcome"
      | "alert.fired" | "research.summary"
  message: string
}

PIPELINE_STAGES (ordered): reasoning → policy → commit-pre → payment → verify → commit-post
```

**Do not break this shape.** The UI binds against these fields; `mapBackend.ts` adapts the Go API.

---

## 5. Live Workflow (how the 6 stages animate)

### Default path — backend SSE

When the API is reachable (`VITE_USE_API` not `false`, backend `serve` running):

1. **Mount** — `TopBar` calls `hydrate()` → `GET /health` + `GET /api/state` → populates `history`.
2. **Execute Flow** — `runScenario('normal')` → `POST /api/scenario/run` (streaming fetch body).
3. Each SSE frame is `{ "type": "<event>", "payload": … }` from `backend/internal/scenario/hero.go`.
4. `handleBackendEvent` in `useMissionStore.ts`:
   - **Advance `pipelineStage`** (0…6) — drives `TrustPipeline`.
   - **Push a `ScenarioEvent`** — appears in `ReasoningFeed`.
   - **Patch `activeDecision`** — e.g. set `txPre`, flip `outcomeStatus`.
5. Each completed purchase (`decision.outcome` or `alert.fired` after `decision.blocked`) is prepended to `history` and metrics are recomputed.
6. **`research.summary`** marks the run complete; the store re-syncs from `GET /api/state`.

**Anomaly** is the same flow with `POST /api/scenario/run?scenario=anomaly` (first knapsack pick gets a 10× price injection → policy block).

**Reset** clears live UI state and calls `POST /api/state/reset`, then reloads history.

### Fallback path — client mock

When `VITE_USE_API=false`, `runScenario` uses `buildScenario()` + `setTimeout` (see `src/lib/mock/scenarios.ts`). Same UI mutations, no network.

### Mock scenarios (fallback only)

- **`normal`** — scripted OpenAI purchase, ~5.4s.
- **`anomaly`** — scripted Twilio block at policy stage, ~2.7s.

The live hero demo (backend) researches **Atlas Robotics GmbH** with a budgeted knapsack over RationAlgo's `/company/*` marketplace — see the root [`README.md`](../README.md#hero-demo).

---

## 6. Determinism & SSR

A pitfall fixed early: `Math.random()` and `Date.now()` at module scope produce **different** values on server vs client → React hydration mismatch.

Mitigations (already in place — keep them):
- `src/lib/mock/generators.ts` exposes `makeRng(seed)` (mulberry32) and a `seededRng()` helper. All seed data uses fixed seeds.
- `src/lib/mock/decisions.ts` uses a fixed `BASE_EPOCH` (UTC) for timestamps.
- Scenario runner only seeds from `Date.now()` **inside the click handler** (client-only, never SSR).

If you add new mock data, follow the same pattern.

---

## 7. Design System (tokens live in `src/styles.css`)

- Background `#0B0D12`, Surface `#12151C`, Surface-2 (hover) slightly lighter, Border `#20242D`.
- Foreground near-white, muted-foreground mid-grey.
- Status colors: success `#10B981`, warn `#F59E0B`, danger `#EF4444`, accent cyan reserved for "live/verified" emphasis.
- `hairline-t`, `hairline-b`, `hairline-r` utilities for 1px dividers — used heavily for the Stripe/Linear density.
- `mono-meta` utility for tiny uppercase mono labels.

### Panel layout (fixed height + scroll)

List panels keep a stable footprint during long hero runs — content scrolls inside the box instead of stretching the page.

| Token / utility | Default | Used by |
|---|---|---|
| `--panel-workspace-h` / `panel-workspace-h` | 480px (600px on `lg+`) | Agent Activity, Current Decision, Trust Pipeline |
| `--panel-history-h` / `panel-history-h` | 360px (400px on `lg+`) | Decision History table |
| `panel-scroll-body` | `flex: 1; overflow-y: auto` | Scrollable body inside any panel |

Tune heights in one place: `:root` and the `lg` media query in `src/styles.css`.

**Rule:** never hardcode colors in components — use the tokens. New colors go in `src/styles.css` first.

---

## 8. Backend Integration (implemented)

The frontend talks to the **Go backend** at `VITE_API_URL` (default `http://localhost:8080`). CORS is enabled on the backend for local dev.

### 8.1 API surface used today

| Method | Path | Purpose |
|---|---|---|
| GET | `/health` | Liveness — top bar **api live** / **offline** |
| GET | `/api/state` | Hydrate `history` on mount and after each scenario run |
| POST | `/api/state/reset` | Reset server seed data; reload history |
| POST | `/api/scenario/run` | Hero demo SSE stream (normal) |
| POST | `/api/scenario/run?scenario=anomaly` | Hero demo with price anomaly on first purchase |

Implementation: `src/lib/api.ts` (`runScenarioStream` parses SSE `data:` lines from a POST response body — **not** `EventSource`, which only supports GET).

### 8.2 SSE event mapping

Backend events (`scenario/hero.go`) → UI updates (`useMissionStore.ts`):

| Backend `type` | UI effect |
|---|---|
| `agent.thinking` | Stage 1; reasoning feed message |
| `decision.pending` | Set `activeDecision` from payload (`mapBackendRecord`) |
| `decision.committed` | Stage 3; patch `txPre`; synthesize `policy.approved` feed event |
| `payment.sent` | Stage 4; x402 settlement message |
| `decision.outcome` | Stage 5–6; patch outcome + `txOutcome`; finalize to `history` |
| `decision.blocked` | Stage 2; patch `policyStatus: blocked` |
| `alert.fired` | Feed alert; finalize blocked decision to `history` |
| `research.summary` | Stage 6; terminal summary message; `running: false` |

Payload shapes follow `backend/internal/models/decision.go` (`DecisionRecord` with snake_case JSON). Mapping lives in `src/lib/mapBackend.ts`.

### 8.3 Algorand specifics

- `txPre` ← `committed_tx`; `txOutcome` ← `outcome_tx` (52-char base32 tx ids).
- `TxHash` truncates and links to Allo explorer; override base via `VITE_ALGORAND_EXPLORER_BASE` if needed.
- Stages 3 and 6 correspond to pre-payment and post-outcome commits.

### 8.4 x402 specifics

- `payment.sent` fires after `PayAndFetch` completes (includes `settlement_tx` when on-chain).
- Blocked purchases never receive `payment.sent` — policy halts before x402.
- Real payments require a funded testnet wallet with USDC ASA `10458941` in `backend/.env`.

### 8.5 Key files

| Concern | File |
|---|---|
| HTTP + SSE client | `src/lib/api.ts` |
| Backend → UI types | `src/lib/mapBackend.ts` |
| Store orchestration | `src/hooks/useMissionStore.ts` |
| Execute Flow button | `src/components/mission-control/TopBar.tsx` |
| Panel heights / scroll | `src/styles.css` (`--panel-workspace-h`, `--panel-history-h`, `panel-scroll-body`) |

### 8.6 Environment variables

| Name | Where | Purpose |
|---|---|---|
| `VITE_API_URL` | client | Backend base URL (default `http://localhost:8080`) |
| `VITE_USE_API` | client | Set to `false` to use mock timers instead of backend |
| `VITE_ALGORAND_EXPLORER_BASE` | client | Optional explorer URL prefix for `TxHash` links |

Future webhook routes (`/api/public/x402/callback`, etc.) remain reserved under `src/routes/api/` — not used by the hero demo today.

---

## 9. Local Dev

**Terminal 1 — backend** (required for live demo):

```bash
cd backend
go run ./cmd/rationalgo serve    # :8080
```

**Terminal 2 — frontend**:

```bash
cd frontend
npm install
npm approve-scripts esbuild   # if npm warns about pending install scripts
npm run dev          # Vite + TanStack Start — http://localhost:3000
```

Uses **npm** only (`package-lock.json`). Requires Node **≥ 20.19**. If install or dev fails with missing modules, delete `node_modules` and run `npm install` again.

Open the preview. The top bar should show **api live** when the backend is reachable.
Click **Execute Flow** for the hero demo (Atlas Robotics knapsack research + real x402).
**Anomaly** runs the blocked-purchase variant. **Reset** clears live UI state and resets
server seed data.

To develop UI without the backend: set `VITE_USE_API=false` — mock timers and seeded
history are used instead.

---

## 10. Non-Goals (deliberate)

- No auth flows, no user management — the product is single-tenant operator UI for the demo.
- No charts, no analytics dashboard look — this is operations, not BI.
- No additional routes beyond `/` and the decision drawer — everything (live feed, active decision, pipeline, history) is on one Mission Control screen.
- No real Algorand / x402 SDK usage in the frontend. That belongs behind server functions.

---

## 11. Quick Map: "Where do I change X?"

| Want to change… | File |
|---|---|
| Color tokens / typography / panel heights | `src/styles.css` |
| Top bar actions or api-live indicator | `src/components/mission-control/TopBar.tsx` |
| Add a stage to the pipeline | `src/lib/types.ts` (`PIPELINE_STAGES`) + `TrustPipeline.tsx` |
| Add a new event type | `src/lib/types.ts` (`ScenarioEventType`) + `ReasoningFeed.tsx` (icon/tone) |
| Wire backend / change event handling | `src/hooks/useMissionStore.ts`, `src/lib/api.ts`, `src/lib/mapBackend.ts` |
| Add a webhook | `src/routes/api/public/<name>.ts` (createFileRoute with `server.handlers`) |
| Add an authenticated RPC | colocate as `*.functions.ts`, call with `useServerFn` |

---

Built for the hackathon. Designed to outlive it.
