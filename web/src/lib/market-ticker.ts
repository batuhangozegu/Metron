import type { Market } from "@/app/markets/_lib/market-data";
import { markets } from "@/lib/sample-data";

export interface TickerQuote {
  pair: string;
  price: string | null;
  change: string | null;
}

export interface TickerSnapshot {
  source: "live" | "sample";
  quotes: TickerQuote[];
}

export const sampleTicker: TickerSnapshot = {
  source: "sample",
  quotes: markets.map(row => ({
    pair: row.pair,
    price: row.price.replaceAll(",", ""),
    change: row.changePercent,
  })),
};

/** An empty feed is not live; never silently mix sample and live prices. */
export function tickerSnapshot(rows: Market[] | null): TickerSnapshot {
  if (!rows?.some(row => row.price !== null)) return sampleTicker;
  return {
    source: "live",
    quotes: rows.map(row => ({
      pair: `${row.base_asset}/${row.quote_asset}`,
      price: row.price,
      change: row.change_24h_percent,
    })),
  };
}
