import { motion } from "framer-motion";
import { Check, Minus } from "lucide-react";
import { useMissionStore } from "@/hooks/useMissionStore";
import { PIPELINE_STAGES } from "@/lib/types";
import { cn } from "@/lib/utils";

export function TrustPipeline() {
  const stage = useMissionStore((s) => s.pipelineStage);
  const decision = useMissionStore((s) => s.activeDecision);
  const blocked = decision?.policyStatus === "blocked";
  const failed = decision?.outcomeStatus === "failed";

  return (
    <section className="flex h-full min-h-0 flex-col">
      <header className="flex h-9 items-center justify-between px-3 hairline-b">
        <h2 className="text-[12px] font-semibold tracking-tight">Trust Pipeline</h2>
        <span className="font-mono text-[11px] tabular-nums text-muted-foreground">
          {Math.min(stage, 6)}/6
        </span>
      </header>
      <ol className="panel-scroll-body px-3 py-3">
        {PIPELINE_STAGES.map((s, i) => {
          const idx = i + 1;
          const active = stage === idx;
          const done = stage > idx;
          const skipped = blocked && idx > 2;
          const errored = (blocked && idx === 2) || (failed && idx === 5);
          const isLast = i === PIPELINE_STAGES.length - 1;

          return (
            <li key={s.key} className="relative grid grid-cols-[16px_1fr] gap-3">
              {/* Rail */}
              {!isLast && (
                <span
                  aria-hidden
                  className={cn(
                    "absolute left-[7px] top-5 h-[calc(100%-12px)] w-px",
                    done && !skipped ? "bg-foreground/40" : "bg-border",
                  )}
                />
              )}

              {/* Node */}
              <div className="relative pt-1.5">
                <div
                  className={cn(
                    "relative z-10 grid h-3.5 w-3.5 place-items-center rounded-full border transition-colors",
                    errored
                      ? "border-[#EF4444] bg-[#EF4444]"
                      : done
                      ? "border-foreground bg-foreground"
                      : active
                      ? "border-foreground bg-background"
                      : skipped
                      ? "border-border bg-background"
                      : "border-border bg-background",
                  )}
                >
                  {done && !errored && <Check className="h-2.5 w-2.5 text-background" />}
                  {skipped && <Minus className="h-2 w-2 text-muted-foreground" />}
                  {active && !errored && (
                    <motion.span
                      className="absolute inset-0 rounded-full ring-2 ring-foreground/40"
                      initial={{ scale: 1, opacity: 0.6 }}
                      animate={{ scale: 1.9, opacity: 0 }}
                      transition={{ duration: 1.4, repeat: Infinity, ease: "easeOut" }}
                    />
                  )}
                </div>
              </div>

              {/* Body */}
              <div className="min-w-0 pb-4 pt-0.5">
                <div className="flex items-center justify-between gap-2">
                  <span
                    className={cn(
                      "text-[13px] font-medium leading-tight",
                      skipped && "text-muted-foreground/60 line-through",
                      errored && "text-[#EF4444]",
                      !skipped && !errored && (done || active) ? "text-foreground" : "text-foreground/70",
                    )}
                  >
                    {s.label}
                  </span>
                  {active && !errored && (
                    <span className="font-mono text-[10px] uppercase tracking-wider text-[#F59E0B]">
                      now
                    </span>
                  )}
                </div>
                <div
                  className={cn(
                    "mt-0.5 text-[12px] leading-snug",
                    skipped ? "text-muted-foreground/50" : "text-muted-foreground",
                    errored && "text-[#EF4444]/80",
                  )}
                >
                  {errored ? "Halted — see alerts" : s.desc}
                </div>
              </div>
            </li>
          );
        })}
      </ol>
    </section>
  );
}
