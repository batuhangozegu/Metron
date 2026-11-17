"use client";

import { Moon, Sun } from "lucide-react";
import { useSyncExternalStore } from "react";
import { useLocale } from "@/components/metron/locale-provider";
import { getMessages } from "@/lib/messages";
import { THEME_STORAGE_KEY } from "@/lib/theme";

function subscribe(update: () => void) {
  const onStorage = (event: StorageEvent) => {
    if (event.key !== THEME_STORAGE_KEY && event.key !== null) return;
    document.documentElement.dataset.theme = event.newValue === "light" ? "light" : "dark";
    update();
  };
  window.addEventListener("storage", onStorage);
  window.addEventListener("metron-theme-change", update);
  return () => {
    window.removeEventListener("storage", onStorage);
    window.removeEventListener("metron-theme-change", update);
  };
}

/** Persistent theme control; CSS selects the correct icon even before hydration. */
export function ThemeToggle() {
  const locale = useLocale();
  const t = getMessages("common", locale);
  const theme = useSyncExternalStore(subscribe, () => document.documentElement.dataset.theme ?? "dark", () => "dark");

  return <button type="button" className="theme-toggle inline-flex size-11 shrink-0 items-center justify-center rounded-control border border-line text-muted hover:bg-raised"
    aria-label={t("theme.toggle")} aria-pressed={theme === "light"}
    onClick={() => {
      const next = document.documentElement.dataset.theme === "light" ? "dark" : "light";
      document.documentElement.dataset.theme = next;
      try { localStorage.setItem(THEME_STORAGE_KEY, next); } catch { /* Keep theme usable without storage. */ }
      window.dispatchEvent(new Event("metron-theme-change"));
    }}>
    <Sun aria-hidden="true" className="theme-light-icon size-5" />
    <Moon aria-hidden="true" className="theme-dark-icon size-5" />
  </button>;
}
