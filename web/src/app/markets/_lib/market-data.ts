import type { CoinClass } from "@/components/metron/class-tag";
import { compareDecimals, decimalProduct, formatAmount } from "@/lib/format";

export interface Market {
  symbol: string;
  base_asset: string;
  quote_asset: string;
  coin_class: CoinClass;
  price: string | null;
  change_24h_percent: string | null;
  quote_volume_24h: string | null;
  price_time: string | null;
  atr_percent: string | null;
  atr: string | null;
  suggested_stop_percent: string | null;
}

function isMarket(value: unknown): value is Market {
  if (typeof value !== "object" || value === null) return false;
  const row = value as Record<string, unknown>;
  return ["symbol", "base_asset", "quote_asset"].every((key) => typeof row[key] === "string")
    && ["core", "large", "small", "stable"].includes(String(row.coin_class))
    && ["price", "change_24h_percent", "quote_volume_24h", "atr_percent", "atr", "suggested_stop_percent"]
      .every((key) => row[key] === null || (typeof row[key] === "string" && /^-?\d+(?:\.\d+)?$/.test(row[key])))
    && (row.price_time === null || typeof row.price_time === "string");
}

export async function fetchMarkets(): Promise<Market[] | null> {
  try {
    const base = (process.env.METRON_API_URL || "http://127.0.0.1:8080").replace(/\/$/, "");
    const response = await fetch(`${base}/v1/markets`, {
      cache: "no-store",
      signal: AbortSignal.timeout(3_000),
    });
    if (!response.ok) return null;
    const data: unknown = await response.json();
    if (typeof data !== "object" || data === null || !("markets" in data)
      || !Array.isArray(data.markets) || !data.markets.every(isMarket)) return null;
    return data.markets;
  } catch {
    return null;
  }
}

/** Compact decimal strings exactly; money never passes through a JS Number. */
export function formatVolume(value: string): string {
  if (compareDecimals(value, "1000000000") >= 0) return `${formatAmount(decimalProduct(value, "0.000000001"), 1).replace(/\.0$/, "")}B`;
  if (compareDecimals(value, "1000000") >= 0) return `${formatAmount(decimalProduct(value, "0.000001"), 0)}M`;
  if (compareDecimals(value, "1000") >= 0) return `${formatAmount(decimalProduct(value, "0.001"), 1).replace(/\.0$/, "")}K`;
  return formatAmount(value, 0);
}
