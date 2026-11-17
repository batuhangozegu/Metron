import { Suspense } from "react";
import { ClassTag } from "@/components/metron/class-tag";
import { PnL } from "@/components/metron/pnl";
import { formatPrice } from "@/lib/format";
import { getLocale } from "@/lib/locale-server";
import { getMessages, type Locale } from "@/lib/messages";
import { AutoRefresh } from "./_components/auto-refresh";
import { fetchMarkets, formatVolume } from "./_lib/market-data";
import { FilteredMarkets } from "./_components/filtered-markets";
import { MovementMeasure } from "./_components/movement-measure";

async function LiveMarkets({ locale }: { locale: Locale }) {
  const t = getMessages("markets", locale);
  const markets = await fetchMarkets();
  if (markets === null) return <p role="status" className="rounded-card border border-line bg-surface p-6 text-muted">{t("unavailable")}</p>;
  if (markets.length === 0) return <p role="status" className="rounded-card border border-line bg-surface p-6 text-muted">{t("waiting")}</p>;
  const cell = "border-t border-line px-3 py-1.5";
  const rows = markets.map((market) => {
    const pair = `${market.base_asset}/${market.quote_asset}`;
    return {
      pair,
      coinClass: market.coin_class,
      content: (
        <tr key={market.symbol} className="text-right">
          <th scope="row" className={`${cell} whitespace-nowrap px-6 text-left font-semibold`}>{pair}</th>
          <td className={`${cell} text-left`}><ClassTag coinClass={market.coin_class} locale={locale} /></td>
          <td className={`${cell} num`}>{market.price === null ? t("missing") : formatPrice(market.price)}</td>
          <td className={cell}>{market.change_24h_percent === null ? t("missing") : <PnL value={market.change_24h_percent} suffix={t("percentSuffix")} locale={locale} />}</td>
          <td className={`${cell} text-left`}>{market.atr_percent === null ? t("missing") : <MovementMeasure locale={locale} market={{
            pair, atr: `${market.atr_percent}%`, atrFill: `${Number(market.atr_percent) * 10}%`,
          }} />}</td>
          <td className={`${cell} ${market.coin_class === "stable" ? "" : "num"}`}>
            {market.coin_class === "stable" ? t("stop.none") : market.suggested_stop_percent === null ? t("missing") : `${market.suggested_stop_percent}%`}
          </td>
          <td className={`${cell} num text-muted`}>{market.quote_volume_24h === null ? t("missing") : formatVolume(market.quote_volume_24h)}</td>
          <td className={`${cell} pr-6`}>
            <a href="/check" aria-label={t("checkPair", { pair })}
              className="inline-flex min-h-control items-center whitespace-nowrap rounded-control px-3.5 py-3 text-small font-medium underline hover:bg-raised">
              {t("checkTrade")}
            </a>
          </td>
        </tr>
      ),
    };
  });
  return <>
    <div className="flex items-center gap-2 text-small text-muted">
      <span aria-hidden="true" className={`size-2 rounded-full ${markets.some((market) => market.price !== null) ? "bg-safe" : "bg-muted"}`} />
      {t("live")}
    </div>
    <FilteredMarkets rows={rows} />
  </>;
}

export default async function MarketsPage() {
  const locale = await getLocale();
  const t = getMessages("markets", locale);
  return <>
    <AutoRefresh />
    <section className="flex flex-wrap items-end justify-between gap-x-8 gap-y-4">
      <div className="flex flex-col gap-2">
        <h1 className="text-title">{t("title")}</h1>
        <p className="max-w-[620px] text-body text-muted">{t("description")}</p>
      </div>
    </section>
    <Suspense fallback={<p role="status" className="text-muted">{t("loading")}</p>}>
      <LiveMarkets locale={locale} />
    </Suspense>
  </>;
}
