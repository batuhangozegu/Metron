"use client";

import { useCallback, useTransition, type ReactNode } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { Card, CardHeader, ClassTag, PnL, Stat } from "@/components/metron";
import { Button } from "@/components/ui/button";
import { useLocale } from "@/components/metron/locale-provider";
import { compareDecimals, formatAmount, formatPercent, formatPrice, formatTradePrice } from "@/lib/format";
import { getMessages, type Locale, type MessageKey } from "@/lib/messages";
import type { Portfolio, PortfolioPeriod } from "@/lib/portfolio";
import { riskRuleName } from "@/lib/risk";
import { tradingError } from "@/lib/trading";
import { accountLabel } from "@/lib/accounts";
import { AccountBadges } from "@/components/metron/account-badges";
import { useTradingAccounts } from "@/lib/use-trading-accounts";
import { useJournal, usePortfolio } from "@/lib/use-portfolio";
import { EquityChart } from "./equity-chart";
import { NumberText } from "./number-text";
import { PeriodSelector } from "./period-selector";
import { Positions } from "./positions";

const table = "w-full border-collapse whitespace-nowrap text-right text-sm [&_th]:border-t [&_th]:border-line [&_th]:px-3 [&_th]:py-2.5 [&_th]:font-medium [&_td]:border-t [&_td]:border-line [&_td]:px-3 [&_td]:py-3.5 [&_th:first-child]:px-6 [&_td:first-child]:px-6 [&_th:last-child]:pr-6 [&_td:last-child]:pr-6";

export function PortfolioDashboard({ initialAccountId, period }: { initialAccountId?: string; period: PortfolioPeriod }) {
  const locale = useLocale();
  const t = getMessages("portfolio", locale);
  const router = useRouter();
  const params = useSearchParams();
  const [selecting, startTransition] = useTransition();
  const { accounts, account, error, reload } = useTradingAccounts(initialAccountId);
  const navigate = (accountId: string | undefined, nextPeriod: PortfolioPeriod) => {
    const query = new URLSearchParams(params.toString());
    if (accountId) query.set("account", accountId);
    query.set("period", nextPeriod);
    startTransition(() => router.replace(`/portfolio?${query}`, { scroll: false }));
  };
  return <div className="flex min-w-0 flex-col gap-6">
    <div className="flex flex-wrap items-center justify-between gap-4">
      {account && <div className="flex min-w-0 flex-wrap items-center gap-3"><label htmlFor="portfolio-account" className="text-small text-muted">{t("account")}</label><select id="portfolio-account" className="min-h-control max-w-full rounded-control border border-line bg-raised px-3 text-text" value={account.id} disabled={selecting} onChange={event => navigate(event.target.value, period)}>{accounts!.map(item => <option key={item.id} value={item.id}>{accountLabel(item, locale)}</option>)}</select><AccountBadges account={account} locale={locale} /></div>}
      <PeriodSelector period={period} disabled={selecting} onChange={value => navigate(account?.id, value)} />
    </div>
    {error !== null ? <BlockState error={error} retry={() => void reload()} locale={locale} />
      : accounts === null ? <p role="status" className="text-muted">{t("loadingAccounts")}</p>
      : !account ? <p className="text-muted">{t("noAccounts")} <Link href="/accounts" className="inline-flex min-h-control items-center underline">{t("createAccount")}</Link></p>
      : <AccountPortfolio key={account.id} accountId={account.id} period={period} locale={locale} />}
  </div>;
}

function BlockState({ error, retry, locale, empty }: { error?: unknown; retry?: () => void; locale: Locale; empty?: MessageKey<"portfolio"> }) {
  const t = getMessages("portfolio", locale);
  return <div className="flex flex-wrap items-center gap-3 p-6">
    {error != null ? <><p role="alert" className="text-limit">{tradingError(error, locale)}</p><Button variant="outline" onClick={retry}>{t("retry")}</Button></>
      : <p role={empty ? undefined : "status"} className="text-muted">{t(empty ?? "loadingPortfolio")}</p>}
  </div>;
}

