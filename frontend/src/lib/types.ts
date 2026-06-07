export interface DecisionRecord {
  id: string;
  timestamp: string;
  task: string;
  vendor: string;
  cost: number;
  confidence: number;
  policyStatus: "approved" | "blocked";
  outcomeStatus: "verified" | "failed" | "pending";
  reasoningSummary: string;
  txPre?: string;
  txOutcome?: string;
  settlementTx?: string;
  policyChecks?: PolicyCheck[];
  predictedOutcome?: string;
  actualOutcome?: string;
}

export interface PolicyCheck {
  name: string;
  passed: boolean;
  detail?: string;
}

export type ScenarioEventType =
  | "agent.thinking"
  | "decision.pending"
  | "policy.approved"
  | "policy.blocked"
  | "decision.committed"
  | "decision.blocked"
  | "payment.sent"
  | "decision.outcome"
  | "alert.fired"
  | "research.summary";

export interface ScenarioEvent {
  id: string;
  type: ScenarioEventType;
  timestamp: string;
  message: string;
}

export const PIPELINE_STAGES = [
  { key: "reasoning", label: "Reasoning Generated", desc: "Agent constructs intent" },
  { key: "policy", label: "Policy Evaluated", desc: "Guardrails & budget checks" },
  { key: "commit-pre", label: "Algorand Commit", desc: "Intent anchored on-chain" },
  { key: "payment", label: "x402 Payment", desc: "Settlement executed" },
  { key: "verify", label: "Outcome Verification", desc: "Result validated vs intent" },
  { key: "commit-post", label: "Outcome Commit", desc: "Receipt anchored on-chain" },
] as const;

export type PipelineStageKey = (typeof PIPELINE_STAGES)[number]["key"];
