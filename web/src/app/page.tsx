import { OverviewDashboard } from "./_components/overview-dashboard";

export default async function Page({ searchParams }: { searchParams: Promise<{ account?: string | string[] }> }) {
  const { account } = await searchParams;
  const accountId = typeof account === "string" ? account : undefined;
  return <OverviewDashboard key={accountId ?? "default"} initialAccountId={accountId} />;
}
