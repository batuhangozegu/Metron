import type { TradingAccount } from "@/lib/auth";
import { getMessages, type Locale } from "@/lib/messages";

export function AccountBadges({ account, locale }: { account: TradingAccount; locale: Locale }) {
  const t = getMessages("accounts", locale);
  return <span className="inline-flex flex-wrap items-center gap-2 text-xs">
    <span className={`border px-2 py-1 font-semibold ${account.type === "live" ? "border-limit text-limit" : "border-line text-muted"}`}>{t(account.type === "live" ? "badge.live" : "paper")}</span>
    {account.type === "live" && account.environment && <span className={`border px-2 py-1 ${account.environment === "testnet" ? "border-caution text-caution" : "border-limit text-limit"}`}>{t(account.environment)}</span>}
    {account.type === "live" && account.key_hint && <span className="num text-muted">{t("keyHint", { hint: account.key_hint })}</span>}
  </span>;
}
