import { ApiError, apiError, errorText } from "@/lib/auth";
import { compareDecimals } from "@/lib/format";
import type { Locale } from "@/lib/messages";
import type { ProfileSource } from "@/lib/risk";

export interface RuleCheck {
  rule_id: string;
  status: "pass" | "caution" | "fail";
  values: Record<string, string>;
  message_en: string;
  message_tr: string;
}

/** API decimals remain strings throughout forms, requests and result rendering. */
export interface CheckResponse {
  profile_source: ProfileSource;
  symbol: string;
  side: "long" | "short";
  coin_class: "core" | "large" | "small" | "stable";
  market: "spot" | "futures";
  leverage: number;
  entry_price: string;
  account_equity: string;
  atr: string;
  atr_multiplier: string;
  stop: string;
  suggested_stop?: string;
  stop_percent: string;
  risk_percent: string;
  risk_amount: string;
  position_size: string;
  notional: string;
  actual_risk: string;
  target_risk_reward: string;
  target_price: string;
  initial_margin?: string;
  liquidation_price?: string;
  checks: RuleCheck[];
  allowed: boolean;
  verdict: "pass" | "caution" | "fail";
  passed: number;
  cautions: number;
  failed: number;
}

export interface OrderRequest {
  account_id: string;
  symbol: string;
  side: "long" | "short";
  manual_stop?: string;
  leverage?: number;
}

export interface Position {
  id: string;
  symbol: string;
  side: "long" | "short";
  quantity: string;
  entry_price: string;
  stop: string;
  leverage: number | null;
  coin_class: string;
  initial_risk: string;
  status: "open" | "closed";
  opened_at: string;
  closed_at: string | null;
  exit_price: string | null;
  realized_pnl: string | null;
  close_reason: "manual" | "stop" | null;
  last_price?: string | null;
  unrealized_pnl?: string | null;
}

export function checkBlocked(check: CheckResponse): boolean {
  return !check.allowed || check.checks.some(rule => rule.status === "fail");
}

export function rejectedCheck(error: unknown): CheckResponse | null {
  if (!(error instanceof ApiError) || error.status !== 422 || tradingErrorCode(error) !== "risk_blocked") return null;
  const check = (error.data as { check?: CheckResponse }).check;
  return check && Array.isArray(check.checks) ? { ...check, allowed: false, verdict: "fail", profile_source: "stored" } : null;
}

export function tradingErrorCode(error: unknown): string | null {
  if (!(error instanceof ApiError) || !error.data || typeof error.data !== "object") return null;
  const code = (error.data as { code?: unknown }).code;
  return typeof code === "string" ? code : null;
}

export function tradingError(error: unknown, locale: Locale): string {
  if (error instanceof ApiError && error.data && typeof error.data === "object") {
    const data = error.data as { error?: { message_en?: string; message_tr?: string } | string; error_tr?: string };
    const message = typeof data.error === "object" && data.error
      ? locale === "tr" ? data.error.message_tr : data.error.message_en
      : locale === "tr" ? data.error_tr : data.error;
    if (typeof message === "string" && message) return message;
  }
  return errorText(apiError(error), locale);
}

export function validStop(value: string): boolean {
  return /^[0-9]{1,13}(\.[0-9]{1,8})?$/.test(value)
    && compareDecimals(value, "0") > 0 && compareDecimals(value, "1000000000000") <= 0;
}

/** Only the final bounded geometry is converted to Number; source decimals stay exact. */
export function meterWidth(value: string, limit: string): number {
  if (!value || !limit || compareDecimals(limit, "0") <= 0 || compareDecimals(value, "0") <= 0) return 0;
  if (compareDecimals(value, limit) >= 0) return 100;
  const scale = Math.max(value.split(".")[1]?.length ?? 0, limit.split(".")[1]?.length ?? 0);
  const integer = (amount: string) => {
    const [whole, fraction = ""] = amount.split(".");
    return BigInt(whole + fraction.padEnd(scale, "0"));
  };
  return Number(integer(value) * BigInt(10000) / integer(limit)) / 100;
}
