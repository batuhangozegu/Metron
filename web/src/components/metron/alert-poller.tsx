"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { useAuth } from "./auth-provider";
import { useLocale } from "./locale-provider";
import { Button } from "@/components/ui/button";
import { ApiError } from "@/lib/auth";
import { alertPollInterval, alertStorageKey, alertsChanged, pollTriggeredAlerts, readAlertCursor, saveAlertCursor, unseenAlerts, type PriceAlert } from "@/lib/alerts";
import { formatPrice } from "@/lib/format";
import { getMessages } from "@/lib/messages";

/** Mount once in the shell; no private requests or notifications outside a signed-in session. */
export function AlertPoller() {
  const { status, user, refresh } = useAuth();
  const locale = useLocale();
  const t = getMessages("alerts", locale);
  const [toasts, setToasts] = useState<{ userId: string; alert: PriceAlert }[]>([]);
  const userId = status === "authenticated" ? user?.id : undefined;
  useEffect(() => {
    if (!userId) return;
    const t = getMessages("alerts", locale);
    const controller = new AbortController();
    const key = alertStorageKey(userId);
    let cursor = new Date().toISOString();
    try { cursor = readAlertCursor(window.localStorage, key, cursor); } catch { /* Access itself may throw. */ }
    const seen = new Set<string>();
    let pending = false;
    let unauthorized = false;
    async function poll() {
      if (pending || unauthorized || controller.signal.aborted) return;
      pending = true;
      // Advance only after success, to the request start so concurrent triggers are not lost.
      const started = new Date().toISOString();
      try {
        const alerts = await pollTriggeredAlerts(cursor, AbortSignal.any([controller.signal, AbortSignal.timeout(15_000)]));
        if (controller.signal.aborted) return;
        const incoming = unseenAlerts(alerts, cursor, seen);
        if (incoming.length) {
          setToasts(previous => [...previous.filter(toast => toast.userId === userId), ...incoming.map(alert => ({ userId: userId!, alert }))]);
          window.dispatchEvent(new Event(alertsChanged));
          for (const alert of incoming) {
            try {
              if ("Notification" in window && Notification.permission === "granted") {
                new Notification(t("notificationTitle"), { tag: alert.id, body: t("notificationBody", { symbol: alert.symbol, price: formatPrice(alert.triggered_price ?? alert.price, undefined, locale), direction: t(alert.direction), target: formatPrice(alert.price, undefined, locale) }) });
              }
            } catch { /* The in-app toast remains available if the browser rejects notifications. */ }
          }
        }
        cursor = started;
        try { saveAlertCursor(window.localStorage, key, cursor); } catch { /* Access itself may throw. */ }
      } catch (error) {
        if (!controller.signal.aborted && error instanceof ApiError && error.status === 401) { unauthorized = true; void refresh(); }
      } finally { pending = false; }
    }
    void poll();
    const timer = window.setInterval(() => void poll(), alertPollInterval);
    return () => { controller.abort(); window.clearInterval(timer); };
  }, [userId, locale, refresh]);

  const visible = toasts.filter(toast => toast.userId === userId);
  if (!userId || !visible.length) return null;
  return <aside aria-label={t("title")} aria-live="polite" aria-relevant="additions" className="fixed right-4 bottom-4 z-50 flex max-h-[60vh] w-[min(92vw,380px)] flex-col gap-3 overflow-y-auto">
    {visible.map(({ alert }) => <div key={alert.id} role="status" className="rounded-card border border-caution bg-surface p-4 shadow-lg">
      <Link href="/alerts" className="font-semibold text-caution">{t("notificationTitle")}</Link>
      <p className="mt-2 text-small">{t("notificationBody", { symbol: alert.symbol, price: formatPrice(alert.triggered_price ?? alert.price, undefined, locale), direction: t(alert.direction), target: formatPrice(alert.price, undefined, locale) })}</p>
      {alert.note && <p className="mt-1 break-words text-small text-muted">{alert.note}</p>}
      <Button variant="outline" className="mt-3" onClick={() => setToasts(previous => previous.filter(toast => toast.alert.id !== alert.id))}>{t("dismiss")}</Button>
    </div>)}
  </aside>;
}
