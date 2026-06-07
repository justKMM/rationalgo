
## RationAlgo — Mission Control Frontend

A dark, fintech/cybersecurity mission-control UI for an Algorand-native governance layer over agentic x402 payments. Desktop-first, mock data only, no backend.

### Stack note
The project is on **TanStack Start + Tailwind v4 + shadcn/ui** (not vanilla Vite React Router). I'll honor that: file-based routes under `src/routes/`, tokens in `src/styles.css` under `@theme`, no `tailwind.config.js`. Functionally identical to the brief.

### Design direction
- Dark by default. Near-black base (`oklch(0.16 0.02 260)`), elevated panels, hairline borders.
- Accent: electric cyan/teal for "live/verified", amber for "pending", red for "blocked".
- Mono display font (JetBrains Mono) for IDs/tx hashes/timestamps; Inter for body.
- Subtle scanline / grid background on the shell, pulsing status dots, animated pipeline progression.
- Motion via framer-motion: stage activation, reasoning feed slide-in, decision card swap.

### Routes
```
src/routes/
  __root.tsx              # shell: top bar, grid background, providers
  index.tsx               # Mission Control (main screen)
```
Decision details = **Sheet/Drawer** (shadcn `sheet`) opened from the history table, not a route. Trust Metrics = strip of 4 KPI cards on the same page above the history table (per brief: "Trust Metrics Panel" with 4 KPIs only).

### Layout (single screen)
```
┌─────────────────────────────────────────────────────────────┐
│ TOP BAR  RationAlgo • ●ONLINE   [Normal] [Anomaly] [Reset]  │
├──────────────┬──────────────────────────┬───────────────────┤
│ Reasoning    │  Active Decision Card    │  Trust Pipeline   │
│ Feed         │  task / vendor / cost    │  6 vertical       │
│ (chat-like   │  confidence meter        │  stages, activate │
│  timeline)   │  policy + decision state │  in sequence      │
│              │                          │                   │
├──────────────┴──────────────────────────┴───────────────────┤
│  [Trust Score] [Successful] [Blocked] [Violations Prev.]    │
├─────────────────────────────────────────────────────────────┤
│ Decision History Table (click row → Drawer)                 │
└─────────────────────────────────────────────────────────────┘
```

### Component architecture
```
src/
  components/
    mission-control/
      TopBar.tsx              # name, status dot, scenario buttons
      ReasoningFeed.tsx       # left panel, animated event list
      ActiveDecisionCard.tsx  # center panel
      TrustPipeline.tsx       # right panel, 6 stages
      TrustMetrics.tsx        # 4 KPI cards
      DecisionHistoryTable.tsx
      DecisionDetailsDrawer.tsx
      StageNode.tsx           # one pipeline stage
      ConfidenceMeter.tsx
      StatusPill.tsx          # approved/blocked/verified/pending/failed
      TxHash.tsx              # truncated mono hash w/ copy
    ui/                       # existing shadcn
  lib/
    mock/
      scenarios.ts            # normal + anomaly scripted event sequences
      decisions.ts            # seed DecisionRecord history
      generators.ts           # fake tx hashes, vendors, timestamps
    types.ts                  # DecisionRecord, ScenarioEvent, PipelineStage
  hooks/
    useScenarioRunner.ts      # drives event playback + pipeline state
    useMissionStore.ts        # zustand store (events, active decision, history, metrics)
```

### State management
Lightweight **zustand** store (`useMissionStore`):
- `events: ScenarioEvent[]`
- `activeDecision: DecisionRecord | null`
- `pipelineStage: 0..6` (0 = idle)
- `history: DecisionRecord[]`
- `metrics: { trustScore, successful, blocked, violationsPrevented }`
- actions: `runScenario('normal'|'anomaly')`, `reset()`, `selectHistory(id)`

`useScenarioRunner` orchestrates a scripted timeline with `setTimeout` ticks (~600–900ms per stage) to:
1. push reasoning events into the feed,
2. advance `pipelineStage`,
3. mutate the active decision (policy result, tx hashes, outcome),
4. on completion, prepend to history + update metrics.

### Scenarios
- **Normal**: agent reasons → policy approved → pre-tx committed → x402 paid → outcome verified → post-tx committed. Confidence ~0.92.
- **Anomaly**: agent reasons → policy **blocked** (budget/vendor-reputation violation) → `alert.fired`, no payment, violations counter +1. Or alt: approved → payment sent → outcome **failed** → post-tx records failure.
- Reset clears active decision + pipeline; keeps seeded history.

### Mock data
- 12 seeded historical decisions across vendors (OpenAI API, AWS S3, Stripe top-up, Twilio SMS, Pinecone, Replicate, etc.), realistic USD costs, mix of approved/blocked/verified/failed.
- Algorand-style tx hashes: 52-char base32 uppercase via generator.
- Timestamps within last 24h.

### Design tokens (added to `src/styles.css`)
```
--background: oklch(0.16 0.02 260)
--panel:      oklch(0.20 0.02 260)
--border:     oklch(0.28 0.02 260)
--accent:     oklch(0.78 0.16 195)   /* cyan */
--success:    oklch(0.72 0.18 155)
--warn:       oklch(0.80 0.16 75)
--danger:     oklch(0.65 0.22 25)
--mono font: JetBrains Mono via <link> in __root head
```

### Drawer contents (DecisionDetailsDrawer)
- Decision ID (mono), timestamp
- Reasoning summary (multi-line)
- Policy results (list with pass/fail chips: budget, vendor allow-list, rate limit, risk)
- Predicted vs actual outcome (side-by-side)
- Pre-payment TX (hash + "View on Algorand" stub link)
- Post-outcome TX (same)

### What I will NOT build
- No auth, no real Algorand/x402 calls, no backend.
- No charts / analytics dashboard look.
- No extra routes beyond the single mission-control screen + drawer.

### Verification
- Typecheck/build runs automatically.
- Visually verify both scenarios end-to-end in the preview, drawer opens from history row, reset returns to idle state.

Ready to switch to build mode and implement.
