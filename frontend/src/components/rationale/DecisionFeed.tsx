import { DecisionCard } from "./DecisionCard";
import type { Decision } from "@/lib/rationale/types";

interface Props {
  decisions: Decision[];
  selectedId: string | null;
  onSelect: (id: string) => void;
}

export function DecisionFeed({ decisions, selectedId, onSelect }: Props) {
  const sorted = [...decisions].sort((a, b) => b.timestamp - a.timestamp);
  return (
    <section className="flex-1 overflow-y-auto">
      <div className="sticky top-0 z-10 flex items-center justify-between border-b border-border bg-background px-5 h-9">
        <h2 className="text-[10px] font-mono uppercase tracking-[0.18em] text-muted-foreground">
          decision feed
        </h2>
        <span className="text-[10px] font-mono tnum text-muted-foreground">
          {decisions.length} records · newest first
        </span>
      </div>
      <div className="flex flex-col gap-2 p-4">
        {sorted.map((d) => (
          <DecisionCard
            key={d.id}
            decision={d}
            selected={selectedId === d.id}
            onClick={() => onSelect(d.id)}
          />
        ))}
      </div>
    </section>
  );
}
