import { CircleCheck, OctagonMinus, TriangleAlert } from "lucide-react";
import { Card, CardHeader, RiskBadge, Stat } from "@/components/metron";
import { decimalProduct, formatAmount, formatPercent, formatPrice, formatTradePrice } from "@/lib/format";
import { getMessages, type Locale, type MessageKey } from "@/lib/messages";
import { checkBlocked, meterWidth, type CheckResponse, type RuleCheck } from "@/lib/trading";
import { PriceMeasure } from "./price-measure";
import { CheckActions } from "./check-actions";

const ruleNames: Record<string, MessageKey<"check">> = {
  class_risk: "rule.classRisk", atr_stop: "rule.atrStop", quantity_step: "rule.quantityStep",
  target_price: "rule.targetPrice", position_limit: "rule.positionLimit", spot_side: "rule.spotSide",
  initial_margin: "rule.initialMargin", liquidation_stop: "rule.liquidationStop",
  total_open_risk: "rule.totalOpenRisk", open_positions: "rule.openPositions",
  losing_trades_today: "rule.losingTradesToday", daily_loss: "rule.dailyLoss",
  stopped_trades_this_week: "rule.stoppedTradesThisWeek", weekly_loss: "rule.weeklyLoss",
};

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
    <Card data-risk={state} className="risk-panel flex flex-col gap-6 px-6 py-7">
      <div className="flex flex-wrap items-baseline justify-between gap-3"><h2 className="text-heading">{check.symbol} · {t(check.side)}</h2><span className="text-small text-subtle">{t("market")}</span></div>
      <PriceMeasure locale={locale} check={check} />
      <div className="flex flex-wrap gap-x-10 gap-y-5 border-t border-line pt-6 [&_.text-figure-medium]:text-xl">
        <Stat label={t("positionSize")} figure={formatPrice(check.position_size, undefined, locale)} />
        <Stat label={t("orderValue")} figure={t("usdt", { value: formatAmount(check.notional, 2, locale) })} />
        <Stat label={t("moneyAtRisk")} figure={t("usdt", { value: formatAmount(check.actual_risk, 2, locale) })} />
        <Stat label={t("riskPerTrade")} figure={formatPercent(check.risk_percent, 2, locale)} />
        <Stat label={t("atrMultiplier")} figure={t("multiplier", { value: formatPrice(check.atr_multiplier, undefined, locale) })} />
        {check.initial_margin != null && <Stat label={t("initialMargin")} figure={t("usdt", { value: formatAmount(check.initial_margin, 2, locale) })} />}
        {check.suggested_stop != null && <Stat label={t("suggestedStop")} figure={formatTradePrice(check.suggested_stop, locale)} />}
        {check.liquidation_price != null && <Stat label={t("liquidationPrice")} figure={formatTradePrice(check.liquidation_price, locale)} />}
      </div>
    </Card>
    <Card>
      <CardHeader title={t("limitsTitle")} action={<span className="text-small text-subtle">{t("limitsNote")}</span>} />
      <ul>{check.checks.map(row => {
        const state = row.status === "fail" ? "limit" : row.status === "caution" ? "caution" : "safe";
        const colors = { safe: "text-safe", caution: "text-caution", limit: "text-limit" };
        const title = ruleNames[row.rule_id] ? t(ruleNames[row.rule_id]) : t("rule.other");
        const metric = ruleMeter(row);
        const usable = metric && metric.before != null && metric.after != null && metric.limit != null;
        const width = usable ? meterWidth(metric.before, metric.limit) : 0;
        const added = usable ? Math.max(0, meterWidth(metric.after, metric.limit) - width) : 0;
        const display = (value: string) => metric?.percent ? formatPercent(value, 2, locale) : formatPrice(value, undefined, locale);
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
