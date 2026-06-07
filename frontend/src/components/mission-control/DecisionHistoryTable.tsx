import { useMissionStore } from "@/hooks/useMissionStore";
import { StatusPill } from "./StatusPill";
import { TxHash } from "./TxHash";

function fmt(ts: string) {
  const d = new Date(ts);
  const months = ["Jan","Feb","Mar","Apr","May","Jun","Jul","Aug","Sep","Oct","Nov","Dec"];
  const M = months[d.getUTCMonth()];
  const D = String(d.getUTCDate()).padStart(2, "0");
  const hh = String(d.getUTCHours()).padStart(2, "0");
  const mm = String(d.getUTCMinutes()).padStart(2, "0");
  return `${M} ${D} · ${hh}:${mm}`;
}

export function DecisionHistoryTable() {
  const history = useMissionStore((s) => s.history);
  const select = useMissionStore((s) => s.selectDecision);

  return (
    <section className="overflow-hidden rounded-[10px] border border-border bg-surface">
      <header className="flex h-10 items-center justify-between px-4 hairline-b">
        <div className="flex items-center gap-2">
          <h2 className="text-[12px] font-semibold tracking-tight">Decision History</h2>
          <span className="font-mono text-[11px] text-muted-foreground">— click row to inspect</span>
        </div>
        <span className="font-mono text-[11px] tabular-nums text-muted-foreground">
          {history.length} records
        </span>
      </header>
      <div className="overflow-x-auto">
        <table className="min-w-full text-[13px]">
          <thead>
            <tr className="text-left mono-meta hairline-b">
              <th className="px-4 py-2 font-normal">Time (UTC)</th>
              <th className="px-4 py-2 font-normal">ID</th>
              <th className="px-4 py-2 font-normal">Vendor</th>
              <th className="px-4 py-2 font-normal">Task</th>
              <th className="px-4 py-2 text-right font-normal">Cost</th>
              <th className="px-4 py-2 text-right font-normal">Conf.</th>
              <th className="px-4 py-2 font-normal">Policy</th>
              <th className="px-4 py-2 font-normal">Outcome</th>
              <th className="px-4 py-2 font-normal">Algorand TX</th>
            </tr>
          </thead>
          <tbody>
            {history.map((d) => (
              <tr
                key={d.id}
                onClick={() => select(d.id)}
                className="cursor-pointer border-b border-border/50 transition-colors last:border-0 hover:bg-surface-2"
              >
                <td className="px-4 py-2 font-mono text-[12px] tabular-nums text-muted-foreground">
                  {fmt(d.timestamp)}
                </td>
                <td className="px-4 py-2 font-mono text-[12px] text-foreground/80">{d.id}</td>
                <td className="px-4 py-2 font-medium text-foreground">{d.vendor}</td>
                <td className="px-4 py-2 max-w-[280px] truncate text-foreground/75">{d.task}</td>
                <td className="px-4 py-2 text-right font-mono tabular-nums">${d.cost.toFixed(2)}</td>
                <td className="px-4 py-2 text-right font-mono tabular-nums text-foreground/80">
                  {Math.round(d.confidence * 100)}%
                </td>
                <td className="px-4 py-2">
                  <StatusPill tone={d.policyStatus === "blocked" ? "blocked" : "approved"}>
                    {d.policyStatus}
                  </StatusPill>
                </td>
                <td className="px-4 py-2">
                  <StatusPill tone={d.outcomeStatus}>{d.outcomeStatus}</StatusPill>
                </td>
                <td className="px-4 py-2">
                  <TxHash hash={d.txPre} />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
}
