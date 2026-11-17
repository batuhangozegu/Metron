import { Card, PnL } from "@/components/metron";
import { formatAmount, formatPercent } from "@/lib/format";
import { getMessages, type Locale } from "@/lib/messages";
import { account, getSampleData, portfolioStatistics as stats } from "@/lib/sample-data";

export function EquityChart({ locale = "en" }: { locale?: Locale }) {
  const t = getMessages("portfolio", locale);
  const { equitySeries: series } = getSampleData(locale);
  // The fixtures contain screen coordinates: smaller y means greater equity.
  let peak = series.account[0];
  let span = { start: peak.x, end: peak.x, fall: 0 };
  for (const point of series.account) {
    if (point.y < peak.y) peak = point;
    if (point.y - peak.y > span.fall) span = { start: peak.x, end: point.x, fall: point.y - peak.y };
  }
  const last = series.account[series.account.length - 1];
  const description = t("chartDescription", {
    initial: formatAmount(stats.initialEquity, 0), equity: formatAmount(stats.equity, 0), currency: account.currency,
    days: stats.periodDays, drawdown: formatAmount(stats.maxDrawdownPercent.replace("-", ""), 1), benchmark: formatAmount(stats.benchmarkEndEquity, 0),
  });
  return <Card className="p-card flex flex-col gap-4">
    <div className="flex flex-wrap items-center justify-between gap-x-6 gap-y-2">
      <h2 className="text-heading">{t("chartTitle")}</h2>
      <div className="flex flex-wrap gap-5 text-small text-muted">
        <span className="flex flex-wrap items-center gap-2"><span aria-hidden="true" className="h-0.5 w-5 bg-text" />{t("yourAccount")}<PnL locale={locale} value={stats.returnPercent} decimals={1} suffix={t("percent")} className="text-text" /></span>
        <span className="flex flex-wrap items-center gap-2"><span aria-hidden="true" className="w-5 border-t-2 border-dashed border-subtle" />{t("benchmark")}<PnL locale={locale} value={stats.benchmarkReturnPercent} decimals={1} suffix={t("percent")} className="text-text" /></span>
      </div>
    </div>
    <svg role="img" aria-labelledby="portfolio-equity-title portfolio-equity-description" viewBox={series.viewBox} className="block h-auto w-full">
      <title id="portfolio-equity-title">{t("chartTitle")}</title>
      <desc id="portfolio-equity-description">{description}</desc>
      <rect x={span.start} y={20} width={span.end - span.start} height={240} className="fill-loss-tint" />
      {series.valueTicks.map(tick => <line key={tick.y} x1={0} y1={tick.y - 4} x2={last.x} y2={tick.y - 4} className="stroke-line" />)}
      <line x1={0} y1={260} x2={last.x} y2={260} className="stroke-subtle" />
      {series.account.map((point, index) => <line key={point.x} x1={point.x} y1={260} x2={point.x} y2={index % 4 === 0 ? 268 : 264} className="stroke-subtle" />)}
      <polyline points={series.benchmark.map(point => `${point.x},${point.y}`).join(" ")} fill="none" strokeWidth={2} strokeDasharray="6 5" className="stroke-subtle" />
      <polyline points={series.account.map(point => `${point.x},${point.y}`).join(" ")} fill="none" strokeWidth={2.5} strokeLinejoin="round" className="stroke-text" />
      <circle cx={last.x} cy={last.y} r={4.5} className="fill-text" />
      <text x={span.start + 8} y={38} className="fill-loss text-[13px] font-medium">{t("drawdownLabel")}<tspan className="num">{formatPercent(stats.maxDrawdownPercent, 1)}</tspan></text>
      {series.valueTicks.map(tick => <text key={tick.y} x={last.x + 12} y={tick.y} className="num fill-subtle text-xs">{tick.label}</text>)}
      {series.dates.map((date, index) => <text key={date.x} x={date.x} y={288} textAnchor={index === 0 ? "start" : index === series.dates.length - 1 ? "end" : "middle"} className="num fill-subtle text-xs">{date.label}</text>)}
    </svg>
  </Card>;
}
