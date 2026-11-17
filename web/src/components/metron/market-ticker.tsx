"use client";

import { useEffect, useState } from "react";
import { Pause, Play } from "lucide-react";
import { formatPrice } from "@/lib/format";
import { getMessages } from "@/lib/messages";
import { sampleTicker, type TickerSnapshot } from "@/lib/market-ticker";
import { useLocale } from "./locale-provider";
import { PnL } from "./pnl";

/** Hydrates with samples; network work is independent of page rendering. */
export function MarketTicker() {
  const locale = useLocale();
  const t = getMessages("common", locale);
  const [snapshot, setSnapshot] = useState(sampleTicker);
  const [paused, setPaused] = useState(false);

  useEffect(() => {
    let disposed = false;
    let timer: ReturnType<typeof setTimeout>;
    let controller: AbortController;
    async function refresh() {
      controller = new AbortController();
      const timeout = setTimeout(() => controller.abort(), 5_000);
      try {
        const response = await fetch("/api/market-ticker", { cache: "no-store", signal: controller.signal });
        if (!response.ok) throw new Error("Ticker unavailable");
        const data: TickerSnapshot = await response.json();
        if (!disposed) setSnapshot(data);
      } catch {
        if (!disposed) setSnapshot(sampleTicker);
      } finally {
        clearTimeout(timeout);
        if (!disposed) timer = setTimeout(refresh, 30_000);
      }
    }
    void refresh();
    return () => { disposed = true; clearTimeout(timer); controller?.abort(); };
  }, []);

  return <section className="market-ticker" aria-label={t("ticker.label")} data-paused={paused}>
    <div className="ticker-controls">
      <span className="ticker-source micro-label" role="status" data-source={snapshot.source}>
        <span aria-hidden="true" className="ticker-beacon" />{t(`ticker.${snapshot.source}`)}
      </span>
      <button type="button" className="ticker-pause" onClick={() => setPaused(value => !value)} aria-label={t(paused ? "ticker.resume" : "ticker.pause")} aria-pressed={paused}>
        {paused ? <Play aria-hidden="true" className="size-3.5" /> : <Pause aria-hidden="true" className="size-3.5" />}
      </button>
    </div>
    <div className="ticker-viewport" tabIndex={0} aria-label={t("ticker.quotes")}>
      <div className="ticker-track">
        {[false, true].map(duplicate => <ul key={String(duplicate)} className={duplicate ? "ticker-quotes ticker-duplicate" : "ticker-quotes"} aria-hidden={duplicate || undefined}>
          {snapshot.quotes.map(quote => <li key={quote.pair}>
            <span className="ticker-pair">{quote.pair}</span>
            <span className="num">{quote.price === null ? t("ticker.missing") : formatPrice(quote.price, undefined, locale)}</span>
            <span className="sr-only">{t("ticker.change")}</span>
            {quote.change === null ? <span className="text-muted">{t("ticker.missing")}</span> : <PnL locale={locale} value={quote.change} suffix="%" />}
          </li>)}
        </ul>)}
      </div>
    </div>
  </section>;
}
