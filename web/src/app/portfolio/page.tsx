import { getLocale } from "@/lib/locale-server";
import { getMessages } from "@/lib/messages";
import { parsePortfolioPeriod } from "@/lib/portfolio";
import { PortfolioDashboard } from "./_components/portfolio-dashboard";

export default async function PortfolioPage({ searchParams }: { searchParams: Promise<{ account?: string | string[]; period?: string | string[] }> }) {
  const locale = await getLocale();
  const t = getMessages("portfolio", locale);
  const { account, period } = await searchParams;
  const accountId = typeof account === "string" ? account : undefined;
  return <>
    <h1 className="text-title">{t("title")}</h1>
    <PortfolioDashboard key={accountId ?? "default"} initialAccountId={accountId} period={parsePortfolioPeriod(period)} />
  </>;
}
