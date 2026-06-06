import type { Dispatch } from "react";
import type { Action } from "./reducer";
import type { Decision } from "./types";
import { mockHash, mockRound, uid } from "./mock";

// ScenarioEvent mirrors the backend event types from scenario/events.go
interface ScenarioEvent {
  type: "decision.pending" | "decision.committed" | "decision.outcome" | "decision.blocked";
  decision?: Decision;
  id?: string;
  tx_id?: string;
  outcome?: {
    predicted: string;
    actual: string;
    verdict: string;
    trustDelta: number;
  };
  alert?: {
    id: string;
    level: "amber" | "red";
    message: string;
    at: number;
  };
}

function mapEventToAction(event: ScenarioEvent): Action | null {
  switch (event.type) {
    case "decision.pending":
      if (!event.decision) return null;
      return { type: "ADD_DECISION", decision: event.decision, select: true };
    case "decision.committed":
      if (!event.id || !event.tx_id) return null;
      return { type: "UPDATE_DECISION", id: event.id, patch: { committedTx: event.tx_id, status: "APPROVED" } };
    case "decision.outcome":
      if (!event.id || !event.outcome) return null;
      return { type: "SET_OUTCOME", id: event.id, outcome: event.outcome };
    case "decision.blocked":
      if (!event.decision) return null;
      return { type: "ADD_DECISION", decision: { ...event.decision, status: "BLOCKED" } };
    default:
      return null;
  }
}

// runLiveScenario subscribes to the backend SSE stream and dispatches events.
// Returns a cleanup function that closes the connection.
export function runLiveScenario(dispatch: Dispatch<Action>): () => void {
  const evtSource = new EventSource(
    `${import.meta.env.VITE_API_URL ?? "http://localhost:8080"}/api/scenario/run`
  );

  evtSource.onmessage = (e: MessageEvent) => {
    try {
      const event = JSON.parse(e.data as string) as ScenarioEvent;
      // Handle blocked decisions: also fire an alert if provided
      if (event.type === "decision.blocked" && event.alert) {
        dispatch({ type: "ADD_ALERT", alert: event.alert });
      }
      const action = mapEventToAction(event);
      if (action) dispatch(action);
    } catch {
      // ignore malformed SSE frames
    }
  };

  evtSource.onerror = () => evtSource.close();

  return () => evtSource.close();
}


  const weatherId = uid();
  const now = Date.now();

  const pending: Decision = {
    id: weatherId,
    vendor: "WeatherAPI",
    status: "PENDING",
    amountEURQ: 0.12,
    intent: "Task needs 24h weather forecast for route optimization (NL/BE corridor)",
    alternatives: [
      { name: "OpenMeteoFree", reason: "95% recent outage rate over last 24h" },
      { name: "cached result", reason: "12h stale — front passing over Rotterdam at 18:00" },
    ],
    expectedValue: "+23% routing confidence",
    confidence: 0.81,
    policy: { budgetOk: true, reputation: 4.2, anomaly: "none", vendorAllowed: true },
    reasoningHash: mockHash(),
    round: mockRound(),
    timestamp: now,
  };

  // t=0: card slides in (PENDING) + drawer opens
  dispatch({ type: "ADD_DECISION", decision: pending, select: true });

  // t=1.2s: flip to APPROVED, charge balance
  timers.push(
    window.setTimeout(() => {
      dispatch({
        type: "UPDATE_DECISION",
        id: weatherId,
        patch: { status: "APPROVED", reasoningHash: mockHash() },
      });
      dispatch({ type: "SPEND", amount: 0.12 });
    }, 1200)
  );

  // t=5s: outcome lands
  timers.push(
    window.setTimeout(() => {
      dispatch({
        type: "SET_OUTCOME",
        id: weatherId,
        outcome: {
          predicted: "+23%",
          actual: "+25%",
          verdict: "Good purchase",
          trustDelta: 0.1,
        },
      });
      dispatch({ type: "ADJUST_TRUST", vendor: "WeatherAPI", delta: 0.1 });
    }, 5000)
  );

  // t=7s: unknown vendor blocked
  timers.push(
    window.setTimeout(() => {
      const blockedId = uid();
      const blocked: Decision = {
        id: blockedId,
        vendor: "MetricsHub.xyz",
        status: "BLOCKED",
        amountEURQ: 1.2,
        intent: "Real-time courier density heatmap for surge dispatch",
        alternatives: [
          { name: "FleetSignal", reason: "allowlisted, same dataset @ 0.11 EURQ" },
        ],
        expectedValue: "—",
        confidence: 0.29,
        policy: { budgetOk: true, reputation: 1.1, anomaly: "flagged", vendorAllowed: false },
        reasoningHash: mockHash(),
        round: mockRound(),
        timestamp: Date.now(),
        blockedReason: "Price anomaly +1000% vs 7d median; vendor not on allowlist",
      };
      dispatch({ type: "ADD_DECISION", decision: blocked });
      dispatch({
        type: "ADD_ALERT",
        alert: {
          id: uid(),
          level: "amber",
          message: "MetricsHub.xyz price +1000% vs 7d avg — flagged",
          at: Date.now(),
        },
      });
    }, 7000)
  );
}
