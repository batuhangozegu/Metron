"use client";

import { useEffect, useRef, useState, useTransition } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { CircleCheck, Octagon, TriangleAlert } from "lucide-react";
import { Card, ClassTag, Measure, PnL, Stat } from "@/components/metron";
import { AccountBadges } from "@/components/metron/account-badges";
import { useAuth } from "@/components/metron/auth-provider";
import { useLocale } from "@/components/metron/locale-provider";
import { Button } from "@/components/ui/button";
import { accountLabel } from "@/lib/accounts";
import { ApiError, apiRequest, type TradingAccount } from "@/lib/auth";
import { compareDecimals, formatAmount, formatPercent, formatPrice, formatTradePrice } from "@/lib/format";
import { getMessages, type Locale } from "@/lib/messages";
import { disciplineState, distanceToStop, formatRiskTime, formatStopDistance, gaugeState, overviewState, remainingAmount, type RiskStatus } from "@/lib/risk-status";
import { tradingError, type Position } from "@/lib/trading";
import { useRiskStatus } from "@/lib/use-risk-status";
import { useTradingAccounts } from "@/lib/use-trading-accounts";
import { DisciplineCounts } from "../rules/_components/discipline-mode";
import { NumberText } from "./number-text";

const gauges = [{ key: "open_risk", label: "open" }, { key: "daily_loss", label: "daily" }, { key: "weekly_loss", label: "weekly" }] as const;
const columns = ["pair", "entry", "price", "stop", "toStop", "atRisk", "unrealized"] as const;

export function OverviewDashboard({ initialAccountId }: { initialAccountId?: string }) {
  const locale = useLocale();
  const t = getMessages("overview", locale);
  const { accounts, account, error, reload } = useTradingAccounts(initialAccountId);
  const router = useRouter();
  const params = useSearchParams();
  const [selecting, startTransition] = useTransition();
  const selectAccount = (id: string) => {
    const query = new URLSearchParams(params.toString());
    query.set("account", id);
    startTransition(() => router.replace(`/?${query}`, { scroll: false }));
  };
  return <>
    {account && <div className="flex min-w-0 flex-wrap items-center gap-3"><label htmlFor="overview-account" className="text-small text-muted">{t("account")}</label><select id="overview-account" value={account.id} disabled={selecting} onChange={event => selectAccount(event.target.value)} className="min-h-control max-w-full rounded-control border border-line bg-raised px-3 text-text">{accounts!.map(item => <option key={item.id} value={item.id}>{accountLabel(item, locale)}</option>)}</select><AccountBadges account={account} locale={locale} /></div>}
    {error !== null ? <LoadError error={error} reload={reload} locale={locale} />
      : accounts === null ? <p role="status" className="text-muted">{t("loadingAccounts")}</p>
      : !account ? <Card className="p-card"><h1 className="text-title">{t("title")}</h1><p className="mt-3 text-muted">{t("noAccounts")} <Link href="/accounts" className="inline-flex min-h-control items-center underline">{t("manageAccounts")}</Link></p></Card>
      : <AccountOverview key={account.id} account={account} locale={locale} />}
  </>;
}

function LoadError({ error, reload, locale, stale = false }: { error: unknown; reload: () => void; locale: Locale; stale?: boolean }) {
  const t = getMessages("overview", locale);
  return <div className="flex flex-wrap items-center gap-3"><p role="alert" className="text-limit">{tradingError(error, locale)} {stale && t("stale")}</p><Button variant="outline" onClick={() => void reload()}>{t("retry")}</Button></div>;
}

