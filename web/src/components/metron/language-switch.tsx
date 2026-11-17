"use client";

import { useTransition } from "react";
import { useRouter } from "next/navigation";
import { getMessages, localeCookie } from "@/lib/messages";
import { useLocale } from "./locale-provider";

export function LanguageSwitch() {
  const locale = useLocale();
  const t = getMessages("common", locale);
  const router = useRouter();
  const [pending, startTransition] = useTransition();
  return <div role="group" aria-label={t("language.label")} aria-busy={pending}
    className="language-switch inline-flex h-11 shrink-0 items-center rounded-control border border-line bg-surface">
    {(["en", "tr"] as const).map(next => <button key={next} type="button" lang={next}
      aria-label={t(next === "en" ? "language.english" : "language.turkish")}
      aria-pressed={locale === next} disabled={pending}
      className="h-full min-h-0 px-2 text-xs font-bold tracking-wider"
      onClick={() => {
        if (next === locale) return;
        document.cookie = `${localeCookie}=${next}; Path=/; Max-Age=31536000; SameSite=Lax${location.protocol === "https:" ? "; Secure" : ""}`;
        startTransition(() => router.refresh());
      }}>{next.toUpperCase()}</button>)}
  </div>;
}
