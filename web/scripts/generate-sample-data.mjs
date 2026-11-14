import { readFileSync, writeFileSync } from "node:fs";

const root = new URL("../", import.meta.url);
const read = (path) => readFileSync(new URL(path, root), "utf8");
const write = (path, text) => writeFileSync(new URL(path, root), text);
const common = JSON.parse(read("src/messages/en/common.json"));
function load(name) {
  const body = read(`../design/mockups/${name}.dc.html`).replaceAll("\r", "").split("renderVals() {")[1].split("\n}\n}")[0];
  // Trusted repository design sources, evaluated only by this optional development generator.
  return new Function(body).call({ props: {} });
}
const main = load("Main"), marketData = load("Markets"), portfolio = load("Portfolio"), rules = load("Rules");
const classes = { "BTC/ETH": "core", "Large-cap": "large", "Small-cap": "small", Stablecoin: "stable" };
const translated = new Set(["label", "note", "cls", "who", "source", "mode", "history", "how", "rules", "name", "detail", "state", "closed"]);
function serialize(value, path, field = "") {
  if (Array.isArray(value)) return "[\n" + value.map((item, i) => serialize(item, `${path}.${i}`)).join(",\n") + "\n]";
  if (value && typeof value === "object") {
    const fields = Object.entries(value).filter(([key]) => key !== "theme").map(([key, item]) => `${key}: ${serialize(item, `${path}.${key}`, key)}`);
    if (value.cls) fields.push(`coinClass: ${JSON.stringify(classes[value.cls])}`);
    // Keep the exact source display, and provide undecorated decimals for PnL/format helpers.
    for (const [display, decimal] of [["pl", "plAmount"], ["change", "changePercent"], ["net", "netAmount"]]) {
      if (value[display]) fields.push(`${decimal}: ${JSON.stringify(value[display].replace(/[▲▼+%,\s]/g, "").replace("−", "-"))}`);
    }
    return "{ " + fields.join(", ") + " }";
  }
  if (typeof value === "string") {
    if (value.startsWith("var(--")) return JSON.stringify(value.replace("var(--", "").replace(")", "").replace("-t", "-tint"));
    if (translated.has(field) || value === "None" || value === "No cap") {
      const key = `sample.${path}`;
      common[key] = value;
      return `t(${JSON.stringify(key)})`;
    }
  }
  return JSON.stringify(value);
}
let source = `import type { CoinClass } from "@/components/metron/class-tag";
import { getMessages } from "@/lib/messages";

/** Source fixtures from design/mockups; money is stored as exact display strings. */
const t = getMessages("common");
export type DisplayTone = "profit" | "loss" | "muted" | "safe" | "caution" | "limit";
export type TintTone = "caution-tint" | "limit-tint";
export interface ChecksToday { run: number; placed: number; blocked: number; abandoned: number }
export interface Cooldown { pair: string; closedAt: string; waitUntil: string; minutesRemaining: number; minutesWaited: number; durationMinutes: number; violationsThisWeek: number }
export interface Account { equity: string; currency: string; realizedToday: string; unrealized: string; openPositions: number; asOf: string; totalOpenRisk: string; checksToday: ChecksToday; cooldown: Cooldown }
export interface AccountLimits { totalOpenRisk: string; dailyLoss: string; weeklyLoss: string; maxPositions: number; targetRiskReward: string; cautionAt: number }
export interface Gauge { label: string; value: string; limit: string; cautionAt: string; fill: string; amount: string; note: string }
export interface OpenPosition { pair: string; cls: string; coinClass: CoinClass; qty: string; entry: string; price: string; stop: string; toStop: string; stopFill: string; risk: string; pl: string; plAmount: string; plColor: DisplayTone }
export interface MarketRow { pair: string; cls: string; coinClass: CoinClass; price: string; change: string; changePercent: string; chColor: DisplayTone; atr: string; atrFill: string; stop: string; vol: string }
export interface ClassRule { cls: string; coinClass: CoinClass; who: string; mult: string; risk: string; maxPos: string; total: string; source: string }
export interface DisciplineRule { name: string; detail: string; mode: string; modeBg: TintTone; modeColor: DisplayTone; history: string }
export interface JournalEntry { pair: string; closed: string; how: string; r: string; pl: string; plAmount: string; color: DisplayTone; rules: string; ruleColor: DisplayTone }
export interface ClassStatistic { cls: string; coinClass: CoinClass; trades: string; win: string; net: string; netAmount: string; color: DisplayTone }
export interface Holding { asset: string; qty: string; cost: string; fees: string }
export interface LimitCheck { name: string; detail: string; before: string; add: string; value: string; state: string; color: DisplayTone; ok: boolean; warn: boolean; over?: boolean }
export interface PreTradeCheck { verdict: "caution" | "limit"; pair: string; coinClass: CoinClass; orderType: "limit"; atr: string; last: string; entry: string; stop: string; target: string; stopPercent: string; targetPercent: string; stopAtr: string; targetAtr: string; riskPercent: string; atrMultiplier: string; targetRiskReward: string; quantity: string; asset: string; orderValue: string; moneyAtRisk: string; stopResult: string; targetResult: string; passed: number; cautions: number; blocked: number; checks: LimitCheck[]; reducedOrderValue?: string; reducedRisk?: string; existingPositionPercent?: string; proposedPositionPercent?: string; positionLimitPercent?: string }
export interface PortfolioStatistics { periodDays: number; initialEquity: string; equity: string; returnPercent: string; netResult: string; maxDrawdownPercent: string; drawdownPeak: string; drawdownTrough: string; winRate: string; wins: number; trades: number; averageWinLoss: string; averageWin: string; averageLoss: string; feesPaid: string; benchmarkReturnPercent: string; benchmarkEndEquity: string }
/** The source provides SVG geometry rather than dated monetary observations; points are copied exactly. */
export interface EquityPoint { x: number; y: number }
export interface EquityAxisLabel { x: number; label: string }
export interface EquityValueTick { y: number; label: string }
export interface EquitySeries { viewBox: string; account: EquityPoint[]; benchmark: EquityPoint[]; dates: EquityAxisLabel[]; valueTicks: EquityValueTick[] }
`;
function emit(name, type, data) { source += `\nexport const ${name}: ${type} = ${serialize(data, name)};\n`; }
emit("account", "Account", {
  equity: "25000.00", currency: "USDT", realizedToday: "-312.50", unrealized: "140.85", openPositions: 3, asOf: "14:20 UTC", totalOpenRisk: "687.50",
  checksToday: { run: 4, placed: 2, blocked: 1, abandoned: 1 },
  cooldown: { pair: "AVAX/USDT", closedAt: "13:42", waitUntil: "14:42", minutesRemaining: 22, minutesWaited: 38, durationMinutes: 60, violationsThisWeek: 1 },
});
emit("limits", "AccountLimits", { totalOpenRisk: "6.00", dailyLoss: "5.00", weeklyLoss: "10.00", maxPositions: 5, targetRiskReward: "2.0", cautionAt: 0.75 });
for (const [name, type, data] of [
  ["gauges", "Gauge[]", main.gauges], ["openPositions", "OpenPosition[]", main.positions], ["markets", "MarketRow[]", marketData.rows],
  ["classRules", "ClassRule[]", rules.classes], ["disciplineRules", "DisciplineRule[]", rules.discipline],
  ["journal", "JournalEntry[]", portfolio.journal], ["classStatistics", "ClassStatistic[]", portfolio.classes], ["holdings", "Holding[]", portfolio.holdings],
]) emit(name, type, data);
emit("portfolioStatistics", "PortfolioStatistics", {
  periodDays: 90, initialEquity: "20000.00", equity: "25000.00", returnPercent: "25.0", netResult: "5000.00", maxDrawdownPercent: "-8.4", drawdownPeak: "23400", drawdownTrough: "21434", winRate: "46.9", wins: 30, trades: 64, averageWinLoss: "1.91", averageWin: "410.33", averageLoss: "-215.00", feesPaid: "186.40", benchmarkReturnPercent: "14.2", benchmarkEndEquity: "22840",
});
const points = [...read("../design/mockups/Portfolio.dc.html").matchAll(/<polyline points="([^"]+)"/g)].map((match) => match[1].split(" ").map((point) => {
  const [x, y] = point.split(",").map(Number); return { x, y };
}));
emit("equitySeries", "EquitySeries", {
  viewBox: "0 0 1060 300", account: points[1], benchmark: points[0],
  dates: [{ x: 0, label: "16 Aug" }, { x: 333, label: "15 Sep" }, { x: 667, label: "15 Oct" }, { x: 1000, label: "14 Nov" }],
  valueTicks: [{ y: 24, label: "26k" }, { y: 92.6, label: "24k" }, { y: 161.1, label: "22k" }, { y: 229.7, label: "20k" }],
});
const base = { orderType: "limit", atrMultiplier: "1.5", targetRiskReward: "2.0", stopAtr: "1.5", targetAtr: "3", cautions: 1 };
emit("preTradeCaution", "PreTradeCheck", {
  ...base, verdict: "caution", pair: "PENDLE/USDT", coinClass: "small", atr: "8.0", last: "5.204", entry: "5.200", stop: "4.576", target: "6.448", stopPercent: "-12.0", targetPercent: "24.0", riskPercent: "0.50", quantity: "200.32", asset: "PENDLE", orderValue: "1041.67", moneyAtRisk: "125.00", stopResult: "-125.00", targetResult: "250.00", passed: 5, blocked: 0, checks: load("PreTrade").checks,
});
emit("preTradeBlocked", "PreTradeCheck", {
  ...base, verdict: "limit", pair: "BTC/USDT", coinClass: "core", atr: "2.0", last: "97846.00", entry: "97846.00", stop: "94910.62", target: "103716.76", stopPercent: "-3.0", targetPercent: "6.0", riskPercent: "1.00", quantity: "0.08517", asset: "BTC", orderValue: "8333.33", moneyAtRisk: "250.00", stopResult: "-250.00", targetResult: "500.00", passed: 3, blocked: 1, reducedOrderValue: "1666.67", reducedRisk: "50.00", existingPositionPercent: "33.33", proposedPositionPercent: "66.67", positionLimitPercent: "40.00", checks: load("PreTradeBlocked").checks,
});
source += `\nexport const sampleData = { account, limits, openPositions, gauges, markets, classRules, disciplineRules, journal, classStatistics, holdings, portfolioStatistics, equitySeries, preTradeCaution, preTradeBlocked };\nexport type SampleData = typeof sampleData;\n`;
write("src/lib/sample-data.ts", source);
write("src/messages/en/common.json", JSON.stringify(common, null, 2) + "\n");
