"use client";

import { Card, RiskBadge } from "@/components/metron";
import { Button } from "@/components/ui/button";
import { formatAmount } from "@/lib/format";
import { getMessages, type Locale } from "@/lib/messages";
import { countState, disciplineState, type RiskStatus } from "@/lib/risk-status";
import { tradingError } from "@/lib/trading";
import { useRiskStatus } from "@/lib/use-risk-status";

export function DisciplineCounts({ status, locale }: { status: RiskStatus; locale: Locale }) {
  const t = getMessages("overview", locale);
  return <dl className="flex flex-col gap-4">{(["losing_trades_today", "stopped_trades_this_week"] as const).map(key => <div key={key} className="flex flex-wrap items-center justify-between gap-3">
    <div className="flex flex-col gap-1"><dt className="text-small text-muted">{t(`discipline.${key}`)}</dt><dd className="num text-figure-medium">{formatAmount(String(status[key].count), 0, locale)} <span className="text-small text-subtle">{t("of", { limit: formatAmount(String(status[key].limit), 0, locale) })}</span></dd></div>
    <RiskBadge state={countState(status[key])} locale={locale} />
  </div>)}</dl>;
}

export function DisciplineMode({ accountId, locale }: { accountId: string; locale: Locale }) {
  const { data, error, reload } = useRiskStatus(accountId);
  const t = getMessages("overview", locale);
  return <Card data-risk={data ? disciplineState(data) : undefined} aria-labelledby="rules-discipline-title" className="risk-panel flex flex-col gap-4 p-card">
    <h2 id="rules-discipline-title" className="text-heading">{t("discipline")}</h2>
    {error !== null && <div className="flex flex-wrap items-center gap-3"><p role="alert" className="text-limit">{tradingError(error, locale)} {data && t("stale")}</p><Button variant="outline" onClick={() => void reload()}>{t("retry")}</Button></div>}
    {data ? <DisciplineCounts status={data} locale={locale} /> : error === null && <p role="status" className="text-muted">{t("loading")}</p>}
    <p className="text-small text-subtle">{t("discipline.note")}</p>
  </Card>;
}
