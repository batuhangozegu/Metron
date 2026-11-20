import { CircleCheck, OctagonMinus, TriangleAlert } from "lucide-react";
import { Card, CardHeader, RiskBadge, Stat } from "@/components/metron";
import { compareDecimals, decimalProduct, formatAmount, formatPercent, formatPrice, formatTradePrice } from "@/lib/format";
import { getMessages, type Locale } from "@/lib/messages";
import { riskRuleName } from "@/lib/risk";
import { checkBlocked, meterWidth, type CheckResponse, type RuleCheck } from "@/lib/trading";
import { PriceMeasure } from "./price-measure";
import { CheckActions } from "./check-actions";

function ruleMeter(row: RuleCheck) {
  const v = row.values;
  if (v.limit_percent) return { before: v.current_percent, after: v.proposed_percent, limit: v.limit_percent, percent: true };
  if (v.limit) return { before: v.current, after: v.proposed, limit: v.limit, percent: false };
  if (row.rule_id === "position_limit" && v.equity && v.max_percent) return {
    before: v.current_class_holding, after: v.total_class_notional,
    limit: decimalProduct(decimalProduct(v.equity, v.max_percent), "0.01"), percent: false,
  };
  if (row.rule_id === "class_risk") return { before: "0", after: v.effective_percent, limit: v.max_percent, percent: true };
  return null;
}

function ruleValue(value: string, row: RuleCheck, locale: Locale, percent = false) {
  if (percent) return formatPercent(value, 2, locale);
  if (row.rule_id === "position_limit" || row.rule_id === "initial_margin") return getMessages("check", locale)("usdt", { value: formatAmount(value, 2, locale) });
  return formatTradePrice(value, locale);
}

function failedValues(row: RuleCheck, locale: Locale): string | null {
  const metric = ruleMeter(row);
  const v = row.values;
  // Non-portfolio checks expose their own comparison keys in the values map.
  const pairs: Record<string, [string | undefined, string | undefined]> = {
    initial_margin: [v.initial_margin, v.equity],
    liquidation_stop: [v.stop, v.liquidation_price],
    quantity_step: [v.size, v.step],
    target_price: [v.price, undefined],
    atr_stop: [v.distance, v.suggested_distance],
  };
  const [current, limit] = metric ? [metric.before, metric.limit] : pairs[row.rule_id] ?? [v.current, v.limit];
  if (current == null && limit == null) return null;
  const t = getMessages("check", locale);
  return t("blockedValues", {
    current: current == null ? t("notApplicable") : ruleValue(current, row, locale, metric?.percent),
    limit: limit == null ? t("notApplicable") : ruleValue(limit, row, locale, metric?.percent),
  });
}

