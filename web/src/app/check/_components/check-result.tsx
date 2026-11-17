import { CircleCheck, OctagonMinus, TriangleAlert } from "lucide-react";
import { Card, CardHeader, Measure, PnL, RiskBadge, Stat } from "@/components/metron";
import { formatAmount, formatPercent, formatPrice } from "@/lib/format";
import { getMessages, type Locale } from "@/lib/messages";
import { account, type PreTradeCheck, type LimitCheck } from "@/lib/sample-data";
import { PriceMeasure } from "./price-measure";
import { CheckActions } from "./check-actions";
import { NumberCopy } from "./number-copy";

export type CheckData = Omit<PreTradeCheck, "verdict"> & { verdict: "safe" | "caution" | "limit" };

function checkValues(row: LimitCheck, t: ReturnType<typeof getMessages<"check">>, locale: Locale) {
  const values = row.value.match(/\d+(?:\.\d+)?/g) ?? [];
  const [first = "0", second = "0", third] = values;
  const limit = Number(values.at(-1));
  const before = third ? Number(first) : row.warn ? Number(first) : 0;
  const after = third ? Number(second) : Number(first);
  const percent = row.value.includes("%");
  const isCooldown = !percent && !third;
  const display = (raw: string) => percent ? formatPercent(raw, raw.split(".")[1]?.length ?? 0, locale) : formatAmount(raw, 0, locale);
  const value = isCooldown ? t("waited", { value: display(first), limit: display(second) })
    : third ? t(percent ? "transitionPercent" : "transitionCount", { before: display(first), after: display(second), limit: display(third) })
    : t("singlePercent", { value: display(first), limit: display(second) });
  return { before: isCooldown ? after : before, added: isCooldown ? 0 : after - before, limit, value, percent, isCooldown };
}

