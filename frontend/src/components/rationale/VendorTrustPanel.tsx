import { useEffect, useState } from "react";
import { ArrowUp, ArrowDown } from "lucide-react";
import type { Vendor } from "@/lib/rationale/types";

function VendorRow({ v }: { v: Vendor }) {
  const [flash, setFlash] = useState(false);
  useEffect(() => {
    if (v.lastDelta) {
      setFlash(true);
      const t = setTimeout(() => setFlash(false), 1500);
      return () => clearTimeout(t);
    }
  }, [v.lastDelta?.at]);

  const scoreColor =
    v.score >= 4 ? "text-approved" : v.score >= 2.5 ? "text-foreground" : "text-blocked";

  return (
    <li className="flex items-center justify-between py-1.5 px-1">
      <span className="font-mono text-xs text-foreground/90">{v.name}</span>
      <div className="flex items-center gap-2">
        {v.lastDelta && flash && (
          <span
            className={`inline-flex items-center gap-0.5 text-[10px] font-mono tnum animate-fade-slide-in ${
              v.lastDelta.dir === "up" ? "text-approved" : "text-blocked"
            }`}
          >
            {v.lastDelta.dir === "up" ? (
              <ArrowUp className="size-3" />
            ) : (
              <ArrowDown className="size-3" />
            )}
            {v.lastDelta.value.toFixed(2)}
          </span>
        )}
        <div className="flex items-center gap-1">
          <span className={`font-mono tnum text-xs ${scoreColor}`}>{v.score.toFixed(1)}</span>
          <span className="text-[10px] text-muted-foreground font-mono">/5</span>
        </div>
      </div>
    </li>
  );
}

export function VendorTrustPanel({ vendors }: { vendors: Vendor[] }) {
  return (
    <section>
      <div className="flex items-center justify-between border-b border-border bg-surface px-4 h-9">
        <h2 className="text-[10px] font-mono uppercase tracking-[0.18em] text-muted-foreground">
          vendor trust
        </h2>
        <span className="text-[10px] font-mono tnum text-muted-foreground">
          pred vs actual
        </span>
      </div>
      <ul className="px-3 py-2 divide-y divide-border/60">
        {vendors.map((v) => (
          <VendorRow key={v.name} v={v} />
        ))}
      </ul>
    </section>
  );
}
