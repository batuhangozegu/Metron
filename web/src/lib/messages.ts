import common from "@/messages/en/common.json";
import overview from "@/messages/en/overview.json";
import check from "@/messages/en/check.json";
import markets from "@/messages/en/markets.json";
import portfolio from "@/messages/en/portfolio.json";
import rules from "@/messages/en/rules.json";

const messages = { common, overview, check, markets, portfolio, rules };
export type MessageNamespace = keyof typeof messages;
export type MessageKey<N extends MessageNamespace> = keyof (typeof messages)[N] & string;
export type MessageValues = Record<string, string | number>;

/** Read a typed English message and replace named {placeholders}; usable on server and client. */
export function message<N extends MessageNamespace>(namespace: N, key: MessageKey<N>, values: MessageValues = {}): string {
  const template = (messages[namespace] as Record<string, string>)[key];
  if (typeof template !== "string") throw new Error(`Missing message: ${namespace}.${key}`);
  return template.replace(/\{(\w+)\}/g, (_, name: string) => {
    if (!(name in values)) throw new Error(`Missing interpolation: ${name}`);
    return String(values[name]);
  });
}

/** Bind a namespace; its key type automatically follows edits to its JSON file. */
export function getMessages<N extends MessageNamespace>(namespace: N) {
  return (key: MessageKey<N>, values?: MessageValues) => message(namespace, key, values);
}