export function CheckResult({ check, locale = "en" }: { check: CheckData; locale?: Locale }) {
  const t = getMessages("check", locale);
  const state = check.verdict;
  const Icon = { safe: CircleCheck, caution: TriangleAlert, limit: OctagonMinus }[state];
  const theme = { safe: "border-safe bg-safe-tint text-safe", caution: "border-caution bg-caution-tint text-caution", limit: "border-limit bg-limit-tint text-limit" }[state];
  const source = check.coinClass === "core" ? "blocked" : "caution";
  return <div className="check-result flex min-w-0 flex-col gap-4" data-verdict={state}>
    <div role={state === "limit" ? "alert" : "status"} data-risk={state} className={`risk-banner risk-panel flex flex-wrap items-center gap-x-5 gap-y-4 rounded-card border px-6 py-5 ${theme}`}>
      <Icon className="size-9 shrink-0" aria-hidden="true" />
      {state === "limit" && <span className="blocked-stamp">{t("blockedStamp")}</span>}
      <div className="flex min-w-0 flex-[1_1_320px] flex-col gap-1 text-text">
        <h2 className="text-[26px] leading-8 font-semibold tracking-[-0.01em]">{t(`banner.${state}`)}</h2>
        <p className="text-body text-muted"><NumberCopy text={t(`description.${state}`, { minutes: formatAmount(String(account.cooldown.minutesRemaining), 0, locale), position: formatPercent(check.proposedPositionPercent ?? "0", 1, locale), limit: formatPercent(check.positionLimitPercent ?? "0", 0, locale) })} /></p>
      </div>
      <span className="num text-small text-muted">{t("summary", { passed: formatAmount(String(check.passed), 0, locale), cautions: formatAmount(String(check.cautions), 0, locale), blocked: formatAmount(String(check.blocked), 0, locale) })}</span>
    </div>
    <Card data-risk={state} className="risk-panel flex flex-col gap-6 px-6 py-7">
      <div className="flex flex-wrap items-baseline justify-between gap-x-6 gap-y-2">
        <h2 className="text-heading">{t("measureTitle", { pair: check.pair })}</h2>
        <span className="text-small text-subtle"><NumberCopy text={t("tickExplanation")} /></span>
      </div>
      <PriceMeasure locale={locale} check={check} />
      <div className="flex flex-wrap justify-between gap-x-6 gap-y-3 text-sm text-muted">
        <span>{t("ifStop")} <PnL locale={locale} value={check.stopResult} suffix={t("currencySuffix", { currency: account.currency })} /> <span className="text-subtle"><NumberCopy text={t("accountRisk", { value: formatPercent(check.riskPercent, 2, locale) })} /></span></span>
        <span>{t("ifTarget")} <PnL locale={locale} value={check.targetResult} suffix={t("currencySuffix", { currency: account.currency })} /> <span className="num text-subtle">{t("reward", { value: formatPrice(check.targetRiskReward, undefined, locale) })}</span></span>
      </div>
      <div className="flex flex-wrap gap-x-10 gap-y-5 border-t border-line pt-6 [&_.text-figure-medium]:text-xl [&_.text-small]:text-xs">
        <Stat label={t("positionSize")} figure={<>{formatPrice(check.quantity, undefined, locale)} <span className="text-small text-subtle">{check.asset}</span></>} />
        <Stat label={t("orderValue")} figure={<>{formatAmount(check.orderValue, 2, locale)} <span className="text-small text-subtle">{account.currency}</span></>} />
        <Stat label={t("moneyAtRisk")} figure={<>{formatAmount(check.moneyAtRisk, 2, locale)} <span className="text-small text-subtle">{account.currency}</span></>} />
        <Stat label={t("sizeFormula")} className="[&_.text-figure-medium]:text-sm [&_.text-figure-medium]:text-muted" figure={t("formula", { risk: formatAmount(check.moneyAtRisk, 2, locale), entry: formatPrice(check.entry, undefined, locale), stop: formatPrice(check.stop, undefined, locale) })} />
      </div>
    </Card>
    <Card>
      <CardHeader title={t("limitsTitle")} action={<span className="text-small text-subtle">{t("limitsNote")}</span>} />
      <ul>
        {check.checks.map((row, i) => {
          const values = checkValues(row, t, locale);
          const key = `${source}.${i}` as const;
          return <li key={row.name} data-risk={row.over ? "limit" : row.warn ? "caution" : "safe"} className="risk-panel flex flex-wrap items-center gap-x-5 gap-y-2.5 border-t border-line px-6 py-3.5">
            <div className="flex min-w-0 flex-[1_1_220px] flex-col gap-0.5">
              <span className="text-sm font-medium">{t(`${key}.name` as Parameters<typeof t>[0])}</span>
              <span className="text-xs text-subtle"><NumberCopy text={t(`${key}.detail` as Parameters<typeof t>[0])} /></span>
            </div>
            {values.isCooldown ? <div role="meter" aria-label={t(`${key}.name` as Parameters<typeof t>[0])} aria-valuemin={0} aria-valuemax={values.limit} aria-valuenow={values.before} aria-valuetext={t(row.warn ? "cooldownStatus" : "cooldownComplete", { value: values.value })} className={`measure min-w-0 flex-[1_1_200px] ${row.warn ? "text-caution" : "text-safe"}`}>
              <div aria-hidden="true" className="measure-ticks" /><div aria-hidden="true" className="measure-track"><span className="measure-solid" style={{ width: `${Math.min(100, values.before / values.limit * 100)}%` }} /></div>
            </div> : <Measure locale={locale} value={values.before} added={values.added} limit={values.limit} formatValue={value => values.percent ? formatPercent(value, 2, locale) : formatAmount(value, 0, locale)} showMarks={false} className="min-w-0 flex-[1_1_200px]" label={t(`${key}.name` as Parameters<typeof t>[0])} />}
            <span className="num min-w-0 flex-[0_1_150px] text-right text-sm">{values.value}</span>
            <div className="flex flex-[0_0_110px] justify-end">{row.over || row.warn ? <span data-risk={row.over ? "limit" : "caution"} className={`risk-badge inline-flex items-center gap-1.5 text-[13px] font-medium ${row.over ? "text-limit" : "text-caution"}`}>{row.over ? <OctagonMinus aria-hidden="true" className="size-4" /> : <TriangleAlert aria-hidden="true" className="size-4" />}{t(row.over ? "exceeded" : values.isCooldown ? "broken" : "caution")}</span> : <RiskBadge locale={locale} state="safe" className="bg-transparent px-0" />}</div>
          </li>;
        })}
      </ul>
    </Card>
    <CheckActions state={state} stop={check.stop} reducedValue={check.reducedOrderValue} reducedRisk={check.reducedRisk} />
  </div>;
}
