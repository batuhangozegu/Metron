import { Measure } from "@/components/metron/measure";
import { getMessages, type Locale } from "@/lib/messages";
import { formatAmount, formatPercent } from "@/lib/format";

/** ATR measures movement, not risk. Reuse shared geometry without its risk annotations. */
export function MovementMeasure({ market, locale = "en" }: { market: { pair: string; atr: string; atrFill: string }; locale?: Locale }) {
  const t = getMessages("markets", locale);
  const value = Number.parseFloat(market.atr);
  const atr = formatPercent(market.atr.replace("%", ""), market.atr.split(".")[1]?.replace("%", "").length ?? 0, locale);
  const limit = Number(t("movement.maximum"));
  return <div className="flex items-center gap-3">
    <div role="meter" aria-label={t("movement.label", { pair: market.pair })}
      aria-valuemin={0} aria-valuemax={limit} aria-valuenow={Math.min(limit, Math.max(0, value))}
      aria-valuetext={t("movement.value", { value: atr, maximum: formatAmount(String(limit), 0, locale) })}
      className="market-movement shrink-0">
      <div aria-hidden="true">
        <Measure locale={locale} value={limit * Number.parseFloat(market.atrFill) / 100} limit={limit} size="sm" showMarks={false}
          className="[&_.measure]:text-muted [&_.measure-mark]:hidden [&_.measure-track]:!animate-none" />
      </div>
    </div>
    <span className="num whitespace-nowrap">{atr}</span>
  </div>;
}
