"use client";

import { createContext, useContext, type ReactNode } from "react";
import type { Locale } from "@/lib/messages";

const LocaleContext = createContext<Locale>("en");
export function LocaleProvider({ locale, children }: { locale: Locale; children: ReactNode }) {
  return <LocaleContext value={locale}>{children}</LocaleContext>;
}
export function useLocale() { return useContext(LocaleContext); }
