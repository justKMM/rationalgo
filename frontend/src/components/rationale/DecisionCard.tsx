import { Lock } from "lucide-react";
import type { Decision } from "@/lib/rationale/types";
import { StatusPill } from "./StatusPill";
import { truncateHash } from "@/lib/rationale/mock";

function formatTime(ts: number) {
  const d = new Date(ts);
  return d.toLocaleTimeString("en-GB", { hour12: false });
}

interface Props {
  decision: Decision;
  selected: boolean;
  onClick: () => void;
}

export function DecisionCard({ decision, selected, onClick }: Props) {
  const accent =
    decision.status === "APPROVED"
      ? "border-l-approved"
      : decision.status === "BLOCKED"
      ? "border-l-blocked"
      : "border-l-alert";

  return (
    <button
      onClick={onClick}
      className={`group w-full text-left border border-border border-l-2 ${accent} bg-surface hover:bg-surface-2 transition-colors animate-fade-slide-in ${
        selected ? "outline outline-1 outline-approved/60" : ""
      }`}
    >
      <div className="px-4 py-3 flex flex-col gap-2">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2.5">
            <span className="font-mono text-xs text-foreground font-medium">
              {decision.vendor}
            </span>
            <StatusPill status={decision.status} />
          </div>
          <div className="flex items-center gap-3 font-mono tnum text-xs">
            <span
              className={
                decision.status === "BLOCKED"
                  ? "text-muted-foreground line-through"
                  : "text-foreground"
              }
            >
              {decision.amountEURQ.toFixed(2)}
              <span className="text-muted-foreground ml-1">EURQ</span>
            </span>
          </div>
        </div>

        <p className="text-sm text-foreground/90 leading-snug">{decision.intent}</p>

        <div className="flex items-center justify-between text-[10px] font-mono text-muted-foreground pt-1">
          <div className="flex items-center gap-2">
            <Lock className="size-3 text-approved/80" />
            <span className="text-foreground/70">{truncateHash(decision.reasoningHash)}</span>
            <span className="uppercase tracking-wider">
              reasoning committed pre-outcome
            </span>
          </div>
          <div className="flex items-center gap-3">
            {decision.outcome && (
              <span className="text-approved uppercase tracking-wider">
                outcome · {decision.outcome.actual}
              </span>
            )}
            <span>{formatTime(decision.timestamp)}</span>
          </div>
        </div>
      </div>
    </button>
  );
}
