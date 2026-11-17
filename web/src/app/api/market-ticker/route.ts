import { fetchMarkets } from "@/app/markets/_lib/market-data";
import { tickerSnapshot } from "@/lib/market-ticker";

export async function GET() {
  return Response.json(tickerSnapshot(await fetchMarkets()), {
    headers: { "Cache-Control": "no-store" },
  });
}
