import { CircleCheck, TriangleAlert } from "lucide-react";
import { Card, ClassTag, Measure, PnL, RiskBadge, Stat, riskState } from "@/components/metron";
import { formatAmount, formatPercent, formatPrice } from "@/lib/format";
import { getMessages } from "@/lib/messages";
import { account, gauges, limits, openPositions } from "@/lib/sample-data";
import { NumberText } from "./_components/number-text";

const t = getMessages("overview");
const gaugeKeys = ["open", "daily", "weekly"] as const;
const columns = ["pair", "entry", "price", "stop", "toStop", "atRisk", "unrealized"] as const;
const checks = ["run", "placed", "blocked", "abandoned"] as const;

export default function Page() {
  const cooldown = account.cooldown;
  return <>
    <section aria-labelledby="overview-title" className="flex flex-wrap items-end justify-between gap-x-8 gap-y-4">
      <div className="flex flex-col gap-2">
        <p className="text-label text-subtle uppercase">{t("status")} <span className="num">{account.asOf}</span></p>
        <div className="flex items-center gap-3.5">
          <CircleCheck aria-hidden="true" className="size-8 shrink-0 text-safe" />
          <h1 id="overview-title" className="text-display">{t("headline")}</h1>
        </div>
        <p className="max-w-[640px] text-body text-muted"><NumberText text={t("summary", { amount: formatAmount(account.totalOpenRisk), currency: account.currency, percent: gauges[0].value })} /></p>
      </div>
      <div className="flex flex-wrap gap-8 [&_.text-figure-medium]:text-xl [&_.text-small]:text-xs">
        <Stat label={t("realized")} figure={<PnL value={account.realizedToday} />} />
        <Stat label={t("unrealized")} figure={<PnL value={account.unrealized} />} />
        <Stat label={t("positions")} figure={<>{formatAmount(String(account.openPositions), 0)} <span className="text-subtle"><NumberText text={t("of", { limit: formatAmount(String(limits.maxPositions), 0) })} /></span></>} />
      </div>
    </section>
    <section aria-label={t("gauges")} className="flex flex-wrap gap-4">
      {gauges.map((gauge, index) => {
        const value = parseFloat(gauge.value), limit = parseFloat(gauge.limit);
        const label = t(`gauge.${gaugeKeys[index]}.label`);
        return <Card key={gaugeKeys[index]} className="flex min-w-0 flex-[1_1_320px] flex-col gap-4 p-card">
          <div className="flex items-center justify-between gap-3"><h2 className="text-sm font-medium text-muted">{label}</h2><RiskBadge state={riskState(value, limit, limits.cautionAt)} /></div>
          <div className="flex flex-wrap items-baseline gap-2.5">
            <span className="num text-figure-large">{formatPercent(gauge.value.replace("%", ""))}</span>
            <span className="text-body text-subtle"><NumberText text={t("of", { limit: formatPercent(gauge.limit.replace("%", "")) })} /></span>
          </div>
          <Measure value={value} limit={limit} cautionAt={limits.cautionAt} label={label} formatValue={(amount) => amount === "0" ? formatAmount(amount, 0) : formatPercent(amount)} />
          <p className="text-small leading-5 text-muted"><span className="num text-text">{gauge.amount}</span> {t(`gauge.${gaugeKeys[index]}.note`)}</p>
        </Card>;
      })}
    </section>
    <section aria-label={t("details")} className="flex min-w-0 flex-wrap items-start gap-4">
      <Card aria-labelledby="positions-title" className="flex min-w-0 flex-[999_1_640px] flex-col overflow-hidden">
        <div className="flex flex-wrap items-center justify-between gap-3 px-6 py-5"><h2 id="positions-title" className="text-heading">{t("positions")}</h2><span className="text-small text-subtle">{t("sorted")}</span></div>
        <div role="region" aria-label={t("positions.table")} tabIndex={0} className="min-w-0 overflow-x-auto">
          <table className="w-full min-w-[760px] border-collapse text-right text-sm">
            <thead className="text-xs text-subtle"><tr>{columns.map((column) => <th key={column} scope="col" className={`border-t border-line px-3 py-2.5 font-medium first:px-6 last:pr-6 ${column === "pair" || column === "toStop" ? "text-left" : ""}`}>{t(`column.${column}`)}</th>)}</tr></thead>
            <tbody>{openPositions.map((position) => <tr key={position.pair} className="[&>td]:border-t [&>td]:border-line [&>td]:px-3 [&>td]:py-3.5">
              <th scope="row" className="border-t border-line px-6 py-3.5 text-left font-normal"><div className="flex flex-col gap-0.5">
                <span className="font-semibold">{position.pair}</span>
                <div className="flex items-center gap-1 text-xs text-subtle"><ClassTag coinClass={position.coinClass} className="rounded-none border-0 p-0 text-subtle" /><span aria-hidden="true">{t("separator")}</span><span className="num whitespace-nowrap">{position.qty}</span></div>
              </div></th>
              <td className="num text-muted">{formatPrice(position.entry)}</td><td className="num">{formatPrice(position.price)}</td><td className="num text-muted">{formatPrice(position.stop)}</td>
              <td className="text-left"><div className="flex items-center gap-2.5"><div aria-hidden="true" className="relative h-1.5 w-[72px] shrink-0 rounded-[2px] bg-raised"><span className="absolute inset-y-0 left-0 rounded-[2px] bg-safe" style={{ width: position.stopFill }} /></div><span className="num text-small">{formatPercent(position.toStop.replace("%", ""))}</span></div></td>
              <td className="num">{formatAmount(position.risk)}</td><td className="num !pr-6"><PnL value={position.plAmount} /></td>
            </tr>)}</tbody>
            <tfoot><tr className="font-medium [&>*]:border-t [&>*]:border-line"><th scope="row" colSpan={5} className="px-6 py-3.5 text-left font-medium text-muted">{t("totalRisk")}</th><td className="num px-3 py-3.5">{formatAmount(account.totalOpenRisk)}</td><td className="num py-3.5 pr-6 pl-3"><PnL value={account.unrealized} /></td></tr></tfoot>
          </table>
        </div>
        <p className="border-t border-line px-6 py-3.5 text-small text-subtle">{t("stopsNote")}</p>
      </Card>
      <aside aria-label={t("discipline")} className="flex min-w-0 flex-[1_1_320px] flex-col gap-4">
        <Card className="flex flex-col gap-4 p-card">
          <div className="flex items-center justify-between gap-3"><h2 className="text-heading">{t("discipline")}</h2><a href="/rules" className="inline-flex min-h-control items-center py-3 text-small text-muted underline">{t("editRules")}</a></div>
          <div className="flex items-start gap-3 rounded-[10px] bg-caution-tint p-3.5">
            <TriangleAlert role="img" aria-label={t("caution")} className="mt-px size-5 shrink-0 text-caution" />
            <div className="flex min-w-0 flex-col gap-1"><h3 className="text-sm font-medium">{t("cooldown.title")}</h3><p className="text-small text-muted"><NumberText text={t("cooldown.body", { pair: cooldown.pair, closedAt: cooldown.closedAt, waitUntil: cooldown.waitUntil, minutes: cooldown.minutesRemaining })} /></p></div>
          </div>
          <ul className="flex flex-col gap-3 text-sm">{(["hours", "stop"] as const).map((rule) => <li key={rule} className="flex items-center gap-2.5"><CircleCheck role="img" aria-label={t("kept")} className="size-4 shrink-0 text-safe" /><span><NumberText text={t(`rule.${rule}`)} /></span></li>)}</ul>
          <p className="border-t border-line pt-3.5 text-small text-subtle"><NumberText text={t("violations", { count: cooldown.violationsThisWeek })} /></p>
        </Card>
        <Card className="flex flex-col gap-3.5 p-card"><h2 className="text-heading">{t("checks")}</h2><dl className="grid grid-cols-[1fr_auto] gap-x-4 gap-y-2.5 text-sm">{checks.map((check) => <div key={check} className="contents"><dt className="text-muted">{t(`checks.${check}`)}</dt><dd className="num text-right">{formatAmount(String(account.checksToday[check]), 0)}</dd></div>)}</dl></Card>
      </aside>
    </section>
  </>;
}
