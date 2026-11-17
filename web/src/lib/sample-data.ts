import type { CoinClass } from "@/components/metron/class-tag";
import { getMessages, type Locale } from "@/lib/messages";

/** Source fixtures from design/mockups; money is stored as exact display strings. */

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

/** Build translated fixture copy per locale; exported constants below remain English. */
export function getSampleData(locale: Locale = "en") {
  const t = getMessages("common", locale);
  const account: Account = { equity: "25000.00", currency: "USDT", realizedToday: "-312.50", unrealized: "140.85", openPositions: 3, asOf: "14:20 UTC", totalOpenRisk: "687.50", checksToday: { run: 4, placed: 2, blocked: 1, abandoned: 1 }, cooldown: { pair: "AVAX/USDT", closedAt: "13:42", waitUntil: "14:42", minutesRemaining: 22, minutesWaited: 38, durationMinutes: 60, violationsThisWeek: 1 } };

  const limits: AccountLimits = { totalOpenRisk: "6.00", dailyLoss: "5.00", weeklyLoss: "10.00", maxPositions: 5, targetRiskReward: "2.0", cautionAt: 0.75 };

  const gauges: Gauge[] = [
  { label: t("sample.gauges.0.label"), value: "2.75%", limit: "6.00%", cautionAt: "4.50%", fill: "45.8%", amount: "687.50 USDT", note: t("sample.gauges.0.note") },
  { label: t("sample.gauges.1.label"), value: "1.25%", limit: "5.00%", cautionAt: "3.75%", fill: "25%", amount: "937.50 USDT", note: t("sample.gauges.1.note") },
  { label: t("sample.gauges.2.label"), value: "1.80%", limit: "10.00%", cautionAt: "7.50%", fill: "18%", amount: "2,050.00 USDT", note: t("sample.gauges.2.note") }
  ];

  const openPositions: OpenPosition[] = [
  { pair: "BTC/USDT", cls: t("sample.openPositions.0.cls"), qty: "0.08645 BTC", entry: "96,400.00", price: "97,846.00", stop: "93,508.00", toStop: "4.43%", stopFill: "74%", risk: "250.00", pl: "▲ +125.01", plColor: "profit", coinClass: "core", plAmount: "125.01" },
  { pair: "SOL/USDT", cls: t("sample.openPositions.1.cls"), qty: "15.229 SOL", entry: "182.40", price: "178.75", stop: "170.09", toStop: "4.84%", stopFill: "81%", risk: "187.50", pl: "▼ −55.59", plColor: "loss", coinClass: "large", plAmount: "-55.59" },
  { pair: "ETH/USDT", cls: t("sample.openPositions.2.cls"), qty: "1.7405 ETH", entry: "3,420.00", price: "3,461.04", stop: "3,276.36", toStop: "5.34%", stopFill: "89%", risk: "250.00", pl: "▲ +71.43", plColor: "profit", coinClass: "core", plAmount: "71.43" }
  ];

  const markets: MarketRow[] = [
  { pair: "BTC/USDT", cls: t("sample.markets.0.cls"), price: "97,846.00", change: "▲ +1.42%", chColor: "profit", atr: "2.0%", atrFill: "20%", stop: "3.0%", vol: "38.2B", coinClass: "core", changePercent: "1.42" },
  { pair: "ETH/USDT", cls: t("sample.markets.1.cls"), price: "3,461.04", change: "▲ +0.86%", chColor: "profit", atr: "2.8%", atrFill: "28%", stop: "4.2%", vol: "17.6B", coinClass: "core", changePercent: "0.86" },
  { pair: "SOL/USDT", cls: t("sample.markets.2.cls"), price: "178.75", change: "▼ −2.10%", chColor: "loss", atr: "4.5%", atrFill: "45%", stop: "6.75%", vol: "4.1B", coinClass: "large", changePercent: "-2.10" },
  { pair: "BNB/USDT", cls: t("sample.markets.3.cls"), price: "642.30", change: "▲ +0.31%", chColor: "profit", atr: "2.4%", atrFill: "24%", stop: "3.6%", vol: "1.9B", coinClass: "large", changePercent: "0.31" },
  { pair: "XRP/USDT", cls: t("sample.markets.4.cls"), price: "1.124", change: "▼ −0.74%", chColor: "loss", atr: "5.1%", atrFill: "51%", stop: "7.65%", vol: "3.3B", coinClass: "large", changePercent: "-0.74" },
  { pair: "LINK/USDT", cls: t("sample.markets.5.cls"), price: "14.82", change: "▲ +2.95%", chColor: "profit", atr: "5.4%", atrFill: "54%", stop: "8.1%", vol: "612M", coinClass: "large", changePercent: "2.95" },
  { pair: "AVAX/USDT", cls: t("sample.markets.6.cls"), price: "34.16", change: "▼ −3.40%", chColor: "loss", atr: "6.2%", atrFill: "62%", stop: "9.3%", vol: "540M", coinClass: "large", changePercent: "-3.40" },
  { pair: "PENDLE/USDT", cls: t("sample.markets.7.cls"), price: "5.204", change: "▲ +4.12%", chColor: "profit", atr: "8.0%", atrFill: "80%", stop: "12.0%", vol: "148M", coinClass: "small", changePercent: "4.12" },
  { pair: "INJ/USDT", cls: t("sample.markets.8.cls"), price: "22.47", change: "▼ −1.88%", chColor: "loss", atr: "7.4%", atrFill: "74%", stop: "11.1%", vol: "96M", coinClass: "small", changePercent: "-1.88" },
  { pair: "USDC/USDT", cls: t("sample.markets.9.cls"), price: "1.0001", change: "0.00%", chColor: "muted", atr: "0.02%", atrFill: "1%", stop: t("sample.markets.9.stop"), vol: "1.2B", coinClass: "stable", changePercent: "0.00" }
  ];

  const classRules: ClassRule[] = [
  { cls: t("sample.classRules.0.cls"), who: t("sample.classRules.0.who"), mult: "1.5×", risk: "1.00%", maxPos: "40%", total: t("sample.classRules.0.total"), source: t("sample.classRules.0.source"), coinClass: "core" },
  { cls: t("sample.classRules.1.cls"), who: t("sample.classRules.1.who"), mult: "1.5×", risk: "0.75%", maxPos: "15%", total: "40%", source: t("sample.classRules.1.source"), coinClass: "large" },
  { cls: t("sample.classRules.2.cls"), who: t("sample.classRules.2.who"), mult: "1.5×", risk: "0.50%", maxPos: "8%", total: "10%", source: t("sample.classRules.2.source"), coinClass: "small" },
  { cls: t("sample.classRules.3.cls"), who: t("sample.classRules.3.who"), mult: t("sample.classRules.3.mult"), risk: t("sample.classRules.3.risk"), maxPos: t("sample.classRules.3.maxPos"), total: t("sample.classRules.3.total"), source: t("sample.classRules.3.source"), coinClass: "stable" }
  ];

  const disciplineRules: DisciplineRule[] = [
  { name: t("sample.disciplineRules.0.name"), detail: t("sample.disciplineRules.0.detail"), mode: t("sample.disciplineRules.0.mode"), modeBg: "caution-tint", modeColor: "caution", history: t("sample.disciplineRules.0.history") },
  { name: t("sample.disciplineRules.1.name"), detail: t("sample.disciplineRules.1.detail"), mode: t("sample.disciplineRules.1.mode"), modeBg: "caution-tint", modeColor: "caution", history: t("sample.disciplineRules.1.history") },
  { name: t("sample.disciplineRules.2.name"), detail: t("sample.disciplineRules.2.detail"), mode: t("sample.disciplineRules.2.mode"), modeBg: "limit-tint", modeColor: "limit", history: t("sample.disciplineRules.2.history") }
  ];

  const journal: JournalEntry[] = [
  { pair: "AVAX/USDT", closed: t("sample.journal.0.closed"), how: t("sample.journal.0.how"), r: "−1.0R", pl: "▼ −187.50", color: "loss", rules: t("sample.journal.0.rules"), ruleColor: "muted", plAmount: "-187.50" },
  { pair: "INJ/USDT", closed: t("sample.journal.1.closed"), how: t("sample.journal.1.how"), r: "−1.0R", pl: "▼ −125.00", color: "loss", rules: t("sample.journal.1.rules"), ruleColor: "muted", plAmount: "-125.00" },
  { pair: "LINK/USDT", closed: t("sample.journal.2.closed"), how: t("sample.journal.2.how"), r: "+1.6R", pl: "▲ +300.00", color: "profit", rules: t("sample.journal.2.rules"), ruleColor: "muted", plAmount: "300.00" },
  { pair: "ETH/USDT", closed: t("sample.journal.3.closed"), how: t("sample.journal.3.how"), r: "−1.0R", pl: "▼ −250.00", color: "loss", rules: t("sample.journal.3.rules"), ruleColor: "muted", plAmount: "-250.00" },
  { pair: "SOL/USDT", closed: t("sample.journal.4.closed"), how: t("sample.journal.4.how"), r: "−1.0R", pl: "▼ −187.50", color: "loss", rules: t("sample.journal.4.rules"), ruleColor: "caution", plAmount: "-187.50" }
  ];

  const classStatistics: ClassStatistic[] = [
  { cls: t("sample.classStatistics.0.cls"), trades: "22", win: "54.5%", net: "▲ +3,140.00", color: "profit", coinClass: "core", netAmount: "3140.00" },
  { cls: t("sample.classStatistics.1.cls"), trades: "26", win: "46.2%", net: "▲ +2,480.00", color: "profit", coinClass: "large", netAmount: "2480.00" },
  { cls: t("sample.classStatistics.2.cls"), trades: "16", win: "37.5%", net: "▼ −620.00", color: "loss", coinClass: "small", netAmount: "-620.00" }
  ];

  const holdings: Holding[] = [
  { asset: "BTC", qty: "0.08645", cost: "96,400.00", fees: "8.33" },
  { asset: "ETH", qty: "1.7405", cost: "3,420.00", fees: "5.95" },
  { asset: "SOL", qty: "15.229", cost: "182.40", fees: "2.78" }
  ];

  const portfolioStatistics: PortfolioStatistics = { periodDays: 90, initialEquity: "20000.00", equity: "25000.00", returnPercent: "25.0", netResult: "5000.00", maxDrawdownPercent: "-8.4", drawdownPeak: "23400", drawdownTrough: "21434", winRate: "46.9", wins: 30, trades: 64, averageWinLoss: "1.91", averageWin: "410.33", averageLoss: "-215.00", feesPaid: "186.40", benchmarkReturnPercent: "14.2", benchmarkEndEquity: "22840" };

  const equitySeries: EquitySeries = { viewBox: "0 0 1060 300", account: [
  { x: 0, y: 225.7 },
  { x: 83, y: 210.3 },
  { x: 167, y: 188 },
  { x: 250, y: 198.3 },
  { x: 333, y: 160.6 },
  { x: 417, y: 136.6 },
  { x: 500, y: 109.1 },
  { x: 583, y: 146.9 },
  { x: 667, y: 176.5 },
  { x: 750, y: 140 },
  { x: 833, y: 102.3 },
  { x: 917, y: 78.3 },
  { x: 1000, y: 54.3 }
  ], benchmark: [
  { x: 0, y: 225.7 },
  { x: 83, y: 205.1 },
  { x: 167, y: 174.3 },
  { x: 250, y: 194.9 },
  { x: 333, y: 181.1 },
  { x: 417, y: 143.4 },
  { x: 500, y: 126.3 },
  { x: 583, y: 184.6 },
  { x: 667, y: 215.4 },
  { x: 750, y: 191.4 },
  { x: 833, y: 160.6 },
  { x: 917, y: 146.9 },
  { x: 1000, y: 128.3 }
  ], dates: [
  { x: 0, label: t("sample.equitySeries.dates.0.label") },
  { x: 333, label: t("sample.equitySeries.dates.1.label") },
  { x: 667, label: t("sample.equitySeries.dates.2.label") },
  { x: 1000, label: t("sample.equitySeries.dates.3.label") }
  ], valueTicks: [
  { y: 24, label: t("sample.equitySeries.valueTicks.0.label") },
  { y: 92.6, label: t("sample.equitySeries.valueTicks.1.label") },
  { y: 161.1, label: t("sample.equitySeries.valueTicks.2.label") },
  { y: 229.7, label: t("sample.equitySeries.valueTicks.3.label") }
  ] };

  const preTradeCaution: PreTradeCheck = { orderType: "limit", atrMultiplier: "1.5", targetRiskReward: "2.0", stopAtr: "1.5", targetAtr: "3", cautions: 2, verdict: "caution", pair: "PENDLE/USDT", coinClass: "small", atr: "8.0", last: "5.204", entry: "5.200", stop: "4.576", target: "6.448", stopPercent: "-12.0", targetPercent: "24.0", riskPercent: "0.50", quantity: "200.32", asset: "PENDLE", orderValue: "1041.67", moneyAtRisk: "125.00", stopResult: "-125.00", targetResult: "250.00", passed: 4, blocked: 0, checks: [
  { name: t("sample.preTradeCaution.checks.0.name"), detail: t("sample.preTradeCaution.checks.0.detail"), before: "45.8%", add: "8.4%", value: "2.75% → 3.25% of 6.00%", state: t("sample.preTradeCaution.checks.0.state"), color: "safe", ok: true, warn: false },
  { name: t("sample.preTradeCaution.checks.1.name"), detail: t("sample.preTradeCaution.checks.1.detail"), before: "25%", add: "10%", value: "1.25% → 1.75% of 5.00%", state: t("sample.preTradeCaution.checks.1.state"), color: "safe", ok: true, warn: false },
  { name: t("sample.preTradeCaution.checks.2.name"), detail: t("sample.preTradeCaution.checks.2.detail"), before: "0%", add: "52.1%", value: "4.17% of 8.00%", state: t("sample.preTradeCaution.checks.2.state"), color: "safe", ok: true, warn: false },
  { name: t("sample.preTradeCaution.checks.3.name"), detail: t("sample.preTradeCaution.checks.3.detail"), before: "0%", add: "41.7%", value: "0% → 4.17% of 10.00%", state: t("sample.preTradeCaution.checks.3.state"), color: "safe", ok: true, warn: false },
  { name: t("sample.preTradeCaution.checks.4.name"), detail: t("sample.preTradeCaution.checks.4.detail"), before: "60%", add: "20%", value: "3 → 4 of 5", state: t("sample.preTradeCaution.checks.4.state"), color: "caution", ok: false, warn: true },
  { name: t("sample.preTradeCaution.checks.5.name"), detail: t("sample.preTradeCaution.checks.5.detail"), before: "63.3%", add: "0%", value: "38 of 60 min waited", state: t("sample.preTradeCaution.checks.5.state"), color: "caution", ok: false, warn: true }
  ] };

  const preTradeBlocked: PreTradeCheck = { orderType: "limit", atrMultiplier: "1.5", targetRiskReward: "2.0", stopAtr: "1.5", targetAtr: "3", cautions: 1, verdict: "limit", pair: "BTC/USDT", coinClass: "core", atr: "2.0", last: "97846.00", entry: "97846.00", stop: "94910.62", target: "103716.76", stopPercent: "-3.0", targetPercent: "6.0", riskPercent: "1.00", quantity: "0.08517", asset: "BTC", orderValue: "8333.33", moneyAtRisk: "250.00", stopResult: "-250.00", targetResult: "500.00", passed: 3, blocked: 1, reducedOrderValue: "1666.67", reducedRisk: "50.00", existingPositionPercent: "33.33", proposedPositionPercent: "66.67", positionLimitPercent: "40.00", checks: [
  { name: t("sample.preTradeBlocked.checks.0.name"), detail: t("sample.preTradeBlocked.checks.0.detail"), before: "83.3%", add: "16.7%", value: "33.33% → 66.67% of 40.00%", state: t("sample.preTradeBlocked.checks.0.state"), color: "limit", ok: false, warn: false, over: true },
  { name: t("sample.preTradeBlocked.checks.1.name"), detail: t("sample.preTradeBlocked.checks.1.detail"), before: "45.8%", add: "16.7%", value: "2.75% → 3.75% of 6.00%", state: t("sample.preTradeBlocked.checks.1.state"), color: "safe", ok: true, warn: false, over: false },
  { name: t("sample.preTradeBlocked.checks.2.name"), detail: t("sample.preTradeBlocked.checks.2.detail"), before: "25%", add: "20%", value: "1.25% → 2.25% of 5.00%", state: t("sample.preTradeBlocked.checks.2.state"), color: "safe", ok: true, warn: false, over: false },
  { name: t("sample.preTradeBlocked.checks.3.name"), detail: t("sample.preTradeBlocked.checks.3.detail"), before: "60%", add: "0%", value: "3 → 3 of 5", state: t("sample.preTradeBlocked.checks.3.state"), color: "safe", ok: true, warn: false, over: false },
  { name: t("sample.preTradeBlocked.checks.4.name"), detail: t("sample.preTradeBlocked.checks.4.detail"), before: "63.3%", add: "0%", value: "38 of 60 min waited", state: t("sample.preTradeBlocked.checks.4.state"), color: "caution", ok: false, warn: true, over: false }
  ] };

  return { account, limits, openPositions, gauges, markets, classRules, disciplineRules, journal, classStatistics, holdings, portfolioStatistics, equitySeries, preTradeCaution, preTradeBlocked };
}

export const sampleData = getSampleData();
export const { account, limits, openPositions, gauges, markets, classRules, disciplineRules, journal, classStatistics, holdings, portfolioStatistics, equitySeries, preTradeCaution, preTradeBlocked } = sampleData;
export type SampleData = typeof sampleData;
