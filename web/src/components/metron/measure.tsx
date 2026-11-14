import { decimalProduct, decimalSum, formatAmount } from "@/lib/format";
import { getMessages } from "@/lib/messages";
import { riskState } from "@/lib/risk";
import { cn } from "@/lib/utils";
import { RiskBadge } from "./risk-badge";

export interface MeasureProps {
  value: number; limit: number; added?: number; cautionAt?: number;
  size?: "md" | "sm"; showMarks?: boolean; className?: string; label?: string;
  formatValue?: (value: string) => string;
}

/** Ten-division ruler: current exposure is solid, proposed exposure hatched; state uses their total. */
export function Measure({ value, limit, added = 0, cautionAt = 0.75, size = "md", showMarks = true, className, label, formatValue = formatAmount }: MeasureProps) {
  const t = getMessages("common");
  const safeValue = Number.isFinite(value) && value >= 0 ? value : 0;
  const safeAdded = Number.isFinite(added) && added >= 0 ? added : 0;
  const total = Number(decimalSum(String(safeValue), String(safeAdded)));
  const state = riskState(Number.isFinite(value) && Number.isFinite(added) && value >= 0 && added >= 0 ? total : NaN, limit, cautionAt);
  const width = limit > 0 ? Math.min(100, safeValue / limit * 100) : 0;
  const extraWidth = limit > 0 ? Math.min(100 - width, safeAdded / limit * 100) : 0;
  const colors = { safe: "text-safe", caution: "text-caution", limit: "text-limit" };
  const valueText = t(added ? "measure.added" : "measure.value", {
    value: formatValue(String(value)), added: formatValue(String(added)), limit: formatValue(String(limit)), state: t(`risk.${state}`),
  });
  return <div className={cn("flex min-w-0 flex-col gap-1.5", className)}>
    <div role="meter" aria-label={label ?? t("measure.label")} aria-valuemin={0} aria-valuemax={limit > 0 ? limit : 1}
      aria-valuenow={Math.min(Math.max(total, 0), limit > 0 ? limit : 1)} aria-valuetext={valueText}
      className={cn("measure", colors[state])} data-size={size}>
      <div className="measure-ticks" aria-hidden="true" />
      <div className="measure-track" aria-hidden="true">
        <span className="measure-solid" style={{ width: `${width}%` }} />
        <span className="measure-added" style={{ left: `${width}%`, width: `${extraWidth}%` }} />
        <span className="measure-mark bg-caution" style={{ left: `${cautionAt * 100}%` }} />
        <span className="measure-mark right-0 bg-limit" />
      </div>
    </div>
    {showMarks && <div className="num relative flex justify-between text-[11px] text-subtle" aria-hidden="true">
      <span>{formatValue("0")}</span>
      <span className="absolute -translate-x-1/2" style={{ left: `${cautionAt * 100}%` }} title={t("measure.caution")}>{formatValue(decimalProduct(String(limit), String(cautionAt)))}</span>
      <span title={t("measure.limit")}>{formatValue(String(limit))}</span>
    </div>}
    <span className="sr-only"><RiskBadge state={state} /></span>
  </div>;
}
