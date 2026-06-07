import { create } from "zustand";
import type {
  DecisionRecord,
  ScenarioEvent,
  ScenarioEventType,
  VendorPlanStep,
  VendorStepStatus,
} from "@/lib/types";
import {
  checkHealth,
  fetchDashboardState,
  isApiConfigured,
  resetDashboardState,
  runScenarioStream,
  type BackendDecision,
  type BackendDecisionRecord,
  type BackendScenarioEvent,
} from "@/lib/api";
import { mapBackendDecision, mapBackendRecord, mergeDecisionRecord } from "@/lib/mapBackend";
import { budgetEur, type BudgetTier } from "@/lib/budget";
import { buildScenario, type ScenarioKind } from "@/lib/mock/scenarios";
import { seedHistory } from "@/lib/mock/decisions";

interface Metrics {
  trustScore: number;
  successful: number;
  blocked: number;
  violationsPrevented: number;
}

interface MissionState {
  events: ScenarioEvent[];
  activeDecision: DecisionRecord | null;
  pipelineStage: number;
  history: DecisionRecord[];
  metrics: Metrics;
  selectedDecisionId: string | null;
  running: boolean;
  apiLive: boolean;
  budgetTier: BudgetTier;
  vendorPlan: VendorPlanStep[];
  error: string | null;
  _timers: ReturnType<typeof setTimeout>[];
  _abort: AbortController | null;

  hydrate: () => Promise<void>;
  setBudgetTier: (tier: BudgetTier) => void;
  runScenario: (kind: ScenarioKind) => void;
  reset: () => void;
  selectDecision: (id: string | null) => void;
}

function computeMetrics(history: DecisionRecord[]): Metrics {
  const successful = history.filter((d) => d.outcomeStatus === "verified").length;
  const blocked = history.filter((d) => d.policyStatus === "blocked").length;
  const failed = history.filter((d) => d.outcomeStatus === "failed").length;
  const total = history.length || 1;
  const trustScore = Math.round(((successful + blocked * 0.6) / total) * 100);
  return {
    trustScore,
    successful,
    blocked,
    violationsPrevented: blocked + failed,
  };
}

function historyFromState(decisions: BackendDecision[]): DecisionRecord[] {
  return [...decisions].reverse().map(mapBackendDecision);
}

const mockHistory = seedHistory();
const useApi = isApiConfigured();

export const useMissionStore = create<MissionState>((set, get) => ({
  events: [],
  activeDecision: null,
  pipelineStage: 0,
  history: useApi ? [] : mockHistory,
  metrics: computeMetrics(useApi ? [] : mockHistory),
  selectedDecisionId: null,
  running: false,
  apiLive: false,
  budgetTier: "mid",
  vendorPlan: [],
  error: null,
  _timers: [],
  _abort: null,

  setBudgetTier: (tier) => set({ budgetTier: tier }),

  hydrate: async () => {
    if (!isApiConfigured()) return;
    const live = await checkHealth();
    set({ apiLive: live });
    if (!live) return;
    try {
      const state = await fetchDashboardState();
      const history = historyFromState(state.decisions);
      set({ history, metrics: computeMetrics(history), error: null });
    } catch (err) {
      set({ apiLive: false, error: err instanceof Error ? err.message : "Failed to load state" });
    }
  },

  runScenario: (kind) => {
    if (isApiConfigured()) {
      void runLiveScenario(kind, set, get);
      return;
    }
    runMockScenario(kind, set, get);
  },

  reset: () => {
    const state = get();
    state._abort?.abort();
    state._timers.forEach(clearTimeout);
    set({
      events: [],
      activeDecision: null,
      pipelineStage: 0,
      vendorPlan: [],
      running: false,
      error: null,
      _timers: [],
      _abort: null,
    });
    if (isApiConfigured()) {
      void resetDashboardState()
        .then((state) => {
          const history = historyFromState(state.decisions);
          set({ history, metrics: computeMetrics(history) });
        })
        .catch((err) => {
          set({ error: err instanceof Error ? err.message : "Reset failed" });
        });
    }
  },

  selectDecision: (id) => set({ selectedDecisionId: id }),
}));

function pushEvent(
  set: (partial: Partial<MissionState> | ((s: MissionState) => Partial<MissionState>)) => void,
  get: () => MissionState,
  type: ScenarioEventType,
  message: string,
): void {
  const idx = get().events.length;
  const evt: ScenarioEvent = {
    id: `evt-${Date.now()}-${idx}`,
    timestamp: new Date().toISOString(),
    type,
    message,
  };
  set((s) => ({ events: [...s.events, evt] }));
}

function setVendorStatus(
  plan: VendorPlanStep[],
  vendorName: string,
  status: VendorStepStatus,
): VendorPlanStep[] {
  return plan.map((step) =>
    step.name === vendorName ? { ...step, status } : step,
  );
}

