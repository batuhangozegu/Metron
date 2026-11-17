"use client";

import { useState } from "react";
import Link from "next/link";
import { formatAmount, formatPrice } from "@/lib/format";
import { useLocale } from "@/components/metron/locale-provider";
import { getMessages } from "@/lib/messages";
import { account } from "@/lib/sample-data";
import { NumberCopy } from "./number-copy";

export function CheckActions({ state, stop, reducedValue, reducedRisk }: { state: "safe" | "caution" | "limit"; stop: string; reducedValue?: string; reducedRisk?: string }) {
  const locale = useLocale();
  const t = getMessages("check", locale);
  // Store the event, so an open confirmation also follows a language refresh.
  const [feedbackKey, setFeedbackKey] = useState<"reduceFeedback" | "waitFeedback" | "placeFeedback" | "violationFeedback" | null>(null);
  const feedback = feedbackKey ? t(feedbackKey, { time: account.cooldown.waitUntil, value: formatAmount(reducedValue ?? "0"), risk: formatAmount(reducedRisk ?? "0"), currency: account.currency }) : "";
  return <div className="flex flex-wrap items-center gap-x-4 gap-y-3">
    <button type="button" onClick={() => setFeedbackKey(state === "limit" ? "reduceFeedback" : state === "caution" ? "waitFeedback" : "placeFeedback")} className="min-h-12 rounded-control bg-text px-[22px] py-3 text-[15px] font-semibold text-inverse"><NumberCopy text={t(state === "limit" ? "reduce" : state === "caution" ? "wait" : "place", { time: account.cooldown.waitUntil })} /></button>
    {state === "caution" && <button type="button" className="min-h-12 rounded-control border border-caution px-[22px] py-3 text-[15px] font-medium" onClick={() => setFeedbackKey("violationFeedback")}>{t("placeAnyway")}</button>}
    <Link href="/" className="inline-flex min-h-12 items-center px-3 py-3 text-[15px] text-muted">{t("discard")}</Link>
    <p className="min-w-0 flex-[1_1_260px] text-small text-subtle"><NumberCopy text={t(state === "limit" ? "blockedNote" : state === "caution" ? "cautionNote" : "safeNote", { value: formatAmount(reducedValue ?? "0"), risk: formatAmount(reducedRisk ?? "0"), currency: account.currency, stop: formatPrice(stop) })} /></p>
    {feedback && <p role="status" className="w-full rounded-control border border-line bg-raised p-3 text-small text-muted"><NumberCopy text={feedback} /></p>}
  </div>;
}