function AccountPortfolio({ accountId, period, locale }: { accountId: string; period: PortfolioPeriod; locale: Locale }) {
  const t = getMessages("portfolio", locale);
  const portfolio = usePortfolio(accountId, period);
  const journal = useJournal(accountId);
  const reloadPortfolio = portfolio.reload, reloadJournal = journal.reload;
  const onClosed = useCallback(() => { reloadPortfolio(); reloadJournal(); }, [reloadPortfolio, reloadJournal]);
  return <>
    {portfolio.data ? <>
      <Summary data={portfolio.data} locale={locale} />
      <EquityChart data={portfolio.data} locale={locale} />
      <div className="grid min-w-0 items-start gap-4 xl:grid-cols-2"><ClassResults data={portfolio.data} locale={locale} /><Holdings data={portfolio.data} locale={locale} /></div>
    </> : <>
      {(["summary", "chartTitle", "classTitle", "holdingsTitle"] as const).map(title => <Card key={title}><CardHeader title={t(title)} /><BlockState error={portfolio.error} retry={portfolio.reload} locale={locale} /></Card>)}
    </>}
    <section aria-label={t("positions")} className="flex min-w-0 flex-col gap-6"><Positions accountId={accountId} onClosed={onClosed} /></section>
    <Card>
      <CardHeader title={t("journalTitle")} note={<NumberText text={t("journalNote")} />} />
      {journal.data === null ? <BlockState error={journal.error} retry={journal.reload} locale={locale} />
        : journal.data.trades.length === 0 ? <BlockState empty="emptyJournal" locale={locale} />
        : <ScrollTable title={t("journalTitle")}><table className={`${table} min-w-[850px]`}>
          <thead className="text-xs text-subtle"><tr>{(["pair", "closedUtc", "howClosed", "result", "pnlUsdt", "rules"] as const).map((key, i) => <th key={key} scope="col" className={i < 3 || i === 5 ? "text-left" : ""}>{t(key)}</th>)}</tr></thead>
          <tbody>{journal.data.trades.map(trade => <tr key={trade.position_id}>
            <th scope="row" className="text-left">{trade.symbol}</th>
            <td className="num text-left text-muted"><time dateTime={trade.closed_at}>{new Intl.DateTimeFormat(locale === "tr" ? "tr-TR" : "en-GB", { dateStyle: "short", timeStyle: "medium", timeZone: "UTC" }).format(new Date(trade.closed_at))}</time></td>
            <td className="text-left text-muted">{t(`reason.${trade.close_reason}`)}</td>
            <td>{trade.r_multiple === null ? t("nullValue") : <PnL locale={locale} value={trade.r_multiple} suffix={t("rUnit")} />}</td>
            <td><PnL locale={locale} value={trade.net_pnl} /></td>
            <td className="text-left">{trade.rules.length === 0 ? <span className="text-muted">{t("rulesKept")}</span> : <ul className="flex flex-col gap-1">{trade.rules.map((rule, index) => <li key={`${rule.rule}:${index}`} className={rule.status === "fail" ? "text-limit" : "text-caution"}>{t(`ruleStatus.${rule.status}`, { rule: riskRuleName(rule.rule, locale) })}</li>)}</ul>}</td>
          </tr>)}</tbody>
        </table></ScrollTable>}
      {journal.moreError != null && <BlockState error={journal.moreError} retry={() => void journal.loadMore()} locale={locale} />}
      {journal.data?.next_before && <div className="border-t border-line p-4"><Button variant="outline" disabled={journal.loadingMore} onClick={() => void journal.loadMore()}>{t(journal.loadingMore ? "loadingMore" : "loadMore")}</Button></div>}
    </Card>
  </>;
}

