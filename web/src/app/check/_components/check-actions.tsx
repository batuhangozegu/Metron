"use client";

import Link from "next/link";
import { Button } from "@/components/ui/button";
import { useLocale } from "@/components/metron/locale-provider";
import { getMessages } from "@/lib/messages";

export function CheckActions({ blocked, busy, onOpen }: { blocked: boolean; busy: boolean; onOpen: () => void }) {
  const t = getMessages("check", useLocale());
  return <div className="flex flex-wrap items-center gap-4">
    <Button type="button" disabled={blocked || busy} onClick={onOpen}>{t(busy ? "opening" : "openOrder")}</Button>
    <Link href="/" className="inline-flex min-h-control items-center px-3 text-muted">{t("discard")}</Link>
    <p className="text-small text-subtle">{t(blocked ? "liveBlockedNote" : "executionNote")}</p>
  </div>;
}
