"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useLocale } from "@/components/metron/locale-provider";
import { getMessages } from "@/lib/messages";
import { cn } from "@/lib/utils";

const routes = [
  ["/", "nav.overview"], ["/check", "nav.check"], ["/markets", "nav.markets"],
  ["/portfolio", "nav.portfolio"], ["/alerts", "nav.alerts"], ["/rules", "nav.rules"],
] as const;

/** Wrapping app navigation with an accessible active-route indicator. */
export function Navigation() {
  const locale = useLocale();
  const t = getMessages("common", locale);
  const accounts = getMessages("accounts", locale);
  const pathname = usePathname();

  return <nav aria-label={t("nav.label")} className="command-nav flex flex-auto flex-wrap gap-1">
    {routes.map(([href, key]) => {
      const active = pathname === href || (href !== "/" && pathname.startsWith(`${href}/`));
      return <Link key={href} href={href} aria-current={active ? "page" : undefined}
        className={cn("inline-flex min-h-control items-center rounded-control px-3.5 py-3 text-sm leading-5 hover:bg-raised", active ? "bg-raised font-medium text-text" : "text-muted")}>{t(key)}</Link>;
    })}
    <Link href="/accounts" aria-current={pathname === "/accounts" ? "page" : undefined}
      className={cn("inline-flex min-h-control items-center rounded-control px-3.5 py-3 text-sm leading-5 hover:bg-raised", pathname === "/accounts" ? "bg-raised font-medium text-text" : "text-muted")}>{accounts("title")}</Link>
  </nav>;
}
