import { Card, CardHeader } from "@/components/metron/card";
import { ClassTag } from "@/components/metron/class-tag";
import { formatAmount, formatPercent } from "@/lib/format";
import { getLocale } from "@/lib/locale-server";
import { getMessages } from "@/lib/messages";
import { AccountLimits } from "./_components/account-limits";

/** Server-side system defaults from docs/RISK.md; not editable per account yet. */
const classRules = [
  { coinClass: "core", multiplier: "1.5", risk: "2", allocation: "40" },
  { coinClass: "large", multiplier: "1.5", risk: "1", allocation: "15" },
  { coinClass: "small", multiplier: "2", risk: "0.5", allocation: "8" },
  { coinClass: "stable", multiplier: null, risk: null, allocation: null },
] as const;

export default async function RulesPage() {
  const locale = await getLocale();
  const t = getMessages("rules", locale);
  const columns = ["class", "multiplier", "risk", "position"] as const;

  return <>
    <AccountLimits />
    <Card aria-labelledby="class-rules-title">
      <CardHeader title={<span id="class-rules-title">{t("classes.title")}</span>} note={t("classes.description")} />
      <div className="overflow-x-auto rounded-b-card" role="region" aria-label={t("classes.scrollLabel")} tabIndex={0}>
        <table className="w-full min-w-[640px] border-collapse text-right text-sm">
          <thead>
            <tr className="text-xs text-subtle">
              {columns.map((column) => <th key={column} scope="col" className={`border-t border-line py-2.5 font-medium ${column === "class" ? "px-6 text-left" : column === "position" ? "pr-6 pl-3" : "px-3"}`}>{t(`classes.${column}`)}</th>)}
            </tr>
          </thead>
          <tbody>
            {classRules.map((rule) => <tr key={rule.coinClass} className="[&>td]:border-t [&>td]:border-line [&>td]:px-3 [&>td]:py-2.5">
              <th scope="row" className="border-t border-line px-6 py-2.5 text-left font-normal">
                <div className="flex flex-col items-start gap-0.5">
                  <ClassTag locale={locale} coinClass={rule.coinClass} className="rounded-none border-0 p-0 text-sm font-semibold text-text" />
                  <span className="text-xs text-subtle">{t(`classes.${rule.coinClass}.who`)}</span>
                </div>
              </th>
              <td className="num">{rule.multiplier ? `${formatAmount(rule.multiplier, rule.multiplier.includes(".") ? 1 : 0, locale)}×` : t("none")}</td>
              <td className="num">{rule.risk ? formatPercent(rule.risk, rule.risk.includes(".") ? 1 : 0, locale) : t("classes.stable.risk")}</td>
              <td className="num !pr-6">{rule.allocation ? formatPercent(rule.allocation, 0, locale) : t("noCap")}</td>
            </tr>)}
          </tbody>
        </table>
      </div>
    </Card>
  </>;
}
