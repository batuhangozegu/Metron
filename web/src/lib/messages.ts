import common from "@/messages/en/common.json";
import overview from "@/messages/en/overview.json";
import check from "@/messages/en/check.json";
import markets from "@/messages/en/markets.json";
import portfolio from "@/messages/en/portfolio.json";
import rules from "@/messages/en/rules.json";
import trCommon from "@/messages/tr/common.json";
import trOverview from "@/messages/tr/overview.json";
import trCheck from "@/messages/tr/check.json";
import trMarkets from "@/messages/tr/markets.json";
import trPortfolio from "@/messages/tr/portfolio.json";
import trRules from "@/messages/tr/rules.json";

const messages = { common, overview, check, markets, portfolio, rules };
const turkish = { common: trCommon, overview: trOverview, check: trCheck, markets: trMarkets, portfolio: trPortfolio, rules: trRules };
export type Locale = "en" | "tr";
export const localeCookie = "metron-locale";
export function parseLocale(value: string | undefined): Locale { return value === "tr" ? "tr" : "en"; }

// Check both directions: structural assignment alone would allow extra Turkish keys.
type SameKeys = { [N in keyof typeof messages]:
  Exclude<keyof (typeof messages)[N], keyof (typeof turkish)[N]> extends never
    ? Exclude<keyof (typeof turkish)[N], keyof (typeof messages)[N]> extends never ? true : false
    : false };
const sameKeys: SameKeys = { common: true, overview: true, check: true, markets: true, portfolio: true, rules: true };
void sameKeys;
export type MessageNamespace = keyof typeof messages;
export type MessageKey<N extends MessageNamespace> = keyof (typeof messages)[N] & string;
export type MessageValues = Record<string, string | number>;

/** Explicit locale keeps server requests isolated; omitted locales retain English compatibility. */
export function message<N extends MessageNamespace>(namespace: N, key: MessageKey<N>, values: MessageValues = {}, locale: Locale = "en"): string {
  const template = ((locale === "tr" ? turkish : messages)[namespace] as Record<string, string>)[key];
  if (typeof template !== "string") throw new Error(`Missing message: ${namespace}.${key}`);
  return template.replace(/\{(\w+)\}/g, (_, name: string) => {
    if (!(name in values)) throw new Error(`Missing interpolation: ${name}`);
    return String(values[name]);
  });
}

/** Bind a namespace; its key type automatically follows edits to its JSON file. */
export function getMessages<N extends MessageNamespace>(namespace: N, locale: Locale = "en") {
  return (key: MessageKey<N>, values?: MessageValues) => message(namespace, key, values, locale);
}
