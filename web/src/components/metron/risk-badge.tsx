import { CircleCheck, Octagon, TriangleAlert } from "lucide-react";
import { getMessages } from "@/lib/messages";
import { cn } from "@/lib/utils";
import type { RiskState } from "@/lib/risk";
export { riskState, type RiskState } from "@/lib/risk";

export interface RiskBadgeProps { state: RiskState; className?: string }
const icons = { safe: CircleCheck, caution: TriangleAlert, limit: Octagon };
const colors = { safe: "text-safe bg-safe-tint", caution: "text-caution bg-caution-tint", limit: "text-limit bg-limit-tint" };

/** Risk is always conveyed together by an icon, a translated word, and a semantic color. */
export function RiskBadge({ state, className }: RiskBadgeProps) {
  const Icon = icons[state];
  return <span className={cn("inline-flex items-center gap-1.5 rounded-tag px-2 py-1 text-xs font-medium", colors[state], className)}>
    <Icon aria-hidden="true" className="size-4 shrink-0" />{getMessages("common")(`risk.${state}`)}
  </span>;
}
