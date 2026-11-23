import { apiRequest } from "@/lib/auth";
import { compareDecimals, decimalProduct } from "@/lib/format";
import { getMessages, type Locale, type MessageKey } from "@/lib/messages";

export type RiskState = "safe" | "caution" | "limit";

export interface RiskProfileInput {
  risk_per_trade_percent: string;
  target_risk_reward: string;
  total_open_risk_limit_percent: string;
  max_open_positions: number;
  max_losing_trades_today: number;
  daily_loss_limit_percent: string;
  max_stopped_trades_this_week: number;
  weekly_loss_limit_percent: string;
}

export interface RiskProfile extends RiskProfileInput { updated_at: string | null }
export type ProfileSource = "stored" | "preview";

export interface RiskCheckHistoryEntry {
  id: string;
  source: "check" | "order";
  order_id: string | null;
  symbol: string;
  side: "long" | "short";
  entry_price: string;
  allowed: boolean;
  verdict: "pass" | "caution" | "fail";
  failed_rules: string[];
  caution_rules: string[];
  profile_source: ProfileSource;
  created_at: string;
}

export function getRiskProfile(accountId: string, signal?: AbortSignal): Promise<RiskProfile> {
  return apiRequest(`/accounts/${encodeURIComponent(accountId)}/risk-profile`, undefined, signal);
}

export function putRiskProfile(accountId: string, profile: RiskProfileInput, signal?: AbortSignal): Promise<RiskProfile> {
  // Pick only the eight writable fields: PUT rejects unknown fields, including updated_at.
  const { risk_per_trade_percent, target_risk_reward, total_open_risk_limit_percent, max_open_positions,
    max_losing_trades_today, daily_loss_limit_percent, max_stopped_trades_this_week, weekly_loss_limit_percent } = profile;
  return apiRequest(`/accounts/${encodeURIComponent(accountId)}/risk-profile`, {
    risk_per_trade_percent, target_risk_reward, total_open_risk_limit_percent, max_open_positions,
    max_losing_trades_today, daily_loss_limit_percent, max_stopped_trades_this_week, weekly_loss_limit_percent,
  }, signal, "PUT");
}

export function getRiskChecks(accountId: string, limit = 20, signal?: AbortSignal): Promise<{ checks: RiskCheckHistoryEntry[] }> {
  return apiRequest(`/accounts/${encodeURIComponent(accountId)}/risk-checks?limit=${limit}`, undefined, signal);
}

export function validRiskDecimal(value: string): boolean {
  return /^[0-9]+(\.[0-9]{1,8})?$/.test(value)
    && compareDecimals(value, "0") > 0 && compareDecimals(value, "100") <= 0;
}

export function validRiskCount(value: string): boolean {
  return /^[0-9]+$/.test(value) && compareDecimals(value, "0") > 0 && compareDecimals(value, "4294967295") <= 0;
}

const ruleNames: Record<string, MessageKey<"check">> = {
  class_risk: "rule.classRisk", atr_stop: "rule.atrStop", quantity_step: "rule.quantityStep",
  target_price: "rule.targetPrice", position_limit: "rule.positionLimit", class_allocation: "rule.positionLimit", spot_side: "rule.spotSide",
  initial_margin: "rule.initialMargin", liquidation_stop: "rule.liquidationStop",
  total_open_risk: "rule.totalOpenRisk", open_positions: "rule.openPositions",
  losing_trades_today: "rule.losingTradesToday", daily_loss: "rule.dailyLoss",
  stopped_trades_this_week: "rule.stoppedTradesThisWeek", weekly_loss: "rule.weeklyLoss",
};

export function riskRuleName(ruleId: string, locale: Locale): string {
  return getMessages("check", locale)(ruleNames[ruleId] ?? "rule.other");
}

/** Caution includes its threshold; a limit is exceeded only strictly above it. Zero capacity fails closed. */
export function riskState(value: number, limit: number, cautionAt = 0.75): RiskState {
  if (!Number.isFinite(value) || !Number.isFinite(limit) || limit <= 0 || value < 0) return "limit";
  if (!Number.isFinite(cautionAt) || cautionAt < 0 || cautionAt > 1) throw new RangeError("Invalid caution fraction");
  if (value > limit) return "limit";
  return compareDecimals(String(value), decimalProduct(String(limit), String(cautionAt))) >= 0 ? "caution" : "safe";
}

