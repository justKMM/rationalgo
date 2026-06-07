import { useMissionStore } from "@/hooks/useMissionStore";

export function OpsCorner() {
  const m = useMissionStore((s) => s.metrics);

  const stats = [
    { label: "Trust", value: m.trustScore },
    { label: "OK", value: m.successful },
    { label: "Blocked", value: m.blocked },
    { label: "Caught", value: m.violationsPrevented },
  ];

  return (
    <aside className="absolute bottom-3 right-3 z-10 w-[132px] rounded-md border border-border bg-surface/95 p-2 shadow-sm backdrop-blur-sm">
      <div className="grid grid-cols-2 gap-x-2 gap-y-1.5">
        {stats.map((s) => (
          <div key={s.label}>
            <div className="font-mono text-[9px] uppercase tracking-wide text-muted-foreground">
              {s.label}
            </div>
            <div className="text-[13px] font-semibold tabular-nums leading-tight">{s.value}</div>
          </div>
        ))}
      </div>
    </aside>
  );
}
