import { getLocale } from "@/lib/locale-server";
import { preTradeBlocked, preTradeCaution } from "@/lib/sample-data";
import { CheckScreen } from "./_components/check-screen";
import { CheckResult } from "./_components/check-result";

export default async function Page() {
  const locale = await getLocale();
  return <CheckScreen caution={<CheckResult locale={locale} check={preTradeCaution} />} blocked={<CheckResult locale={locale} check={preTradeBlocked} />} />;
}
