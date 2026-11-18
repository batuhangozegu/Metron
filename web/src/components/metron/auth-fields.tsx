"use client";

import type { ComponentProps } from "react";
import { Input } from "@/components/ui/input";
import { errorText, type UiError } from "@/lib/auth";
import { getMessages } from "@/lib/messages";
import { useLocale } from "./locale-provider";

export function FormError({ error }: { error: UiError | null }) {
  const locale = useLocale();
  return error ? <p role="alert" className="auth-error border border-limit bg-limit-tint p-3 text-sm text-limit">{errorText(error, locale)}</p> : null;
}

export function CodeField(props: ComponentProps<"input">) {
  const t = getMessages("auth", useLocale());
  const id = props.id ?? "auth-code";
  return <div className="flex flex-col gap-2">
    <label htmlFor={id} className="micro-label text-muted">{t("code")}</label>
    <Input id={id} name="code" type="text" inputMode="numeric" autoComplete="one-time-code" pattern="[0-9]{6}" maxLength={6} required aria-describedby={`${id}-hint`} className="num tracking-[0.25em]" {...props} />
    <p id={`${id}-hint`} className="text-small text-subtle">{t("code.hint")}</p>
  </div>;
}
