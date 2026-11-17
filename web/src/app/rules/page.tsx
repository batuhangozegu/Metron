import { getLocale } from "@/lib/locale-server";
import { Card, CardHeader } from "@/components/metron/card";
import { ClassTag } from "@/components/metron/class-tag";
import { getMessages } from "@/lib/messages";
import { classRules, disciplineRules, limits } from "@/lib/sample-data";
import { AccountLimits } from "./_components/account-limits";
import { DisciplineMode } from "./_components/discipline-mode";
import { NumericCopy } from "./_components/numeric-copy";

export default async function RulesPage() {
  const locale = await getLocale();
  const t = getMessages("rules", locale);

  const columns = ["class", "multiplier", "risk", "position", "total", "source", "action"] as const;

  return <>
    <AccountLimits initialLimits={limits} />
    <Card aria-labelledby="class-rules-title">
      <CardHeader title={<span id="class-rules-title">{t("classes.title")}</span>} note={t("classes.description")} />
      <div className="overflow-x-auto rounded-b-card" role="region" aria-label={t("classes.scrollLabel")} tabIndex={0}>
        <table className="w-full min-w-[820px] border-collapse text-right text-sm">
          <thead>
            <tr className="text-xs text-subtle">
              {columns.map((column) => <th key={column} scope="col" className={`border-t border-line py-2.5 font-medium ${column === "class" ? "px-6 text-left" : column === "source" ? "px-3 text-left" : column === "action" ? "pr-6 pl-3" : "px-3"}`}>
                <span className={column === "action" ? "sr-only" : undefined}>{t(`classes.${column}`)}</span>
              </th>)}
            </tr>
          </thead>
          <tbody>
            {classRules.map((rule) => <tr key={rule.coinClass} className="[&>td]:border-t [&>td]:border-line [&>td]:px-3 [&>td]:py-1.5">
              <th scope="row" className="border-t border-line px-6 py-1.5 text-left font-normal">
                <div className="flex flex-col items-start gap-0.5">
                  <ClassTag locale={locale} coinClass={rule.coinClass} className="rounded-none border-0 p-0 text-sm font-semibold text-text" />
                  <span className="text-xs text-subtle"><NumericCopy text={t(`classes.${rule.coinClass}.who`, { count: rule.who.match(/\d+/)?.[0] ?? "" })} /></span>
                </div>
              </th>
              <td><NumericCopy text={rule.coinClass === "stable" ? t("none") : rule.mult} /></td>
              <td><NumericCopy text={rule.coinClass === "stable" ? t("none") : rule.risk} /></td>
              <td><NumericCopy text={rule.coinClass === "stable" ? t("noCap") : rule.maxPos} /></td>
              <td><NumericCopy text={rule.coinClass === "stable" || rule.coinClass === "core" ? t("noCap") : rule.total} /></td>
              <td className="text-left text-small text-muted">{t(`classes.${rule.coinClass}.source`)}</td>
              <td className="!pr-6"><button type="button" className="min-h-control px-3.5 text-small font-medium underline" aria-label={t("classes.editLabel", { name: getMessages("common", locale)(`class.${rule.coinClass}`) })}>{t("edit")}</button></td>
            </tr>)}
          </tbody>
        </table>
      </div>
    </Card>
    <Card aria-labelledby="discipline-title">
      <CardHeader title={<span id="discipline-title">{t("discipline.title")}</span>} note={t("discipline.description")} action={<button type="button" className="min-h-control rounded-control border border-line px-4 text-sm font-medium">{t("discipline.add")}</button>} />
      <ul>
        {disciplineRules.map((rule, index) => <li key={rule.name} className="flex flex-wrap items-center gap-x-6 gap-y-3 border-t border-line px-6 py-4">
          <div className="flex min-w-0 flex-[1_1_320px] flex-col gap-[3px]">
            <span className="text-sm font-medium"><NumericCopy text={disciplineCopy(locale, index, "name", rule.name)} /></span>
            <span className="text-small text-subtle"><NumericCopy text={disciplineCopy(locale, index, "detail", rule.detail)} /></span>
          </div>
          <DisciplineMode locale={locale} rule={rule} />
          <span className="flex-[0_0_190px] text-small text-muted"><NumericCopy text={disciplineCopy(locale, index, "history", rule.history)} /></span>
          <button type="button" className="min-h-control px-3.5 text-small font-medium underline" aria-label={t("discipline.editLabel", { name: disciplineCopy(locale, index, "name", rule.name) })}>{t("edit")}</button>
        </li>)}
      </ul>
      <p className="border-t border-line px-6 py-3.5 text-small text-subtle">{t("discipline.footer")}</p>
    </Card>
  </>;
}

/** Numeric values stay in the shared fixtures; sentence templates belong to this screen. */
function disciplineCopy(locale: import("@/lib/messages").Locale, index: number, field: "name" | "detail" | "history", source: string) {
  const keys = ["cooldown", "hours", "stop"] as const;
  const values = Object.fromEntries((source.match(/\d+(?::\d+)*/g) ?? []).map((number, position) => [`n${position}`, number]));
  return getMessages("rules", locale)(`discipline.${keys[index]}.${field}`, values);
}
