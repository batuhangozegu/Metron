"use client";

import { Card, CardHeader, RiskBadge } from "@/components/metron";
import { Button } from "@/components/ui/button";
import { useLocale } from "@/components/metron/locale-provider";
import { getMessages } from "@/lib/messages";
import { riskRuleName, type RiskCheckHistoryEntry } from "@/lib/risk";
import { tradingError } from "@/lib/trading";

export function RiskCheckHistory({ checks, error, onRetry }: { checks: RiskCheckHistoryEntry[] | null; error: unknown; onRetry: () => void }) {
  const locale = useLocale();
  const t = getMessages("check", locale);
  const time = new Intl.DateTimeFormat(locale === "tr" ? "tr-TR" : "en-GB", { dateStyle: "short", timeStyle: "medium", timeZone: "UTC" });
  return <Card aria-labelledby="risk-history-title">
    <CardHeader title={<span id="risk-history-title">{t("history.title")}</span>} note={t("history.note")} />
    {error !== null ? <div className="flex flex-wrap items-center gap-3 border-t border-line px-6 py-4"><p role="alert" className="text-limit">{tradingError(error, locale)}</p><Button variant="outline" onClick={onRetry}>{t("retry")}</Button></div>
      : checks === null ? <p role="status" className="border-t border-line px-6 py-4 text-small text-muted">{t("history.loading")}</p>
      : checks.length === 0 ? <p className="border-t border-line px-6 py-4 text-small text-muted">{t("history.empty")}</p>
      : <ul>{checks.slice(0, 20).map(row => <li key={row.id} className="flex flex-wrap items-center gap-x-5 gap-y-2 border-t border-line px-6 py-3 text-small">
        <time dateTime={row.created_at} className="num text-muted">{t("history.time", { time: time.format(new Date(row.created_at)) })}</time>
        <span className="font-medium">{row.symbol} · {t(row.side)}</span>
        <span className="text-muted">{t(`history.source.${row.source}`)} · {t(`profileSource.${row.profile_source}`)}</span>
        <RiskBadge locale={locale} state={!row.allowed || row.verdict === "fail" ? "limit" : row.verdict === "caution" ? "caution" : "safe"} />
        <span className={`min-w-0 flex-[1_1_220px] ${row.failed_rules.length ? "text-limit" : row.caution_rules.length ? "text-caution" : "text-subtle"}`}>{
          row.failed_rules.length ? t("history.failed", { rules: row.failed_rules.map(rule => riskRuleName(rule, locale)).join(t("ruleSeparator")) })
            : row.caution_rules.length ? t("history.cautions", { rules: row.caution_rules.map(rule => riskRuleName(rule, locale)).join(t("ruleSeparator")) })
            : t("history.passed")}</span>
      </li>)}</ul>}
  </Card>;
}
