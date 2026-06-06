export type DecisionStatus = "APPROVED" | "BLOCKED" | "PENDING";

export interface Alternative {
  name: string;
  reason: string;
}

export interface PolicyChecks {
  budgetOk: boolean;
  reputation: number;
  anomaly: "none" | "flagged";
  vendorAllowed: boolean;
}

export interface Outcome {
  predicted: string;
  actual: string;
  verdict: string;
  trustDelta: number;
}

export interface Decision {
  id: string;
  vendor: string;
  status: DecisionStatus;
  amountEURQ: number;
  intent: string;
  alternatives: Alternative[];
  expectedValue: string;
  confidence: number;
  policy: PolicyChecks;
  reasoningHash: string;
  round: number;
  timestamp: number;
  outcome?: Outcome;
  blockedReason?: string;
}

export interface Vendor {
  name: string;
  score: number;
  lastDelta?: { dir: "up" | "down"; value: number; at: number };
}

export interface Alert {
  id: string;
  level: "amber" | "red";
  message: string;
  at: number;
}

export interface AppState {
  agent: string;
  balance: number;
  spent: number;
  dailyLimit: number;
  decisions: Decision[];
  vendors: Vendor[];
  allowedVendors: string[];
  blockedVendors: string[];
  alerts: Alert[];
  selectedId: string | null;
}
