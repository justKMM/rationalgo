import type { DecisionRecord, ScenarioEvent } from "../types";
import { algoTx, seededRng, shortId } from "./generators";

export type ScenarioKind = "normal" | "anomaly";

export interface ScenarioStep {
  delay: number;
  stage: number;
  event?: Omit<ScenarioEvent, "id" | "timestamp">;
  patch?: Partial<DecisionRecord>;
  finalize?: boolean;
}

export interface ScenarioScript {
  decision: DecisionRecord;
  steps: ScenarioStep[];
}

export function buildScenario(kind: ScenarioKind): ScenarioScript {
  // Use Date.now seed here — only called on user click (client-only), no SSR concern.
  const rng = seededRng(Date.now() & 0xFFFF);
  if (kind === "normal") {
    const id = shortId(rng);
    const decision: DecisionRecord = {
      id,
      timestamp: new Date().toISOString(),
      task: "Procure GPT-4o inference for nightly summary batch",
      vendor: "OpenAI API",
      cost: 4.82,
      confidence: 0.93,
      policyStatus: "approved",
      outcomeStatus: "pending",
      reasoningSummary:
        "Nightly summary job requires 1.2M tokens. OpenAI gpt-4o ranked best on quality/$ vs Anthropic and local Llama-3. Within budget envelope.",
      policyChecks: [
        { name: "Budget envelope", passed: true, detail: "$4.82 of $25 ceiling" },
        { name: "Vendor allow-list", passed: true, detail: "Verified vendor" },
        { name: "Rate limit", passed: true, detail: "OK" },
        { name: "Risk model", passed: true, detail: "Risk score 0.08" },
      ],
      predictedOutcome: "1.2M tokens processed, summary artifact returned within 90s.",
    };
    const txPre = algoTx(rng);
    const txOutcome = algoTx(rng);
    return {
      decision,
      steps: [
        { delay: 0, stage: 1, event: { type: "agent.thinking", message: "Evaluating vendors for nightly summary batch." } },
        { delay: 700, stage: 1, event: { type: "agent.thinking", message: "Scoring gpt-4o vs claude-3.5 vs Llama-3 (local)." } },
        { delay: 1400, stage: 1, event: { type: "decision.pending", message: "Selected OpenAI gpt-4o · $4.82 · confidence 0.93" } },
        { delay: 2100, stage: 2, event: { type: "policy.approved", message: "Policy: 4/4 checks passed." } },
        { delay: 2800, stage: 3, event: { type: "decision.committed", message: `Reasoning anchored on Algorand · ${txPre.slice(0, 10)}…` }, patch: { txPre } },
        { delay: 3700, stage: 4, event: { type: "payment.sent", message: "x402 payment broadcast to vendor endpoint." } },
        { delay: 4600, stage: 5, event: { type: "decision.outcome", message: "Outcome verified · artifact hash matches schema." }, patch: { outcomeStatus: "verified", actualOutcome: "Summary artifact returned in 78s. Hash matches predicted schema." } },
        { delay: 5400, stage: 6, event: { type: "decision.committed", message: `Outcome receipt anchored · ${txOutcome.slice(0, 10)}…` }, patch: { txOutcome }, finalize: true },
      ],
    };
  }

  const id = shortId(rng);
  const decision: DecisionRecord = {
    id,
    timestamp: new Date().toISOString(),
    task: "Bulk SMS campaign to 48,000 unverified numbers",
    vendor: "Twilio SMS",
    cost: 312.4,
    confidence: 0.48,
    policyStatus: "approved",
    outcomeStatus: "pending",
    reasoningSummary:
      "Agent proposed bulk SMS to 48k recipients sourced from scraped list. Selected Twilio for throughput. Estimated cost $312.40.",
    policyChecks: [
      { name: "Budget envelope", passed: false, detail: "$312.40 exceeds $25 per-task ceiling" },
      { name: "Vendor allow-list", passed: true, detail: "Verified vendor" },
      { name: "Rate limit", passed: false, detail: "Recipient volume 12× window cap" },
      { name: "Risk model", passed: false, detail: "Risk score 0.81 — unverified recipient list" },
    ],
    predictedOutcome: "48k SMS dispatched.",
  };
  return {
    decision,
    steps: [
      { delay: 0, stage: 1, event: { type: "agent.thinking", message: "Planning outbound SMS campaign." } },
      { delay: 700, stage: 1, event: { type: "agent.thinking", message: "Recipient list scraped · 48,000 entries · verification unknown." } },
      { delay: 1400, stage: 1, event: { type: "decision.pending", message: "Selected Twilio SMS · estimated $312.40." } },
      { delay: 2100, stage: 2, event: { type: "policy.blocked", message: "Blocked: budget, rate-limit, risk checks failed." } },
      { delay: 2700, stage: 2, event: { type: "alert.fired", message: "Operator alerted. No on-chain commit, no payment sent." }, patch: { policyStatus: "blocked", outcomeStatus: "pending", actualOutcome: "Not executed — blocked at policy layer before any x402 spend." }, finalize: true },
    ],
  };
}
