"use client";

import { useCallback, useEffect, useRef, useState, type FormEvent } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { ClassTag } from "@/components/metron";
import { Button } from "@/components/ui/button";
import { useAuth } from "@/components/metron/auth-provider";
import { useLocale } from "@/components/metron/locale-provider";
import { ApiError, apiRequest } from "@/lib/auth";
import { formatTradePrice } from "@/lib/format";
import { getMessages } from "@/lib/messages";
import { checkBlocked, rejectedCheck, tradingError, validStop, type CheckResponse, type OrderRequest } from "@/lib/trading";
import { usePaperAccounts } from "@/lib/use-paper-accounts";
import type { Market } from "@/app/markets/_lib/market-data";
import { CheckResult } from "./check-result";

const control = "min-h-control w-full rounded-control border border-line bg-raised px-3 text-text";

export function CheckScreen() {
  const locale = useLocale();
  const t = getMessages("check", locale);
  const at = getMessages("accounts", locale);
  const router = useRouter();
  const { refresh } = useAuth();
  const { accounts, account, setSelectedId, error: accountError, reload } = usePaperAccounts();
  const [markets, setMarkets] = useState<Market[] | null>(null);
  const [marketError, setMarketError] = useState<unknown>(null);
  const [symbol, setSymbol] = useState("");
  const [side, setSide] = useState<"long" | "short">("long");
  const [stop, setStop] = useState("");
  const [leverage, setLeverage] = useState("1");
  const [check, setCheck] = useState<CheckResponse | null>(null);
  const [checkedOrder, setCheckedOrder] = useState<OrderRequest | null>(null);
  const [busy, setBusy] = useState<"check" | "order" | null>(null);
  const [error, setError] = useState<unknown>(null);
  const [invalid, setInvalid] = useState(false);
  const pending = useRef(false);
  const lifetime = useRef<AbortController | null>(null);
  useEffect(() => {
    const controller = new AbortController();
    lifetime.current = controller;
    return () => controller.abort();
  }, []);

  const loadMarkets = useCallback((signal?: AbortSignal) => {
    return apiRequest<{ markets: Market[] }>("/markets", undefined, signal).then(data => {
      if (signal?.aborted) return;
      setMarkets(data.markets.filter(market => market.quote_asset === "USDT" && market.coin_class !== "stable"));
      setMarketError(null);
    }, error => {
      if (!signal?.aborted) setMarketError(error);
    });
  }, []);
  useEffect(() => {
    const controller = new AbortController();
    void loadMarkets(controller.signal);
    return () => controller.abort();
  }, [loadMarkets]);
  const market = markets?.find(market => market.symbol === symbol)
    ?? markets?.find(market => market.symbol === "BTCUSDT") ?? markets?.[0];
  const futures = account?.market === "futures";

  function invalidate() { setCheck(null); setCheckedOrder(null); setError(null); setInvalid(false); }

  async function runCheck(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (pending.current || !account || !market) return;
    if ((stop && !validStop(stop)) || (futures && (!/^\d{1,3}$/.test(leverage) || Number(leverage) < 1 || Number(leverage) > 125))) {
      setInvalid(true); return;
    }
    pending.current = true;
    setBusy("check");
    invalidate();
    const order: OrderRequest = { account_id: account.id, symbol: market.symbol, side: futures ? side : "long",
      ...(stop ? { manual_stop: stop } : {}), ...(futures ? { leverage: Number(leverage) } : {}) };
    const signal = AbortSignal.any([lifetime.current!.signal, AbortSignal.timeout(15_000)]);
    try {
      // Refresh the ticker immediately before the preview; order fills are priced by Go.
      const data = await apiRequest<{ markets: Market[] }>("/markets", undefined, signal);
      const latest = data.markets.find(item => item.symbol === order.symbol);
      if (!latest?.price) { setInvalid(true); return; }
      const result = await apiRequest<CheckResponse>("/risk/check", { ...order, entry_price: latest.price }, signal);
      if (signal.aborted) return;
      setCheck(result);
      setCheckedOrder(order);
      setMarkets(data.markets.filter(item => item.quote_asset === "USDT" && item.coin_class !== "stable"));
    } catch (error) {
      if (lifetime.current?.signal.aborted) return;
      setError(error);
      if (error instanceof ApiError && error.status === 401) void refresh();
    } finally { pending.current = false; setBusy(null); }
  }

  async function openOrder() {
    if (pending.current || !check || !checkedOrder || checkBlocked(check)) return;
    pending.current = true;
    setBusy("order");
    setError(null);
    try {
      await apiRequest("/orders", checkedOrder, AbortSignal.any([lifetime.current!.signal, AbortSignal.timeout(15_000)]));
      if (!lifetime.current?.signal.aborted) {
        invalidate();
        router.push(`/portfolio?account=${encodeURIComponent(checkedOrder.account_id)}`);
      }
    } catch (error) {
      if (lifetime.current?.signal.aborted) return;
      const blocked = rejectedCheck(error);
      if (blocked) setCheck(blocked);
      else setError(error);
      if (error instanceof ApiError && error.status === 401) void refresh();
    } finally { pending.current = false; setBusy(null); }
  }

  const loadError = accountError || marketError;
  if (loadError) return <div className="flex flex-col items-start gap-4"><p role="alert" className="text-limit">{tradingError(loadError, locale)}</p><Button onClick={() => { void reload(); void loadMarkets(); }}>{t("retry")}</Button></div>;
  if (accounts === null || markets === null) return <p role="status" className="text-muted">{t("loading")}</p>;
  if (!account) return <p className="text-muted">{t("noAccounts")} <Link href="/accounts" className="inline-flex min-h-control items-center underline">{t("createAccount")}</Link></p>;
  if (!market) return <div className="flex flex-col items-start gap-3"><p className="text-muted">{t("noMarkets")}</p><Button onClick={() => void loadMarkets()}>{t("retry")}</Button></div>;

  return <div className="flex min-w-0 flex-wrap items-start gap-6">
    <form aria-label={t("order")} onSubmit={runCheck} aria-busy={busy !== null} className="min-w-0 flex-[1_1_320px] rounded-card border border-line bg-surface p-card">
      <fieldset disabled={busy !== null} className="flex min-w-0 flex-col gap-5">
        <h1 className="text-heading">{t("order")}</h1>
        <div className="flex flex-col gap-2"><label htmlFor="check-account" className="text-small text-muted">{t("account")}</label>
          <select id="check-account" className={control} value={account.id} onChange={event => { setSelectedId(event.target.value); setSide("long"); invalidate(); }}>{accounts.map(item => <option key={item.id} value={item.id}>{item.name} · {at(item.market === "futures" ? "market.futures" : "market.spot")}</option>)}</select>
        </div>
        <div className="flex flex-col gap-2"><label htmlFor="pair" className="text-small text-muted">{t("pair")}</label>
          <select id="pair" className={control} value={market.symbol} onChange={event => { setSymbol(event.target.value); invalidate(); }}>{markets.map(item => <option key={item.symbol}>{item.symbol}</option>)}</select>
          <div className="flex flex-wrap items-center gap-2 text-xs text-muted"><ClassTag locale={locale} coinClass={market.coin_class} /><span>{t("last")} <span className="num">{market.price === null ? t("unavailable") : formatTradePrice(market.price, locale)}</span></span></div>
        </div>
        <div className="flex flex-col gap-2"><label htmlFor="side" className="text-small text-muted">{t("side")}</label>
          <select id="side" className={control} value={futures ? side : "long"} onChange={event => { setSide(event.target.value as "long" | "short"); invalidate(); }}><option value="long">{t("long")}</option>{futures && <option value="short">{t("short")}</option>}</select>
        </div>
        <div className="flex flex-col gap-2"><label htmlFor="manual-stop" className="text-small text-muted">{t("manualStop")}</label><input id="manual-stop" inputMode="decimal" className={`${control} num`} value={stop} onChange={event => { setStop(event.target.value); invalidate(); }} aria-describedby="stop-hint" /><p id="stop-hint" className="text-small text-subtle">{t("stopHint")}</p></div>
        {futures && <div className="flex flex-col gap-2"><label htmlFor="leverage" className="text-small text-muted">{t("leverage")}</label><input id="leverage" type="number" min="1" max="125" step="1" required className={`${control} num`} value={leverage} onChange={event => { setLeverage(event.target.value); invalidate(); }} /></div>}
        <p className="text-small text-subtle">{t("previewNote")}</p>
        {invalid && <p role="alert" className="text-small text-limit">{t("invalid")}</p>}
        <Button type="submit">{t(busy === "check" ? "checking" : "runCheck")}</Button>
      </fieldset>
    </form>
    <section aria-label={t("results")} className="flex min-w-0 flex-[999_1_600px] flex-col gap-4">
      {error !== null && <p role="alert" className="text-limit">{tradingError(error, locale)}</p>}
      {check ? <CheckResult locale={locale} check={check} busy={busy === "order"} onOpen={() => void openOrder()} /> : <p role="status" className="rounded-card border border-line bg-surface p-card text-muted">{t(busy === "check" ? "checking" : "awaitingCheck")}</p>}
    </section>
  </div>;
}
