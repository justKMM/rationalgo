import { AlertTriangle } from "lucide-react";
import type { Alert } from "@/lib/rationale/types";

interface Props {
  dailyLimit: number;
  spent: number;
  allowed: string[];
  blocked: string[];
  alerts: Alert[];
}

export function PolicyPanel({ dailyLimit, spent, allowed, blocked, alerts }: Props) {
  return (
    <section className="border-b border-border">
      <div className="flex items-center justify-between border-b border-border bg-surface px-4 h-9">
        <h2 className="text-[10px] font-mono uppercase tracking-[0.18em] text-muted-foreground">
          policy
        </h2>
        <span className="text-[10px] font-mono tnum text-muted-foreground">enforced</span>
      </div>

      <div className="px-4 py-3 space-y-3">
        <div>
          <div className="text-[10px] font-mono uppercase tracking-wider text-muted-foreground mb-1">
            daily limit
          </div>
          <div className="font-mono tnum text-xs text-foreground">
            {spent.toFixed(2)}
            <span className="text-muted-foreground"> / {dailyLimit.toFixed(2)} EURQ</span>
          </div>
        </div>

        <div>
          <div className="text-[10px] font-mono uppercase tracking-wider text-muted-foreground mb-1">
            allowed vendors
          </div>
          <div className="flex flex-wrap gap-1">
            {allowed.map((v) => (
              <span
                key={v}
                className="border border-approved/30 text-approved px-1.5 py-0.5 text-[10px] font-mono"
              >
                {v}
              </span>
            ))}
          </div>
        </div>

        <div>
          <div className="text-[10px] font-mono uppercase tracking-wider text-muted-foreground mb-1">
            blocked vendors
          </div>
          <div className="flex flex-wrap gap-1">
            {blocked.map((v) => (
              <span
                key={v}
                className="border border-blocked/30 text-blocked px-1.5 py-0.5 text-[10px] font-mono"
              >
                {v}
              </span>
            ))}
          </div>
        </div>

        <div>
          <div className="text-[10px] font-mono uppercase tracking-wider text-muted-foreground mb-1.5">
            active alerts
          </div>
          <ul className="space-y-1">
            {alerts.length === 0 && (
              <li className="text-xs text-muted-foreground font-mono">— none</li>
            )}
            {alerts.map((a) => (
              <li
                key={a.id}
                className={`flex items-start gap-2 border border-l-2 px-2 py-1.5 animate-fade-slide-in ${
                  a.level === "amber"
                    ? "border-alert/30 border-l-alert bg-alert/5"
                    : "border-blocked/30 border-l-blocked bg-blocked/5"
                }`}
              >
                <AlertTriangle
                  className={`size-3 mt-0.5 shrink-0 ${
                    a.level === "amber" ? "text-alert" : "text-blocked"
                  }`}
                />
                <span className="text-[11px] text-foreground/90 leading-snug">
                  {a.message}
                </span>
              </li>
            ))}
          </ul>
        </div>
      </div>
    </section>
  );
}
