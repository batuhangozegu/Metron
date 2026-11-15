import { preTradeBlocked, preTradeCaution } from "@/lib/sample-data";
import { CheckScreen } from "./_components/check-screen";
import { CheckResult } from "./_components/check-result";

export default function Page() {
  return <CheckScreen caution={<CheckResult check={preTradeCaution} />} blocked={<CheckResult check={preTradeBlocked} />} />;
}
