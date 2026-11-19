import { getLocale } from "@/lib/locale-server";
import { getMessages } from "@/lib/messages";
import { EquityChart } from "./_components/equity-chart";
import { Positions } from "./_components/positions";

export default async function PortfolioPage({ searchParams }: { searchParams: Promise<{ account?: string | string[] }> }) {
  const locale = await getLocale();
  const t = getMessages("portfolio", locale);
  const { account } = await searchParams;
  const accountId = typeof account === "string" ? account : undefined;
  return <>
    <h1 className="text-title">{t("title")}</h1>
    <Positions key={accountId ?? "default"} initialAccountId={accountId} />
    <EquityChart locale={locale} />
  </>;
}
