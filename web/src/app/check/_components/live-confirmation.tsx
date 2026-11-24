"use client";

import { useEffect, useRef } from "react";
import { Button } from "@/components/ui/button";
import { formatPrice } from "@/lib/format";
import { getMessages, type Locale } from "@/lib/messages";
import type { CheckResponse } from "@/lib/trading";

export function LiveConfirmation({ check, locale, onCancel, onConfirm }: { check: CheckResponse; locale: Locale; onCancel: () => void; onConfirm: () => void }) {
  const dialog = useRef<HTMLDialogElement>(null);
  const t = getMessages("check", locale);
  useEffect(() => {
    const element = dialog.current!;
    const previous = document.activeElement;
    element.showModal();
    return () => { element.close(); if (previous instanceof HTMLElement) previous.focus(); };
  }, []);
  return <dialog ref={dialog} aria-labelledby="live-confirm-title" aria-describedby="live-confirm-note" onCancel={event => { event.preventDefault(); onCancel(); }} className="m-auto max-h-[90vh] w-[min(92vw,560px)] overflow-y-auto rounded-card border-2 border-limit bg-surface p-6 text-text backdrop:bg-black/75">
    <h2 id="live-confirm-title" className="text-heading font-semibold text-limit">{t("confirm.title")}</h2>
    <p id="live-confirm-note" className="mt-3 text-small text-limit">{t("confirm.note")}</p>
    <dl className="my-6 grid grid-cols-2 gap-3 text-sm">
      {([
        [t("confirm.symbol"), check.symbol], [t("side"), t(check.side)],
        [t("confirm.quantity"), formatPrice(check.position_size, undefined, locale)],
        [t("confirm.notional"), formatPrice(check.notional, undefined, locale)],
        [t("stop"), formatPrice(check.stop, undefined, locale)],
      ]).map(([label, value]) => <div key={label}><dt className="text-muted">{label}</dt><dd className="num mt-1 break-all">{value}</dd></div>)}
    </dl>
    <div className="flex flex-wrap justify-end gap-3"><Button variant="outline" autoFocus onClick={onCancel}>{t("confirm.cancel")}</Button><Button className="border border-limit bg-limit text-inverse hover:bg-limit" onClick={onConfirm}>{t("confirm.submit")}</Button></div>
  </dialog>;
}
