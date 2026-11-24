"use client";

import { useCallback, useEffect, useRef, useState, type FormEvent } from "react";
import { Bell } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardHeader } from "@/components/metron/card";
import { useAuth } from "@/components/metron/auth-provider";
import { useLocale } from "@/components/metron/locale-provider";
import { ApiError, apiRequest } from "@/lib/auth";
import { alertsChanged, validAlertPrice, type PriceAlert } from "@/lib/alerts";
import { formatPrice } from "@/lib/format";
import { getMessages } from "@/lib/messages";
import { tradingError } from "@/lib/trading";
import type { Market } from "@/app/markets/_lib/market-data";

const control = "min-h-control w-full min-w-0 rounded-control border border-line bg-raised px-3 text-text";

export default function AlertsPage() {
  const locale = useLocale();
  const t = getMessages("alerts", locale);
  const { refresh } = useAuth();
  const [alerts, setAlerts] = useState<PriceAlert[] | null>(null);
  const [markets, setMarkets] = useState<Market[] | null>(null);
  const [loadError, setLoadError] = useState<unknown>(null);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState<string | null>(null);
  const [created, setCreated] = useState(false);
  const [permission, setPermission] = useState<NotificationPermission | null>(null);
  const pending = useRef(false);
  const lifetime = useRef<AbortController | null>(null);
  const load = useCallback((signal?: AbortSignal) => {
    return Promise.all([
      apiRequest<PriceAlert[]>("/alerts", undefined, signal),
      apiRequest<{ markets: Market[] }>("/markets", undefined, signal),
    ]).then(([alerts, data]) => {
      if (signal?.aborted) return;
      setAlerts(alerts); setMarkets(data.markets); setLoadError(null);
    }, error => {
      if (signal?.aborted) return;
      setLoadError(error);
      if (error instanceof ApiError && error.status === 401) void refresh();
    });
  }, [refresh]);
  useEffect(() => {
    const controller = new AbortController();
    lifetime.current = controller;
    void load(controller.signal);
    const onTriggered = () => void load(controller.signal);
    window.addEventListener(alertsChanged, onTriggered);
    return () => { controller.abort(); window.removeEventListener(alertsChanged, onTriggered); };
  }, [load]);

  async function create(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (pending.current) return;
    const form = event.currentTarget;
    const fields = new FormData(form);
    const symbol = String(fields.get("symbol") ?? "");
    const price = String(fields.get("price") ?? "");
    setCreated(false);
    if (!markets?.some(market => market.symbol === symbol) || !validAlertPrice(price)) { setError(t("invalid")); return; }
    pending.current = true; setBusy("create"); setError(null);
    try {
      await apiRequest<PriceAlert>("/alerts", { symbol, direction: fields.get("direction"), price, note: String(fields.get("note") ?? "").trim() }, AbortSignal.any([lifetime.current!.signal, AbortSignal.timeout(15_000)]));
      if (lifetime.current?.signal.aborted) return;
      form.reset(); setCreated(true); await load(lifetime.current?.signal);
    } catch (error) {
      if (lifetime.current?.signal.aborted) return;
      setError(error instanceof ApiError && error.status === 409 ? t("limit") : tradingError(error, locale));
      if (error instanceof ApiError && error.status === 401) void refresh();
    } finally { pending.current = false; setBusy(null); }
  }

  async function cancel(id: string) {
    if (pending.current) return;
    pending.current = true; setBusy(id); setError(null); setCreated(false);
    try {
      await apiRequest(`/alerts/${encodeURIComponent(id)}`, undefined, AbortSignal.any([lifetime.current!.signal, AbortSignal.timeout(15_000)]), "DELETE");
      if (!lifetime.current?.signal.aborted) await load(lifetime.current?.signal);
    } catch (error) {
      if (lifetime.current?.signal.aborted) return;
      setError(error instanceof ApiError && error.status === 404 ? t("error.notFound") : tradingError(error, locale));
      if (error instanceof ApiError && error.status === 401) void refresh();
    } finally { pending.current = false; setBusy(null); }
  }

  async function enableNotifications() {
    try { if ("Notification" in window) setPermission(await Notification.requestPermission()); } catch { setPermission("denied"); }
  }

  return <section className="flex min-w-0 flex-col gap-6">
    <div><p className="micro-label mb-3 flex items-center gap-2 text-subtle"><Bell aria-hidden="true" className="size-4" />{t("title")}</p><h1>{t("title")}</h1><p className="mt-3 text-muted">{t("note")}</p></div>
    <div className="flex flex-wrap items-center gap-3"><Button variant="outline" onClick={() => void enableNotifications()}>{t("notifications")}</Button>{permission === "denied" && <p role="status" className="text-small text-caution">{t("permissionDenied")}</p>}</div>
    {loadError !== null && <div className="flex flex-wrap items-center gap-3"><p role="alert" className="text-limit">{tradingError(loadError, locale)}</p><Button variant="outline" onClick={() => void load(lifetime.current?.signal)}>{t("retry")}</Button></div>}
    {error && <p role="alert" className="text-limit">{error}</p>}
    {created && <p role="status" className="text-profit">{t("created")}</p>}
    {alerts === null || markets === null ? loadError === null && <p role="status" className="text-muted">{t("loading")}</p> : <div className="grid min-w-0 items-start gap-6 lg:grid-cols-2">
      <Card><CardHeader title={t("list")} action={<Button variant="outline" onClick={() => void load(lifetime.current?.signal)}>{t("retry")}</Button>} />
        {!alerts.length ? <p className="border-t border-line p-6 text-muted">{t("empty")}</p> : <ul className="divide-y divide-line border-t border-line">{alerts.map(alert => <li key={alert.id} className="flex min-w-0 flex-col gap-3 p-6">
          <div className="flex flex-wrap items-center justify-between gap-3"><h2 className="font-semibold">{alert.symbol}</h2><span className={`micro-label border px-2 py-1 ${alert.status === "triggered" ? "border-caution text-caution" : "border-line text-muted"}`}>{t(alert.status)}</span></div>
          <p className="text-small">{t(alert.direction)} <span className="num">{formatPrice(alert.price, undefined, locale)}</span></p>
          {alert.triggered_price && <p className="num text-small text-caution">{t("triggeredPrice", { price: formatPrice(alert.triggered_price, undefined, locale) })}</p>}
          {alert.note && <p className="break-words text-small text-muted">{alert.note}</p>}
          <time className="num text-xs text-subtle" dateTime={alert.triggered_at ?? alert.created_at}>{t("time", { time: new Intl.DateTimeFormat(locale === "tr" ? "tr-TR" : "en-GB", { dateStyle: "short", timeStyle: "medium", timeZone: "UTC" }).format(new Date(alert.triggered_at ?? alert.created_at)) })}</time>
          {alert.status === "active" && <Button variant="outline" className="self-start" disabled={busy !== null} onClick={() => void cancel(alert.id)}>{t(busy === alert.id ? "working" : "cancel")}</Button>}
        </li>)}</ul>}
      </Card>
      <Card><CardHeader title={t("create")} /><form onSubmit={create} className="border-t border-line p-6" aria-busy={busy === "create"}>
        <fieldset disabled={busy !== null || !markets.length} className="flex min-w-0 flex-col gap-5">
          <div className="flex flex-col gap-2"><label htmlFor="alert-symbol" className="micro-label text-muted">{t("symbol")}</label><select id="alert-symbol" name="symbol" required className={control}>{markets.map(market => <option key={market.symbol} value={market.symbol}>{market.symbol}</option>)}</select></div>
          <div className="flex flex-col gap-2"><label htmlFor="alert-direction" className="micro-label text-muted">{t("direction")}</label><select id="alert-direction" name="direction" className={control}><option value="above">{t("above")}</option><option value="below">{t("below")}</option></select></div>
          <div className="flex flex-col gap-2"><label htmlFor="alert-price" className="micro-label text-muted">{t("price")}</label><Input id="alert-price" name="price" inputMode="decimal" required aria-describedby="alert-price-hint" /><p id="alert-price-hint" className="text-small text-subtle">{t("priceHint")}</p></div>
          <div className="flex flex-col gap-2"><label htmlFor="alert-note" className="micro-label text-muted">{t("noteLabel")}</label><Input id="alert-note" name="note" /></div>
          <Button type="submit" className="self-start">{t(busy === "create" ? "working" : "create")}</Button>
        </fieldset>
        {!markets.length && <p className="mt-3 text-muted">{t("marketEmpty")}</p>}
      </form></Card>
    </div>}
  </section>;
}
