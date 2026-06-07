import type {
  BackendDecision,
  BackendDecisionRecord,
  BackendPolicyResult,
} from "@/lib/api";
import type { DecisionRecord, PolicyCheck } from "@/lib/types";

function policyChecksFromResult(
  policy: BackendPolicyResult | undefined,
  cost: number,
  budget?: number,
): PolicyCheck[] {
  if (!policy) {
    return [];
  }
  const ceiling = budget != null ? `$${cost.toFixed(2)} of $${budget.toFixed(2)} ceiling` : `$${cost.toFixed(2)}`;
  return [
    {
      name: "Budget envelope",
      passed: policy.budget_ok,
      detail: policy.budget_ok ? ceiling : policy.block_reason ?? "Over budget",
    },
    {
      name: "Price anomaly",
      passed: !policy.price_anomaly,
      detail: policy.price_anomaly ? (policy.block_reason ?? "Price spike detected") : "Within tolerance",
    },
  ];
}

function policyChecksFromDashboard(
  policy: BackendDecision["policy"],
  cost: number,
): PolicyCheck[] {
  return [
    {
      name: "Budget envelope",
      passed: policy.budgetOk,
      detail: policy.budgetOk ? `$${cost.toFixed(2)} within envelope` : "Over budget",
    },
    {
      name: "Price anomaly",
      passed: policy.anomaly === "none",
      detail: policy.anomaly === "none" ? "Within tolerance" : "Price anomaly flagged",
    },
    {
      name: "Reputation",
      passed: (policy.reputation ?? 0) >= 2.5,
      detail: policy.reputation != null ? `Score ${policy.reputation.toFixed(1)}` : undefined,
    },
  ];
}

function outcomeStatusFromRecord(
  status: BackendDecisionRecord["status"],
  outcome?: BackendDecisionRecord["outcome"],
): DecisionRecord["outcomeStatus"] {
  if (status === "BLOCKED") return "pending";
  if (!outcome) return "pending";
  if (outcome.verdict.toLowerCase().includes("good")) return "verified";
  if (outcome.verdict.toLowerCase().includes("within")) return "verified";
  return "failed";
}

function reasoningSummary(record: BackendDecisionRecord): string {
  const alt = record.alternatives?.[0];
  if (alt) {
    return `${record.task_intent} Passed over ${alt.vendor.name}: ${alt.reason_rejected}`;
  }
  return record.task_intent;
}

export function mapBackendRecord(
  payload: unknown,
  budget?: number,
): DecisionRecord {
  const record = payload as BackendDecisionRecord;
  const cost = record.vendor_chosen?.price_eurq ?? 0;
  const policyStatus: DecisionRecord["policyStatus"] =
    record.status === "BLOCKED" || record.policy?.approved === false ? "blocked" : "approved";

  return {
    id: record.id,
    timestamp: new Date(record.timestamp).toISOString(),
    task: record.task_intent,
    vendor: record.vendor_chosen.name,
    cost,
    confidence: record.confidence,
    policyStatus,
    outcomeStatus: outcomeStatusFromRecord(record.status, record.outcome),
    reasoningSummary: reasoningSummary(record),
    txPre: record.committed_tx,
    txOutcome: record.outcome_tx,
    settlementTx: record.settlement_tx,
    policyChecks: policyChecksFromResult(record.policy, cost, budget),
    predictedOutcome: record.expected_value ?? record.outcome?.predicted,
    actualOutcome: record.outcome?.actual,
  };
}

export function mapBackendDecision(decision: BackendDecision): DecisionRecord {
  const policyStatus: DecisionRecord["policyStatus"] =
    decision.status === "BLOCKED" ? "blocked" : "approved";
  let outcomeStatus: DecisionRecord["outcomeStatus"] = "pending";
  if (decision.outcome) {
    outcomeStatus = decision.outcome.verdict.toLowerCase().includes("good")
      || decision.outcome.verdict.toLowerCase().includes("within")
      ? "verified"
      : "failed";
  } else if (decision.status === "BLOCKED") {
    outcomeStatus = "pending";
  }

  return {
    id: decision.id,
    timestamp: new Date(decision.timestamp).toISOString(),
    task: decision.intent,
    vendor: decision.vendor,
    cost: decision.amountEURQ,
    confidence: decision.confidence,
    policyStatus,
    outcomeStatus,
    reasoningSummary: decision.blockedReason
      ? `${decision.intent} Blocked: ${decision.blockedReason}`
      : decision.intent,
    txPre: decision.committedTx,
    settlementTx: decision.settlementTx,
    txOutcome: decision.outcomeTx,
    policyChecks: policyChecksFromDashboard(decision.policy, decision.amountEURQ),
    predictedOutcome: decision.outcome?.predicted,
    actualOutcome: decision.outcome?.actual ?? (decision.blockedReason ? "Not executed — blocked at policy layer." : undefined),
  };
}

export function mergeDecisionRecord(
  current: DecisionRecord,
  patch: DecisionRecord,
): DecisionRecord {
  return {
    ...current,
    ...patch,
    policyChecks: patch.policyChecks?.length ? patch.policyChecks : current.policyChecks,
    txPre: patch.txPre ?? current.txPre,
    txOutcome: patch.txOutcome ?? current.txOutcome,
    settlementTx: patch.settlementTx ?? current.settlementTx,
  };
}
