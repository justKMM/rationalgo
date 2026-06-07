import { useEffect } from "react";
import { Play, AlertTriangle, RotateCcw } from "lucide-react";
import { useMissionStore } from "@/hooks/useMissionStore";
import { isApiConfigured } from "@/lib/api";
import { cn } from "@/lib/utils";

export function TopBar() {
  const { runScenario, reset, running, apiLive, hydrate, error } = useMissionStore();
  const apiMode = isApiConfigured();
  const flowDisabled = running || (apiMode && !apiLive);
  const statusLive = apiMode ? apiLive : true;

  useEffect(() => {
    void hydrate();
  }, [hydrate]);

  return (
    <header className="sticky top-0 z-40 hairline-b bg-background/85 backdrop-blur-md">
      <div className="mx-auto flex h-12 max-w-[1600px] items-center gap-6 px-4 lg:px-6">
        <div className="flex items-center gap-2">
          <div className="grid h-5 w-5 place-items-center rounded-[5px] bg-foreground text-background">
            <span className="text-[10px] font-bold leading-none">R</span>
          </div>
          <span className="text-[13px] font-semibold tracking-tight">RationAlgo</span>
          <span className="text-[12px] text-muted-foreground">Mission Control</span>
        </div>

        <div className="ml-auto flex items-center gap-3">
          {/* Status */}
          <div className="flex items-center gap-1.5 text-[12px] text-muted-foreground">
            <span className="relative inline-flex h-1.5 w-1.5">
              <span
                className={cn(
                  "absolute inline-flex h-full w-full rounded-full opacity-75 pulse-dot-glow",
                  statusLive ? "bg-[#10B981] text-[#10B981]" : "bg-muted-foreground text-muted-foreground",
                )}
              />
              <span
                className={cn(
                  "relative inline-flex h-1.5 w-1.5 rounded-full",
                  statusLive ? "bg-[#10B981]" : "bg-muted-foreground",
                )}
              />
            </span>
            <span>{apiMode ? (apiLive ? "api live" : "offline") : "mock"}</span>
          </div>

          {error && (
            <span className="max-w-[200px] truncate text-[11px] text-[#EF4444]" title={error}>
              {error}
            </span>
          )}

          <div className="h-4 w-px bg-border" />

          <button
            disabled={flowDisabled}
            onClick={() => runScenario("normal")}
            className="inline-flex h-7 items-center gap-1.5 rounded-md border border-border bg-surface px-2.5 text-[12px] font-medium text-foreground transition hover:bg-surface-2 disabled:opacity-50"
          >
            <Play className="h-3 w-3" />
            Execute Flow
          </button>
          <button
            disabled={flowDisabled}
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
