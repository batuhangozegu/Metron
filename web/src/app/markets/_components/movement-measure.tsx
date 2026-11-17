import { Measure } from "@/components/metron/measure";
import { getMessages } from "@/lib/messages";

/** ATR measures movement, not risk. Reuse shared geometry without its risk annotations. */
export function MovementMeasure({ market }: { market: { pair: string; atr: string; atrFill: string } }) {
  const t = getMessages("markets");
  const value = Number.parseFloat(market.atr);
  const limit = Number(t("movement.maximum"));
  return <div className="flex items-center gap-3">
    <div role="meter" aria-label={t("movement.label", { pair: market.pair })}
      aria-valuemin={0} aria-valuemax={limit} aria-valuenow={Math.min(limit, Math.max(0, value))}
      aria-valuetext={t("movement.value", { value: market.atr, maximum: limit })}
      className="w-[140px] shrink-0">
      <div aria-hidden="true">
        <Measure value={limit * Number.parseFloat(market.atrFill) / 100} limit={limit} size="sm" showMarks={false}
          className="[&_.measure]:text-muted [&_.measure-mark]:hidden" />
      </div>
    </div>
    <span className="num whitespace-nowrap">{market.atr}</span>
  </div>;
}