function finalizeActive(
  set: (partial: Partial<MissionState> | ((s: MissionState) => Partial<MissionState>)) => void,
  get: () => MissionState,
): void {
  const s = get();
  if (!s.activeDecision) return;
  const newHistory = [s.activeDecision, ...s.history];
  set({
    history: newHistory,
    metrics: computeMetrics(newHistory),
    activeDecision: null,
    pipelineStage: 0,
  });
}

function handleBackendEvent(
  event: BackendScenarioEvent,
  set: (partial: Partial<MissionState> | ((s: MissionState) => Partial<MissionState>)) => void,
  get: () => MissionState,
  budget?: number,
): void {
  switch (event.type) {
    case "research.plan": {
      const payload = event.payload as {
        vendors?: Array<{
          id: string;
          name: string;
          price_eurq: number;
          order: number;
        }>;
      };
      const vendorPlan: VendorPlanStep[] = (payload.vendors ?? []).map((v) => ({
        id: v.id,
        name: v.name,
        price: v.price_eurq,
        order: v.order,
        status: "pending",
      }));
      set({ vendorPlan, pipelineStage: 1 });
      pushEvent(
        set,
        get,
        "agent.thinking",
        `Plan ready · ${vendorPlan.length} vendor${vendorPlan.length === 1 ? "" : "s"} in queue`,
      );
      break;
    }
    case "agent.thinking": {
      const payload = event.payload as { intent?: string; company?: string; budget_eurq?: number };
      set({ pipelineStage: 1, vendorPlan: [] });
      pushEvent(
        set,
        get,
        "agent.thinking",
        payload.intent ?? `Researching ${payload.company ?? "target company"} within budget.`,
      );
      break;
    }
    case "decision.pending": {
      const record = mapBackendRecord(event.payload, budget);
      set((s) => ({
        activeDecision: record,
        pipelineStage: 1,
        vendorPlan: s.vendorPlan.map((step) => ({
          ...step,
          status:
            step.name === record.vendor
              ? "active"
              : step.status === "active"
              ? "pending"
              : step.status,
        })),
      }));
      pushEvent(
        set,
        get,
        "decision.pending",
        `Selected ${record.vendor} · $${record.cost.toFixed(2)} · confidence ${record.confidence.toFixed(2)}`,
      );
      break;
    }
    case "decision.committed": {
      const payload = event.payload as
        | BackendDecisionRecord
        | { record?: BackendDecisionRecord; committed_tx?: string; commit_error?: string };
      const raw =
        payload && typeof payload === "object" && "record" in payload && payload.record
          ? payload.record
          : (payload as BackendDecisionRecord);
      const record = mapBackendRecord(
        {
          ...raw,
          committed_tx:
            ("committed_tx" in payload && payload.committed_tx) || raw.committed_tx,
        },
        budget,
      );
      const passed = record.policyChecks?.filter((c) => c.passed).length ?? 0;
      const total = record.policyChecks?.length ?? 0;
      const commitError =
        payload && typeof payload === "object" && "commit_error" in payload
          ? payload.commit_error
          : undefined;
      set((s) => ({
        pipelineStage: 3,
        activeDecision: s.activeDecision
          ? mergeDecisionRecord(s.activeDecision, record)
          : record,
      }));
      pushEvent(set, get, "policy.approved", `Policy: ${passed}/${total} checks passed.`);
      if (record.txPre) {
        pushEvent(
          set,
          get,
          "decision.committed",
          `Reasoning anchored on Algorand · ${record.txPre.slice(0, 10)}…`,
        );
      } else if (commitError) {
        pushEvent(
          set,
          get,
          "decision.committed",
          `Algorand commit failed · ${commitError}`,
        );
      } else {
        pushEvent(set, get, "decision.committed", "Reasoning committed (no on-chain tx).");
      }
      break;
    }
    case "payment.sent": {
      const payload = event.payload as {
        vendor?: string;
        amount?: number;
        paid?: boolean;
        settlement_tx?: string;
        error?: string;
      };
      set((s) => ({
        pipelineStage: 4,
        activeDecision: s.activeDecision && payload.settlement_tx
          ? { ...s.activeDecision, settlementTx: payload.settlement_tx }
          : s.activeDecision,
      }));
      if (payload.paid) {
        pushEvent(
          set,
          get,
          "payment.sent",
          payload.settlement_tx
            ? `x402 payment settled · ${payload.settlement_tx.slice(0, 10)}…`
            : `x402 payment sent to ${payload.vendor ?? "vendor"}.`,
        );
      } else {
        pushEvent(
          set,
          get,
          "payment.sent",
          payload.error ?? "x402 payment failed.",
        );
      }
      break;
    }
    case "decision.outcome": {
      const record = mapBackendRecord(event.payload, budget);
      set((s) => ({
        pipelineStage: record.txOutcome ? 6 : 5,
        vendorPlan: setVendorStatus(s.vendorPlan, record.vendor, "completed"),
        activeDecision: s.activeDecision
          ? mergeDecisionRecord(s.activeDecision, {
              ...record,
              outcomeStatus: record.outcomeStatus,
              actualOutcome: record.actualOutcome ?? record.predictedOutcome,
            })
          : record,
      }));
      pushEvent(
        set,
        get,
        "decision.outcome",
        record.actualOutcome
          ? `Outcome verified · ${record.actualOutcome}`
          : "Outcome recorded.",
      );
      if (record.txOutcome) {
        pushEvent(
          set,
          get,
          "decision.committed",
          `Outcome receipt anchored · ${record.txOutcome.slice(0, 10)}…`,
        );
      }
      finalizeActive(set, get);
      break;
    }
    case "decision.blocked": {
      const record = mapBackendRecord(event.payload, budget);
      set((s) => ({
        pipelineStage: 2,
        vendorPlan: setVendorStatus(s.vendorPlan, record.vendor, "blocked"),
        activeDecision: s.activeDecision
          ? mergeDecisionRecord(s.activeDecision, { ...record, policyStatus: "blocked" })
          : { ...record, policyStatus: "blocked" },
      }));
      pushEvent(
        set,
        get,
        "decision.blocked",
        record.policyChecks?.find((c) => !c.passed)?.detail ?? "Purchase blocked by policy.",
      );
      break;
    }
    case "alert.fired": {
      const alert = event.payload as { message?: string };
      pushEvent(set, get, "alert.fired", alert.message ?? "Policy alert fired.");
      finalizeActive(set, get);
      break;
    }
    case "research.summary": {
      const summary = event.payload as {
        company?: string;
        purchased?: string[];
        blocked?: string[];
        spent_eurq?: number;
        budget_eurq?: number;
      };
      const purchased = summary.purchased?.length ?? 0;
      const blocked = summary.blocked?.length ?? 0;
      set({ pipelineStage: 6, running: false });
      pushEvent(
        set,
        get,
        "research.summary",
        `Research complete for ${summary.company ?? "company"} · ${purchased} purchased · ${blocked} blocked · $${(summary.spent_eurq ?? 0).toFixed(2)} spent of $${(summary.budget_eurq ?? 0).toFixed(2)}.`,
      );
      break;
    }
    default:
      break;
  }
}

