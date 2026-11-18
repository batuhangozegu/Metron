import { compareDecimals, formatAmount } from "@/lib/format";
import { getMessages, type Locale, type MessageKey } from "@/lib/messages";

export interface User {
  id: string;
  email: string;
  two_factor_enabled: boolean;
  created_at: string;
}

export interface TradingAccount {
  id: string;
  user_id: string;
  type: "paper" | "live";
  market?: "spot" | "futures";
  name: string;
  quote_asset: string;
  paper_balance: string | null;
  created_at: string;
}

export class ApiError extends Error {
  constructor(public status: number, public retryAfter: string | null = null) {
    super(`API request failed (${status})`);
  }
}

/** Only same-origin requests: the proxy forwards HttpOnly cookies and refreshes them. */
export async function apiRequest<T>(path: string, body?: object, signal?: AbortSignal): Promise<T> {
  const response = await fetch(`/v1${path}`, {
    method: body === undefined ? "GET" : "POST",
    credentials: "same-origin",
    cache: "no-store",
    headers: body === undefined ? undefined : { "Content-Type": "application/json" },
    body: body === undefined ? undefined : JSON.stringify(body),
    signal: signal ?? AbortSignal.timeout(15_000),
  });
  if (!response.ok) throw new ApiError(response.status, response.headers.get("Retry-After"));
  return response.json() as Promise<T>;
}

export type ErrorContext = "session" | "login" | "register" | "code" | "disable" | "settings" | "accounts";
export interface UiError { key: MessageKey<"auth">; seconds?: string }

export function apiError(error: unknown, context: ErrorContext = "session"): UiError {
  if (!(error instanceof ApiError)) return { key: "error.network" };
  switch (error.status) {
    case 400: return { key: "error.badRequest" };
    case 401: return { key: context === "login" ? "error.credentials" : context === "code" ? "error.code" : context === "disable" ? "error.disable" : "error.session" };
    case 409: return { key: context === "register" ? "error.duplicate" : "error.conflict" };
    case 429: {
      const value = error.retryAfter;
      const seconds = value && /^\d+$/.test(value) ? value
        : value && Number.isFinite(Date.parse(value)) ? String(Math.max(1, Math.ceil((Date.parse(value) - Date.now()) / 1000))) : "900";
      return { key: "error.rateLimit", seconds };
    }
    case 503: return { key: "error.unavailable" };
    default: return { key: "error.unknown" };
  }
}

export function errorText(error: UiError, locale: Locale): string {
  return getMessages("auth", locale)(error.key, { seconds: formatAmount(error.seconds ?? "0", 0, locale) });
}

export function validPassword(value: string): boolean {
  const length = Array.from(value).length;
  return length >= 12 && length <= 128;
}

export function validCode(value: string): boolean { return /^[0-9]{6}$/.test(value); }

/** Match the API's plain decimal syntax and compare without converting money to Number. */
export function validPaperBalance(value: string): boolean {
  return /^[0-9]{1,8}(\.[0-9]{1,8})?$/.test(value)
    && compareDecimals(value, "0") > 0 && compareDecimals(value, "10000000") <= 0;
}
