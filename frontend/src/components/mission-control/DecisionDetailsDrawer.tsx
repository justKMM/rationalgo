import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetDescription } from "@/components/ui/sheet";
import { useMissionStore } from "@/hooks/useMissionStore";
import { StatusPill } from "./StatusPill";
import { TxHash } from "./TxHash";
import { Check, X, ExternalLink } from "lucide-react";

function fmt(ts: string) {
  const d = new Date(ts);
  return d.toISOString().replace("T", " · ").slice(0, 19) + " UTC";
}

export function DecisionDetailsDrawer() {
  const id = useMissionStore((s) => s.selectedDecisionId);
  const history = useMissionStore((s) => s.history);
  const close = () => useMissionStore.getState().selectDecision(null);
  const d = history.find((x) => x.id === id) ?? null;

  return (
    <Sheet open={!!d} onOpenChange={(o) => !o && close()}>
      <SheetContent className="w-full sm:max-w-lg overflow-y-auto bg-surface border-l border-border p-0">
        {d && (
          <>
            <SheetHeader className="space-y-2 p-5 hairline-b">
              <div className="flex items-center gap-2">
                <span className="mono-meta">DECISION</span>
                <span className="font-mono text-[12px] text-foreground/80">{d.id}</span>
              </div>
              <SheetTitle className="text-[15px] font-medium leading-snug">{d.task}</SheetTitle>
              <SheetDescription className="font-mono text-[11px] text-muted-foreground">
                {fmt(d.timestamp)}
              </SheetDescription>
              <div className="flex flex-wrap gap-1.5 pt-1">
                <StatusPill tone={d.policyStatus === "blocked" ? "blocked" : "approved"}>
                  policy · {d.policyStatus}
                </StatusPill>
                <StatusPill tone={d.outcomeStatus}>outcome · {d.outcomeStatus}</StatusPill>
              </div>
            </SheetHeader>

            <div className="divide-y divide-border/60">
              <dl className="divide-y divide-border/60">
                <KV label="Vendor" value={d.vendor} />
                <KV label="Cost" value={`$${d.cost.toFixed(2)}`} mono />
                <KV label="Confidence" value={`${Math.round(d.confidence * 100)}%`} mono />
              </dl>

              <Section title="Reasoning">
                <p className="text-[13px] leading-relaxed text-foreground/85">{d.reasoningSummary}</p>
              </Section>

              <Section title="Policy checks">
                <ul className="divide-y divide-border/60 rounded-md border border-border">
                  {(d.policyChecks ?? []).map((c) => (
                    <li
                      key={c.name}
                      className="grid grid-cols-[14px_140px_1fr] items-center gap-2 px-3 py-1.5"
                    >
                      {c.passed ? (
                        <Check className="h-3 w-3 text-[#10B981]" />
                      ) : (
                        <X className="h-3 w-3 text-[#EF4444]" />
                      )}
                      <span className="text-[13px]">{c.name}</span>
                      <span className="truncate text-[12px] text-muted-foreground">{c.detail}</span>
                    </li>
                  ))}
                </ul>
              </Section>

              <Section title="Predicted vs actual">
                <div className="space-y-2">
                  <Bubble title="Predicted">{d.predictedOutcome ?? "—"}</Bubble>
                  <Bubble title="Actual" tone={d.outcomeStatus}>
                    {d.actualOutcome ?? "Pending."}
                  </Bubble>
                </div>
              </Section>

              <Section title="On-chain provenance">
                <div className="space-y-1.5">
                  <TxLine label="Pre-payment commit" hash={d.txPre} />
                  <TxLine label="Post-outcome commit" hash={d.txOutcome} />
                </div>
              </Section>
            </div>
          </>
        )}
      </SheetContent>
    </Sheet>
  );
}

function KV({ label, value, mono }: { label: string; value: string; mono?: boolean }) {
  return (
    <div className="grid grid-cols-[140px_1fr] items-center gap-2 px-5 py-2">
      <dt className="mono-meta">{label}</dt>
      <dd className={mono ? "font-mono text-[13px] tabular-nums text-foreground" : "text-[13px] text-foreground"}>
        {value}
      </dd>
    </div>
  );
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="px-5 py-4">
      <h3 className="mono-meta mb-2">{title}</h3>
      {children}
    </div>
  );
}

function Bubble({ title, children, tone }: { title: string; children: React.ReactNode; tone?: "verified" | "failed" | "pending" }) {
  const ring =
    tone === "verified"
      ? "ring-[#10B981]/25"
      : tone === "failed"
      ? "ring-[#EF4444]/25"
      : "ring-border";
  return (
    <div className={`rounded-md bg-surface-2 p-2.5 ring-1 ring-inset ${ring}`}>
      <div className="mono-meta">{title}</div>
      <div className="mt-0.5 text-[13px] text-foreground/90">{children}</div>
    </div>
  );
}

function TxLine({ label, hash }: { label: string; hash?: string }) {
  return (
    <div className="flex items-center justify-between gap-3 rounded-md bg-surface-2 px-3 py-2">
      <div className="min-w-0">
        <div className="mono-meta">{label}</div>
        <div className="mt-0.5"><TxHash hash={hash} /></div>
      </div>
      {hash && (
        <a
          href={`https://allo.info/tx/${hash}`}
          target="_blank"
          rel="noreferrer"
          className="inline-flex items-center gap-1 text-[11px] text-muted-foreground hover:text-foreground"
          onClick={(e) => e.stopPropagation()}
        >
          Explorer <ExternalLink className="h-3 w-3" />
        </a>
      )}
    </div>
  );
}
