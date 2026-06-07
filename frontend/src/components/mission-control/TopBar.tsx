import { Play, AlertTriangle, RotateCcw } from "lucide-react";
import { useMissionStore } from "@/hooks/useMissionStore";
import { cn } from "@/lib/utils";

const TABS = ["Mission Control", "Decision History", "Audit Trail"] as const;

export function TopBar() {
  const { runScenario, reset, running } = useMissionStore();
  return (
    <header className="sticky top-0 z-40 hairline-b bg-background/85 backdrop-blur-md">
      <div className="mx-auto flex h-12 max-w-[1600px] items-center gap-6 px-4 lg:px-6">
        {/* Brand */}
        <div className="flex items-center gap-2">
          <div className="grid h-5 w-5 place-items-center rounded-[5px] bg-foreground text-background">
            <span className="text-[10px] font-bold leading-none">R</span>
          </div>
          <span className="text-[13px] font-semibold tracking-tight">RationAlgo</span>
        </div>

        {/* Tabs */}
        <nav className="flex items-center gap-0.5">
          {TABS.map((t, i) => (
            <button
              key={t}
              className={cn(
                "relative h-12 px-3 text-[13px] transition-colors",
                i === 0 ? "text-foreground" : "text-muted-foreground hover:text-foreground",
              )}
            >
              {t}
              {i === 0 && (
                <span className="absolute inset-x-3 bottom-0 h-px bg-foreground" />
              )}
            </button>
          ))}
        </nav>

        <div className="ml-auto flex items-center gap-3">
          {/* Status */}
          <div className="flex items-center gap-1.5 text-[12px] text-muted-foreground">
            <span className="relative inline-flex h-1.5 w-1.5">
              <span className="absolute inline-flex h-full w-full rounded-full bg-[#10B981] opacity-75 pulse-dot-glow text-[#10B981]" />
              <span className="relative inline-flex h-1.5 w-1.5 rounded-full bg-[#10B981]" />
            </span>
            <span>Online</span>
          </div>

          <div className="h-4 w-px bg-border" />

          <button
            disabled={running}
            onClick={() => runScenario("normal")}
            className="inline-flex h-7 items-center gap-1.5 rounded-md border border-border bg-surface px-2.5 text-[12px] font-medium text-foreground transition hover:bg-surface-2 disabled:opacity-50"
          >
            <Play className="h-3 w-3" />
            Run Scenario
          </button>
          <button
            disabled={running}
            onClick={() => runScenario("anomaly")}
            className="inline-flex h-7 items-center gap-1.5 rounded-md border border-border bg-surface px-2.5 text-[12px] font-medium text-[#F59E0B] transition hover:bg-surface-2 disabled:opacity-50"
          >
            <AlertTriangle className="h-3 w-3" />
            Anomaly
          </button>
          <button
            onClick={reset}
            className="inline-flex h-7 items-center gap-1.5 rounded-md px-2 text-[12px] text-muted-foreground transition hover:text-foreground"
          >
            <RotateCcw className="h-3 w-3" />
            Reset
          </button>
        </div>
      </div>
    </header>
  );
}
