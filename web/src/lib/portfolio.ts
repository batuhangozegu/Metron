import { apiRequest } from "@/lib/auth";
import type { CoinClass } from "@/components/metron/class-tag";

export type PortfolioPeriod = "30d" | "90d" | "all";
export function parsePortfolioPeriod(value: unknown): PortfolioPeriod {
  return value === "90d" || value === "all" ? value : "30d";
}

export interface EquityPoint { at: string; equity: string }
export interface Portfolio {
  account_id: string;
  market: "spot" | "futures";
  currency: "USDT";
  period: PortfolioPeriod;
  summary: {
    starting_balance: string; balance: string; cash: string; unrealized_pnl: string;
    equity: string; net_result: string; return_percent: string; fees_paid: string;
  };
  equity_series: EquityPoint[];
  benchmark_series: EquityPoint[];
  benchmark_return_percent: string | null;
  max_drawdown: { percent: string; peak: string; trough: string; peak_at: string; trough_at: string } | null;
  stats: {
    trades: number; wins: number; losses: number;
    win_rate: string | null; average_win: string | null; average_loss: string | null; win_loss_ratio: string | null;
    by_class: { coin_class: CoinClass; trades: number; wins: number; win_rate: string | null; net_pnl: string }[];
  };
  holdings: { symbol: string; side: "long" | "short"; quantity: string; average_cost: string; fees_paid: string }[];
}

export interface JournalTrade {
  position_id: string; symbol: string; side: "long" | "short"; coin_class: CoinClass;
  opened_at: string; closed_at: string; close_reason: "manual" | "stop";
  quantity: string; entry_price: string; exit_price: string; fees: string; net_pnl: string;
  r_multiple: string | null;
  rules: { rule: string; status: "caution" | "fail" }[];
}
export interface Journal { trades: JournalTrade[]; next_before: string | null }

/** Preserve API decimals; only chart geometry converts them to numbers. */
export function getPortfolio(accountId: string, period: PortfolioPeriod = "30d", signal?: AbortSignal): Promise<Portfolio> {
  return apiRequest(`/accounts/${encodeURIComponent(accountId)}/portfolio?period=${period}`, undefined, signal);
}

export function getJournal(accountId: string, before?: string, signal?: AbortSignal, limit = 50): Promise<Journal> {
  const query = new URLSearchParams({ limit: String(limit) });
  if (before) query.set("before", before);
  return apiRequest(`/accounts/${encodeURIComponent(accountId)}/journal?${query}`, undefined, signal);
}
