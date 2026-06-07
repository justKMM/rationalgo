import { createFileRoute } from "@tanstack/react-router";
import { TopBar } from "@/components/mission-control/TopBar";
import { BudgetTierBar } from "@/components/mission-control/BudgetTierBar";
import { TrustPipeline } from "@/components/mission-control/TrustPipeline";
import { VendorPipeline } from "@/components/mission-control/VendorPipeline";
import { OpsCorner } from "@/components/mission-control/TrustMetrics";
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

      <main className="mx-auto w-full max-w-[1280px] px-4 py-4 lg:px-6 lg:py-5">
        <section className="overflow-hidden rounded-[10px] border border-border bg-surface">
          <BudgetTierBar />

          <div className="panel-pipeline-h relative grid min-h-0 grid-cols-1 pb-14 lg:grid-cols-2">
            <OpsCorner />
            <div className="flex min-h-0 flex-col hairline-b lg:hairline-r lg:border-b-0">
              <TrustPipeline />
            </div>
            <div className="flex min-h-0 flex-col">
              <VendorPipeline />
            </div>
          </div>
        </section>

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
