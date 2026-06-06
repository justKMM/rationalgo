# Rationale — Agent Decision Audit Dashboard

A single-page, frontend-only demo. All state in React, all hashes/amounts mocked. Dark, dense, technical "control room" aesthetic — explicitly not a SaaS landing page.

## Scope

- One route: replace `src/routes/index.tsx` with the dashboard.
- No backend, no Cloud, no blockchain calls. Pure React state + `setTimeout` for the scripted demo.
- Update `__root.tsx` head meta (title "Rationale — Agent Decision Audit", description) and load JetBrains Mono via `<link>` in the root head; register `--font-mono` and `--font-sans` in `src/styles.css` `@theme`.

## Design tokens (src/styles.css)

Override `:root` and `.dark` with the control-room palette; default the app to dark (add `class="dark"` on `<html>` in `__root.tsx`).

- `--background: oklch(0.16 0.005 250)` (~#0B0D0F)
- `--foreground: oklch(0.95 0 0)`
- `--card: oklch(0.19 0.005 250)`
- `--muted-foreground: oklch(0.62 0.01 250)`
- `--border: oklch(0.28 0.005 250)` (thin 1px dividers)
- `--primary` / accent-approved: `oklch(0.82 0.21 150)` (~#3DDC84 electric green)
- `--destructive` / blocked: `oklch(0.62 0.22 25)` (red)
- `--accent-alert: oklch(0.78 0.16 75)` (amber)
- Tabular numerals utility via `@utility tnum { font-variant-numeric: tabular-nums; }`
- Subtle `@keyframes fade-in-up` already available via existing animations.

## Layout

```text
┌─────────────────────────────────────────────────────────────┐
│ TOP BAR: fleet-router-01 │ EURQ 9.41 │ spent 0.59/10 │ [Run]│
├──────────────────────────────────────────┬──────────────────┤
│                                          │  POLICY          │
│   DECISION FEED (newest on top)          │  - daily limit   │
│   ┌──────────────────────────────────┐   │  - allowed       │
│   │ vendor ● APPROVED   0.12 EURQ    │   │  - blocked       │
│   │ intent line…                     │   │  - alerts        │
│   │ 🔒 0xa31f…b2 committed  12:04:31 │   ├──────────────────┤
│   └──────────────────────────────────┘   │  VENDOR TRUST    │
│   …                                      │  WeatherAPI 4.2  │
│                                          │  …               │
└──────────────────────────────────────────┴──────────────────┘
       (drawer slides in from right on card click)
```

- CSS grid: `grid-cols-[1fr_360px]`, `gap-px bg-border` for hairline dividers between zones.
- Drawer: fixed right overlay, ~520px, scrollable, overlays the right rail. Close on backdrop click / Esc.

## Component breakdown (all under `src/components/rationale/`)

- `TopBar.tsx` — agent name, EURQ balance (mono), `SpendGauge` (thin horizontal bar), `RunDemoButton`, `ResetButton`.
- `DecisionFeed.tsx` — maps decisions, renders `DecisionCard`.
- `DecisionCard.tsx` — vendor, `StatusPill` (APPROVED/BLOCKED/PENDING), amount mono, intent line, commit row (truncated hash + lock icon + "reasoning committed pre-outcome"), timestamp. `animate-fade-in` on mount.
- `DecisionDrawer.tsx` — full record: intent, alternatives list (each with rejection reason), expected value, confidence, cost, policy checks (budget/reputation/anomaly), on-chain reasoning hash + round number, `OutcomeBlock` (renders once `decision.outcome` is set).
- `PolicyPanel.tsx` — daily limit row, allowed vendors, blocked vendors, alerts list (amber rows).
- `VendorTrustPanel.tsx` — list of 4-5 vendors, score (mono), tiny predicted-vs-actual sparkline-style indicator (▲/▼ + delta) that animates on update.
- `lib/mock.ts` — seed data, hash generator (`0x` + 6 hex … 4 hex), types.
- `lib/demoScenario.ts` — function `runDemo(dispatch)` orchestrating the 5 steps with timeouts.

## State shape

```ts
type Decision = {
  id: string;
  vendor: string;
  status: 'APPROVED' | 'BLOCKED' | 'PENDING';
  amountEURQ: number;
  intent: string;
  alternatives: { name: string; reason: string }[];
  expectedValue: string;       // "+23% routing confidence"
  confidence: number;          // 0.81
  policy: { budgetOk: boolean; reputation: number; anomaly: 'none'|'flagged' };
  reasoningHash: string;       // mocked
  round: number;               // mocked Algorand round
  timestamp: number;
  outcome?: { predicted: string; actual: string; verdict: string; trustDelta: number };
};
```

Single `useReducer` in `index.tsx` managing `{ decisions, balance, spent, vendors, alerts, selectedId }`. Actions: `ADD_DECISION`, `UPDATE_DECISION`, `SET_OUTCOME`, `ADJUST_TRUST`, `ADD_ALERT`, `RESET`.

## Demo scenario (≈10s)

Triggered by `Run demo scenario`:

- t=0: dispatch `ADD_DECISION` (PENDING) for WeatherAPI, 0.12 EURQ, intent "Task needs 24h weather forecast for route optimization". Card slides in. Auto-open drawer.
- t=0.4s: populate alternatives + expected value in drawer (already in payload; reveal via small staged animation).
- t=1.2s: flip to APPROVED, set `reasoningHash`, decrement balance by 0.12, advance spend gauge.
- t=5s: `SET_OUTCOME` predicted +23% / actual +25%, verdict "Good purchase"; `ADJUST_TRUST` WeatherAPI +0.1 (tick ▲ in right rail).
- t=7s: `ADD_DECISION` from unknown vendor "MetricsHub.xyz" at 1.20 EURQ (10× normal), immediately BLOCKED red card; `ADD_ALERT` amber: "MetricsHub.xyz price +1000% vs 7d avg — flagged".

`Reset demo` clears scenario-added items and restores seed state + balance/spent.

## Seed data (4 prior decisions)

Pre-populate feed with realistic vendor/intent pairs, e.g.:
- OSRM-Pro — 0.08 EURQ — "Recompute route after traffic spike on A4" — APPROVED
- TollGuru — 0.04 EURQ — "Toll cost lookup BE→NL corridor" — APPROVED
- FuelPriceAPI — 0.15 EURQ — "Fuel index for fleet rebalancing" — APPROVED
- ScrapeShack — 0.90 EURQ — "Competitor pricing scrape" — BLOCKED (vendor not allowlisted)

Vendors in trust panel: WeatherAPI 4.2, OSRM-Pro 4.7, TollGuru 4.5, FuelPriceAPI 4.0, MetricsHub.xyz 1.1.

## Out of scope

- Real Algorand SDK, x402 wiring, wallets, persistence, auth.
- Mobile layout polish beyond "doesn't break" (desktop-first control room).
- Tests.

## Files touched

- `src/routes/index.tsx` (rewrite)
- `src/routes/__root.tsx` (head meta + JetBrains Mono link + `<html class="dark">`)
- `src/styles.css` (palette overrides, mono token, tnum utility)
- New: `src/components/rationale/{TopBar,DecisionFeed,DecisionCard,DecisionDrawer,PolicyPanel,VendorTrustPanel,StatusPill,SpendGauge}.tsx`
- New: `src/lib/rationale/{types.ts,mock.ts,demoScenario.ts,reducer.ts}`
