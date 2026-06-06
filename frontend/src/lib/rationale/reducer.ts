import type { AppState, Decision, Alert, Outcome } from "./types";
import { initialState } from "./mock";

export type Action =
  | { type: "ADD_DECISION"; decision: Decision; select?: boolean }
  | { type: "UPDATE_DECISION"; id: string; patch: Partial<Decision> }
  | { type: "SET_OUTCOME"; id: string; outcome: Outcome }
  | { type: "ADJUST_TRUST"; vendor: string; delta: number }
  | { type: "ADD_ALERT"; alert: Alert }
  | { type: "SELECT"; id: string | null }
  | { type: "SPEND"; amount: number }
  | { type: "RESET" };

export function reducer(state: AppState, action: Action): AppState {
  switch (action.type) {
    case "ADD_DECISION":
      return {
        ...state,
        decisions: [action.decision, ...state.decisions],
        selectedId: action.select ? action.decision.id : state.selectedId,
      };
    case "UPDATE_DECISION":
      return {
        ...state,
        decisions: state.decisions.map((d) =>
          d.id === action.id ? { ...d, ...action.patch } : d
        ),
      };
    case "SET_OUTCOME":
      return {
        ...state,
        decisions: state.decisions.map((d) =>
          d.id === action.id ? { ...d, outcome: action.outcome } : d
        ),
      };
    case "ADJUST_TRUST":
      return {
        ...state,
        vendors: state.vendors.map((v) =>
          v.name === action.vendor
            ? {
                ...v,
                score: Math.max(0, Math.min(5, +(v.score + action.delta).toFixed(2))),
                lastDelta: {
                  dir: action.delta >= 0 ? "up" : "down",
                  value: Math.abs(action.delta),
                  at: Date.now(),
                },
              }
            : v
        ),
      };
    case "ADD_ALERT":
      return { ...state, alerts: [action.alert, ...state.alerts] };
    case "SELECT":
      return { ...state, selectedId: action.id };
    case "SPEND":
      return {
        ...state,
        balance: +(state.balance - action.amount).toFixed(2),
        spent: +(state.spent + action.amount).toFixed(2),
      };
    case "RESET":
      return { ...initialState, selectedId: null };
  }
}
