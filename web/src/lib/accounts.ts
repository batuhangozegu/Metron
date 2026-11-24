import type { TradingAccount } from "@/lib/auth";
import { getMessages, type Locale } from "@/lib/messages";
import type { OrderRequest } from "@/lib/trading";

export function tradableAccounts(accounts: TradingAccount[]): TradingAccount[] {
  return accounts.filter(account => account.quote_asset === "USDT" && (account.type === "paper" || account.type === "live"));
}

export function requiresLiveConfirmation(account: TradingAccount): boolean {
  return account.type === "live" && account.environment === "mainnet";
}

/** Keep the previous paper default when a user also has live accounts. */
export function selectTradingAccount(accounts: TradingAccount[], selectedId: string): TradingAccount | undefined {
  return accounts.find(account => account.id === selectedId) ?? accounts.find(account => account.type === "paper") ?? accounts[0];
}

/** A mainnet preview cannot become an executable request without explicit confirmation. */
export function orderForExecution(account: TradingAccount, order: OrderRequest, confirmed: boolean): OrderRequest | null {
  if (requiresLiveConfirmation(account)) return confirmed ? { ...order, confirm_live: true } : null;
  return order;
}

/** Native select options support text only; repeat the selected account's chips beside them. */
export function accountLabel(account: TradingAccount, locale: Locale): string {
  const t = getMessages("accounts", locale);
  return [account.name, t(account.market === "futures" ? "market.futures" : "market.spot"),
    ...(account.type === "live" ? [t("badge.live"), account.environment ? t(account.environment) : "", account.key_hint ?? ""] : [t("paper")])].filter(Boolean).join(" · ");
}
