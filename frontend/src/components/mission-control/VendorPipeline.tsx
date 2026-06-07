import { motion } from "framer-motion";
import { Ban, Check, Circle, Loader2 } from "lucide-react";
import { useMissionStore } from "@/hooks/useMissionStore";
import type { VendorPlanStep } from "@/lib/types";
import { cn } from "@/lib/utils";

function StepIcon({ status }: { status: VendorPlanStep["status"] }) {
  if (status === "completed") {
    return <Check className="h-3 w-3 text-background" />;
  }
  if (status === "blocked") {
    return <Ban className="h-3 w-3 text-background" />;
  }
  if (status === "active") {
    return <Loader2 className="h-3 w-3 animate-spin text-foreground" />;
  }
  return <Circle className="h-2 w-2 fill-muted-foreground/30 text-muted-foreground/30" />;
}

export function VendorPipeline() {
  const plan = useMissionStore((s) => s.vendorPlan);
  const running = useMissionStore((s) => s.running);
  const completed = plan.filter((v) => v.status === "completed").length;
  const blocked = plan.filter((v) => v.status === "blocked").length;

  return (
    <section className="flex h-full min-h-0 flex-col">
      <header className="flex shrink-0 flex-col gap-1 px-4 py-4 hairline-b">
        <div className="flex items-center justify-between gap-2">
          <h2 className="text-[15px] font-semibold tracking-tight">Vendor Queue</h2>
          {plan.length > 0 && (
            <span className="font-mono text-[11px] tabular-nums text-muted-foreground">
              {completed + blocked}/{plan.length} done
            </span>
          )}
        </div>
        <p className="text-[12px] text-muted-foreground">
          Knapsack-ordered sources — processed one at a time
        </p>
      </header>

      <div className="panel-scroll-body px-4 py-4">
        {plan.length === 0 ? (
          <div className="flex h-full min-h-[200px] flex-col items-center justify-center gap-2 text-center">
            <p className="text-[12px] text-muted-foreground">
              {running
                ? "Agent is reasoning which vendors fit the budget…"
                : "Run Execute Flow to see the vendor research plan."}
            </p>
          </div>
        ) : (
          <ol className="space-y-0">
            {plan.map((step, i) => (
              <li key={step.id} className="relative">
                {i < plan.length - 1 && (
                  <span
                    aria-hidden
                    className={cn(
                      "absolute left-[11px] top-8 h-[calc(100%-8px)] w-px",
                      step.status === "completed" ? "bg-[#10B981]/50" : "bg-border",
                    )}
                  />
                )}
                <div
                  className={cn(
                    "relative grid grid-cols-[24px_1fr_auto] items-start gap-3 rounded-md px-1 py-2 transition-colors",
                    step.status === "active" && "bg-surface-2/80 ring-1 ring-[#F59E0B]/30",
                    step.status === "blocked" && "bg-[#EF4444]/5",
                  )}
                >
                  <div className="relative pt-0.5">
                    <div
                      className={cn(
                        "relative z-10 grid h-6 w-6 place-items-center rounded-full border transition-colors",
                        step.status === "completed" && "border-[#10B981] bg-[#10B981]",
                        step.status === "blocked" && "border-[#EF4444] bg-[#EF4444]",
                        step.status === "active" && "border-[#F59E0B] bg-background",
                        step.status === "pending" && "border-border bg-background",
                      )}
                    >
                      <StepIcon status={step.status} />
                      {step.status === "active" && (
                        <motion.span
                          className="absolute inset-0 rounded-full ring-2 ring-[#F59E0B]/40"
                          initial={{ scale: 1, opacity: 0.6 }}
                          animate={{ scale: 1.8, opacity: 0 }}
                          transition={{ duration: 1.4, repeat: Infinity, ease: "easeOut" }}
                        />
                      )}
                    </div>
                  </div>

                  <div className="min-w-0 pt-0.5">
                    <div className="flex items-center gap-2">
                      <span className="font-mono text-[10px] tabular-nums text-muted-foreground">
                        #{step.order}
                      </span>
                      <span
                        className={cn(
                          "truncate text-[13px] font-medium",
                          step.status === "pending" && "text-foreground/70",
                          step.status === "active" && "text-foreground",
                          step.status === "completed" && "text-foreground",
                          step.status === "blocked" && "text-[#EF4444]",
                        )}
                      >
                        {step.name}
                      </span>
                    </div>
                    <div className="mt-0.5 text-[11px] text-muted-foreground">
                      {step.status === "pending" && "Queued"}
                      {step.status === "active" && "In progress — trust pipeline running"}
                      {step.status === "completed" && "Purchased & verified"}
                      {step.status === "blocked" && "Blocked by policy"}
                    </div>
                  </div>

                  <span className="pt-0.5 font-mono text-[12px] tabular-nums text-muted-foreground">
                    ${step.price.toFixed(2)}
                  </span>
                </div>
              </li>
            ))}
          </ol>
        )}
      </div>
    </section>
  );
}
