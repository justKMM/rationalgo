import type { BudgetTier } from "@/lib/budget";

const defaultBase = "http://localhost:8080";

export function apiBase(): string {
  return import.meta.env.VITE_API_URL ?? defaultBase;
}

export function isApiConfigured(): boolean {
  return import.meta.env.VITE_USE_API !== "false";
}

export interface BackendPolicyResult {
  approved: boolean;
  budget_ok: boolean;
  vendor_allowed: boolean;
  price_anomaly: boolean;
  block_reason?: string;
}

export interface BackendVendorOption {
  id: string;
  name: string;
  price_eurq: number;
  description?: string;
}

export interface BackendOutcomeRecord {
  predicted: string;
  actual: string;
  verdict: string;
  trust_delta?: number;
}

export interface BackendDecisionRecord {
  id: string;
  task_intent: string;
  vendor_chosen: BackendVendorOption;
  confidence: number;
  policy?: BackendPolicyResult;
  status: "APPROVED" | "BLOCKED" | "PENDING";
  committed_tx?: string;
  settlement_tx?: string;
  outcome_tx?: string;
  outcome?: BackendOutcomeRecord;
  expected_value?: string;
  timestamp: number;
  alternatives?: Array<{ vendor: BackendVendorOption; reason_rejected: string }>;
}

export interface BackendDecision {
  id: string;
  vendor: string;
  status: "APPROVED" | "BLOCKED" | "PENDING";
  amountEURQ: number;
  intent: string;
  confidence: number;
  policy: {
    budgetOk: boolean;
    vendorAllowed: boolean;
    anomaly: string;
    reputation?: number;
  };
  committedTx?: string;
  settlementTx?: string;
  outcomeTx?: string;
  outcome?: BackendOutcomeRecord;
  blockedReason?: string;
  timestamp: number;
}

export interface BackendAppState {
  agent: string;
  balance: number;
  spent: number;
  dailyLimit: number;
  decisions: BackendDecision[];
}

export type BackendScenarioEventType =
  | "agent.thinking"
  | "decision.pending"
  | "decision.committed"
  | "payment.sent"
  | "decision.outcome"
  | "decision.blocked"
  | "alert.fired"
  | "research.plan"
  | "research.summary";

export interface BackendScenarioEvent {
  type: BackendScenarioEventType;
  payload: unknown;
}

export async function checkHealth(): Promise<boolean> {
  try {
    const res = await fetch(`${apiBase()}/health`);
    return res.ok;
  } catch {
    return false;
  }
}

export async function fetchDashboardState(): Promise<BackendAppState> {
  const res = await fetch(`${apiBase()}/api/state`);
  if (!res.ok) {
    throw new Error(`GET /api/state returned ${res.status}`);
  }
  return res.json() as Promise<BackendAppState>;
}

export async function resetDashboardState(): Promise<BackendAppState> {
  const res = await fetch(`${apiBase()}/api/state/reset`, { method: "POST" });
  if (!res.ok) {
    throw new Error(`POST /api/state/reset returned ${res.status}`);
  }
  return res.json() as Promise<BackendAppState>;
}

function parseSSEChunk(chunk: string): BackendScenarioEvent[] {
  const events: BackendScenarioEvent[] = [];
  for (const block of chunk.split("\n\n")) {
    const trimmed = block.trim();
    if (!trimmed) continue;
    const dataLine = trimmed.split("\n").find((line) => line.startsWith("data: "));
    if (!dataLine) continue;
    try {
      events.push(JSON.parse(dataLine.slice(6)) as BackendScenarioEvent);
    } catch {
      // ignore malformed frames
    }
  }
  return events;
}

export async function runScenarioStream(
  kind: "normal" | "anomaly",
  budget: BudgetTier,
  onEvent: (event: BackendScenarioEvent) => void,
  signal?: AbortSignal,
): Promise<void> {
  const params = new URLSearchParams({ budget });
  if (kind === "anomaly") {
    params.set("scenario", "anomaly");
  }
  const url = `${apiBase()}/api/scenario/run?${params}`;

  const res = await fetch(url, { method: "POST", signal });
  if (!res.ok) {
    const text = await res.text().catch(() => "");
    throw new Error(`POST /api/scenario/run failed (${res.status})${text ? `: ${text}` : ""}`);
  }
  if (!res.body) {
    throw new Error("Scenario stream has no response body");
  }

  const reader = res.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";

  while (true) {
    if (signal?.aborted) {
      await reader.cancel().catch(() => undefined);
      return;
    }
    const { done, value } = await reader.read();
    if (done) break;
    buffer += decoder.decode(value, { stream: true });

    const parts = buffer.split("\n\n");
    buffer = parts.pop() ?? "";
    for (const part of parts) {
      for (const event of parseSSEChunk(part + "\n\n")) {
        onEvent(event);
      }
    }
  }

  if (buffer.trim()) {
    for (const event of parseSSEChunk(buffer)) {
      onEvent(event);
    }
  }
}