export function CheckResult({ check, locale = "en", busy = false, onOpen }: { check: CheckResponse; locale?: Locale; busy?: boolean; onOpen: () => void }) {
  const t = getMessages("check", locale);
  const blocked = checkBlocked(check);
  const state = blocked ? "limit" : check.checks.some(rule => rule.status === "caution") ? "caution" : "safe";
  const Icon = { safe: CircleCheck, caution: TriangleAlert, limit: OctagonMinus }[state];
  const theme = { safe: "border-safe bg-safe-tint text-safe", caution: "border-caution bg-caution-tint text-caution", limit: "border-limit bg-limit-tint text-limit" }[state];
  return <div className="check-result flex min-w-0 flex-col gap-4" data-verdict={state}>
    <div role={blocked ? "alert" : "status"} data-risk={state} className={`risk-banner risk-panel flex flex-wrap items-center gap-x-5 gap-y-4 rounded-card border px-6 py-5 ${theme}`}>
      <Icon className="size-9 shrink-0" aria-hidden="true" />
      {blocked && <span className="blocked-stamp">{t("blockedStamp")}</span>}
      <div className="flex min-w-0 flex-[1_1_320px] flex-col gap-1 text-text"><h2 className="text-[26px] leading-8 font-semibold">{t(`liveBanner.${state}`)}</h2><p className="text-body text-muted">{t(`liveDescription.${state}`)}</p></div>
      <span className="num text-small text-muted">{t("summary", {
        passed: formatAmount(String(check.checks.filter(rule => rule.status === "pass").length), 0, locale),
        cautions: formatAmount(String(check.checks.filter(rule => rule.status === "caution").length), 0, locale),
        blocked: formatAmount(String(check.checks.filter(rule => rule.status === "fail").length), 0, locale),
      })}</span>
    </div>
    {blocked && <Card className="risk-panel border-limit bg-limit-tint px-6 py-5" aria-labelledby="failed-rules-title">
      <h3 id="failed-rules-title" className="font-semibold text-limit">{t("failedRulesTitle")}</h3>
      <ul className="mt-3 flex flex-col gap-3">{check.checks.filter(row => row.status === "fail").map(row => <li key={row.rule_id} className="flex flex-col gap-1 text-small">
        <span className="font-medium">{riskRuleName(row.rule_id, locale)}</span>
        <span className="text-muted">{locale === "tr" ? row.message_tr : row.message_en}</span>
        {failedValues(row, locale) && <span className="num text-limit">{failedValues(row, locale)}</span>}
      </li>)}</ul>
    </Card>}
    <Card data-risk={state} className="risk-panel flex flex-col gap-6 px-6 py-7">
      <div className="flex flex-wrap items-baseline justify-between gap-3"><h2 className="text-heading">{check.symbol} · {t(check.side)}</h2><span className="rounded-tag border border-line px-2.5 py-1 text-small text-muted">{t(`profileSource.${check.profile_source}`)}</span><span className="text-small text-subtle">{t("market")}</span></div>
      <PriceMeasure locale={locale} check={check} />
      <div className="flex flex-wrap gap-x-10 gap-y-5 border-t border-line pt-6 [&_.text-figure-medium]:text-xl">
        <Stat label={t("positionSize")} figure={formatPrice(check.position_size, undefined, locale)} />
        <Stat label={t("orderValue")} figure={t("usdt", { value: formatAmount(check.notional, 2, locale) })} />
        <Stat label={t("moneyAtRisk")} figure={t("usdt", { value: formatAmount(check.actual_risk, 2, locale) })} />
        <Stat label={t("riskPerTrade")} figure={formatPercent(check.risk_percent, 2, locale)} />
        <Stat label={t("atrMultiplier")} figure={t("multiplier", { value: formatPrice(check.atr_multiplier, undefined, locale) })} />
        {check.initial_margin != null && <Stat label={t("initialMargin")} figure={t("usdt", { value: formatAmount(check.initial_margin, 2, locale) })} />}
        {check.suggested_stop != null && <Stat label={t("suggestedStop")} figure={formatTradePrice(check.suggested_stop, locale)} />}
        {check.market === "futures" && <Stat label={t("liquidationPrice")} figure={check.liquidation_price && compareDecimals(check.liquidation_price, "0") !== 0 ? formatTradePrice(check.liquidation_price, locale) : t("notApplicable")} />}
      </div>
    </Card>
    <Card>
      <CardHeader title={t("limitsTitle")} action={<span className="text-small text-subtle">{t("limitsNote")}</span>} />
      <ul>{check.checks.map(row => {
        const state = row.status === "fail" ? "limit" : row.status === "caution" ? "caution" : "safe";
        const colors = { safe: "text-safe", caution: "text-caution", limit: "text-limit" };
        const title = riskRuleName(row.rule_id, locale);
        const metric = ruleMeter(row);
        const usable = metric && metric.before != null && metric.after != null && metric.limit != null;
        const width = usable ? meterWidth(metric.before, metric.limit) : 0;
        const added = usable ? Math.max(0, meterWidth(metric.after, metric.limit) - width) : 0;
        const display = (value: string) => ruleValue(value, row, locale, metric?.percent);
        const valueText = usable ? t("ruleValues", { before: display(metric.before), after: display(metric.after), limit: display(metric.limit) }) : "";
        return <li key={row.rule_id} data-risk={state} className="risk-panel flex flex-wrap items-center gap-x-5 gap-y-2.5 border-t border-line px-6 py-3.5">
          <div className="flex min-w-0 flex-[1_1_220px] flex-col gap-0.5"><span className="text-sm font-medium">{title}</span><span className="text-xs text-subtle">{locale === "tr" ? row.message_tr : row.message_en}</span></div>
          {usable && <><div role="meter" aria-label={title} aria-valuemin={0} aria-valuemax={100} aria-valuenow={meterWidth(metric.after, metric.limit)} aria-valuetext={valueText} data-risk={state} className={`measure min-w-0 flex-[1_1_200px] ${colors[state]}`}><div className="measure-ticks" aria-hidden="true" /><div className="measure-track" aria-hidden="true"><span className="measure-solid" style={{ width: `${width}%` }} /><span className="measure-added" style={{ left: `${width}%`, width: `${added}%` }} /></div></div><span className="num min-w-0 flex-[0_1_180px] text-right text-sm">{valueText}</span></>}
          <div className="ml-auto flex flex-[0_0_110px] justify-end"><RiskBadge locale={locale} state={state} className="bg-transparent px-0" /></div>
        </li>;
      })}</ul>
    </Card>
    <CheckActions blocked={blocked} busy={busy} onOpen={onOpen} />
  </div>;
}
