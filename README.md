# RationAlgo

**Algorand-native policy & transparency layer for agentic commerce.**

Before an AI agent pays via x402, RationAlgo commits structured reasoning on Algorand — creating a tamper-evident audit trail. After payment, outcomes are compared to predictions so humans can trust agent spending decisions.

Built for the [Algorand x402 Agentic Commerce Hackathon](https://luma.com/agentic-commerce-hack) (Infrastructure + EURQ tracks).

**Infrastructure deliverable:** [`backend/pkg/provenance/`](backend/pkg/provenance/) — the **RAv1** note-field standard (`RAv1:` pre-payment, `RAv1out:` post-outcome). See [`SPEC.md`](backend/pkg/provenance/SPEC.md).

---

## Current status

| Component | Status | Notes |
|-----------|--------|-------|
| `pkg/provenance/` (RAv1) | ✅ | Encode/decode, tests, SPEC, standalone example |
| x402 probe (`spike x402`) | ✅ | HTTP 402 from GoPlausible `/avm/weather` |
| x402 pay (`spike x402 pay`) | ✅ | Real ASA payment via GoPlausible facilitator (testnet or mainnet) |
| Algorand legacy spike (`spike algorand`) | ⚠️ | Needs valid Pera **Algorand** passphrase + funded wallet |
| RAv1 spike (`spike provenance`) | ⚠️ | Same wallet requirement; commits `RAv1:` envelope |
| Vendor catalog | ✅ | `internal/catalog/` + `services/vendor/` — 10 priced `/company/*` research endpoints |
| x402 seller (`/company/*`, `/pricing`) | ✅ | Real on-chain ASA paywall — `services/x402.Seller` + `services/research/` |
| Reasoning / policy / outcome | ✅ | Hero uses deterministic per-purchase reasoning; `POST /api/decide` uses Anthropic when key is set |
| Hero orchestrator | ✅ | `scenario/hero.go` — budgeted multi-endpoint knapsack research (normal + anomaly) |
| HTTP API | ✅ | State, decisions, scenario SSE stream |
| Dashboard → API | ✅ | Hydrates from `GET /api/state` on mount; **Execute Flow** runs hero demo |
| Real EURQ `PayAndFetch` | ✅ | Hero demo pays local `/company/*` seller; GoPlausible for spikes only |
| Frontend → scenario SSE | ✅ | `POST /api/scenario/run` streamed via `src/lib/api.ts` → `useMissionStore` |

---

## Hero demo

**Task:** *"Research Atlas Robotics GmbH within a data budget — which paid sources are worth buying?"*

Pick a budget tier in Mission Control (or pass `?budget=` on the API):

| Tier | Budget |
|------|--------|
| **Cheapass** | €5 |
| **Mid** | €10 (default) |
| **Luxury Pro VIP** | €15 |

RationAlgo hosts its own **x402-protected company-research marketplace**: 10 priced
`/company/*` endpoints (basic info, industry, top products, reviews, competitors, news
sentiment, growth rate, revenue estimate, security incidents, legal issues — $0.01 to
$1.00 each; see [Company-research marketplace](#company-research-marketplace-x402-seller)).
The agent runs a **0/1 knapsack** (`value = importance × confidence`, `score = value / price`)
to pick the best-value subset that fits the budget, then buys them one at a time — each
purchase running the full reasoning → policy → on-chain provenance commit → real x402
payment → outcome-verification pipeline, producing its own audited `DecisionRecord`.

| Flow | What happens |
|------|----------------|
| **Normal** | Knapsack orders endpoints by value/price → each approved purchase: RAv1 commit on Algorand → real on-chain x402 ASA payment to RationAlgo's own seller → confidence-vs-expectation outcome check → RAv1out commit → final `research.summary` |
| **Anomaly** | The first selected endpoint's price is injected at 10× → policy blocks **that** purchase (alert fires, no Algorand tx, no x402 call) → the agent keeps buying the rest of the plan with its remaining budget |

Trigger via API or the dashboard (**`serve` must be running** — the hero demo pays your own `/company/*` endpoints over HTTP):

```bash
go run ./cmd/rationalgo serve   # terminal 1
curl.exe -N -X POST "http://localhost:8080/api/scenario/run?budget=mid"
curl.exe -N -X POST "http://localhost:8080/api/scenario/run?budget=cheapass"
curl.exe -N -X POST "http://localhost:8080/api/scenario/run?scenario=anomaly&budget=luxury"
```

`budget` must be `cheapass`, `mid`, or `luxury`. Omit it to use the dashboard seed limit (€10).

Or start the frontend (`cd frontend && npm run dev`), choose a budget tier in the top bar, then click **Execute Flow** / **Anomaly**.

On PowerShell, use `curl.exe` (not `curl` — that's an alias for `Invoke-WebRequest` and doesn't support `-N`).

---

## Repo layout

| Path | Purpose |
|------|---------|
| `backend/pkg/provenance/` | **RAv1 standard** — importable, stdlib-only |
| `backend/cmd/rationalgo/` | CLI — `status`, `serve`, `spike` |
| `backend/internal/catalog/` | Vendor registry — derives RationAlgo's own `/company/*` x402 endpoints from `services/research` pricing |
| `backend/internal/scenario/` | Hero demo orchestrator (budgeted multi-endpoint research) + SSE events |
| `backend/internal/models/` | Unified types in `decision.go` — `DecisionRecord`, `VendorOption`, `PolicyResult`, dashboard `Decision` |
| `backend/internal/services/` | `algorand`, `x402` (client + seller), `decision`, `vendor`, `reasoning`, `policy`, `outcome`, `research` (company-research data + paywalled handlers) |
| `backend/internal/api/` | HTTP handlers (stdlib `net/http`) |
| `backend/internal/repository/` | Thread-safe in-memory store |
| `backend/internal/store/` | Dashboard seed data |
| `backend/internal/util/` | Explorer URLs, mnemonic normalization, Pera Universal Wallet (BIP39) key resolution |
| `frontend/` | React audit dashboard (npm, TanStack Start) |

---

## Architecture

```mermaid
flowchart TB
    subgraph Infra["Infrastructure deliverable"]
        RAv1[pkg/provenance RAv1]
        SPEC[SPEC.md]
    end

    subgraph CLI["cmd/rationalgo"]
        SPIKE[spike algorand / provenance / x402]
        SERVE[serve]
    end

    subgraph Backend["Go backend"]
        CAT[catalog + vendor]
        REASON[reasoning]
        POL[policy]
        OUT[outcome]
        SCEN[scenario/hero]
        ALG[algorand.Client]
        X402C[x402 client — PayAndFetch]
        RES["research: /company/* + /pricing\nbehind x402.Seller paywall"]
        API[api.Server]
        STORE[(repository.Store)]
    end

    subgraph External["External"]
        TN[Algorand Testnet]
        GP["GoPlausible x402\n(spike x402 pay only)"]
    end

    subgraph UI["React dashboard"]
        FE[index.tsx]
    end

    RAv1 --> ALG
    SPIKE --> ALG --> TN
    SPIKE --> X402C --> GP
    SERVE --> API --> SCEN
    SERVE --> API --> RES
    RES --> ALG
    SCEN --> REASON --> POL
    SCEN --> ALG
    SCEN -->|real x402 payment, self-referential| X402C -->|HTTP 402 handshake| RES
    SCEN --> STORE
    CAT --> REASON
    FE -->|GET /api/state| API
    FE -->|POST /api/scenario/run SSE| API
```

The hero demo's x402 payments are now **self-referential**: the agent pays RationAlgo's
own `/company/*` endpoints, hosted by the same backend. `GoPlausible` remains only as the
external target for the standalone `spike x402` / `spike x402 pay` integration spikes.

### Provenance on-chain (judge story)

Every approved spend writes **two** Algorand transactions:

1. **Pre-payment** — note: `RAv1:<base64url(JSON)>` (reasoning before spend)
2. **Post-outcome** — note: `RAv1out:<base64url(JSON)>` (links actual result to original tx)

Query via Algorand Indexer: `note-prefix=RAv1:` — no app database required.

Legacy Phase 0 spike still uses `RationAlgo:commit:<hash>` via `CommitHash`; new code uses `CommitProvenance` / `CommitOutcome`.

---

## Quick start

### Backend

```bash
cd backend
cp .env.example .env
# edit .env — wallet address + mnemonic (see below)
go build -o bin/rationalgo ./cmd/rationalgo
go run ./cmd/rationalgo              # config status
go run ./cmd/rationalgo spike all    # integration spikes
go run ./cmd/rationalgo serve        # HTTP API :8080 (Ctrl+C to stop)
```

Stop the server with **Ctrl+C** before starting another instance. If `:8080` is already in use, a previous `rationalgo.exe` may still be running — see [Troubleshooting](#troubleshooting).

### Provenance package (no wallet needed)

```bash
cd backend
go test ./pkg/provenance/...
go run ./pkg/provenance/example
```

### Frontend

```bash
cd frontend
npm install
npm approve-scripts esbuild   # if npm warns about pending install scripts
npm run dev
```

With `serve` running, the dashboard top bar shows **api live**. Click **Execute Flow** to
trigger the hero demo (`POST /api/scenario/run`); the UI streams SSE events live and syncs
decision history from the backend when the run completes. **Anomaly** appends
`?scenario=anomaly`. Set `VITE_USE_API=false` to fall back to client-side mock timers.

**Node ≥ 20.19** required. Frontend uses **npm** only (`package-lock.json`); do not use Bun or yarn.

---

## CLI commands

| Command | Purpose |
|---------|---------|
| `go run ./cmd/rationalgo` | Config status + spike readiness |
| `go run ./cmd/rationalgo serve` | Start HTTP API |
| `go run ./cmd/rationalgo spike algorand` | Legacy hash commit (`RationAlgo:commit:…`) |
| `go run ./cmd/rationalgo spike provenance` | RAv1 envelope commit on testnet |
| `go run ./cmd/rationalgo spike x402` | Unpaid 402 probe (GoPlausible) |
| `go run ./cmd/rationalgo spike x402 pay` | Real payment + fetch (GoPlausible facilitator) |
| `go run ./cmd/rationalgo spike x402 pay-local` | Real payment against local `/company/*` (`serve` must be running) |
| `go run ./cmd/rationalgo spike all` | All spikes in sequence |

---

## HTTP API

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/health` | Liveness (`{"status":"ok","phase":"2"}`) |
| GET | `/api/state` | Full dashboard state |
| GET | `/api/decisions` | Decision feed only |
| POST | `/api/state/reset` | Reset to seed data |
| POST | `/api/scenario/run?budget=<tier>` | SSE stream — normal hero demo; `budget` = `cheapass` (€5) \| `mid` (€10) \| `luxury` (€15) |
| POST | `/api/scenario/run?scenario=anomaly&budget=<tier>` | SSE stream — blocked purchase demo |
| POST | `/api/decide` | LLM reasoning pipeline — returns `DecisionRecord` (requires `RATIONALGO_ANTHROPIC_KEY`) |
| GET | `/pricing` | Unprotected x402 pricing discovery — `{"pricing": [...]}` |
| GET | `/company/*` | 10 x402-protected company-research endpoints (see [Company-research marketplace](#company-research-marketplace-x402-seller)) |

CORS enabled for local frontend dev (`Access-Control-Allow-Origin: *`).

---

## Company-research marketplace (x402 seller)

RationAlgo isn't just an x402 **buyer** — it hosts a real x402 **seller**. Ten priced,
on-chain-paywalled `/company/*` endpoints plus an unprotected `/pricing` discovery route
are served by [`internal/services/research/`](backend/internal/services/research/) and
protected by [`internal/services/x402.Seller`](backend/internal/services/x402/seller.go).

| Path | Price (USD) | Returns |
|------|-------------|---------|
| `/pricing` | free | Full catalog with `price_usd` / `price_microalgo` |
| `/company/basic-info?company=...` | $0.01 | Jurisdiction, company number, incorporation date |
| `/company/industry?company=...` | $0.01 | Industry tags + NAICS code |
| `/company/top-products?company=...` | $0.02 | Top product list |
| `/company/reviews-summary?company=...` | $0.10 | Average rating, sample count, highlights |
| `/company/competitors?company=...` | $0.10 | Competitor names |
| `/company/news-sentiment?company=...` | $0.15 | Sentiment score + sample headlines |
| `/company/growth-rate?company=...` | $0.20 | YoY growth estimate |
| `/company/revenue-estimate?company=...` | $0.50 | Revenue range |
| `/company/security-incidents?company=...` | $0.50 | Security incident history |
| `/company/legal-issues?company=...` | $1.00 | Legal issue history |

Every `/company/*` response is `{"data": ..., "metadata": {...}, "confidence": 0.0-1.0}`,
backed by **deterministic mock data** (seeded from the company name — same input always
produces the same output, "Determinism for demo") plus provenance fields (`source: "mock"`,
`retrieved_at`).

**Real on-chain x402 handshake:** an unpaid request gets `402 Payment Required` with a
`PAYMENT-REQUIRED` header describing an ASA transfer (testnet USDC `10458941` by default —
override with `RATIONALGO_SETTLEMENT_ASSET_ID`); the client signs and resubmits with a
`PAYMENT-SIGNATURE` header, the seller decodes the signed transaction, verifies it on-chain
(asset id, receiver, amount, signature present, not replayed), broadcasts + confirms it via
`algorand.Client.SubmitSignedTxn`, and only then serves the data with a `PAYMENT-RESPONSE`
header.

**Demo simplification:** the seller and the buying agent share the **same configured
wallet** (`RATIONALGO_WALLET_ADDRESS` / `RATIONALGO_MNEMONIC`) — a self-payment, just like
the existing `CommitHash` provenance pattern. In production, buyer and seller would be
different parties with separate wallets.

---

## Wallet setup

Edit `backend/.env`:

```env
RATIONALGO_WALLET_ADDRESS=<58-char Pera testnet address>
RATIONALGO_MNEMONIC=<24 or 25 words, space-separated, same account>
RATIONALGO_ALGOD_TOKEN=          # leave empty for public AlgoNode
```

**Mnemonic notes:**

- **Pera Universal Wallet** (newer): 24-word **BIP39** seed — RationAlgo derives keys at `m/44'/283'/0'/0/0` automatically.
- **Legacy Algo25 wallet**: 25-word Algorand passphrase — Pera often shows 24 words; RationAlgo auto-derives the checksum (`internal/util/mnemonic.go`).
- The address derived from the mnemonic must match `RATIONALGO_WALLET_ADDRESS`.
- Fund via the [Algorand Testnet dispenser](https://bank.testnet.algorand.network/) and opt into testnet USDC ASA `10458941` for x402 payments.

### Troubleshooting

| Error | Fix |
|-------|-----|
| `mnemonic address … does not match` | Mnemonic and address must be the **same** Pera account |
| `account info: HTTP 403` / daily quota | Switch `RATIONALGO_ALGOD_URL` (e.g. Tatum testnet algod) or wait for AlgoNode reset |
| `HTTP 429` on Tatum | Free tier is 5 req/min — wait 60s and retry |
| `account info: …` / insufficient balance | Fund via testnet dispenser |
| x402 pay fails: insufficient ASA | Opt-in + fund testnet USDC ASA `10458941`. Mainnet USDC is ASA `31566704` |
| x402 pay-local fails | Run `serve` first; fund testnet USDC ASA `10458941` |
| x402 returns 404 | Use `/avm/weather` not `/api/json` (GoPlausible probe only) |
| `payment failed with 402` / `underflow … sender amount 0` | Wallet has 0 USDC — add ASA `10458941` in Pera and fund via testnet faucet |
| `listen tcp :8080: bind: … address already in use` | Stop the old server (**Ctrl+C**) or kill the stale process: `netstat -ano \| findstr :8080` then `taskkill /PID <pid> /F`. Or set `RATIONALGO_HTTP_ADDR=:8081` in `.env`. |
| `RATIONALGO_ANTHROPIC_KEY not set` | Harmless for hero demo and spikes; required only for `POST /api/decide` |

---

## How the codebase works

### `pkg/provenance/` — RAv1 standard

Stdlib-only package. `Encode` / `Decode` for pre-payment envelopes; `EncodeOutcome` / `DecodeOutcome` for post-outcome. Used by `algorand.Client.CommitProvenance` and `CommitOutcome`.

### `internal/catalog/` — vendor registry

Derives its 10 entries directly from `services/research.Pricing` — RationAlgo's own
x402-protected `/company/*` company-research endpoints (basic info, industry, top
products, reviews, competitors, news sentiment, growth rate, revenue estimate, security
incidents, legal issues; $0.01–$1.00, confidence 0.55–0.92). `services/vendor/` adapts
catalog entries to `models.VendorOption`, anchors each `EndpointURL` at
`cfg.PublicBaseURL()`, and supplies flat price history for anomaly detection.

### `internal/scenario/hero.go` — orchestrator

Runs a **budgeted multi-endpoint research flow** with 600ms delays between SSE events.
A 0/1 knapsack (`services/research.Select`) first picks the best-value subset of
`/company/*` endpoints for the demo budget; the agent then buys them one at a time:

```
agent.thinking → [per selected endpoint, in value/price order]
  decision.pending → [policy: live remaining budget + 5× price anomaly check]
    → approved: decision.committed → payment.sent (real x402 PayAndFetch + settlement_tx) → decision.outcome → store
    → blocked:  decision.blocked → alert.fired → store (continue with the rest of the plan)
→ research.summary
```

### Services

| Service | Role |
|---------|------|
| `reasoning` | `GenerateResearchDecision` — assembles a `DecisionRecord` per knapsack-selected purchase (deterministic, no API key). `GenerateDecision` — Anthropic LLM for `POST /api/decide` |
| `policy` | Budget + 5× price anomaly (`services/policy/service.go`) |
| `outcome` | Verifies the confidence a purchased endpoint actually returned vs. what the agent expected to get for the price |
| `research` | RationAlgo's own company-research data + handlers — 10 priced endpoints, deterministic mock payloads, `/pricing` discovery, knapsack selection |
| `x402` | **Client:** `RunProbe` + `PayAndFetch` (402 → sign ASA → `PAYMENT-SIGNATURE` → 200). **Seller:** `Seller.Protect` — verifies + settles on-chain ASA payments on `/company/*` |
| `algorand` | `CommitHash`, `CommitProvenance`, `CommitOutcome`, `SubmitSignedTxn` — on-chain commits and settlement via the same wallet |

### Frontend

Mission Control lives in `frontend/` (`src/routes/index.tsx`) — a single-page ops console.
On mount it hydrates from `GET /api/state`. Layout: centered **Trust Pipeline** (live stage
ladder), compact **Ops corner** (budget tier + KPIs), and **Decision History** below (click a
row for the audit drawer). The top bar shows **api live** / **offline** plus **Execute Flow**,
**Anomaly**, and **Reset**.

**Execute Flow** / **Anomaly** call `POST /api/scenario/run?budget=<tier>` (SSE over fetch —
the endpoint requires POST, not `EventSource`) and map backend events into the Zustand store
(`src/hooks/useMissionStore.ts`). Backend payloads are normalized in `src/lib/mapBackend.ts`.
Mock timers in `src/lib/mock/scenarios.ts` remain available when `VITE_USE_API=false`.

List panels (Agent Activity, Current Decision, Trust Pipeline, Decision History) use fixed
heights with internal scroll — tune `--panel-workspace-h` and `--panel-history-h` in
`frontend/src/styles.css`. See [`frontend/README.md`](frontend/README.md) for layout details.

---

## Roadmap

| Milestone | Deliverable | Status |
|-----------|-------------|--------|
| **0** | Algorand + x402 integration spikes | ✅ probe; ⚠️ on-chain needs valid wallet |
| **1** | HTTP API + dashboard hydration | ✅ |
| **Infra** | `pkg/provenance/` RAv1 + SPEC | ✅ |
| **2** | Catalog, services, hero orchestrator, SSE API | ✅ |
| **3** | Real EURQ `PayAndFetch` | ✅ |
| **4** | Frontend scenario SSE + live demo UI | ✅ |

---

## Environment variables

| Variable | Required | Description |
|----------|----------|-------------|
| `RATIONALGO_WALLET_ADDRESS` | Yes (spikes) | 58-character Pera Testnet address |
| `RATIONALGO_MNEMONIC` | Yes (spikes) | 24-word Pera Universal Wallet seed or legacy 25-word Algorand passphrase |
| `RATIONALGO_ALGOD_TOKEN` | No | Empty for public AlgoNode testnet |
| `RATIONALGO_ALGOD_URL` | No | Default: `https://testnet-api.algonode.cloud` — switch if free quota exceeded |
| `RATIONALGO_X402_PROBE_URL` | No | Default: `…/avm/weather` (used by `spike x402`, not the hero demo) |
| `RATIONALGO_SETTLEMENT_ASSET_ID` | No | ASA the `/company/*` x402 seller charges in. Default: testnet USDC `10458941` |
| `RATIONALGO_HTTP_ADDR` | No | Default: `:8080` — also used to build the seller's self-referential `EndpointURL`s (`cfg.PublicBaseURL()`) |
| `RATIONALGO_ANTHROPIC_KEY` | No | Anthropic API key for `POST /api/decide` (hero demo works without it) |
| `VITE_API_URL` | No | Frontend API base (default: `http://localhost:8080`) |
| `VITE_USE_API` | No | Default: API enabled. Set to `false` for client-side mock scenario only |

Never commit `backend/.env`.

---

## License

Hackathon submission — MIT (TBD).
