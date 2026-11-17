"use client";

import { useState, type ReactNode } from "react";
import { Card } from "@/components/metron/card";
import type { CoinClass } from "@/components/metron/class-tag";
import { getMessages } from "@/lib/messages";
import { useLocale } from "@/components/metron/locale-provider";

interface FilterableRow {
  pair: string;
  coinClass: CoinClass;
  content: ReactNode;
}
const classes = ["all", "core", "large", "small", "stable"] as const;
type ClassFilter = (typeof classes)[number];

export function FilteredMarkets({ rows }: { rows: FilterableRow[] }) {
  const t = getMessages("markets", useLocale());
  const [coinClass, setCoinClass] = useState<ClassFilter>("all");
  const [search, setSearch] = useState("");
  const query = search.trim().toLowerCase();
  const visibleRows = rows.filter((row) =>
    (coinClass === "all" || row.coinClass === coinClass) && row.pair.toLowerCase().includes(query),
  );
  return <>
    <section aria-label={t("filters")} className="flex min-w-0 flex-wrap items-center justify-between gap-x-6 gap-y-3">
      <div role="group" aria-label={t("coinClass")} className="flex flex-wrap gap-1.5">
        {classes.map((value) => <button key={value} type="button" aria-pressed={coinClass === value}
          onClick={() => setCoinClass(value)}
          className={`min-h-control rounded-control border px-4 text-sm font-medium ${coinClass === value
            ? "border-text bg-text text-inverse"
            : "border-line bg-transparent text-muted hover:bg-raised"}`}>
          {t(`class.${value}`)}
        </button>)}
      </div>
      <div className="min-w-0 flex-[0_1_280px]">
        <label htmlFor="markets-search" className="sr-only">{t("search")}</label>
        <input id="markets-search" type="search" placeholder={t("search")} value={search}
          onChange={(event) => setSearch(event.target.value)} aria-controls="markets-table"
          className="min-h-control w-full rounded-control border border-line bg-surface px-3.5 text-sm text-text placeholder:text-muted" />
      </div>
    </section>
    <Card className="overflow-hidden">
      <div role="region" aria-label={t("table.scroll")} tabIndex={0} className="overflow-x-auto">
        <table id="markets-table" className="w-full min-w-[920px] border-collapse text-sm">
          <caption className="sr-only">{t("table.caption")}</caption>
          <thead>
            <tr className="text-right text-xs text-subtle">
              <th scope="col" className="px-6 py-3.5 text-left font-medium">{t("column.pair")}</th>
              <th scope="col" className="px-3 py-3.5 text-left font-medium">{t("column.class")}</th>
              <th scope="col" className="px-3 py-3.5 font-medium">{t("column.price")}</th>
              <th scope="col" className="px-3 py-3.5 font-medium"><span className="num">{t("column.change.period")}</span>{t("column.hours")}</th>
              <th scope="col" className="px-3 py-3.5 text-left font-medium">{t("column.movement")} <span className="num">{t("column.movement.period")}</span>{t("column.movement.end")}</th>
              <th scope="col" className="px-3 py-3.5 font-medium">{t("column.stop")}</th>
              <th scope="col" className="px-3 py-3.5 font-medium"><span className="num">{t("column.change.period")}</span>{t("column.volume")}</th>
              <th scope="col" className="py-3.5 pr-6 pl-3 font-medium"><span className="sr-only">{t("column.action")}</span></th>
            </tr>
          </thead>
          <tbody>
            {visibleRows.map((row) => row.content)}
            {visibleRows.length === 0 && <tr><td colSpan={8} className="border-t border-line px-6 py-8 text-center text-muted">{t("empty")}</td></tr>}
          </tbody>
        </table>
      </div>
      <p role="status" aria-live="polite" className="sr-only num">{t("results", { count: visibleRows.length })}</p>
      <p className="border-t border-line px-6 py-3.5 text-small text-subtle">
        {t("note.start")}<span className="num">{t("note.minimum")}</span>{t("note.to")}<span className="num">{t("note.maximum")}</span>{t("note.end")}
      </p>
    </Card>
  </>;
}