function AccountOverview({ account, locale }: { account: TradingAccount; locale: Locale }) {
  const { data, positions, error, reload } = useRiskStatus(account.id, true);
  const t = getMessages("overview", locale);
  return <>
    {error !== null && <LoadError error={error} reload={reload} locale={locale} stale={data !== null} />}
    {!data || !positions ? error === null && <p role="status" className="text-muted">{t("loading")}</p> : <>
      <Headline data={data} locale={locale} />
      <section aria-label={t("gauges")} className="flex flex-wrap gap-4">{gauges.map(({ key, label: name }) => {
        const gauge = data[key], label = t(`gauge.${name}.label`);
        return <Card key={key} data-risk={gaugeState(gauge)} className="risk-panel flex min-w-0 flex-[1_1_320px] flex-col gap-4 p-card">
          <h2 className="risk-label text-muted">{label}</h2>
          <div className="flex flex-wrap items-baseline gap-2.5"><span className="gauge-figure num text-figure-large">{formatPercent(gauge.percent, 2, locale)}</span><span className="text-body text-subtle"><NumberText text={t("of", { limit: formatPercent(gauge.limit_percent, 2, locale) })} /></span></div>
          <Measure locale={locale} value={Number(gauge.percent)} limit={Number(gauge.limit_percent)} label={label} formatValue={amount => amount === "0" ? formatAmount(amount, 0, locale) : formatPercent(amount, 2, locale)} />
          <p className="text-small leading-5 text-muted"><span className="num text-text">{formatAmount(remainingAmount(data.equity, gauge), 2, locale)} {data.currency}</span> {t(`gauge.${name}.note`)}</p>
        </Card>;
      })}</section>
      <section aria-label={t("details")} className="positions-layout min-w-0 items-start gap-4">
        <div className="flex min-w-0 flex-col gap-4">
          <OpenPositions account={account} data={data} positions={positions} locale={locale} reload={reload} />
          <Card aria-labelledby="checks-title" className="flex flex-col gap-4 px-6 py-5"><h2 id="checks-title" className="text-label text-subtle uppercase">{t("checks")}</h2><dl className="grid grid-cols-3 gap-x-6 gap-y-4">{(["run", "orders", "blocked"] as const).map(check => <div key={check} className="flex min-w-0 flex-col-reverse gap-1"><dt className="text-small text-muted">{t(`checks.${check}`)}</dt><dd className="num text-figure-medium">{formatAmount(String(data.checks_today[check]), 0, locale)}</dd></div>)}</dl></Card>
        </div>
        <aside aria-label={t("discipline")} className="positions-aside min-w-0 flex-col gap-4"><Card data-risk={disciplineState(data)} className="risk-panel flex flex-col gap-4 p-card">
          <div className="flex items-center justify-between gap-3"><h2 className="text-heading">{t("discipline")}</h2><Link href={`/rules?account=${encodeURIComponent(account.id)}`} className="inline-flex min-h-control items-center py-3 text-small text-muted underline">{t("editRules")}</Link></div>
          <DisciplineCounts status={data} locale={locale} /><p className="border-t border-line pt-3.5 text-small text-subtle">{t("discipline.note")}</p>
        </Card></aside>
      </section>
    </>}
  </>;
}

function Headline({ data, locale }: { data: RiskStatus; locale: Locale }) {
  const t = getMessages("overview", locale);
  const state = overviewState(data);
  const Icon = { safe: CircleCheck, caution: TriangleAlert, limit: Octagon }[state];
  return <section aria-labelledby="overview-title" data-risk={state} className="overview-command risk-panel flex flex-wrap items-end justify-between gap-x-8 gap-y-4">
    <div className="flex flex-col gap-2"><p className="text-label text-subtle uppercase">{t("status")} <time dateTime={data.as_of} className="num">{formatRiskTime(data.as_of, locale)} UTC</time></p>
      <div className="flex items-center gap-3.5"><Icon aria-hidden="true" className={`size-8 shrink-0 ${{ safe: "text-safe", caution: "text-caution", limit: "text-limit" }[state]}`} /><h1 id="overview-title" className="text-display">{t(state === "safe" ? "headline" : `headline.${state}`)}</h1></div>
      <p className="max-w-[640px] text-body text-muted"><NumberText text={t("summary", { amount: formatAmount(data.open_risk.amount, 2, locale), currency: data.currency, percent: formatPercent(data.open_risk.percent, 2, locale) })} /></p>
    </div>
    <div className="flex flex-wrap gap-8 [&_.text-figure-medium]:text-xl [&_.text-small]:text-xs">
      <Stat label={t("realized")} figure={<PnL locale={locale} value={data.realized_today} />} />
      <Stat label={t("unrealized")} figure={<PnL locale={locale} value={data.unrealized} />} />
      <Stat label={t("positions")} figure={<>{formatAmount(String(data.open_positions.count), 0, locale)} <span className="text-subtle"><NumberText text={t("of", { limit: formatAmount(String(data.open_positions.limit), 0, locale) })} /></span></>} />
    </div>
  </section>;
}

