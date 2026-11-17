import { formatAmount, formatPercent, type Decimal } from "@/lib/format";
import { getMessages, type Locale } from "@/lib/messages";
import { cn } from "@/lib/utils";

export interface PnLProps { locale?: Locale; value: Decimal; decimals?: number; suffix?: string; className?: string }

/** Signed trade result with both direction and sign; zero is neutral and has no arrow. */
export function PnL({ value, decimals = 2, suffix = "", className, locale = "en" }: PnLProps) {
  const formatted = suffix === "%" ? formatPercent(value, decimals, locale) : formatAmount(value, decimals, locale);
  const negative = formatted.startsWith("−");
  const amount = negative ? formatted.slice(1) : formatted;
  const state = !/[1-9]/.test(amount) ? "zero" : negative ? "negative" : "positive";
  return <span key={formatted} data-pnl={state} className={cn("pnl num whitespace-nowrap", { positive: "text-profit", negative: "text-loss", zero: "text-muted" }[state], className)}>
    {getMessages("common", locale)(`pnl.${state}`, { amount, suffix: suffix === "%" ? "" : suffix })}
  </span>;
}
