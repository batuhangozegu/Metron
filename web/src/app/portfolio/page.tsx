import { getLocale } from "@/lib/locale-server";
import { Card, CardHeader, ClassTag, PnL, Stat } from "@/components/metron";
import { formatAmount, formatPercent, formatPrice } from "@/lib/format";
import { getMessages } from "@/lib/messages";
import { account, classStatistics, holdings, getSampleData, portfolioStatistics as stats } from "@/lib/sample-data";
import { EquityChart } from "./_components/equity-chart";
import { PeriodSelector } from "./_components/period-selector";
import { NumberText } from "./_components/number-text";

const table = "w-full border-collapse whitespace-nowrap text-sm text-right [&_th]:border-t [&_th]:border-line [&_th]:px-3 [&_th]:py-2.5 [&_th]:font-medium [&_tbody_th]:py-3.5 [&_td]:border-t [&_td]:border-line [&_td]:px-3 [&_td]:py-3.5 [&_th:first-child]:px-6 [&_td:first-child]:px-6 [&_th:last-child]:pr-6 [&_td:last-child]:pr-6";
const footnote = "border-t border-line px-6 py-3.5 text-small text-subtle";

export default async function PortfolioPage() {
  const locale = await getLocale();
  const { journal } = getSampleData(locale);
  const t = getMessages("portfolio", locale);
  return <>
    <section className="flex flex-wrap items-center justify-between gap-x-8 gap-y-4">
      <h1 className="text-title">{t("title")}</h1>
      <PeriodSelector days={stats.periodDays} />
    </section>
    <section aria-label={t("summary")} className="flex flex-wrap rounded-card border border-line bg-surface py-2 [&>div]:flex-[1_1_200px] [&>div]:gap-1.5 [&>div]:px-6 [&>div]:py-4">
      <Stat label={t("accountValue")} figure={formatAmount(stats.equity)} note={<><PnL locale={locale} value={stats.returnPercent} decimals={1} suffix={t("percent")} />{t("separator")}<PnL locale={locale} value={stats.netResult} /></>} />
      <Stat label={t("maxDrawdown")} figure={formatPercent(stats.maxDrawdownPercent, 1)} note={<span className="num">{t("drawdownRange", { peak: formatAmount(stats.drawdownPeak, 0), trough: formatAmount(stats.drawdownTrough, 0) })}</span>} />
      <Stat label={t("winRate")} figure={formatPercent(stats.winRate, 1)} note={<NumberText text={t("tradeCount", { wins: stats.wins, trades: stats.trades })} />} />
      <Stat label={t("averageWinLoss")} figure={t("multiple", { value: formatAmount(stats.averageWinLoss) })} note={<><PnL locale={locale} value={stats.averageWin} />{t("slash")}<PnL locale={locale} value={stats.averageLoss} /></>} />
      <Stat label={t("feesPaid")} figure={formatAmount(stats.feesPaid)} note={t("feesIncluded")} />
    </section>
    <EquityChart locale={locale} />
    <section aria-label={t("breakdown")} className="flex flex-wrap items-start gap-4 [&>article]:flex-[1_1_440px]">
      <Card>
        <CardHeader title={t("classTitle")} />
        <div role="region" aria-label={t("classTitle")} tabIndex={0} className="overflow-x-auto">
          <table className={`${table} min-w-[420px]`}>
            <thead className="text-xs text-subtle"><tr>
              <th scope="col" className="text-left">{t("class")}</th><th scope="col">{t("trades")}</th><th scope="col">{t("winRate")}</th><th scope="col">{t("netResult")}</th>
            </tr></thead>
            <tbody>{classStatistics.map(row => <tr key={row.coinClass}>
              <th scope="row" className="text-left"><ClassTag locale={locale} coinClass={row.coinClass} className="rounded-none border-0 p-0 text-sm font-medium text-text" /></th>
              <td className="num">{formatAmount(row.trades, 0)}</td><td className="num">{formatPercent(row.win.replace("%", ""), 1)}</td><td><PnL locale={locale} value={row.netAmount} /></td>
            </tr>)}</tbody>
          </table>
        </div>
        <p className={footnote}>{t("classNote")}</p>
      </Card>
      <Card>
        <CardHeader title={t("holdingsTitle")} />
        <div role="region" aria-label={t("holdingsTitle")} tabIndex={0} className="overflow-x-auto">
          <table className={`${table} min-w-[420px]`}>
            <thead className="text-xs text-subtle"><tr>
              <th scope="col" className="text-left">{t("asset")}</th><th scope="col">{t("quantity")}</th><th scope="col">{t("averageCost")}</th><th scope="col">{t("fees")}</th>
            </tr></thead>
            <tbody>{holdings.map(row => <tr key={row.asset}>
              <th scope="row" className="text-left">{row.asset}</th><td className="num">{formatPrice(row.qty)}</td><td className="num">{formatAmount(row.cost)}</td><td className="num text-muted">{formatAmount(row.fees)}</td>
            </tr>)}</tbody>
          </table>
        </div>
        <p className={footnote}>{t("holdingsNote")}</p>
      </Card>
    </section>
    <Card>
      <CardHeader className="items-baseline" title={t("journalTitle")} action={<span className="text-small text-subtle"><NumberText text={t("journalNote")} /></span>} />
      <div role="region" aria-label={t("journalTitle")} tabIndex={0} className="overflow-x-auto">
        <table className={`${table} min-w-[760px]`}>
          <thead className="text-xs text-subtle"><tr>
            <th scope="col" className="text-left">{t("pair")}</th><th scope="col" className="text-left">{t("closed")}</th><th scope="col" className="text-left">{t("howClosed")}</th><th scope="col">{t("result")}</th><th scope="col">{account.currency}</th><th scope="col" className="text-left">{t("rules")}</th>
          </tr></thead>
          <tbody>{journal.map(row => <tr key={`${row.pair}-${row.closed}`}>
            <th scope="row" className="text-left font-semibold">{row.pair}</th>
            <td className="num text-left text-muted">{row.closed}</td>
            <td className="text-left text-muted">{row.how}</td>
            <td><PnL locale={locale} value={row.r.replace("R", "")} decimals={1} suffix={t("rUnit")} /></td>
            <td><PnL locale={locale} value={row.plAmount} /></td>
            <td className={`text-left ${row.ruleColor === "caution" ? "text-caution" : "text-muted"}`}>
              {row.ruleColor === "caution" && <svg aria-hidden="true" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={1.75} className="mr-1.5 inline-block size-4 align-text-bottom"><path d="M12 3 2 21h20L12 3Z" /><path d="M12 9v5m0 3v1" /></svg>}
              <NumberText text={row.rules} />
            </td>
          </tr>)}</tbody>
        </table>
      </div>
    </Card>
  </>;
}
