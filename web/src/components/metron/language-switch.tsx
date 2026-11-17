"use client";

import { useTransition } from "react";
import { useRouter } from "next/navigation";
import { getMessages, localeCookie, parseLocale } from "@/lib/messages";
import { useLocale } from "./locale-provider";

export function LanguageSwitch() {
  const locale = useLocale();
  const t = getMessages("common", locale);
  const router = useRouter();
  const [pending, startTransition] = useTransition();
  return <label className="inline-flex min-h-control items-center gap-2 text-sm">
    <span className="sr-only">{t("language.label")}</span>
    <select aria-label={t("language.label")} value={locale} disabled={pending}
      className="min-h-control rounded-control border border-line bg-surface px-2 text-text"
      onChange={event => {
        const next = parseLocale(event.target.value);
        document.cookie = `${localeCookie}=${next}; Path=/; Max-Age=31536000; SameSite=Lax${location.protocol === "https:" ? "; Secure" : ""}`;
        startTransition(() => router.refresh());
      }}>
      <option value="en" lang="en">{t("language.english")}</option>
      <option value="tr" lang="tr">{t("language.turkish")}</option>
    </select>
  </label>;
}