function OpenPositions({ account, data, positions, locale, reload }: { account: TradingAccount; data: RiskStatus; positions: Position[]; locale: Locale; reload: () => Promise<void> }) {
  const t = getMessages("overview", locale);
  const portfolio = getMessages("portfolio", locale);
  const { refresh } = useAuth();
  const [confirmId, setConfirmId] = useState<string | null>(null);
  const [closingId, setClosingId] = useState<string | null>(null);
  const [error, setError] = useState<unknown>(null);
  const [alreadyClosed, setAlreadyClosed] = useState(false);
  const pending = useRef(false);
  const lifetime = useRef<AbortController | null>(null);
  useEffect(() => {
    const controller = new AbortController();
    lifetime.current = controller;
    return () => controller.abort();
  }, []);
  async function closePosition(id: string) {
    if (pending.current) return;
    const signal = lifetime.current!.signal;
    pending.current = true;
    setClosingId(id); setError(null); setAlreadyClosed(false);
    try {
      await apiRequest(`/positions/${encodeURIComponent(id)}/close`, {}, AbortSignal.any([signal, AbortSignal.timeout(15_000)]));
      if (signal.aborted) return;
      setConfirmId(null);
      await reload();
    } catch (error) {
      if (signal.aborted) return;
      if (error instanceof ApiError && error.status === 409) { setAlreadyClosed(true); setConfirmId(null); await reload(); }
      else setError(error);
      if (error instanceof ApiError && error.status === 401) void refresh();
    } finally { pending.current = false; if (!signal.aborted) setClosingId(null); }
  }
  const rows = positions.map(position => ({ position, distance: distanceToStop(position.last_price, position.stop, position.side) })).sort((a, b) => a.distance === null ? b.distance === null ? 0 : 1 : b.distance === null ? -1 : compareDecimals(a.distance, b.distance));
  const price = (value: string | null | undefined) => value == null ? portfolio("unavailable") : formatTradePrice(value, locale);
  return <Card aria-labelledby="positions-title" className="flex min-w-0 flex-col overflow-hidden">
    <div className="flex flex-wrap items-center justify-between gap-3 px-6 py-5"><h2 id="positions-title" className="text-heading">{t("positions")}</h2><span className="text-small text-subtle">{t("sorted")}</span></div>
    {error !== null && <p role="alert" className="px-6 pb-4 text-limit">{tradingError(error, locale)}</p>}
    {alreadyClosed && <p role="status" className="px-6 pb-4 text-caution">{portfolio("alreadyClosed")}</p>}
    {rows.length === 0 ? <p className="border-t border-line p-6 text-muted">{t("emptyPositions")} <Link href={`/check?account=${encodeURIComponent(account.id)}`} className="inline-flex min-h-control items-center underline">{t("preTradeCheck")}</Link></p> : <div role="region" aria-label={t("positions.table")} tabIndex={0} className="min-w-0 overflow-x-auto">
      <table className="positions-table w-full border-collapse whitespace-nowrap text-right">
        <thead className="text-xs text-subtle"><tr>{columns.map(column => <th key={column} scope="col" className={`border-t border-line px-3 py-2.5 font-medium first:px-6 ${column === "pair" || column === "toStop" ? "text-left" : ""}`}>{t(`column.${column}`)}</th>)}{account.type === "live" && <th scope="col" className="border-t border-line px-3 py-2.5 font-medium">{portfolio("exchangeStop")}</th>}<th scope="col" className="border-t border-line py-2.5 pr-6 pl-3 font-medium">{portfolio("actions")}</th></tr></thead>
        <tbody>{rows.map(({ position, distance }) => <tr key={position.id} className="[&>td]:border-t [&>td]:border-line [&>td]:px-3 [&>td]:py-3.5">
          <th scope="row" className="border-t border-line px-6 py-3.5 text-left font-normal"><div className="flex flex-col gap-0.5"><span className="font-semibold">{position.symbol}</span><div className="flex flex-wrap items-center gap-1 text-xs text-subtle">{(position.coin_class === "core" || position.coin_class === "large" || position.coin_class === "small" || position.coin_class === "stable") && <ClassTag locale={locale} coinClass={position.coin_class} className="rounded-none border-0 p-0 text-subtle" />}<span aria-hidden="true">{t("separator")}</span><span>{portfolio(position.side)}</span><span className="num">{formatPrice(position.quantity, undefined, locale)}</span></div></div></th>
          <td className="num text-muted">{price(position.entry_price)}</td><td className="num">{price(position.last_price)}</td><td className="num text-muted">{price(position.stop)}</td>
          <td className={`num text-left ${distance !== null && compareDecimals(distance, "0") <= 0 ? "text-limit" : ""}`}>{formatStopDistance(distance, locale)}</td>
          <td className="num">{formatAmount(position.initial_risk, 2, locale)}</td><td className="num">{position.unrealized_pnl == null ? portfolio("unavailable") : <PnL locale={locale} value={position.unrealized_pnl} />}</td>
          {account.type === "live" && <td className={`max-w-xs whitespace-normal text-small ${position.exchange_stop?.status === "failed" ? "text-limit" : position.exchange_stop?.status === "placed" ? "text-safe" : "text-muted"}`}><span role={position.exchange_stop?.status === "failed" ? "alert" : undefined}>{!position.exchange_stop ? portfolio("unavailable") : position.exchange_stop.status === "placed" ? portfolio("exchangeStop.placed", { id: position.exchange_stop.order_id ?? portfolio("unavailable") }) : portfolio(`exchangeStop.${position.exchange_stop.status}`)}</span></td>}
          <td className="!pr-6">{confirmId === position.id ? <div role="group" aria-label={t("closeConfirm", { symbol: position.symbol })} className="flex max-w-xs flex-col items-end gap-2 whitespace-normal"><p className="text-small">{t("closeConfirm", { symbol: position.symbol })}</p><div className="flex gap-2"><Button variant="outline" disabled={closingId !== null} onClick={() => setConfirmId(null)}>{t("cancel")}</Button><Button disabled={closingId !== null} onClick={() => void closePosition(position.id)}>{t(closingId === position.id ? "closing" : "confirmClose")}</Button></div></div> : <Button variant="outline" disabled={closingId !== null} onClick={() => { setConfirmId(position.id); setError(null); }} aria-label={portfolio("closePosition", { symbol: position.symbol })}>{portfolio("close")}</Button>}</td>
        </tr>)}</tbody>
        <tfoot><tr className="font-medium [&>*]:border-t [&>*]:border-line"><th scope="row" colSpan={5} className="px-6 py-3.5 text-left font-medium text-muted">{t("totalRisk")}</th><td className="num px-3 py-3.5">{formatAmount(data.open_risk.amount, 2, locale)}</td><td className="num px-3 py-3.5"><PnL locale={locale} value={data.unrealized} /></td><td colSpan={account.type === "live" ? 2 : 1} /></tr></tfoot>
      </table>
    </div>}
    <p className="border-t border-line px-6 py-3.5 text-small text-subtle">{t("sorted")} {t("separator")} {t("refreshNote")}</p>
  </Card>;
}
