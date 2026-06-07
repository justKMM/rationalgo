import { useState } from "react";
import { Copy, Check } from "lucide-react";
import { cn } from "@/lib/utils";

export function TxHash({ hash, className }: { hash?: string; className?: string }) {
  const [copied, setCopied] = useState(false);
  if (!hash) {
    return <span className={cn("font-mono text-[11px] text-muted-foreground", className)}>—</span>;
  }
  const short = `${hash.slice(0, 6)}…${hash.slice(-4)}`;
  return (
    <button
      onClick={(e) => {
        e.stopPropagation();
        navigator.clipboard.writeText(hash);
        setCopied(true);
        setTimeout(() => setCopied(false), 1200);
      }}
      className={cn(
        "group inline-flex items-center gap-1 font-mono text-[11px] text-foreground/75 transition hover:text-foreground",
        className,
      )}
      title={hash}
    >
      <span className="border-b border-dashed border-border group-hover:border-foreground/40">{short}</span>
      {copied ? (
        <Check className="h-3 w-3 text-[#10B981]" />
      ) : (
        <Copy className="h-3 w-3 opacity-0 transition group-hover:opacity-60" />
      )}
    </button>
  );
}
