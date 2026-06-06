import type { DecisionStatus } from "@/lib/rationale/types";

const styles: Record<DecisionStatus, string> = {
  APPROVED: "border-approved/40 text-approved bg-approved/10",
  BLOCKED: "border-blocked/40 text-blocked bg-blocked/10",
  PENDING: "border-alert/40 text-alert bg-alert/10",
};

const dot: Record<DecisionStatus, string> = {
  APPROVED: "bg-approved",
  BLOCKED: "bg-blocked",
  PENDING: "bg-alert animate-pulse-dot",
};

export function StatusPill({ status }: { status: DecisionStatus }) {
  return (
    <span
      className={`inline-flex items-center gap-1.5 border px-1.5 py-0.5 text-[10px] font-mono font-medium uppercase tracking-wider ${styles[status]}`}
    >
      <span className={`size-1.5 ${dot[status]}`} />
      {status}
    </span>
  );
}
