import { motion, AnimatePresence } from "framer-motion";
import { Brain, AlertOctagon, CheckCircle2, Send, Anchor, Bell, Clock, ShieldCheck } from "lucide-react";
import { useMissionStore } from "@/hooks/useMissionStore";
import type { ScenarioEvent, ScenarioEventType } from "@/lib/types";
import { cn } from "@/lib/utils";

const META: Record<ScenarioEventType, { icon: React.ComponentType<{ className?: string }>; tone: string; label: string }> = {
  "agent.thinking":    { icon: Brain,         tone: "text-muted-foreground",      label: "agent" },
  "decision.pending":  { icon: Clock,         tone: "text-[#F59E0B]",             label: "decision" },
  "policy.approved":   { icon: ShieldCheck,   tone: "text-[#10B981]",             label: "policy" },
  "policy.blocked":    { icon: AlertOctagon,  tone: "text-[#EF4444]",             label: "policy" },
  "decision.committed":{ icon: Anchor,        tone: "text-foreground",            label: "algorand" },
  "payment.sent":      { icon: Send,          tone: "text-foreground",            label: "x402" },
  "decision.outcome":  { icon: CheckCircle2,  tone: "text-[#10B981]",             label: "outcome" },
  "alert.fired":       { icon: Bell,          tone: "text-[#EF4444]",             label: "alert" },
};

function fmt(ts: string) {
  const d = new Date(ts);
  const hh = String(d.getUTCHours()).padStart(2, "0");
  const mm = String(d.getUTCMinutes()).padStart(2, "0");
  const ss = String(d.getUTCSeconds()).padStart(2, "0");
  return `${hh}:${mm}:${ss}`;
}

function EventRow({ evt }: { evt: ScenarioEvent }) {
  const meta = META[evt.type];
  const Icon = meta.icon;
  return (
    <motion.li
      layout
      initial={{ opacity: 0, y: -4 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.18, ease: "easeOut" }}
      className="row-in grid grid-cols-[56px_14px_1fr] items-start gap-2 py-1.5 text-[13px]"
    >
      <span className="pt-0.5 font-mono text-[11px] tabular-nums text-muted-foreground">
        {fmt(evt.timestamp)}
      </span>
      <Icon className={cn("mt-0.5 h-3.5 w-3.5", meta.tone)} />
      <div className="min-w-0">
        <span className={cn("mr-1.5 font-mono text-[10px] uppercase tracking-wide", meta.tone)}>
          {meta.label}
        </span>
        <span className="text-foreground/90">{evt.message}</span>
      </div>
    </motion.li>
  );
}

export function ReasoningFeed() {
  const events = useMissionStore((s) => s.events);
  return (
    <section className="flex h-full min-h-0 flex-col">
      <header className="flex h-9 items-center justify-between px-3 hairline-b">
        <div className="flex items-center gap-2">
          <h2 className="text-[12px] font-semibold tracking-tight">Agent Activity</h2>
          <span className="font-mono text-[10px] uppercase tracking-wider text-muted-foreground">
            live
          </span>
        </div>
        <span className="font-mono text-[11px] tabular-nums text-muted-foreground">
          {events.length}
        </span>
      </header>
      <div className="min-h-0 flex-1 overflow-y-auto px-3 py-2">
        {events.length === 0 ? (
          <div className="flex h-full flex-col items-center justify-center gap-2 py-8 text-center">
            <Brain className="h-4 w-4 text-muted-foreground/50" />
            <p className="text-[12px] text-muted-foreground">
              Idle. Trigger a scenario to stream agent reasoning.
            </p>
          </div>
        ) : (
          <ul className="divide-y divide-border/60">
            <AnimatePresence initial={false}>
              {events.map((e) => (
                <EventRow key={e.id} evt={e} />
              ))}
            </AnimatePresence>
          </ul>
        )}
      </div>
    </section>
  );
}
