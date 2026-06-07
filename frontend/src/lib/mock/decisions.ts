import type { DecisionRecord } from "../types";
import { algoTx, seededRng } from "./generators";

const VENDORS = [
  { name: "OpenAI API", task: "Generate marketing copy variants" },
  { name: "AWS S3", task: "Store inference artifacts" },
  { name: "Pinecone", task: "Upsert embedding batch" },
  { name: "Twilio SMS", task: "Send customer verification" },
  { name: "Replicate", task: "Run image upscaler model" },
  { name: "Stripe Top-up", task: "Replenish operating float" },
  { name: "Anthropic API", task: "Long-context document review" },
  { name: "Cloudflare R2", task: "Archive transcript bundle" },
  { name: "Mapbox", task: "Geocode delivery batch" },
  { name: "SendGrid", task: "Dispatch transactional email" },
  { name: "Datadog", task: "Push custom metrics" },
  { name: "Linode GPU", task: "Spin up fine-tune worker" },
];

// Fixed epoch base so SSR and client render identical timestamps.
const BASE_EPOCH = Date.UTC(2026, 5, 6, 12, 0, 0);

function hoursAgo(h: number) {
  return new Date(BASE_EPOCH - h * 3600_000).toISOString();
}

export function seedHistory(): DecisionRecord[] {
  const rng = seededRng(0xC0FFEE);
  return VENDORS.map((v, i): DecisionRecord => {
    const blocked = i === 2 || i === 7;
    const failed = i === 5;
    const cost = Math.round((0.42 + rng() * 18) * 100) / 100;
    const confidence = blocked
      ? 0.41 + rng() * 0.15
      : 0.82 + rng() * 0.16;
    const id = `DEC-${(0x1A4F + i * 137).toString(16).toUpperCase().padStart(6, "0")}`;
    return {
      id,
      timestamp: hoursAgo(0.6 + i * 1.7),
      task: v.task,
      vendor: v.name,
      cost,
      confidence: Math.round(confidence * 100) / 100,
      policyStatus: blocked ? "blocked" : "approved",
      outcomeStatus: blocked ? "pending" : failed ? "failed" : "verified",
      reasoningSummary: `Agent selected ${v.name} to ${v.task.toLowerCase()}. Cost-per-utility ranked best vs. ${2 + (i % 3)} alternatives; vendor within trust allow-list.`,
      txPre: blocked ? undefined : algoTx(rng),
      txOutcome: blocked ? undefined : failed ? algoTx(rng) : algoTx(rng),
      policyChecks: [
        { name: "Budget envelope", passed: !blocked, detail: blocked ? "Exceeds per-task ceiling ($25)" : "Within $25 ceiling" },
        { name: "Vendor allow-list", passed: i !== 7, detail: i === 7 ? "Vendor reputation below threshold" : "Verified vendor" },
        { name: "Rate limit", passed: i !== 2, detail: i === 2 ? "Call rate exceeded window" : "OK" },
        { name: "Risk model", passed: !blocked, detail: blocked ? "Risk score 0.71" : "Risk score < 0.3" },
      ],
      predictedOutcome: "Task completes within SLA, artifact returned.",
      actualOutcome: failed
        ? "Vendor returned 5xx after retry; no artifact."
        : blocked
        ? "Not executed — blocked at policy layer."
        : "Artifact returned, hash matches predicted schema.",
    };
  });
}
