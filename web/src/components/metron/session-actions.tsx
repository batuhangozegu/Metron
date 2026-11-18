"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useRef, useState } from "react";
import { UserRound } from "lucide-react";
import { Button } from "@/components/ui/button";
import { apiError, apiRequest, type UiError } from "@/lib/auth";
import { getMessages } from "@/lib/messages";
import { useAuth } from "./auth-provider";
import { useLocale } from "./locale-provider";
import { FormError } from "./auth-fields";

export function SignOut() {
  const { acceptUser } = useAuth();
  const router = useRouter();
  const t = getMessages("auth", useLocale());
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<UiError | null>(null);
  async function signOut() {
    setBusy(true);
    setError(null);
    try {
      await apiRequest("/auth/logout", {});
      acceptUser(null);
      router.replace("/sign-in");
    } catch (error) { setError(apiError(error)); }
    finally { setBusy(false); }
  }
  return <div className="flex flex-col gap-2">
    <Button variant="outline" disabled={busy} onClick={() => void signOut()}>{t(busy ? "working" : "signOut")}</Button>
    <FormError error={error} />
  </div>;
}

export function SessionActions() {
  const { user, status } = useAuth();
  const t = getMessages("auth", useLocale());
  const menu = useRef<HTMLDetailsElement>(null);
  useEffect(() => {
    const close = (event: PointerEvent) => {
      if (!menu.current?.contains(event.target as Node)) menu.current?.removeAttribute("open");
    };
    document.addEventListener("pointerdown", close);
    return () => document.removeEventListener("pointerdown", close);
  }, []);
  if (status === "loading") return <span role="status" className="text-small text-subtle">{t("loading")}</span>;
  return user ? <details ref={menu} className="account-menu relative shrink-0" onKeyDown={event => {
    if (event.key === "Escape") { menu.current?.removeAttribute("open"); menu.current?.querySelector("summary")?.focus(); }
  }}>
    <summary aria-label={t("accountMenu")} title={t("accountMenu")} className="flex size-11 cursor-pointer list-none items-center justify-center border border-line text-muted hover:bg-raised"><UserRound aria-hidden="true" className="size-5" /></summary>
    <div className="account-panel absolute right-0 top-full z-50 mt-2 flex w-64 max-w-[calc(100vw-32px)] flex-col gap-3 border border-line bg-surface p-4 shadow-lg">
      <p className="break-all text-small text-muted">{user.email}</p>
      <Link href="/settings" onClick={() => menu.current?.removeAttribute("open")} className="inline-flex min-h-control items-center text-sm underline">{t("settings")}</Link>
      <SignOut />
    </div>
  </details> : <Link href="/sign-in" className="inline-flex min-h-control items-center whitespace-nowrap text-sm">{t("signIn")}</Link>;
}
