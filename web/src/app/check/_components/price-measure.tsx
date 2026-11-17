import { PnL } from "@/components/metron";
import { formatAmount, formatPrice } from "@/lib/format";
import { getMessages, type Locale } from "@/lib/messages";
import type { PreTradeCheck } from "@/lib/sample-data";

/** The domain is stop to target; integer ATR offsets form the intervening ticks. */
export function PriceMeasure({ check, locale = "en" }: { check: Omit<PreTradeCheck, "verdict">; locale?: Locale }) {
  const t = getMessages("check", locale);
  const stop = -Number(check.stopAtr), target = Number(check.targetAtr);
  const position = (atr: number) => (atr - stop) / (target - stop) * 100;
  const ticks = [stop, ...Array.from({ length: Math.ceil(target) - Math.ceil(stop) }, (_, i) => Math.ceil(stop) + i).filter(n => n > stop && n < target), target];
  return <>
    <div className="flex flex-wrap justify-between gap-4">
      {(["stop", "entry", "target"] as const).map(kind => <div key={kind} className={`flex min-w-0 flex-[1_1_190px] flex-col gap-1 ${kind === "target" ? "text-right" : ""}`}>
        <span className={`text-label uppercase ${kind === "stop" ? "text-loss" : kind === "target" ? "text-profit" : "text-subtle"}`}>{t(kind)}</span>
        <span className="num text-[32px] leading-[38px] font-medium tracking-[-0.02em]">{formatPrice(check[kind])}</span>
        <span className="text-small text-muted">{kind === "entry" ? t(check.orderType) : <><PnL locale={locale} value={kind === "stop" ? check.stopPercent : check.targetPercent} decimals={1} suffix="%" /> <span className="num">{t("atrDistance", { value: kind === "stop" ? check.stopAtr : check.targetAtr })}</span></>}</span>
      </div>)}
    </div>
    <div aria-hidden="true" className="flex flex-col gap-1.5">
      <div className="relative h-3.5">{ticks.map(tick => <span key={tick} className={`absolute bottom-0 w-px ${tick === stop || tick === 0 || tick === target ? "top-0 bg-muted" : "top-1.5 bg-subtle"}`} style={{ left: `calc(${position(tick)}% - ${tick === target ? 1 : 0}px)` }} />)}</div>
      <div className="relative flex h-7 overflow-hidden rounded">
        <span className="border-r-2 border-text bg-loss-tint" style={{ width: `${position(0)}%` }} /><span className="flex-1 bg-profit-tint" />
        <span className="absolute inset-y-0 left-0 w-1 bg-loss" /><span className="absolute inset-y-0 right-0 w-1 bg-profit" />
      </div>
      <div className="num relative h-4 text-[11px] text-subtle">{ticks.map(tick => <span key={tick} className={`absolute ${tick === stop ? "left-0" : tick === target ? "right-0" : "-translate-x-1/2"}`} style={tick !== stop && tick !== target ? { left: `${position(tick)}%` } : undefined}>{t(tick > 0 ? "positiveTick" : "tick", { value: formatAmount(String(tick), Number.isInteger(tick) ? 0 : 1) })}{tick === target ? t("atrUnit") : ""}</span>)}</div>
    </div>
  </>;
}
