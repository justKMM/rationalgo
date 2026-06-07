import { useMissionStore } from "@/hooks/useMissionStore";
import { BUDGET_TIERS, type BudgetTier } from "@/lib/budget";
import { cn } from "@/lib/utils";

const TIER_ORDER: BudgetTier[] = ["cheapass", "mid", "luxury"];

const TIER_STYLE: Record<
  BudgetTier,
  { active: string; ring: string; idle: string }
> = {
  cheapass: {
    active: "bg-[#10B981] text-white shadow-[0_0_20px_rgba(16,185,129,0.35)]",
    ring: "ring-[#10B981]/60",
    idle: "text-[#10B981] hover:bg-[#10B981]/10",
  },
  mid: {
    active: "bg-[#06B6D4] text-white shadow-[0_0_20px_rgba(6,182,212,0.4)]",
    ring: "ring-[#06B6D4]/60",
    idle: "text-[#06B6D4] hover:bg-[#06B6D4]/10",
  },
  luxury: {
    active: "bg-[#F59E0B] text-[#0B0D12] shadow-[0_0_20px_rgba(245,158,11,0.35)]",
    ring: "ring-[#F59E0B]/60",
    idle: "text-[#F59E0B] hover:bg-[#F59E0B]/10",
  },
};

export function BudgetTierBar() {
  const running = useMissionStore((s) => s.running);
  const budgetTier = useMissionStore((s) => s.budgetTier);
  const setBudgetTier = useMissionStore((s) => s.setBudgetTier);

  return (
    <div className="hairline-b bg-surface-2/80 px-4 py-3">
      <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <div className="text-[11px] font-semibold uppercase tracking-[0.14em] text-foreground">
            Research budget
          </div>
          <p className="mt-0.5 text-[12px] text-muted-foreground">
            Knapsack picks vendors that fit this envelope — change before Execute Flow
          </p>
        </div>
        <div
          className="flex flex-wrap gap-2"
          role="group"
          aria-label="Research budget tier"
        >
          {TIER_ORDER.map((tier) => {
            const meta = BUDGET_TIERS[tier];
            const style = TIER_STYLE[tier];
            const active = budgetTier === tier;
            return (
              <button
                key={tier}
                type="button"
                disabled={running}
                onClick={() => setBudgetTier(tier)}
                className={cn(
                  "inline-flex min-w-[132px] flex-col items-center rounded-lg border px-4 py-2.5 text-center transition disabled:opacity-50",
                  active
                    ? cn("border-transparent ring-2", style.active, style.ring)
                    : cn("border-border bg-surface", style.idle),
                )}
              >
                <span className="text-[13px] font-semibold leading-tight">{meta.label}</span>
                <span className="mt-0.5 font-mono text-[15px] font-bold tabular-nums leading-none">
                  €{meta.eur}
                </span>
              </button>
            );
          })}
        </div>
      </div>
    </div>
  );
}
