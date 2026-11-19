"use client";

import { useCallback, useEffect, useRef, useState, useTransition } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Card, CardHeader, PnL } from "@/components/metron";
import { Button } from "@/components/ui/button";
import { useLocale } from "@/components/metron/locale-provider";
import { useAuth } from "@/components/metron/auth-provider";
import { ApiError, apiRequest } from "@/lib/auth";
import { formatPrice, formatTradePrice } from "@/lib/format";
import { getMessages, type Locale } from "@/lib/messages";
import { tradingError, type Position } from "@/lib/trading";
import { usePaperAccounts } from "@/lib/use-paper-accounts";

const table = "w-full border-collapse whitespace-nowrap text-sm text-right [&_th]:border-t [&_th]:border-line [&_th]:px-3 [&_th]:py-2.5 [&_th]:font-medium [&_td]:border-t [&_td]:border-line [&_td]:px-3 [&_td]:py-3.5 [&_th:first-child]:px-6 [&_td:first-child]:px-6 [&_th:last-child]:pr-6 [&_td:last-child]:pr-6";

export function Positions({ initialAccountId }: { initialAccountId?: string }) {
  const locale = useLocale();
  const t = getMessages("portfolio", locale);
  const at = getMessages("accounts", locale);
  const router = useRouter();
  const [selecting, startTransition] = useTransition();
  const { accounts, account, error, reload } = usePaperAccounts(initialAccountId);
  if (error !== null) return <div className="flex flex-col items-start gap-4"><p role="alert" className="text-limit">{tradingError(error, locale)}</p><Button onClick={() => void reload()}>{t("retry")}</Button></div>;
  if (accounts === null) return <p role="status" className="text-muted">{t("loading")}</p>;
  if (!account) return <p className="text-muted">{t("noAccounts")} <Link href="/accounts" className="inline-flex min-h-control items-center underline">{t("createAccount")}</Link></p>;
  return <section className="flex min-w-0 flex-col gap-6" aria-label={t("positions")}>
    <div className="flex flex-wrap items-center gap-3"><label htmlFor="portfolio-account" className="text-small text-muted">{t("account")}</label><select id="portfolio-account" className="min-h-control max-w-full rounded-control border border-line bg-raised px-3 text-text" value={account.id} disabled={selecting} onChange={event => {
      const id = event.target.value;
      startTransition(() => router.replace(`/portfolio?account=${encodeURIComponent(id)}`, { scroll: false }));
    }}>{accounts.map(item => <option key={item.id} value={item.id}>{item.name} · {at(item.market === "futures" ? "market.futures" : "market.spot")}</option>)}</select></div>
    <AccountPositions key={account.id} accountId={account.id} />
  </section>;
}

function AccountPositions({ accountId }: { accountId: string }) {
  const locale = useLocale();
  const t = getMessages("portfolio", locale);
  const { refresh } = useAuth();
  const [snapshot, setSnapshot] = useState<{ open: Position[]; closed: Position[] } | null>(null);
  const [loadError, setLoadError] = useState<unknown>(null);
  const [actionError, setActionError] = useState<unknown>(null);
  const [alreadyClosed, setAlreadyClosed] = useState(false);
  const [closingId, setClosingId] = useState<string | null>(null);
  const request = useRef<AbortController | null>(null);
  const lifetime = useRef<AbortController | null>(null);
  const pendingClose = useRef(false);

  const load = useCallback(async (replace = false) => {
    if (request.current && !replace) return;
    request.current?.abort();
    const controller = new AbortController();
    request.current = controller;
    const signal = AbortSignal.any([controller.signal, AbortSignal.timeout(15_000)]);
    try {
      const [open, closed] = await Promise.all([
        apiRequest<{ positions: Position[] }>(`/accounts/${encodeURIComponent(accountId)}/positions?status=open`, undefined, signal),
        apiRequest<{ positions: Position[] }>(`/accounts/${encodeURIComponent(accountId)}/positions?status=closed`, undefined, signal),
      ]);
      if (controller.signal.aborted) return;
      setSnapshot({ open: open.positions, closed: closed.positions });
      setLoadError(null);
    } catch (error) {
      if (controller.signal.aborted) return;
      setLoadError(error);
      if (error instanceof ApiError && error.status === 401) void refresh();
    } finally {
      if (request.current === controller) request.current = null;
    }
  }, [accountId, refresh]);

  useEffect(() => {
    const controller = new AbortController();
    lifetime.current = controller;
    const visibleRefresh = () => { if (document.visibilityState === "visible") void load(); };
    visibleRefresh();
    const timer = window.setInterval(visibleRefresh, 5_000);
    document.addEventListener("visibilitychange", visibleRefresh);
    return () => {
      controller.abort();
      request.current?.abort();
      request.current = null;
      window.clearInterval(timer);
      document.removeEventListener("visibilitychange", visibleRefresh);
    };
  }, [load]);

  const closePosition = useCallback(async (positionId: string) => {
    if (pendingClose.current) return;
    pendingClose.current = true;
    setClosingId(positionId);
    setActionError(null);
    setAlreadyClosed(false);
    try {
      await apiRequest(`/positions/${encodeURIComponent(positionId)}/close`, {}, AbortSignal.any([lifetime.current!.signal, AbortSignal.timeout(15_000)]));
      if (!lifetime.current?.signal.aborted) await load(true);
    } catch (error) {
      if (lifetime.current?.signal.aborted) return;
      if (error instanceof ApiError && error.status === 409) {
        setAlreadyClosed(true);
        await load(true);
      } else setActionError(error);
      if (error instanceof ApiError && error.status === 401) void refresh();
    } finally { pendingClose.current = false; setClosingId(null); }
  }, [load, refresh]);

  return <>
    {loadError !== null && <div className="flex flex-wrap items-center gap-3"><p role="alert" className="text-limit">{tradingError(loadError, locale)} {snapshot && t("stale")}</p><Button variant="outline" onClick={() => void load(true)}>{t("retry")}</Button></div>}
    {actionError !== null && <p role="alert" className="text-limit">{tradingError(actionError, locale)}</p>}
    {alreadyClosed && <p role="status" className="text-caution">{t("alreadyClosed")}</p>}
    {snapshot === null ? loadError === null && <p role="status" className="text-muted">{t("loading")}</p> : ([false, true] as const).map(closed => {
      const rows = closed ? snapshot.closed : snapshot.open;
      const title = t(closed ? "closedPositions" : "openPositions");
      return <Card key={String(closed)}>
        <CardHeader title={title} note={t(closed ? "closedNote" : "refreshNote")} />
        <PositionTable rows={rows} closed={closed} locale={locale} closingId={closingId} onClose={closePosition} />
      </Card>;
    })}
  </>;
}

