import Link from "next/link";
import { Plus } from "lucide-react";
import type { ReactNode } from "react";
import { getMessages, type Locale } from "@/lib/messages";
import { Navigation } from "./navigation";
import { LanguageSwitch } from "./language-switch";
import { ThemeToggle } from "./theme-toggle";
import { MarketTicker } from "./market-ticker";
import { AuthGate } from "./auth-provider";
import { SessionActions } from "./session-actions";

export interface AppShellProps { locale?: Locale; children: ReactNode }

/** Shared wrapping header and constrained page canvas; rendered once by the root layout. */
export function AppShell({ children, locale = "en" }: AppShellProps) {
  const t = getMessages("common", locale);
  return <>
    <a href="#main-content" className="sr-only focus:not-sr-only focus:absolute focus:z-50 focus:rounded-control focus:bg-text focus:p-3 focus:text-inverse">{t("skip")}</a>
    <header className="command-header flex flex-wrap items-center gap-x-3 gap-y-3 border-b border-line bg-surface px-page py-2.5">
      <Link href="/" aria-label={t("brand")} className="flex min-h-control items-center gap-2.5 text-[17px] font-semibold tracking-[0.02em]">
        <svg aria-hidden="true" width="22" height="22" viewBox="0 0 22 22" fill="none" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round"><path d="M2 16h18M2 16V7M6.5 16v-4M11 16V7M15.5 16v-4M20 16V7" /></svg>
        <span className="brand-name">{t("brand")}</span>
      </Link>
      <Navigation />
      <div className="header-actions ml-auto flex shrink-0 items-center gap-2">
        <SessionActions />
        <Link href="/check" aria-label={t("order.new")} className="new-order inline-flex min-h-control items-center whitespace-nowrap rounded-control bg-text px-3 py-3 text-sm leading-5 font-semibold text-inverse"><Plus aria-hidden="true" className="new-order-icon hidden size-5" /><span>{t("order.new")}</span></Link>
        <LanguageSwitch />
        <ThemeToggle />
      </div>
    </header>
    <MarketTicker />
    <main id="main-content" tabIndex={-1} className="mx-auto flex w-full max-w-canvas flex-col gap-6 p-page outline-none"><AuthGate>{children}</AuthGate></main>
  </>;
}
