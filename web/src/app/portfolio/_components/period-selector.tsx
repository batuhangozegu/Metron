"use client";

import { useState } from "react";
import { getMessages } from "@/lib/messages";
import { NumberText } from "./number-text";

const t = getMessages("portfolio");

export function PeriodSelector({ days }: { days: number }) {
  const [selected, setSelected] = useState("available");
  const options = [
    { id: "month", label: t("month") },
    { id: "available", label: t("days", { days }) },
    { id: "all", label: t("allTime") },
  ];
  return <div className="flex max-w-full flex-col items-end gap-2">
    <div role="group" aria-label={t("period")} className="flex flex-wrap gap-1 rounded-[10px] bg-raised p-1">
      {options.map(option => <button key={option.id} type="button" aria-pressed={selected === option.id} aria-describedby="portfolio-period-status" onClick={() => setSelected(option.id)} className={`min-h-control rounded-[7px] px-4 text-sm font-medium ${selected === option.id ? "bg-surface text-text ring-1 ring-line" : "text-muted hover:text-text"}`}><NumberText text={option.label} /></button>)}
    </div>
    <p id="portfolio-period-status" role="status" className={selected === "available" ? "sr-only" : "max-w-80 text-small text-subtle"}><NumberText text={t(selected === "available" ? "showingPeriod" : "unavailablePeriod", { days })} /></p>
  </div>;
}
