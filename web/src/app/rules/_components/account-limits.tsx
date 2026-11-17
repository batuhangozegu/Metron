"use client";

import { useState } from "react";
import { Card } from "@/components/metron/card";
import { decimalProduct, formatAmount, reformatInput } from "@/lib/format";
import { useLocale } from "@/components/metron/locale-provider";
import { getMessages } from "@/lib/messages";
import type { AccountLimits as Limits } from "@/lib/sample-data";
import { NumericCopy } from "./numeric-copy";

export function AccountLimits({ initialLimits }: { initialLimits: Limits }) {
  const locale = useLocale();
  const t = getMessages("rules", locale);

  const [inputs, setInputs] = useState(() => ({ locale, values: {
    totalOpenRisk: formatAmount(initialLimits.totalOpenRisk, 2, locale),
    dailyLoss: formatAmount(initialLimits.dailyLoss, 2, locale),
    weeklyLoss: formatAmount(initialLimits.weeklyLoss, 2, locale),
    maxPositions: formatAmount(String(initialLimits.maxPositions), 0, locale),
    targetRiskReward: formatAmount(initialLimits.targetRiskReward, 1, locale),
    cautionAt: formatAmount(decimalProduct(String(initialLimits.cautionAt), "100"), 0, locale),
  } }));
  if (inputs.locale !== locale) setInputs({ locale, values: Object.fromEntries(Object.entries(inputs.values).map(([field, value]) => [field, reformatInput(value, inputs.locale, locale)])) as typeof inputs.values });
  const [saved, setSaved] = useState(false);
  const fields = ["totalOpenRisk", "dailyLoss", "weeklyLoss", "maxPositions", "targetRiskReward", "cautionAt"] as const;
  const hints = {
    totalOpenRisk: t("limits.totalOpenRisk.hint"),
    dailyLoss: t("limits.dailyLoss.hint"),
    weeklyLoss: t("limits.weeklyLoss.hint"),
  };

  return <form className="flex min-w-0 flex-col gap-6" onSubmit={(event) => { event.preventDefault(); setSaved(true); }}>
    <section className="flex flex-wrap items-end justify-between gap-x-8 gap-y-4">
      <div className="flex flex-col gap-2">
        <h1 className="text-title">{t("title")}</h1>
        <p className="max-w-[620px] text-body text-muted">{t("description")}</p>
      </div>
      <div className="flex flex-wrap items-center gap-3">
        <span role="status" className="text-small text-muted">{saved ? t("saved") : ""}</span>
        <button type="submit" className="min-h-control rounded-control bg-text px-5 text-sm font-semibold text-inverse">{t("save")}</button>
      </div>
    </section>
    <Card className="flex flex-col gap-5 p-card" aria-labelledby="account-limits-title">
      <div className="flex flex-col gap-1">
        <h2 id="account-limits-title" className="text-heading">{t("limits.title")}</h2>
        <p className="text-small text-subtle">{t("limits.description")}</p>
      </div>
      <div className="grid grid-cols-[repeat(auto-fit,minmax(min(100%,220px),1fr))] gap-x-6 gap-y-5">
        {fields.map((field) => {
          const hint = field in hints ? hints[field as keyof typeof hints] : undefined;
          return <div key={field} className="flex min-w-0 flex-col gap-2">
            <label htmlFor={field} className="text-small text-muted">{t(`limits.${field}`)}</label>
            <input id={field} name={field} inputMode={field === "maxPositions" || field === "cautionAt" ? "numeric" : "decimal"} value={inputs.values[field]} aria-describedby={hint ? `${field}-hint` : undefined} onChange={(event) => { setInputs((previous) => ({ locale, values: { ...previous.values, [field]: event.target.value } })); setSaved(false); }} className="num min-h-control w-full min-w-0 rounded-control border border-line bg-raised px-3 text-base text-text" />
            {hint && <span id={`${field}-hint`} className="text-xs leading-[17px] text-subtle"><NumericCopy text={hint} /></span>}
          </div>;
        })}
      </div>
    </Card>
  </form>;
}
