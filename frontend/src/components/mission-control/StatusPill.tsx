import { cn } from "@/lib/utils";

type Tone = "approved" | "blocked" | "verified" | "failed" | "pending" | "online" | "neutral";

const TONES: Record<Tone, string> = {
  approved: "text-[#10B981] bg-[#10B981]/10 ring-[#10B981]/20",
  verified: "text-[#10B981] bg-[#10B981]/10 ring-[#10B981]/20",
  online: "text-[#10B981] bg-[#10B981]/10 ring-[#10B981]/20",
  blocked: "text-[#EF4444] bg-[#EF4444]/10 ring-[#EF4444]/25",
  failed: "text-[#EF4444] bg-[#EF4444]/10 ring-[#EF4444]/25",
  pending: "text-[#F59E0B] bg-[#F59E0B]/10 ring-[#F59E0B]/25",
  neutral: "text-muted-foreground bg-surface-2 ring-border",
};

export function StatusPill({
  tone,
  children,
  pulse = false,
  className,
}: {
  tone: Tone;
  children: React.ReactNode;
  pulse?: boolean;
  className?: string;
}) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 rounded-md px-1.5 py-0.5 text-[11px] font-medium tracking-tight ring-1 ring-inset",
        TONES[tone],
        className,
      )}
    >
      <span
        className={cn("h-1.5 w-1.5 rounded-full bg-current", pulse && "pulse-dot-glow")}
      />
      <span className="capitalize">{children}</span>
    </span>
  );
}
