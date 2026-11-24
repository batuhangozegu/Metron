import { apiRequest } from "@/lib/auth";
import { compareDecimals, decimalProduct, decimalSum, formatPercent } from "@/lib/format";
import type { Locale } from "@/lib/messages";
import type { RiskState } from "@/lib/risk";

export interface RiskGauge { amount: string; percent: string; limit_percent: string }
export interface RiskCount { count: number; limit: number }
/** GET /accounts/{id}/risk-status: decimal values are plain strings. */
export interface RiskStatus {
  account_id: string;
  market: "spot" | "futures";
  as_of: string;
  currency: string;
  equity: string;
  realized_today: string;
  unrealized: string;
  open_risk: RiskGauge;
  daily_loss: RiskGauge;
  weekly_loss: RiskGauge;
  open_positions: RiskCount;
  losing_trades_today: RiskCount;
  stopped_trades_this_week: RiskCount;
  checks_today: { run: number; orders: number; blocked: number };
}

export function getRiskStatus(accountId: string, signal?: AbortSignal): Promise<RiskStatus> {
  return apiRequest(`/accounts/${encodeURIComponent(accountId)}/risk-status`, undefined, signal);
}

/** Match the existing rulers: caution at 75%, exceeded strictly above the limit. */
export function gaugeState(gauge: RiskGauge): RiskState {
  if (compareDecimals(gauge.limit_percent, "0") <= 0 || compareDecimals(gauge.percent, gauge.limit_percent) > 0) return "limit";
  return compareDecimals(gauge.percent, decimalProduct(gauge.limit_percent, "0.75")) >= 0 ? "caution" : "safe";
}

export function countState({ count, limit }: RiskCount): RiskState {
  return count >= limit ? "limit" : count === limit - 1 ? "caution" : "safe";
}

export function worstState(states: RiskState[]): RiskState {
  return states.includes("limit") ? "limit" : states.includes("caution") ? "caution" : "safe";
}

export function disciplineState(status: RiskStatus): RiskState {
  return worstState([countState(status.losing_trades_today), countState(status.stopped_trades_this_week)]);
}

export function overviewState(status: RiskStatus): RiskState {
  return worstState([gaugeState(status.open_risk), gaugeState(status.daily_loss), gaugeState(status.weekly_loss), disciplineState(status)]);
}

/** Equity × limit percentage minus current amount, clamped to zero. */
export function remainingAmount(equity: string, gauge: RiskGauge): string {
  const remaining = decimalSum(decimalProduct(decimalProduct(equity, gauge.limit_percent), "0.01"), decimalProduct(gauge.amount, "-1"));
  return compareDecimals(remaining, "0") > 0 ? remaining : "0";
}

/** Signed distance in the position's direction; negative means the stop was crossed.
 * Fixed-point division preserves tiny token prices and large quotes. */
export function distanceToStop(lastPrice: string | null | undefined, stop: string, side: "long" | "short"): string | null {
  if (lastPrice == null || compareDecimals(lastPrice, "0") <= 0) return null;
  const difference = decimalSum(side === "long" ? lastPrice : stop, decimalProduct(side === "long" ? stop : lastPrice, "-1"));
  const scale = Math.max(difference.split(".")[1]?.length ?? 0, lastPrice.split(".")[1]?.length ?? 0);
  const integer = (value: string) => {
    const [whole, fraction = ""] = value.split(".");
    return BigInt(whole + fraction.padEnd(scale, "0"));
  };
  const result = integer(difference) * BigInt(10_000_000_000) / integer(lastPrice);
  const digits = String(result < BigInt(0) ? -result : result).padStart(9, "0");
  return `${result < BigInt(0) ? "-" : ""}${digits.slice(0, -8)}.${digits.slice(-8)}`;
}

export function formatStopDistance(distance: string | null, locale: Locale): string {
  return distance === null ? "—" : formatPercent(distance, 2, locale);
}

export function formatRiskTime(asOf: string, locale: Locale): string {
  return new Intl.DateTimeFormat(locale === "tr" ? "tr-TR" : "en-GB", { dateStyle: "short", timeStyle: "medium", timeZone: "UTC" }).format(new Date(asOf));
}
