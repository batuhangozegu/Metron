import { getMessages } from "@/lib/messages";
import { cn } from "@/lib/utils";

export type CoinClass = "core" | "large" | "small" | "stable";
export interface ClassTagProps { coinClass: CoinClass; className?: string }

/** Neutral coin-class chip; class is informational and never uses risk colors. */
export function ClassTag({ coinClass, className }: ClassTagProps) {
  return <span className={cn("inline-flex whitespace-nowrap rounded-tag border border-line px-2 py-0.5 text-xs text-muted", className)}>{getMessages("common")(`class.${coinClass}`)}</span>;
}
