import Link from "next/link";
import type { ReactNode } from "react";
import { getMessages, type Locale } from "@/lib/messages";
import { formatAmount } from "@/lib/format";
import { account } from "@/lib/sample-data";
import { Navigation } from "./navigation";
import { LanguageSwitch } from "./language-switch";
import { ThemeToggle } from "./theme-toggle";

export interface AppShellProps { locale?: Locale; children: ReactNode }

/** Shared wrapping header and constrained page canvas; rendered once by the root layout. */
export function AppShell({ children, locale = "en" }: AppShellProps) {
  const t = getMessages("common", locale);
  return <>
    <a href="#main-content" className="sr-only focus:not-sr-only focus:absolute focus:z-50 focus:rounded-control focus:bg-text focus:p-3 focus:text-inverse">{t("skip")}</a>
    <header className="flex flex-wrap items-center gap-x-8 gap-y-3 border-b border-line bg-surface px-page py-2.5">
      <Link href="/" className="flex min-h-control items-center gap-2.5 text-[17px] font-semibold tracking-[0.02em]">
        <svg aria-hidden="true" width="22" height="22" viewBox="0 0 22 22" fill="none" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round"><path d="M2 16h18M2 16V7M6.5 16v-4M11 16V7M15.5 16v-4M20 16V7" /></svg>
        {t("brand")}
      </Link>
      <Navigation />
      <div className="flex flex-wrap items-center gap-4">
        <span className="rounded-tag border border-line px-2 py-1 text-xs font-medium tracking-[0.06em] text-muted uppercase">{t("account.paper")}</span>
        <span className="num text-sm">{formatAmount(account.equity, 2, locale)} <span className="text-subtle">{t("account.currency")}</span></span>
        <Link href="/check" className="inline-flex min-h-control items-center rounded-control bg-text px-[18px] py-3 text-sm leading-5 font-semibold text-inverse">{t("order.new")}</Link>
        <LanguageSwitch />
        <ThemeToggle />
      </div>
    </header>
    <main id="main-content" tabIndex={-1} className="mx-auto flex w-full max-w-canvas flex-col gap-6 p-page outline-none">{children}</main>
  </>;
}
