import { useMissionStore } from "@/hooks/useMissionStore";
import { cn } from "@/lib/utils";

export function TrustMetrics() {
  const m = useMissionStore((s) => s.metrics);
  const items = [
    { label: "Trust Score", value: `${m.trustScore}`, suffix: " / 100", tone: "text-foreground" },
    { label: "Successful", value: m.successful.toString(), tone: "text-foreground" },
    { label: "Blocked", value: m.blocked.toString(), tone: "text-foreground" },
    { label: "Violations prevented", value: m.violationsPrevented.toString(), tone: "text-foreground" },
  ];
  return (
    <section className="grid grid-cols-2 divide-x divide-border border border-border bg-surface md:grid-cols-4 rounded-[10px] overflow-hidden">
      {items.map((it) => (
        <div key={it.label} className="px-4 py-3">
          <div className="mono-meta">{it.label}</div>
          <div className="mt-1 flex items-baseline gap-1">
            <span className={cn("text-[22px] font-semibold tracking-tight tabular-nums", it.tone)}>
              {it.value}
            </span>
            {it.suffix && <span className="text-[12px] text-muted-foreground">{it.suffix}</span>}
          </div>
        </div>
      ))}
    </section>
  );
}