async function runLiveScenario(
  kind: ScenarioKind,
  set: (partial: Partial<MissionState> | ((s: MissionState) => Partial<MissionState>)) => void,
  get: () => MissionState,
): Promise<void> {
  const state = get();
  state._abort?.abort();
  state._timers.forEach(clearTimeout);

  const abort = new AbortController();
  set({
    events: [],
    activeDecision: null,
    pipelineStage: 0,
    vendorPlan: [],
    running: true,
    error: null,
    _timers: [],
    _abort: abort,
  });

  const budget = budgetEur(get().budgetTier);

  try {
    await runScenarioStream(
      kind,
      get().budgetTier,
      (event) => handleBackendEvent(event, set, get, budget),
      abort.signal,
    );
    set({ running: false, _abort: null });
    void get().hydrate();
  } catch (err) {
    if (abort.signal.aborted) return;
    set({
      running: false,
      _abort: null,
      error: err instanceof Error ? err.message : "Scenario failed",
    });
  }
}

function runMockScenario(
  kind: ScenarioKind,
  set: (partial: Partial<MissionState> | ((s: MissionState) => Partial<MissionState>)) => void,
  get: () => MissionState,
): void {
  const state = get();
  state._timers.forEach(clearTimeout);
  const script = buildScenario(kind);
  const mockPlan: VendorPlanStep[] = [
    {
      id: "mock-vendor",
      name: script.decision.vendor,
      price: script.decision.cost,
      order: 1,
      status: "pending",
    },
  ];
  set({
    events: [],
    activeDecision: script.decision,
    pipelineStage: 0,
    vendorPlan: mockPlan,
    running: true,
    _timers: [],
  });

  const timers: ReturnType<typeof setTimeout>[] = [];
  script.steps.forEach((step, idx) => {
    const t = setTimeout(() => {
      const s = get();
      const evt: ScenarioEvent | null = step.event
        ? {
            id: `${script.decision.id}-${idx}`,
            timestamp: new Date().toISOString(),
            ...step.event,
          }
        : null;
      const nextDecision = step.patch && s.activeDecision
        ? { ...s.activeDecision, ...step.patch }
        : s.activeDecision;
      const vendorPlan = [...s.vendorPlan];
      if (step.event?.type === "decision.pending" && vendorPlan.length > 0) {
        vendorPlan[0] = { ...vendorPlan[0], status: "active" };
      }
      if (step.finalize && vendorPlan.length > 0) {
        vendorPlan[0] = {
          ...vendorPlan[0],
          status: nextDecision?.policyStatus === "blocked" ? "blocked" : "completed",
        };
      }

      set({
        pipelineStage: step.stage,
        events: evt ? [...s.events, evt] : s.events,
        activeDecision: nextDecision,
        vendorPlan,
      });
      if (step.finalize && nextDecision) {
        const newHistory = [nextDecision, ...s.history];
        set({
          history: newHistory,
          metrics: computeMetrics(newHistory),
          running: false,
        });
      }
    }, step.delay);
    timers.push(t);
  });
  set({ _timers: timers });
}
