import type { Dispatch } from "react";
import type { Action } from "./reducer";
import type { Decision } from "./types";
import { mockHash, mockRound, uid } from "./mock";

export function runDemo(dispatch: Dispatch<Action>, timers: number[]) {
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
