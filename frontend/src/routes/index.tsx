import { createFileRoute } from "@tanstack/react-router";
import { TopBar } from "@/components/mission-control/TopBar";
import { ReasoningFeed } from "@/components/mission-control/ReasoningFeed";
import { ActiveDecisionCard } from "@/components/mission-control/ActiveDecisionCard";
import { TrustPipeline } from "@/components/mission-control/TrustPipeline";
import { TrustMetrics } from "@/components/mission-control/TrustMetrics";
import { DecisionHistoryTable } from "@/components/mission-control/DecisionHistoryTable";
import { DecisionDetailsDrawer } from "@/components/mission-control/DecisionDetailsDrawer";

export const Route = createFileRoute("/")({
  head: () => ({
    meta: [
      { title: "RationAlgo — Mission Control" },
      {
        name: "description",
        content:
          "Mission control for AI agents spending via x402 — every decision reasoned, policy-checked, and anchored on Algorand.",
      },
    ],
  }),
  component: MissionControl,
});

function MissionControl() {
  return (
    <div className="min-h-screen bg-background text-foreground">
      <TopBar />

      <main className="mx-auto w-full max-w-[1600px] px-4 py-4 lg:px-6 lg:py-5">
        {/* Page header */}
        <div className="mb-4">
          <h1 className="text-[20px] font-semibold tracking-tight">Mission Control</h1>
          <p className="mt-0.5 text-[12px] text-muted-foreground">
            AI Agent · Policy · Algorand Provenance · x402 Payment · Outcome Verification
          </p>
        </div>

        {/* Metrics strip */}
        <TrustMetrics />

        {/* 3-column workspace */}
        <section className="mt-4 grid grid-cols-1 gap-0 overflow-hidden rounded-[10px] border border-border bg-surface lg:grid-cols-12 lg:[grid-auto-rows:minmax(0,1fr)]">
          <div className="panel-workspace-h lg:col-span-3 lg:hairline-r flex min-h-0 flex-col hairline-b lg:border-b-0">
            <ReasoningFeed />
          </div>
          <div className="panel-workspace-h lg:col-span-6 lg:hairline-r flex min-h-0 flex-col hairline-b lg:border-b-0">
            <ActiveDecisionCard />
          </div>
          <div className="panel-workspace-h lg:col-span-3 flex min-h-0 flex-col">
            <TrustPipeline />
          </div>
        </section>

        {/* History */}
        <div className="mt-4">
          <DecisionHistoryTable />
        </div>

        <footer className="mt-6 pb-4 text-center font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">
          Agent · Policy · Algorand · x402 · Verification
        </footer>
      </main>

      <DecisionDetailsDrawer />
    </div>
  );
}
