import { apiRequest } from "@/lib/auth";

export interface PriceAlert {
  id: string;
  symbol: string;
  direction: "above" | "below";
  price: string;
  note: string | null;
  status: "active" | "triggered" | "cancelled";
  created_at: string;
  triggered_at: string | null;
  triggered_price: string | null;
}

export const alertPollInterval = 15_000;
export const alertsChanged = "metron-alerts-triggered";
export const alertStorageKey = (userId: string) => `metron-alerts-seen:${userId}`;

/** Validate plain positive decimal syntax without rounding or imposing order-stop limits. */
export function validAlertPrice(value: string): boolean {
  return /^\d+(?:\.\d+)?$/.test(value) && /[1-9]/.test(value);
}

export function readAlertCursor(storage: Pick<Storage, "getItem">, key: string, now: string): string {
  try {
    const saved = storage.getItem(key);
    if (saved && Number.isFinite(Date.parse(saved)) && Date.parse(saved) <= Date.parse(now)) return new Date(saved).toISOString();
  } catch { /* Storage can be disabled by browser privacy settings. */ }
  return now;
}

export function saveAlertCursor(storage: Pick<Storage, "setItem">, key: string, cursor: string): void {
  try { storage.setItem(key, cursor); } catch { /* Notifications still work without persistence. */ }
}

export async function pollTriggeredAlerts(cursor: string, signal: AbortSignal): Promise<PriceAlert[]> {
  return apiRequest<PriceAlert[]>(`/alerts?${new URLSearchParams({ triggered_since: cursor })}`, undefined, signal);
}

export function unseenAlerts(alerts: PriceAlert[], cursor: string, seen: Set<string>): PriceAlert[] {
  return alerts.filter(alert => {
    if (alert.status !== "triggered" || !alert.triggered_at || Date.parse(alert.triggered_at) <= Date.parse(cursor) || seen.has(alert.id)) return false;
    seen.add(alert.id);
    return true;
  });
}
