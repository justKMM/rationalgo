import { create } from "zustand";
import type { DecisionRecord, ScenarioEvent } from "@/lib/types";
import { seedHistory } from "@/lib/mock/decisions";
import { buildScenario, type ScenarioKind } from "@/lib/mock/scenarios";

interface Metrics {
  trustScore: number;
  successful: number;
  blocked: number;
  violationsPrevented: number;
}

interface MissionState {
  events: ScenarioEvent[];
  activeDecision: DecisionRecord | null;
  pipelineStage: number; // 0..6
  history: DecisionRecord[];
  metrics: Metrics;
  selectedDecisionId: string | null;
  running: boolean;
  _timers: ReturnType<typeof setTimeout>[];

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

const initialHistory = seedHistory();

export const useMissionStore = create<MissionState>((set, get) => ({
  events: [],
  activeDecision: null,
  pipelineStage: 0,
  history: initialHistory,
  metrics: computeMetrics(initialHistory),
  selectedDecisionId: null,
  running: false,
  _timers: [],

  runScenario: (kind) => {
    const state = get();
    state._timers.forEach(clearTimeout);
    const script = buildScenario(kind);
    set({
      events: [],
      activeDecision: script.decision,
      pipelineStage: 0,
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
        set({
          pipelineStage: step.stage,
          events: evt ? [...s.events, evt] : s.events,
          activeDecision: nextDecision,
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
  },

  reset: () => {
    get()._timers.forEach(clearTimeout);
    set({
      events: [],
      activeDecision: null,
      pipelineStage: 0,
      running: false,
      _timers: [],
    });
  },

  selectDecision: (id) => set({ selectedDecisionId: id }),
}));
