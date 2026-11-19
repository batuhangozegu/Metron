import { compareDecimals, decimalProduct, decimalSum, formatTradePrice } from "@/lib/format";
import { getMessages, type Locale } from "@/lib/messages";
import { meterWidth, type CheckResponse } from "@/lib/trading";

export function PriceMeasure({ check, locale = "en" }: { check: CheckResponse; locale?: Locale }) {
  const t = getMessages("check", locale);
  const short = check.side === "short";
  const low = short ? check.target_price : check.stop;
  const high = short ? check.stop : check.target_price;
  const distance = decimalSum(high, decimalProduct(low, "-1"));
  const entryOffset = decimalSum(check.entry_price, decimalProduct(low, "-1"));
  const width = meterWidth(entryOffset, distance);
  const valid = compareDecimals(distance, "0") > 0;
  return <>
    <div className="flex flex-wrap justify-between gap-4">
      {(["stop", "entry", "target"] as const).map(kind => <div key={kind} className={`flex min-w-0 flex-[1_1_190px] flex-col gap-1 ${kind === "target" ? "text-right" : ""}`}>
        <span className={`text-label uppercase ${kind === "stop" ? "text-loss" : kind === "target" ? "text-profit" : "text-subtle"}`}>{t(kind)}</span>
        <span className="num break-all text-[32px] leading-[38px] font-bold tracking-[-0.04em]">{formatTradePrice(kind === "stop" ? check.stop : kind === "entry" ? check.entry_price : check.target_price, locale)}</span>
      </div>)}
    </div>
    {valid && <div aria-hidden="true" className="measure"><div className="measure-ticks" /><div className="relative flex h-7 overflow-hidden rounded"><span className={`border-r-2 border-text ${short ? "bg-profit-tint" : "bg-loss-tint"}`} style={{ width: `${width}%` }} /><span className={`flex-1 ${short ? "bg-loss-tint" : "bg-profit-tint"}`} /><span className={`absolute inset-y-0 left-0 w-1 ${short ? "bg-profit" : "bg-loss"}`} /><span className={`absolute inset-y-0 right-0 w-1 ${short ? "bg-loss" : "bg-profit"}`} /></div></div>}
  </>;
}