export function PositionTable({ rows, closed, locale, closingId, onClose }: { rows: Position[]; closed: boolean; locale: Locale; closingId: string | null; onClose: (id: string) => void }) {
  const t = getMessages("portfolio", locale);
  const title = t(closed ? "closedPositions" : "openPositions");
  const amount = (value: string | null | undefined) => value == null ? t("unavailable") : formatPrice(value, undefined, locale);
  const price = (value: string | null | undefined) => value == null ? t("unavailable") : formatTradePrice(value, locale);
  const pnl = (position: Position) => {
    const value = closed ? position.realized_pnl : position.unrealized_pnl;
    return value == null ? t("unavailable") : <PnL locale={locale} value={value} />;
  };
  const closedTime = (position: Position) => position.closed_at ? new Intl.DateTimeFormat(locale === "tr" ? "tr-TR" : "en-GB", { dateStyle: "short", timeStyle: "medium", timeZone: "UTC" }).format(new Date(position.closed_at)) : t("unavailable");
  const closeButton = (position: Position) => <Button variant="outline" disabled={closingId !== null} onClick={() => onClose(position.id)} aria-label={t("closePosition", { symbol: position.symbol })}>{t(closingId === position.id ? "closing" : "close")}</Button>;
  const headers = (closed: boolean) => <tr><th scope="col" className="text-left">{t("pair")}</th><th scope="col">{t("side")}</th><th scope="col">{t("quantity")}</th><th scope="col">{t("entry")}</th><th scope="col">{t("stop")}</th><th scope="col">{t(closed ? "exit" : "last")}</th><th scope="col">{t(closed ? "realizedPnl" : "unrealizedPnl")}</th>{closed ? <><th scope="col">{t("howClosed")}</th><th scope="col">{t("closedUtc")}</th></> : <th scope="col">{t("actions")}</th>}</tr>;
  return <>{rows.length === 0 ? <p className="border-t border-line p-6 text-muted">{t(closed ? "emptyClosed" : "emptyOpen")}</p> : <>
    <ul aria-label={title} className="min-w-0 sm:hidden">{rows.map(position => <li key={position.id} className="flex min-w-0 flex-col gap-4 border-t border-line p-5">
      <div className="flex min-w-0 flex-wrap items-baseline justify-between gap-2"><h3 className="break-all font-semibold">{position.symbol}</h3><span className="text-small text-muted">{t(position.side)}</span></div>
      <dl className="grid min-w-0 grid-cols-2 gap-x-4 gap-y-3 text-sm">
        {([
          [t("quantity"), amount(position.quantity)],
          [t("entry"), price(position.entry_price)],
          [t("stop"), price(position.stop)],
          [t(closed ? "exit" : "last"), price(closed ? position.exit_price : position.last_price)],
          [t(closed ? "realizedPnl" : "unrealizedPnl"), pnl(position)],
          ...(closed ? [[t("howClosed"), position.close_reason ? t(`reason.${position.close_reason}`) : t("unavailable")], [t("closedUtc"), closedTime(position)]] : []),
        ]).map(([label, value], index) => <div key={index} className="min-w-0"><dt className="text-xs text-subtle">{label}</dt><dd className="num mt-1 break-words [overflow-wrap:anywhere] [&_*]:whitespace-normal">{value}</dd></div>)}
      </dl>
      {!closed && closeButton(position)}
    </li>)}</ul>
    <div role="region" aria-label={title} tabIndex={0} className="hidden overflow-x-auto sm:block"><table className={`${table} min-w-[850px]`}>
          <thead className="text-xs text-subtle">{headers(closed)}</thead>
          <tbody>{rows.map(position => <tr key={position.id}>
            <th scope="row" className="text-left font-semibold">{position.symbol}</th>
            <td>{t(position.side)}</td><td className="num">{amount(position.quantity)}</td><td className="num">{price(position.entry_price)}</td><td className="num">{price(position.stop)}</td><td className="num">{price(closed ? position.exit_price : position.last_price)}</td>
            <td>{pnl(position)}</td>
            {closed ? <><td>{position.close_reason ? t(`reason.${position.close_reason}`) : t("unavailable")}</td><td className="num">{closedTime(position)}</td></> : <td>{closeButton(position)}</td>}
          </tr>)}</tbody>
        </table></div></>}</>;
}
