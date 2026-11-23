"use client";

import { useLocale } from "@/components/metron/locale-provider";
import { formatAmount } from "@/lib/format";
import { getMessages } from "@/lib/messages";
import type { PortfolioPeriod } from "@/lib/portfolio";
import { NumberText } from "./number-text";

export function PeriodSelector({ period, onChange, disabled }: { period: PortfolioPeriod; onChange: (period: PortfolioPeriod) => void; disabled?: boolean }) {
  const locale = useLocale();
  const t = getMessages("portfolio", locale);
  const options = [
    { id: "30d", label: t("days", { days: formatAmount("30", 0, locale) }) },
    { id: "90d", label: t("days", { days: formatAmount("90", 0, locale) }) },
    { id: "all", label: t("allTime") },
  ] as const;
  return <div className="flex max-w-full flex-col items-end gap-2">
    <div role="group" aria-label={t("period")} className="flex flex-wrap gap-1 rounded-[10px] bg-raised p-1">
      {options.map(option => <button key={option.id} type="button" disabled={disabled} aria-pressed={period === option.id} onClick={() => onChange(option.id)} className={`min-h-control rounded-[7px] px-4 text-sm font-medium disabled:opacity-60 ${period === option.id ? "bg-surface text-text ring-1 ring-line" : "text-muted hover:text-text"}`}><NumberText text={option.label} /></button>)}
    </div>
  </div>;
}
