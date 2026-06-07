import { cn } from "@/lib/utils";

export function ConfidenceMeter({ value, className }: { value: number; className?: string }) {
  const pct = Math.round(value * 100);
  const bar =
    value >= 0.8 ? "bg-[#10B981]" : value >= 0.6 ? "bg-[#F59E0B]" : "bg-[#EF4444]";
  return (
    <div className={cn("flex items-center gap-2", className)}>
      <div className="relative h-1 flex-1 overflow-hidden rounded-full bg-surface-2">
        <div
          className={cn("h-full rounded-full transition-all duration-700", bar)}
          style={{ width: `${pct}%` }}
        />
      </div>
      <span className="font-mono text-[11px] tabular-nums text-foreground/85">{pct}%</span>
    </div>
  );
}