function Summary({ data, locale }: { data: Portfolio; locale: Locale }) {
  const t = getMessages("portfolio", locale);
  const { summary, stats, max_drawdown: drawdown } = data;
  const amount = (value: string | null) => value === null ? t("nullValue") : formatAmount(value, 2, locale);
  return <Card aria-label={t("summary")} className="grid gap-y-4 py-6 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5 [&>div]:px-6">
    <Stat label={t("accountValue")} figure={<>{amount(summary.equity)} <span className="text-small text-subtle">{data.currency}</span></>} note={<><PnL locale={locale} value={summary.return_percent} decimals={1} suffix="%" />{t("separator")}<PnL locale={locale} value={summary.net_result} /></>} />
    <Stat label={t("maxDrawdown")} figure={drawdown ? formatPercent(drawdown.percent, 1, locale) : t("nullValue")} note={drawdown ? <NumberText text={t("drawdownRange", { peak: amount(drawdown.peak), trough: amount(drawdown.trough) })} /> : t("nullValue")} />
    <Stat label={t("winRate")} figure={stats.win_rate === null ? t("nullValue") : formatPercent(stats.win_rate, 1, locale)} note={<NumberText text={t("tradeCount", { wins: formatAmount(String(stats.wins), 0, locale), trades: formatAmount(String(stats.trades), 0, locale) })} />} />
    <Stat label={t("averageWinLoss")} figure={stats.win_loss_ratio === null ? t("nullValue") : t("multiple", { value: amount(stats.win_loss_ratio) })} note={<>{stats.average_win === null ? t("nullValue") : <PnL locale={locale} value={stats.average_win} />}{t("slash")}{stats.average_loss === null ? t("nullValue") : <PnL locale={locale} value={stats.average_loss} />}</>} />
    <Stat label={t("feesPaid")} figure={amount(summary.fees_paid)} note={t("feesIncluded")} />
  </Card>;
}

function ScrollTable({ title, children }: { title: string; children: ReactNode }) {
  return <div role="region" aria-label={title} tabIndex={0} className="overflow-x-auto">{children}</div>;
}

function ClassResults({ data, locale }: { data: Portfolio; locale: Locale }) {
  const t = getMessages("portfolio", locale);
  const common = getMessages("common", locale);
  const rows = data.stats.by_class;
  const losing = rows.filter(row => compareDecimals(row.net_pnl, "0") < 0).map(row => common(`class.${row.coin_class}`));
  return <Card><CardHeader title={t("classTitle")} />
    {rows.length === 0 ? <BlockState empty="emptyClasses" locale={locale} /> : <ScrollTable title={t("classTitle")}><table className={`${table} min-w-[440px]`}>
      <thead className="text-xs text-subtle"><tr><th scope="col" className="text-left">{t("class")}</th><th scope="col">{t("trades")}</th><th scope="col">{t("winRate")}</th><th scope="col">{t("netResult")}</th></tr></thead>
      <tbody>{rows.map(row => <tr key={row.coin_class}><th scope="row" className="text-left"><ClassTag coinClass={row.coin_class} locale={locale} /></th><td className="num">{formatAmount(String(row.trades), 0, locale)}</td><td className="num">{row.win_rate === null ? t("nullValue") : formatPercent(row.win_rate, 1, locale)}</td><td><PnL locale={locale} value={row.net_pnl} /></td></tr>)}</tbody>
    </table></ScrollTable>}
    {losing.length > 0 && <p className="border-t border-line px-6 py-3.5 text-small text-subtle">{t("losingClasses", { classes: new Intl.ListFormat(locale === "tr" ? "tr-TR" : "en-GB", { style: "long", type: "conjunction" }).format(losing) })}</p>}
  </Card>;
}

function Holdings({ data, locale }: { data: Portfolio; locale: Locale }) {
  const t = getMessages("portfolio", locale);
  return <Card><CardHeader title={t("holdingsTitle")} />
    {data.holdings.length === 0 ? <BlockState empty="emptyHoldings" locale={locale} /> : <ScrollTable title={t("holdingsTitle")}><table className={`${table} min-w-[560px]`}>
      <thead className="text-xs text-subtle"><tr>{(["asset", "side", "quantity", "averageCost", "fees"] as const).map((key, index) => <th key={key} scope="col" className={index === 0 ? "text-left" : ""}>{t(key)}</th>)}</tr></thead>
      <tbody>{data.holdings.map(row => <tr key={`${row.symbol}:${row.side}`}><th scope="row" className="text-left">{row.symbol}</th><td>{t(row.side)}</td><td className="num">{formatPrice(row.quantity, undefined, locale)}</td><td className="num">{formatTradePrice(row.average_cost, locale)}</td><td className="num text-muted">{formatAmount(row.fees_paid, 2, locale)}</td></tr>)}</tbody>
    </table></ScrollTable>}
    <p className="border-t border-line px-6 py-3.5 text-small text-subtle">{t("holdingsNote")}</p>
  </Card>;
}
