import { useEffect } from "react";
import { X, Lock, Check, AlertTriangle, Ban } from "lucide-react";
import type { Decision } from "@/lib/rationale/types";
import { StatusPill } from "./StatusPill";

interface Props {
  decision: Decision | null;
  onClose: () => void;
}

function Row({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="flex items-baseline justify-between gap-4 py-2 border-b border-border/60">
      <span className="text-[10px] font-mono uppercase tracking-wider text-muted-foreground shrink-0">
        {label}
      </span>
      <span className="text-xs text-foreground font-mono tnum text-right">{children}</span>
    </div>
  );
}

function CheckRow({ ok, label, value }: { ok: boolean; label: string; value: string }) {
  return (
    <div className="flex items-center justify-between py-1.5 text-xs font-mono">
      <div className="flex items-center gap-2 text-foreground/80">
        {ok ? (
          <Check className="size-3 text-approved" />
        ) : (
          <Ban className="size-3 text-blocked" />
        )}
        <span className="uppercase tracking-wider text-[10px] text-muted-foreground">
          {label}
        </span>
      </div>
      <span className={ok ? "text-foreground" : "text-blocked"}>{value}</span>
    </div>
  );
}

export function DecisionDrawer({ decision, onClose }: Props) {
  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      if (e.key === "Escape") onClose();
    }
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [onClose]);

  if (!decision) return null;

  return (
    <>
      <div
        className="fixed inset-0 z-40 bg-black/50"
        onClick={onClose}
        aria-hidden
      />
      <aside className="fixed right-0 top-0 z-50 h-full w-full max-w-[560px] bg-surface border-l border-border overflow-y-auto animate-drawer-in">
        <div className="sticky top-0 flex items-center justify-between border-b border-border bg-surface px-5 h-12 z-10">
          <div className="flex items-center gap-3">
            <span className="text-[10px] font-mono uppercase tracking-[0.18em] text-muted-foreground">
              decision record
            </span>
            <span className="font-mono text-xs text-foreground">{decision.vendor}</span>
            <StatusPill status={decision.status} />
          </div>
          <button
            onClick={onClose}
            className="text-muted-foreground hover:text-foreground"
            aria-label="Close"
          >
            <X className="size-4" />
          </button>
        </div>

        <div className="px-5 py-4 space-y-5">
          <section>
            <h3 className="text-[10px] font-mono uppercase tracking-wider text-muted-foreground mb-2">
              intent
            </h3>
            <p className="text-sm text-foreground leading-relaxed">{decision.intent}</p>
          </section>

          <section>
            <h3 className="text-[10px] font-mono uppercase tracking-wider text-muted-foreground mb-2">
              alternatives considered
            </h3>
            <ul className="space-y-1.5">
              {decision.alternatives.map((alt, i) => (
                <li
                  key={i}
                  className="border border-border border-l-2 border-l-blocked/60 bg-background/40 px-3 py-2"
                >
                  <div className="flex items-baseline justify-between gap-3">
                    <span className="font-mono text-xs text-foreground">{alt.name}</span>
                    <span className="text-[10px] font-mono uppercase tracking-wider text-blocked">
                      rejected
                    </span>
                  </div>
                  <p className="text-xs text-muted-foreground mt-0.5">{alt.reason}</p>
                </li>
              ))}
            </ul>
          </section>

          <section className="grid grid-cols-2 gap-x-6">
            <Row label="expected value">{decision.expectedValue}</Row>
            <Row label="confidence">{decision.confidence.toFixed(2)}</Row>
            <Row label="cost">
              {decision.amountEURQ.toFixed(2)}
              <span className="text-muted-foreground ml-1">EURQ</span>
            </Row>
            <Row label="timestamp">
              {new Date(decision.timestamp).toLocaleTimeString("en-GB", { hour12: false })}
            </Row>
          </section>

          <section>
            <h3 className="text-[10px] font-mono uppercase tracking-wider text-muted-foreground mb-2">
              policy checks
            </h3>
            <div className="border border-border bg-background/40 px-3 py-2">
              <CheckRow
                ok={decision.policy.budgetOk}
                label="daily budget"
                value={decision.policy.budgetOk ? "ok" : "exceeded"}
              />
              <CheckRow
                ok={decision.policy.vendorAllowed}
                label="vendor allowlist"
                value={decision.policy.vendorAllowed ? "allowed" : "not allowed"}
              />
              <CheckRow
                ok={decision.policy.reputation >= 2.5}
                label="vendor reputation"
                value={`${decision.policy.reputation.toFixed(1)} / 5`}
              />
              <CheckRow
                ok={decision.policy.anomaly === "none"}
                label="price anomaly"
                value={decision.policy.anomaly}
              />
            </div>
            {decision.blockedReason && (
              <p className="mt-2 flex items-start gap-2 text-xs text-blocked">
                <AlertTriangle className="size-3 mt-0.5 shrink-0" />
                {decision.blockedReason}
              </p>
            )}
          </section>

          <section>
            <h3 className="text-[10px] font-mono uppercase tracking-wider text-muted-foreground mb-2 flex items-center gap-2">
              <Lock className="size-3 text-approved" />
              on-chain reasoning commit
            </h3>
            <div className="border border-border bg-background/40 px-3 py-2 space-y-1.5">
              <Row label="reasoning hash">
                <span className="text-foreground break-all">{decision.reasoningHash}</span>
              </Row>
              <Row label="algorand round">{decision.round.toLocaleString()}</Row>
              <Row label="commit order">pre-outcome</Row>
            </div>
          </section>

          {decision.outcome ? (
            <section className="border border-approved/40 bg-approved/5 p-3">
              <h3 className="text-[10px] font-mono uppercase tracking-wider text-approved mb-2">
                outcome
              </h3>
              <div className="grid grid-cols-3 gap-3 font-mono">
                <div>
                  <div className="text-[10px] uppercase text-muted-foreground">predicted</div>
                  <div className="text-base tnum text-foreground">
                    {decision.outcome.predicted}
                  </div>
                </div>
                <div>
                  <div className="text-[10px] uppercase text-muted-foreground">actual</div>
                  <div className="text-base tnum text-approved">{decision.outcome.actual}</div>
                </div>
                <div>
                  <div className="text-[10px] uppercase text-muted-foreground">trust Δ</div>
                  <div
                    className={`text-base tnum ${
                      decision.outcome.trustDelta >= 0 ? "text-approved" : "text-blocked"
                    }`}
                  >
                    {decision.outcome.trustDelta >= 0 ? "+" : ""}
                    {decision.outcome.trustDelta.toFixed(2)}
                  </div>
                </div>
              </div>
              <div className="mt-2 text-xs text-foreground/80">
                verdict: <span className="text-approved">{decision.outcome.verdict}</span>
              </div>
            </section>
          ) : decision.status === "PENDING" || decision.status === "APPROVED" ? (
            <section className="border border-dashed border-border p-3 text-xs font-mono text-muted-foreground uppercase tracking-wider">
              awaiting outcome…
            </section>
          ) : null}
        </div>
      </aside>
    </>
  );
}
