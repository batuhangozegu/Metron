import { Card, PnL } from "@/components/metron";
import { formatAmount, formatPercent } from "@/lib/format";
import { getMessages, type Locale } from "@/lib/messages";
import type { EquityPoint, Portfolio } from "@/lib/portfolio";

/** Numeric conversion is confined to SVG geometry; displayed API amounts stay exact. */
export function equityGeometry(account: EquityPoint[], benchmark: EquityPoint[]) {
  const points = [...account, ...benchmark];
  const times = points.map(point => Date.parse(point.at));
  const values = points.map(point => Number(point.equity));
  const start = Math.min(...times), end = Math.max(...times);
  const minimum = Math.min(...values), maximum = Math.max(...values);
  const padding = (maximum - minimum || Math.max(Math.abs(maximum), 1) * 0.02) * 0.1;
  const low = minimum - padding, high = maximum + padding;
  const x = (at: string | number) => 12 + ((typeof at === "number" ? at : Date.parse(at)) - start) / (end - start || 1) * 888;
  const y = (equity: string | number) => 250 - (Number(equity) - low) / (high - low) * 226;
  const line = (series: EquityPoint[]) => series.map(point => `${x(point.at)},${y(point.equity)}`).join(" ");
  return { start, end, low, high, x, y, line };
}

export function EquityChart({ data, locale }: { data: Portfolio; locale: Locale }) {
  const t = getMessages("portfolio", locale);
  const series = data.equity_series;
  if (series.length < 2) return <Card className="p-card"><h2 className="text-heading">{t("chartTitle")}</h2><p className="mt-4 text-muted">{t("notEnoughHistory")}</p></Card>;
  const g = equityGeometry(series, data.benchmark_series);
  const last = series[series.length - 1];
  const drawdown = data.max_drawdown;
  const date = new Intl.DateTimeFormat(locale === "tr" ? "tr-TR" : "en-GB", { day: "numeric", month: "short", ...(g.end - g.start >= 31_536_000_000 ? { year: "numeric" } as const : {}), ...(g.end - g.start < 172_800_000 ? { hour: "2-digit", minute: "2-digit" } as const : {}), timeZone: "UTC" });
  const description = t("chartDescription", { initial: formatAmount(series[0].equity, 2, locale), equity: formatAmount(last.equity, 2, locale), currency: data.currency, from: date.format(new Date(series[0].at)), to: date.format(new Date(last.at)) });
  return <Card className="p-card flex flex-col gap-4">
    <div className="flex flex-wrap items-center justify-between gap-x-6 gap-y-2">
      <h2 className="text-heading">{t("chartTitle")}</h2>
      <div className="flex flex-wrap gap-5 text-small text-muted">
        <span className="flex flex-wrap items-center gap-2"><span aria-hidden="true" className="h-0.5 w-5 bg-text" />{t("yourAccount")}<PnL locale={locale} value={data.summary.return_percent} decimals={1} suffix="%" /></span>
        <span className="flex flex-wrap items-center gap-2"><span aria-hidden="true" className="w-5 border-t-2 border-dashed border-subtle" />{t("benchmark")}{data.benchmark_return_percent === null ? t("nullValue") : <PnL locale={locale} value={data.benchmark_return_percent} decimals={1} suffix="%" />}</span>
      </div>
    </div>
    <div role="region" aria-label={t("chartTitle")} tabIndex={0} className="overflow-x-auto">
      <svg role="img" aria-labelledby="portfolio-equity-title portfolio-equity-description" viewBox="0 0 1060 300" className="block h-auto w-full min-w-[600px]">
        <title id="portfolio-equity-title">{t("chartTitle")}</title>
        <desc id="portfolio-equity-description">{description}{drawdown && ` ${t("maxDrawdown")}: ${formatPercent(drawdown.percent, 1, locale)}.`}</desc>
        {drawdown && <rect x={g.x(drawdown.peak_at)} y={20} width={Math.max(0, g.x(drawdown.trough_at) - g.x(drawdown.peak_at))} height={230} className="fill-loss-tint" />}
        {[0, 1, 2, 3].map(index => {
          const value = g.low + (g.high - g.low) * index / 3;
          return <g key={index}><line x1={12} y1={g.y(value)} x2={900} y2={g.y(value)} className="stroke-line" /><text x={914} y={g.y(value) + 4} className="num fill-subtle text-xs">{formatAmount(String(value), 2, locale)}</text></g>;
        })}
        <line x1={12} y1={250} x2={900} y2={250} className="stroke-subtle" />
        {data.benchmark_series.length >= 2 && <polyline points={g.line(data.benchmark_series)} fill="none" strokeWidth={2} strokeDasharray="6 5" className="stroke-subtle" />}
        <polyline points={g.line(series)} fill="none" strokeWidth={2.5} strokeLinejoin="round" className="stroke-text" />
        <circle cx={g.x(last.at)} cy={g.y(last.equity)} r={4.5} className="fill-text" />
        {drawdown && <text x={Math.min(g.x(drawdown.peak_at) + 8, 650)} y={16} className="fill-loss text-[13px] font-medium">{t("drawdownLabel")}<tspan className="num">{formatPercent(drawdown.percent, 1, locale)}</tspan></text>}
        {[0, 1, 2, 3].map(index => {
          const at = g.start + (g.end - g.start) * index / 3;
          return <g key={index}><line x1={g.x(at)} y1={250} x2={g.x(at)} y2={258} className="stroke-subtle" /><text x={g.x(at)} y={286} textAnchor={index === 0 ? "start" : index === 3 ? "end" : "middle"} className="num fill-subtle text-xs">{date.format(new Date(at))}</text></g>;
        })}
      </svg>
    </div>
  </Card>;
}
