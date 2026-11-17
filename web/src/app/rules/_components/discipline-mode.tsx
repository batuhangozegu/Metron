import { Octagon, TriangleAlert } from "lucide-react";
import type { DisciplineRule } from "@/lib/sample-data";
import { getMessages, type Locale } from "@/lib/messages";

/** Discipline actions have different wording from generic risk verdicts. */
export function DisciplineMode({ rule, locale = "en" }: { rule: DisciplineRule; locale?: Locale }) {
  const t = getMessages("rules", locale);
  const blocking = rule.modeColor === "limit";
  const Icon = blocking ? Octagon : TriangleAlert;
  return <span className={`inline-flex flex-[0_0_150px] items-center justify-center gap-1.5 rounded-tag px-2.5 py-[5px] text-small font-medium ${blocking ? "bg-limit-tint text-limit" : "bg-caution-tint text-caution"}`}>
    <Icon aria-hidden="true" className="size-4 shrink-0" />
    {t(blocking ? "discipline.block" : "discipline.warn")}
  </span>;
}
