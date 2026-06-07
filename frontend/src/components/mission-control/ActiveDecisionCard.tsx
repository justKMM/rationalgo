import { motion } from "framer-motion";
import { Inbox, Check, X } from "lucide-react";
import { useMissionStore } from "@/hooks/useMissionStore";
import { ConfidenceMeter } from "./ConfidenceMeter";
import { StatusPill } from "./StatusPill";
import { TxHash } from "./TxHash";

export function ActiveDecisionCard() {
  const decision = useMissionStore((s) => s.activeDecision);
  const stage = useMissionStore((s) => s.pipelineStage);

  if (!decision) {
    return (
      <section className="flex h-full flex-col">
        <SectionHeader title="Current Decision" meta="standby" />
        <div className="flex flex-1 flex-col items-center justify-center gap-3 p-8 text-center">
          <Inbox className="h-5 w-5 text-muted-foreground/50" />
          <p className="max-w-xs text-[13px] text-muted-foreground">
            No active decision. When the agent proposes a spend, it surfaces here and runs
            through reasoning, policy, on-chain commit, payment, and verification.
          </p>
        </div>
      </section>
    );
  }

  const policyTone = decision.policyStatus === "blocked" ? "blocked" : "approved";
  const overallTone =
    decision.policyStatus === "blocked"
      ? "blocked"
      : stage === 0
      ? "pending"
      : stage < 6
      ? "pending"
      : "verified";
  const overallLabel =
    decision.policyStatus === "blocked"
      ? "Blocked"
      : stage === 0
      ? "Queued"
      : stage < 6
      ? "In flight"
      : "Settled";

  return (
    <motion.section
      key={decision.id}
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.2 }}
      className="flex h-full min-h-0 flex-col"
    >
      <SectionHeader
        title="Current Decision"
        meta={
          <span className="flex items-center gap-2">
            <span className="font-mono text-[11px] text-muted-foreground">{decision.id}</span>
            <StatusPill tone={overallTone} pulse={stage > 0 && stage < 6}>
              {overallLabel}
            </StatusPill>
          </span>
        }
      />

      <div className="min-h-0 flex-1 overflow-y-auto">
        {/* Task headline */}
        <div className="px-4 py-3 hairline-b">
          <div className="mono-meta mb-1">Task</div>
          <h3 className="text-[15px] font-medium leading-snug text-foreground">
            {decision.task}
          </h3>
        </div>

        {/* Key/value rows — table-aligned */}
        <dl className="divide-y divide-border/60">
          <KVRow label="Vendor" value={decision.vendor} />
          <KVRow label="Cost" value={`$${decision.cost.toFixed(2)}`} mono />
          <KVRow
            label="Confidence"
            valueNode={<ConfidenceMeter value={decision.confidence} className="w-44" />}
          />
          <KVRow
            label="Policy"
            valueNode={
              <div className="flex items-center gap-2">
                <StatusPill tone={policyTone}>{decision.policyStatus}</StatusPill>
                <span className="text-[12px] text-muted-foreground">
                  {(decision.policyChecks ?? []).filter((c) => c.passed).length}
                  /{(decision.policyChecks ?? []).length} checks
                </span>
              </div>
            }
          />
          <KVRow
            label="Outcome"
            valueNode={
              <StatusPill tone={decision.outcomeStatus} pulse={decision.outcomeStatus === "pending" && stage > 0 && stage < 6}>
                {decision.outcomeStatus}
              </StatusPill>
            }
          />
          <KVRow
            label="Pre-payment TX"
            valueNode={<TxHash hash={decision.txPre} />}
          />
          <KVRow
            label="Outcome TX"
            valueNode={<TxHash hash={decision.txOutcome} />}
          />
        </dl>

        {/* Reasoning */}
        <div className="px-4 py-3 hairline-t">
          <div className="mono-meta mb-1.5">Reasoning</div>
          <p className="text-[13px] leading-relaxed text-foreground/85">
            {decision.reasoningSummary}
          </p>
        </div>

        {/* Policy checks compact list */}
        <div className="px-4 pb-4">
          <div className="mono-meta mb-1.5">Policy checks</div>
          <ul className="divide-y divide-border/60 rounded-md border border-border">
            {(decision.policyChecks ?? []).map((c) => (
              <li key={c.name} className="grid grid-cols-[14px_140px_1fr] items-center gap-2 px-3 py-1.5">
                {c.passed ? (
                  <Check className="h-3 w-3 text-[#10B981]" />
                ) : (
                  <X className="h-3 w-3 text-[#EF4444]" />
                )}
                <span className="text-[13px] text-foreground">{c.name}</span>
                <span className="truncate text-[12px] text-muted-foreground">{c.detail}</span>
              </li>
            ))}
          </ul>
        </div>
      </div>
    </motion.section>
  );
}

function SectionHeader({ title, meta }: { title: string; meta?: React.ReactNode }) {
  return (
    <header className="flex h-9 items-center justify-between px-4 hairline-b">
      <h2 className="text-[12px] font-semibold tracking-tight">{title}</h2>
      {meta && <div className="flex items-center">{meta}</div>}
    </header>
  );
}

function KVRow({
  label,
  value,
  valueNode,
  mono,
}: {
  label: string;
  value?: string;
  valueNode?: React.ReactNode;
  mono?: boolean;
}) {
  return (
    <div className="grid grid-cols-[140px_1fr] items-center gap-2 px-4 py-2">
      <dt className="mono-meta">{label}</dt>
      <dd className={mono ? "font-mono text-[13px] tabular-nums text-foreground" : "text-[13px] text-foreground"}>
        {valueNode ?? value}
      </dd>
    </div>
  );
}
