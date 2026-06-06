import { createFileRoute } from "@tanstack/react-router";
import { useReducer, useRef, useState, useCallback, useEffect } from "react";
import { TopBar } from "@/components/rationale/TopBar";
import { DecisionFeed } from "@/components/rationale/DecisionFeed";
import { DecisionDrawer } from "@/components/rationale/DecisionDrawer";
import { PolicyPanel } from "@/components/rationale/PolicyPanel";
import { VendorTrustPanel } from "@/components/rationale/VendorTrustPanel";
import { reducer } from "@/lib/rationale/reducer";
import { initialState } from "@/lib/rationale/mock";
import { runLiveScenario } from "@/lib/rationale/demoScenario";
import { fetchState, isApiConfigured } from "@/lib/rationale/api";

export const Route = createFileRoute("/")({
  head: () => ({
    meta: [
      { title: "Rationale — Agent Decision Audit" },
      {
        name: "description",
        content:
          "Tamper-evident reasoning for autonomous agent spend on Algorand via x402.",
      },
    ],
  }),
  component: RationaleApp,
});

function RationaleApp() {
  const [state, dispatch] = useReducer(reducer, initialState);
  const [running, setRunning] = useState(false);
  const [apiStatus, setApiStatus] = useState<"loading" | "live" | "mock">(
    isApiConfigured() ? "loading" : "mock"
  );
  const timersRef = useRef<number[]>([]);

  const cleanupRef = useRef<(() => void) | null>(null);

  const clearTimers = useCallback(() => {
    timersRef.current.forEach((t) => window.clearTimeout(t));
    timersRef.current = [];
    if (cleanupRef.current) {
      cleanupRef.current();
      cleanupRef.current = null;
    }
  }, []);

  useEffect(() => clearTimers, [clearTimers]);

  useEffect(() => {
    if (!isApiConfigured()) return;
    let cancelled = false;
    fetchState()
      .then((remote) => {
        if (cancelled) return;
        dispatch({ type: "HYDRATE", state: remote });
        setApiStatus("live");
      })
      .catch(() => {
        if (cancelled) return;
        setApiStatus("mock");
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const onRun = () => {
    if (running) return;
    setRunning(true);
    cleanupRef.current = runLiveScenario(dispatch);
  };

  const onReset = () => {
    clearTimers();
    setRunning(false);
    dispatch({ type: "RESET" });
  };

  const selected = state.decisions.find((d) => d.id === state.selectedId) ?? null;

  return (
    <div className="flex flex-col h-screen bg-background text-foreground">
      <TopBar
        agent={state.agent}
        balance={state.balance}
        spent={state.spent}
        limit={state.dailyLimit}
        running={running}
        apiStatus={apiStatus}
        onRun={onRun}
        onReset={onReset}
      />

      <main className="flex-1 grid grid-cols-[1fr_360px] overflow-hidden">
        <div className="flex flex-col overflow-hidden border-r border-border">
          <DecisionFeed
            decisions={state.decisions}
            selectedId={state.selectedId}
            onSelect={(id) => dispatch({ type: "SELECT", id })}
          />
        </div>

        <aside className="flex flex-col overflow-y-auto bg-background">
          <PolicyPanel
            dailyLimit={state.dailyLimit}
            spent={state.spent}
            allowed={state.allowedVendors}
            blocked={state.blockedVendors}
            alerts={state.alerts}
          />
          <VendorTrustPanel vendors={state.vendors} />
          <div className="px-4 py-3 mt-auto border-t border-border">
            <p className="text-[10px] font-mono uppercase tracking-wider text-muted-foreground leading-relaxed">
              every decision's reasoning is hashed and committed to algorand{" "}
              <span className="text-approved">before</span> the outcome is known.
              tamper-evident by construction.
            </p>
          </div>
        </aside>
      </main>

      <DecisionDrawer
        decision={selected}
        onClose={() => dispatch({ type: "SELECT", id: null })}
      />
    </div>
  );
}
