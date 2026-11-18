"use client";

import { useState, type ReactNode } from "react";
import Link from "next/link";
import { ClassTag } from "@/components/metron";
import { formatPercent, formatPrice, reformatInput } from "@/lib/format";
import { useLocale } from "@/components/metron/locale-provider";
import { getMessages } from "@/lib/messages";
import { account, preTradeCaution, preTradeBlocked } from "@/lib/sample-data";

const control = "min-h-control w-full rounded-control border border-line bg-raised px-3 text-text";

export function CheckScreen({ caution, blocked }: { caution: ReactNode; blocked: ReactNode }) {
  const locale = useLocale();
  const t = getMessages("check", locale);
  const [pair, setPair] = useState(preTradeCaution.pair);
  const [orderType, setOrderType] = useState<"limit" | "market">("limit");
  const [price, setPrice] = useState({ value: formatPrice(preTradeCaution.entry, undefined, locale), locale });
  if (price.locale !== locale) setPrice({ value: reformatInput(price.value, price.locale, locale), locale });
  const check = pair === preTradeBlocked.pair ? preTradeBlocked : preTradeCaution;
  return <div className="flex min-w-0 flex-wrap items-start gap-6">
    <form aria-label={t("order")} onSubmit={event => event.preventDefault()} className="flex min-w-0 flex-[1_1_320px] flex-col gap-5 rounded-card border border-line bg-surface p-card">
      <h1 className="text-heading">{t("order")}</h1>
      <p className="text-small text-muted">{t("accountMarket")}: <span className="rounded-tag border border-line px-2 py-1">{getMessages("accounts", locale)(`market.${check.market}`)}</span></p>
      <div className="flex flex-col gap-2">
        <label htmlFor="pair" className="text-small text-muted">{t("pair")}</label>
        <select id="pair" className={`${control} text-[15px] font-medium`} value={pair} onChange={event => {
          const next = event.target.value === preTradeBlocked.pair ? preTradeBlocked : preTradeCaution;
          setPair(next.pair); setPrice({ value: formatPrice(next.entry, undefined, locale), locale });
        }}>
          <option>{preTradeCaution.pair}</option><option>{preTradeBlocked.pair}</option>
        </select>
        <div className="flex flex-wrap gap-2 text-xs text-muted">
          <ClassTag locale={locale} coinClass={check.coinClass} />
          <span className="rounded-tag border border-line px-2 py-[3px]">{t("atrPrefix")} <span className="num">{formatPercent(check.atr, 1, locale)}</span> {t("perDay")}</span>
          <span className="rounded-tag border border-line px-2 py-[3px]">{t("last")} <span className="num">{formatPrice(check.last, undefined, locale)}</span></span>
        </div>
      </div>
      <fieldset>
        <legend className="mb-2 text-small text-muted">{t("orderType")}</legend>
        <div className="grid grid-cols-2 gap-1 rounded-[10px] bg-raised p-1">
          {(["limit", "market"] as const).map(type => <button key={type} type="button" aria-pressed={orderType === type} onClick={() => setOrderType(type)} className={`min-h-control rounded-[7px] border text-sm font-medium ${orderType === type ? "border-line bg-surface text-text" : "border-transparent text-muted"}`}>{t(type)}</button>)}
        </div>
      </fieldset>
      <div className="flex flex-col gap-2">
        <label htmlFor="entry" className="text-small text-muted">{t("entryLabel", { currency: account.currency })}</label>
        <input id="entry" inputMode="decimal" className={`${control} num text-base`} value={price.value} onChange={event => setPrice({ value: event.target.value, locale })} />
      </div>
      <div className="flex flex-col gap-3 border-t border-line pt-5">
        <span className="text-label text-subtle uppercase">{t("setByRules")}</span>
        <dl className="grid grid-cols-[1fr_auto] gap-x-4 gap-y-2.5 text-sm">
          <dt className="text-muted">{t("riskPerTrade")}</dt><dd className="num text-right">{formatPercent(check.riskPercent, 2, locale)}</dd>
          <dt className="text-muted">{t("atrMultiplier")}</dt><dd className="num text-right">{t("multiplier", { value: formatPrice(check.atrMultiplier, undefined, locale) })}</dd>
          <dt className="text-muted">{t("targetRiskReward")}</dt><dd className="num text-right">{formatPrice(check.targetRiskReward, undefined, locale)}</dd>
        </dl>
        <p className="text-small text-subtle">{t(check.coinClass === "core" ? "rulesCore" : "rulesSmall")} <Link href="/rules" className="inline-flex min-h-control items-center text-muted underline">{t("changeRules")}</Link></p>
      </div>
    </form>
    <section aria-label={t("results")} className="flex min-w-0 flex-[999_1_600px] flex-col gap-4">{pair === preTradeBlocked.pair ? blocked : caution}</section>
  </div>;
}
