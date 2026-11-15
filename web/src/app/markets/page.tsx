import { ClassTag } from "@/components/metron/class-tag";
import { PnL } from "@/components/metron/pnl";
import { formatPrice } from "@/lib/format";
import { getMessages } from "@/lib/messages";
import { markets } from "@/lib/sample-data";
import { FilteredMarkets } from "./_components/filtered-markets";
import { MovementMeasure } from "./_components/movement-measure";

export default function MarketsPage() {
  const t = getMessages("markets");
  const cell = "border-t border-line px-3 py-1.5";
  const rows = markets.map((market) => ({
    pair: market.pair,
    coinClass: market.coinClass,
    content: (
      <tr key={market.pair} className="text-right">
        <th scope="row" className={`${cell} whitespace-nowrap px-6 text-left font-semibold`}>{market.pair}</th>
        <td className={`${cell} text-left`}><ClassTag coinClass={market.coinClass} /></td>
        <td className={`${cell} num`}>{formatPrice(market.price)}</td>
        <td className={cell}><PnL value={market.changePercent} suffix={t("percentSuffix")} /></td>
        <td className={`${cell} text-left`}><MovementMeasure market={market} /></td>
        <td className={`${cell} ${market.coinClass === "stable" ? "" : "num"}`}>
          {market.coinClass === "stable" ? t("stop.none") : market.stop}
        </td>
        <td className={`${cell} num text-muted`}>{market.vol}</td>
        <td className={`${cell} pr-6`}>
          <a href="/check" aria-label={t("checkPair", { pair: market.pair })}
            className="inline-flex min-h-control items-center whitespace-nowrap rounded-control px-3.5 py-3 text-small font-medium underline hover:bg-raised">
            {t("checkTrade")}
          </a>
        </td>
      </tr>
    ),
  }));
  return <>
    <section className="flex flex-wrap items-end justify-between gap-x-8 gap-y-4">
      <div className="flex flex-col gap-2">
        <h1 className="text-title">{t("title")}</h1>
        <p className="max-w-[620px] text-body text-muted">{t("description")}</p>
      </div>
      <div className="flex items-center gap-2 text-small text-muted">
        <span aria-hidden="true" className="size-2 rounded-full bg-safe" />
        {t("live")}
      </div>
    </section>
    <FilteredMarkets rows={rows} />
  </>;
}
