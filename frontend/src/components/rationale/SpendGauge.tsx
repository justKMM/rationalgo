export function SpendGauge({ spent, limit }: { spent: number; limit: number }) {
  const pct = Math.min(100, (spent / limit) * 100);
  const color =
    pct > 90 ? "bg-blocked" : pct > 70 ? "bg-alert" : "bg-approved";
  return (
    <div className="flex items-center gap-3">
      <span className="text-[10px] font-mono uppercase tracking-wider text-muted-foreground">
        daily
      </span>
      <div className="relative h-1.5 w-40 bg-surface-2 border border-border">
        <div
          className={`absolute inset-y-0 left-0 ${color} transition-all duration-500`}
          style={{ width: `${pct}%` }}
        />
      </div>
      <span className="font-mono tnum text-xs text-foreground">
        {spent.toFixed(2)}
        <span className="text-muted-foreground"> / {limit.toFixed(2)}</span>
        <span className="ml-1 text-muted-foreground">EURQ</span>
      </span>
    </div>
  );
}
