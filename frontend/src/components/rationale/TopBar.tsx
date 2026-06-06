import { Play, RotateCcw } from "lucide-react";
import { SpendGauge } from "./SpendGauge";

interface Props {
  agent: string;
  balance: number;
  spent: number;
  limit: number;
  running: boolean;
  apiStatus?: "loading" | "live" | "mock";
  onRun: () => void;
  onReset: () => void;
}

export function TopBar({
  agent,
  balance,
  spent,
  limit,
  running,
  apiStatus = "mock",
  onRun,
  onReset,
}: Props) {
  return (
    <header className="flex items-center justify-between border-b border-border bg-surface px-5 h-12">
      <div className="flex items-center gap-5">
        <div className="flex items-center gap-2">
          <div className="size-2 bg-approved animate-pulse-dot" />
          <span className="font-mono text-xs font-semibold tracking-wide text-foreground">
            RATIONALE
          </span>
          <span className="text-[10px] uppercase tracking-[0.18em] text-muted-foreground">
            agent decision audit · algorand x402
          </span>
          {apiStatus === "live" && (
            <span className="text-[10px] font-mono uppercase tracking-wider text-approved">
              api live
            </span>
          )}
          {apiStatus === "loading" && (
            <span className="text-[10px] font-mono uppercase tracking-wider text-muted-foreground">
              api…
            </span>
          )}
        </div>
        <span className="text-border">|</span>
        <div className="flex items-center gap-2">
          <span className="text-[10px] uppercase tracking-wider text-muted-foreground">
            agent
          </span>
          <span className="font-mono text-xs text-foreground">{agent}</span>
        </div>
        <span className="text-border">|</span>
        <div className="flex items-center gap-2">
          <span className="text-[10px] uppercase tracking-wider text-muted-foreground">
            balance
          </span>
          <span className="font-mono tnum text-xs text-foreground">
            {balance.toFixed(2)}
            <span className="text-muted-foreground ml-1">EURQ</span>
          </span>
        </div>
      </div>

      <div className="flex items-center gap-5">
        <SpendGauge spent={spent} limit={limit} />
        <button
          onClick={onReset}
          className="inline-flex items-center gap-1.5 border border-border bg-transparent px-2 h-7 text-[11px] font-mono uppercase tracking-wider text-muted-foreground hover:text-foreground hover:border-muted-foreground transition-colors"
        >
          <RotateCcw className="size-3" />
          reset
        </button>
        <button
          onClick={onRun}
          disabled={running}
          className="inline-flex items-center gap-1.5 border border-approved bg-approved/15 px-3 h-7 text-[11px] font-mono uppercase tracking-wider text-approved hover:bg-approved/25 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          <Play className="size-3" />
          {running ? "running…" : "run demo scenario"}
        </button>
      </div>
    </header>
  );
}
