import type { Metadata } from "next";
import { Suspense, type ReactNode } from "react";
import { Inter, JetBrains_Mono, Oswald } from "next/font/google";
import { AppShell } from "@/components/metron/app-shell";
import { getMessages } from "@/lib/messages";
import { themeBootstrap } from "@/lib/theme";
import { getLocale } from "@/lib/locale-server";
import { LocaleProvider } from "@/components/metron/locale-provider";
import "./globals.css";

const inter = Inter({
  variable: "--font-inter",
  subsets: ["latin"],
});

const jetBrainsMono = JetBrains_Mono({
  variable: "--font-jetbrains-mono",
  subsets: ["latin"],
});

// Latin Extended includes Turkish dotted/dotless I and ş, ğ, ç, ö, ü.
const display = Oswald({
  variable: "--font-display",
  subsets: ["latin", "latin-ext"],
  weight: ["600", "700"],
  display: "swap",
});

export const metadata: Metadata = {
  title: getMessages("common")("brand"),
  description: getMessages("common")("description"),
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return <Suspense fallback={<html lang="en"><body /></html>}><LocalizedDocument>{children}</LocalizedDocument></Suspense>;
}

/** The cookie determines even html.lang, so Suspense encloses the entire runtime
 * document read. This trades a localized static shell for request-rendered HTML;
 * server copy gets an explicit locale and client controls receive it via context. */
async function LocalizedDocument({ children }: { children: ReactNode }) {
  const locale = await getLocale();
  return (
    <html
      lang={locale}
      data-theme="dark"
      suppressHydrationWarning
      className={`${inter.variable} ${jetBrainsMono.variable} ${display.variable} h-full antialiased`}
    >
      <head><script dangerouslySetInnerHTML={{ __html: themeBootstrap }} /></head>
      <body className="min-h-full"><LocaleProvider locale={locale}><AppShell locale={locale}>{children}</AppShell></LocaleProvider></body>
    </html>
  );
}
