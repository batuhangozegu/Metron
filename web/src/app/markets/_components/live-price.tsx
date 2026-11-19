"use client";

import { useState } from "react";
import { compareDecimals, formatTradePrice } from "@/lib/format";
import { getMessages } from "@/lib/messages";
import { useLocale } from "@/components/metron/locale-provider";

/** router.refresh preserves this component's state, so ticks compare exact prices. */
export function LivePrice({ value }: { value: string | null }) {
  const locale = useLocale();
  const [tick, setTick] = useState({ value, direction: 0, revision: 0 });
  if (tick.value !== value) {
    setTick({ value, direction: value !== null && tick.value !== null ? compareDecimals(value, tick.value) : 0, revision: tick.revision + 1 });
  }
  return <span key={tick.revision} className="price-tick" data-direction={tick.direction > 0 ? "up" : tick.direction < 0 ? "down" : undefined}>
    <span aria-hidden="true" className="price-direction w-2 text-[9px]">{tick.direction > 0 ? "▲" : tick.direction < 0 ? "▼" : ""}</span>
    {value === null ? getMessages("markets", locale)("missing") : formatTradePrice(value, locale)}
  </span>;
}
